/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"errors"
	"testing"

	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/protocol/v2"
)

func TestBlockchain_startNetService(t *testing.T) {
	type fields struct {
		log          *logger.CMLogger
		chainId      string
		netService   protocol.NetService
		initModules  map[string]struct{}
		startModules map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "chain1",
				netService: func() protocol.NetService {
					netService := newMockNetService(t)
					netService.EXPECT().Start().AnyTimes()
					return netService
				}(),
				initModules: nil,
				startModules: map[string]struct{}{
					moduleNameNetService: {},
				},
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				chainId: "chain1",
				netService: func() protocol.NetService {
					netService := newMockNetService(t)
					netService.EXPECT().Start().Return(errors.New("test error")).AnyTimes()
					return netService
				}(),
				initModules:  nil,
				startModules: map[string]struct{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:          tt.fields.log,
				chainId:      tt.fields.chainId,
				netService:   tt.fields.netService,
				initModules:  tt.fields.initModules,
				startModules: tt.fields.startModules,
			}
			if err := bc.startNetService(); (err != nil) != tt.wantErr {
				t.Errorf("startNetService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_startConsensus(t *testing.T) {
	type fields struct {
		log          *logger.CMLogger
		chainId      string
		consensus    protocol.ConsensusEngine
		initModules  map[string]struct{}
		startModules map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "chain1",
				consensus: func() protocol.ConsensusEngine {
					return nil
				}(),
				initModules:  nil,
				startModules: nil,
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				chainId: "chain1",
				consensus: func() protocol.ConsensusEngine {
					consensus := newMockConsensusEngine(t)
					consensus.EXPECT().Start().Return(errors.New("test error")).AnyTimes()
					return consensus
				}(),
				initModules:  nil,
				startModules: nil,
			},
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				log:     log,
				chainId: "chain1",
				consensus: func() protocol.ConsensusEngine {
					consensus := newMockConsensusEngine(t)
					consensus.EXPECT().Start().AnyTimes()
					return consensus
				}(),
				initModules: nil,
				startModules: map[string]struct{}{
					moduleNameConsensus: {},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:          tt.fields.log,
				chainId:      tt.fields.chainId,
				consensus:    tt.fields.consensus,
				initModules:  tt.fields.initModules,
				startModules: tt.fields.startModules,
			}
			if err := bc.startConsensus(); (err != nil) != tt.wantErr {
				t.Errorf("startConsensus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStartCoreEngine(t *testing.T) {
	t.Log("TestStartCoreEngine")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.startCoreEngine()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestBlockchain_startSyncService(t *testing.T) {
	type fields struct {
		log          *logger.CMLogger
		chainId      string
		syncServer   protocol.SyncService
		initModules  map[string]struct{}
		startModules map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "chain1",
				syncServer: func() protocol.SyncService {
					syncService := newMockSyncService(t)
					syncService.EXPECT().Start().AnyTimes()
					return syncService
				}(),
				initModules: nil,
				startModules: map[string]struct{}{
					moduleNameSync: {},
				},
			},
			wantErr: false,
		},
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "chain1",
				syncServer: func() protocol.SyncService {
					syncService := newMockSyncService(t)
					syncService.EXPECT().Start().Return(errors.New("test error")).AnyTimes()
					return syncService
				}(),
				initModules: nil,
				startModules: map[string]struct{}{
					moduleNameSync: {},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:          tt.fields.log,
				chainId:      tt.fields.chainId,
				syncServer:   tt.fields.syncServer,
				initModules:  tt.fields.initModules,
				startModules: tt.fields.startModules,
			}
			if err := bc.startSyncService(); (err != nil) != tt.wantErr {
				t.Errorf("startSyncService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_startTxPool(t *testing.T) {
	type fields struct {
		log     *logger.CMLogger
		chainId string
		txPool  protocol.TxPool
		//initModules  map[string]struct{}
		startModules map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     nil,
				chainId: "",
				txPool: func() protocol.TxPool {
					txPool := newMockTxPool(t)
					txPool.EXPECT().Start().AnyTimes()
					return txPool
				}(),
				startModules: map[string]struct{}{},
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				chainId: "",
				txPool: func() protocol.TxPool {
					txPool := newMockTxPool(t)
					txPool.EXPECT().Start().Return(errors.New("test error")).AnyTimes()
					return txPool
				}(),
				startModules: map[string]struct{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:          tt.fields.log,
				chainId:      tt.fields.chainId,
				txPool:       tt.fields.txPool,
				startModules: tt.fields.startModules,
			}
			if err := bc.startTxPool(); (err != nil) != tt.wantErr {
				t.Errorf("startTxPool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_startVM(t *testing.T) {
	type fields struct {
		log          *logger.CMLogger
		chainId      string
		vmMgr        protocol.VmManager
		initModules  map[string]struct{}
		startModules map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "chain1",
				vmMgr: func() protocol.VmManager {
					vmManager := newMockVmManager(t)
					vmManager.EXPECT().Start().AnyTimes()
					return vmManager
				}(),
				startModules: map[string]struct{}{},
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				chainId: "chain1",
				vmMgr: func() protocol.VmManager {
					vmManager := newMockVmManager(t)
					vmManager.EXPECT().Start().Return(errors.New("test error")).AnyTimes()
					return vmManager
				}(),
				startModules: map[string]struct{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:          tt.fields.log,
				chainId:      tt.fields.chainId,
				vmMgr:        tt.fields.vmMgr,
				initModules:  tt.fields.initModules,
				startModules: tt.fields.startModules,
			}
			if err := bc.startVM(); (err != nil) != tt.wantErr {
				t.Errorf("startVM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsModuleStartUp(t *testing.T) {
	t.Log("TestIsModuleStartUp")

	blockchainList := createBlockChain(t)

	moduleNameSlice := []string{
		moduleNameSubscriber,
		moduleNameStore,
		moduleNameLedger,
		moduleNameChainConf,
		moduleNameAccessControl,
		moduleNameNetService,
		moduleNameVM,
		moduleNameTxPool,
		moduleNameCore,
		moduleNameConsensus,
		moduleNameSync,
	}
	for _, moduleName := range moduleNameSlice {
		res := blockchainList[0].isModuleStartUp(moduleName)
		t.Log(res)
	}
}
