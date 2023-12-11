/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"encoding/hex"

	"chainmaker.org/chainmaker/pb-go/v2/config"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"github.com/gogo/protobuf/proto"
)

var _ msgbus.Subscriber = (*Blockchain)(nil)

// OnMessage contract event data is a []string, hexToString(proto.Marshal(data))
func (bc *Blockchain) OnMessage(msg *msgbus.Message) {
	switch msg.Topic {
	case msgbus.ChainConfig:
		bc.log.Infof("[Blockchain] receive msg, topic: %s", msg.Topic.String())
		bc.updateChainConfig(msg)
		if err := bc.Init(); err != nil {
			bc.log.Errorf("blockchain init failed when the configuration of blockchain updating, %s", err)
			return
		}
		bc.StopOnRequirements()
		if err := bc.Start(); err != nil {
			bc.log.Errorf("blockchain start failed when the configuration of blockchain updating, %s", err)
			return
		}
	}
}

func (bc *Blockchain) OnQuit() {
	// nothing for implement interface msgbus.Subscriber
}

func (bc *Blockchain) updateChainConfig(msg *msgbus.Message) {
	cfg := &config.ChainConfig{}
	dataStr, ok := msg.Payload.([]string)
	if !ok {
		return
	}
	dataBytes, err := hex.DecodeString(dataStr[0])
	if err != nil {
		bc.log.Error(err)
		panic(err)
	}
	err = proto.Unmarshal(dataBytes, cfg)
	if err != nil {
		bc.log.Error(err)
		panic(err)
	}
	_ = bc.chainConf.SetChainConfig(cfg)
}
