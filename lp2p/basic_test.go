package lp2p

import "testing"

func TestInitBasicHost(t *testing.T) {
	err := InitBasicHost()
	if err != nil {
		t.Fatal(err)
	}
}
