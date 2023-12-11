/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cache

import (
	"reflect"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/stretchr/testify/require"
)

func TestLedger(t *testing.T) {
	ledgerCache := NewLedgerCache("Chain1")
	ledgerCache.SetLastCommittedBlock(CreateNewTestBlock(0))
	modulea := &moduleA{}
	moduleb := &moduleB{
		ledgerCache: ledgerCache,
	}
	modulea.setLedgerCache(ledgerCache)
	b := ledgerCache.GetLastCommittedBlock()
	b.Header.BlockHeight = 100
	ledgerCache.SetLastCommittedBlock(b)
	require.Equal(t, uint64(100), modulea.getBlock().Header.BlockHeight)
	b = modulea.getBlock()
	b.Header.BlockHeight = 200
	modulea.updateBlock(b)
	require.Equal(t, uint64(200), moduleb.getBlock().Header.BlockHeight)
}

type moduleA struct {
	ledgerCache protocol.LedgerCache
}

func (m *moduleA) setLedgerCache(cache protocol.LedgerCache) {
	m.ledgerCache = cache
}

func (m *moduleA) updateBlock(block *commonpb.Block) {
	m.ledgerCache.SetLastCommittedBlock(block)
}

func (m *moduleA) getBlock() *commonpb.Block {
	return m.ledgerCache.GetLastCommittedBlock()
}

type moduleB struct {
	ledgerCache protocol.LedgerCache
}

func (m *moduleB) updateBlock(block *commonpb.Block) {
	m.ledgerCache.SetLastCommittedBlock(block)
}

func (m *moduleB) getBlock() *commonpb.Block {
	return m.ledgerCache.GetLastCommittedBlock()
}

func CreateNewTestBlock(height uint64) *commonpb.Block {
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
	tx := CreateNewTestTx()
	txs := make([]*commonpb.Transaction, 1)
	txs[0] = tx
	block.Txs = txs
	return block
}

func CreateNewTestBlockByHash(height uint64, hash []byte) *commonpb.Block {
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
	tx := CreateNewTestTx()
	txs := make([]*commonpb.Transaction, 1)
	txs[0] = tx
	block.Txs = txs
	return block
}

func CreateNewTestTx() *commonpb.Transaction {
	var hash = []byte("0123456789")
	return &commonpb.Transaction{
		Payload: &commonpb.Payload{
			ChainId: "",
			//Sender:         nil,
			TxType:         0,
			TxId:           "",
			Timestamp:      0,
			ExpirationTime: 0,
		},
		//RequestPayload:   hash,
		//RequestSignature: hash,
		Sender: &commonpb.EndorsementEntry{Signature: hash},
		Result: &commonpb.Result{
			Code:           commonpb.TxStatusCode_SUCCESS,
			ContractResult: nil,
			RwSetHash:      nil,
		},
	}
}

/*
 * test unit ledger cache NewLedgerCache func
 */
func TestNewLedgerCache(t *testing.T) {
	type args struct {
		chainId string
	}
	tests := []struct {
		name string
		args args
		want protocol.LedgerCache
	}{
		{
			name: "test0", // test case 0,chainId equal chainId
			args: args{
				chainId: "12345",
			},
			want: &LedgerCache{
				chainId: "12345",
			},
		},
		{
			name: "test1", // test case 1,chainId equal chainId
			args: args{
				chainId: "6789",
			},
			want: &LedgerCache{
				chainId: "6789",
			},
		},
		{
			name: "test2", // test case 2,chainId equal chainId
			args: args{
				chainId: "abcd",
			},
			want: &LedgerCache{
				chainId: "abcd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLedgerCache(tt.args.chainId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLedgerCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ledger cache GetLastCommittedBlock func
 */
func TestLedgerCache_GetLastCommittedBlock(t *testing.T) {
	type fields struct {
		chainId            string
		lastCommittedBlock *commonpb.Block
	}
	tests := []struct {
		name   string
		fields fields
		want   *commonpb.Block
	}{
		{
			name: "test0", // test case 0,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "11111",
				lastCommittedBlock: CreateNewTestBlock(0),
			},
			want: CreateNewTestBlock(0),
		},
		{
			name: "test1", // test case 1,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "123456",
				lastCommittedBlock: CreateNewTestBlock(1),
			},
			want: CreateNewTestBlock(1),
		},
		{
			name: "test2", // test case 2,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "7890",
				lastCommittedBlock: CreateNewTestBlock(2),
			},
			want: CreateNewTestBlock(2),
		},
		{
			name: "test3", // test case 3,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "abcd",
				lastCommittedBlock: CreateNewTestBlock(3),
			},
			want: CreateNewTestBlock(3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &LedgerCache{
				chainId:            tt.fields.chainId,
				lastCommittedBlock: tt.fields.lastCommittedBlock,
			}
			if got := lc.GetLastCommittedBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastCommittedBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ledger cache SetLastCommittedBlock func
 */
func TestLedgerCache_SetLastCommittedBlock(t *testing.T) {
	type fields struct {
		chainId            string
		lastCommittedBlock *commonpb.Block
	}
	type args struct {
		b *commonpb.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0", // test case 0,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "11111",
				lastCommittedBlock: CreateNewTestBlock(0),
			},
			args: args{
				b: CreateNewTestBlock(uint64(0)),
			},
		},
		{
			name: "test1", // test case 1,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "123456",
				lastCommittedBlock: CreateNewTestBlock(1),
			},
			args: args{
				b: CreateNewTestBlock(uint64(1)),
			},
		},
		{
			name: "test2", // test case 2,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "7890",
				lastCommittedBlock: CreateNewTestBlock(2),
			},
			args: args{
				b: CreateNewTestBlock(uint64(2)),
			},
		},
		{
			name: "test3", // test case 3,field lastCommittedBlock equal want block
			fields: fields{
				chainId:            "abcd",
				lastCommittedBlock: CreateNewTestBlock(3),
			},
			args: args{
				b: CreateNewTestBlock(uint64(3)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &LedgerCache{
				chainId:            tt.fields.chainId,
				lastCommittedBlock: tt.fields.lastCommittedBlock,
			}
			lc.SetLastCommittedBlock(tt.args.b)
		})
	}
}

/*
 * test unit ledger cache CurrentHeight func
 */
func TestLedgerCache_CurrentHeight(t *testing.T) {
	type fields struct {
		chainId            string
		lastCommittedBlock *commonpb.Block
	}
	tests := []struct {
		name    string
		fields  fields
		want    uint64
		wantErr bool
	}{
		{
			name: "test0", // test case 0, want 0, wantErr: false,
			fields: fields{
				chainId:            "11111",
				lastCommittedBlock: CreateNewTestBlock(0),
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "test1", // test case 1, want 1, wantErr: false,
			fields: fields{
				chainId:            "123456",
				lastCommittedBlock: CreateNewTestBlock(1),
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "test2", // test case 2, want 2, wantErr: false,
			fields: fields{
				chainId:            "7890",
				lastCommittedBlock: CreateNewTestBlock(2),
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "test3", // test case 3, want 3, wantErr: false,
			fields: fields{
				chainId:            "abcd",
				lastCommittedBlock: CreateNewTestBlock(3),
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &LedgerCache{
				chainId:            tt.fields.chainId,
				lastCommittedBlock: tt.fields.lastCommittedBlock,
			}
			got, err := lc.CurrentHeight()
			if (err != nil) != tt.wantErr {
				t.Errorf("CurrentHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CurrentHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}
