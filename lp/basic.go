package lp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/RTradeLtd/Temporal/utils"
	libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
)

const (
	TemporalProtocol = "/temporal/1.0.0"
)

// LibPeerManager is a generalized Temporal libp2p host
type LibPeerManager struct {
	PrivateKey ci.PrivKey
	Host       host.Host
}

// GenerateLibPeerManager is used to generate our basic Temporal libp2p host
func GenerateLibPeerManager(listenAddress string, keyType, keyBits int) (*LibPeerManager, error) {
	host, pk, err := initCustomHost(listenAddress, keyType, keyBits)
	if err != nil {
		return nil, err
	}
	lpm := LibPeerManager{
		PrivateKey: pk,
		Host:       host,
	}
	return &lpm, nil
}

func (lpm *LibPeerManager) Run(targetAddress string) {
	if targetAddress == "" {
		// run for select
		for {
		}
	}
}
func (lpm *LibPeerManager) GenerateTemporalProtocol() {
	lpm.Host.SetStreamHandler(TemporalProtocol, func(s net.Stream) {
		log.Println("Got a new stream!")
		if err := doEcho(s); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})
}

// InitBasicHost is used to generate a basic host
func initBasicHost() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host, err := libp2p.New(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Host ID is %s\n", host.ID())
	return nil
}

// InitCustomHost is used to generate a customized libp2p host
func initCustomHost(listenAddress string, keyType, keyBits int) (host.Host, ci.PrivKey, error) {
	// vlaidate the key parameters
	if keyType == ci.RSA && keyBits > 4096 {
		return nil, nil, errors.New("for RSA keys larger than 4096 contact your temporal admin")
	}
	// validate the listen address
	_, err := utils.GenerateMultiAddrFromString(listenAddress)
	if err != nil {
		return nil, nil, err
	}
	pk, _, err := ci.GenerateKeyPair(keyType, keyBits)
	if err != nil {
		return nil, nil, err
	}

	//  create a background context for which the host will run in
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// generate our options
	opts := []libp2p.Option{
		libp2p.Identity(pk),
		libp2p.ListenAddrStrings(listenAddress),
	}
	// generate our libp2p host with custom options
	host, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	return host, pk, nil
}

// doEcho reads a line of data a stream and writes it back
// borrowed from https://github.com/libp2p/go-libp2p/blob/master/examples/echo/main.go
func doEcho(s net.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}
