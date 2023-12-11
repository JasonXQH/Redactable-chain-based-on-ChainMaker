/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// syncmode means commit new block in sync way
package syncmode

import (
	"reflect"
	"testing"

	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker-go/module/core/common/scheduler"
	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker-go/module/core/syncmode/proposer"
	"chainmaker.org/chainmaker-go/module/core/syncmode/verifier"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	msgbusMock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
)

func TestNewCoreEngine(t *testing.T) {
	type args struct {
		cf *conf.CoreEngineConfig
	}

	var (
		chainId         = "123456"
		blockchainStore = newMockBlockchainStore(t)
		snapshotManager = newMockSnapshotManager(t)
		txPool          = newMockTxPool(t)
		ledgerCache     = newMockLedgerCache(t)
		proposedCache   = newMockProposalCache(t)
		chainConf       = newMockChainConf(t)
		msgBus          = msgbus.NewMessageBus()
		storeHelper     = newMockStoreHelper(t)
		log             = newMockLogger(t)
		signMember      = newMockSigningMember(t)
		ac              = newMockAccessControlProvider(t)
		vmMgr           = newMockVmManager(t)
	)

	chainConfig := &configpb.ChainConfig{
		Core: &configpb.CoreConfig{
			ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
				ConsensusMessageTurbo: true,
			},
		},
		Block: &configpb.BlockConfig{
			BlockInterval: 10,
		},
		Crypto: &configpb.CryptoConfig{
			Hash: "SHA256",
		},
		AuthType: protocol.Identity,
	}
	chainConf.EXPECT().AddWatch(gomock.Any()).AnyTimes().Return()
	chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
	signMember.EXPECT().GetMember().Return(nil, nil).AnyTimes()

	cf := &conf.CoreEngineConfig{
		ChainId:         chainId,
		TxPool:          txPool,
		SnapshotManager: snapshotManager,
		MsgBus:          msgBus,
		Identity:        signMember,
		LedgerCache:     ledgerCache,
		ProposalCache:   proposedCache,
		ChainConf:       chainConf,
		AC:              ac,
		BlockchainStore: blockchainStore,
		Log:             log,
		VmMgr:           vmMgr,
		StoreHelper:     storeHelper,
	}

	tests := []struct {
		name    string
		args    args
		want    *CoreEngine
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				cf: cf,
			},
			want: func() *CoreEngine {
				core := &CoreEngine{
					msgBus:          cf.MsgBus,
					txPool:          cf.TxPool,
					vmMgr:           cf.VmMgr,
					blockchainStore: cf.BlockchainStore,
					snapshotManager: cf.SnapshotManager,
					proposedCache:   cf.ProposalCache,
					chainConf:       cf.ChainConf,
					log:             cf.Log,
				}

				var schedulerFactory scheduler.TxSchedulerFactory
				core.txScheduler = schedulerFactory.NewTxScheduler(
					cf.VmMgr, cf.ChainConf, cf.StoreHelper, ledgerCache)
				core.quitC = make(<-chan interface{})

				var err error
				proposerConfig := proposer.BlockProposerConfig{
					ChainId:         cf.ChainId,
					TxPool:          cf.TxPool,
					SnapshotManager: cf.SnapshotManager,
					MsgBus:          cf.MsgBus,
					Identity:        cf.Identity,
					LedgerCache:     cf.LedgerCache,
					TxScheduler:     core.txScheduler,
					ProposalCache:   cf.ProposalCache,
					ChainConf:       cf.ChainConf,
					AC:              cf.AC,
					BlockchainStore: cf.BlockchainStore,
					StoreHelper:     cf.StoreHelper,
				}
				core.blockProposer, err = proposer.NewBlockProposer(proposerConfig, cf.Log)
				if err != nil {
					t.Error(err)
					return nil
				}

				verifierConfig := verifier.BlockVerifierConfig{
					ChainId:         cf.ChainId,
					MsgBus:          cf.MsgBus,
					SnapshotManager: cf.SnapshotManager,
					BlockchainStore: cf.BlockchainStore,
					LedgerCache:     cf.LedgerCache,
					TxScheduler:     core.txScheduler,
					ProposedCache:   cf.ProposalCache,
					ChainConf:       cf.ChainConf,
					AC:              cf.AC,
					TxPool:          cf.TxPool,
					VmMgr:           cf.VmMgr,
					StoreHelper:     cf.StoreHelper,
				}
				core.BlockVerifier, err = verifier.NewBlockVerifier(verifierConfig, cf.Log)
				if err != nil {
					t.Error(err)
					return nil
				}

				committerConfig := common.BlockCommitterConfig{
					ChainId:         cf.ChainId,
					BlockchainStore: cf.BlockchainStore,
					SnapshotManager: cf.SnapshotManager,
					TxPool:          cf.TxPool,
					LedgerCache:     cf.LedgerCache,
					ProposedCache:   cf.ProposalCache,
					ChainConf:       cf.ChainConf,
					MsgBus:          cf.MsgBus,
					Subscriber:      cf.Subscriber,
					Verifier:        core.BlockVerifier,
					StoreHelper:     cf.StoreHelper,
				}
				core.BlockCommitter, err = common.NewBlockCommitter(committerConfig, cf.Log)
				if err != nil {
					return nil
				}

				return core
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCoreEngine(tt.args.cf)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCoreEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if reflect.DeepEqual(got.txScheduler, tt.want.txScheduler) {
				t.Errorf("NewCoreEngine() got txScheduler = %v, equal want txScheduler %v", got, tt.want)
			}

			if reflect.DeepEqual(got.blockProposer, tt.want.blockProposer) {
				t.Errorf("NewCoreEngine() got blockProposer = %v, equal want blockProposer %v", got, tt.want)
			}

			if reflect.DeepEqual(got.BlockVerifier, tt.want.BlockVerifier) {
				t.Errorf("NewCoreEngine() got BlockVerifier = %v equal want BlockVerifier %v", got, tt.want)
			}

			if reflect.DeepEqual(got.BlockCommitter, tt.want.BlockCommitter) {
				t.Errorf("NewCoreEngine() got BlockCommitter = %v equal want BlockCommitter %v", got, tt.want)
			}
		})
	}
}

func TestCoreEngine_OnQuit(t *testing.T) {
	type fields struct {
		log protocol.Logger
	}

	log := newMockLogger(t)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				log: log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				log: tt.fields.log,
			}
			c.OnQuit()
		})
	}
}

func TestCoreEngine_OnMessage(t *testing.T) {
	type fields struct {
		msgBus msgbus.MessageBus
	}
	type args struct {
		message *msgbus.Message
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "test0",
			fields: fields{},
			args: args{
				message: func() *msgbus.Message {
					mBus := &msgbus.Message{
						Topic: msgbus.ProposeState,
					}
					return mBus
				}(),
			},
		},
		{
			name:   "test1",
			fields: fields{},
			args: args{
				message: func() *msgbus.Message {
					mBus := &msgbus.Message{
						Topic: msgbus.VerifyBlock,
					}
					return mBus
				}(),
			},
		},
		{
			name:   "test2",
			fields: fields{},
			args: args{
				message: func() *msgbus.Message {
					mBus := &msgbus.Message{
						Topic: msgbus.CommitBlock,
					}
					return mBus
				}(),
			},
		},
		{
			name:   "test3",
			fields: fields{},
			args: args{
				message: func() *msgbus.Message {
					mBus := &msgbus.Message{
						Topic: msgbus.TxPoolSignal,
					}
					return mBus
				}(),
			},
		},
		{
			name:   "test4",
			fields: fields{},
			args: args{
				message: func() *msgbus.Message {
					mBus := &msgbus.Message{
						Topic: msgbus.BuildProposal,
					}
					return mBus
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				msgBus: tt.fields.msgBus,
			}
			c.OnMessage(tt.args.message)
		})
	}
}

func TestCoreEngine_Start(t *testing.T) {
	type fields struct {
		//msgBus        msgbus.MessageBus
		//blockProposer protocol.BlockProposer
		//log           protocol.Logger
	}

	var (
		msgBus   = newMockMessageBus(t)
		proposer = newMockBlockProposer(t)
		log      = newMockLogger(t)
	)
	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
	proposer.EXPECT().Start().AnyTimes()
	log.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "test0",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				msgBus:        msgBus,
				blockProposer: proposer,
				log:           log,
			}
			c.Start()
		})
	}
}

func TestCoreEngine_Stop(t *testing.T) {
	type fields struct {
		msgBus        msgbus.MessageBus
		blockProposer protocol.BlockProposer
		log           protocol.Logger
	}

	var (
		msgBus   = newMockMessageBus(t)
		proposer = newMockBlockProposer(t)
		log      = newMockLogger(t)
	)
	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
	proposer.EXPECT().Stop().AnyTimes()
	log.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				msgBus:        msgBus,
				blockProposer: proposer,
				log:           log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				msgBus:        tt.fields.msgBus,
				blockProposer: tt.fields.blockProposer,
				log:           tt.fields.log,
			}
			c.Stop()
		})
	}
}

func TestCoreEngine_GetBlockCommitter(t *testing.T) {
	type fields struct {
		BlockCommitter protocol.BlockCommitter
	}
	tests := []struct {
		name   string
		fields fields
		want   protocol.BlockCommitter
	}{
		{
			name: "test0",
			fields: fields{
				BlockCommitter: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				BlockCommitter: tt.fields.BlockCommitter,
			}
			if got := c.GetBlockCommitter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockCommitter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoreEngine_GetBlockVerifier(t *testing.T) {
	type fields struct {
		BlockVerifier protocol.BlockVerifier
	}
	tests := []struct {
		name   string
		fields fields
		want   protocol.BlockVerifier
	}{
		{
			name: "test0",
			fields: fields{
				BlockVerifier: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				BlockVerifier: tt.fields.BlockVerifier,
			}
			if got := c.GetBlockVerifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockVerifier() = %v, want %v", got, tt.want)
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

func newMockStoreHelper(t *testing.T) *mock.MockStoreHelper {
	ctrl := gomock.NewController(t)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	return storeHelper
}

func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

func newMockVmManager(t *testing.T) *mock.MockVmManager {
	ctrl := gomock.NewController(t)
	vmManager := mock.NewMockVmManager(ctrl)
	vmManager.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonpb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonpb.TxStatusCode_SUCCESS).AnyTimes()
	return vmManager
}

func newMockTxPool(t *testing.T) *mock.MockTxPool {
	ctrl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctrl)
	return txPool
}

func newMockSnapshotManager(t *testing.T) *mock.MockSnapshotManager {
	ctrl := gomock.NewController(t)
	snapshotManager := mock.NewMockSnapshotManager(ctrl)
	return snapshotManager
}

func newMockLedgerCache(t *testing.T) *mock.MockLedgerCache {
	ctrl := gomock.NewController(t)
	newMockLedgerCache := mock.NewMockLedgerCache(ctrl)
	return newMockLedgerCache
}

func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
	ctrl := gomock.NewController(t)
	proposalCache := mock.NewMockProposalCache(ctrl)
	return proposalCache
}

func newMockBlockVerifier(t *testing.T) *mock.MockBlockVerifier {
	ctrl := gomock.NewController(t)
	blockVerifier := mock.NewMockBlockVerifier(ctrl)
	return blockVerifier
}

func newMockBlockCommitter(t *testing.T) *mock.MockBlockCommitter {
	ctrl := gomock.NewController(t)
	blockCommitter := mock.NewMockBlockCommitter(ctrl)
	return blockCommitter
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

func newMockTxScheduler(t *testing.T) *mock.MockTxScheduler {
	ctrl := gomock.NewController(t)
	txScheduler := mock.NewMockTxScheduler(ctrl)
	return txScheduler
}

func newMockMessageBus(t *testing.T) *msgbusMock.MockMessageBus {
	ctrl := gomock.NewController(t)
	messageBus := msgbusMock.NewMockMessageBus(ctrl)
	return messageBus
}

func newMockBlockProposer(t *testing.T) *mock.MockBlockProposer {
	ctrl := gomock.NewController(t)
	blockProposer := mock.NewMockBlockProposer(ctrl)
	return blockProposer
}

//func createBlockByHash(height uint64, hash []byte) *commonpb.Block {
//	//var hash = []byte("0123456789")
//	var version = uint32(1)
//	var block = &commonpb.Block{
//		Header: &commonpb.BlockHeader{
//			ChainId:        "Chain1",
//			BlockHeight:    height,
//			PreBlockHash:   hash,
//			BlockHash:      hash,
//			PreConfHeight:  0,
//			BlockVersion:   version,
//			DagHash:        hash,
//			RwSetRoot:      hash,
//			TxRoot:         hash,
//			BlockTimestamp: 0,
//			Proposer:       &accesscontrol.Member{MemberInfo: hash},
//			ConsensusArgs:  nil,
//			TxCount:        1,
//			Signature:      []byte(""),
//		},
//		Dag: &commonpb.DAG{
//			Vertexes: nil,
//		},
//		Txs: nil,
//	}
//
//	return block
//}

func TestCoreEngine_GetMaxbftHelper(t *testing.T) {
	type fields struct {
		MaxbftHelper protocol.MaxbftHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   protocol.MaxbftHelper
	}{
		{
			name:   "test0",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoreEngine{
				MaxbftHelper: tt.fields.MaxbftHelper,
			}
			if got := c.GetMaxbftHelper(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaxbftHelper() = %v, want %v", got, tt.want)
			}
		})
	}
}
