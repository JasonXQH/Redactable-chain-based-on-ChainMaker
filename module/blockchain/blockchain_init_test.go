/*
   Copyright (C) BABEC. All rights reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"errors"
	"testing"

	"chainmaker.org/chainmaker-go/module/subscriber"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	msgbusMock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/logger/v2"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
)

func TestBlockchain_initCache(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}

	var (
		ctrl            = gomock.NewController(t)
		chainConf       = mock.NewMockChainConf(ctrl)
		ac              = mock.NewMockAccessControlProvider(ctrl)
		syncService     = mock.NewMockSyncService(ctrl)
		coreEngine      = mock.NewMockCoreEngine(ctrl)
		txpool          = mock.NewMockTxPool(ctrl)
		snapshotManager = mock.NewMockSnapshotManager(ctrl)
		singMemer       = mock.NewMockSigningMember(ctrl)
		ledgerCache     = mock.NewMockLedgerCache(ctrl)
		vmMgr           = mock.NewMockVmManager(ctrl)
		proposalCache   = mock.NewMockProposalCache(ctrl)
		net             = mock.NewMockNet(ctrl)
		netService      = mock.NewMockNetService(ctrl)
		consensus       = mock.NewMockConsensusEngine(ctrl)
		msgBus          = msgbusMock.NewMockMessageBus(ctrl)
	)

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:        log,
				genesis:    "chain1",
				chainId:    "chain1",
				msgBus:     msgBus,
				net:        net,
				netService: netService,
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					return blockChainStore
				}(),
				oldStore:        nil,
				consensus:       consensus,
				txPool:          txpool,
				coreEngine:      coreEngine,
				vmMgr:           vmMgr,
				identity:        singMemer,
				ac:              ac,
				syncServer:      syncService,
				ledgerCache:     ledgerCache,
				proposalCache:   proposalCache,
				snapshotManager: snapshotManager,
				chainConf:       chainConf,
				initModules: map[string]struct{}{
					moduleNameLedger: {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:        log,
				genesis:    "chain1",
				chainId:    "chain1",
				msgBus:     msgBus,
				net:        net,
				netService: netService,
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					block := createNewBlock(uint64(1), int64(1), "chain1")
					blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, nil)
					return blockChainStore
				}(),
				oldStore:        nil,
				consensus:       consensus,
				txPool:          txpool,
				coreEngine:      coreEngine,
				vmMgr:           vmMgr,
				identity:        singMemer,
				ac:              ac,
				syncServer:      syncService,
				ledgerCache:     ledgerCache,
				proposalCache:   proposalCache,
				snapshotManager: snapshotManager,
				chainConf:       chainConf,
				initModules:     map[string]struct{}{},
				startModules:    nil,
			},
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				log:        log,
				genesis:    "chain1",
				chainId:    "chain1",
				msgBus:     msgBus,
				net:        net,
				netService: netService,
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					block := createNewBlock(uint64(1), int64(1), "chain1")
					blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, errors.New("blockChainStore err test"))

					return blockChainStore
				}(),
				oldStore:        nil,
				consensus:       consensus,
				txPool:          txpool,
				coreEngine:      coreEngine,
				vmMgr:           vmMgr,
				identity:        singMemer,
				ac:              ac,
				syncServer:      syncService,
				ledgerCache:     ledgerCache,
				proposalCache:   proposalCache,
				snapshotManager: snapshotManager,
				chainConf: func() protocol.ChainConf {
					chainConf = mock.NewMockChainConf(ctrl)
					chainConfig.AuthType = protocol.Identity
					chainConfig.Snapshot = &configpb.SnapshotConfig{
						EnableEvidence: true,
					}
					chainConf.EXPECT().Init().AnyTimes().Return(nil)
					chainConf.EXPECT().GetChainConfigFromFuture(gomock.Any()).AnyTimes().Return(nil, nil)
					chainConf.EXPECT().GetChainConfigAt(gomock.Any()).AnyTimes().Return(nil, nil)
					chainConf.EXPECT().GetConsensusNodeIdList().AnyTimes().Return(nil, nil)
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()

					return chainConf
				}(),
				initModules:  map[string]struct{}{},
				startModules: nil,
			},
			wantErr: false,
		},
		{
			name: "test3",
			fields: fields{
				log:        log,
				genesis:    "chain1",
				chainId:    "chain1",
				msgBus:     msgBus,
				net:        net,
				netService: netService,
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					block := createNewBlock(uint64(1), int64(1), "chain1")
					blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, errors.New("blockChainStore GetDbTransaction test err msg"))
					blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, errors.New("blockChainStore err test"))
					return blockChainStore
				}(),
				oldStore:        nil,
				consensus:       consensus,
				txPool:          txpool,
				coreEngine:      coreEngine,
				vmMgr:           vmMgr,
				identity:        singMemer,
				ac:              ac,
				syncServer:      syncService,
				ledgerCache:     ledgerCache,
				proposalCache:   proposalCache,
				snapshotManager: snapshotManager,
				chainConf:       chainConf,
				initModules:     map[string]struct{}{},
				startModules:    nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initCache(); (err != nil) != tt.wantErr {
				t.Errorf("initCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_InitForRebuildDbs(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	localconf.ChainMakerConfig = &localconf.CMConfig{
		StorageConfig: map[string]interface{}{
			"store_path": "./mypath1",
			"blockdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"statedb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"historydb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"resultdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"txexistdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"disable_contract_eventdb": true,
			"contract_eventdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
		},
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
				genesis: "",
				chainId: "chain1",
				msgBus: func() msgbus.MessageBus {
					msgbus := newMockMessageBus(t)
					msgbus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
					return msgbus
				}(),
				net:             nil,
				netService:      nil,
				store:           nil,
				oldStore:        nil,
				consensus:       nil,
				txPool:          nil,
				coreEngine:      nil,
				vmMgr:           nil,
				identity:        nil,
				ac:              newMockAccessControlProvider(t),
				syncServer:      nil,
				ledgerCache:     nil,
				proposalCache:   nil,
				snapshotManager: nil,
				lastBlock:       nil,
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConf.EXPECT().AddWatch(gomock.Any()).AnyTimes()
					chainConf.EXPECT().AddVmWatch(gomock.Any()).AnyTimes()
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				chainNodeList:   nil,
				eventSubscriber: nil,
				initModules: map[string]struct{}{
					moduleNameSubscriber:    {},
					moduleNameStore:         {},
					moduleNameLedger:        {},
					moduleNameChainConf:     {},
					moduleNameAccessControl: {},
					moduleNameVM:            {},
					moduleNameTxPool:        {},
					moduleNameCore:          {},
					moduleNameConsensus:     {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
	}

	// init sync service module
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.InitForRebuildDbs(); (err != nil) != tt.wantErr {
				t.Errorf("InitForRebuildDbs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initBaseModules(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	type args struct {
		baseModules []map[string]func() error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				genesis: "",
				chainId: "chain1",
				msgBus: func() msgbus.MessageBus {
					msgBus := newMockMessageBus(t)
					msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
					return msgBus
				}(),
				initModules: map[string]struct{}{
					moduleNameSubscriber: {},
				},
				startModules: nil,
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			tt.args.baseModules = []map[string]func() error{
				{moduleNameSubscriber: bc.initSubscriber},
			}
			if err := bc.initBaseModules(tt.args.baseModules); (err != nil) != tt.wantErr {
				t.Errorf("initBaseModules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initExtModules(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	type args struct {
		extModules []map[string]func() error
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				genesis: "",
				chainId: "chin1",
				msgBus: func() msgbus.MessageBus {
					msgBus := newMockMessageBus(t)
					msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
					return msgBus
				}(),
				initModules: map[string]struct{}{
					moduleNameSubscriber: {},
				},
				startModules: nil,
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			tt.args.extModules = []map[string]func() error{
				{moduleNameSubscriber: bc.initSubscriber},
			}
			if err := bc.initExtModules(tt.args.extModules); (err != nil) != tt.wantErr {
				t.Errorf("initExtModules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_Init(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}

	localconf.ChainMakerConfig = &localconf.CMConfig{
		StorageConfig: map[string]interface{}{
			"store_path": "./mypath1",
			"blockdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"statedb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"historydb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"resultdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"txexistdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"disable_contract_eventdb": true,
			"contract_eventdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
		},
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
				genesis: "chain1",
				chainId: "chain1",
				msgBus: func() msgbus.MessageBus {
					msgbus := newMockMessageBus(t)
					msgbus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
					return msgbus
				}(),
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConf.EXPECT().AddWatch(gomock.Any()).AnyTimes()
					chainConf.EXPECT().AddVmWatch(gomock.Any()).AnyTimes()
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				chainNodeList:   nil,
				eventSubscriber: nil,
				initModules: map[string]struct{}{
					moduleNameSubscriber:    {},
					moduleNameLedger:        {},
					moduleNameChainConf:     {},
					moduleNameVM:            {},
					moduleNameTxPool:        {},
					moduleNameAccessControl: {},
					moduleNameNetService:    {},
					moduleNameCore:          {},
					moduleNameConsensus:     {},
					moduleNameSync:          {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initNetService(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:          log,
				genesis:      "",
				chainId:      "chain1",
				initModules:  map[string]struct{}{},
				startModules: nil,
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				genesis: "",
				chainId: "chain1",
				initModules: map[string]struct{}{
					moduleNameNetService: {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initNetService(); (err != nil) != tt.wantErr {
				t.Errorf("initNetService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initStore(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}

	localconf.ChainMakerConfig = &localconf.CMConfig{
		StorageConfig: map[string]interface{}{
			"store_path": "./mypath1",
			"blockdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"statedb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"historydb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"resultdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"txexistdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
			"disable_contract_eventdb": true,
			"contract_eventdb_config": map[string]interface{}{
				"provider": "leveldb",
				"leveldb_config": map[string]interface{}{
					"store_path": "./mypath1",
				},
			},
		},
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
				genesis: "",
				chainId: "chain1",
				store: func() protocol.BlockchainStore {
					store := newMockBlockchainStore(t)
					store.EXPECT().GetDBHandle(gomock.Any()).AnyTimes()
					store.EXPECT().GetContractByName(gomock.Any()).AnyTimes()
					return store
				}(),
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				initModules: map[string]struct{}{
					moduleNameStore: {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				genesis: "",
				chainId: "chain1",
				store: func() protocol.BlockchainStore {
					store := newMockBlockchainStore(t)
					return store
				}(),
				oldStore: func() protocol.BlockchainStore {
					store := newMockBlockchainStore(t)
					store.EXPECT().GetDBHandle(gomock.Any()).AnyTimes()
					store.EXPECT().GetContractByName(gomock.Any()).AnyTimes()
					return store
				}(),
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				initModules: map[string]struct{}{
					moduleNameStore: {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initStore(); (err != nil) != tt.wantErr {
				t.Errorf("initStore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initChainConf(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}

	var (
		ctrl      = gomock.NewController(t)
		chainConf = mock.NewMockChainConf(ctrl)
		//log = mock.NewMockLogger(ctrl)
		//ac        = mock.NewMockAccessControlProvider(ctrl)
		//syncService = mock.NewMockSyncService(ctrl)
		//coreEngine  = mock.NewMockCoreEngine(ctrl)
		//txpool          = mock.NewMockTxPool(ctrl)
		//snapshotManager = mock.NewMockSnapshotManager(ctrl)
		//singMemer       = mock.NewMockSigningMember(ctrl)
		//ledgerCache     = mock.NewMockLedgerCache(ctrl)
		//vmMgr           = mock.NewMockVmManager(ctrl)
		//proposalCache   = mock.NewMockProposalCache(ctrl)
		//net             = mock.NewMockNet(ctrl)
		//netService      = mock.NewMockNetService(ctrl)
		//consensus       = mock.NewMockConsensusEngine(ctrl)
		//msgBus          = msgbusMock.NewMockMessageBus(ctrl)
	)

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:     log,
				chainId: "",
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					block := createNewBlock(uint64(1), int64(1), "chain1")
					blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, errors.New("blockChainStore GetDbTransaction test err msg"))
					blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, errors.New("blockChainStore err test"))
					blockChainStore.EXPECT().ReadObject(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
					return blockChainStore
				}(),
				chainConf: chainConf,
			},
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				log:        log,
				genesis:    "",
				chainId:    "chain1",
				msgBus:     newMockMessageBus(t),
				net:        nil,
				netService: nil,
				store: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					block := createNewBlock(uint64(1), int64(1), "chain1")
					blockChainStore.EXPECT().GetDbTransaction(gomock.Any()).AnyTimes().Return(nil, errors.New("blockChainStore GetDbTransaction test err msg"))
					blockChainStore.EXPECT().GetLastBlock().AnyTimes().Return(block, errors.New("blockChainStore err test"))
					blockChainStore.EXPECT().ReadObject(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
					return blockChainStore
				}(),
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						AuthType: protocol.Identity,
						Consensus: &configpb.ConsensusConfig{
							Type: consensus.ConsensusType_DPOS,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initChainConf(); (err != nil) != tt.wantErr {
				t.Errorf("initChainConf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initAC(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log: log,
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						AuthType: protocol.Identity,
						Consensus: &configpb.ConsensusConfig{
							Type: consensus.ConsensusType_DPOS,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				log: log,
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						AuthType: protocol.PermissionedWithKey,
						Consensus: &configpb.ConsensusConfig{
							Type: consensus.ConsensusType_DPOS,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				log: log,
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						AuthType: protocol.Public,
						Consensus: &configpb.ConsensusConfig{
							Type: consensus.ConsensusType_RAFT,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			wantErr: true,
		},
		{
			name: "test3",
			fields: fields{
				log: log,
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						AuthType: protocol.Public,
						Consensus: &configpb.ConsensusConfig{
							Type: consensus.ConsensusType_TBFT,
							Nodes: []*configpb.OrgConfig{
								{
									OrgId:  "org1",
									NodeId: []string{"chain1"},
								},
							},
						},
						Crypto: &configpb.CryptoConfig{
							Hash: "SHA256",
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				store:  newMockBlockchainStore(t),
				msgBus: newMockMessageBus(t),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initAC(); (err != nil) != tt.wantErr {
				t.Errorf("initAC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initTxPool(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				log:          log,
				chainId:      "chain1",
				initModules:  nil,
				startModules: nil,
			},
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				log:     log,
				genesis: "",
				chainId: "chain1",
				initModules: map[string]struct{}{
					moduleNameTxPool: {},
				},
				startModules: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initTxPool(); (err != nil) != tt.wantErr {
				t.Errorf("initTxPool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initVM(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	msgBus := msgbusMock.NewMockMessageBus(c)
	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).Return().AnyTimes()
	iter := mock.NewMockStateIterator(c)
	iter.EXPECT().Release().Return().AnyTimes()
	iter.EXPECT().Next().Return(false).AnyTimes()
	store := mock.NewMockBlockchainStore(c)
	store.EXPECT().SelectObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(iter, nil).AnyTimes()

	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					ctrl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig = &configpb.ChainConfig{
						Vm: &configpb.Vm{},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				initModules: map[string]struct{}{},
				msgBus:      msgBus,
				store:       store,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initVM(); (err != nil) != tt.wantErr {
				t.Errorf("initVM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetChainNodesInfo(t *testing.T) {
	t.Log("TestGetChainNodesInfo")

	for i := 0; i < 6; i++ {
		chainNoInfo, err := (&soloChainNodesInfoProvider{}).GetChainNodesInfo()

		if err != nil {
			t.Log(err)
		}

		t.Log(chainNoInfo)
	}

}

func TestBlockchain_initCore(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
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
				msgBus: func() msgbus.MessageBus {
					msgbus := newMockMessageBus(t)
					msgbus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
					return msgbus
				}(),
				identity: func() protocol.SigningMember {
					signMember := newMockSigningMember(t)
					signMember.EXPECT().GetMember().AnyTimes()
					return signMember
				}(),
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConfig = &configpb.ChainConfig{
						ChainId: "chain1",
						Consensus: &configpb.ConsensusConfig{
							Type: consensusType,
							Nodes: []*configpb.OrgConfig{
								{
									OrgId:  org1Id,
									NodeId: []string{id},
								},
							},
						},
						Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
						Contract: &configpb.ContractConfig{
							EnableSqlSupport: true,
						},
						Block: &configpb.BlockConfig{
							BlockInterval: 5,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					chainConf.EXPECT().AddWatch(gomock.Any()).Return().AnyTimes()
					return chainConf
				}(),
				initModules: map[string]struct{}{
					moduleNameCore: {},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initCore(); (err != nil) != tt.wantErr {
				t.Errorf("initCore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchain_initConsensus(t *testing.T) {
	type fields struct {
		log             *logger.CMLogger
		genesis         string
		chainId         string
		msgBus          msgbus.MessageBus
		net             protocol.Net
		netService      protocol.NetService
		store           protocol.BlockchainStore
		oldStore        protocol.BlockchainStore
		consensus       protocol.ConsensusEngine
		txPool          protocol.TxPool
		coreEngine      protocol.CoreEngine
		vmMgr           protocol.VmManager
		identity        protocol.SigningMember
		ac              protocol.AccessControlProvider
		syncServer      protocol.SyncService
		ledgerCache     protocol.LedgerCache
		proposalCache   protocol.ProposalCache
		snapshotManager protocol.SnapshotManager
		lastBlock       *commonPb.Block
		chainConf       protocol.ChainConf
		chainNodeList   []string
		eventSubscriber *subscriber.EventSubscriber
		initModules     map[string]struct{}
		startModules    map[string]struct{}
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
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConfig = &configpb.ChainConfig{
						ChainId: "chain1",
						Consensus: &configpb.ConsensusConfig{
							Type: consensusType,
							Nodes: []*configpb.OrgConfig{
								{
									OrgId:  org1Id,
									NodeId: []string{id},
								},
							},
						},
						Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
						Contract: &configpb.ContractConfig{
							EnableSqlSupport: true,
						},
						Block: &configpb.BlockConfig{
							BlockInterval: 5,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					chainConf.EXPECT().AddWatch(gomock.Any()).Return().AnyTimes()
					return chainConf
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				log:             tt.fields.log,
				genesis:         tt.fields.genesis,
				chainId:         tt.fields.chainId,
				msgBus:          tt.fields.msgBus,
				net:             tt.fields.net,
				netService:      tt.fields.netService,
				store:           tt.fields.store,
				oldStore:        tt.fields.oldStore,
				consensus:       tt.fields.consensus,
				txPool:          tt.fields.txPool,
				coreEngine:      tt.fields.coreEngine,
				vmMgr:           tt.fields.vmMgr,
				identity:        tt.fields.identity,
				ac:              tt.fields.ac,
				syncServer:      tt.fields.syncServer,
				ledgerCache:     tt.fields.ledgerCache,
				proposalCache:   tt.fields.proposalCache,
				snapshotManager: tt.fields.snapshotManager,
				lastBlock:       tt.fields.lastBlock,
				chainConf:       tt.fields.chainConf,
				chainNodeList:   tt.fields.chainNodeList,
				eventSubscriber: tt.fields.eventSubscriber,
				initModules:     tt.fields.initModules,
				startModules:    tt.fields.startModules,
			}
			if err := bc.initConsensus(); (err != nil) != tt.wantErr {
				t.Errorf("initConsensus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitSync(t *testing.T) {
	t.Log("TestInitSync")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initSync()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestInitSubscriber(t *testing.T) {
	t.Log("TestInitSubscriber")

	blockchainList := createBlockChain(t)
	for _, blockchain := range blockchainList {
		err := blockchain.initSubscriber()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestBlockchain_isModuleInit(t *testing.T) {
	type fields struct {
		initModules map[string]struct{}
	}
	type args struct {
		moduleName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				initModules: map[string]struct{}{},
			},
			args: args{
				moduleName: moduleNameAccessControl,
			},
			want: false,
		},
		{
			name: "test1",
			fields: fields{
				initModules: map[string]struct{}{
					moduleNameAccessControl: {},
				},
			},
			args: args{
				moduleName: moduleNameAccessControl,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &Blockchain{
				initModules: tt.fields.initModules,
			}
			if got := bc.isModuleInit(tt.args.moduleName); got != tt.want {
				t.Errorf("isModuleInit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newMockChainConf(t *testing.T) *mock.MockChainConf {
	ctrl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctrl)
	return chainConf
}

func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}

func newMockTxPool(t *testing.T) *mock.MockTxPool {
	ctrl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctrl)
	return txPool
}

func newMockSigningMember(t *testing.T) *mock.MockSigningMember {
	ctrl := gomock.NewController(t)
	signingMember := mock.NewMockSigningMember(ctrl)
	return signingMember
}

func newMockAccessControlProvider(t *testing.T) *mock.MockAccessControlProvider {
	ctrl := gomock.NewController(t)
	ac := mock.NewMockAccessControlProvider(ctrl)
	return ac
}

func newMockMessageBus(t *testing.T) *msgbusMock.MockMessageBus {
	ctrl := gomock.NewController(t)
	messageBus := msgbusMock.NewMockMessageBus(ctrl)
	return messageBus
}

func newMockVmManager(t *testing.T) *mock.MockVmManager {
	ctrl := gomock.NewController(t)
	vmManager := mock.NewMockVmManager(ctrl)
	return vmManager
}

func newMockSyncService(t *testing.T) *mock.MockSyncService {
	ctrl := gomock.NewController(t)
	syncService := mock.NewMockSyncService(ctrl)
	return syncService
}

func newMockConsensusEngine(t *testing.T) *mock.MockConsensusEngine {
	ctrl := gomock.NewController(t)
	consensus := mock.NewMockConsensusEngine(ctrl)
	return consensus
}

func newMockNetService(t *testing.T) *mock.MockNetService {
	ctrl := gomock.NewController(t)
	netService := mock.NewMockNetService(ctrl)
	return netService
}

func newMockNet(t *testing.T) *mock.MockNet {
	ctrl := gomock.NewController(t)
	net := mock.NewMockNet(ctrl)
	return net
}
