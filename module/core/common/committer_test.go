package common

//
//func TestCommitBlock_CommitBlock(t *testing.T) {
//
//	ctl := gomock.NewController(t)
//	log := logger.GetLoggerByChain(logger.MODULE_CORE, "chain1")
//	block := createNewTestBlock(12)
//
//	// snapshotManager
//	snapshotManager := mock.NewMockSnapshotManager(ctl)
//	snapshotManager.EXPECT().NotifyBlockCommitted(block).Return(nil)
//
//	// 	ledgerCache
//	ledgerCache := mock.NewMockLedgerCache(ctl)
//	ledgerCache.EXPECT().SetLastCommittedBlock(block)
//
//	// msgbus
//	msgbus := mbusmock.NewMockMessageBus(ctl)
//	//msgbus.EXPECT().PublishSafe(gomock.Any(), gomock.Any()).Return()
//
//	//chainConf mock
//	config := &config.ChainConfig{
//		ChainId: "chain1",
//		Crypto: &config.CryptoConfig{
//			Hash: "SHA256",
//		},
//		Block: &config.BlockConfig{
//			BlockTxCapacity: 1000,
//			BlockSize:       1,
//			BlockInterval:   DEFAULTDURATION,
//		},
//		Consensus: &config.ConsensusConfig{
//			Type: consensusPb.ConsensusType_RAFT,
//		},
//	}
//	chainConf := mock.NewMockChainConf(ctl)
//	chainConf.EXPECT().ChainConfig().AnyTimes().Return(config)
//
//	txRWSetMap := make(map[string]*commonpb.TxRWSet)
//	tx0 := block.Txs[0]
//	contractName := "testContract"
//	txRWSetMap[tx0.Payload.TxId] = &commonpb.TxRWSet{
//		TxId: tx0.Payload.TxId,
//		TxReads: []*commonpb.TxRead{{
//			ContractName: contractName,
//			Key:          []byte("K1"),
//			Value:        []byte("V"),
//		}},
//		TxWrites: []*commonpb.TxWrite{{
//			ContractName: contractName,
//			Key:          []byte("K2"),
//			Value:        []byte("V"),
//		}},
//	}
//
//	// Mock blockChain Store
//	store := mock.NewMockBlockchainStore(ctl)
//	txRWSets := []*commonpb.TxRWSet{
//		txRWSetMap[tx0.Payload.TxId],
//	}
//	log.Infof("init block(%d,%s)", block.Header.BlockHeight, hex.EncodeToString(block.Header.BlockHash))
//	store.EXPECT().PutBlock(block, txRWSets).Return(nil)
//
//	cbConf := &CommitBlockConf{
//		Store:           store,
//		Log:             log,
//		SnapshotManager: snapshotManager,
//		LedgerCache:     ledgerCache,
//		ChainConf:       chainConf,
//		MsgBus:          msgbus,
//	}
//
//	conEventMap := make(map[string][]*commonpb.ContractEvent)
//
//	commiter := NewCommitBlock(cbConf)
//	_, _, _, _, _, _, _, err := commiter.CommitBlock(block, txRWSetMap, conEventMap)
//	if err != nil {
//		panic(err)
//	}
//}
//
//func createNewTestBlock(height uint64) *commonpb.Block {
//	var hash = []byte("0123456789")
//	var block = &commonpb.Block{
//		Header: &commonpb.BlockHeader{
//			ChainId:        "Chain1",
//			BlockHeight:    height,
//			PreBlockHash:   hash,
//			BlockHash:      hash,
//			PreConfHeight:  0,
//			BlockVersion:   1,
//			DagHash:        hash,
//			RwSetRoot:      hash,
//			TxRoot:         hash,
//			BlockTimestamp: 0,
//			Proposer: &accesscontrol.Member{
//				OrgId:      "org1",
//				MemberType: 0,
//				MemberInfo: nil,
//			},
//			TxCount:   0,
//			Signature: nil,
//		},
//		Dag: &commonpb.DAG{
//			Vertexes: nil,
//		},
//		Txs: nil,
//	}
//	tx := createNewTestTx("0123456789")
//	txs := make([]*commonpb.Transaction, 1)
//	txs[0] = tx
//	block.Txs = txs
//	return block
//}
//
//func TestCommitBlock_MonitorCommit(t *testing.T) {
//	type fields struct {
//		store                 protocol.BlockchainStore
//		log                   protocol.Logger
//		snapshotManager       protocol.SnapshotManager
//		ledgerCache           protocol.LedgerCache
//		chainConf             protocol.ChainConf
//		msgBus                msgbus.MessageBus
//		metricBlockSize       *prometheus.HistogramVec
//		metricBlockCounter    *prometheus.CounterVec
//		metricTxCounter       *prometheus.CounterVec
//		metricBlockCommitTime *prometheus.HistogramVec
//	}
//	type args struct {
//		bi *commonpb.BlockInfo
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				store:                 newMockBlockchainStore(t),
//				log:                   newMockLogger(t),
//				snapshotManager:       newMockSnapshotManager(t),
//				ledgerCache:           newMockLedgerCache(t),
//				chainConf:             newMockChainConf(t),
//				msgBus:                msgbus.NewMessageBus(),
//				metricBlockSize:       nil,
//				metricBlockCounter:    nil,
//				metricTxCounter:       nil,
//				metricBlockCommitTime: nil,
//			},
//			args: args{
//				bi: &commonpb.BlockInfo{
//					Block:     createBlock(0),
//					RwsetList: RearrangeRWSet(createBlock(0), map[string]*commonpb.TxRWSet{}),
//				},
//			},
//			wantErr: false,
//		},
//		//{
//		//	name:    "test1", // TODO monitor
//		//	fields:  fields{
//		//		store:                 newMockBlockchainStore(t),
//		//		log:                   newMockLogger(t),
//		//		snapshotManager:       newMockSnapshotManager(t),
//		//		ledgerCache:           newMockLedgerCache(t),
//		//		chainConf:             newMockChainConf(t),
//		//		msgBus:                msgbus.NewMessageBus(),
//		//		metricBlockSize:       nil,
//		//		metricBlockCounter:    nil,
//		//		metricTxCounter:       nil,
//		//		metricBlockCommitTime: nil,
//		//	},
//		//	args:    args{
//		//		bi: &commonpb.BlockInfo{
//		//			Block: func() *commonpb.Block {
//		//				localconf.ChainMakerConfig.MonitorConfig.Enabled = true
//		//
//		//				block := createBlock(0)
//		//				return block
//		//			}(),
//		//			RwsetList: RearrangeRWSet(createBlock(0), map[string]*commonpb.TxRWSet{}),
//		//		},
//		//	},
//		//	wantErr: false,
//		//},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			cb := &CommitBlock{
//				store:                 tt.fields.store,
//				log:                   tt.fields.log,
//				snapshotManager:       tt.fields.snapshotManager,
//				ledgerCache:           tt.fields.ledgerCache,
//				chainConf:             tt.fields.chainConf,
//				msgBus:                tt.fields.msgBus,
//				metricBlockSize:       tt.fields.metricBlockSize,
//				metricBlockCounter:    tt.fields.metricBlockCounter,
//				metricTxCounter:       tt.fields.metricTxCounter,
//				metricBlockCommitTime: tt.fields.metricBlockCommitTime,
//			}
//			cb.MonitorCommit(tt.args.bi)
//		})
//	}
//}
//
//func TestNotifyChainConf(t *testing.T) {
//	type args struct {
//		block     *commonpb.Block
//		chainConf protocol.ChainConf
//	}
//
//	block := createBlock(0)
//	block.Header.ConsensusArgs = []byte("test123456")
//	block.Txs = []*commonpb.Transaction{
//		{
//			Payload: &commonpb.Payload{
//				TxId: "123456",
//			},
//		},
//	}
//
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			args: args{
//				block:     createBlock(0),
//				chainConf: newMockChainConf(t),
//			},
//			wantErr: false,
//		},
//		{
//			name: "test1",
//			args: args{
//				block: block,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensusPb.ConsensusType_DPOS,
//						},
//					}
//
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					chainConf.EXPECT().CompleteBlock(block).Return(nil).AnyTimes()
//					return chainConf
//				}(),
//			},
//			wantErr: false,
//		},
//		{
//			name: "test1",
//			args: args{
//				block: block,
//				chainConf: func() protocol.ChainConf {
//					chainConf := newMockChainConf(t)
//					chainConfig := &configpb.ChainConfig{
//						Consensus: &configpb.ConsensusConfig{
//							Type: consensusPb.ConsensusType_DPOS,
//						},
//					}
//					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
//					chainConf.EXPECT().CompleteBlock(block).Return(errors.New("chainconf block complete")).AnyTimes()
//					return chainConf
//				}(),
//			},
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := NotifyChainConf(tt.args.block, tt.args.chainConf); (err != nil) != tt.wantErr {
//				t.Errorf("NotifyChainConf() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_rearrangeContractEvent(t *testing.T) {
//	type args struct {
//		block       *commonpb.Block
//		conEventMap map[string][]*commonpb.ContractEvent
//	}
//	tests := []struct {
//		name string
//		args args
//		want []*commonpb.ContractEvent
//	}{
//		{
//			name: "test0",
//			args: args{
//				block:       createBlock(0),
//				conEventMap: nil,
//			},
//			want: make([]*commonpb.ContractEvent, 0),
//		},
//		{
//			name: "test1",
//			args: args{
//				block: func() *commonpb.Block {
//					block := createBlock(0)
//					block.Txs = []*commonpb.Transaction{
//						{
//							Payload: &commonpb.Payload{
//								TxId: "123456",
//							},
//						},
//					}
//					return block
//				}(),
//				conEventMap: map[string][]*commonpb.ContractEvent{
//					"test": {
//						{
//							TxId: "123456",
//						},
//					},
//				},
//			},
//			want: make([]*commonpb.ContractEvent, 0),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := rearrangeContractEvent(tt.args.block, tt.args.conEventMap); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("rearrangeContractEvent() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
