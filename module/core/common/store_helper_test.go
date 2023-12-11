/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package common

import (
	"errors"
	"reflect"
	"runtime"
	"testing"

	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
)

/*
 * test unit NewKVStoreHelper func
 */
func TestNewKVStoreHelper(t *testing.T) {
	type args struct {
		chainId string
	}
	tests := []struct {
		name string
		args args
		want *KVStoreHelper
	}{
		{
			name: "test0",
			args: args{
				chainId: "123456",
			},
			want: &KVStoreHelper{
				chainId: "123456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKVStoreHelper(tt.args.chainId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKVStoreHelper() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit KVStoreHelper RollBack func
 */
func TestKVStoreHelper_RollBack(t *testing.T) {
	type fields struct {
		chainId string
	}
	type args struct {
		block           *commonpb.Block
		blockchainStore protocol.BlockchainStore
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
				chainId: "123456",
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := &KVStoreHelper{
				chainId: tt.fields.chainId,
			}
			if err := kv.RollBack(tt.args.block, tt.args.blockchainStore); (err != nil) != tt.wantErr {
				t.Errorf("RollBack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

/*
 * test unit KVStoreHelper BeginDbTransaction func
 */
func TestKVStoreHelper_BeginDbTransaction(t *testing.T) {
	type fields struct {
		chainId string
	}
	type args struct {
		blockchainStore protocol.BlockchainStore
		txKey           string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				chainId: "123456",
			},
			args: args{
				blockchainStore: nil,
				txKey:           "",
			},
		},
	}

	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := &KVStoreHelper{
				chainId: tt.fields.chainId,
			}
			kv.RollBack(createBlock(0), blockchainStore)
		})
	}
}

/*
 * test unit KVStoreHelper GetPoolCapacity func
 */
func TestKVStoreHelper_GetPoolCapacity(t *testing.T) {
	type fields struct {
		chainId string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test0",
			fields: fields{
				chainId: "123456",
			},
			want: runtime.NumCPU() * 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := &KVStoreHelper{
				chainId: tt.fields.chainId,
			}
			if got := kv.GetPoolCapacity(); got != tt.want {
				t.Errorf("GetPoolCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit NewSQLStoreHelper func
 */
func TestNewSQLStoreHelper(t *testing.T) {
	type args struct {
		chainId string
	}
	tests := []struct {
		name string
		args args
		want *SQLStoreHelper
	}{
		{
			name: "test0",
			args: args{
				chainId: "123456",
			},
			want: &SQLStoreHelper{
				chainId: "123456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSQLStoreHelper(tt.args.chainId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLStoreHelper() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit SQLStoreHelper RollBack func
 */
func TestSQLStoreHelper_RollBack(t *testing.T) {
	type fields struct {
		chainId string
	}
	type args struct {
		block           *commonpb.Block
		blockchainStore protocol.BlockchainStore
	}

	ctrl := gomock.NewController(t)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				chainId: "123456",
			},
			args: args{
				block: createBlock(0),
				blockchainStore: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					blockChainStore.EXPECT().RollbackDbTransaction("ee29eb4a8725678278ac439cf7abfd2a849cdc7378a6b6316017b81c51d720e7").Return(nil).AnyTimes()
					return blockChainStore
				}(),
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				chainId: "123456",
			},
			args: args{
				block: createBlock(0),
				blockchainStore: func() protocol.BlockchainStore {
					blockChainStore := mock.NewMockBlockchainStore(ctrl)
					blockChainStore.EXPECT().RollbackDbTransaction("ee29eb4a8725678278ac439cf7abfd2a849cdc7378a6b6316017b81c51d720e7").Return(errors.New("data is nil")).AnyTimes()
					return blockChainStore
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := &SQLStoreHelper{
				chainId: tt.fields.chainId,
			}
			if err := sql.RollBack(tt.args.block, tt.args.blockchainStore); (err != nil) != tt.wantErr {
				t.Errorf("RollBack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

/*
 * test unit SQLStoreHelper BeginDbTransaction func
 */
func TestSQLStoreHelper_BeginDbTransaction(t *testing.T) {
	type fields struct {
		chainId string
	}
	type args struct {
		blockchainStore protocol.BlockchainStore
		txKey           string
	}
	ctrl := gomock.NewController(t)
	blockChainStore := mock.NewMockBlockchainStore(ctrl)
	sqlDBTransaction := mock.NewMockSqlDBTransaction(ctrl)
	blockChainStore.EXPECT().BeginDbTransaction(gomock.Any()).Return(sqlDBTransaction, nil).AnyTimes()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				chainId: "123456",
			},
			args: args{
				blockchainStore: blockChainStore,
				txKey:           "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := &SQLStoreHelper{
				chainId: tt.fields.chainId,
			}
			sql.BeginDbTransaction(tt.args.blockchainStore, tt.args.txKey)
		})
	}
}

/*
 * test unit SQLStoreHelper GetPoolCapacity func
 */
func TestSQLStoreHelper_GetPoolCapacity(t *testing.T) {
	type fields struct {
		chainId string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test0",
			fields: fields{
				chainId: "123456",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := &SQLStoreHelper{
				chainId: tt.fields.chainId,
			}
			if got := sql.GetPoolCapacity(); got != tt.want {
				t.Errorf("GetPoolCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}
