package cache

import (
	"reflect"
	"sync"
	"testing"

	"chainmaker.org/chainmaker/logger/v2"

	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

/*
 * test unit proposal cache ClearTheBlock func
 */
var (
	log = logger.GetLoggerByChain(logger.MODULE_CORE, "chain1")
)

func TestProposalCache_ClearTheBlock(t *testing.T) {

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil)

	proposalCache := NewProposalCache(chainConf, ledgerCache, log)

	rwSetMap := make(map[string]*commonpb.TxRWSet)
	contractEvenMap := make(map[string][]*commonpb.ContractEvent)

	hash := []byte("123")
	block := &commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockVersion:   1,
			BlockType:      0,
			ChainId:        "chain1",
			BlockHeight:    0,
			BlockHash:      hash,
			PreBlockHash:   nil,
			PreConfHeight:  0,
			TxCount:        0,
			TxRoot:         nil,
			DagHash:        nil,
			RwSetRoot:      nil,
			BlockTimestamp: 0,
			ConsensusArgs:  nil,
			Proposer:       nil,
			Signature:      nil,
		},
		Dag:            nil,
		Txs:            nil,
		AdditionalData: nil,
	}

	err := proposalCache.SetProposedBlock(block, rwSetMap, contractEvenMap, false)
	require.Nil(t, err)

	b0 := proposalCache.GetProposedBlocksAt(0)
	require.NotNil(t, b0)

	b1 := proposalCache.GetProposedBlocksAt(1)
	require.Nil(t, b1)

	b2, txRWSet := proposalCache.GetProposedBlockByHashAndHeight(hash, 0)
	require.NotNil(t, b2)
	require.NotNil(t, txRWSet)

	b3 := proposalCache.GetSelfProposedBlockAt(0)
	require.Nil(t, b3)

	proposalCache.ClearProposedBlockAt(0)

	b2_1, txRWSet_2 := proposalCache.GetProposedBlockByHashAndHeight(hash, 0)
	require.Nil(t, b2_1)
	require.Nil(t, txRWSet_2)

	b0_1 := proposalCache.GetProposedBlocksAt(0)
	require.Nil(t, b0_1)

}

/*
 * test unit proposal cache ClearProposedBlockAt func
 */
func TestProposalCache_ClearProposedBlockAt(t *testing.T) {

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	ledgerCache.EXPECT().CurrentHeight().Return(uint64(0), nil)

	proposalCache := NewProposalCache(chainConf, ledgerCache, log)

	rwSetMap := make(map[string]*commonpb.TxRWSet)
	contractEvenMap := make(map[string][]*commonpb.ContractEvent)

	hash := []byte("123")
	block := &commonpb.Block{
		Header: &commonpb.BlockHeader{
			BlockVersion:   1,
			BlockType:      0,
			ChainId:        "chain1",
			BlockHeight:    0,
			BlockHash:      hash,
			PreBlockHash:   nil,
			PreConfHeight:  0,
			TxCount:        0,
			TxRoot:         nil,
			DagHash:        nil,
			RwSetRoot:      nil,
			BlockTimestamp: 0,
			ConsensusArgs:  nil,
			Proposer:       nil,
			Signature:      nil,
		},
		Dag:            nil,
		Txs:            nil,
		AdditionalData: nil,
	}

	err := proposalCache.SetProposedBlock(block, rwSetMap, contractEvenMap, true)
	require.Nil(t, err)

	require.Equal(t, proposalCache.IsProposedAt(block.Header.BlockHeight), true)

	b0 := proposalCache.GetSelfProposedBlockAt(block.Header.BlockHeight)
	require.NotNil(t, b0)

	b1 := proposalCache.GetSelfProposedBlockAt(block.Header.BlockHeight + 1)
	require.Nil(t, b1)

	require.Equal(t, proposalCache.HasProposedBlockAt(block.Header.BlockHeight), true)

	//proposalCache.ResetProposedAt(block.Header.BlockHeight)
	//
	//b2 := proposalCache.GetSelfProposedBlockAt(block.Header.BlockHeight)
	//require.Nil(t, b2)
	//
	//proposalCache.SetProposedAt(block.Header.BlockHeight)
	//b3 := proposalCache.GetSelfProposedBlockAt(block.Header.BlockHeight)
	//require.Nil(t, b3)

	//proposalCache.ClearProposedBlockAt(0)
	//
	//b2_1, txRWSet_2 := proposalCache.GetProposedBlockByHashAndHeight(hash, 0)
	//require.Nil(t, b2_1)
	//require.Nil(t, txRWSet_2)
	//
	//b0_1 := proposalCache.GetProposedBlocksAt(0)
	//require.Nil(t, b0_1)

}

/*
 * test unit proposal cache NewProposalCache func
 */
func TestNewProposalCache(t *testing.T) {
	type args struct {
		chainConf   protocol.ChainConf
		ledgerCache protocol.LedgerCache
	}

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)

	tests := []struct {
		name string
		args args
		want protocol.ProposalCache
	}{
		{
			name: "test0",
			args: args{
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId: "12345",
				},
			},
			want: &ProposalCache{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId: "12345",
				},
				logger: log,
			},
		},
		{
			name: "test1",
			args: args{
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId:            "6789",
					lastCommittedBlock: CreateNewTestBlock(1),
				},
			},
			want: &ProposalCache{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId:            "6789",
					lastCommittedBlock: CreateNewTestBlock(1),
				},
				logger: log,
			},
		},
		{
			name: "test2",
			args: args{
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId:            "6789",
					lastCommittedBlock: CreateNewTestBlock(2),
					rwMu:               sync.RWMutex{},
				},
			},
			want: &ProposalCache{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId:            "6789",
					lastCommittedBlock: CreateNewTestBlock(2),
				},
				logger: log,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProposalCache(tt.args.chainConf, tt.args.ledgerCache, log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProposalCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit proposal cache SetProposedAt func
 */
func TestProposalCache_SetProposedAt(t *testing.T) {
	type fields struct {
		lastProposedBlock map[uint64]map[string]*blockProposal
		chainConf         protocol.ChainConf
		ledgerCache       protocol.LedgerCache
	}
	type args struct {
		height uint64
	}

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			args: args{
				height: 0,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"0": &blockProposal{
							block:          CreateNewTestBlock(uint64(0)),
							isSelfProposed: false,
						},
					},
				},
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId: "12345",
				},
			},
		},
		{
			name: "test1",
			args: args{
				height: 1,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"0": &blockProposal{
							block:          CreateNewTestBlock(uint64(1)),
							isSelfProposed: true,
						},
					},
				},
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId: "6789",
				},
			},
		},
		{
			name: "test2",
			args: args{
				height: 2,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId: "abcd",
				},
			},
		},
		{
			name: "test3",
			args: args{
				height: 3,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId: "defg",
				},
			},
		},
	}

	for height, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProposalCache{
				lastProposedBlock: tt.fields.lastProposedBlock,
				chainConf:         tt.fields.chainConf,
				ledgerCache:       tt.fields.ledgerCache,
			}
			pc.SetProposedAt(uint64(height))
		})
	}
}

/*
 * test unit proposal cache ResetProposedAt func
 */
func TestProposalCache_ResetProposedAt(t *testing.T) {
	type fields struct {
		lastProposedBlock map[uint64]map[string]*blockProposal
		chainConf         protocol.ChainConf
		ledgerCache       protocol.LedgerCache
	}
	type args struct {
		height uint64
	}

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			args: args{
				height: 0,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"0": &blockProposal{
							block:          CreateNewTestBlock(uint64(0)),
							isSelfProposed: true,
						},
					},
				},
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId: "abcd",
				},
			},
		},
		{
			name: "test1",
			args: args{
				height: 1,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"0": &blockProposal{
							block:          CreateNewTestBlock(uint64(1)),
							isSelfProposed: false,
						},
					},
				},
				chainConf: chainConf,
				ledgerCache: &LedgerCache{
					chainId: "abcd",
				},
			},
		},
		{
			name: "test2",
			args: args{
				height: 2,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId: "abcd",
				},
			},
		},
		{
			name: "test2",
			args: args{
				height: 2,
			},
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{},
				chainConf:         chainConf,
				ledgerCache: &LedgerCache{
					chainId: "abcd",
				},
			},
		},
	}
	for height, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProposalCache{
				lastProposedBlock: tt.fields.lastProposedBlock,
				chainConf:         tt.fields.chainConf,
				ledgerCache:       tt.fields.ledgerCache,
			}
			pc.ResetProposedAt(uint64(height))
		})
	}
}

/*
 * test unit proposal cache IsProposedAt func
 */
func TestProposalCache_IsProposedAt(t *testing.T) {
	type fields struct {
		lastProposedBlock map[uint64]map[string]*blockProposal
		chainConf         protocol.ChainConf
		ledgerCache       protocol.LedgerCache
	}
	type args struct {
		height uint64
	}

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"0": &blockProposal{
							block:                CreateNewTestBlock(uint64(0)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(0),
			},
			want: true,
		},
		{
			name: "test1",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"1": &blockProposal{
							block:          CreateNewTestBlock(uint64(2)),
							isSelfProposed: true,
						},
					},
				},
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(1),
			},
			want: false,
		},
		{
			name: "test2",
			fields: fields{
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(2),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProposalCache{
				lastProposedBlock: tt.fields.lastProposedBlock,
				chainConf:         tt.fields.chainConf,
				ledgerCache:       tt.fields.ledgerCache,
			}
			if got := pc.IsProposedAt(tt.args.height); got != tt.want {
				t.Errorf("IsProposedAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit proposal cache KeepProposedBlock func
 */
func TestProposalCache_KeepProposedBlock(t *testing.T) {
	type fields struct {
		lastProposedBlock map[uint64]map[string]*blockProposal
		chainConf         protocol.ChainConf
		ledgerCache       protocol.LedgerCache
	}
	type args struct {
		hash   []byte
		height uint64
	}

	ctl := gomock.NewController(t)
	chainConf := mock.NewMockChainConf(ctl)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*commonpb.Block
	}{
		{
			name: "test0",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"0": &blockProposal{
							block:                CreateNewTestBlockByHash(uint64(0), []byte("12345")),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(0),
			},
			want: []*commonpb.Block{
				CreateNewTestBlockByHash(uint64(0), []byte("12345")),
			},
		},
		{
			name: "test1",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"1": &blockProposal{
							block:                CreateNewTestBlockByHash(uint64(0), []byte("12345")),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(1),
				hash:   []byte("12345"),
			},
			want: []*commonpb.Block{},
		},
		{
			name: "test2",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"2": &blockProposal{
							block:                CreateNewTestBlockByHash(uint64(0), []byte("12345")),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf:   chainConf,
				ledgerCache: NewLedgerCache("12345"),
			},
			args: args{
				height: uint64(2),
				hash:   []byte("6789"),
			},
			want: []*commonpb.Block{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProposalCache{
				lastProposedBlock: tt.fields.lastProposedBlock,
				chainConf:         tt.fields.chainConf,
				ledgerCache:       tt.fields.ledgerCache,
				logger:            log,
			}
			if got := pc.KeepProposedBlock(tt.args.hash, tt.args.height); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeepProposedBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit proposal cache getHashType func
 */
func TestProposalCache_getHashType(t *testing.T) {
	type fields struct {
		lastProposedBlock map[uint64]map[string]*blockProposal
		chainConf         protocol.ChainConf
		ledgerCache       protocol.LedgerCache
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test0",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(0): {
						"0": &blockProposal{
							block:                CreateNewTestBlock(uint64(0)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf: func() protocol.ChainConf {
					var chainConfig = &configpb.ChainConfig{
						Crypto: &configpb.CryptoConfig{Hash: "SHA256"},
					}
					ctl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctl)
					chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
					return chainConf
				}(),
				ledgerCache: NewLedgerCache("12345"),
			},
			want: "SHA256",
		},
		{
			name: "test1",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"1": &blockProposal{
							block:                CreateNewTestBlock(uint64(1)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf: func() protocol.ChainConf {
					var chainConfig = &configpb.ChainConfig{
						Crypto: &configpb.CryptoConfig{Hash: "SM3"},
					}
					ctl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctl)
					chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
					return chainConf
				}(),
				ledgerCache: NewLedgerCache("12345"),
			},
			want: "SM3",
		},
		{
			name: "test1",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"1": &blockProposal{
							block:                CreateNewTestBlock(uint64(1)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf: func() protocol.ChainConf {
					var chainConfig = &configpb.ChainConfig{
						Crypto: &configpb.CryptoConfig{Hash: "SM3"},
					}
					ctl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctl)
					chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
					return chainConf
				}(),
				ledgerCache: NewLedgerCache("12345"),
			},
			want: "SM3",
		},
		{
			name: "test2",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"1": &blockProposal{
							block:                CreateNewTestBlock(uint64(1)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf: func() protocol.ChainConf {
					var chainConfig = &configpb.ChainConfig{
						Crypto: &configpb.CryptoConfig{Hash: "SHA_256"},
					}
					ctl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctl)
					chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
					return chainConf
				}(),
				ledgerCache: NewLedgerCache("12345"),
			},
			want: "SHA_256",
		},
		{
			name: "test2",
			fields: fields{
				lastProposedBlock: map[uint64]map[string]*blockProposal{
					uint64(1): {
						"1": &blockProposal{
							block:                CreateNewTestBlock(uint64(1)),
							isSelfProposed:       true,
							hasProposedThisRound: true,
						},
					},
				},
				chainConf: func() protocol.ChainConf {
					ctl := gomock.NewController(t)
					chainConf := mock.NewMockChainConf(ctl)

					var chainConfig = &configpb.ChainConfig{
						Crypto: &configpb.CryptoConfig{Hash: "SHA_256"},
					}
					chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
					return chainConf
				}(),
				ledgerCache: NewLedgerCache("12345"),
			},
			want: "SHA_256",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProposalCache{
				lastProposedBlock: tt.fields.lastProposedBlock,
				chainConf:         tt.fields.chainConf,
				ledgerCache:       tt.fields.ledgerCache,
			}
			if got := pc.getHashType(); got != tt.want {
				t.Errorf("getHashType() = %v, want %v", got, tt.want)
			}
		})
	}
}
