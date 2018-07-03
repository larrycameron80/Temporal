package lp

import (
	"context"
	"fmt"
	"testing"

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
	lpm, err := GenerateLibPeerManager(listenAddress, ci.RSA, 1024)
	if err != nil {
		t.Fatal(err)
	}
	lpm.GenerateTemporalProtocol()
	go func() { lpm.Run("") }()
	for _, v := range lpm.Host.Addrs() {
		fmt.Println(v)
	}
	p1 := lpm.Host.Peerstore().Peers()[0]
	_, err = lpm.Host.NewStream(context.TODO(), p1, TemporalProtocol)
	if err != nil {
		t.Fatal(err)
	}
}
