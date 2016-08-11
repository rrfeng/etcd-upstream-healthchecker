package main

import (
	"testing"
)

func Test_EqualTo(t *testing.T) {
	p1 := &Peer{Name: "name1", Ip: "1.2.3.4", Port: 15, Info: &PeerInfo{}}
	p2 := &Peer{Name: "name1", Ip: "1.2.3.4", Port: 15, Info: &PeerInfo{}}
	p3 := &Peer{Name: "name2", Ip: "1.2.3.4", Port: 15, Info: &PeerInfo{}}
	p4 := &Peer{Name: "name2", Ip: "1.2.3.4", Port: 16}

	if p1.EqualTo(p2) {
		t.Log("Test passed.")
	}

	if p3.EqualTo(p4) {
		t.Error("You shall not pass.")
	}
}
