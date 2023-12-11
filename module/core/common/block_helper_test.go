/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"
	"testing"

	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	mock2 "chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/golang/mock/gomock"
)

//
//import (
//	"errors"
//	"fmt"
//	"reflect"
//	"testing"
//	"time"
//
//	"chainmaker.org/chainmaker-go/module/core/cache"
//	"chainmaker.org/chainmaker-go/module/core/provider/conf"
//	"chainmaker.org/chainmaker-go/module/subscriber"
//	"chainmaker.org/chainmaker/common/v2/crypto/hash"
//	"chainmaker.org/chainmaker/common/v2/msgbus"
//	"chainmaker.org/chainmaker/localconf/v2"
//	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
//	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
//	"chainmaker.org/chainmaker/pb-go/v2/config"
//	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
//	"chainmaker.org/chainmaker/pb-go/v2/consensus"
//	"chainmaker.org/chainmaker/protocol/v2"
//	"chainmaker.org/chainmaker/protocol/v2/mock"
//	"chainmaker.org/chainmaker/utils/v2"
//
//	"github.com/golang/mock/gomock"
//	"github.com/golang/protobuf/proto"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/stretchr/testify/require"
//)
//
////  statistic the time consuming of finalizeBlock between sync and async
//// logLevel: Debug TxNum: 1000000; async:3037 ; sync: 4264
//// logLevel: Info  TxNum: 1000000; async:224 ; sync: 251
////func TestFinalizeBlock_Async(t *testing.T) {
////
////	log := logger.GetLogger("core")
////	block := createBlock(10)
////	txs := make([]*commonpb.Transaction, 0)
////	txRWSetMap := make(map[string]*commonpb.TxRWSet)
////	for i := 0; i < 100; i++ {
////		txId := "0x123456789" + fmt.Sprint(i)
////		tx := createNewTestTx(txId)
////		txs = append(txs, tx)
////		txRWSetMap[txId] = &commonpb.TxRWSet{
////			TxId:    txId,
////			TxReads: nil,
////			TxWrites: []*commonpb.TxWrite{{
////				Key:          []byte(fmt.Sprintf("key%d", i)),
////				Value:        []byte(fmt.Sprintf("value[%d]", i)),
////				ContractName: "TestContract",
////			}},
////		}
////	}
////	block.Txs = txs
////	var err error
////
////	asyncTimeStart := CurrentTimeMillisSeconds()
////	err = FinalizeBlockSequence(block, txRWSetMap, nil, "SM3", log)
////	t.Logf("sync mode cost:[%d]", CurrentTimeMillisSeconds()-asyncTimeStart)
////	t.Logf("%x,%x,%x", block.Header.RwSetRoot, block.Header.TxRoot, block.Header.DagHash)
////	rwSetRoot := block.Header.RwSetRoot
////	//blockHash := block.Header.BlockHash
////	dagHash := block.Header.DagHash
////	txRoot := block.Header.TxRoot
////	asyncTimeStart = CurrentTimeMillisSeconds()
////	block.Header.RwSetRoot = nil
////	block.Header.BlockHash = nil
////	block.Header.DagHash = nil
////	block.Header.TxRoot = nil
////	err = FinalizeBlock(block, txRWSetMap, nil, "SM3", log)
////	asyncTimeEnd := CurrentTimeMillisSeconds()
////	require.Equal(t, nil, err)
////	t.Logf("concurrent mode cost:[%d]", asyncTimeEnd-asyncTimeStart)
////	assert.EqualValues(t, rwSetRoot, block.Header.RwSetRoot, "RwSetRoot")
////	//assert.EqualValues(t, blockHash, block.Header.BlockHash, "BlockHash")
////	assert.EqualValues(t, dagHash, block.Header.DagHash, "DagHash")
////	assert.EqualValues(t, txRoot, block.Header.TxRoot, "TxRoot")
////
////	////
////	//syncTimeStart := CurrentTimeMillisSeconds()
////	//err = FinalizeBlock(block, txRWSetMap, nil, "SHA256", log)
////	//syncTimeEnd := CurrentTimeMillisSeconds()
////	//require.Equal(t, nil, err)
////	//t.Log(fmt.Sprintf("sync mode cost:[%d]", syncTimeEnd-syncTimeStart))
////	////
////	//require.Equal(t, rwSetRoot, block.Header.RwSetRoot)
////	//require.Equal(t, blockHash, block.Header.BlockHash)
////	//require.Equal(t, dagHash, block.Header.DagHash)
////	//
////	//log.Infof(fmt.Sprintf("async mode cost:[%d], sync mode cost:[%d]", asyncTimeEnd-asyncTimeStart, syncTimeEnd-syncTimeStart))
////
////}
//
//func TestBlockBuilder_InitNewBlock(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	lastBlock := createBlock(9)
//
//	ledgerCache := mock.NewMockLedgerCache(ctrl)
//	ledgerCache.EXPECT().CurrentHeight().Return(uint64(9), nil).AnyTimes()
//	ledgerCache.EXPECT().GetLastCommittedBlock().Return(lastBlock).AnyTimes()
//
//	identity := mock.NewMockSigningMember(ctrl)
//	identity.EXPECT().GetMember().Return(nil, nil).AnyTimes()
//
//	snapshotManager := mock.NewMockSnapshotManager(ctrl)
//	//var snapshot *protocol.Snapshot
//
//	//storeHelper := mock.NewMockStoreHelper(ctrl)
//	snapshot := mock.NewMockSnapshot(ctrl)
//	snapshot.EXPECT().GetBlockchainStore().AnyTimes()
//	snapshotManager.EXPECT().NewSnapshot(lastBlock, gomock.Any()).Return(snapshot).AnyTimes()
//	//storeHelper.EXPECT().BeginDbTransaction(gomock.Any(), gomock.Any())
//
//	tx1 := createNewTestTx("0x987654321")
//	txBatch := []*commonpb.Transaction{tx1}
//
//	txScheduler := mock.NewMockTxScheduler(ctrl)
//	txScheduler.EXPECT().Schedule(gomock.Any(), txBatch, snapshot).AnyTimes()
//
//	chainConf := mock.NewMockChainConf(ctrl)
//	cf := config.ChainConfig{Consensus: &config.ConsensusConfig{Type: 0}}
//	chainConf.EXPECT().ChainConfig().Return(&cf).AnyTimes()
//
//	//conf := &BlockBuilderConf{
//	//	ChainId:         "chain1",
//	//	TxPool:          nil,
//	//	TxScheduler:     txScheduler,
//	//	SnapshotManager: snapshotManager,
//	//	Identity:        identity,
//	//	LedgerCache:     ledgerCache,
//	//	ProposalCache:   nil,
//	//	ChainConf:       chainConf,
//	//	Log:             nil,
//	//	StoreHelper:     storeHelper,
//	//}
//
//	block, err := initNewBlock(lastBlock, identity, "chain1", chainConf, false)
//	require.Nil(t, err)
//	require.NotNil(t, block)
//
//}
//
func createBlock(height uint64) *commonpb.Block {
	var hash = []byte("0123456789")
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

func createNewTestTx(txID string) *commonpb.Transaction {
	//var hash = []byte("0123456789")
	return &commonpb.Transaction{
		Payload: &commonpb.Payload{
			ChainId:        "Chain1",
			TxType:         0,
			TxId:           txID,
			Timestamp:      utils.CurrentTimeMillisSeconds(),
			ExpirationTime: 0,
		},
		Result: &commonpb.Result{
			Code:           commonpb.TxStatusCode_SUCCESS,
			ContractResult: nil,
			RwSetHash:      nil,
		},
	}
}

//
//// the sync way fo finalize block
//func FinalizeBlockSync(
//	block *commonpb.Block,
//	txRWSetMap map[string]*commonpb.TxRWSet,
//	aclFailTxs []*commonpb.Transaction,
//	hashType string,
//	logger protocol.Logger) error {
//
//	if aclFailTxs != nil && len(aclFailTxs) > 0 { //nolint: gosimple
//		// append acl check failed txs to the end of block.Txs
//		block.Txs = append(block.Txs, aclFailTxs...)
//	}
//
//	// TxCount contains acl verify failed txs and invoked contract txs
//	txCount := len(block.Txs)
//	block.Header.TxCount = uint32(txCount)
//
//	// TxRoot/RwSetRoot
//	var err error
//	txHashes := make([][]byte, txCount)
//	for i, tx := range block.Txs {
//		// finalize tx, put rwsethash into tx.Result
//		rwSet := txRWSetMap[tx.Payload.TxId]
//		if rwSet == nil {
//			rwSet = &commonpb.TxRWSet{
//				TxId:     tx.Payload.TxId,
//				TxReads:  nil,
//				TxWrites: nil,
//			}
//		}
//		var rwSetHash []byte
//		rwSetHash, err = utils.CalcRWSetHash(hashType, rwSet)
//		logger.DebugDynamic(func() string {
//			return fmt.Sprintf("CalcRWSetHash rwset: %+v ,hash: %x", rwSet, rwSetHash)
//		})
//		if err != nil {
//			return err
//		}
//		if tx.Result == nil {
//			// in case tx.Result is nil, avoid panic
//			e := fmt.Errorf("tx(%s) result == nil", tx.Payload.TxId)
//			logger.Error(e.Error())
//			return e
//		}
//		tx.Result.RwSetHash = rwSetHash
//		// calculate complete tx hash, include tx.Header, tx.Payload, tx.Result
//		var txHash []byte
//		txHash, err = utils.CalcTxHash(hashType, tx)
//		if err != nil {
//			return err
//		}
//		txHashes[i] = txHash
//	}
//
//	block.Header.TxRoot, err = hash.GetMerkleRoot(hashType, txHashes)
//	if err != nil {
//		logger.Warnf("get tx merkle root error %s", err)
//		return err
//	}
//	block.Header.RwSetRoot, err = utils.CalcRWSetRoot(hashType, block.Txs)
//	if err != nil {
//		logger.Warnf("get rwset merkle root error %s", err)
//		return err
//	}
//
//	// DagDigest
//	dagHash, err := utils.CalcDagHash(hashType, block.Dag)
//	if err != nil {
//		logger.Warnf("get dag hash error %s", err)
//		return err
//	}
//	block.Header.DagHash = dagHash
//
//	return nil
//}
//
//func TestBlockBuilder_GenerateNewBlock(t *testing.T) {
//	type fields struct {
//		chainId         string
//		txPool          protocol.TxPool
//		txScheduler     protocol.TxScheduler
//		snapshotManager protocol.SnapshotManager
//		identity        protocol.SigningMember
//		ledgerCache     protocol.LedgerCache
//		proposalCache   protocol.ProposalCache
//		chainConf       protocol.ChainConf
//		log             protocol.Logger
//		storeHelper     conf.StoreHelper
//	}
//	type args struct {
//		proposingHeight uint64
//		preHash         []byte
//		txBatch         []*commonpb.Transaction
//	}
//
//	var (
//		chainId         = "123456"
//		txPool          = newMockTxPool(t)
//		txScheduler     = newMockTxScheduler(t)
//		snapshotManager = newMockSnapshotManager(t)
//		identity        = newMockSigningMember(t)
//		ledgerCache     = newMockLedgerCache(t)
//		proposalCache   = newMockProposalCache(t)
//		chainConf       = newMockChainConf(t)
//		log             = newMockLogger(t)
//		storeHelper     = newMockStoreHelper(t)
//		snapshot        = newMockSnapshot(t)
//	)
//
//	ledgerCache.EXPECT().CurrentHeight().AnyTimes()
//
//	block := createBlock(0)
//	ledgerCache.EXPECT().GetLastCommittedBlock().Return(block).AnyTimes()
//	proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).AnyTimes()
//	identity.EXPECT().GetMember().AnyTimes()
//	snapshot.EXPECT().GetBlockchainStore().AnyTimes()
//	snapshotManager.EXPECT().NewSnapshot(gomock.Any(), gomock.Any()).Return(snapshot).AnyTimes()
//	storeHelper.EXPECT().BeginDbTransaction(gomock.Any(), gomock.Any()).AnyTimes()
//	txScheduler.EXPECT().Schedule(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
//
//	//timeNow := time.Now().Unix()
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *commonpb.Block
//		want1   []int64
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         chainId,
//				txPool:          txPool,
//				txScheduler:     txScheduler,
//				snapshotManager: snapshotManager,
//				identity:        identity,
//				ledgerCache:     ledgerCache,
//				proposalCache:   proposalCache,
//				chainConf:       chainConf,
//				log:             log,
//				storeHelper:     storeHelper,
//			},
//			args: args{
//				proposingHeight: 0,
//				preHash:         []byte("123456"),
//				txBatch:         nil,
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId:         chainId,
//				txPool:          txPool,
//				txScheduler:     txScheduler,
//				snapshotManager: snapshotManager,
//				identity:        identity,
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache = newMockLedgerCache(t)
//					ledgerCache.EXPECT().CurrentHeight().AnyTimes()
//					ledgerCache.EXPECT().GetLastCommittedBlock().AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache: proposalCache,
//				chainConf:     chainConf,
//				log:           log,
//				storeHelper:   storeHelper,
//			},
//			args: args{
//				proposingHeight: 1,
//				preHash:         block.Hash(),
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:         chainId,
//				txPool:          txPool,
//				txScheduler:     txScheduler,
//				snapshotManager: snapshotManager,
//				identity:        identity,
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache = newMockLedgerCache(t)
//
//					block = createBlock(1)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(block).AnyTimes()
//					ledgerCache.EXPECT().CurrentHeight().Return(uint64(2), nil).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache = newMockProposalCache(t)
//
//					proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).Return(createBlock(2), nil).AnyTimes()
//					return proposalCache
//				}(),
//				chainConf: func() protocol.ChainConf {
//					var chainConfig = &configpb.ChainConfig{
//						Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensus.ConsensusType_TBFT,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//				log:         log,
//				storeHelper: storeHelper,
//			},
//			args: args{
//				proposingHeight: 1,
//				preHash:         block.Hash(),
//			},
//			want:  nil,
//			want1: []int64{0, 0, 0},
//
//			wantErr: true,
//		},
//		//{
//		//	name: "test3",
//		//	fields: fields{
//		//		chainId: chainId,
//		//		txPool:  txPool,
//		//		txScheduler: func() protocol.TxScheduler {
//		//
//		//			//block := &commonpb.Block{
//		//			//	Header: &commonpb.BlockHeader{
//		//			//		BlockVersion: uint32(220),
//		//			//		ChainId: "123456",
//		//			//		BlockHeight: 3,
//		//			//		PreBlockHash: []byte("0123456789"),
//		//			//		BlockTimestamp: timeNow,
//		//			//	},
//		//			//	Dag: &commonpb.DAG{
//		//			//		Vertexes: nil,
//		//			//	},
//		//			//}
//		//			//
//		//			//txBatch := []*commonpb.Transaction{
//		//			//	{
//		//			//		Payload: &commonpb.Payload{
//		//			//			ChainId: "chain1",
//		//			//		},
//		//			//	},
//		//			//}
//		//			//block.Txs = txBatch
//		//			txScheduler := newMockTxScheduler(t)
//		//			txScheduler.EXPECT().Schedule(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
//		//			//txScheduler.EXPECT().Schedule(block, txBatch, snapshot).Return(map[string]*commonpb.TxRWSet{
//		//			//	"123456": {
//		//			//		TxId: "123456",
//		//			//		TxReads: []*commonpb.TxRead{
//		//			//
//		//			//		},
//		//			//		TxWrites: []*commonpb.TxWrite{
//		//			//
//		//			//		},
//		//			//	},
//		//			//}, map[string][]*commonpb.ContractEvent{
//		//			//
//		//			//}, nil).AnyTimes()
//		//
//		//			return txScheduler
//		//		}(),
//		//		snapshotManager: snapshotManager,
//		//		identity:        identity,
//		//		ledgerCache: func() protocol.LedgerCache {
//		//			ledgerCache = newMockLedgerCache(t)
//		//
//		//			block = createBlock(1)
//		//
//		//			block.Txs = []*commonpb.Transaction{
//		//				{
//		//					Payload: &commonpb.Payload{
//		//						ChainId: "123456",
//		//					},
//		//				},
//		//			}
//		//			block.Header.BlockTimestamp = timeNow
//		//			ledgerCache.EXPECT().GetLastCommittedBlock().Return(block).AnyTimes()
//		//			ledgerCache.EXPECT().CurrentHeight().Return(uint64(2), nil).AnyTimes()
//		//			return ledgerCache
//		//		}(),
//		//		proposalCache: func() protocol.ProposalCache {
//		//			proposalCache = newMockProposalCache(t)
//		//
//		//			lastBlock := createBlock(2)
//		//			lastBlock.Txs = []*commonpb.Transaction{
//		//				{
//		//					Payload: &commonpb.Payload{
//		//						ChainId: "123456",
//		//					},
//		//				},
//		//			}
//		//			lastBlock.Dag = &commonpb.DAG{
//		//				Vertexes: nil,
//		//			}
//		//
//		//			proposalCache.EXPECT().GetProposedBlockByHashAndHeight(lastBlock.Header.PreBlockHash, uint64(0)).Return(lastBlock, nil).AnyTimes()
//		//			return proposalCache
//		//		}(),
//		//		chainConf: func() protocol.ChainConf {
//		//			chainConf = newMockChainConf(t)
//		//			var chainConfig = &configpb.ChainConfig{
//		//				Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
//		//				Consensus: &configpb.ConsensusConfig{
//		//					Type: consensus.ConsensusType_TBFT,
//		//				},
//		//			}
//		//			chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//		//			return chainConf
//		//		}(),
//		//		log:         log,
//		//		storeHelper: storeHelper,
//		//	},
//		//	args: args{
//		//		proposingHeight: 1,
//		//		preHash:         block.Hash(),
//		//		txBatch: []*commonpb.Transaction{
//		//			{
//		//				Payload: &commonpb.Payload{
//		//					ChainId: "chain1",
//		//				},
//		//			},
//		//		},
//		//	},
//		//	want:    nil,
//		//	want1:   []int64{0, 0, 0},
//		//	wantErr: true,
//		//},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			bb := &BlockBuilder{
//				chainId:         tt.fields.chainId,
//				txPool:          tt.fields.txPool,
//				txScheduler:     tt.fields.txScheduler,
//				snapshotManager: tt.fields.snapshotManager,
//				identity:        tt.fields.identity,
//				ledgerCache:     tt.fields.ledgerCache,
//				proposalCache:   tt.fields.proposalCache,
//				chainConf:       tt.fields.chainConf,
//				log:             tt.fields.log,
//				storeHelper:     tt.fields.storeHelper,
//			}
//			got, got1, err := bb.GenerateNewBlock(tt.args.proposingHeight, tt.args.preHash, tt.args.txBatch, []string{"abc"})
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GenerateNewBlock() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GenerateNewBlock() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("GenerateNewBlock() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}
//
//func TestBlockBuilder_findLastBlockFromCache(t *testing.T) {
//	type fields struct {
//		ledgerCache   protocol.LedgerCache
//		proposalCache protocol.ProposalCache
//	}
//	type args struct {
//		proposingHeight uint64
//		preHash         []byte
//		currentHeight   uint64
//	}
//
//	ledgerCache := newMockLedgerCache(t)
//	proposalCache := newMockProposalCache(t)
//	block := createBlock(0)
//
//	proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).AnyTimes()
//	ledgerCache.EXPECT().GetLastCommittedBlock().Return(block).AnyTimes()
//
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   *commonpb.Block
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				ledgerCache:   ledgerCache,
//				proposalCache: proposalCache,
//			},
//			args: args{
//				proposingHeight: 0,
//				preHash:         []byte("123456"),
//				currentHeight:   1,
//			},
//			want: nil,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				ledgerCache:   ledgerCache,
//				proposalCache: proposalCache,
//			},
//			args: args{
//				proposingHeight: 1,
//				preHash:         block.Hash(),
//				currentHeight:   0,
//			},
//			want: block,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			bb := &BlockBuilder{
//				ledgerCache:   tt.fields.ledgerCache,
//				proposalCache: tt.fields.proposalCache,
//			}
//			if got := bb.findLastBlockFromCache(tt.args.proposingHeight, tt.args.preHash, tt.args.currentHeight); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("findLastBlockFromCache() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestNewBlockCommitter(t *testing.T) {
//	type args struct {
//		config BlockCommitterConfig
//		log    protocol.Logger
//	}
//
//	var (
//		chainId         = "123456"
//		blockchainStore = newMockBlockchainStore(t)
//		snapshotManager = newMockSnapshotManager(t)
//		txPool          = newMockTxPool(t)
//		ledgerCache     = newMockLedgerCache(t)
//		proposedCache   = newMockProposalCache(t)
//		chainConf       = newMockChainConf(t)
//		msgBus          = msgbus.NewMessageBus()
//		//subscriber      = nil
//		verifier    = newMockBlockVerifier(t)
//		storeHelper = newMockStoreHelper(t)
//		log         = newMockLogger(t)
//	)
//
//	tests := []struct {
//		name    string
//		args    args
//		want    protocol.BlockCommitter
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			args: args{
//				config: BlockCommitterConfig{
//					ChainId:         chainId,
//					BlockchainStore: blockchainStore,
//					SnapshotManager: snapshotManager,
//					TxPool:          txPool,
//					LedgerCache:     ledgerCache,
//					ProposedCache:   proposedCache,
//					ChainConf:       chainConf,
//					MsgBus:          msgBus,
//					Subscriber:      nil,
//					Verifier:        verifier,
//					StoreHelper:     storeHelper,
//				},
//				log: log,
//			},
//			want: func() protocol.BlockCommitter {
//
//				blockchain := &BlockCommitterImpl{
//					chainId:         chainId,
//					blockchainStore: blockchainStore,
//					snapshotManager: snapshotManager,
//					txPool:          txPool,
//					ledgerCache:     ledgerCache,
//					proposalCache:   proposedCache,
//					log:             log,
//					chainConf:       chainConf,
//					msgBus:          msgBus,
//					subscriber:      nil,
//					verifier:        verifier,
//					storeHelper:     storeHelper,
//				}
//
//				cbConf := &CommitBlockConf{
//					Store:                 blockchain.blockchainStore,
//					Log:                   blockchain.log,
//					SnapshotManager:       blockchain.snapshotManager,
//					TxPool:                blockchain.txPool,
//					LedgerCache:           blockchain.ledgerCache,
//					ChainConf:             blockchain.chainConf,
//					MsgBus:                blockchain.msgBus,
//					MetricBlockCommitTime: blockchain.metricBlockCommitTime,
//					MetricBlockCounter:    blockchain.metricBlockCounter,
//					MetricBlockSize:       blockchain.metricBlockSize,
//					MetricTxCounter:       blockchain.metricTxCounter,
//				}
//
//				blockchain.commonCommit = NewCommitBlock(cbConf)
//				return blockchain
//			}(),
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := NewBlockCommitter(tt.args.config, tt.args.log)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("NewBlockCommitter() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("NewBlockCommitter() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestBlockCommitterImpl_AddBlock(t *testing.T) {
//	type fields struct {
//		chainId               string
//		blockchainStore       protocol.BlockchainStore
//		snapshotManager       protocol.SnapshotManager
//		txPool                protocol.TxPool
//		chainConf             protocol.ChainConf
//		ledgerCache           protocol.LedgerCache
//		proposalCache         protocol.ProposalCache
//		log                   protocol.Logger
//		msgBus                msgbus.MessageBus
//		subscriber            *subscriber.EventSubscriber
//		verifier              protocol.BlockVerifier
//		commonCommit          *CommitBlock
//		metricBlockSize       *prometheus.HistogramVec
//		metricBlockCounter    *prometheus.CounterVec
//		metricTxCounter       *prometheus.CounterVec
//		metricBlockCommitTime *prometheus.HistogramVec
//		storeHelper           conf.StoreHelper
//		blockInterval         int64
//	}
//	type args struct {
//		block *commonpb.Block
//	}
//
//	var (
//		chainId         = "123456"
//		blockchainStore = newMockBlockchainStore(t)
//		snapshotManager = newMockSnapshotManager(t)
//		txPool          = newMockTxPool(t)
//		ledgerCache     = newMockLedgerCache(t)
//		proposedCache   = newMockProposalCache(t)
//		chainConf       = newMockChainConf(t)
//		msgBus          = msgbus.NewMessageBus()
//		//subscriber      = nil
//		verifier    = newMockBlockVerifier(t)
//		storeHelper = newMockStoreHelper(t)
//		log         = newMockLogger(t)
//	)
//
//	storeHelper.EXPECT().RollBack(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
//	blockchainStore.EXPECT().PutBlock(gomock.Any(), gomock.Any()).AnyTimes()
//	snapshotManager.EXPECT().NotifyBlockCommitted(gomock.Any()).AnyTimes()
//	verifier.EXPECT().VerifyBlock(gomock.Any(), gomock.Any()).AnyTimes()
//
//	blockHash := []byte("123456")
//	commonBlock := createBlockByHash(1, blockHash)
//	lastBlock := createBlockByHash(1, blockHash)
//
//	ledgerCache.EXPECT().GetLastCommittedBlock().Return(commonBlock).AnyTimes()
//	ledgerCache.EXPECT().SetLastCommittedBlock(gomock.Any()).AnyTimes()
//	ledgerCache.EXPECT().CurrentHeight().AnyTimes()
//	txPool.EXPECT().RetryAndRemoveTxs(gomock.Any(), gomock.Any()).AnyTimes()
//
//	chainConfig := &configpb.ChainConfig{
//		Core: &configpb.CoreConfig{
//			ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//				ConsensusMessageTurbo: true,
//			},
//		},
//		Crypto: &configpb.CryptoConfig{
//			Hash: "SHA256",
//		},
//	}
//	chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//
//	log.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
//	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
//	log.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
//	log.EXPECT().Warn(gomock.Any()).AnyTimes()
//
//	cbConf := &CommitBlockConf{
//		Store:           blockchainStore,
//		Log:             log,
//		SnapshotManager: snapshotManager,
//		TxPool:          txPool,
//		LedgerCache:     ledgerCache,
//		ChainConf:       chainConf,
//		MsgBus:          msgBus,
//	}
//
//	committer := NewCommitBlock(cbConf)
//
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache:     ledgerCache,
//				proposalCache:   proposedCache,
//				log:             log,
//				msgBus:          msgBus,
//				verifier:        verifier,
//				storeHelper:     storeHelper,
//			},
//			wantErr: true,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache:     ledgerCache,
//				proposalCache:   proposedCache,
//				log:             log,
//				msgBus:          msgBus,
//				verifier:        verifier,
//				storeHelper:     storeHelper,
//			},
//			args: args{
//				block: commonBlock,
//			},
//			wantErr: true,
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache:     ledgerCache,
//				proposalCache:   proposedCache,
//				log:             log,
//				msgBus:          msgBus,
//				verifier:        verifier,
//				storeHelper:     storeHelper,
//			},
//			args: args{
//				block: createBlock(2),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test3",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache:     ledgerCache,
//				proposalCache:   proposedCache,
//				log:             log,
//				msgBus:          msgBus,
//				verifier:        verifier,
//				storeHelper:     storeHelper,
//			},
//			args: args{
//				block: createBlock(1),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test4",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache = newMockLedgerCache(t)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(lastBlock).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := cache.NewProposalCache(chainConf, ledgerCache)
//					return proposalCache
//				}(),
//				log:          log,
//				msgBus:       msgBus,
//				verifier:     verifier,
//				commonCommit: committer,
//				storeHelper:  storeHelper,
//			},
//			args: args{
//				block: func() *commonpb.Block {
//					blockHash := []byte("123456")
//					block := createBlockByHash(2, blockHash)
//					header := *block.Header
//					header.Signature = nil
//					header.BlockHash = nil
//					blockHash, _ = proto.Marshal(&header)
//					hash, _ := hash.GetByStrType("SHA256", blockHash)
//					block.Header.BlockHash = hash
//					return block
//				}(),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test5",
//			fields: fields{
//				chainId:         chainId,
//				blockchainStore: blockchainStore,
//				snapshotManager: snapshotManager,
//				txPool:          txPool,
//				chainConf:       chainConf,
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache = newMockLedgerCache(t)
//					ledgerCache.EXPECT().CurrentHeight().AnyTimes()
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(lastBlock).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := cache.NewProposalCache(chainConf, ledgerCache)
//					rwSetMap := make(map[string]*commonpb.TxRWSet)
//					contractEvenMap := make(map[string][]*commonpb.ContractEvent)
//					proposalCache.SetProposedBlock(createBlockByHash(2, []byte("123456")), rwSetMap, contractEvenMap, false)
//					return proposalCache
//				}(),
//				log:          log,
//				msgBus:       msgBus,
//				verifier:     verifier,
//				commonCommit: committer,
//				storeHelper:  storeHelper,
//			},
//			args: args{
//				block: func() *commonpb.Block {
//					blockHash := []byte("123456")
//					block := createBlockByHash(2, blockHash)
//					header := *block.Header
//					header.Signature = nil
//					header.BlockHash = nil
//
//					blockHash, _ = proto.Marshal(&header)
//					hash, _ := hash.GetByStrType("SHA256", blockHash)
//
//					block.Header.BlockHash = hash
//					return block
//				}(),
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			chain := &BlockCommitterImpl{
//				chainId:               tt.fields.chainId,
//				blockchainStore:       tt.fields.blockchainStore,
//				snapshotManager:       tt.fields.snapshotManager,
//				txPool:                tt.fields.txPool,
//				chainConf:             tt.fields.chainConf,
//				ledgerCache:           tt.fields.ledgerCache,
//				proposalCache:         tt.fields.proposalCache,
//				log:                   tt.fields.log,
//				msgBus:                tt.fields.msgBus,
//				subscriber:            tt.fields.subscriber,
//				verifier:              tt.fields.verifier,
//				commonCommit:          tt.fields.commonCommit,
//				metricBlockSize:       tt.fields.metricBlockSize,
//				metricBlockCounter:    tt.fields.metricBlockCounter,
//				metricTxCounter:       tt.fields.metricTxCounter,
//				metricBlockCommitTime: tt.fields.metricBlockCommitTime,
//				storeHelper:           tt.fields.storeHelper,
//				blockInterval:         tt.fields.blockInterval,
//			}
//			if err := chain.AddBlock(tt.args.block); (err != nil) != tt.wantErr {
//				t.Errorf("AddBlock() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestBlockCommitterImpl_isBlockLegal(t *testing.T) {
//	type fields struct {
//		chainId               string
//		blockchainStore       protocol.BlockchainStore
//		snapshotManager       protocol.SnapshotManager
//		txPool                protocol.TxPool
//		chainConf             protocol.ChainConf
//		ledgerCache           protocol.LedgerCache
//		proposalCache         protocol.ProposalCache
//		log                   protocol.Logger
//		msgBus                msgbus.MessageBus
//		subscriber            *subscriber.EventSubscriber
//		verifier              protocol.BlockVerifier
//		commonCommit          *CommitBlock
//		metricBlockSize       *prometheus.HistogramVec
//		metricBlockCounter    *prometheus.CounterVec
//		metricTxCounter       *prometheus.CounterVec
//		metricBlockCommitTime *prometheus.HistogramVec
//		storeHelper           conf.StoreHelper
//		blockInterval         int64
//	}
//	type args struct {
//		blk *commonpb.Block
//	}
//
//	blockHash := []byte("123456")
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         "123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache := newMockLedgerCache(t)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(nil).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache:         newMockProposalCache(t),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				blk: createBlock(0),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId:         "123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache := newMockLedgerCache(t)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(createBlock(0)).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache:         newMockProposalCache(t),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				blk: createBlock(0),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:         "123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf: func() protocol.ChainConf {
//
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Crypto: &configpb.CryptoConfig{
//							Hash: "SHA256",
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache := newMockLedgerCache(t)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(createBlock(1)).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache:         newMockProposalCache(t),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				blk: createBlock(2),
//			},
//			wantErr: true,
//		},
//		{
//			name: "test3",
//			fields: fields{
//				chainId:         "123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf: func() protocol.ChainConf {
//
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Crypto: &configpb.CryptoConfig{
//							Hash: "SHA256",
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//				ledgerCache: func() protocol.LedgerCache {
//					ledgerCache := newMockLedgerCache(t)
//					lastBlock := createBlockByHash(1, blockHash)
//					ledgerCache.EXPECT().GetLastCommittedBlock().Return(lastBlock).AnyTimes()
//					return ledgerCache
//				}(),
//				proposalCache:         newMockProposalCache(t),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				blk: func() *commonpb.Block {
//					block := createBlockByHash(2, blockHash)
//					header := *block.Header
//					header.Signature = nil
//					header.BlockHash = nil
//
//					blockHash, _ := proto.Marshal(&header)
//					hash, _ := hash.GetByStrType("SHA256", blockHash)
//					block.Header.BlockHash = hash
//					return block
//				}(),
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			chain := &BlockCommitterImpl{
//				chainId:               tt.fields.chainId,
//				blockchainStore:       tt.fields.blockchainStore,
//				snapshotManager:       tt.fields.snapshotManager,
//				txPool:                tt.fields.txPool,
//				chainConf:             tt.fields.chainConf,
//				ledgerCache:           tt.fields.ledgerCache,
//				proposalCache:         tt.fields.proposalCache,
//				log:                   tt.fields.log,
//				msgBus:                tt.fields.msgBus,
//				subscriber:            tt.fields.subscriber,
//				verifier:              tt.fields.verifier,
//				commonCommit:          tt.fields.commonCommit,
//				metricBlockSize:       tt.fields.metricBlockSize,
//				metricBlockCounter:    tt.fields.metricBlockCounter,
//				metricTxCounter:       tt.fields.metricTxCounter,
//				metricBlockCommitTime: tt.fields.metricBlockCommitTime,
//				storeHelper:           tt.fields.storeHelper,
//				blockInterval:         tt.fields.blockInterval,
//			}
//			if err := chain.isBlockLegal(tt.args.blk); (err != nil) != tt.wantErr {
//				t.Errorf("isBlockLegal() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestBlockCommitterImpl_syncWithTxPool(t *testing.T) {
//	type fields struct {
//		chainId               string
//		blockchainStore       protocol.BlockchainStore
//		snapshotManager       protocol.SnapshotManager
//		txPool                protocol.TxPool
//		chainConf             protocol.ChainConf
//		ledgerCache           protocol.LedgerCache
//		proposalCache         protocol.ProposalCache
//		log                   protocol.Logger
//		msgBus                msgbus.MessageBus
//		subscriber            *subscriber.EventSubscriber
//		verifier              protocol.BlockVerifier
//		commonCommit          *CommitBlock
//		metricBlockSize       *prometheus.HistogramVec
//		metricBlockCounter    *prometheus.CounterVec
//		metricTxCounter       *prometheus.CounterVec
//		metricBlockCommitTime *prometheus.HistogramVec
//		storeHelper           conf.StoreHelper
//		blockInterval         int64
//	}
//	type args struct {
//		block  *commonpb.Block
//		height uint64
//	}
//
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   []*commonpb.Transaction
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         "test0",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					proposalCache.EXPECT().GetProposedBlocksAt(gomock.Any()).Return([]*commonpb.Block{}).AnyTimes()
//					return proposalCache
//				}(),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				block: func() *commonpb.Block {
//					block := createBlock(0)
//					block.Txs = make([]*commonpb.Transaction, 0)
//					return block
//				}(),
//				height: 0,
//			},
//			want: []*commonpb.Transaction{},
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId:         "test1",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					proposalCache.EXPECT().GetProposedBlocksAt(uint64(0)).Return([]*commonpb.Block{
//						createBlock(0),
//					}).AnyTimes()
//					return proposalCache
//				}(),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				block: func() *commonpb.Block {
//					block := createBlock(0)
//					block.Txs = make([]*commonpb.Transaction, 0)
//					return block
//				}(),
//				height: 0,
//			},
//			want: make([]*commonpb.Transaction, 0),
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:         "test2",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					block := createBlock(0)
//					block.Header.BlockHash = []byte("6789")
//					block.Txs = []*commonpb.Transaction{
//						{
//							Payload: &commonpb.Payload{
//								TxId: "1234567890",
//							},
//						},
//					}
//					proposalCache.EXPECT().GetProposedBlocksAt(uint64(0)).Return([]*commonpb.Block{
//						block,
//					}).AnyTimes()
//					return proposalCache
//				}(),
//				log:                   newMockLogger(t),
//				msgBus:                msgbus.NewMessageBus(),
//				subscriber:            nil,
//				verifier:              newMockBlockVerifier(t),
//				commonCommit:          nil,
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           newMockStoreHelper(t),
//				blockInterval:         0,
//			},
//			args: args{
//				block: func() *commonpb.Block {
//					block := createBlock(0)
//					block.Txs = make([]*commonpb.Transaction, 0)
//					return block
//				}(),
//				height: 0,
//			},
//			want: []*commonpb.Transaction{
//				{
//					Payload: &commonpb.Payload{
//						TxId: "1234567890",
//					},
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			chain := &BlockCommitterImpl{
//				chainId:               tt.fields.chainId,
//				blockchainStore:       tt.fields.blockchainStore,
//				snapshotManager:       tt.fields.snapshotManager,
//				txPool:                tt.fields.txPool,
//				chainConf:             tt.fields.chainConf,
//				ledgerCache:           tt.fields.ledgerCache,
//				proposalCache:         tt.fields.proposalCache,
//				log:                   tt.fields.log,
//				msgBus:                tt.fields.msgBus,
//				subscriber:            tt.fields.subscriber,
//				verifier:              tt.fields.verifier,
//				commonCommit:          tt.fields.commonCommit,
//				metricBlockSize:       tt.fields.metricBlockSize,
//				metricBlockCounter:    tt.fields.metricBlockCounter,
//				metricTxCounter:       tt.fields.metricTxCounter,
//				metricBlockCommitTime: tt.fields.metricBlockCommitTime,
//				storeHelper:           tt.fields.storeHelper,
//				blockInterval:         tt.fields.blockInterval,
//			}
//			if got, _ := chain.syncWithTxPool(tt.args.block, tt.args.height); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("syncWithTxPool() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestBlockCommitterImpl_checkLastProposedBlock(t *testing.T) {
//	type fields struct {
//		chainId               string
//		blockchainStore       protocol.BlockchainStore
//		snapshotManager       protocol.SnapshotManager
//		txPool                protocol.TxPool
//		chainConf             protocol.ChainConf
//		ledgerCache           protocol.LedgerCache
//		proposalCache         protocol.ProposalCache
//		log                   protocol.Logger
//		msgBus                msgbus.MessageBus
//		subscriber            *subscriber.EventSubscriber
//		verifier              protocol.BlockVerifier
//		commonCommit          *CommitBlock
//		metricBlockSize       *prometheus.HistogramVec
//		metricBlockCounter    *prometheus.CounterVec
//		metricTxCounter       *prometheus.CounterVec
//		metricBlockCommitTime *prometheus.HistogramVec
//		storeHelper           conf.StoreHelper
//		blockInterval         int64
//	}
//	type args struct {
//		block *commonpb.Block
//	}
//
//	block := createBlock(0)
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *commonpb.Block
//		want1   map[string]*commonpb.TxRWSet
//		want2   map[string][]*commonpb.ContractEvent
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				chainId:         "test123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					proposalCache.EXPECT().GetProposedBlock(gomock.Any()).Return(block, map[string]*commonpb.TxRWSet{
//						"test": {
//							TxId: "123456",
//						},
//					}, map[string][]*commonpb.ContractEvent{
//						"test": {
//							&commonpb.ContractEvent{
//								TxId: "123456",
//							},
//						},
//					}).AnyTimes()
//					return proposalCache
//				}(),
//				log:        newMockLogger(t),
//				msgBus:     msgbus.NewMessageBus(),
//				subscriber: nil,
//				verifier: func() protocol.BlockVerifier {
//					verifier := newMockBlockVerifier(t)
//					verifier.EXPECT().VerifyBlock(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
//					return verifier
//				}(),
//				commonCommit: &CommitBlock{
//					log: newMockLogger(t),
//				},
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           nil,
//				blockInterval:         0,
//			},
//			args: args{
//				block: block,
//			},
//			want: block,
//			want1: map[string]*commonpb.TxRWSet{
//				"test": {
//					TxId: "123456",
//				},
//			},
//			want2: map[string][]*commonpb.ContractEvent{
//				"test": {
//					&commonpb.ContractEvent{
//						TxId: "123456",
//					},
//				},
//			},
//			wantErr: false,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId:         "test123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					proposalCache.EXPECT().GetProposedBlock(gomock.Any()).Return(block, map[string]*commonpb.TxRWSet{
//						"test": {
//							TxId: "123456",
//						},
//					}, map[string][]*commonpb.ContractEvent{
//						"test": {
//							&commonpb.ContractEvent{
//								TxId: "123456",
//							},
//						},
//					}).AnyTimes()
//					return proposalCache
//				}(),
//				log:        newMockLogger(t),
//				msgBus:     msgbus.NewMessageBus(),
//				subscriber: nil,
//				verifier: func() protocol.BlockVerifier {
//					verifier := newMockBlockVerifier(t)
//					verifier.EXPECT().VerifyBlock(gomock.Any(), gomock.Any()).Return(errors.New("verify block error")).AnyTimes()
//					return verifier
//				}(),
//				commonCommit: &CommitBlock{
//					log: newMockLogger(t),
//				},
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           nil,
//				blockInterval:         0,
//			},
//			args: args{
//				block: block,
//			},
//			want:    nil,
//			want1:   nil,
//			want2:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:         "test123456",
//				blockchainStore: newMockBlockchainStore(t),
//				snapshotManager: newMockSnapshotManager(t),
//				txPool:          newMockTxPool(t),
//				chainConf:       newMockChainConf(t),
//				ledgerCache:     newMockLedgerCache(t),
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache := newMockProposalCache(t)
//					proposalCache.EXPECT().GetProposedBlock(gomock.Any()).Return(nil, map[string]*commonpb.TxRWSet{
//						"test": {
//							TxId: "123456",
//						},
//					}, map[string][]*commonpb.ContractEvent{
//						"test": {
//							&commonpb.ContractEvent{
//								TxId: "123456",
//							},
//						},
//					}).AnyTimes()
//					return proposalCache
//				}(),
//				log:        newMockLogger(t),
//				msgBus:     msgbus.NewMessageBus(),
//				subscriber: nil,
//				verifier: func() protocol.BlockVerifier {
//					verifier := newMockBlockVerifier(t)
//					verifier.EXPECT().VerifyBlock(gomock.Any(), gomock.Any()).Return(errors.New("verify block error")).AnyTimes()
//					return verifier
//				}(),
//				commonCommit: &CommitBlock{
//					log: newMockLogger(t),
//				},
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//				storeHelper:           nil,
//				blockInterval:         0,
//			},
//			args: args{
//				block: block,
//			},
//			want:    nil,
//			want1:   nil,
//			want2:   nil,
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			chain := &BlockCommitterImpl{
//				chainId:               tt.fields.chainId,
//				blockchainStore:       tt.fields.blockchainStore,
//				snapshotManager:       tt.fields.snapshotManager,
//				txPool:                tt.fields.txPool,
//				chainConf:             tt.fields.chainConf,
//				ledgerCache:           tt.fields.ledgerCache,
//				proposalCache:         tt.fields.proposalCache,
//				log:                   tt.fields.log,
//				msgBus:                tt.fields.msgBus,
//				subscriber:            tt.fields.subscriber,
//				verifier:              tt.fields.verifier,
//				commonCommit:          tt.fields.commonCommit,
//				metricBlockSize:       tt.fields.metricBlockSize,
//				metricBlockCounter:    tt.fields.metricBlockCounter,
//				metricTxCounter:       tt.fields.metricTxCounter,
//				metricBlockCommitTime: tt.fields.metricBlockCommitTime,
//				storeHelper:           tt.fields.storeHelper,
//				blockInterval:         tt.fields.blockInterval,
//			}
//			got, got1, got2, err := chain.checkLastProposedBlock(tt.args.block)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("checkLastProposedBlock() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("checkLastProposedBlock() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("checkLastProposedBlock() got1 = %v, want %v", got1, tt.want1)
//			}
//			if !reflect.DeepEqual(got2, tt.want2) {
//				t.Errorf("checkLastProposedBlock() got2 = %v, want %v", got2, tt.want2)
//			}
//		})
//	}
//}
//
//func TestIfOpenConsensusMessageTurbo(t *testing.T) {
//	type args struct {
//		chainConf protocol.ChainConf
//	}
//	tests := []struct {
//		name string
//		args args
//		want bool
//	}{
//		{
//			name: "test0",
//			args: args{
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: false,
//							},
//						},
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensus.ConsensusType_SOLO,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//			},
//			want: false,
//		},
//		{
//			name: "test1",
//			args: args{
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensus.ConsensusType_TBFT,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//			},
//			want: true,
//		},
//	}
//	for k, tt := range tests {
//
//		if k == 0 {
//			localconf.ChainMakerConfig = &localconf.CMConfig{
//				TxPoolConfig: map[string]interface{}{
//					"pool_type": "test", //batch.TxPoolType
//				},
//			}
//		}
//		if k == 1 {
//			localconf.ChainMakerConfig = &localconf.CMConfig{
//				TxPoolConfig: map[string]interface{}{
//					"pool_type": "batch", //batch.TxPoolType
//				},
//			}
//		}
//		t.Run(tt.name, func(t *testing.T) {
//			if got := IfOpenConsensusMessageTurbo(tt.args.chainConf); got != tt.want {
//				t.Errorf("IfOpenConsensusMessageTurbo() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestRecoverBlock(t *testing.T) {
//	type args struct {
//		block      *commonpb.Block
//		mode       protocol.VerifyMode
//		chainConf  protocol.ChainConf
//		txPool     protocol.TxPool
//		logger     protocol.Logger
//		proposeId  string
//		netService protocol.NetService
//		ac         protocol.AccessControlProvider
//	}
//
//	block := createBlock(1)
//	tests := []struct {
//		name    string
//		args    args
//		want    *commonpb.Block
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			args: args{
//				block: createBlock(0),
//				mode:  0,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensus.ConsensusType_TBFT,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//				txPool: func() protocol.TxPool {
//					txPool := newMockTxPool(t)
//					txPool.EXPECT().GetTxsByTxIds(gomock.Any()).Return(map[string]*commonpb.Transaction{}, map[string]struct{}{}).AnyTimes()
//					txsReSet := map[string]*commonpb.Transaction{
//						"chain1": {
//							Payload: &commonpb.Payload{
//								ChainId: "chain1",
//								TxId:    "chain1",
//							},
//						},
//					}
//					txPool.EXPECT().GetAllTxsByTxIds(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(txsReSet, nil).AnyTimes()
//					return txPool
//				}(),
//				logger:    newMockLogger(t),
//				proposeId: "test0",
//				ac: func() protocol.AccessControlProvider {
//					ac := newMockAccessControlProvider(t)
//					ctrl := gomock.NewController(t)
//					member := mock.NewMockMember(ctrl)
//					member.EXPECT().GetMemberId().Return("chain1").AnyTimes()
//					ac.EXPECT().NewMember(gomock.Any()).Return(member, nil).AnyTimes()
//					return ac
//				}(),
//				netService: func() protocol.NetService {
//					ctrl := gomock.NewController(t)
//					netService := mock.NewMockNetService(ctrl)
//					netService.EXPECT().GetNodeUidByCertId(gomock.Any()).AnyTimes()
//					return netService
//				}(),
//			},
//			want:    createBlock(0),
//			wantErr: false,
//		},
//		{ // 批量交易
//			name: "test1",
//			args: args{
//				block: block,
//				mode:  protocol.CONSENSUS_VERIFY,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Core: &configpb.CoreConfig{
//							ConsensusTurboConfig: &configpb.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//								RetryTime:             100,
//								RetryInterval:         100,
//							},
//						},
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensus.ConsensusType_TBFT,
//						},
//					}
//
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//				txPool: func() protocol.TxPool {
//					txPool := newMockTxPool(t)
//					txPool.EXPECT().GetTxsByTxIds(gomock.Any()).Return(map[string]*commonpb.Transaction{}, map[string]struct{}{}).AnyTimes()
//					txsReSet := map[string]*commonpb.Transaction{
//						"chain1": {
//							Payload: &commonpb.Payload{
//								ChainId: "chain1",
//								TxId:    "chain1",
//							},
//						},
//					}
//					txPool.EXPECT().GetAllTxsByTxIds(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(txsReSet, nil).AnyTimes()
//					return txPool
//				}(),
//				logger:    newMockLogger(t),
//				proposeId: "test1",
//				ac: func() protocol.AccessControlProvider {
//					ac := newMockAccessControlProvider(t)
//					ctrl := gomock.NewController(t)
//					member := mock.NewMockMember(ctrl)
//					member.EXPECT().GetMemberId().Return("chain1").AnyTimes()
//					ac.EXPECT().NewMember(gomock.Any()).Return(member, nil).AnyTimes()
//					return ac
//				}(),
//				netService: func() protocol.NetService {
//					ctrl := gomock.NewController(t)
//					netService := mock.NewMockNetService(ctrl)
//					netService.EXPECT().GetNodeUidByCertId(gomock.Any()).AnyTimes()
//					return netService
//				}(),
//			},
//			want:    block,
//			wantErr: false,
//		},
//	}
//	for k, tt := range tests {
//
//		if k == 0 {
//			localconf.ChainMakerConfig = &localconf.CMConfig{
//				TxPoolConfig: map[string]interface{}{
//					"pool_type": "test", //batch.TxPoolType
//				},
//			}
//		}
//		if k == 1 {
//			localconf.ChainMakerConfig = &localconf.CMConfig{
//				TxPoolConfig: map[string]interface{}{
//					"pool_type": "batch", //batch.TxPoolType
//				},
//			}
//		}
//
//		t.Run(tt.name, func(t *testing.T) {
//			got, _, err := RecoverBlock(tt.args.block, tt.args.mode, tt.args.chainConf, tt.args.txPool, tt.args.ac, tt.args.netService, tt.args.logger)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("RecoverBlock() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got.Header, tt.want.Header) {
//				t.Errorf("RecoverBlock() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func newMockChainConf(t *testing.T) *mock.MockChainConf {
//	ctrl := gomock.NewController(t)
//	chainConf := mock.NewMockChainConf(ctrl)
//	return chainConf
//}
//
//func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
//	ctrl := gomock.NewController(t)
//	blockchainStore := mock.NewMockBlockchainStore(ctrl)
//	return blockchainStore
//}
//
//func newMockStoreHelper(t *testing.T) *mock.MockStoreHelper {
//	ctrl := gomock.NewController(t)
//	storeHelper := mock.NewMockStoreHelper(ctrl)
//	return storeHelper
//}
//
//func newMockLogger(t *testing.T) *mock.MockLogger {
//	ctrl := gomock.NewController(t)
//	logger := mock.NewMockLogger(ctrl)
//	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
//	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
//	logger.EXPECT().Error(gomock.Any()).AnyTimes()
//
//	return logger
//}
//
//func newMockVmManager(t *testing.T) *mock.MockVmManager {
//	ctrl := gomock.NewController(t)
//	vmManager := mock.NewMockVmManager(ctrl)
//	vmManager.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonpb.ContractResult{
//		Code: 0,
//	}, protocol.ExecOrderTxTypeNormal, commonpb.TxStatusCode_SUCCESS).AnyTimes()
//	return vmManager
//}
//
//func newMockTxPool(t *testing.T) *mock.MockTxPool {
//	ctrl := gomock.NewController(t)
//	txPool := mock.NewMockTxPool(ctrl)
//	return txPool
//}
//
//func newMockSnapshotManager(t *testing.T) *mock.MockSnapshotManager {
//	ctrl := gomock.NewController(t)
//	snapshotManager := mock.NewMockSnapshotManager(ctrl)
//	return snapshotManager
//}
//
//func newMockLedgerCache(t *testing.T) *mock.MockLedgerCache {
//	ctrl := gomock.NewController(t)
//	newMockLedgerCache := mock.NewMockLedgerCache(ctrl)
//	return newMockLedgerCache
//}
//
//func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
//	ctrl := gomock.NewController(t)
//	proposalCache := mock.NewMockProposalCache(ctrl)
//	return proposalCache
//}
//
//func newMockBlockVerifier(t *testing.T) *mock.MockBlockVerifier {
//	ctrl := gomock.NewController(t)
//	blockVerifier := mock.NewMockBlockVerifier(ctrl)
//	return blockVerifier
//}
//
//func newMockBlockCommitter(t *testing.T) *mock.MockBlockCommitter {
//	ctrl := gomock.NewController(t)
//	blockCommitter := mock.NewMockBlockCommitter(ctrl)
//	return blockCommitter
//}
//
//func newMockTxScheduler(t *testing.T) *mock.MockTxScheduler {
//	ctrl := gomock.NewController(t)
//	txScheduler := mock.NewMockTxScheduler(ctrl)
//	return txScheduler
//}
//
//func newMockSigningMember(t *testing.T) *mock.MockSigningMember {
//	ctrl := gomock.NewController(t)
//	signingMember := mock.NewMockSigningMember(ctrl)
//	return signingMember
//}
//
//func newMockSnapshot(t *testing.T) *mock.MockSnapshot {
//	ctrl := gomock.NewController(t)
//	snapshot := mock.NewMockSnapshot(ctrl)
//	return snapshot
//
//}
//
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
//
//func newMockAccessControlProvider(t *testing.T) *mock.MockAccessControlProvider {
//	ctrl := gomock.NewController(t)
//	accessControlProvider := mock.NewMockAccessControlProvider(ctrl)
//	return accessControlProvider
//}
//
//func newMockNetService(t *testing.T) *mock.MockNetService {
//	ctrl := gomock.NewController(t)
//	netService := mock.NewMockNetService(ctrl)
//	return netService
//}

func TestName(t *testing.T) {
	log := logger.GetLogger("core")

	batchIds := []string{"bac", "sss"}

	tx1 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "1",
		},
	}

	tx2 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "2",
		},
	}

	tx3 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "3",
		},
	}

	tx4 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "4",
		},
	}

	tx5 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "5",
		},
	}

	tx6 := &commonpb.Transaction{
		Payload: &commonpb.Payload{
			TxId: "6",
		},
	}

	txs := []*commonpb.Transaction{tx1, tx2, tx3, tx4, tx5, tx6}

	fetchBatches := make([][]*commonpb.Transaction, 0)
	fetchBatches = append(fetchBatches, []*commonpb.Transaction{tx4, tx6, tx1})
	fetchBatches = append(fetchBatches, []*commonpb.Transaction{tx2, tx3, tx5})

	by, err := SerializeTxBatchInfo(batchIds, txs, fetchBatches, log)
	if err != nil {
		panic(err)
	}

	bi, err := DeserializeTxBatchInfo(by)
	if err != nil {
		panic(err)
	}

	fmt.Println("txs:", txs)
	fmt.Println("fetchBatches:", fetchBatches)

	for _, v := range fetchBatches {
		fmt.Println(v)
	}

	fmt.Printf("1, bId: %s, index:%v \n", bi.BatchIds[0], bi.Index)
	fmt.Printf("2, bId: %s, index:%v \n", bi.BatchIds[1], bi.Index)
}

func Test3(t *testing.T) {
	hash, err := utils.CalcTxHashWithVersion("SHA256", &commonpb.Transaction{
		Payload:   &commonpb.Payload{TxId: "123"},
		Sender:    nil,
		Endorsers: nil,
		Result:    nil,
	}, int(protocol.DefaultBlockVersion))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x", hash)

}

func Test2(t *testing.T) {

	logger := logger.GetLogger("core")

	newBlock := new(commonpb.Block)
	txIds := []string{
		"16f48b96364bdf22cacd42a708a721aad356b4f05a0046d485b0f9a52c3d3fc4",
		"16f48b96362ce3afca05da3b2169f5a972eac0b9a127497f9f5b279bd63d61e8",
		"16f48b96362eb470cad0cc7760331b66e54f5426d4ee4a6f90073feb3d42492a",
		"16f48b96362ec7acca3138d6d342b051af860cf09c4d4b368e2170b7c31aab2a",
		"16f48b963630be30ca39b216cbc50e7360aa939d2c1346ee86ed8489106661a7",
		"16f48b963630ceddcaa32eaf936401e27d2ceb0ce85b4030b276a0e3068a692c",
		"16f48b9636443becca279a96879a4f3610d821bc4e0a42b48eaed950bdc9198d",
		"16f48b9636446b3ccab15e0501ebc34b0ad24f97532f42a995a67e961f19f1d9",
		"16f48b96364a6f8bcaa410fd6718f227685e92f4909e40ed9ee787ba34bbc26b",
		"16f48b96364af325caa3d38540dc22292c74778205814acdaf233137b48fd177",
	}

	newBlock.Txs = make([]*commonpb.Transaction, len(txIds))
	txsa := getTestTx(txIds)
	indexs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}

	txs := [][]*commonpb.Transaction{txsa}
	newTxs := make([]*commonpb.Transaction, 0)
	for _, tx := range txs {
		newTxs = append(newTxs, tx...)
	}

	for _, v := range newTxs {
		logger.Infof("<lzw> verifier, newTxs: %v", v.Payload.TxId)
	}

	logger.Infof("<lzw> verifier, indexs: %v", indexs)
	for i, v := range indexs {
		logger.Infof("<lzw> verifier,i:%d ,v:%d, newBlock: %v", i, v, newTxs[v])
		newBlock.Txs[i] = newTxs[v]
	}

	logger.Infof("<lzw> verifier, indexs: %v", indexs)

	for _, v := range newBlock.Txs {
		logger.Infof("<lzw> verifier, newBlock: %v", v.Payload.TxId)
	}

}

func getTestTx(txIds []string) []*commonpb.Transaction {
	txs := make([]*commonpb.Transaction, 0)
	for _, v := range txIds {
		txs = append(txs, &commonpb.Transaction{
			Payload: &commonpb.Payload{TxId: v},
		})
	}

	return txs
}

func TestFinalizeBlock(t *testing.T) {
	type args struct {
		block      *commonpb.Block
		txRWSetMap map[string]*commonpb.TxRWSet
		aclFailTxs []*commonpb.Transaction
		hashType   string
		logger     protocol.Logger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				block:      createBlock(1),
				txRWSetMap: map[string]*commonpb.TxRWSet{},
				aclFailTxs: nil,
				hashType:   "SHA256",
				logger:     logger.GetLogger("core"),
			},
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				block: func() *commonpb.Block {
					block := createBlock(1)
					block.Dag = nil
					return block
				}(),
				txRWSetMap: map[string]*commonpb.TxRWSet{},
				aclFailTxs: nil,
				hashType:   "SHA256",
				logger:     logger.GetLogger("core"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FinalizeBlock(tt.args.block, tt.args.txRWSetMap, tt.args.aclFailTxs, tt.args.hashType, tt.args.logger); (err != nil) != tt.wantErr {
				t.Errorf("FinalizeBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockCommitterImpl_AddBlock(t *testing.T) {
	controller := gomock.NewController(t)
	helper := mock2.NewMockStoreHelper(controller)
	helper.EXPECT().RollBack(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	type fields struct {
		storeHelper conf.StoreHelper
	}
	type args struct {
		block *commonpb.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "异常流",
			fields: fields{
				storeHelper: helper,
			},
			args: args{
				block: &commonpb.Block{
					Header: &commonpb.BlockHeader{BlockHeight: 666},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if panicError := recover(); panicError == nil {
					t.Errorf("AddBlock() panic is nil")
				}
			}()
			chain := &BlockCommitterImpl{
				storeHelper: tt.fields.storeHelper,
			}
			// 不需要关注err
			_ = chain.AddBlock(tt.args.block)
		})
	}
}
