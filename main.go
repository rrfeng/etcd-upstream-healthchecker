package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var ETCD *string = flag.String("e", "127.0.0.1:2379", "The etcd endpoints")
var SVCDIR *string = flag.String("s", "/v1/pre/services", "The service dir in etcd")
var TIMEOUT *int = flag.Int("t", 1, "Time out of the check")
var INTERVAL *int = flag.Int("i", 10, "Interval of check")
var CONCURRENCY *int = flag.Int("c", 20, "Concurrency of check")

func HandleResult(c chan CheckResult) {
	for res := range c {
		if res.Result.Error() == "Peer Up" {
			log.Println(res.Target.Name, res.Target.Ip, res.Target.Port, "Peer recover!")
			SetPeerUp(res.Target)
		} else {
			log.Println(res.Target.Name, res.Target.Ip, res.Target.Port, "Checked down!", res.Result.Error())
			SetPeerDown(res.Target)
		}
	}
}

func RunCheck(ck chan Checker, cp chan *Peer, ce chan CheckResult) {
	for {
		c := <-ck
		go func(c Checker) {
			p := <-cp
			err := c.Check(p)
			if err != nil {
				result := CheckResult{Target: p, Result: err}
				ce <- result
			}
			ck <- c
		}(c)
	}
}

func main() {
	flag.Parse()

	ps := [](*Peer){}
	for i := 0; i < 100; i++ {
		p := Peer{Name: "hockey", Ip: "172.16.1.5", Port: 3000 + i, Info: &PeerInfo{}}
		ps = append(ps, &p)
	}

	fmt.Println(len(ps))

	chTask := make(chan *Peer)
	chCPool := make(chan Checker, *CONCURRENCY)
	chResult := make(chan CheckResult)

	for i := 0; i < *CONCURRENCY; i++ {
		c := Checker{}
		chCPool <- c
	}

	go HandleResult(chResult)
	go RunCheck(chCPool, chTask, chResult)

	for {
		for _, p := range ps {
			chTask <- p
		}
		time.Sleep(time.Second * time.Duration(*INTERVAL))
	}
}
