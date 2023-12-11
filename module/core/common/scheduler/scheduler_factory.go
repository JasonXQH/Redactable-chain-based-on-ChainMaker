/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	"fmt"
	"regexp"
	"sync"

	"chainmaker.org/chainmaker-go/module/accesscontrol"
	"chainmaker.org/chainmaker/pb-go/v2/config"

	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/common/v2/monitor"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/protocol/v2"
)

type TxSchedulerFactory struct {
}

// NewTxScheduler building a transaction scheduler
func (sf TxSchedulerFactory) NewTxScheduler(vmMgr protocol.VmManager, chainConf protocol.ChainConf,
	storeHelper conf.StoreHelper, ledgerCache protocol.LedgerCache) protocol.TxScheduler {
	if chainConf.ChainConfig().Scheduler != nil && chainConf.ChainConfig().Scheduler.EnableEvidence {
		return newTxSchedulerEvidence(vmMgr, chainConf, storeHelper, ledgerCache)
	}
	return newTxScheduler(vmMgr, chainConf, storeHelper, ledgerCache)
}

// newTxScheduler building a regular transaction scheduler
func newTxScheduler(vmMgr protocol.VmManager, chainConf protocol.ChainConf,
	storeHelper conf.StoreHelper, cache protocol.LedgerCache) *TxScheduler {
	log := logger.GetLoggerByChain(logger.MODULE_CORE, chainConf.ChainConfig().ChainId)
	log.Debugf("use the common TxScheduler.")
	var txScheduler = &TxScheduler{
		lock:            sync.Mutex{},
		VmManager:       vmMgr,
		scheduleFinishC: make(chan bool),
		log:             log,
		chainConf:       chainConf,
		StoreHelper:     storeHelper,
		ledgerCache:     cache,
		contractCache:   &sync.Map{},
	}
	var err error
	txScheduler.keyReg, err = regexp.Compile(protocol.DefaultStateRegex)
	if err != nil {
		log.Fatalf("compile default state regex error %v", err)
	}
	txScheduler.signer, err = initSigner(chainConf.ChainConfig(), localconf.ChainMakerConfig, log)
	if err != nil {
		log.Fatalf("init signer of TxScheduler failed: err = %v", err)
	}

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		txScheduler.metricVMRunTime = monitor.NewHistogramVec(monitor.SUBSYSTEM_CORE_PROPOSER_SCHEDULER, "metric_vm_run_time",
			"VM run time metric", []float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 2, 5, 10}, "chainId")
	}
	return txScheduler
}

// init a signer with node private key
func initSigner(
	chainConfig *config.ChainConfig,
	cmConfig *localconf.CMConfig,
	log protocol.Logger) (protocol.SigningMember, error) {
	var err error
	var signingMember protocol.SigningMember
	nodeConfig := cmConfig.NodeConfig

	switch chainConfig.AuthType {
	case protocol.PermissionedWithCert, protocol.Identity:
		signingMember, err = accesscontrol.InitCertSigningMember(
			chainConfig,
			nodeConfig.OrgId,
			nodeConfig.PrivKeyFile,
			nodeConfig.PrivKeyPassword,
			nodeConfig.CertFile)
		if err != nil {
			return nil, fmt.Errorf("InitCertSigningMember failed: err = %v", err)
		}
	case protocol.PermissionedWithKey, protocol.Public:
		signingMember, err = accesscontrol.InitPKSigningMember(
			chainConfig.Crypto.Hash,
			nodeConfig.OrgId,
			nodeConfig.PrivKeyFile,
			nodeConfig.PrivKeyPassword)
		if err != nil {
			return nil, fmt.Errorf("InitPKSigningMember failed: err = %v", err)
		}
	default:
		return nil, fmt.Errorf("unknown auth type: %v", chainConfig.AuthType)
	}

	return signingMember, nil
}

// newTxSchedulerEvidence building a evidence transaction scheduler
func newTxSchedulerEvidence(vmMgr protocol.VmManager, chainConf protocol.ChainConf,
	storeHelper conf.StoreHelper, cache protocol.LedgerCache) *TxSchedulerEvidence {
	log := logger.GetLoggerByChain(logger.MODULE_CORE, chainConf.ChainConfig().ChainId)
	log.Debugf("use the evidence TxScheduler.")
	txSchedulerEvidence := &TxSchedulerEvidence{
		delegate: &TxScheduler{
			lock:            sync.Mutex{},
			VmManager:       vmMgr,
			scheduleFinishC: make(chan bool),
			log:             log,
			chainConf:       chainConf,
			StoreHelper:     storeHelper,
			ledgerCache:     cache,
			contractCache:   &sync.Map{},
		},
	}
	var err error
	txSchedulerEvidence.delegate.keyReg, err = regexp.Compile(protocol.DefaultStateRegex)
	if err != nil {
		log.Fatalf("compile default state regex error %v", err)
	}
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		txSchedulerEvidence.delegate.metricVMRunTime = monitor.NewHistogramVec(
			monitor.SUBSYSTEM_CORE_PROPOSER_SCHEDULER,
			"metric_vm_run_time",
			"VM run time metric",
			[]float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 2, 5, 10},
			"chainId",
		)
	}
	return txSchedulerEvidence
}
