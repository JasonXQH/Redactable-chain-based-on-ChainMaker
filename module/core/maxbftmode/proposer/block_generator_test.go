/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package proposer

//
//import (
//	"reflect"
//	"testing"
//	"time"
//
//	"chainmaker.org/chainmaker-go/module/core/common"
//	"chainmaker.org/chainmaker-go/module/core/provider/conf"
//	"chainmaker.org/chainmaker-go/module/snapshot"
//	pbac "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
//	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
//	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
//	"chainmaker.org/chainmaker/pb-go/v2/consensus"
//	"chainmaker.org/chainmaker/protocol/v2"
//	"chainmaker.org/chainmaker/protocol/v2/mock"
//	"chainmaker.org/chainmaker/protocol/v2/test"
//
//	"github.com/golang/mock/gomock"
//)
//
///*
// * test unit BlockProposerImpl generateNewBlock func
// */
//func TestBlockProposerImpl_generateNewBlock(t *testing.T) {
//	type fields struct {
//		chainId      string
//		txScheduler  protocol.TxScheduler
//		blockBuilder *common.BlockBuilder
//		storeHelper  conf.StoreHelper
//	}
//	type args struct {
//		proposingHeight uint64
//		preHash         []byte
//		txBatch         []*commonpb.Transaction
//	}
//
//	var (
//		storeHelper = newMockStoreHelper(t)
//		txScheduler = newMockTxScheduler(t)
//		chainConf   = newMockChainConf(t)
//	)
//	storeHelper.EXPECT().BeginDbTransaction(gomock.Any(), gomock.Any()).AnyTimes()
//
//	txBatch := []*commonpb.Transaction{createNewTestTx("chain1")}
//
//	blockTest3 := func() *commonpb.Block {
//
//		var hash = []byte("0123456789")
//		var block = &commonpb.Block{
//			Header: &commonpb.BlockHeader{
//				BlockVersion:   220,
//				ChainId:        "chain1",
//				BlockHeight:    1,
//				PreBlockHash:   hash,
//				BlockTimestamp: time.Now().Unix(),
//				Proposer:       &pbac.Member{},
//			},
//			Dag: &commonpb.DAG{Vertexes: nil},
//		}
//		return block
//	}()
//
//	txScheduler.EXPECT().Schedule(blockTest3, txBatch, gomock.Any()).AnyTimes()
//
//	chainConfig := &configpb.ChainConfig{
//		Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
//		Consensus: &configpb.ConsensusConfig{
//			Type: consensus.ConsensusType_TBFT,
//		},
//	}
//	chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
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
//				chainId: "test0",
//				blockBuilder: func() *common.BlockBuilder {
//					bbConf := &common.BlockBuilderConf{
//						ChainId:         "chain1",
//						TxPool:          newMockTxPool(t),
//						TxScheduler:     newMockTxScheduler(t),
//						SnapshotManager: newMockSnapshotManager(t),
//						Identity:        newMockSigningMember(t),
//						LedgerCache: func() protocol.LedgerCache {
//							ledgerCache := newMockLedgerCache(t)
//							ledgerCache.EXPECT().GetLastCommittedBlock().Return(nil).AnyTimes()
//							ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
//							return ledgerCache
//						}(),
//						ProposalCache: newMockProposalCache(t),
//						ChainConf:     newMockChainConf(t),
//						Log:           newMockLogger(t),
//						StoreHelper:   newMockStoreHelper(t),
//					}
//					blockBuilder := common.NewBlockBuilder(bbConf)
//					return blockBuilder
//				}(),
//				storeHelper: newMockStoreHelper(t),
//			},
//			args: args{
//				proposingHeight: 1,
//				preHash:         []byte("test0"),
//				txBatch: []*commonpb.Transaction{
//					{
//						Payload: &commonpb.Payload{
//							TxId: "test0",
//						},
//					},
//				},
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//		{
//			name: "test1",
//			fields: fields{
//				chainId: "chain1",
//				blockBuilder: func() *common.BlockBuilder {
//					bbConf := &common.BlockBuilderConf{
//						ChainId: "chain1",
//						TxPool:  newMockTxPool(t),
//						TxScheduler: func() protocol.TxScheduler {
//							txScheduler = newMockTxScheduler(t)
//							txScheduler.EXPECT().Schedule(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
//							return txScheduler
//						}(),
//						SnapshotManager: func() protocol.SnapshotManager {
//							snapshotManager := newMockSnapshotManager(t)
//							block0 := createNewTestBlock(0)
//							block1 := func() *commonpb.Block {
//								var hash = []byte("0123456789")
//								var block = &commonpb.Block{
//									Header: &commonpb.BlockHeader{
//										BlockVersion:   protocol.DefaultBlockVersion,
//										ChainId:        "chain1",
//										BlockHeight:    1,
//										PreBlockHash:   hash,
//										BlockTimestamp: time.Now().Unix(),
//										Proposer:       &pbac.Member{},
//									},
//									Dag: &commonpb.DAG{Vertexes: nil},
//								}
//								return block
//							}()
//
//							var (
//								snapshotFactory snapshot.Factory
//								log             = &test.GoLogger{}
//								ctl             = gomock.NewController(t)
//								store           = mock.NewMockBlockchainStore(ctl)
//							)
//							snapshotMgr := snapshotFactory.NewSnapshotManager(store, log)
//							mockSnapShot := snapshotMgr.NewSnapshot(block0, block1)
//							snapshotManager.EXPECT().NewSnapshot(block0, block1).Return(mockSnapShot).AnyTimes()
//							return snapshotManager
//						}(),
//						Identity: func() protocol.SigningMember {
//							signMember := newMockSigningMember(t)
//							signMember.EXPECT().GetMember().Return(&pbac.Member{}, nil).AnyTimes()
//							return signMember
//						}(),
//						LedgerCache: func() protocol.LedgerCache {
//							ledgerCache := newMockLedgerCache(t)
//							ledgerCache.EXPECT().GetLastCommittedBlock().Return(createNewTestBlock(0)).AnyTimes()
//							ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
//							ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
//							return ledgerCache
//						}(),
//						ProposalCache: func() protocol.ProposalCache {
//							proposalCache := newMockProposalCache(t)
//							proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).Return(createNewTestBlock(0), map[string]*commonpb.TxRWSet{}).AnyTimes()
//							return proposalCache
//						}(),
//						ChainConf:   chainConf,
//						Log:         newMockLogger(t),
//						StoreHelper: storeHelper,
//					}
//					blockBuilder := common.NewBlockBuilder(bbConf)
//					return blockBuilder
//				}(),
//				storeHelper: nil,
//			},
//			args: args{
//				proposingHeight: 0,
//				preHash:         []byte("123456"),
//				txBatch: []*commonpb.Transaction{
//					{
//						Payload: &commonpb.Payload{
//							TxId: "123456",
//						},
//					},
//				},
//			},
//			want:    nil,
//			want1:   []int64{0, 0, 0},
//			wantErr: true,
//		},
//		{
//			name: "test2",
//			fields: fields{
//				chainId:     "test2",
//				txScheduler: txScheduler,
//				blockBuilder: func() *common.BlockBuilder {
//					bbConf := &common.BlockBuilderConf{
//						ChainId:     "chain1",
//						TxPool:      newMockTxPool(t),
//						TxScheduler: txScheduler,
//						SnapshotManager: func() protocol.SnapshotManager {
//							snapshotManager := newMockSnapshotManager(t)
//							block0 := createNewTestBlock(0)
//
//							block1 := func() *commonpb.Block {
//
//								var hash = []byte("0123456789")
//								var block = &commonpb.Block{
//									Header: &commonpb.BlockHeader{
//										BlockVersion:   protocol.DefaultBlockVersion,
//										ChainId:        "chain1",
//										BlockHeight:    1,
//										PreBlockHash:   hash,
//										BlockTimestamp: time.Now().Unix(),
//										Proposer:       &pbac.Member{},
//									},
//									Dag: &commonpb.DAG{Vertexes: nil},
//								}
//
//								return block
//							}()
//							var (
//								snapshotFactory snapshot.Factory
//								log             = &test.GoLogger{}
//								ctl             = gomock.NewController(t)
//								store           = mock.NewMockBlockchainStore(ctl)
//							)
//							snapshotMgr := snapshotFactory.NewSnapshotManager(store, log)
//
//							mockSnapShot := snapshotMgr.NewSnapshot(block0, block1)
//							snapshotManager.EXPECT().NewSnapshot(block0, block1).Return(mockSnapShot).AnyTimes()
//							return snapshotManager
//						}(),
//						Identity: func() protocol.SigningMember {
//							signMember := newMockSigningMember(t)
//							signMember.EXPECT().GetMember().Return(&pbac.Member{}, nil).AnyTimes()
//							return signMember
//						}(),
//						LedgerCache: func() protocol.LedgerCache {
//							ledgerCache := newMockLedgerCache(t)
//							ledgerCache.EXPECT().GetLastCommittedBlock().Return(createNewTestBlock(0)).AnyTimes()
//							ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
//							ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil).AnyTimes()
//							return ledgerCache
//						}(),
//						ProposalCache: func() protocol.ProposalCache {
//							proposalCache := newMockProposalCache(t)
//							proposalCache.EXPECT().GetProposedBlockByHashAndHeight(gomock.Any(), gomock.Any()).Return(createNewTestBlock(0), map[string]*commonpb.TxRWSet{}).AnyTimes()
//							return proposalCache
//						}(),
//						ChainConf:   chainConf,
//						Log:         newMockLogger(t),
//						StoreHelper: storeHelper,
//					}
//
//					blockBuilder := common.NewBlockBuilder(bbConf)
//					return blockBuilder
//				}(),
//				storeHelper: nil,
//			},
//			args: args{
//				proposingHeight: 0,
//				preHash:         []byte("123456"),
//				txBatch: []*commonpb.Transaction{
//					{
//						Payload: &commonpb.Payload{
//							TxId:         "chain1",
//							ChainId:      "chain1",
//							ContractName: "fact",
//							Method:       "set",
//							Sequence:     1,
//						},
//						Sender: &commonpb.EndorsementEntry{},
//						Result: &commonpb.Result{
//							RwSetHash: []byte("0123456789"),
//						},
//					},
//				},
//			},
//			want:    nil,
//			want1:   []int64{0, 0, 0},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			bp := &BlockProposerImpl{
//				chainId:      tt.fields.chainId,
//				blockBuilder: tt.fields.blockBuilder,
//				storeHelper:  tt.fields.storeHelper,
//				txScheduler:  tt.fields.txScheduler,
//			}
//			got, _, err := bp.generateNewBlock(tt.args.proposingHeight, tt.args.preHash, tt.args.txBatch, []string{})
//			if (err != nil) != tt.wantErr {
//				t.Errorf("generateNewBlock() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("generateNewBlock() got = %v, want %v", got, tt.want)
//			}
//			//if !reflect.DeepEqual(got1, tt.want1) {
//			//	t.Errorf("generateNewBlock() got1 = %v, want %v", got1, tt.want1)
//			//}
//		})
//	}
//}
