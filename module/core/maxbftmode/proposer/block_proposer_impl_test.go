/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package proposer

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"chainmaker.org/chainmaker-go/module/core/cache"
	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/common/v2/json"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	mbusmock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	pbac "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/pb-go/v2/consensus/maxbft"
	txpoolpb "chainmaker.org/chainmaker/pb-go/v2/txpool"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"chainmaker.org/chainmaker/utils/v2"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

var (
	chainId      = "Chain1"
	contractName = "contractName"
	log          = logger.GetLoggerByChain(logger.MODULE_CORE, chainId)
)

/*
 * test unit ProposeStatusChange func
 */
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
	txPool.EXPECT().RetryAndRemoveTxs(gomock.Any(), gomock.Any()).AnyTimes()

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

	blockBuilderConf := &common.BlockBuilderConf{
		ChainId:         "chain1",
		TxPool:          txPool,
		TxScheduler:     txScheduler,
		SnapshotManager: snapshotMgr,
		Identity:        identity,
		LedgerCache:     ledgerCache,
		ProposalCache:   proposedCache,
		ChainConf:       chainConf,
		Log:             log,
		StoreHelper:     storeHelper,
	}
	blockBuilder := common.NewBlockBuilder(blockBuilderConf)

	blockProposer := &BlockProposerImpl{
		chainId:         chainId,
		idle:            true,
		msgBus:          msgBus,
		exitC:           make(chan bool),
		txPool:          txPool,
		snapshotManager: snapshotMgr,
		txScheduler:     txScheduler,
		identity:        identity,
		ledgerCache:     ledgerCache,
		proposalCache:   proposedCache,
		log:             log,
		finishProposeC:  make(chan bool),
		blockchainStore: blockChainStore,
		chainConf:       chainConf,

		blockBuilder: blockBuilder,
		storeHelper:  storeHelper,
	}

	blockProposer.OnReceiveYieldProposeSignal(true)
}

/*
 * test unit ShouldPropose func
 */
func TestShouldPropose(t *testing.T) {
	ledgerCache := cache.NewLedgerCache(chainId)
	proposedCache := cache.NewProposalCache(nil, ledgerCache, log)
	ledgerCache.SetLastCommittedBlock(createNewTestBlock(0))

	b0 := createNewTestBlock(0)
	ledgerCache.SetLastCommittedBlock(b0)

	b := createNewTestBlock(1)
	proposedCache.SetProposedBlock(b, nil, nil, false)
	require.Nil(t, proposedCache.GetSelfProposedBlockAt(1))
	b1, _, _ := proposedCache.GetProposedBlock(b)
	require.NotNil(t, b1)

	b2 := createNewTestBlock(1)
	b2.Header.BlockHash = nil
	proposedCache.SetProposedBlock(b2, nil, nil, true)

	require.NotNil(t, proposedCache.GetSelfProposedBlockAt(1))
	ledgerCache.SetLastCommittedBlock(b2)

	b3, _, _ := proposedCache.GetProposedBlock(b2)
	require.NotNil(t, b3)

	proposedCache.SetProposedAt(b3.Header.BlockHeight)
}

/*
 * test unit ShouldProposeByMaxBFT func
 */
func TestShouldProposeByMaxBFT(t *testing.T) {
	ctl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctl)
	snapshotMgr := mock.NewMockSnapshotManager(ctl)
	msgBus := mbusmock.NewMockMessageBus(ctl)
	identity := mock.NewMockSigningMember(ctl)
	ledgerCache := cache.NewLedgerCache(chainId)
	proposedCache := cache.NewProposalCache(nil, ledgerCache, log)
	txScheduler := mock.NewMockTxScheduler(ctl)

	ledgerCache.SetLastCommittedBlock(createNewTestBlock(0))

	log := newMockLogger(t)
	log.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	blockProposer := &BlockProposerImpl{
		chainId:         chainId,
		idle:            true,
		msgBus:          msgBus,
		exitC:           make(chan bool),
		txPool:          txPool,
		snapshotManager: snapshotMgr,
		txScheduler:     txScheduler,
		identity:        identity,
		ledgerCache:     ledgerCache,
		proposalCache:   proposedCache,
		log:             log,
	}

	b0 := createNewTestBlock(0)
	ledgerCache.SetLastCommittedBlock(b0)
	require.True(t, blockProposer.shouldProposeByMaxBFT(b0.Header.BlockHeight+1, b0.Header.BlockHash))
	require.False(t, blockProposer.shouldProposeByMaxBFT(b0.Header.BlockHeight+1, []byte("xyz")))
	require.False(t, blockProposer.shouldProposeByMaxBFT(b0.Header.BlockHeight, b0.Header.PreBlockHash))

	b := createNewTestBlock(1)
	proposedCache.SetProposedBlock(b, nil, nil, false)
	require.Nil(t, proposedCache.GetSelfProposedBlockAt(1))
	b1, _, _ := proposedCache.GetProposedBlock(b)
	require.NotNil(t, b1)

	b2 := createNewTestBlock(1)
	b2.Header.BlockHash = nil
	proposedCache.SetProposedBlock(b2, nil, nil, true)
	require.NotNil(t, proposedCache.GetSelfProposedBlockAt(1))
	require.True(t, blockProposer.shouldProposeByMaxBFT(b2.Header.BlockHeight, b0.Header.BlockHash))

	b3, _, _ := proposedCache.GetProposedBlock(b2)
	require.NotNil(t, b3)

}

/*
 * test unit YieldGoRountine func
 */
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

/*
 * test unit Hash func
 */
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

/*
 * test unit Finalize func
 */
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

/*
 * test unit finalizeBlockRoots func
 */
func finalizeBlockRoots() (interface{}, interface{}) {
	return nil, nil
}

/*
 * test unit parseTxs func
 */
func parseTxs(num int) []*commonpb.Transaction {
	txs := make([]*commonpb.Transaction, 0)
	for i := 0; i < num; i++ {
		txId := uuid.GetUUID() + uuid.GetUUID()
		payload := parsePayload(txId)
		payloadBytes, _ := json.Marshal(payload)
		txs = append(txs, parseTx(txId, payloadBytes))
	}
	return txs
}

/*
 * test unit parsePayload func
 */
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

/*
 * test unit parseTx func
 */
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

/*
 * test unit createNewTestBlock func
 */
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
			BlockTimestamp: 0,
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

/*
 * test unit createNewTestTx func
 */
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

/*
 * test unit BlockProposerImpl OnReceiveTxPoolSignal func
 */
func TestBlockProposerImpl_OnReceiveTxPoolSignal(t *testing.T) {
	//type fields struct {
	//	chainId                string
	//	txPool                 protocol.TxPool
	//	txScheduler            protocol.TxScheduler
	//	snapshotManager        protocol.SnapshotManager
	//	identity               protocol.SigningMember
	//	ledgerCache            protocol.LedgerCache
	//	msgBus                 msgbus.MessageBus
	//	ac                     protocol.AccessControlProvider
	//	blockchainStore        protocol.BlockchainStore
	//	idle                   bool
	//	exitC                  chan bool
	//	proposalCache          protocol.ProposalCache
	//	chainConf              protocol.ChainConf
	//	log                    protocol.Logger
	//	finishProposeC         chan bool
	//	metricBlockPackageTime *prometheus.HistogramVec
	//	proposer               *pbac.Member
	//	blockBuilder           *common.BlockBuilder
	//	storeHelper            conf.StoreHelper
	//}
	//type args struct {
	//	txPoolSignal *txpoolpb.TxPoolSignal
	//}
	//tests := []struct {
	//	name   string
	//	fields fields
	//	args   args
	//}{
	//	{
	//		name: "test0",
	//		fields: fields{
	//			chainId:                "test0",
	//			txPool:                 nil,
	//			txScheduler:            nil,
	//			snapshotManager:        nil,
	//			identity:               nil,
	//			ledgerCache:            nil,
	//			msgBus:                 nil,
	//			ac:                     nil,
	//			blockchainStore:        nil,
	//			idle:                   false,
	//			exitC:                  nil,
	//			proposalCache:          nil,
	//			chainConf:              nil,
	//			log:                    nil,
	//			finishProposeC:         nil,
	//			metricBlockPackageTime: nil,
	//			proposer:               nil,
	//			blockBuilder:           nil,
	//			storeHelper:            nil,
	//		},
	//		args: args{
	//			txPoolSignal: &txpoolpb.TxPoolSignal{
	//				SignalType: txpoolpb.SignalType_BLOCK_PROPOSE,
	//				ChainId:    "test123456",
	//			},
	//		},
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		bp := &BlockProposerImpl{
	//			chainId:                tt.fields.chainId,
	//			txPool:                 tt.fields.txPool,
	//			txScheduler:            tt.fields.txScheduler,
	//			snapshotManager:        tt.fields.snapshotManager,
	//			identity:               tt.fields.identity,
	//			ledgerCache:            tt.fields.ledgerCache,
	//			msgBus:                 tt.fields.msgBus,
	//			ac:                     tt.fields.ac,
	//			blockchainStore:        tt.fields.blockchainStore,
	//			idle:                   tt.fields.idle,
	//			exitC:                  tt.fields.exitC,
	//			proposalCache:          tt.fields.proposalCache,
	//			chainConf:              tt.fields.chainConf,
	//			log:                    tt.fields.log,
	//			finishProposeC:         tt.fields.finishProposeC,
	//			metricBlockPackageTime: tt.fields.metricBlockPackageTime,
	//			proposer:               tt.fields.proposer,
	//			blockBuilder:           tt.fields.blockBuilder,
	//			storeHelper:            tt.fields.storeHelper,
	//		}
	//		bp.OnReceiveTxPoolSignal(tt.args.txPoolSignal)
	//	})
	//}
}

/*
 * test unit BlockProposerImpl OnReceiveProposeStatusChange func
 */
func TestBlockProposerImpl_OnReceiveProposeStatusChange(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		idle                   bool
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	type args struct {
		proposeStatus bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				idle:                   false,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    newMockLogger(t),
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			args: args{
				proposeStatus: false,
			},
		},
		{
			name: "test1",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				idle:                   false,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    newMockLogger(t),
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer: &accesscontrol.Member{
					OrgId:      "org1",
					MemberType: 0,
					MemberInfo: nil,
				},
				blockBuilder: nil,
				storeHelper:  nil,
			},
			args: args{
				proposeStatus: true,
			},
		},
		{
			name: "test2",
			fields: fields{
				chainId:         "123456",
				txPool:          nil,
				txScheduler:     nil,
				snapshotManager: nil,
				identity:        nil,
				ledgerCache: func() protocol.LedgerCache {
					ledgerCache := newMockLedgerCache(t)
					ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
					return ledgerCache
				}(),
				msgBus:          nil,
				ac:              nil,
				blockchainStore: nil,
				idle:            false,
				exitC:           nil,
				proposalCache: func() protocol.ProposalCache {
					proposalCache := newMockProposalCache(t)
					proposalCache.EXPECT().ResetProposedAt(gomock.Any()).AnyTimes()
					return proposalCache
				}(),
				chainConf:              nil,
				log:                    newMockLogger(t),
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer: &accesscontrol.Member{
					OrgId:      "org1",
					MemberType: 0,
					MemberInfo: nil,
				},
				blockBuilder: nil,
				storeHelper:  nil,
			},
			args: args{
				proposeStatus: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
				proposeTimer:           time.NewTimer(10),
			}
			bp.OnReceiveProposeStatusChange(tt.args.proposeStatus)
		})
	}
}

/*
 * test unit BlockProposerImpl OnReceiveMaxBFTProposal func
 */
func TestBlockProposerImpl_OnReceiveMaxBFTProposal(t *testing.T) {
	type fields struct {
		chainId     string
		ledgerCache protocol.LedgerCache
		log         protocol.Logger
	}
	type args struct {
		proposal *maxbft.BuildProposal
	}

	var (
		block0      = createNewTestBlock(3)
		ledgerCache = newMockLedgerCache(t)
		log         = newMockLogger(t)
	)

	ledgerCache.EXPECT().GetLastCommittedBlock().Return(block0).AnyTimes()
	log.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				chainId:     "test0",
				ledgerCache: ledgerCache,
				log:         log,
			},
			args: args{
				proposal: &maxbft.BuildProposal{
					Height: 0,
				},
			},
		},
		{
			name: "test1",
			fields: fields{
				chainId:     "test1",
				ledgerCache: ledgerCache,
				log:         log,
			},
			args: args{
				proposal: &maxbft.BuildProposal{
					Height: 4,
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				chainId:     "test2",
				ledgerCache: ledgerCache,
				log:         log,
			},
			args: args{
				proposal: &maxbft.BuildProposal{
					Height:  4,
					PreHash: []byte("0123456789"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:     tt.fields.chainId,
				ledgerCache: tt.fields.ledgerCache,
				log:         log,
			}
			bp.OnReceiveMaxBFTProposal(tt.args.proposal)
		})
	}
}

/*
 * test unit BlockProposerImpl yieldProposing func
 */
func TestBlockProposerImpl_yieldProposing(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		idle                   bool
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				chainId:         "123456",
				txPool:          nil,
				txScheduler:     nil,
				snapshotManager: nil,
				identity:        nil,
				ledgerCache:     nil,
				msgBus:          nil,
				ac:              nil,
				blockchainStore: nil,
				idle:            true,
				exitC:           nil,
				proposalCache:   nil,
				chainConf:       nil,
				log:             nil,
				finishProposeC: func() chan bool {
					c := make(chan bool)
					return c
				}(),
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
			}
			if got := bp.yieldProposing(); got != tt.want {
				t.Errorf("yieldProposing() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit BlockProposerImpl getDuration func
 */
func TestBlockProposerImpl_getDuration(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		isProposer             bool
		idle                   bool
		proposeTimer           *time.Timer
		canProposeC            chan bool
		txPoolSignalC          chan *txpoolpb.TxPoolSignal
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name: "test0",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				isProposer:             false,
				idle:                   false,
				proposeTimer:           nil,
				canProposeC:            nil,
				txPoolSignalC:          nil,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    nil,
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: 1 * time.Second,
		},
		{
			name: "test1",
			fields: fields{
				chainId:         "123456",
				txPool:          nil,
				txScheduler:     nil,
				snapshotManager: nil,
				identity:        nil,
				ledgerCache:     nil,
				msgBus:          nil,
				ac:              nil,
				blockchainStore: nil,
				isProposer:      false,
				idle:            false,
				proposeTimer:    nil,
				canProposeC:     nil,
				txPoolSignalC:   nil,
				exitC:           nil,
				proposalCache:   nil,
				chainConf: func() protocol.ChainConf {
					chainConf := newMockChainConf(t)
					chainConfig := &configpb.ChainConfig{
						Block: &configpb.BlockConfig{
							BlockInterval: DEFAULTDURATION,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log:                    nil,
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
			}
			if got := bp.getDuration(); got != tt.want {
				t.Errorf("getDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit BlockProposerImpl getChainVersion func
 */
func TestBlockProposerImpl_getChainVersion(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		isProposer             bool
		idle                   bool
		proposeTimer           *time.Timer
		canProposeC            chan bool
		txPoolSignalC          chan *txpoolpb.TxPoolSignal
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "test0",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				isProposer:             false,
				idle:                   false,
				proposeTimer:           nil,
				canProposeC:            nil,
				txPoolSignalC:          nil,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    &test.HoleLogger{},
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: protocol.DefaultBlockVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
			}
			if got := bp.getChainVersion(); got != tt.want {
				t.Errorf("getChainVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit BlockProposerImpl setNotIdle func
 */
func TestBlockProposerImpl_setNotIdle(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		isProposer             bool
		idle                   bool
		proposeTimer           *time.Timer
		canProposeC            chan bool
		txPoolSignalC          chan *txpoolpb.TxPoolSignal
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				isProposer:             false,
				idle:                   false,
				proposeTimer:           nil,
				canProposeC:            nil,
				txPoolSignalC:          nil,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    &test.HoleLogger{},
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: false,
		},
		{
			name: "test0",
			fields: fields{
				chainId:                "123456",
				txPool:                 nil,
				txScheduler:            nil,
				snapshotManager:        nil,
				identity:               nil,
				ledgerCache:            nil,
				msgBus:                 nil,
				ac:                     nil,
				blockchainStore:        nil,
				isProposer:             false,
				idle:                   true,
				proposeTimer:           nil,
				canProposeC:            nil,
				txPoolSignalC:          nil,
				exitC:                  nil,
				proposalCache:          nil,
				chainConf:              nil,
				log:                    &test.HoleLogger{},
				finishProposeC:         nil,
				metricBlockPackageTime: nil,
				proposer:               nil,
				blockBuilder:           nil,
				storeHelper:            nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
			}
			if got := bp.setNotIdle(); got != tt.want {
				t.Errorf("setNotIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit BlockProposerImpl setIdle func
 */
func TestBlockProposerImpl_setIdle(t *testing.T) {
	type fields struct {
		chainId                string
		txPool                 protocol.TxPool
		txScheduler            protocol.TxScheduler
		snapshotManager        protocol.SnapshotManager
		identity               protocol.SigningMember
		ledgerCache            protocol.LedgerCache
		msgBus                 msgbus.MessageBus
		ac                     protocol.AccessControlProvider
		blockchainStore        protocol.BlockchainStore
		idle                   bool
		exitC                  chan bool
		proposalCache          protocol.ProposalCache
		chainConf              protocol.ChainConf
		log                    protocol.Logger
		finishProposeC         chan bool
		metricBlockPackageTime *prometheus.HistogramVec
		proposer               *pbac.Member
		blockBuilder           *common.BlockBuilder
		storeHelper            conf.StoreHelper
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockProposerImpl{
				chainId:                tt.fields.chainId,
				txPool:                 tt.fields.txPool,
				txScheduler:            tt.fields.txScheduler,
				snapshotManager:        tt.fields.snapshotManager,
				identity:               tt.fields.identity,
				ledgerCache:            tt.fields.ledgerCache,
				msgBus:                 tt.fields.msgBus,
				ac:                     tt.fields.ac,
				blockchainStore:        tt.fields.blockchainStore,
				idle:                   tt.fields.idle,
				exitC:                  tt.fields.exitC,
				proposalCache:          tt.fields.proposalCache,
				chainConf:              tt.fields.chainConf,
				log:                    tt.fields.log,
				finishProposeC:         tt.fields.finishProposeC,
				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
				proposer:               tt.fields.proposer,
				blockBuilder:           tt.fields.blockBuilder,
				storeHelper:            tt.fields.storeHelper,
			}
			bp.setIdle()
		})
	}
}

/*
 * test unit mock newMockChainConf
 */
func newMockChainConf(t *testing.T) *mock.MockChainConf {
	ctrl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctrl)
	return chainConf
}

/*
 * test unit mock newMockBlockchainStore
 */
func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}

/*
 * test unit mock newMockStoreHelper
 */
func newMockStoreHelper(t *testing.T) *mock.MockStoreHelper {
	ctrl := gomock.NewController(t)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	return storeHelper
}

/*
 * test unit mock newMockLogger
 */
func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

/*
 * test unit mock newMockVmManager
 */
func newMockVmManager(t *testing.T) *mock.MockVmManager {
	ctrl := gomock.NewController(t)
	vmManager := mock.NewMockVmManager(ctrl)
	vmManager.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonpb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonpb.TxStatusCode_SUCCESS).AnyTimes()
	return vmManager
}

/*
 * test unit mock newMockTxPool
 */
func newMockTxPool(t *testing.T) *mock.MockTxPool {
	ctrl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctrl)
	return txPool
}

/*
 * test unit mock newMockSnapshotManager
 */
func newMockSnapshotManager(t *testing.T) *mock.MockSnapshotManager {
	ctrl := gomock.NewController(t)
	snapshotManager := mock.NewMockSnapshotManager(ctrl)
	return snapshotManager
}

/*
 * test unit mock newMockLedgerCache
 */
func newMockLedgerCache(t *testing.T) *mock.MockLedgerCache {
	ctrl := gomock.NewController(t)
	newMockLedgerCache := mock.NewMockLedgerCache(ctrl)
	return newMockLedgerCache
}

/*
 * test unit mock newMockProposalCache
 */
func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
	ctrl := gomock.NewController(t)
	proposalCache := mock.NewMockProposalCache(ctrl)
	return proposalCache
}

/*
 * test unit mock newMockBlockVerifier
 */
func newMockBlockVerifier(t *testing.T) *mock.MockBlockVerifier {
	ctrl := gomock.NewController(t)
	blockVerifier := mock.NewMockBlockVerifier(ctrl)
	return blockVerifier
}

/*
 * test unit mock newMockBlockCommitter
 */
func newMockBlockCommitter(t *testing.T) *mock.MockBlockCommitter {
	ctrl := gomock.NewController(t)
	blockCommitter := mock.NewMockBlockCommitter(ctrl)
	return blockCommitter
}

/*
 * test unit mock newMockTxScheduler
 */
func newMockTxScheduler(t *testing.T) *mock.MockTxScheduler {
	ctrl := gomock.NewController(t)
	txScheduler := mock.NewMockTxScheduler(ctrl)
	return txScheduler
}

/*
 * test unit mock newMockSigningMember
 */
func newMockSigningMember(t *testing.T) *mock.MockSigningMember {
	ctrl := gomock.NewController(t)
	signingMember := mock.NewMockSigningMember(ctrl)
	return signingMember
}

/*
 * test unit mock newMockAccessControlProvider
 */
func newMockAccessControlProvider(t *testing.T) *mock.MockAccessControlProvider {
	ctrl := gomock.NewController(t)
	accessControlProvider := mock.NewMockAccessControlProvider(ctrl)
	return accessControlProvider
}

/*
 * test unit mock newMockBlockProposer
 */
func newMockBlockProposer(t *testing.T) *mock.MockBlockProposer {
	ctrl := gomock.NewController(t)
	blockProposer := mock.NewMockBlockProposer(ctrl)
	return blockProposer
}

/*
 * test unit mock newMockSnapshot
 */
func newMockSnapshot(t *testing.T) *mock.MockSnapshot {
	ctrl := gomock.NewController(t)
	snapshot := mock.NewMockSnapshot(ctrl)
	return snapshot
}

func newMockMessageBus(t *testing.T) *mbusmock.MockMessageBus {
	ctrl := gomock.NewController(t)
	messageBus := mbusmock.NewMockMessageBus(ctrl)
	return messageBus
}

//func TestBlockProposerImpl_startProposingLoop(t *testing.T) {
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
//		txFilter               protocol.TxFilter
//		isProposer             bool
//		idle                   bool
//		proposeTimer           *time.Timer
//		txPoolSignalC          chan *txpoolpb.TxPoolSignal
//		exitC                  chan bool
//		proposalCache          protocol.ProposalCache
//		chainConf              protocol.ChainConf
//		log                    protocol.Logger
//		finishProposeC         chan bool
//		metricBlockPackageTime *prometheus.HistogramVec
//		blockBuilder           *common.BlockBuilder
//		storeHelper            conf.StoreHelper
//	}
//
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         "chain1",
//				txPool:          newMockTxPool(t),
//				txScheduler:     newMockTxScheduler(t),
//				snapshotManager: newMockSnapshotManager(t),
//				ledgerCache:     newMockLedgerCache(t),
//				msgBus: func() msgbus.MessageBus {
//					msgBus := newMockMessageBus(t)
//					msgBus.EXPECT().Publish(msgbus.ProposeBlock, &maxbft.ProposeBlock{IsPropose: true}).AnyTimes()
//					return msgBus
//				}(),
//				ac:              newMockAccessControlProvider(t),
//				blockchainStore: newMockBlockchainStore(t),
//				isProposer:      false,
//				idle:            false,
//				proposeTimer:    time.NewTimer(5 * time.Second),
//				txPoolSignalC:   make(chan *txpoolpb.TxPoolSignal),
//				log:             newMockLogger(t),
//			},
//			wantErr: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			bp := &BlockProposerImpl{
//				chainId:                tt.fields.chainId,
//				txPool:                 tt.fields.txPool,
//				txScheduler:            tt.fields.txScheduler,
//				snapshotManager:        tt.fields.snapshotManager,
//				identity:               tt.fields.identity,
//				ledgerCache:            tt.fields.ledgerCache,
//				msgBus:                 tt.fields.msgBus,
//				ac:                     tt.fields.ac,
//				blockchainStore:        tt.fields.blockchainStore,
//				txFilter:               tt.fields.txFilter,
//				isProposer:             tt.fields.isProposer,
//				idle:                   tt.fields.idle,
//				proposeTimer:           tt.fields.proposeTimer,
//				txPoolSignalC:          tt.fields.txPoolSignalC,
//				exitC:                  tt.fields.exitC,
//				proposalCache:          tt.fields.proposalCache,
//				chainConf:              tt.fields.chainConf,
//				log:                    tt.fields.log,
//				finishProposeC:         tt.fields.finishProposeC,
//				metricBlockPackageTime: tt.fields.metricBlockPackageTime,
//				blockBuilder:           tt.fields.blockBuilder,
//				storeHelper:            tt.fields.storeHelper,
//			}
//			bp.startProposingLoop()
//		})
//	}
//}
