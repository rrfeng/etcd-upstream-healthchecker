package main

import (
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"strconv"
	"strings"
	"time"
)

type EtcdApi struct {
	Api  client.KeysAPI
	Ctx  context.Context
	Path string
}

func (e *EtcdApi) FetchUpstreamList() ([]string, uint64, error) {
	var index uint64
	uss := []string{}

	ops := &client.GetOptions{Recursive: false}

	resp, err := e.Api.Get(e.Ctx, e.Path, ops)

	if err != nil {
		return nil, 0, err
	}

	if !resp.Node.Dir {
		return nil, index, errors.New("Not a dir")
	}

	index = resp.Node.ModifiedIndex

	for _, n := range resp.Node.Nodes {

		if n.ModifiedIndex > index {
			index = n.ModifiedIndex
		}

		if !n.Dir {
			continue
		}

		uss = append(uss, n.Key)
	}

	return uss, index, nil
}

func (e *EtcdApi) FetchUpstreamPeers(uss []string) ([](*Peer), uint64, error) {
	var index uint64 = 0
	peers := [](*Peer){}

	ops := &client.GetOptions{Recursive: false}

	for _, us := range uss {

		resp, err := e.Api.Get(e.Ctx, us, ops)

		if err != nil {
			return peers, index, err
		}

		if !resp.Node.Dir {
			continue
		}

		for _, n := range resp.Node.Nodes {
			if n.ModifiedIndex > index {
				index = n.ModifiedIndex
			}

			if n.Dir {
				continue
			}

			peer, err := NewPeer(n.Key, n.Value)
			if err != nil {
				return peers, index, err
			}

			ok, _ := peer.IndexOf(peers)
			if !ok {
				peers = append(peers, peer)
			}
		}
	}

	return peers, index, nil
}

func (e *EtcdApi) SetPeerDown(p *Peer) error {

	if p.Info.Status == "down" {
		return nil
	}

	p.Info.Status = "down"

	info, err := json.Marshal(p.Info)
	if err != nil {
		return err
	}

	value := string(info)
	key := e.Path + "/" + p.Name + "/" + p.Ip + ":" + strconv.Itoa(p.Port)

	_, err = e.Api.Set(e.Ctx, key, value, nil)
	if err != nil {
		return err
	}

	return nil
}

func (e *EtcdApi) SetPeerUp(p *Peer) error {

	if p.Info.Status == "up" {
		return nil
	}

	p.Info.Status = "up"

	info, err := json.Marshal(p.Info)
	if err != nil {
		return err
	}

	value := string(info)
	key := e.Path + "/" + p.Name + "/" + p.Ip + ":" + strconv.Itoa(p.Port)

	_, err = e.Api.Set(e.Ctx, key, value, nil)
	if err != nil {
		return err
	}

	return nil
}

func (e *EtcdApi) StartWatch(index uint64, ps *[](*Peer)) {
	opts := &client.WatcherOptions{AfterIndex: index, Recursive: true}
	w := e.Api.Watcher(e.Path, opts)
	go func(w client.Watcher, ps *[](*Peer)) {
		for {
			resp, err := w.Next(e.Ctx)
			if err != nil {
				log.Println("Watching:", e.Path, err.Error())
				time.Sleep(1)
				continue
			}

			if resp.Node.Dir {
				continue
			}

			p, err := NewPeer(resp.Node.Key, resp.Node.Value)
			if err != nil {
				log.Println("Watching:", e.Path, err.Error())
				continue
			}

			if resp.Action == "delete" {
				ok, i := p.IndexOf(*ps)
				if ok {
					(*ps)[i] = (*ps)[len(*ps)-1]
					(*ps)[len(*ps)-1] = nil
					*ps = (*ps)[:len(*ps)-1]
					log.Println("Delete a peer:", p)
				}
			} else {
				ok, i := p.IndexOf(*ps)
				if ok {
					(*ps)[i] = p
				} else {
					*ps = append(*ps, p)
					log.Println("Add a peer:", p)
				}
			}
		}
	}(w, ps)
}

func NewPeer(k, v string) (*Peer, error) {

	p := &Peer{}

	s := strings.Split(k, "/")
	name := s[len(s)-2]

	ip, port, err := ParseIpPort(k)
	if err != nil {
		return p, err
	}

	info, err := ParsePeerInfo(v)
	if err != nil {
		return p, err
	}

	p = &Peer{Name: name, Ip: ip, Port: port, Info: &info}

	return p, nil
}

func ParseIpPort(key string) (string, int, error) {
	key = strings.TrimRight(key, "/")
	s := strings.Split(key, "/")
	ip_port := strings.Split(s[len(s)-1], ":")
	if len(ip_port) < 2 {
		return "127.0.0.1", 0, errors.New(s[len(s)-1] + " is not correct ip:port.")
	}

	ip := ip_port[0]
	v := strings.Split(ip, ".")
	if len(v) != 4 {
		return "127.0.0.1", 0, errors.New(ip + " is not correct ip address.")
	} else {
		for _, i := range v {
			t, err := strconv.Atoi(i)
			if err != nil || t > 255 || t < 0 {
				return "127.0.0.1", 0, errors.New(ip + " is not correct ip address.")
			}
		}
	}

	port, err := strconv.Atoi(ip_port[1])
	if err != nil || port > 65535 || port < 1 {
		return ip, 0, errors.New(ip_port[1] + " is not correct ip address.")
	}

	return ip, port, nil
}

func ParsePeerInfo(s string) (PeerInfo, error) {
	pi := PeerInfo{}
	err := json.Unmarshal([]byte(s), &pi)
	if err != nil {
		pi.Weight = 1
		pi.Status = "up"
	}

	return pi, nil
}
