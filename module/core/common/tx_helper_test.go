package common

import (
	"encoding/hex"
	"fmt"
	"testing"

	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//func TestValidateTx(t *testing.T) {
//	verifyTx, block := txPrepare(t)
//
//	chainConf := newMockChainConf(t)
//	config := &config.ChainConfig{
//		ChainId:   "chain1",
//		Version:   "1.0",
//		Crypto:    &config.CryptoConfig{Hash: "SHA256"},
//		Consensus: &config.ConsensusConfig{Type: 0},
//		Core: &config.CoreConfig{
//			ConsensusTurboConfig: &config.ConsensusTurboConfig{
//				ConsensusMessageTurbo: true,
//			},
//		},
//		Block: &config.BlockConfig{},
//	}
//
//	proposalCache := newMockProposalCache(t)
//	proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(),
//	gomock.Any()).Return(nil, nil).AnyTimes()
//
//	chainConf.EXPECT().ChainConfig().AnyTimes().Return(config)
//
//	verifyTx.chainConf = chainConf
//	verifyTx.proposalCache = proposalCache
//	hashes, _, _, err := verifyTx.verifierTxs(block, protocol.SYNC_VERIFY, QuickSyncVerifyMode)
//	require.Nil(t, err)
//
//	for _, hash := range hashes {
//		fmt.Println("test hash: ", hex.EncodeToString(hash))
//	}
//}

func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
	ctrl := gomock.NewController(t)
	proposalCache := mock.NewMockProposalCache(ctrl)
	return proposalCache
}

func newMockChainConf(t *testing.T) *mock.MockChainConf {
	ctrl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctrl)
	return chainConf
}

func newTx(txId string, contractId *commonpb.Contract, parameterMap map[string]string) *commonpb.Transaction {

	var parameters []*commonpb.KeyValuePair
	for key, value := range parameterMap {
		parameters = append(parameters, &commonpb.KeyValuePair{
			Key:   key,
			Value: []byte(value),
		})
	}

	return &commonpb.Transaction{
		Payload: &commonpb.Payload{
			ChainId:        "chain1",
			TxType:         commonpb.TxType_QUERY_CONTRACT,
			TxId:           txId,
			Timestamp:      0,
			ExpirationTime: 0,
			ContractName:   contractId.Name,
			Method:         "set",
			Parameters:     parameters,
			Sequence:       0,
			Limit:          nil,
		},
		Sender:    nil,
		Endorsers: nil,
		Result: &commonpb.Result{
			Code: 0,
			ContractResult: &commonpb.ContractResult{
				Code:          0,
				Result:        nil,
				Message:       "",
				GasUsed:       0,
				ContractEvent: nil,
			},
			RwSetHash: nil,
		},
	}

}

func txPrepare(t *testing.T) (*VerifierTx, *commonpb.Block) {
	block := newBlock()
	//contractId := &commonpb.Contract{
	//	ContractName:    "ContractName",
	//	ContractVersion: "1",
	//	RuntimeType:     commonpb.RuntimeType_WASMER,
	//}
	contractId := &commonpb.Contract{
		Name:        "ContractName",
		Version:     "1",
		RuntimeType: commonpb.RuntimeType_DOCKER_GO,
		Status:      0,
		Creator:     nil,
	}

	parameters := make(map[string]string, 8)
	tx0 := newTx("a0000000000000000000000000000000", contractId, parameters)
	txs := make([]*commonpb.Transaction, 0)
	txs = append(txs, tx0)
	block.Txs = txs

	var txRWSetMap = make(map[string]*commonpb.TxRWSet, 3)
	txRWSetMap[tx0.Payload.TxId] = &commonpb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonpb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonpb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K2"),
			Value:        []byte("V"),
		}},
	}

	rwHash, _ := hex.DecodeString("d02f421ed76e0e26e9def824a8b84c7c223d484762d6d060a8b71e1649d1abbf")
	result := &commonpb.Result{
		Code: commonpb.TxStatusCode_SUCCESS,
		ContractResult: &commonpb.ContractResult{
			Code:    0,
			Result:  nil,
			Message: "",
			GasUsed: 0,
		},
		RwSetHash: rwHash,
	}
	tx0.Result = result
	txResultMap := make(map[string]*commonpb.Result, 1)
	txResultMap[tx0.Payload.TxId] = result

	log := logger.GetLoggerByChain(logger.MODULE_CORE, "chain1")

	ctl := gomock.NewController(t)
	store := mock.NewMockBlockchainStore(ctl)
	txPool := mock.NewMockTxPool(ctl)

	ac := mock.NewMockAccessControlProvider(ctl)
	chainConf := mock.NewMockChainConf(ctl)

	store.EXPECT().TxExists(tx0).AnyTimes().Return(false, nil)

	txsMap := make(map[string]*commonpb.Transaction)

	txsMap[tx0.Payload.TxId] = tx0

	//txPool.EXPECT().GetTxsByTxIds([]string{tx0.Payload.TxId}).Return(txsMap, nil)
	txPool.EXPECT().GetTxsByTxIds(gomock.Any()).Return(txsMap, nil).AnyTimes()
	//config := &config.ChainConfig{
	//	ChainId: "chain1",
	//	Crypto: &config.CryptoConfig{
	//		Hash: "SHA256",
	//	},
	//}
	config := &config.ChainConfig{
		ChainId:   "chain1",
		Version:   "1.0",
		Crypto:    &config.CryptoConfig{Hash: "SHA256"},
		Consensus: &config.ConsensusConfig{Type: 0},
	}

	chainConf.EXPECT().ChainConfig().AnyTimes().Return(config)

	principal := mock.NewMockPrincipal(ctl)
	ac.EXPECT().CreatePrincipal("123", nil, nil).AnyTimes().Return(principal, nil)
	ac.EXPECT().VerifyPrincipal(principal).AnyTimes().Return(true, nil)
	verifyTxConf := &VerifierTxConfig{
		Block:       block,
		TxRWSetMap:  txRWSetMap,
		TxResultMap: txResultMap,
		TxPool:      txPool,
		Ac:          ac,
		ChainConf:   chainConf,
		Log:         log,
	}
	return NewVerifierTx(verifyTxConf), block
}

func newBlock() *commonpb.Block {
	var hash = []byte("0123456789")
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockVersion:   1,
			BlockType:      0,
			ChainId:        "chain1",
			BlockHeight:    3,
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

func TestIfExitInSameBranch(t *testing.T) {

	tx1 := createNewTestTx("123456")
	tx2 := createNewTestTx("1234567")
	tx3 := createNewTestTx("1234568")
	tx4 := createNewTestTx("1234569")
	tx5 := createNewTestTx("12345610")

	b0 := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  9,
			BlockHash:    []byte("012345"),
			PreBlockHash: []byte("012345"),
		},
		Txs: nil,
	}

	b1 := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  10,
			BlockHash:    []byte("0123456"),
			PreBlockHash: []byte("012345"),
		},
		Txs: []*commonpb.Transaction{tx1, tx2},
	}

	b2 := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  11,
			BlockHash:    []byte("123"),
			PreBlockHash: []byte("0123456"),
		},
		Txs: []*commonpb.Transaction{tx3},
	}

	b2a := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  11,
			BlockHash:    []byte("123a"),
			PreBlockHash: []byte("0123456"),
		},
		Txs: []*commonpb.Transaction{tx2},
	}

	b3 := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  12,
			BlockHash:    []byte("1234"),
			PreBlockHash: []byte("123"),
		},
		Txs: []*commonpb.Transaction{tx4},
	}

	b3a := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  12,
			BlockHash:    []byte("1234a"),
			PreBlockHash: []byte("123"),
		},
		Txs: []*commonpb.Transaction{tx2},
	}

	b3b := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  12,
			BlockHash:    []byte("1234b"),
			PreBlockHash: []byte("123"),
		},
		Txs: []*commonpb.Transaction{tx5, tx3},
	}

	b4 := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  13,
			BlockHash:    []byte("12345"),
			PreBlockHash: []byte("1234"),
		},
		Txs: []*commonpb.Transaction{tx1},
	}

	b4a := commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockHeight:  13,
			BlockHash:    []byte("12345"),
			PreBlockHash: []byte("1234"),
		},
		Txs: []*commonpb.Transaction{tx5},
	}

	ctl := gomock.NewController(t)
	proposalCache := mock.NewMockProposalCache(ctl)
	proposalCache.EXPECT().GetProposedBlockByHashAndHeight(b0.Header.BlockHash, b0.Header.BlockHeight).Return(nil, nil).AnyTimes()
	cases := []struct {
		b0       *commonpb.Block
		b1       *commonpb.Block
		preBlock *commonpb.Block
		block    *commonpb.Block
		doc      string
		expected bool // expected result
	}{
		/**
								-> b3a
		 						-> b3b

						 -> b2

								-> b3  ->   b4
									   ->	b4a
				b0 -> b1
						 -> b2a

		*/
		{nil, nil, &b1, &b2, "区块b2里的交易与前面的区块的交易不重复", false},
		{nil, nil, &b1, &b2a, "区块b2a里的交易与b1的区块的交易重复", true},
		{nil, &b1, &b2, &b3a, "区块b3a里的交易与b1的区块的交易重复", true},
		{nil, &b1, &b2, &b3b, "区块b3b里的交易与b2的区块的交易重复", true},
		{nil, &b1, &b2, &b3, "区块b3里的交易与前面的区块的交易不重复", false},
		//
		{&b1, &b2, &b3, &b4, "区块b4里的交易与b1的区块的交易重复", true},
		{&b1, &b2, &b3, &b4a, "区块b4a里的交易与b3b的区块的交易重复", false},
	}

	for i, v := range cases {
		proposalCachePrepare(proposalCache, v.b0, v.b1, v.preBlock)

		var finalResult bool
		for _, tx := range v.block.Txs {
			result, _ := IfExitInSameBranch(
				v.block.Header.BlockHeight,
				tx.Payload.TxId,
				proposalCache,
				v.block.Header.PreBlockHash)

			if result {
				finalResult = true
			}
		}

		if finalResult != v.expected {
			fmt.Printf("Case:%d fail \n", i)
			require.NotEqual(t, v.expected, finalResult)
		} else {
			fmt.Printf("Case:%d pass \n", i)
		}

	}

}

func proposalCachePrepare(proposalCache *mock.MockProposalCache, b0, b1, preBlock *commonpb.Block) {
	proposalCache.EXPECT().GetProposedBlockByHashAndHeight(
		preBlock.Header.BlockHash,
		preBlock.Header.BlockHeight).
		Return(preBlock, nil).AnyTimes()

	if b0 != nil {
		proposalCache.EXPECT().GetProposedBlockByHashAndHeight(
			b0.Header.BlockHash,
			b0.Header.BlockHeight).
			Return(b0, nil).AnyTimes()
	}

	if b1 != nil {
		proposalCache.EXPECT().GetProposedBlockByHashAndHeight(
			b1.Header.BlockHash,
			b1.Header.BlockHeight).
			Return(b1, nil).AnyTimes()
	}

}

//func TestValidateTx1(t *testing.T) {
//	type args struct {
//		txsRet        map[string]*commonpb.Transaction
//		tx            *commonpb.Transaction
//		blockHeight   uint64
//		stat          *VerifyStat
//		newAddTxs     []*commonpb.Transaction
//		block         *commonpb.Block
//		consensusType consensusPb.ConsensusType
//		chainId       string
//		ac            protocol.AccessControlProvider
//		proposalCache protocol.ProposalCache
//		txFilter      protocol.TxFilter
//	}
//
//	var (
//		txFilter      = mock.NewMockTxFilter(gomock.NewController(t))
//		ac            = newMockAccessControlProvider(t)
//		proposalCache = newMockProposalCache(t)
//		block0        = createBlock(0)
//	)
//	txFilter.EXPECT().IsExists(gomock.Any()).AnyTimes()
//
//	ctrl := gomock.NewController(t)
//	principal := mock.NewMockPrincipal(ctrl)
//	ac.EXPECT().LookUpExceptionalPolicy(gomock.Any()).Return(&accesscontrol.Policy{}, nil).AnyTimes()
//	ac.EXPECT().CreatePrincipal("123", nil, nil).AnyTimes().Return(nil, nil)
//	ac.EXPECT().VerifyPrincipal(principal).AnyTimes().Return(true, nil)
//	ac.EXPECT().LookUpPolicy(gomock.Any()).AnyTimes()
//
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test1",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test1": {
//							Payload: &commonpb.Payload{
//								TxId: "test1",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						TxId: "test1",
//					},
//				},
//				blockHeight:   0,
//				stat:          nil,
//				newAddTxs:     nil,
//				block:         block0,
//				consensusType: 0,
//				txFilter:      txFilter,
//				chainId:       "",
//				ac:            ac,
//				proposalCache: proposalCache,
//			},
//			wantErr: false,
//		},
//		{
//			name: "test2",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test2": {
//							Payload: &commonpb.Payload{
//								TxId: "test2",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						TxId: "test2",
//					},
//				},
//				blockHeight: 2,
//				stat:        nil,
//				newAddTxs:   nil,
//				block: func() *commonpb.Block {
//					block := createBlock(1)
//					return block
//				}(),
//				consensusType: consensusPb.ConsensusType_MAXBFT,
//				txFilter:      txFilter,
//				ac:            ac,
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache = newMockProposalCache(t)
//					block0 := createNewTestBlock(0)
//					proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).Return(block0, nil).AnyTimes()
//					return proposalCache
//				}(),
//			},
//			wantErr: false,
//		},
//		{
//			name: "test3",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test3": {
//							Payload: &commonpb.Payload{
//								TxId: "test3",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						TxId: "test3",
//					},
//				},
//				blockHeight: 1,
//				stat:        nil,
//				newAddTxs:   nil,
//				block: func() *commonpb.Block {
//					block := createBlock(1)
//					return block
//				}(),
//				consensusType: consensusPb.ConsensusType_MAXBFT,
//				txFilter:      txFilter,
//				ac:            ac,
//				proposalCache: func() protocol.ProposalCache {
//					proposalCache = newMockProposalCache(t)
//					block0 := createNewTestBlock(0)
//					proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).Return(block0, nil).AnyTimes()
//					return proposalCache
//				}(),
//			},
//			wantErr: false,
//		},
//		{
//			name: "test4",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test4": {
//							Payload: &commonpb.Payload{
//								TxId: "test4",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						TxId: "test4",
//					},
//				},
//				blockHeight: 1,
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				newAddTxs: nil,
//				block: func() *commonpb.Block {
//					block := createBlock(1)
//					return block
//				}(),
//				consensusType: consensusPb.ConsensusType_TBFT,
//				txFilter:      txFilter,
//				ac:            ac,
//				proposalCache: proposalCache,
//			},
//			wantErr: false,
//		},
//		{
//			name: "test5",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test5": {
//							Payload: &commonpb.Payload{
//								ChainId:      "test5",
//								TxId:         "test5",
//								TxType:       commonpb.TxType_INVOKE_CONTRACT,
//								ContractName: "test5",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						ChainId:      "test5",
//						TxId:         "test5",
//						ContractName: "123",
//						Method:       "test",
//						TxType:       commonpb.TxType_INVOKE_CONTRACT,
//					},
//				},
//				blockHeight: 1,
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				newAddTxs: nil,
//				block: func() *commonpb.Block {
//					block := createBlock(1)
//					return block
//				}(),
//				consensusType: consensusPb.ConsensusType_TBFT,
//				txFilter:      txFilter,
//				ac:            ac,
//				proposalCache: proposalCache,
//			},
//			wantErr: true,
//		},
//		{
//			name: "test6",
//			args: args{
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test6": {
//							Payload: &commonpb.Payload{
//								ChainId:      "test6",
//								TxId:         "test6",
//								TxType:       commonpb.TxType_INVOKE_CONTRACT,
//								ContractName: "test6",
//								Method:       "test",
//							},
//						},
//					}
//					return txRet
//				}(),
//				tx: &commonpb.Transaction{
//					Payload: &commonpb.Payload{
//						ChainId:      "test6",
//						TxId:         "test6",
//						ContractName: "test6",
//						Method:       "test",
//						TxType:       commonpb.TxType_INVOKE_CONTRACT,
//					},
//				},
//				blockHeight: 1,
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				newAddTxs: nil,
//				block: func() *commonpb.Block {
//					block := createBlock(1)
//					return block
//				}(),
//				consensusType: consensusPb.ConsensusType_TBFT,
//				txFilter:      txFilter,
//				ac:            ac,
//				proposalCache: proposalCache,
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := ValidateTx(tt.args.txsRet, tt.args.tx, tt.args.stat, tt.args.newAddTxs, tt.args.block, tt.args.consensusType, txFilter, tt.args.chainId, tt.args.ac, tt.args.proposalCache, protocol.SYNC_VERIFY, QuickSyncVerifyMode); (err != nil) != tt.wantErr {
//				t.Errorf("ValidateTx() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestVerifierTx_verifyTx(t *testing.T) {
//	type fields struct {
//		block       *commonpb.Block
//		txRWSetMap  map[string]*commonpb.TxRWSet
//		txResultMap map[string]*commonpb.Result
//		log         protocol.Logger
//		txFilter    protocol.TxFilter
//		txPool      protocol.TxPool
//		ac          protocol.AccessControlProvider
//		chainConf   protocol.ChainConf
//	}
//	type args struct {
//		txs          []*commonpb.Transaction
//		txsRet       map[string]*commonpb.Transaction
//		txsHeightRet map[string]uint64
//		stat         *VerifyStat
//		block        *commonpb.Block
//		mode         protocol.VerifyMode
//	}
//
//	var (
//		block0   = newBlock()
//		log      = newMockLogger(t)
//		txFilter = mock.NewMockTxFilter(gomock.NewController(t))
//		txPool   = newMockTxPool(t)
//		ac       = newMockAccessControlProvider(t)
//	)
//	txFilter.EXPECT().IsExists(gomock.Any()).AnyTimes()
//	log.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
//	log.EXPECT().Warn(gomock.Any()).AnyTimes()
//	log.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
//
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    [][]byte
//		want1   []*commonpb.Transaction
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				block:       block0,
//				txRWSetMap:  nil,
//				txResultMap: nil,
//				log:         log,
//				txFilter:    txFilter,
//				txPool:      txPool,
//				ac:          ac,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &config.ChainConfig{
//						Crypto: &config.CryptoConfig{Hash: "SHA256"},
//						Consensus: &config.ConsensusConfig{
//							Type: consensusPb.ConsensusType_TBFT,
//						},
//						Core: &config.CoreConfig{
//							ConsensusTurboConfig: &config.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Block: &config.BlockConfig{},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//			},
//			args: args{
//				txs: func() []*commonpb.Transaction {
//					txs := make([]*commonpb.Transaction, 0)
//					txs = []*commonpb.Transaction{
//						{
//							Payload: &commonpb.Payload{
//								ChainId: "test0",
//							},
//						},
//					}
//					return txs
//				}(),
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test0": {
//							Payload: &commonpb.Payload{
//								TxId: "test0",
//							},
//						},
//					}
//					return txRet
//				}(),
//				txsHeightRet: nil,
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				block: block0,
//				mode:  protocol.SYNC_VERIFY,
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				block: block0,
//				txRWSetMap: func() map[string]*commonpb.TxRWSet {
//					txRet := map[string]*commonpb.TxRWSet{
//						"test1": {
//							TxReads:  []*commonpb.TxRead{},
//							TxWrites: []*commonpb.TxWrite{},
//						},
//					}
//					return txRet
//				}(),
//				txResultMap: func() map[string]*commonpb.Result {
//					result := map[string]*commonpb.Result{
//						"test1": {
//							RwSetHash: []byte("test1"),
//						},
//					}
//					return result
//				}(),
//				log:      log,
//				txFilter: txFilter,
//				txPool:   txPool,
//				ac:       ac,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &config.ChainConfig{
//						Crypto: &config.CryptoConfig{Hash: "SHA256"},
//						Consensus: &config.ConsensusConfig{
//							Type: consensusPb.ConsensusType_TBFT,
//						},
//						Core: &config.CoreConfig{
//							ConsensusTurboConfig: &config.ConsensusTurboConfig{
//								ConsensusMessageTurbo: false,
//							},
//						},
//						Block: &config.BlockConfig{},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//			},
//			args: args{
//				txs: func() []*commonpb.Transaction {
//					txs := make([]*commonpb.Transaction, 0)
//					txs = []*commonpb.Transaction{
//						{
//							Payload: &commonpb.Payload{
//								ChainId:      "test1",
//								TxId:         "test1",
//								TxType:       commonpb.TxType_INVOKE_CONTRACT,
//								ContractName: "test1",
//								Method:       "test",
//							},
//							Result: &commonpb.Result{
//								RwSetHash: []byte("test1"),
//							},
//						},
//					}
//					return txs
//				}(),
//				txsRet: func() map[string]*commonpb.Transaction {
//					txRet := map[string]*commonpb.Transaction{
//						"test1": {
//							Payload: &commonpb.Payload{
//								ChainId:      "test1",
//								TxId:         "test1",
//								TxType:       commonpb.TxType_INVOKE_CONTRACT,
//								ContractName: "test1",
//								Method:       "test",
//							},
//						},
//					}
//					return txRet
//				}(),
//				txsHeightRet: map[string]uint64{
//					"test1": 0,
//				},
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				block: nil,
//				mode:  protocol.SYNC_VERIFY,
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test2", // test2 case verify tx timestamp is expired
//			fields: fields{
//				block: block0,
//				txRWSetMap: map[string]*commonpb.TxRWSet{
//					"test2": {
//						TxReads:  []*commonpb.TxRead{},
//						TxWrites: []*commonpb.TxWrite{},
//					},
//				},
//				txResultMap: map[string]*commonpb.Result{
//					"test2": {
//						RwSetHash: []byte("test2"),
//					},
//				},
//				log:      log,
//				txFilter: txFilter,
//				txPool:   txPool,
//				ac:       ac,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &config.ChainConfig{
//						Crypto: &config.CryptoConfig{Hash: "SHA256"},
//						Consensus: &config.ConsensusConfig{
//							Type: consensusPb.ConsensusType_TBFT,
//						},
//						Core: &config.CoreConfig{
//							ConsensusTurboConfig: &config.ConsensusTurboConfig{
//								ConsensusMessageTurbo: true,
//							},
//						},
//						Block: &config.BlockConfig{
//							TxTimestampVerify: true,
//							TxTimeout:         30,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					return chainConf
//				}(),
//			},
//			args: args{
//				txs: []*commonpb.Transaction{
//					{
//						Payload: &commonpb.Payload{
//							ChainId:      "test2",
//							TxId:         "test2",
//							TxType:       commonpb.TxType_INVOKE_CONTRACT,
//							ContractName: "test2",
//							Method:       "test2",
//							Timestamp:    utils.CurrentTimeSeconds() - 60,
//						},
//						Result: &commonpb.Result{
//							RwSetHash: []byte("test2"),
//						},
//					},
//				},
//				txsRet: map[string]*commonpb.Transaction{
//					"test2": {
//						Payload: &commonpb.Payload{
//							ChainId:      "test2",
//							TxId:         "test2",
//							TxType:       commonpb.TxType_INVOKE_CONTRACT,
//							ContractName: "test2",
//							Method:       "test2",
//							Timestamp:    utils.CurrentTimeSeconds() - 60,
//						},
//					},
//				},
//				txsHeightRet: map[string]uint64{
//					"test1": 0,
//				},
//				stat: &VerifyStat{
//					TotalCount: 20,
//				},
//				block: nil,
//				mode:  protocol.CONSENSUS_VERIFY,
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			vt := &VerifierTx{
//				block:       tt.fields.block,
//				txRWSetMap:  tt.fields.txRWSetMap,
//				txResultMap: tt.fields.txResultMap,
//				log:         tt.fields.log,
//				txFilter:    tt.fields.txFilter,
//				txPool:      tt.fields.txPool,
//				ac:          tt.fields.ac,
//				chainConf:   tt.fields.chainConf,
//			}
//			got, got1, _, err := vt.verifyTx(tt.args.txs, tt.args.txsRet, tt.args.stat, tt.args.block, protocol.SYNC_VERIFY, QuickSyncVerifyMode)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("verifyTx() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("verifyTx() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("verifyTx() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}
//
//func TestValidateTxRules(t *testing.T) {
//	var txs []*commonpb.Transaction
//	for i := 0; i < 100; i++ {
//		txs = append(txs, &commonpb.Transaction{
//			Payload: &commonpb.Payload{
//				TxId: utils.GetTimestampTxId(),
//			},
//		})
//	}
//	type args struct {
//		filter protocol.TxFilter
//		txs    []*commonpb.Transaction
//	}
//	tests := []struct {
//		name          string
//		args          args
//		wantRemoveTxs []*commonpb.Transaction
//		wantRemainTxs []*commonpb.Transaction
//	}{
//		{
//			name: "正常流",
//			args: args{
//				filter: func() protocol.TxFilter {
//					filter := mock.NewMockTxFilter(gomock.NewController(t))
//					filter.EXPECT().ValidateRule(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(txId string, ruleType ...bn.RuleType) error {
//						if []byte(txId)[63]%2 == 1 {
//							return nil
//						}
//						return bn.ErrKeyTimeIsNotInTheFilterRange
//					})
//					return filter
//				}(),
//				txs: txs,
//			},
//			wantRemoveTxs: func() []*commonpb.Transaction {
//				var arr []*commonpb.Transaction
//				for _, tx := range txs {
//					if []byte(tx.Payload.TxId)[63]%2 == 0 {
//						arr = append(arr, tx)
//					}
//				}
//				return arr
//			}(),
//			wantRemainTxs: func() []*commonpb.Transaction {
//				var arr []*commonpb.Transaction
//				for _, tx := range txs {
//					if []byte(tx.Payload.TxId)[63]%2 == 1 {
//						arr = append(arr, tx)
//					}
//				}
//				return arr
//			}(),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotRemoveTxs, gotRemainTxs := ValidateTxRules(tt.args.filter, tt.args.txs)
//			if !reflect.DeepEqual(gotRemoveTxs, tt.wantRemoveTxs) {
//				t.Errorf("ValidateTxRules() gotRemoveTxs = %v, want %v", gotRemoveTxs, tt.wantRemoveTxs)
//			}
//			if !reflect.DeepEqual(gotRemainTxs, tt.wantRemainTxs) {
//				t.Errorf("ValidateTxRules() gotRemainTxs = %v, want %v", gotRemainTxs, tt.wantRemainTxs)
//			}
//		})
//	}
//}
