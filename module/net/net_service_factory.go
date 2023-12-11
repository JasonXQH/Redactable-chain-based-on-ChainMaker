/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package net

import (
	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/protocol/v2"
)

// NetServiceFactory is a net service instance factory.
type NetServiceFactory struct {
}

// NewNetService create a new net service instance.
func (nsf *NetServiceFactory) NewNetService(
	net protocol.Net,
	chainId string,
	ac protocol.AccessControlProvider,
	chainConf protocol.ChainConf,
	opts ...NetServiceOption) (protocol.NetService, error) {
	//初始化工厂实例
	ns := NewNetService(chainId, net, ac)
	if err := ns.Apply(opts...); err != nil {
		return nil, err
	}
	if chainConf != nil {
		if err := nsf.setAllConsensusNodeIds(ns, chainConf); err != nil {
			return nil, err
		}
		// set contract event subscribe
		ns.msgBus.Register(msgbus.ChainConfig, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.CertManageCertsRevoke, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.CertManageCertsFreeze, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.CertManageCertsUnfreeze, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.CertManageCertsAliasUpdate, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.CertManageCertsAliasDelete, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.PubkeyManageAdd, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.PubkeyManageDelete, ns.NetConfigSubscribe())
		ns.msgBus.Register(msgbus.MaxbftEpochConf, ns.NetConfigSubscribe())

		// v220_compat Deprecated
		{
			// set config watcher
			chainConf.AddWatch(ns.ConfigWatcher()) //nolint: staticcheck
			// set vm watcher
			chainConf.AddVmWatch(ns.VmWatcher()) //nolint: staticcheck
		}
	}
	return ns, nil
}

func (nsf *NetServiceFactory) setAllConsensusNodeIds(ns *NetService, chainConf protocol.ChainConf) error {
	consensusNodeUidList := make([]string, 0)
	// add all the seeds
	for _, node := range chainConf.ChainConfig().Consensus.Nodes {
		consensusNodeUidList = append(consensusNodeUidList, node.NodeId...)
	}
	// set all consensus node id for net service
	err := ns.Apply(WithConsensusNodeUid(consensusNodeUidList...))
	if err != nil {
		return err
	}
	ns.logger.Infof("[NetServiceFactory] set consensus node uid list ok(chain-id:%s)", ns.chainId)
	return nil
}
