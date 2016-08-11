package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
)

type Checker struct {
	Client http.Client
}

type CheckResult struct {
	Target *Peer
	Result error
}

func (c *Checker) Check(p *Peer, config *Config) error {
	url := p.GetCheckUrl()
	resp, err := c.Client.Get(url)

	if err != nil {
		return err
	} else if !StatusIn(resp.StatusCode, config.OkStatus) {
		return errors.New("Return Status: " + strconv.Itoa(resp.StatusCode))
	}

	if p.Info.Status == "down" && StatusIn(resp.StatusCode, config.OkStatus) {
		return errors.New("Peer Up")
	}

	return nil
}

func StatusIn(s int, as []int) bool {
	l := len(as)
	for i := 0; i < l; i++ {
		if as[i] == s {
			return true
		}
	}

	return false
}

func HandleResult(etcd EtcdApi, c chan CheckResult, config *Config) {
	for res := range c {
		if res.Result.Error() == "Peer Up" {
			err := etcd.SetPeerUp(res.Target)
			if err != nil {
				log.Println("Set peer up error: ", err.Error())
			} else {
				log.Println("Peer recover:", res.Target)
			}
		} else {
			if res.Target.Fails >= config.MaxFails {
				err := etcd.SetPeerDown(res.Target)
				if err != nil {
					log.Println("Set peer down error: ", err.Error())
				} else {
					log.Println("Checked down:", res.Target, res.Result.Error())
				}
			} else {
				log.Println("Peer fail: ", res.Target, res.Result.Error())
			}
		}
	}
}

func RunCheck(ck chan Checker, cp chan *Peer, ce chan CheckResult, config *Config) {
	for {
		c := <-ck
		go func(c Checker) {
			p := <-cp
			err := c.Check(p, config)
			if err != nil {
				p.Fails++
				result := CheckResult{Target: p, Result: err}
				ce <- result
			} else {
				p.Fails = 0
			}
			ck <- c
		}(c)
	}
}
