package net

import (
	"testing"

	"chainmaker.org/chainmaker/protocol/v2"
)

func TestLiquidConfig(t *testing.T) {

	var netFactory NetFactory
	netType := protocol.Liquid
	net, err := netFactory.NewNet(
		netType,
		WithStunClient(
			"/ip4/127.0.0.1/tcp/11351",
			"/ip4/127.0.0.1/tcp/11352",
			"udp",
			true),
		WithStunServer(
			true,
			true,
			"/ip4/127.0.0.1/tcp/11352",
			"127.0.0.1/notify",
			"127.0.0.1/notify",
			"/ip4/127.0.0.1/tcp/10001",
			"/ip4/127.0.0.1/tcp/10002",
			"/ip4/127.0.0.1/tcp/10003",
			"/ip4/127.0.0.1/tcp/10004",
			"udp"),
		WithHolePunch(true),
	)
	if err != nil {
		t.Error()
	}
	if net == nil {
		t.Error()
	}

}
