/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"encoding/hex"
	"fmt"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"github.com/gogo/protobuf/proto"
)

var _ msgbus.Subscriber = (*pkACProvider)(nil)

// OnMessage contract event data is a []string, hexToString(proto.Marshal(data))
func (p *pkACProvider) OnMessage(msg *msgbus.Message) {
	switch msg.Topic {
	case msgbus.ChainConfig:
		p.log.Infof("[AC_PK] receive msg, topic: %s", msg.Topic.String())
		p.onMessageChainConfig(msg)
	}

}

func (p *pkACProvider) OnQuit() {

}

// onMessageChainConfig used to handle chain conf message
func (p *pkACProvider) onMessageChainConfig(msg *msgbus.Message) {
	dataStr, _ := msg.Payload.([]string)
	dataBytes, err := hex.DecodeString(dataStr[0])
	if err != nil {
		p.log.Error(err)
		return
	}
	chainConfig := &config.ChainConfig{}
	_ = proto.Unmarshal(dataBytes, chainConfig)

	p.hashType = chainConfig.GetCrypto().GetHash()
	err = p.initAdminMembers(chainConfig.TrustRoots)
	if err != nil {
		err = fmt.Errorf("new public AC provider failed: %s", err.Error())
		p.log.Error(err)
	}

	err = p.initConsensusMember(chainConfig)
	if err != nil {
		err = fmt.Errorf("new public AC provider failed: %s", err.Error())
		p.log.Error(err)
	}
	p.memberCache.Clear()

}
