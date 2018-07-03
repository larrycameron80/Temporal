package lp2p

import (
	"context"
	"errors"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	//libp2p "github.com/libp2p/go-libp2p"
	ci "github.com/libp2p/go-libp2p-crypto"
)

func InitBasicHost() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host, err := libp2p.New(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Host ID is %s\n", host.ID())
	return nil
}

func InitCustomHost(listenAddress string, keyType, keyBits int) error {
	if keyType == ci.RSA && keyBits > 4096 {
		return errors.New("for RSA keys larger than 4096 contact your temporal admin")
	}
	//  create a background context for which the host will run in
	pk, _, err := ci.GenerateKeyPair(keyType, keyBits)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	host, err := libp2p.New(ctx, libp2p.Identity(pk), libp2p.ListenAddrStrings(listenAddress))
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", host)
	return nil
}
