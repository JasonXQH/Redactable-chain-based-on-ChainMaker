/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txpool

import (
	normal "chainmaker.org/chainmaker/txpool-normal/v2"
	"strings"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/protocol/v2" // nolint: typecheck
)

const (
	// TypeDefault NORMAL
	TypeDefault = normal.TxPoolType
)

// Provider provide protocol.TxPool implement
type Provider func(
	nodeId string,
	chainId string,
	txFilter protocol.TxFilter,
	chainStore protocol.BlockchainStore,
	msgBus msgbus.MessageBus,
	chainConf protocol.ChainConf,
	singer protocol.SigningMember,
	ac protocol.AccessControlProvider,
	netService protocol.NetService,
	log protocol.Logger,
	monitorEnabled bool,
	poolConfig map[string]interface{}) (protocol.TxPool, error)

var txPoolProviders = make(map[string]Provider)

// RegisterTxPoolProvider register tx pool provider
// @param t
// @param f
func RegisterTxPoolProvider(t string, f Provider) {
	txPoolProviders[strings.ToUpper(t)] = f
}

// GetTxPoolProvider get provider by name
// @param t
// @return Provider
func GetTxPoolProvider(t string) Provider {
	provider, ok := txPoolProviders[strings.ToUpper(t)]
	if !ok {
		return nil
	}
	return provider
}
