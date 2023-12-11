/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0

This file is for version compatibility
*/

package net

import (
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
)

const (
	moduleSync = "NetService"
)

// ConfigWatcher return a implementation of protocol.Watcher. It is used for refreshing the config.
func (ns *NetService) ConfigWatcher() protocol.Watcher {
	if ns.configWatcher == nil {
		ns.configWatcher = &ConfigWatcher{ns: ns}
	}
	return ns.configWatcher
}

var _ protocol.Watcher = (*ConfigWatcher)(nil)

// ConfigWatcher is a implementation of protocol.Watcher.
type ConfigWatcher struct {
	ns *NetService
}

// Module
func (cw *ConfigWatcher) Module() string {
	return moduleSync
}

// Watch
func (cw *ConfigWatcher) Watch(chainConfig *configPb.ChainConfig) error {
	// refresh chainConfig
	cw.ns.logger.Infof("[NetService] refreshing chain config...")
	// 1.refresh consensus nodeIds
	// 1.1 get all new nodeIds
	newConsensusNodeIds := make(map[string]struct{})
	for _, node := range chainConfig.Consensus.Nodes {
		for _, nodeId := range node.NodeId {
			newConsensusNodeIds[nodeId] = struct{}{}
		}
	}
	// 1.2 refresh consensus nodeIds
	cw.ns.consensusNodeIdsLock.Lock()
	cw.ns.consensusNodeIds = newConsensusNodeIds
	cw.ns.consensusNodeIdsLock.Unlock()
	cw.ns.logger.Infof("[NetService] refresh ids of consensus nodes ok ")
	// 2.re-verify peers
	cw.ns.localNet.ReVerifyPeers(cw.ns.chainId)
	cw.ns.logger.Infof("[NetService] re-verify peers ok")
	cw.ns.logger.Infof("[NetService] refresh chain config ok")
	return nil
}

// VmWatcher return an implementation of protocol.VmWatcher.
// It is used for refreshing revoked peer which use revoked tls cert.
func (ns *NetService) VmWatcher() protocol.VmWatcher {
	if ns.vmWatcher == nil {
		ns.vmWatcher = &VmWatcher{ns: ns}
	}
	return ns.vmWatcher
}

var _ protocol.VmWatcher = (*VmWatcher)(nil)

// VmWatcher NetService callback certificate management
type VmWatcher struct {
	ns *NetService
}

func (v *VmWatcher) Module() string {
	return moduleSync
}

func (v *VmWatcher) ContractNames() []string {
	return []string{syscontract.SystemContract_CERT_MANAGE.String(),
		syscontract.SystemContract_PUBKEY_MANAGE.String()}
}

func (v *VmWatcher) Callback(contractName string, _ []byte) error {
	switch contractName {
	case syscontract.SystemContract_CERT_MANAGE.String():
		v.ns.logger.Infof("[NetService] callback msg, contract name: %s", contractName)
		v.ns.localNet.ReVerifyPeers(v.ns.chainId)
		return nil
	case syscontract.SystemContract_PUBKEY_MANAGE.String():
		v.ns.logger.Infof("[NetService] callback msg, contract name: %s", contractName)
		v.ns.localNet.ReVerifyPeers(v.ns.chainId)
		return nil
	default:
		return nil
	}
}
