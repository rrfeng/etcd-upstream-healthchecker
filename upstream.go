package main

import (
	"fmt"
)

type Peer struct {
	Name  string
	Ip    string
	Port  int
	Info  *PeerInfo
	Fails int
}

type PeerInfo struct {
	CheckUrl string `json:"checkurl"`
	Weight   int    `json:"weight"`
	Status   string `json:"status"`
}

func (p *Peer) GetCheckUrl() string {
	if p.Info.CheckUrl != "" {
		return fmt.Sprintf("http://%s:%d%s", p.Ip, p.Port, p.Info.CheckUrl)
	} else {
		return fmt.Sprintf("http://%s:%d/", p.Ip, p.Port)
	}
}

func (p *Peer) EqualTo(s *Peer) bool {
	if p.Name != s.Name || p.Ip != s.Ip || p.Port != s.Port {
		return false
	}
	return true
}

func (p *Peer) IndexOf(ps [](*Peer)) (bool, int) {
	l := len(ps)
	for i := 0; i < l; i++ {
		if p.EqualTo(ps[i]) {
			return true, i
		}
	}
	return false, l
}

func (p *Peer) String() string {
	return fmt.Sprintf("[%s %s:%d fails:%d]", p.Name, p.Ip, p.Port, p.Fails)
}
