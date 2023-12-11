/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"encoding/hex"
	"fmt"

	"chainmaker.org/chainmaker/common/v2/crypto/asym"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"github.com/gogo/protobuf/proto"
)

var _ msgbus.Subscriber = (*permissionedPkACProvider)(nil)

// OnMessage contract event data is a []string, hexToString(proto.Marshal(data))
func (pp *permissionedPkACProvider) OnMessage(msg *msgbus.Message) {
	switch msg.Topic {
	case msgbus.ChainConfig:
		pp.acService.log.Infof("[AC_PWK] receive msg, topic: %s", msg.Topic.String())
		pp.onMessageChainConfig(msg)
	case msgbus.PubkeyManageDelete:
		pp.acService.log.Infof("[AC_PWK] receive msg, topic: %s", msg.Topic.String())
		pp.onMessagePublishKeyManageDelete(msg)
	}

}

func (pp *permissionedPkACProvider) OnQuit() {

}

func (pp *permissionedPkACProvider) onMessageChainConfig(msg *msgbus.Message) {
	dataStr, _ := msg.Payload.([]string)
	dataBytes, err := hex.DecodeString(dataStr[0])
	if err != nil {
		pp.acService.log.Error(err)
		return
	}
	chainConfig := &config.ChainConfig{}
	_ = proto.Unmarshal(dataBytes, chainConfig)

	pp.acService.hashType = chainConfig.GetCrypto().GetHash()

	err = pp.initAdminMembers(chainConfig.TrustRoots)
	if err != nil {
		err = fmt.Errorf("update chainconfig error: %s", err.Error())
		pp.acService.log.Error(err)
	}

	err = pp.initConsensusMember(chainConfig.Consensus.Nodes)
	if err != nil {
		err = fmt.Errorf("update chainconfig error: %s", err.Error())
		pp.acService.log.Error(err)
	}

	pp.acService.initResourcePolicy(chainConfig.ResourcePolicies, pp.localOrg)

	pp.acService.memberCache.Clear()
}

func (pp *permissionedPkACProvider) onMessagePublishKeyManageDelete(msg *msgbus.Message) {
	data, _ := msg.Payload.([]string)
	publishKey := data[1]

	pk, err := asym.PublicKeyFromPEM([]byte(publishKey))
	if err != nil {
		err = fmt.Errorf("delete member cache failed, [%v]", err.Error())
		pp.acService.log.Error(err)
	}
	pkStr, err := pk.String()
	if err != nil {
		err = fmt.Errorf("delete member cache failed, [%v]", err.Error())
		pp.acService.log.Error(err)
	}
	pp.acService.memberCache.Remove(pkStr)
	pp.acService.log.Debugf("The public key was removed from the cache,[%v]", pkStr)
}
