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

var c *string = flag.String("c", "./default.yml", "The config file path.")

func main() {
	flag.Parse()

	config, err := ReadConfig(*c)
	if err != nil {
		log.Fatalln("Read config error:", err.Error())
	}

	PATH := strings.TrimSuffix(config.ServiceDir, "/")

	var cfg = client.Config{
		Endpoints: config.EtcdEndpoints,
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
	chChecker := make(chan Checker, config.Concurrency)
	chResult := make(chan CheckResult)

	for i := 0; i < config.Concurrency; i++ {
		c := Checker{Client: http.Client{Timeout: time.Duration(config.CheckTimeout) * time.Millisecond}}
		chChecker <- c
	}

	go HandleResult(etcd, chResult, config)
	go RunCheck(chChecker, chTask, chResult, config)

	for {
		for _, p := range ps {
			chTask <- p
		}
		time.Sleep(time.Duration(config.CheckInterval) * time.Millisecond)
	}
}
