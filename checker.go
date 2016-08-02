package main

import (
	"errors"
	"net/http"
	"strconv"
)

type Checker struct {
}

type CheckResult struct {
	Target *Peer
	Result error
}

var OkStatus = map[int]bool{100: true,
	101: true,
	200: true,
	201: true,
	202: true,
	203: true,
	204: true,
	205: true,
	206: true,
	300: true,
	301: true,
	302: true,
	303: true,
	304: true,
	305: true,
	307: true,
	400: true,
	401: true,
	402: true,
	403: true,
	404: true,
	405: true,
	406: true,
	407: true,
	408: true,
	409: true,
	410: true,
	411: true,
	412: true,
	413: true,
	414: true,
	415: true,
	416: true,
	417: true,
	418: true,
	428: true,
	429: true,
	431: true,
	451: true,
	500: false,
	501: false,
	502: false,
	503: false,
	504: false,
	505: false,
	511: false}

func (_ *Checker) Check(p *Peer) error {
	url := p.GetCheckUrl()
	resp, err := http.Get(url)

	if err != nil {
		return err
	} else if OkStatus[resp.StatusCode] == false {
		return errors.New("Return Status: " + strconv.Itoa(resp.StatusCode))
	}

	if p.Info.Weight != p.Info.CheckWeight && OkStatus[resp.StatusCode] == true {
		return errors.New("Peer Up")
	}

	return nil
}
