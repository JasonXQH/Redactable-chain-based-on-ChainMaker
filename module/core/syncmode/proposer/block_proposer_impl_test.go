/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package proposer

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"chainmaker.org/chainmaker-go/module/core/cache"
	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker/common/v2/json"
	mbusmock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/pb-go/v2/consensus/maxbft"
	txpoolpb "chainmaker.org/chainmaker/pb-go/v2/txpool"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/utils/v2"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	chainId      = "Chain1"
	contractName = "contractName"
	log          = logger.GetLoggerByChain(logger.MODULE_CORE, "Chain1")
)

func TestProposeStatusChange(t *testing.T) {
	ctl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctl)
	snapshotMgr := mock.NewMockSnapshotManager(ctl)
	msgBus := mbusmock.NewMockMessageBus(ctl)
	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
	identity := mock.NewMockSigningMember(ctl)
	ledgerCache := cache.NewLedgerCache(chainId)
	proposedCache := cache.NewProposalCache(nil, ledgerCache, log)
	txScheduler := mock.NewMockTxScheduler(ctl)
	blockChainStore := mock.NewMockBlockchainStore(ctl)
	chainConf := mock.NewMockChainConf(ctl)
	storeHelper := common.NewKVStoreHelper("chain1")
	filter := mock.NewMockTxFilter(ctl)
	ledgerCache.SetLastCommittedBlock(createNewTestBlock(0))

	txs := make([]*commonpb.Transaction, 0)

	for i := 0; i < 5; i++ {
		txs = append(txs, createNewTestTx("txId"+fmt.Sprint(i)))
		blockChainStore.EXPECT().TxExists("txId" + fmt.Sprint(i)).AnyTimes()
	}

	identity.EXPECT().GetMember().AnyTimes()
	txPool.EXPECT().FetchTxs(gomock.Any()).Return(txs).AnyTimes()
	txPool.EXPECT().GetPoolStatus().Return(&txpoolpb.TxPoolStatus{}).AnyTimes()
	//msgBus.EXPECT().Publish(gomock.Any(), gomock.Any())
	txPool.EXPECT().RetryAndRemoveTxs(gomock.Any(), gomock.Any())
	filter.EXPECT().ValidateRule(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	consensus := configpb.ConsensusConfig{
		Type: consensus.ConsensusType_TBFT,
	}
	blockConf := configpb.BlockConfig{
		TxTimestampVerify: false,
		TxTimeout:         1000000000,
		BlockTxCapacity:   100,
		BlockSize:         100000,
		BlockInterval:     1000,
	}
	crypro := configpb.CryptoConfig{Hash: "SHA256"}
	contract := configpb.ContractConfig{EnableSqlSupport: false}
	chainConfig := configpb.ChainConfig{
		Consensus: &consensus,
		Block:     &blockConf,
		Contract:  &contract,
		Crypto:    &crypro,
		Core: &configpb.CoreConfig{
			TxSchedulerTimeout:         0,
			TxSchedulerValidateTimeout: 0,
			ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
				ConsensusMessageTurbo: false,
				RetryTime:             0,
				RetryInterval:         0,
			},
		}}
	chainConf.EXPECT().ChainConfig().Return(&chainConfig).AnyTimes()
	snapshot := mock.NewMockSnapshot(ctl)
	snapshot.EXPECT().GetBlockchainStore().AnyTimes()
	snapshotMgr.EXPECT().NewSnapshot(gomock.Any(), gomock.Any()).AnyTimes().Return(snapshot)
	txScheduler.EXPECT().Schedule(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	logger := logger.GetLoggerByChain(logger.MODULE_CORE, chainId)
	blockBuilderConf := &common.BlockBuilderConf{
		ChainId:         "chain1",
		TxPool:          txPool,
		TxScheduler:     txScheduler,
		SnapshotManager: snapshotMgr,
		Identity:        identity,
		LedgerCache:     ledgerCache,
		ProposalCache:   proposedCache,
		ChainConf:       chainConf,
		Log:             logger,
		StoreHelper:     storeHelper,
	}
	blockBuilder := common.NewBlockBuilder(blockBuilderConf)

	blockProposer := &BlockProposerImpl{
		chainId:         chainId,
		isProposer:      false, // not proposer when initialized
		idle:            true,
		msgBus:          msgBus,
		canProposeC:     make(chan bool),
		txPoolSignalC:   make(chan *txpoolpb.TxPoolSignal),
		proposeTimer:    nil,
		exitC:           make(chan bool),
		txPool:          txPool,
		snapshotManager: snapshotMgr,
		txScheduler:     txScheduler,
		identity:        identity,
		ledgerCache:     ledgerCache,
		proposalCache:   proposedCache,
		log:             logger,
		finishProposeC:  make(chan bool),
		blockchainStore: blockChainStore,
		chainConf:       chainConf,
		txFilter:        filter,
		blockBuilder:    blockBuilder,
		storeHelper:     storeHelper,
	}
	require.False(t, blockProposer.isProposer)
	require.Nil(t, blockProposer.proposeTimer)

	blockProposer.proposeBlock()
	blockProposer.OnReceiveYieldProposeSignal(true)
}

func TestShouldPropose(t *testing.T) {
	ctl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctl)
	snapshotMgr := mock.NewMockSnapshotManager(ctl)
	msgBus := mbusmock.NewMockMessageBus(ctl)
	identity := mock.NewMockSigningMember(ctl)
	ledgerCache := cache.NewLedgerCache(chainId)
	proposedCache := cache.NewProposalCache(nil, ledgerCache, log)
	txScheduler := mock.NewMockTxScheduler(ctl)

	ledgerCache.SetLastCommittedBlock(createNewTestBlock(0))
	blockProposer := &BlockProposerImpl{
		chainId:         chainId,
		isProposer:      false, // not proposer when initialized
		idle:            true,
		msgBus:          msgBus,
		canProposeC:     make(chan bool),
		txPoolSignalC:   make(chan *txpoolpb.TxPoolSignal),
		proposeTimer:    nil,
		exitC:           make(chan bool),
		txPool:          txPool,
		snapshotManager: snapshotMgr,
		txScheduler:     txScheduler,
		identity:        identity,
		ledgerCache:     ledgerCache,
		proposalCache:   proposedCache,
		log:             logger.GetLoggerByChain(logger.MODULE_CORE, chainId),
	}

	b0 := createNewTestBlock(0)
	ledgerCache.SetLastCommittedBlock(b0)
	require.True(t, blockProposer.shouldProposeByBFT(b0.Header.BlockHeight+1))

	b := createNewTestBlock(1)
	proposedCache.SetProposedBlock(b, nil, nil, false)
	require.Nil(t, proposedCache.GetSelfProposedBlockAt(1))
	b1, _, _ := proposedCache.GetProposedBlock(b)
	require.NotNil(t, b1)

	b2 := createNewTestBlock(1)
	b2.Header.BlockHash = nil
	proposedCache.SetProposedBlock(b2, nil, nil, true)
	require.False(t, blockProposer.shouldProposeByBFT(b2.Header.BlockHeight+1))
	require.NotNil(t, proposedCache.GetSelfProposedBlockAt(1))
	ledgerCache.SetLastCommittedBlock(b2)
	require.True(t, blockProposer.shouldProposeByBFT(b2.Header.BlockHeight+1))

	b3, _, _ := proposedCache.GetProposedBlock(b2)
	require.NotNil(t, b3)

	proposedCache.SetProposedAt(b3.Header.BlockHeight)
	require.False(t, blockProposer.shouldProposeByBFT(b3.Header.BlockHeight))
}

func TestYieldGoRountine(t *testing.T) {
	exitC := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second)
		exitC <- true
	}()

	sig := <-exitC
	require.True(t, sig)
	fmt.Println("exit1")
}

func TestHash(t *testing.T) {
	txCount := 50000
	txs := make([][]byte, 0)
	for i := 0; i < txCount; i++ {
		txId := uuid.GetUUID() + uuid.GetUUID()
		txs = append(txs, []byte(txId))
	}
	require.Equal(t, txCount, len(txs))
	hf := sha256.New()

	start := utils.CurrentTimeMillisSeconds()
	for _, txId := range txs {
		hf.Write(txId)
		hf.Sum(nil)
		hf.Reset()
	}
	fmt.Println(utils.CurrentTimeMillisSeconds() - start)
}

func TestFinalize(t *testing.T) {
	txCount := 50000
	dag := &commonpb.DAG{Vertexes: make([]*commonpb.DAG_Neighbor, txCount)}
	txRead := &commonpb.TxRead{
		Key:          []byte("key"),
		Value:        []byte("value"),
		ContractName: contractName,
		Version:      nil,
	}
	txReads := make([]*commonpb.TxRead, 5)
	for i := 0; i < 5; i++ {
		txReads[i] = txRead
	}
	block := &commonpb.Block{
		Header: &commonpb.BlockHeader{
			ChainId:        "chain1",
			BlockHeight:    0,
			PreBlockHash:   nil,
			BlockHash:      nil,
			PreConfHeight:  0,
			BlockVersion:   1,
			DagHash:        nil,
			RwSetRoot:      nil,
			TxRoot:         nil,
			BlockTimestamp: 0,
			Proposer: &accesscontrol.Member{
				OrgId:      "org1",
				MemberType: 0,
				MemberInfo: nil,
			},
			ConsensusArgs: nil,
			TxCount:       uint32(txCount),
			Signature:     nil,
		},
		Dag:            nil,
		Txs:            nil,
		AdditionalData: nil,
	}
	txs := make([]*commonpb.Transaction, 0)
	rwSetMap := make(map[string]*commonpb.TxRWSet)
	for i := 0; i < txCount; i++ {
		dag.Vertexes[i] = &commonpb.DAG_Neighbor{
			Neighbors: nil,
		}
		txId := uuid.GetUUID() + uuid.GetUUID()
		payload := parsePayload(txId)
		payloadBytes, _ := json.Marshal(payload)
		tx := parseTx(txId, payloadBytes)
		txs = append(txs, tx)
		txWrite := &commonpb.TxWrite{
			Key:          []byte(txId),
			Value:        payloadBytes,
			ContractName: contractName,
		}
		txWrites := make([]*commonpb.TxWrite, 0)
		txWrites = append(txWrites, txWrite)
		rwSetMap[txId] = &commonpb.TxRWSet{
			TxId:     txId,
			TxReads:  txReads,
			TxWrites: txWrites,
		}
	}
	require.Equal(t, txCount, len(txs))
	block.Txs = txs
	block.Dag = dag

	kvs := []*configpb.ConfigKeyValue{
		{Key: "IsExtreme", Value: "true"},
	}

	err := localconf.UpdateDebugConfig(kvs)
	require.Nil(t, err)
}

func parsePayload(txId string) *commonpb.Payload {
	pairs := []*commonpb.KeyValuePair{
		{
			Key:   "file_hash",
			Value: []byte(txId)[len(txId)/2:],
		},
	}

	return &commonpb.Payload{
		ChainId:        "chain1",
		TxType:         0,
		TxId:           "txId1",
		Timestamp:      0,
		ExpirationTime: 0,
		ContractName:   contractName,
		Method:         "save",
		Parameters:     pairs,
		Sequence:       1,
		Limit:          nil,
	}
}

func parseTx(txId string, payloadBytes []byte) *commonpb.Transaction {
	return &commonpb.Transaction{
		Payload: &commonpb.Payload{
			ChainId:        "chain1",
			TxType:         0,
			TxId:           "txId1",
			Timestamp:      0,
			ExpirationTime: 0,
			ContractName:   contractName,
			Method:         "save",
			Parameters:     nil,
			Sequence:       1,
			Limit:          nil,
		},
		Sender:    nil,
		Endorsers: nil,
		Result: &commonpb.Result{
			Code: 0,
			ContractResult: &commonpb.ContractResult{
				Code:    0,
				Result:  payloadBytes,
				Message: "SUCCESS",
				GasUsed: 0,
			},
			RwSetHash: nil,
		},
	}

}

func createNewTestBlock(height uint64) *commonpb.Block {
	var hash = []byte("0123456789")
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockVersion:   1,
			BlockType:      0,
			ChainId:        "chain1",
			BlockHeight:    height,
			BlockHash:      hash,
			PreBlockHash:   hash,
			PreConfHeight:  0,
			TxCount:        0,
			TxRoot:         hash,
			DagHash:        hash,
			RwSetRoot:      hash,
			BlockTimestamp: utils.CurrentTimeMillisSeconds(),
			ConsensusArgs:  hash,
			Proposer: &accesscontrol.Member{
				OrgId:      "org1",
				MemberType: 0,
				MemberInfo: hash,
			},
			Signature: hash,
		},
		Dag:            &commonpb.DAG{Vertexes: nil},
		Txs:            nil,
		AdditionalData: nil,
	}

	tx := createNewTestTx("txId1")
	txs := make([]*commonpb.Transaction, 1)
	txs[0] = tx
	block.Txs = txs
	return block
}

func createNewTestTx(txId string) *commonpb.Transaction {
	var hash = []byte("0123456789")
	return &commonpb.Transaction{
		Payload: &commonpb.Payload{
			ChainId:        "chain1",
			TxType:         0,
			TxId:           txId,
			Timestamp:      0,
			ExpirationTime: 0,
			ContractName:   "fact",
			Method:         "set",
			Parameters:     nil,
			Sequence:       1,
			Limit:          nil,
		},
		Sender: &commonpb.EndorsementEntry{
			Signer:    nil,
			Signature: nil,
		},
		Endorsers: nil,
		Result: &commonpb.Result{
			Code:           0,
			ContractResult: nil,
			RwSetHash:      hash,
			Message:        "",
		},
	}
}

//func TestBlockProposerImpl_proposing(t *testing.T) {
//	type fields struct {
//		chainId                string
//		txPool                 protocol.TxPool
//		txScheduler            protocol.TxScheduler
//		snapshotManager        protocol.SnapshotManager
//		identity               protocol.SigningMember
//		ledgerCache            protocol.LedgerCache
//		msgBus                 msgbus.MessageBus
//		ac                     protocol.AccessControlProvider
//		blockchainStore        protocol.BlockchainStore
//		isProposer             bool
//		idle                   bool
//		proposeTimer           *time.Timer
//		canProposeC            chan bool
//		txPoolSignalC          chan *txpoolpb.TxPoolSignal
//		exitC                  chan bool
//		proposalCache          protocol.ProposalCache
//		chainConf              protocol.ChainConf
//		idleMu                 sync.Mutex
//		statusMu               sync.Mutex
//		proposerMu             sync.RWMutex
//		log                    protocol.Logger
//		finishProposeC         chan bool
//		metricBlockPackageTime *prometheus.HistogramVec
//		//proposer               *pbac.Member
//		blockBuilder *common.BlockBuilder
//		storeHelper  conf.StoreHelper
//	}
//
//	txPool := newMockTxPool(t)
//	snapshotMgr := newMockSnapshotManager(t)
//	msgBus := newMockMessageBus(t)
//	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
//	identity := newMockSigningMember(t)
//	ledgerCache := newMockLedgerCache(t)
//	proposedCache := newMockProposalCache(t)
//	txScheduler := newMockTxScheduler(t)
//	blockChainStore := newMockBlockchainStore(t)
//	chainConf := newMockChainConf(t)
//	storeHelper := newMockStoreHelper(t)
//	//signingMember := newMockSigningMember(t)
//	ac := newMockAccessControlProvider(t)
//	log := newMockLogger(t)
//	//ledgerCache.SetLastCommittedBlock(createNewTestBlock(0))
//
//	type args struct {
//		height  uint64
//		preHash []byte
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   *commonpb.Block
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         "123456",
//				txPool:          txPool,
//				txScheduler:     txScheduler,
//				snapshotManager: snapshotMgr,
//				identity:        identity,
//				ledgerCache:     ledgerCache,
//				msgBus:          msgBus,
//				ac:              ac,
//				blockchainStore: blockChainStore,
//				isProposer:      false,
//				idle:            false,
//				proposeTimer:    nil,
//				canProposeC:     nil,
//				txPoolSignalC:   nil,
//				exitC:           nil,
//				proposalCache:   proposedCache,
//				chainConf:       chainConf,
//				//idleMu:                 sync.Mutex{},
//				//statusMu:               sync.Mutex{},
//				//proposerMu:             sync.RWMutex{},
//				log:                    log,
//				finishProposeC:         nil,
//				metricBlockPackageTime: nil,
//				//proposer:               nil,
//				blockBuilder: nil,
//				storeHelper:  storeHelper,
//			},
//			args: args{
//
//			},
//			want: nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			bp := &BlockProposerImpl{
//				chainId:         tt.fields.chainId,
//				txPool:          tt.fields.txPool,
//				txScheduler:     tt.fields.txScheduler,
//				snapshotManager: tt.fields.snapshotManager,
//				identity:        tt.fields.identity,
//				ledgerCache:     tt.fields.ledgerCache,
//				msgBus:          tt.fields.msgBus,
//				ac:              tt.fields.ac,
//				blockchainStore: tt.fields.blockchainStore,
//				isProposer:      tt.fields.isProposer,
//				idle:            tt.fields.idle,
//				//proposeTimer:           tt.fields.proposeTimer,
//				//canProposeC:            tt.fields.canProposeC,
//				//txPoolSignalC:          tt.fields.txPoolSignalC,
//				//exitC:                  tt.fields.exitC,
//				proposalCache: tt.fields.proposalCache,
//				chainConf:     tt.fields.chainConf,
//				//idleMu:                 tt.fields.idleMu,
//				//statusMu:               tt.fields.statusMu,
//				//proposerMu:             tt.fields.proposerMu,
//				log:                    tt.fields.log,
//				finishProposeC:         tt.fields.finishProposeC,
//				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
//				//proposer:               tt.fields.proposer,
//				blockBuilder: tt.fields.blockBuilder,
//				storeHelper:  tt.fields.storeHelper,
//			}
//			if got := bp.proposing(tt.args.height, tt.args.preHash); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("proposing() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestBlockProposerImpl_OnReceiveProposeStatusChange(t *testing.T) {
	type fields struct {
		log           protocol.Logger
		isProposer    bool
		ledgerCache   protocol.LedgerCache
		proposalCache protocol.ProposalCache
		proposeTimer  *time.Timer
		idle          bool
	}
	type args struct {
		proposeStatus bool
	}

	ledgerCache := newMockLedgerCache(t)
	log := newMockLogger(t)
	proposalCache := newMockProposalCache(t)

	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Debug(gomock.Any()).AnyTimes()

	ledgerCache.EXPECT().CurrentHeight().Return(uint64(3), nil).AnyTimes()
	proposalCache.EXPECT().ResetProposedAt(gomock.Any()).AnyTimes()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				log:         log,
				ledgerCache: ledgerCache,
				isProposer:  false,
			},
			args: args{
				proposeStatus: false,
			},
		},
		{
			name: "test1",
			fields: fields{
				log:           log,
				ledgerCache:   ledgerCache,
				isProposer:    false,
				proposalCache: proposalCache,
				proposeTimer:  time.NewTimer(1 * time.Second),
			},
			args: args{
				proposeStatus: true,
			},
		},
		{
			name: "test2",
			fields: fields{
				log:           log,
				ledgerCache:   ledgerCache,
				isProposer:    true,
				proposalCache: proposalCache,
				proposeTimer:  time.NewTimer(2 * time.Second),
				idle:          true,
			},
			args: args{
				proposeStatus: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				log:           tt.fields.log,
				ledgerCache:   tt.fields.ledgerCache,
				isProposer:    tt.fields.isProposer,
				proposalCache: tt.fields.proposalCache,
				proposeTimer:  tt.fields.proposeTimer,
				idle:          tt.fields.idle,
			}
			bp.OnReceiveProposeStatusChange(tt.args.proposeStatus)
		})
	}
}

func TestBlockProposerImpl_OnReceiveMaxBFTProposal(t *testing.T) {
	type fields struct {
		isProposer bool
	}
	type args struct {
		proposal *maxbft.BuildProposal
	}

	preBlock := createNewTestBlock(0)
	ledgerCache := newMockLedgerCache(t)
	log := newMockLogger(t)
	commonBlock := createNewTestBlock(2)

	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	header := *preBlock.Header
	header.Signature = nil
	header.BlockHash = nil
	preHash, _ := proto.Marshal(&header)

	commonBlock.Header.BlockHash = preHash
	ledgerCache.EXPECT().GetLastCommittedBlock().Return(commonBlock).AnyTimes()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				isProposer: false,
			},
			args: args{
				proposal: &maxbft.BuildProposal{
					Height:  3,
					PreHash: preHash,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				ledgerCache: ledgerCache,
				log:         log,
			}
			bp.OnReceiveMaxBFTProposal(tt.args.proposal)
		})
	}
}

func TestBlockProposerImpl_yieldProposing(t *testing.T) {
	type fields struct {
		idle           bool
		finishProposeC chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				idle: false,
				finishProposeC: func() chan bool {
					finishProposeC := make(chan bool, 1)
					return finishProposeC
				}(),
			},
			want: true,
		},
		{
			name: "test1",
			fields: fields{
				idle: true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				idle:           tt.fields.idle,
				finishProposeC: tt.fields.finishProposeC,
			}

			if got := bp.yieldProposing(); got != tt.want {
				t.Errorf("yieldProposing() = %v, want %v", got, tt.want)
			} else {
				if tt.fields.finishProposeC != nil {
					res := <-tt.fields.finishProposeC
					t.Log(res)
					close(tt.fields.finishProposeC)
				}
			}
		})
	}
}

func TestBlockProposerImpl_getDuration(t *testing.T) {
	type fields struct {
		chainConf protocol.ChainConf
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name:   "test0",
			fields: fields{},
			want:   1000000000,
		},
		{
			name: "test1",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConfig := &configpb.ChainConfig{
						Block: &configpb.BlockConfig{
							BlockInterval: 10,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			want: 10000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainConf: tt.fields.chainConf,
			}
			if got := bp.getDuration(); got != tt.want {
				t.Errorf("getDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockProposerImpl_getChainVersion(t *testing.T) {
	type fields struct {
		chainConf protocol.ChainConf
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "test0",
			fields: fields{
				chainConf: nil,
			},
			want: []byte(DEFAULTVERSION),
		},
		{
			name: "test1",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConfig := &configpb.ChainConfig{
						Version: "v1.1.1",
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
			},
			want: []byte("v1.1.1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainConf: tt.fields.chainConf,
			}
			if got := bp.getChainVersion(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getChainVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockProposerImpl_setNotIdle(t *testing.T) {
	type fields struct {
		idle bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				idle: false,
			},
			want: false,
		},
		{
			name: "test1",
			fields: fields{
				idle: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				idle: tt.fields.idle,
			}
			if got := bp.setNotIdle(); got != tt.want {
				t.Errorf("setNotIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockProposerImpl_isIdle(t *testing.T) {
	type fields struct {
		idle bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				idle: false,
			},
			want: false,
		},
		{
			name: "test1",
			fields: fields{
				idle: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				idle: tt.fields.idle,
			}
			if got := bp.isIdle(); got != tt.want {
				t.Errorf("isIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockProposerImpl_setIdle(t *testing.T) {
	type fields struct {
		idle bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				idle: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				idle: tt.fields.idle,
			}
			bp.setIdle()
		})
	}
}

func TestBlockProposerImpl_setIsSelfProposer(t *testing.T) {
	type fields struct {
		isProposer   bool
		proposeTimer *time.Timer
	}
	type args struct{}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				isProposer:   false,
				proposeTimer: time.NewTimer(1 * time.Second),
			},
			args: args{},
		},
		{
			name: "test1",
			fields: fields{
				isProposer:   true,
				proposeTimer: time.NewTimer(1 * time.Second),
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				isProposer:   tt.fields.isProposer,
				proposeTimer: tt.fields.proposeTimer,
			}
			bp.setIsSelfProposer(tt.fields.isProposer)
		})
	}
}

func TestBlockProposerImpl_isSelfProposer(t *testing.T) {
	type fields struct {
		isProposer bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "test0",
			fields: fields{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				isProposer: tt.fields.isProposer,
			}
			if got := bp.isSelfProposer(); got != tt.want {
				t.Errorf("isSelfProposer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newMockChainConf(t *testing.T) *mock.MockChainConf {
	ctrl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctrl)
	return chainConf
}

func newMockLedgerCache(t *testing.T) *mock.MockLedgerCache {
	ctrl := gomock.NewController(t)
	newMockLedgerCache := mock.NewMockLedgerCache(ctrl)
	return newMockLedgerCache
}

func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
	ctrl := gomock.NewController(t)
	proposalCache := mock.NewMockProposalCache(ctrl)
	return proposalCache
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

func newMockTxScheduler(t *testing.T) *mock.MockTxScheduler {
	ctrl := gomock.NewController(t)
	txScheduler := mock.NewMockTxScheduler(ctrl)
	return txScheduler
}

func newMockAccessControlProvider(t *testing.T) *mock.MockAccessControlProvider {
	ctrl := gomock.NewController(t)
	ac := mock.NewMockAccessControlProvider(ctrl)
	return ac
}

func newMockMessageBus(t *testing.T) *mbusmock.MockMessageBus {
	ctrl := gomock.NewController(t)
	messageBus := mbusmock.NewMockMessageBus(ctrl)
	return messageBus
}

func createBlockByHash(height uint64, hash []byte) *commonpb.Block {
	//var hash = []byte("0123456789")
	var version = uint32(1)
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			ChainId:        "Chain1",
			BlockHeight:    height,
			PreBlockHash:   hash,
			BlockHash:      hash,
			PreConfHeight:  0,
			BlockVersion:   version,
			DagHash:        hash,
			RwSetRoot:      hash,
			TxRoot:         hash,
			BlockTimestamp: 0,
			Proposer:       &accesscontrol.Member{MemberInfo: hash},
			ConsensusArgs:  nil,
			TxCount:        1,
			Signature:      []byte(""),
		},
		Dag: &commonpb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}

	return block
}

// run this test with `-race`
func TestBlockProposerImpl_getLastProposeTimeByBlockFinger_raceCondition(t *testing.T) {
	finger1 := "test finger 1"
	bp := &BlockProposerImpl{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		got, err := bp.getLastProposeTimeByBlockFinger(finger1)
		assert.Nil(t, err)
		assert.Greater(t, got, int64(0))
		wg.Done()
	}()
	go func() {
		common.ClearProposeRepeatTimerMap()
		wg.Done()
	}()
	wg.Wait()
}

func TestProposeBlock(t *testing.T) {

	log := logger.GetLoggerByChain("[Core_UT]", "chain1")
	ctl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctl)
	snapshotManager := mock.NewMockSnapshotManager(ctl)
	msgBus := mbusmock.NewMockMessageBus(ctl)
	identity := mock.NewMockSigningMember(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	//consensus := mock.NewMockConsensusEngine(ctl)
	proposedCache := mock.NewMockProposalCache(ctl)
	txScheduler := mock.NewMockTxScheduler(ctl)
	blockChainStore := mock.NewMockBlockchainStore(ctl)
	chainConf := mock.NewMockChainConf(ctl)
	storeHelper := mock.NewMockStoreHelper(ctl)
	ac := mock.NewMockAccessControlProvider(ctl)
	txFilter := mock.NewMockTxFilter(ctl)

	b1 := createNewTestBlock(2)
	ledgerCache.EXPECT().GetLastCommittedBlock().Return(b1).AnyTimes()

	b2 := createNewTestBlock(3)
	proposedCache.EXPECT().GetSelfProposedBlockAt(b2.Header.BlockHeight).Return(b2).AnyTimes()
	proposedCache.EXPECT().SetProposedAt(gomock.Any()).AnyTimes()
	proposedCache.EXPECT().GetProposedBlock(b2).AnyTimes()

	c1 := &configpb.ChainConfig{
		Block: &configpb.BlockConfig{
			TxTimeout: 600,
		},
		Core: &configpb.CoreConfig{
			ConsensusTurboConfig: nil,
		},
	}
	chainConf.EXPECT().ChainConfig().Return(c1).AnyTimes()

	msgBus.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()

	blockProposerImpl := &BlockProposerImpl{
		chainId:         "chain1",
		isProposer:      false, // not proposer when initialized
		idle:            true,
		msgBus:          msgBus,
		blockchainStore: blockChainStore,
		canProposeC:     make(chan bool),
		txPoolSignalC:   make(chan *txpoolpb.TxPoolSignal),
		exitC:           make(chan bool),
		txPool:          txPool,
		snapshotManager: snapshotManager,
		txScheduler:     txScheduler,
		identity:        identity,
		ledgerCache:     ledgerCache,
		proposalCache:   proposedCache,
		chainConf:       chainConf,
		ac:              ac,
		log:             log,
		finishProposeC:  make(chan bool),
		storeHelper:     storeHelper,
		txFilter:        txFilter,
	}

	blockProposerImpl.proposeBlock()

	for i := 0; i < 50; i++ {
		blockProposerImpl.proposeBlock()

		// 模拟重复提案
		time.Sleep(time.Millisecond * 200)
	}

}
