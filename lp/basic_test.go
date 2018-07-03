package lp

import (
	"context"
	"testing"
	"time"

	ci "github.com/libp2p/go-libp2p-crypto"
)

var (
	listenAddress = "/ip4/0.0.0.0/tcp/9090"
	targetAddress = "/ip4/127.0.0.1/9090"
)

func TestInitBasicHost(t *testing.T) {
	err := initBasicHost()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitCustomHost(t *testing.T) {

	go func() {
		lpm, err := GenerateLibPeerManager(listenAddress, ci.RSA, 1024)
		if err != nil {
			t.Fatal(err)
		}
		lpm.GenerateTemporalProtocol()
		go func() { lpm.Run("") }()
		p1 := lpm.Host.Peerstore().Peers()[0]
		lpm.Host.Peerstore().AddAddr(p1, lpm.Host.Addrs()[1], time.Hour)
		_, err = lpm.Host.NewStream(context.TODO(), p1, TemporalProtocol)
		if err != nil {
			t.Fatal(err)
		}
		for {
		}
	}()
	lpm, err := GenerateLibPeerManager(listenAddress, ci.RSA, 1024)
	if err != nil {
		t.Fatal(err)
	}
	lpm.GenerateTemporalProtocol()
	go func() { lpm.Run("") }()
	p1 := lpm.Host.Peerstore().Peers()[0]
	lpm.Host.Peerstore().AddAddr(p1, lpm.Host.Addrs()[1], time.Hour)
	_, err = lpm.Host.NewStream(context.TODO(), p1, TemporalProtocol)
	if err != nil {
		t.Fatal(err)
	}
	for {
	}
}
