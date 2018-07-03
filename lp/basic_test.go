package lp

import (
	"fmt"
	"testing"

	ci "github.com/libp2p/go-libp2p-crypto"
)

var (
	listenAddress = "/ip4/0.0.0.0/tcp/9090"
)

func TestInitBasicHost(t *testing.T) {
	err := initBasicHost()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitCustomHost(t *testing.T) {
	host, _, err := initCustomHost(listenAddress, ci.RSA, 2048)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", host)

	for _, v := range host.Addrs() {
		fmt.Println(v)
	}
}
