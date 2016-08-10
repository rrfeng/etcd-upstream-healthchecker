package main

import (
	"flag"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strings"
	"time"
)

var ETCD *string = flag.String("e", "127.0.0.1:2379", "The etcd endpoints")
var SVCDIR *string = flag.String("s", "/v1/pre/services", "The service dir in etcd")
var TIMEOUT *int64 = flag.Int64("t", 1000, "Timeout of the check, by ms, at least 10")
var INTERVAL *int64 = flag.Int64("i", 5000, "Interval of check, by ms, at least 100")
var CONCURRENCY *int = flag.Int("c", 50, "Concurrency of check")

func main() {
	flag.Parse()

	if *INTERVAL < 100 || *TIMEOUT < 10 {
		log.Fatalln("Timeout or interval out of range, may cause problem.")
	}

	PATH := strings.TrimSuffix(*SVCDIR, "/")

	var cfg = client.Config{
		Endpoints: strings.Split(*ETCD, ","),
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatalln(err.Error())
	}

	etcd := EtcdApi{
		Api:  client.NewKeysAPI(c),
		Ctx:  context.Background(),
		Path: PATH,
	}

	uss, u_index, err := etcd.FetchUpstreamList()
	if err != nil {
		log.Fatalln("Fetch Upstream: ", err.Error())
	}

	ps, index, err := etcd.FetchUpstreamPeers(uss)
	if err != nil {
		log.Fatalln("Fetch Peers: ", err.Error())
	}

	if u_index > index {
		index = u_index
	}

	go etcd.StartWatch(index, &ps)

	chTask := make(chan *Peer)
	chChecker := make(chan Checker, *CONCURRENCY)
	chResult := make(chan CheckResult)

	for i := 0; i < *CONCURRENCY; i++ {
		c := Checker{Client: http.Client{Timeout: time.Duration(*TIMEOUT) * time.Millisecond}}
		chChecker <- c
	}

	go HandleResult(etcd, chResult)
	go RunCheck(chChecker, chTask, chResult)

	for {
		for _, p := range ps {
			chTask <- p
		}
		time.Sleep(time.Duration(*INTERVAL) * time.Millisecond)
	}
}
