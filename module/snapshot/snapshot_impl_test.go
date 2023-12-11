/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"chainmaker.org/chainmaker/logger/v2"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	vmPb "chainmaker.org/chainmaker/pb-go/v2/vm"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/test"
	uatomic "go.uber.org/atomic"
	uberAtomic "go.uber.org/atomic"
)

var _ protocol.TxSimContext = (*MockSimContextImpl)(nil)

// Storage interface for smart contracts
type MockSimContextImpl struct {
	txExecSeq    int32
	tx           *commonPb.Transaction
	txRwSet      *commonPb.TxRWSet
	currentDepth int
	txResult     *commonPb.Result
	gasRemaining uint64
}

func (s *MockSimContextImpl) SubtractGas(gasUsed uint64) error {
	if int64(s.gasRemaining) < 0 {
		return errors.New("gas is not enough")
	}
	s.gasRemaining -= gasUsed
	if int64(s.gasRemaining) < 0 {
		return errors.New("gas is not enough")
	}
	return nil
}

func (s *MockSimContextImpl) GetGasRemaining() uint64 {
	//TODO implement me
	panic("implement me")
}

func (s *MockSimContextImpl) GetBlockFingerprint() string {
	//TODO implement me
	panic("implement me")
}

func (s *MockSimContextImpl) GetKeys(keys []*vmPb.BatchKey) ([]*vmPb.BatchKey, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MockSimContextImpl) GetBlockTimestamp() int64 {
	panic("implement me")
}

func (s *MockSimContextImpl) GetHistoryIterForKey(contractName string, key []byte) (protocol.KeyHistoryIterator, error) {
	panic("implement me")
}

func (s *MockSimContextImpl) SetIterHandle(index int32, iter interface{}) {
	panic("implement me")
}

func (s *MockSimContextImpl) GetIterHandle(index int32) (interface{}, bool) {
	panic("implement me")
}

func (s *MockSimContextImpl) PutIntoReadSet(contractName string, key []byte, value []byte) {
	panic("implement me")
}

func (s *MockSimContextImpl) GetContractByName(name string) (*commonPb.Contract, error) {
	panic("implement me")
}

func (s *MockSimContextImpl) GetContractBytecode(name string) ([]byte, error) {
	panic("implement me")
}

const implement_me = "implement me"

func (s *MockSimContextImpl) GetTxExecSeq() int {
	return int(s.txExecSeq)
}

func (s *MockSimContextImpl) GetDepth() int {
	return s.currentDepth
}
func (s *MockSimContextImpl) GetBlockVersion() uint32 {
	return protocol.DefaultBlockVersion
}

func (s *MockSimContextImpl) CallContract(caller *commonPb.Contract, contract *commonPb.Contract, method string, byteCode []byte,
	parameter map[string][]byte, gasUsed uint64, refTxType commonPb.TxType) (*commonPb.ContractResult,
	protocol.ExecOrderTxType, commonPb.TxStatusCode) {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetCurrentResult() []byte {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetCreator(namespace string) *acPb.Member {
	panic(implement_me)
}

func (s *MockSimContextImpl) Select(namespace string, startKey []byte, limit []byte) (protocol.StateIterator, error) {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetBlockHeight() uint64 {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetBlockProposer() *acPb.Member {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetTxResult() *commonPb.Result {
	return s.txResult
}

func (s *MockSimContextImpl) SetTxResult(result *commonPb.Result) {
	s.txResult = result
}

func (s *MockSimContextImpl) GetSender() *acPb.Member {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetBlockchainStore() protocol.BlockchainStore {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetLastChainConfig() *config.ChainConfig {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetAccessControl() (protocol.AccessControlProvider, error) {
	panic(implement_me)
}

func (s *MockSimContextImpl) GetChainNodesInfoProvider() (protocol.ChainNodesInfoProvider, error) {
	panic(implement_me)
}

// StateDB & ReadWriteSet
// 获取合约账户状态、Code
func (s *MockSimContextImpl) Get(contractName string, key []byte) ([]byte, error) {
	return nil, nil
}
func (s *MockSimContextImpl) GetNoRecord(contractName string, key []byte) ([]byte, error) {
	return nil, nil
}

// 写入合约账户状态
func (s *MockSimContextImpl) Put(contractName string, key []byte, value []byte) error {
	return nil
}

func (s *MockSimContextImpl) PutRecord(contractName string, value []byte, sqlType protocol.SqlType) {
}

// 删除合约账户状态
func (s *MockSimContextImpl) Del(contractName string, key []byte) error {
	return nil
}

func (s *MockSimContextImpl) Done() bool {
	return true
}

func (s *MockSimContextImpl) GetTx() *commonPb.Transaction {
	return s.tx
}
func (s *MockSimContextImpl) GetTxRWSet(runVmSuccess bool) *commonPb.TxRWSet {
	return s.txRwSet
}
func (s *MockSimContextImpl) SetTxExecSeq(txExecSeq int) {
	s.txExecSeq = int32(txExecSeq)
}

func (s *MockSimContextImpl) SetStateSqlHandle(index int32, rows protocol.SqlRows) {
	panic("impl me")

}

func (s *MockSimContextImpl) GetStateSqlHandle(index int32) (protocol.SqlRows, bool) {
	panic("impl me")
}

func (s *MockSimContextImpl) GetStateKvHandle(index int32) (protocol.StateIterator, bool) {
	panic("impl me")
}

func (s *MockSimContextImpl) SetStateKvHandle(index int32, rows protocol.StateIterator) {
	panic("impl me")
}

func (s *MockSimContextImpl) GetTxRWMapByContractName(contractName string) (map[string]*commonPb.TxRead, map[string]*commonPb.TxWrite) {
	panic("impl me")
}
func (s *MockSimContextImpl) GetCrossInfo() uint64 {
	panic("impl me")
}
func (s *MockSimContextImpl) HasUsed(runtimeType commonPb.RuntimeType) bool {
	panic("impl me")
}
func (s *MockSimContextImpl) RecordRuntimeTypeIntoCrossInfo(runtimeType commonPb.RuntimeType) {
	panic("impl me")
}
func (s *MockSimContextImpl) RemoveRuntimeTypeFromCrossInfo() {
	panic("impl me")
}

// GetStrAddrFromPbMember calculate string address from pb Member
func (s *MockSimContextImpl) GetStrAddrFromPbMember(pbMember *acPb.Member) (string, error) {
	panic("impl me")
}

// GetSnapshot
func (s *MockSimContextImpl) GetSnapshot() protocol.Snapshot {
	panic("impl me")
}

func TestKey(t *testing.T) {
	s0 := "你好"
	b0 := []byte(s0)
	s1 := string(b0)
	println("s0 equal s1 ", s0 == s1)
	println("len s1 ", len(s1))

	s2 := string([]byte{0x00, 0x01, 0x02})
	s3 := string([]byte{0x00, 0x01, 0x02})

	println("s2 equal s3 ", s2 == s3)
	println("len s2", len(s2))
}

func TestSnapshot(t *testing.T) {
	for i := 0; i < 20; i++ {
		testSnapshot(t, i)
	}
}

func testSnapshot(t *testing.T, i int) {
	snapshot := &SnapshotImpl{
		lock:            sync.RWMutex{},
		blockchainStore: nil,
		sealed:          uberAtomic.NewBool(false),
		chainId:         "",
		blockTimestamp:  0,
		blockProposer:   nil,
		blockHeight:     100,
		blockVersion:    210,
		preSnapshot:     nil,
		txRWSetTable:    nil,
		txTable:         make([]*commonPb.Transaction, 0, 2048),
		txResultMap:     make(map[string]*commonPb.Result, 256),
		readTable:       make(map[string]*sv, 256),
		writeTable:      make(map[string]*sv, 256),
		log:             &test.GoLogger{},
	}

	txSimContext := &MockSimContextImpl{
		tx: &commonPb.Transaction{
			Payload: &commonPb.Payload{
				TxId: "tx id in snapshot",
			},
		},
		txResult: &commonPb.Result{},
	}

	txCount := 4000
	start := time.Now()

	var count int64
	wg := sync.WaitGroup{}

	for i := 0; i < txCount; i++ {
		wg.Add(1)
		go func() {
			//fmt.Printf("tx:%d\t", i)
			readKey := randKey()
			writeKey := randKey()
			txSimContext.txRwSet = genRwSet(readKey, writeKey)
			// TODO: Use of weak random number generator (math/rand instead of crypto/rand) ?
			txSimContext.txExecSeq = int32(rand.Intn(len(snapshot.txTable) + 1)) //nolint: gosec

			applyResult, _ := snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
			atomic.AddInt64(&count, 1)
			if !applyResult {
				fmt.Printf("!!!")
				for {

					randNum := len(snapshot.txTable) - int(txSimContext.txExecSeq) + 1

					if randNum > 0 {
						txSimContext.txRwSet = genRwSet(readKey, writeKey)
						// TODO: Use of weak random number generator (math/rand instead of crypto/rand) ?
						// nolint: gosec
						txSimContext.txExecSeq = txSimContext.txExecSeq +
							int32(
								rand.Intn(
									randNum,
								),
							)
						applyResult, _ = snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)

						atomic.AddInt64(&count, 1)
						if applyResult {
							break
						}
					}

				}
			}
			wg.Done()
		}()

		////fmt.Printf("apply read write set %v, size %d, txExecSeq %d, ", applyResult, len(snapshot.txTable), txExecSeq)
		//for _, txRead := range readKey {
		//	fmt.Printf("%s ", txRead)
		//}
		//fmt.Print("\t")
		//for _, txWrite := range writeKey {
		//	fmt.Printf("%s ", txWrite)
		//}
		//fmt.Println("")
	}
	wg.Wait()
	timeCost := time.Since(start)
	snapshot.Seal()
	//dump(snapshot)
	//dumpDAG(snapshot.BuildDAG())

	fmt.Printf("Cost:%v, count:%d\n", timeCost, count)
}

func randKey() []string {
	kRange := 1000000000
	// TODO: Use of weak random number generator (math/rand instead of crypto/rand) ?
	size := rand.Intn(5) + 1 //nolint: gosec

	var keySlice []string
	for i := 0; i < size; i++ {
		// TODO: Use of weak random number generator (math/rand instead of crypto/rand) ?
		kId := rand.Intn(kRange) //nolint: gosec
		key := "K" + strconv.Itoa(kId)
		keySlice = append(keySlice, key)
	}
	return keySlice
}

func genRwSet(readKeySet []string, writeKeySet []string) *commonPb.TxRWSet {
	var txReads []*commonPb.TxRead
	var txWrites []*commonPb.TxWrite
	for _, key := range readKeySet {
		txRead := &commonPb.TxRead{
			Key:     []byte(key),
			Value:   []byte("value"),
			Version: nil,
		}
		txReads = append(txReads, txRead)
	}
	for _, key := range writeKeySet {
		txWrite := &commonPb.TxWrite{
			Key:   []byte(key),
			Value: []byte("value"),
		}
		txWrites = append(txWrites, txWrite)
	}

	txRwSet := &commonPb.TxRWSet{
		TxReads:  txReads,
		TxWrites: txWrites,
	}

	return txRwSet
}

func testApply(txSimContext protocol.TxSimContext, snapshot *SnapshotImpl, txExecSeq int, readKeySet []string, writeKeySet []string) (bool, int) {
	return snapshot.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
}

func dump(snapshot *SnapshotImpl) {
	fmt.Printf("tableSize %+v\n", len(snapshot.txTable))
	fmt.Printf("txTable %+v\n", snapshot.txTable)
	fmt.Printf("readTable %+v\n", snapshot.readTable)
	fmt.Printf("writeTable %+v\n", snapshot.writeTable)
}

func dumpDAG(dag *commonPb.DAG) {
	fmt.Println("digraph DAG {")
	for i, ns := range dag.Vertexes {
		if len(ns.Neighbors) == 0 {
			fmt.Printf("tx%d -> begin%d\n", i, 0)
			continue
		}
		for _, n := range ns.Neighbors {
			fmt.Printf("tx%d -> tx%d\n", i, n)
		}
	}
	fmt.Println("}")
}

var snapshot = &SnapshotImpl{
	lock:            sync.RWMutex{},
	blockchainStore: nil,
	sealed:          uatomic.NewBool(false),
	chainId:         "",
	blockTimestamp:  0,
	blockProposer:   nil,
	blockHeight:     100,
	preSnapshot:     nil,
	txRWSetTable:    nil,
	txTable:         make([]*commonPb.Transaction, 0, 2048),
	txResultMap:     make(map[string]*commonPb.Result, 256),
	readTable:       make(map[string]*sv, 256),
	writeTable:      make(map[string]*sv, 256),
	log:             &test.GoLogger{},
}

type fields struct {
	//lock            sync.RWMutex
	blockchainStore protocol.BlockchainStore
	log             protocol.Logger
	sealed          *uatomic.Bool
	chainId         string
	blockTimestamp  int64
	blockProposer   *acPb.Member
	blockHeight     uint64
	blockVersion    uint32
	preBlockHash    []byte
	preSnapshot     protocol.Snapshot
	txRWSetTable    []*commonPb.TxRWSet
	txTable         []*commonPb.Transaction
	specialTxTable  []*commonPb.Transaction
	txResultMap     map[string]*commonPb.Result
	readTable       map[string]*sv
	writeTable      map[string]*sv
	txRoot          []byte
	dagHash         []byte
	rwSetHash       []byte
}
type args struct {
	isSql bool
}

func TestSnapshotImpl_BuildDAG(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *commonPb.DAG
	}{
		// TODO: Add test cases.
		{
			name: "txRWSetTable",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "11111111",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key1"),
								Value: []byte("value of key1"),
							},
						},
					},
					{
						TxId: "222222222",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key1"),
								Value: []byte("value of key1"),
							},
						},
					},
					{
						TxId: "333333333",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key2"),
								Value: []byte("value of key2"),
							},
						},
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
				},
				blockHeight:  1,
				blockVersion: 221,
				txTable: []*commonPb.Transaction{
					{},
					{},
					{},
				},
				log: logger.GetLogger("test"),
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{0, 1},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SnapshotImpl{
				lock:            sync.RWMutex{},
				blockchainStore: tt.fields.blockchainStore,
				log:             tt.fields.log,
				sealed:          tt.fields.sealed,
				chainId:         tt.fields.chainId,
				blockTimestamp:  tt.fields.blockTimestamp,
				blockProposer:   tt.fields.blockProposer,
				blockHeight:     tt.fields.blockHeight,
				preBlockHash:    tt.fields.preBlockHash,
				preSnapshot:     tt.fields.preSnapshot,
				txRWSetTable:    tt.fields.txRWSetTable,
				txTable:         tt.fields.txTable,
				specialTxTable:  tt.fields.specialTxTable,
				txResultMap:     tt.fields.txResultMap,
				readTable:       tt.fields.readTable,
				writeTable:      tt.fields.writeTable,
				txRoot:          tt.fields.txRoot,
				dagHash:         tt.fields.dagHash,
				rwSetHash:       tt.fields.rwSetHash,
			}
			if got := s.BuildDAG(tt.args.isSql, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildDAG() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReBuildDag(t *testing.T) {
	tests := []struct {
		name               string
		fields             fields
		args               args
		blockDagRwSetTable []*commonPb.TxRWSet
		want               *commonPb.DAG
	}{
		{
			name: "differentOrderOfApply",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "3",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key2"),
								Value: []byte("value of key2"),
							},
						},
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
					{
						TxId: "1",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key1"),
								Value: []byte("value of key1"),
							},
						},
					},
					{
						TxId: "2",
						TxReads: []*commonPb.TxRead{
							{
								Key:   []byte("key1"),
								Value: []byte("value of key1"),
							},
						},
					},
				},
				log:    logger.GetLogger("test"),
				sealed: uatomic.NewBool(false),
			},
			blockDagRwSetTable: []*commonPb.TxRWSet{
				{
					TxId: "1",
					TxReads: []*commonPb.TxRead{
						{
							Key:   []byte("key1"),
							Value: []byte("value of key1"),
						},
					},
				},
				{
					TxId: "2",
					TxReads: []*commonPb.TxRead{
						{
							Key:   []byte("key1"),
							Value: []byte("value of key1"),
						},
					},
				},
				{
					TxId: "3",
					TxReads: []*commonPb.TxRead{
						{
							Key:   []byte("key2"),
							Value: []byte("value of key2"),
						},
					},
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{0, 1},
					},
				},
			},
		},
		{
			name: "differentOrderOfApply2",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "3",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
					{
						TxId: "1",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
					{
						TxId: "2",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
				},
				log:    logger.GetLogger("test"),
				sealed: uatomic.NewBool(false),
			},
			blockDagRwSetTable: []*commonPb.TxRWSet{
				{
					TxId: "1",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "2",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "3",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{0},
					},
					{
						Neighbors: []uint32{1},
					},
				},
			},
		},
		{
			name: "differentOrderOfApply3",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "3",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
					{
						TxId: "2",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
					{
						TxId: "1",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
				},
				log:    logger.GetLogger("test"),
				sealed: uatomic.NewBool(false),
			},
			blockDagRwSetTable: []*commonPb.TxRWSet{
				{
					TxId: "1",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "2",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "3",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{0},
					},
					{
						Neighbors: []uint32{1},
					},
				},
			},
		},
		{
			name: "orderNotMatter",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "3",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key3"),
								Value: []byte("new value of key3"),
							},
						},
					},
					{
						TxId: "1",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key2"),
							},
						},
					},
					{
						TxId: "2",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key2"),
								Value: []byte("new value of key1"),
							},
						},
					},
				},
				log:    logger.GetLogger("test"),
				sealed: uatomic.NewBool(false),
			},
			blockDagRwSetTable: []*commonPb.TxRWSet{
				{
					TxId: "1",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "2",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key2"),
							Value: []byte("new value of key2"),
						},
					},
				},
				{
					TxId: "3",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key3"),
							Value: []byte("new value of key3"),
						},
					},
				},
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
				},
			},
		},
		{
			name: "orderNotMatter2",
			fields: fields{
				txRWSetTable: []*commonPb.TxRWSet{
					{
						TxId: "3",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key3"),
								Value: []byte("new value of key3"),
							},
						},
					},
					{
						TxId: "2",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key2"),
								Value: []byte("new value of key2"),
							},
						},
					},
					{
						TxId: "1",
						TxWrites: []*commonPb.TxWrite{
							{
								Key:   []byte("key1"),
								Value: []byte("new value of key1"),
							},
						},
					},
				},
				log:    logger.GetLogger("test"),
				sealed: uatomic.NewBool(false),
			},
			blockDagRwSetTable: []*commonPb.TxRWSet{
				{
					TxId: "1",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key1"),
							Value: []byte("new value of key1"),
						},
					},
				},
				{
					TxId: "2",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key2"),
							Value: []byte("new value of key2"),
						},
					},
				},
				{
					TxId: "3",
					TxWrites: []*commonPb.TxWrite{
						{
							Key:   []byte("key3"),
							Value: []byte("new value of key3"),
						},
					},
				},
			},
			args: args{
				isSql: false,
			},
			want: &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SnapshotImpl{
				lock:            sync.RWMutex{},
				blockchainStore: tt.fields.blockchainStore,
				log:             tt.fields.log,
				sealed:          tt.fields.sealed,
				chainId:         tt.fields.chainId,
				blockTimestamp:  tt.fields.blockTimestamp,
				blockProposer:   tt.fields.blockProposer,
				blockHeight:     tt.fields.blockHeight,
				preBlockHash:    tt.fields.preBlockHash,
				preSnapshot:     tt.fields.preSnapshot,
				txRWSetTable:    []*commonPb.TxRWSet{},
				txTable:         []*commonPb.Transaction{},
				specialTxTable:  tt.fields.specialTxTable,
				txResultMap:     map[string]*commonPb.Result{},
				readTable:       map[string]*sv{},
				writeTable:      map[string]*sv{},
				txRoot:          tt.fields.txRoot,
				dagHash:         tt.fields.dagHash,
				rwSetHash:       tt.fields.rwSetHash,
			}
			for i, t := range tt.fields.txRWSetTable {
				txSimContext := &MockSimContextImpl{
					txExecSeq:    int32(i),
					tx:           &commonPb.Transaction{Payload: &commonPb.Payload{TxId: t.GetTxId()}},
					txRwSet:      t,
					currentDepth: 0,
					txResult:     nil,
				}
				s.ApplyTxSimContext(txSimContext, protocol.ExecOrderTxTypeNormal, true, false)
			}
			if got := s.BuildDAG(tt.args.isSql, tt.blockDagRwSetTable); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildDAG() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstructKey(t *testing.T) {
	contractName := "save"
	key := []byte("name")
	fmt.Println(constructKey(contractName, key))
}
