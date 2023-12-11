/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"chainmaker.org/chainmaker/logger/v2"

	"chainmaker.org/chainmaker/pb-go/v2/consensus"

	crypto2 "chainmaker.org/chainmaker/common/v2/crypto"

	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/localconf/v2"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/utils/v2"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

const (
	txId0 = "a0000000000000000000000000000000"
	txId1 = "a0000000000000000000000000000001"
	txId2 = "a0000000000000000000000000000002"
	txId3 = "a0000000000000000000000000000003"
	txId4 = "a0000000000000000000000000000004"
)

var (
	TestPrivKeyFile = "../../../../config/wx-org1/certs/node/consensus1/consensus1.sign.key"
	TestCertFile    = "../../../../config/wx-org1/certs/node/consensus1/consensus1.sign.crt"
)

//func TestDag(t *testing.T) {
//	for i := 0; i < 10; i++ {
//
//		neb1 := &commonPb.DAG_Neighbor{
//			Neighbors: []int32{1, 2, 3, 4},
//		}
//		neb2 := &commonPb.DAG_Neighbor{
//			Neighbors: []int32{1, 2, 3, 4},
//		}
//		neb3 := &commonPb.DAG_Neighbor{
//			Neighbors: []int32{1, 2, 3, 4},
//		}
//		vs := make([]*commonPb.DAG_Neighbor, 3)
//		vs[0] = neb1
//		vs[1] = neb2
//		vs[2] = neb3
//		dag := &commonPb.DAG{
//			Vertexes: vs,
//		}
//		marshal, _ := proto.Marshal(dag)
//		println("Dag", hex.EncodeToString(marshal))
//	}
//}
//
func newTx(txId string, contractId *commonPb.Contract, parameterMap map[string]string) *commonPb.Transaction {

	var parameters []*commonPb.KeyValuePair
	for key, value := range parameterMap {
		parameters = append(parameters, &commonPb.KeyValuePair{
			Key:   key,
			Value: []byte(value),
		})
	}

	return &commonPb.Transaction{
		Payload: &commonPb.Payload{
			ChainId:        "Chain1",
			TxType:         0,
			TxId:           txId,
			ContractName:   contractId.Name,
			Method:         "method",
			Parameters:     parameters,
			Timestamp:      0,
			ExpirationTime: 0,
			Limit:          &commonPb.Limit{GasLimit: 0},
		},
		Result: &commonPb.Result{
			Code: commonPb.TxStatusCode_SUCCESS,
			ContractResult: &commonPb.ContractResult{
				Code:          0,
				Result:        nil,
				Message:       "",
				GasUsed:       0,
				ContractEvent: nil,
			},
			RwSetHash: nil,
		},
		Sender: &commonPb.EndorsementEntry{Signer: &acPb.Member{OrgId: "org1", MemberInfo: []byte("-----BEGIN CERTIFICATE-----\nMIICdDCCAhqgAwIBAgIDDqVbMAoGCCqGSM49BAMCMIGIMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEeMBwGA1UEChMVd3gt\nb3JnLmNoYWlubWFrZXIub3JnMRIwEAYDVQQLEwlyb290LWNlcnQxITAfBgNVBAMT\nGGNhLnd4LW9yZy5jaGFpbm1ha2VyLm9yZzAeFw0yMjA4MjMwMzExNTlaFw0yNzA4\nMjIwMzExNTlaMIGPMQswCQYDVQQGEwJDTjEQMA4GA1UECBMHQmVpamluZzEQMA4G\nA1UEBxMHQmVpamluZzEeMBwGA1UEChMVd3gtb3JnLmNoYWlubWFrZXIub3JnMQ8w\nDQYDVQQLEwZjbGllbnQxKzApBgNVBAMTImNsaWVudDEuc2lnbi53eC1vcmcuY2hh\naW5tYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASk26B+moZKChBk\nhoZ35x21a9YkRGldmVn32pDPSBKYgML57B7aJke0Sh38ULVZUfHMsYin4sYBbyY9\njqQaIR5bo2owaDAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIEIOYu3/LFYp+U2FaF\nQehWEd9wYY4zx4hiDb/YhYysSA+7MCsGA1UdIwQkMCKAIEKT6aAzVPy7Fhu6s1GJ\nWFp2pZRwxoHdLivN2/18vPckMAoGCCqGSM49BAMCA0gAMEUCIQDOFsdUTe9XQ8iz\nfEjdeQGYyS80PCFnobU70bfFbGu9bQIgNLcZhs1nG6RM/bEdwDQeSSJExBK9NkuC\nch16zyC5Krk=\n-----END CERTIFICATE-----")},
			Signature: []byte("sign1"),
		},
	}

}

func newTxWithPubKeyAndGasLimit(txId string, contractId *commonPb.Contract, parameterMap map[string]string, gasLimit uint64) *commonPb.Transaction {

	var parameters []*commonPb.KeyValuePair
	for key, value := range parameterMap {
		parameters = append(parameters, &commonPb.KeyValuePair{
			Key:   key,
			Value: []byte(value),
		})
	}

	return &commonPb.Transaction{
		Payload: &commonPb.Payload{
			ChainId:        "Chain1",
			TxType:         0,
			TxId:           txId,
			ContractName:   contractId.Name,
			Method:         "method",
			Parameters:     parameters,
			Timestamp:      0,
			ExpirationTime: 0,
			Limit:          &commonPb.Limit{GasLimit: gasLimit},
		},
		Result: &commonPb.Result{
			Code: commonPb.TxStatusCode_SUCCESS,
			ContractResult: &commonPb.ContractResult{
				Code:          0,
				Result:        nil,
				Message:       "",
				GasUsed:       0,
				ContractEvent: nil,
			},
			RwSetHash: nil,
		},
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_PUBLIC_KEY,
				MemberInfo: []byte("-----BEGIN PUBLIC KEY-----\nMIIBCgKCAQEAvIU7PHVzanE3V6GHHS5OQLYRAh8gjKIzSVI+UKPRcy6hB8u/z7Is\n2oNPeOLW/N9umreCgi1nBhcjczOlbpIzq8YIMP/7HN3gnyPpsSp4y6GelKzl0YNy\nAN5huqyNU8dn2Du0xFeyzK6UGqmKb9Le1nfLZq6YtVB0NEfPfxzkTG15RrJg/eRn\nc0Lywl8tMwAptRE3ZJA791/aEJWdJLB52vqhM+fGn5+ol6OO/0mQAHdopIutYrZI\nzvM9GBZHdDEdz3f+44IRmc9qmzhoEEp5epD2LJDCtfNnwbKP/cwBaTMNCMqSibA4\nlMMMSwU88dmY6ZH4RCxDXaI9suMGzFh/fwIDAQAB\n-----END PUBLIC KEY-----"),
			},
			Signature: []byte("sign1"),
		},
	}

}

func newBlock() *commonPb.Block {
	return &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:        "",
			BlockHeight:    10,
			PreBlockHash:   nil,
			BlockHash:      nil,
			BlockVersion:   2300,
			DagHash:        nil,
			RwSetRoot:      nil,
			TxRoot:         nil,
			BlockTimestamp: 0,
			Proposer:       nil,
			ConsensusArgs:  nil,
			TxCount:        0,
			Signature:      nil,
		},
		Dag: &commonPb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
		AdditionalData: &commonPb.AdditionalData{
			ExtraData: nil,
		},
	}
}

func prepare(t *testing.T, enableSenderGroup, enableConflictsBitWindow bool, txCount int, setVM bool) (*mock.MockVmManager, []*commonPb.TxRWSet, []*commonPb.Transaction,
	*mock.MockSnapshot, protocol.TxScheduler, *commonPb.Contract, *commonPb.Block) {
	var txRWSetTable = make([]*commonPb.TxRWSet, txCount)
	for i := 0; i < txCount; i++ {
		txRWSetTable[i] = &commonPb.TxRWSet{TxId: fmt.Sprintf("a000000000000000000000000000%04d", i)}
	}
	var txTable = make([]*commonPb.Transaction, txCount)

	ctl := gomock.NewController(t)
	snapshot := mock.NewMockSnapshot(ctl)
	vmMgr := mock.NewMockVmManager(ctl)
	vmMgr.EXPECT().BeforeSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	vmMgr.EXPECT().AfterSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	chainConf := mock.NewMockChainConf(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	crypto := configpb.CryptoConfig{
		Hash: crypto2.CRYPTO_ALGO_SHA256,
	}
	contractConf := configpb.ContractConfig{EnableSqlSupport: false}
	chainConfig := &configpb.ChainConfig{
		Crypto:   &crypto,
		Contract: &contractConf,
		Core: &configpb.CoreConfig{
			EnableSenderGroup:        enableSenderGroup,
			EnableConflictsBitWindow: enableConflictsBitWindow,
		},
		AuthType: protocol.Identity,
		Vm: &configpb.Vm{
			AddrType: configpb.AddrType_CHAINMAKER,
		},
		Consensus: &configpb.ConsensusConfig{Type: consensus.ConsensusType_TBFT},
	}
	chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()

	storeHelper := mock.NewMockStoreHelper(ctl)
	storeHelper.EXPECT().GetPoolCapacity().Return(runtime.NumCPU() * 4).AnyTimes()
	var schedulerFactory TxSchedulerFactory
	scheduler := schedulerFactory.NewTxScheduler(vmMgr, chainConf, storeHelper, ledgerCache)
	contractId := &commonPb.Contract{
		Name:        "ContractName",
		Version:     "1",
		RuntimeType: commonPb.RuntimeType_WASMER,
	}

	contractResult := &commonPb.ContractResult{
		Code:    0,
		Result:  nil,
		Message: "",
	}
	block := newBlock()

	snapshot.EXPECT().GetTxTable().AnyTimes().Return(txTable)
	snapshot.EXPECT().GetTxRWSetTable().AnyTimes().Return(txRWSetTable)
	snapshot.EXPECT().GetSnapshotSize().AnyTimes().Return(len(txTable))
	snapshot.EXPECT().GetSpecialTxTable().AnyTimes().Return([]*commonPb.Transaction{})
	snapshot.EXPECT().GetBlockFingerprint().AnyTimes().Return(strconv.FormatUint(block.Header.BlockHeight, 10))
	snapshot.EXPECT().GetLastChainConfig().Return(chainConfig).AnyTimes()
	blockChainStore := mock.NewMockBlockchainStore(ctl)
	blockChainStore.EXPECT().GetContractByName(contractId.Name).Return(contractId, nil).AnyTimes()
	blockChainStore.EXPECT().GetContractBytecode(contractId.Name).AnyTimes()
	ledgerCache.EXPECT().CurrentHeight().Return(block.Header.BlockHeight-1, nil).AnyTimes()

	snapshot.EXPECT().GetBlockchainStore().AnyTimes().Return(blockChainStore)
	//snapshot.EXPECT().Seal()

	if setVM {
		vmMgr.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS)
	}
	return vmMgr, txRWSetTable, txTable, snapshot, scheduler, contractId, block
}

// prepare4 is used only by TestSchedule4
func prepare4(t *testing.T, enableOptimizeChargeGas, enableSenderGroup, enableConflictsBitWindow bool, txCount int, setVM bool) (
	*mock.MockVmManager, []*commonPb.TxRWSet, []*commonPb.Transaction,
	*mock.MockSnapshot, protocol.TxScheduler, *commonPb.Contract, *commonPb.Block) {
	var txRWSetTable = make([]*commonPb.TxRWSet, txCount)
	for i := 0; i < txCount; i++ {
		txRWSetTable[i] = &commonPb.TxRWSet{TxId: fmt.Sprintf("a000000000000000000000000000%04d", i)}
	}
	var txTable = make([]*commonPb.Transaction, txCount)

	ctl := gomock.NewController(t)
	snapshot := mock.NewMockSnapshot(ctl)
	vmMgr := mock.NewMockVmManager(ctl)
	vmMgr.EXPECT().BeforeSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	vmMgr.EXPECT().AfterSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	chainConf := mock.NewMockChainConf(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	crypto := configpb.CryptoConfig{
		Hash: crypto2.CRYPTO_ALGO_SHA256,
	}
	contractConf := configpb.ContractConfig{EnableSqlSupport: false}
	chainConfig := &configpb.ChainConfig{
		Crypto:   &crypto,
		Contract: &contractConf,
		AuthType: protocol.Identity,
		Core: &configpb.CoreConfig{
			EnableOptimizeChargeGas:  enableOptimizeChargeGas,
			EnableSenderGroup:        enableSenderGroup,
			EnableConflictsBitWindow: enableConflictsBitWindow,
		},
		AccountConfig: &configpb.GasAccountConfig{
			EnableGas: true,
		},
		Vm: &configpb.Vm{
			AddrType: configpb.AddrType_ZXL,
		},
		Consensus: &configpb.ConsensusConfig{Type: consensus.ConsensusType_TBFT},
	}
	chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)

	storeHelper := mock.NewMockStoreHelper(ctl)
	storeHelper.EXPECT().GetPoolCapacity().Return(runtime.NumCPU() * 4).AnyTimes()
	var schedulerFactory TxSchedulerFactory
	scheduler := schedulerFactory.NewTxScheduler(vmMgr, chainConf, storeHelper, ledgerCache)
	contractId := &commonPb.Contract{
		Name:        "ContractName",
		Version:     "1",
		RuntimeType: commonPb.RuntimeType_WASMER,
	}

	sysContractId := &commonPb.Contract{
		Name:        syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		Version:     "1",
		RuntimeType: commonPb.RuntimeType_NATIVE,
	}

	contractResult := &commonPb.ContractResult{
		Code:    0,
		Result:  nil,
		Message: "",
	}
	block := newBlock()

	snapshot.EXPECT().GetTxTable().AnyTimes().Return(txTable)
	snapshot.EXPECT().GetTxRWSetTable().AnyTimes().Return(txRWSetTable)
	snapshot.EXPECT().GetSnapshotSize().AnyTimes().Return(len(txTable))
	snapshot.EXPECT().GetSpecialTxTable().AnyTimes().Return([]*commonPb.Transaction{})
	snapshot.EXPECT().GetBlockFingerprint().AnyTimes().Return(strconv.FormatUint(block.Header.BlockHeight, 10))
	snapshot.EXPECT().GetKey(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]byte("1000000000"), nil)
	snapshot.EXPECT().GetLastChainConfig().Return(chainConfig).AnyTimes()
	blockChainStore := mock.NewMockBlockchainStore(ctl)
	// simulate calling GetContractByName(...) 3 times
	blockChainStore.EXPECT().GetContractByName(gomock.Eq(contractId.Name)).Return(contractId, nil).AnyTimes()
	blockChainStore.EXPECT().GetContractByName(gomock.Eq(sysContractId.Name)).Return(sysContractId, nil).AnyTimes()
	blockChainStore.EXPECT().GetContractBytecode(contractId.Name).AnyTimes()
	blockChainStore.EXPECT().GetContractBytecode(sysContractId.Name).AnyTimes()
	ledgerCache.EXPECT().CurrentHeight().Return(block.Header.BlockHeight-1, nil).AnyTimes()

	snapshot.EXPECT().GetBlockchainStore().AnyTimes().Return(blockChainStore)
	//snapshot.EXPECT().Seal()

	if setVM {
		vmMgr.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS)
	}
	return vmMgr, txRWSetTable, txTable, snapshot, scheduler, contractId, block
}

// prepare5 is used only by TestSchedule5
func prepare5(t *testing.T, enableOptimizeChargeGas, enableSenderGroup, enableConflictsBitWindow bool, txCount int, setVM bool) (
	*mock.MockVmManager, []*commonPb.TxRWSet, []*commonPb.Transaction,
	*mock.MockSnapshot, protocol.TxScheduler, *commonPb.Contract, *commonPb.Block) {
	var txRWSetTable = make([]*commonPb.TxRWSet, txCount)
	for i := 0; i < txCount; i++ {
		txRWSetTable[i] = &commonPb.TxRWSet{TxId: fmt.Sprintf("a000000000000000000000000000%04d", i)}
	}
	var txTable = make([]*commonPb.Transaction, txCount)

	ctl := gomock.NewController(t)
	snapshot := mock.NewMockSnapshot(ctl)
	vmMgr := mock.NewMockVmManager(ctl)
	vmMgr.EXPECT().BeforeSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	vmMgr.EXPECT().AfterSchedule(gomock.Any(), gomock.Any()).Return().AnyTimes()
	chainConf := mock.NewMockChainConf(ctl)
	ledgerCache := mock.NewMockLedgerCache(ctl)
	crypto := configpb.CryptoConfig{
		Hash: crypto2.CRYPTO_ALGO_SHA256,
	}
	contractConf := configpb.ContractConfig{EnableSqlSupport: false}
	chainConfig := &configpb.ChainConfig{
		Crypto:   &crypto,
		Contract: &contractConf,
		AuthType: protocol.Identity,
		Core: &configpb.CoreConfig{
			EnableOptimizeChargeGas:  enableOptimizeChargeGas,
			EnableSenderGroup:        enableSenderGroup,
			EnableConflictsBitWindow: enableConflictsBitWindow,
		},
		AccountConfig: &configpb.GasAccountConfig{
			EnableGas: true,
		},
		Vm: &configpb.Vm{
			AddrType: configpb.AddrType_ZXL,
		},
		Consensus: &configpb.ConsensusConfig{Type: consensus.ConsensusType_TBFT},
	}
	chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)

	storeHelper := mock.NewMockStoreHelper(ctl)
	storeHelper.EXPECT().GetPoolCapacity().Return(runtime.NumCPU() * 4).AnyTimes()
	var schedulerFactory TxSchedulerFactory
	scheduler := schedulerFactory.NewTxScheduler(vmMgr, chainConf, storeHelper, ledgerCache)
	contractId := &commonPb.Contract{
		Name:        "ContractName",
		Version:     "1",
		RuntimeType: commonPb.RuntimeType_WASMER,
	}

	sysContractId := &commonPb.Contract{
		Name:        syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		Version:     "1",
		RuntimeType: commonPb.RuntimeType_NATIVE,
	}

	contractResult := &commonPb.ContractResult{
		Code:    0,
		Result:  nil,
		Message: "",
	}
	block := newBlock()

	snapshot.EXPECT().GetTxTable().AnyTimes().Return(txTable)
	snapshot.EXPECT().GetTxRWSetTable().AnyTimes().Return(txRWSetTable)
	snapshot.EXPECT().GetSnapshotSize().AnyTimes().Return(len(txTable))
	snapshot.EXPECT().GetSpecialTxTable().AnyTimes().Return([]*commonPb.Transaction{})
	snapshot.EXPECT().GetBlockFingerprint().AnyTimes().Return(strconv.FormatUint(block.Header.BlockHeight, 10))
	snapshot.EXPECT().GetKey(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]byte("1000000000"), nil)
	snapshot.EXPECT().GetLastChainConfig().Return(chainConfig).AnyTimes()
	blockChainStore := mock.NewMockBlockchainStore(ctl)
	// simulate calling GetContractByName(...) 3 times
	blockChainStore.EXPECT().GetContractByName(gomock.Eq(contractId.Name)).Return(contractId, nil).AnyTimes()
	blockChainStore.EXPECT().GetContractByName(gomock.Eq(sysContractId.Name)).Return(sysContractId, nil).AnyTimes()
	blockChainStore.EXPECT().GetContractBytecode(contractId.Name).AnyTimes()
	blockChainStore.EXPECT().GetContractBytecode(sysContractId.Name).AnyTimes()
	ledgerCache.EXPECT().CurrentHeight().Return(block.Header.BlockHeight-1, nil).AnyTimes()

	snapshot.EXPECT().GetBlockchainStore().AnyTimes().Return(blockChainStore)
	//snapshot.EXPECT().Seal()

	if setVM {
		vmMgr.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS)
	}
	return vmMgr, txRWSetTable, txTable, snapshot, scheduler, contractId, block
}

func TestSchedule(t *testing.T) {

	_, txRWSetTable, txTable, snapshot, scheduler, contractId, block := prepare(t, false, false, 2, true)

	parameters := make(map[string]string, 8)
	tx0 := newTx("a0000000000000000000000000000001", contractId, parameters)
	tx1 := newTx("a0000000000000000000000000000002", contractId, parameters)

	txTable[0] = tx0
	txTable[1] = tx1
	txRWSetTable[0] = &commonPb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonPb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K2"),
			Value:        []byte("V"),
		}},
	}
	txRWSetTable[1] = &commonPb.TxRWSet{
		TxId: tx1.Payload.TxId,
		TxReads: []*commonPb.TxRead{
			{
				ContractName: contractId.Name,
				Key:          []byte("K2"),
				Value:        []byte("V"),
			},
			{
				ContractName: contractId.Name,
				Key:          []byte("K2"),
				Value:        []byte("V"),
			},
		},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K3"),
			Value:        []byte("V"),
		}},
	}

	snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 2).AnyTimes()
	snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
	snapshot.EXPECT().Seal().Return()

	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{{}},
	}
	snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).Return(dag)

	txBatch := []*commonPb.Transaction{tx0, tx1}
	txSet, contractEven, err := scheduler.Schedule(block, txBatch, snapshot)
	require.Nil(t, err)
	require.NotNil(t, txSet)
	require.NotNil(t, contractEven)

	fmt.Println(txSet)
	fmt.Println(contractEven)
}

func TestSchedule2(t *testing.T) {

	_, txRWSetTable, txTable, snapshot, scheduler, contractId, block := prepare(t, true, false, 1, true)

	parameters := make(map[string]string, 8)
	tx0 := newTx("a0000000000000000000000000000001", contractId, parameters)
	//tx1 := newTx("a0000000000000000000000000000002", contractId, parameters)

	txTable[0] = tx0
	//txTable[1] = tx1
	txRWSetTable[0] = &commonPb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonPb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K2"),
			Value:        []byte("V"),
		}},
	}
	//txRWSetTable[1] = &commonPb.TxRWSet{
	//	TxId: tx1.Payload.TxId,
	//	TxReads: []*commonPb.TxRead{
	//		{
	//			ContractName: contractId.Name,
	//			Key:          []byte("K2"),
	//			Value:        []byte("V"),
	//		},
	//		{
	//			ContractName: contractId.Name,
	//			Key:          []byte("K2"),
	//			Value:        []byte("V"),
	//		},
	//	},
	//	TxWrites: []*commonPb.TxWrite{{
	//		ContractName: contractId.Name,
	//		Key:          []byte("K3"),
	//		Value:        []byte("V"),
	//	}},
	//}

	snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 1).AnyTimes()
	snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
	snapshot.EXPECT().Seal().Return()

	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{{}},
	}
	snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).Return(dag)

	txBatch := []*commonPb.Transaction{tx0}
	txSet, contractEven, err := scheduler.Schedule(block, txBatch, snapshot)
	require.Nil(t, err)
	require.NotNil(t, txSet)
	require.NotNil(t, contractEven)

	fmt.Println(txSet)
	fmt.Println(contractEven)
}

func TestSchedule3(t *testing.T) {

	_, txRWSetTable, txTable, snapshot, scheduler, contractId, block := prepare(t, true, true, 1, true)

	parameters := make(map[string]string, 8)
	tx0 := newTx("a0000000000000000000000000000001", contractId, parameters)
	//tx1 := newTx("a0000000000000000000000000000002", contractId, parameters)

	txTable[0] = tx0
	//txTable[1] = tx1
	txRWSetTable[0] = &commonPb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonPb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K2"),
			Value:        []byte("V"),
		}},
	}
	//txRWSetTable[1] = &commonPb.TxRWSet{
	//	TxId: tx1.Payload.TxId,
	//	TxReads: []*commonPb.TxRead{
	//		{
	//			ContractName: contractId.Name,
	//			Key:          []byte("K2"),
	//			Value:        []byte("V"),
	//		},
	//		{
	//			ContractName: contractId.Name,
	//			Key:          []byte("K2"),
	//			Value:        []byte("V"),
	//		},
	//	},
	//	TxWrites: []*commonPb.TxWrite{{
	//		ContractName: contractId.Name,
	//		Key:          []byte("K3"),
	//		Value:        []byte("V"),
	//	}},
	//}

	snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 1).AnyTimes()
	snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
	snapshot.EXPECT().Seal().Return()

	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{{}},
	}
	snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).Return(dag)

	txBatch := []*commonPb.Transaction{tx0}
	txSet, contractEven, err := scheduler.Schedule(block, txBatch, snapshot)
	require.Nil(t, err)
	require.NotNil(t, txSet)
	require.NotNil(t, contractEven)

	fmt.Println(txSet)
	fmt.Println(contractEven)
}

// TestSchedule4 test the flag `enableOptimizeChargeGas` is opened.
func TestSchedule4(t *testing.T) {

	fmt.Println("===== TestSchedule4() begin ==== ")
	localconf.ChainMakerConfig.NodeConfig.PrivKeyFile = TestPrivKeyFile
	localconf.ChainMakerConfig.NodeConfig.CertFile = TestCertFile
	localconf.ChainMakerConfig.NodeConfig.PrivKeyPassword = "11111111"
	_, txRWSetTable, txTable, snapshot, scheduler, contractId, block := prepare4(t, true, false, false, 2, true)

	parameters := make(map[string]string, 8)
	tx0 := newTxWithPubKeyAndGasLimit("a0000000000000000000000000000001", contractId, parameters, 101)
	tx1 := newTxWithPubKeyAndGasLimit("a0000000000000000000000000000002", contractId, parameters, 102)

	txTable[0] = tx0
	txTable[1] = tx1
	txRWSetTable[0] = &commonPb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonPb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V1"),
		}},
	}
	txRWSetTable[1] = &commonPb.TxRWSet{
		TxId: tx1.Payload.TxId,
		TxReads: []*commonPb.TxRead{
			{
				ContractName: contractId.Name,
				Key:          []byte("K1"),
				Value:        []byte("V"),
			},
			{
				ContractName: contractId.Name,
				Key:          []byte("K2"),
				Value:        []byte("V"),
			},
		},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K2"),
			Value:        []byte("V2"),
		}},
	}

	// simulate Calling ApplyTxSimContext(...) 3 times
	preCall1 := snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 1).Times(1)
	preCall2 := snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).After(preCall1).Return(true, 2).Times(1)
	snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).After(preCall2).Return(true, 3).Times(1)
	snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
	snapshot.EXPECT().Seal().Return()

	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{{}},
	}
	snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).Return(dag)

	txBatch := []*commonPb.Transaction{tx0, tx1}
	txSet, contractEven, err := scheduler.Schedule(block, txBatch, snapshot)
	require.Nil(t, err)
	require.NotNil(t, txSet)
	require.NotNil(t, contractEven)
}

// TestSchedule5 test the conflictsBitWindows features under flag `enableOptimizeChargeGas` is opened.
func TestSchedule5(t *testing.T) {

	fmt.Println("===== TestSchedule5() begin ==== ")
	localconf.ChainMakerConfig.NodeConfig.PrivKeyFile = TestPrivKeyFile
	localconf.ChainMakerConfig.NodeConfig.CertFile = TestCertFile
	localconf.ChainMakerConfig.NodeConfig.PrivKeyPassword = "11111111"
	_, txRWSetTable, txTable, snapshot, scheduler, contractId, block := prepare5(t, true, false, true, 2, true)

	parameters := make(map[string]string, 8)
	tx0 := newTxWithPubKeyAndGasLimit("a0000000000000000000000000000001", contractId, parameters, 101)
	tx1 := newTxWithPubKeyAndGasLimit("a0000000000000000000000000000002", contractId, parameters, 102)

	txTable[0] = tx0
	txTable[1] = tx1
	txRWSetTable[0] = &commonPb.TxRWSet{
		TxId: tx0.Payload.TxId,
		TxReads: []*commonPb.TxRead{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V"),
		}},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K1"),
			Value:        []byte("V1"),
		}},
	}
	txRWSetTable[1] = &commonPb.TxRWSet{
		TxId: tx1.Payload.TxId,
		TxReads: []*commonPb.TxRead{
			{
				ContractName: contractId.Name,
				Key:          []byte("K3"),
				Value:        []byte("V"),
			},
			{
				ContractName: contractId.Name,
				Key:          []byte("K4"),
				Value:        []byte("V"),
			},
		},
		TxWrites: []*commonPb.TxWrite{{
			ContractName: contractId.Name,
			Key:          []byte("K3"),
			Value:        []byte("V3"),
		}},
	}

	// simulate Calling ApplyTxSimContext(...) 3 times
	preCall1 := snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, 1).Times(1)
	preCall2 := snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).After(preCall1).Return(true, 2).Times(1)
	snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).After(preCall2).Return(true, 3).Times(1)
	snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
	snapshot.EXPECT().Seal().Return()

	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{{}},
	}
	snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).Return(dag)

	txBatch := []*commonPb.Transaction{tx0, tx1}
	txSet, contractEven, err := scheduler.Schedule(block, txBatch, snapshot)
	require.Nil(t, err)
	require.NotNil(t, txSet)
	require.NotNil(t, contractEven)
}

func TestSimulateWithDag(t *testing.T) {

	dagNormal := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: nil,
			},
			{
				Neighbors: []uint32{0},
			},
			{
				Neighbors: []uint32{1},
			},
		},
	}
	dagDupVertex := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: nil,
			},
			{
				//malformed dag, should cause error
				Neighbors: []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			{
				Neighbors: []uint32{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			},
		},
	}
	dagCycle := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: []uint32{2},
			},
			{
				Neighbors: []uint32{0},
			},
			{
				Neighbors: []uint32{1},
			},
		},
	}
	dagOutOfBound := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{0},
			},
			{
				Neighbors: []uint32{3},
			},
		},
	}
	dagMissingVertex := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: []uint32{},
			},
		},
	}
	applyTxSimContextNormal := func(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
		runVmSuccess bool, applySpecialTx bool) (bool, int) {
		switch txSimContext.GetTx().Payload.TxId {
		case txId0:
			return true, 1
		case txId1:
			return true, 2
		case txId2:
			return true, 3
		default:
			panic("Test shouldn't reach here")
		}
	}
	contractResult := &commonPb.ContractResult{
		Code:    0,
		Result:  nil,
		Message: "",
	}
	runContractNormal := func(*commonPb.Contract, string, []byte, map[string][]byte, protocol.TxSimContext,
		uint64, commonPb.TxType) (*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
		return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
	}
	tests := []struct {
		name              string
		dag               *commonPb.DAG
		applyTxSimContext func(protocol.TxSimContext, protocol.ExecOrderTxType, bool, bool) (bool, int)
		runContract       func(*commonPb.Contract, string, []byte, map[string][]byte, protocol.TxSimContext,
			uint64, commonPb.TxType) (*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode)
		sealTimes int
		wantErr   bool
	}{
		{
			name:              "test0",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         1,
			wantErr:           false,
		},
		{
			name: "testApplyTxSimContextFail",
			dag:  dagNormal,
			applyTxSimContext: func(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
				runVmSuccess bool, applySpecialTx bool) (bool, int) {
				switch txSimContext.GetTx().Payload.TxId {
				case txId0:
					return true, 1
				case txId1:
					return true, 2
				case txId2:
					// simulate that tx2 has conflict with others, return false
					return false, 3
				default:
					panic("Test shouldn't reach here")
				}
			},
			runContract: runContractNormal,
			sealTimes:   1,
			wantErr:     false, // in real case, compareDag will return err since buildDag will build a different dag!
		},
		{
			name:              "testDagHasDuplicates",
			dag:               dagDupVertex,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         0,
			wantErr:           true,
		},
		{
			name:              "testDagHasCycle",
			dag:               dagCycle,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         0,
			wantErr:           true,
		},
		{
			name:              "testDagHasOutOfBoundIndex",
			dag:               dagOutOfBound,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         0,
			wantErr:           true,
		},
		{
			name:              "testDagMissesVertex",
			dag:               dagMissingVertex,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         0,
			wantErr:           true,
		},
		{
			name:              "testIteratorTx",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract: func(contract *commonPb.Contract, method string, byteCode []byte, parameters map[string][]byte,
				txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (
				*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
				txId := txContext.GetTx().GetPayload().GetTxId()
				if txId == txId0 {
					return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
				}
				return contractResult, protocol.ExecOrderTxTypeIterator, commonPb.TxStatusCode_SUCCESS
			},
			sealTimes: 1,
			wantErr:   false,
		},
		{
			name:              "testIteratorTxAtBeginning",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract: func(contract *commonPb.Contract, method string, byteCode []byte, parameters map[string][]byte,
				txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (
				*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
				txId := txContext.GetTx().GetPayload().GetTxId()
				if txId != txId0 {
					return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
				}
				return contractResult, protocol.ExecOrderTxTypeIterator, commonPb.TxStatusCode_SUCCESS
			},
			sealTimes: 1,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmMgr, _, _, snapshot, scheduler, contractId, block := prepare(t, false, false, 3, false)

			parameters := make(map[string]string, 8)
			tx0 := newTx(txId0, contractId, parameters)
			tx1 := newTx(txId1, contractId, parameters)
			tx2 := newTx(txId2, contractId, parameters)

			block.Txs = []*commonPb.Transaction{tx0, tx1, tx2}
			block.Dag = tt.dag

			snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
			snapshot.EXPECT().Seal().Return().Times(tt.sealTimes)
			snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(tt.applyTxSimContext)
			txResults := make(map[string]*commonPb.Result, len(block.Txs))
			snapshot.EXPECT().GetTxResultMap().AnyTimes().Return(txResults)
			dagCopy := &commonPb.DAG{}
			dagBytes, _ := proto.Marshal(tt.dag)
			proto.Unmarshal(dagBytes, dagCopy)
			if tt.name == "testIteratorTx" {
				dagCopy = &commonPb.DAG{
					Vertexes: []*commonPb.DAG_Neighbor{
						{
							Neighbors: []uint32{},
						},
					},
				}
			}
			snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).AnyTimes().Return(dagCopy)

			vmMgr.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(tt.runContract)

			txRwSet, result, err := scheduler.SimulateWithDag(block, snapshot)
			if tt.wantErr {
				require.NotNil(t, err)
				fmt.Println("err: ", err)
			} else {
				require.Nil(t, err)
				require.NotNil(t, txRwSet)
				require.NotNil(t, result)
				fmt.Println("txRWSet: ", txRwSet)
				fmt.Println("result: ", result)
			}
		})
	}
}

func TestSimulateWithDagUnderGasEnabled(t *testing.T) {

	dagNormal := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: nil,
			},
			{
				Neighbors: []uint32{0},
			},
			{
				Neighbors: []uint32{0, 1},
			},
		},
	}
	applyTxSimContextNormal := func(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
		runVmSuccess bool, applySpecialTx bool) (bool, int) {
		switch txSimContext.GetTx().Payload.TxId {
		case txId0:
			return true, 1
		case txId1:
			return true, 2
		case txId2:
			return true, 3
		default:
			panic("Test shouldn't reach here")
		}
	}
	contractResult := &commonPb.ContractResult{
		Code:    0,
		Result:  nil,
		Message: "",
	}
	runContractNormal := func(*commonPb.Contract, string, []byte, map[string][]byte, protocol.TxSimContext,
		uint64, commonPb.TxType) (*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
		return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
	}
	tests := []struct {
		name              string
		dag               *commonPb.DAG
		applyTxSimContext func(protocol.TxSimContext, protocol.ExecOrderTxType, bool, bool) (bool, int)
		runContract       func(*commonPb.Contract, string, []byte, map[string][]byte, protocol.TxSimContext,
			uint64, commonPb.TxType) (*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode)
		sealTimes int
		wantErr   bool
	}{
		{
			name:              "test0",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract:       runContractNormal,
			sealTimes:         1,
			wantErr:           true, // last tx should be gas type
		},
		{
			name:              "test1",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract: func(contract *commonPb.Contract, method string, byteCode []byte, parameters map[string][]byte,
				txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (
				*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
				txId := txContext.GetTx().GetPayload().GetTxId()
				if txId == txId0 {
					return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
				} else if txId == txId1 {
					return contractResult, protocol.ExecOrderTxTypeIterator, commonPb.TxStatusCode_SUCCESS
				} else {
					return contractResult, protocol.ExecOrderTxTypeChargeGas, commonPb.TxStatusCode_SUCCESS
				}
			},
			sealTimes: 1,
			wantErr:   false,
		},
		{
			name:              "test2",
			dag:               dagNormal,
			applyTxSimContext: applyTxSimContextNormal,
			runContract: func(contract *commonPb.Contract, method string, byteCode []byte, parameters map[string][]byte,
				txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (
				*commonPb.ContractResult, protocol.ExecOrderTxType, commonPb.TxStatusCode) {
				txId := txContext.GetTx().GetPayload().GetTxId()
				if txId == txId0 {
					return contractResult, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS
				}
				return contractResult, protocol.ExecOrderTxTypeIterator, commonPb.TxStatusCode_SUCCESS
			},
			sealTimes: 1,
			wantErr:   true, // last tx should be gas type
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vmMgr, _, _, snapshot, scheduler, contractId, block := prepare4(t, true, false, false, 3, false)

			parameters := make(map[string]string, 8)
			tx0 := newTx(txId0, contractId, parameters)
			tx1 := newTx(txId1, contractId, parameters)
			tx2 := newTx(txId2, contractId, parameters)
			tx2.Payload.ContractName = syscontract.SystemContract_ACCOUNT_MANAGER.String()

			block.Txs = []*commonPb.Transaction{tx0, tx1, tx2}
			block.Dag = tt.dag

			snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
			snapshot.EXPECT().Seal().Return().Times(tt.sealTimes)
			snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(tt.applyTxSimContext)
			txResults := make(map[string]*commonPb.Result, len(block.Txs))
			snapshot.EXPECT().GetTxResultMap().AnyTimes().Return(txResults)
			dagCopy := &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
				},
			}
			snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).AnyTimes().Return(dagCopy)

			vmMgr.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(tt.runContract)

			txRwSet, result, err := scheduler.SimulateWithDag(block, snapshot)
			if tt.wantErr {
				require.NotNil(t, err)
				fmt.Println("err: ", err)
			} else {
				require.Nil(t, err)
				require.NotNil(t, txRwSet)
				require.NotNil(t, result)
				fmt.Println("txRWSet: ", txRwSet)
				fmt.Println("result: ", result)
			}
		})
	}
}
func TestMarshalDag(t *testing.T) {
	dag := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{0},
			},
			{
				Neighbors: []uint32{0, 1},
			},
		},
	}

	mar, _ := proto.Marshal(dag)

	dag2 := &commonPb.DAG{}
	proto.Unmarshal(mar, dag2)
	equal, err := utils.IsDagEqual(dag, dag2)
	require.NoError(t, err)
	require.Truef(t, equal, "dag:%+v, dag2:%+v, mar:%s(len:%d)", dag, dag2, mar, len(mar))
	//require.Truef(t, reflect.DeepEqual(dag, dag2), "dag:%+v, dag2:%+v, mar:%s(len:%d)", dag, dag2, mar, len(mar))
	//DeepEqual false due to empty slice/nil issue

	dag = &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{},
	}
	mar, _ = proto.Marshal(dag)
	dag2 = &commonPb.DAG{}
	proto.Unmarshal(mar, dag2)
	equal, err = utils.IsDagEqual(dag, dag2)
	require.NoError(t, err)
	require.Truef(t, equal, "dag:%+v, dag2:%+v, mar:%s(len:%d)", dag, dag2, mar, len(mar))
	//require.Truef(t, reflect.DeepEqual(dag, dag2), "dag:%+v, dag2:%+v, mar:%s(len:%d)", dag, dag2, mar, len(mar))
	//DeepEqual false due to empty slice/nil issue
}

func Test_errResult(t *testing.T) {
	type args struct {
		result *commonPb.Result
		err    error
	}
	tests := []struct {
		name    string
		args    args
		want    *commonPb.Result
		want1   protocol.ExecOrderTxType
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				result: &commonPb.Result{
					ContractResult: &commonPb.ContractResult{
						Message: "test err",
					},
				},
				err: errors.New("test err"),
			},
			want: &commonPb.Result{
				Code: commonPb.TxStatusCode_INVALID_PARAMETER,
				ContractResult: &commonPb.ContractResult{
					Code:    uint32(commonPb.TxStatusCode_TIMEOUT),
					Message: "test err",
				},
			},
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := errResult(tt.args.result, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("errResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("errResult() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("errResult() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTxScheduler_parseParameter(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		parameterPairs []*commonPb.KeyValuePair
	}

	ctrl := gomock.NewController(t)
	vmM := mock.NewMockVmManager(ctrl)
	chainConf := mock.NewMockChainConf(ctrl)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	vmM.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonPb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS).AnyTimes()
	keyReg, _ := regexp.Compile(protocol.DefaultStateRegex)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string][]byte
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          keyReg,
			},
			args: args{
				parameterPairs: []*commonPb.KeyValuePair{
					{
						Key:   "test",
						Value: []byte("get"),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          keyReg,
			},
			args: args{
				parameterPairs: []*commonPb.KeyValuePair{
					{
						Key:   "123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890123456890",
						Value: []byte("get"),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			got, err := ts.parseParameter2220(tt.args.parameterPairs, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseParameter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxScheduler_dumpDAG(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		dag *commonPb.DAG
		txs []*commonPb.Transaction
	}

	ctrl := gomock.NewController(t)
	vmM := mock.NewMockVmManager(ctrl)
	chainConf := mock.NewMockChainConf(ctrl)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	vmM.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonPb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS).AnyTimes()

	//_, _, _, _, _, contractId, _ := prepare(t, false, false, 2)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          nil,
			},
			args: args{
				dag: &commonPb.DAG{},
				txs: []*commonPb.Transaction{},
			},
		},
		//{
		//	name: "test1",
		//	fields: fields{
		//		lock:            sync.Mutex{},
		//		VmManager:       vmM,
		//		scheduleFinishC: make(chan bool),
		//		log:             logger,
		//		chainConf:       chainConf,
		//		metricVMRunTime: nil,
		//		StoreHelper:     storeHelper,
		//		keyReg:          nil,
		//	},
		//	args: args{
		//		dag: &commonPb.DAG{
		//			Vertexes: []*commonPb.DAG_Neighbor{
		//				{
		//					Neighbors: []uint32{0, 1, 2, 3},
		//				},
		//				{
		//					Neighbors: []uint32{4, 5, 6, 7},
		//				},
		//				{
		//					Neighbors: []uint32{8, 9},
		//				},
		//			},
		//		},
		//		txs: []*commonPb.Transaction{
		//			newTx("a0000000000000000000000000000001", contractId, make(map[string]string, 8)),
		//		},
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			ts.dumpDAG(tt.args.dag, tt.args.txs)
		})
	}
}

func TestTxScheduler_chargeGasLimit(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		accountMangerContract *commonPb.Contract
		tx                    *commonPb.Transaction
		txSimContext          protocol.TxSimContext
		contractName          string
		method                string
		pk                    []byte
		result                *commonPb.Result
	}

	ctrl := gomock.NewController(t)
	vmM := mock.NewMockVmManager(ctrl)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	vmM.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonPb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS).AnyTimes()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRe  *commonPb.Result
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          nil,
			},
			args: args{
				accountMangerContract: &commonPb.Contract{
					Name: syscontract.InitContract_CONTRACT_NAME.String(),
				},
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
						Limit: &commonPb.Limit{
							GasLimit: 100,
						},
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code: 0,
				},
			},
			wantRe:  &commonPb.Result{},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          nil,
			},
			args: args{
				accountMangerContract: &commonPb.Contract{
					Name: syscontract.InitContract_CONTRACT_NAME.String(),
				},
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code:    0,
					Message: "tx payload limit is nil",
				},
			},
			wantRe: &commonPb.Result{
				Message: "tx payload limit is nil",
			},
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				VmManager: func() protocol.VmManager {

					vmM := mock.NewMockVmManager(ctrl)
					vmM.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonPb.ContractResult{
						Code:    uint32(commonPb.TxStatusCode_CONTRACT_FAIL),
						Message: "invoke contract fail",
					}, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_CONTRACT_FAIL).AnyTimes()

					return vmM
				}(),
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				metricVMRunTime: nil,
				StoreHelper:     storeHelper,
				keyReg:          nil,
			},
			args: args{
				accountMangerContract: &commonPb.Contract{
					Name: syscontract.InitContract_CONTRACT_NAME.String(),
				},
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
						Limit: &commonPb.Limit{
							GasLimit: 100,
						},
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code:    commonPb.TxStatusCode_CONTRACT_FAIL,
					Message: "invoke contract fail",
				},
			},
			wantRe: &commonPb.Result{
				Code:    commonPb.TxStatusCode_CONTRACT_FAIL,
				Message: "invoke contract fail",
				ContractResult: &commonPb.ContractResult{
					Code:    uint32(commonPb.TxStatusCode_CONTRACT_FAIL),
					Message: "invoke contract fail",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			gotRe, err := ts.chargeGasLimit(tt.args.accountMangerContract, tt.args.tx, tt.args.txSimContext, tt.args.contractName, tt.args.method, tt.args.pk, tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("chargeGasLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRe, tt.wantRe) {
				t.Errorf("chargeGasLimit() gotRe = %v, want %v", gotRe, tt.wantRe)
			}
		})
	}
}

func TestTxScheduler_refundGas(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		accountMangerContract *commonPb.Contract
		tx                    *commonPb.Transaction
		txSimContext          protocol.TxSimContext
		contractName          string
		method                string
		pk                    []byte
		result                *commonPb.Result
		contractResultPayload *commonPb.ContractResult
	}

	ctrl := gomock.NewController(t)
	vmM := mock.NewMockVmManager(ctrl)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	vmM.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonPb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonPb.TxStatusCode_SUCCESS).AnyTimes()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRe  *commonPb.Result
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				VmManager:   vmM,
				StoreHelper: storeHelper,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code: 0,
				},
				contractResultPayload: &commonPb.ContractResult{
					Code:   0,
					Result: []byte("success"),
				},
			},
			wantRe: &commonPb.Result{
				Code:    0,
				Message: "tx payload limit is nil",
			},
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				VmManager:   vmM,
				StoreHelper: storeHelper,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
						Limit: &commonPb.Limit{
							GasLimit: 10,
						},
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code: 0,
				},
				contractResultPayload: &commonPb.ContractResult{
					Code:    0,
					Result:  []byte("success"),
					GasUsed: 100,
				},
			},
			wantRe: &commonPb.Result{
				Code:    0,
				Message: "gas limit is not enough, [limit:10]/[gasUsed:100]",
			},
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				VmManager:   vmM,
				StoreHelper: storeHelper,
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
						Limit: &commonPb.Limit{
							GasLimit: 200,
						},
					},
				},
				pk: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					return txSimContext
				}(),
				result: &commonPb.Result{
					Code: 0,
				},
				contractResultPayload: &commonPb.ContractResult{
					Code:    0,
					Result:  []byte("success"),
					GasUsed: 100,
				},
			},
			wantRe: &commonPb.Result{
				Code: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			gotRe, err := ts.refundGas(tt.args.accountMangerContract, tt.args.tx, tt.args.txSimContext, tt.args.contractName, tt.args.method, tt.args.pk, tt.args.result, tt.args.contractResultPayload)
			if (err != nil) != tt.wantErr {
				t.Errorf("refundGas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRe, tt.wantRe) {
				t.Errorf("refundGas() gotRe = %v, want %v", gotRe, tt.wantRe)
			}
		})
	}
}

const cert = "-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"

func TestTxScheduler_getAccountMgrContractAndPk(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		txSimContext protocol.TxSimContext
		tx           *commonPb.Transaction
		contractName string
		method       string
	}

	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		wantAccountMangerContract *commonPb.Contract
		wantPk                    []byte
		wantErr                   bool
	}{
		{
			name: "test0",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
					},
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_CERT,
							MemberInfo: []byte(cert),
						},
					},
				},
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_CERT,
					//	MemberInfo: []byte(cert),
					//})
					txSimContext.EXPECT().GetContractByName(syscontract.SystemContract_ACCOUNT_MANAGER.String()).Return(&commonPb.Contract{
						Name: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
					}, nil)

					return txSimContext
				}(),
			},
			wantAccountMangerContract: func() *commonPb.Contract {

				return &commonPb.Contract{
					Name: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
				}
			}(),
			wantPk: func() []byte {
				member := cert
				pubKeyBytes, err := publicKeyFromCert([]byte(member))
				if err != nil {
					t.Log(err)
					return nil
				}

				return pubKeyBytes
			}(),
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
					},
				},
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetBlockVersion().Return(uint32(2030102)).AnyTimes()
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetContractByName(syscontract.SystemContract_ACCOUNT_MANAGER.String()).Return(nil, errors.New("txSimContext GetContractByName data is nil"))
					return txSimContext
				}(),
			},
			wantAccountMangerContract: func() *commonPb.Contract {
				return nil
			}(),
			wantPk: func() []byte {
				return nil
			}(),
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: false,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			args: args{
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
				tx: &commonPb.Transaction{
					Payload: &commonPb.Payload{
						TxType: commonPb.TxType_INVOKE_CONTRACT,
					},
				},
				txSimContext: func() protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().GetSnapshot().Return(mock.NewMockSnapshot(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetBlockchainStore().Return(mock.NewMockBlockchainStore(ctrl)).AnyTimes()
					txSimContext.EXPECT().GetContractByName(gomock.Any()).AnyTimes()
					//txSimContext.EXPECT().GetSender().AnyTimes()
					return txSimContext
				}(),
			},
			wantAccountMangerContract: func() *commonPb.Contract {
				return nil
			}(),
			wantPk: func() []byte {
				return nil
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			gotAccountMangerContract, gotPk, err := ts.getAccountMgrContractAndPk(tt.args.txSimContext, tt.args.tx, tt.args.contractName, tt.args.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAccountMgrContractAndPk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAccountMangerContract, tt.wantAccountMangerContract) {
				t.Errorf("getAccountMgrContractAndPk() gotAccountMangerContract = %v, want %v", gotAccountMangerContract, tt.wantAccountMangerContract)
			}
			if !reflect.DeepEqual(gotPk, tt.wantPk) {
				t.Errorf("getAccountMgrContractAndPk() gotPk = %v, want %v", gotPk, tt.wantPk)
			}
		})
	}
}

func TestTxScheduler_checkGasEnable(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: true,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()

					return chainConf
				}(),
				log: logger,
			},
			want: true,
		},
		{
			name: "test1",
			fields: fields{
				chainConf: func() protocol.ChainConf {
					chainConf := mock.NewMockChainConf(ctrl)
					chainConfig := &configpb.ChainConfig{
						AccountConfig: &configpb.GasAccountConfig{
							EnableGas: false,
						},
					}
					chainConf.EXPECT().ChainConfig().Return(chainConfig).AnyTimes()
					return chainConf
				}(),
				log: logger,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			if got := ts.checkGasEnable(); got != tt.want {
				t.Errorf("checkGasEnable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTxScheduler_checkNativeFilter(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		blockVersion uint32
		contractName string
		method       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test0",
			fields: fields{
				log: logger.GetLogger("unit-test"),
			},
			args: args{
				blockVersion: uint32(blockVersion2312),
				contractName: syscontract.InitContract_CONTRACT_NAME.String(),
				method:       syscontract.InitContract_CONTRACT_VERSION.String(),
			},
			want: true,
		},
		{
			name: "test1",
			fields: fields{
				log: logger.GetLogger("unit-test"),
			},
			args: args{
				blockVersion: uint32(2030102),
				contractName: syscontract.SystemContract_CHAIN_QUERY.String(),
				method:       syscontract.ChainQueryFunction_GET_BLOCK_BY_HASH.String(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}
			if got := ts.checkNativeFilter(tt.args.blockVersion, tt.args.contractName, tt.args.method, nil, nil); got != tt.want {
				t.Errorf("checkNativeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_publicKeyFromCert(t *testing.T) {
	type args struct {
		member []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				member: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
			},
			want: func() []byte {

				member := "-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"

				certificate, err := utils.ParseCert([]byte(member))
				if err != nil {
					t.Log(err)
					return nil
				}
				pubKeyStr, err := certificate.PublicKey.String()
				if err != nil {
					t.Log(err)
					return nil
				}

				return []byte(pubKeyStr)
			}(),
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				member: []byte("-----BEGIN CERTIFICATE-----\\nMIICiTCCAi+gAwIBAgIDA+zYMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\\nb3JnMi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\\nExljYS53eC1vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcyLmNoYWlubWFrZXIub3Jn\\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcy\\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZd92CJez\\nCiOMzLSTrJfX5vIUArCycg05uKru2qFaX0uvZUCwNxbfSuNvkHRXE8qIBUhTbg1Q\\nR9rOlfDY1WfgMaN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\\nKQYDVR0OBCIEICfLatSyyebzRsLbnkNKZJULB2bZOtG+88NqvAHCsXa3MCsGA1Ud\\nIwQkMCKAIPGP1bPT4/Lns2PnYudZ9/qHscm0pGL6Kfy+1CAFWG0hMAoGCCqGSM49\\nBAMCA0gAMEUCIQDzHrEHrGNtoNfB8jSJrGJU1qcxhse74wmDgIdoGjvfTwIgabRJ\\nJNvZKRpa/VyfYi3TXa5nhHRIn91ioF1dQroHQFc=\\n-----END CERTIFICATE-----"),
			},
			want: func() []byte {

				member := "-----BEGIN CERTIFICATE-----\\nMIICiTCCAi+gAwIBAgIDA+zYMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\\nb3JnMi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\\nExljYS53eC1vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcyLmNoYWlubWFrZXIub3Jn\\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcy\\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZd92CJez\\nCiOMzLSTrJfX5vIUArCycg05uKru2qFaX0uvZUCwNxbfSuNvkHRXE8qIBUhTbg1Q\\nR9rOlfDY1WfgMaN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\\nKQYDVR0OBCIEICfLatSyyebzRsLbnkNKZJULB2bZOtG+88NqvAHCsXa3MCsGA1Ud\\nIwQkMCKAIPGP1bPT4/Lns2PnYudZ9/qHscm0pGL6Kfy+1CAFWG0hMAoGCCqGSM49\\nBAMCA0gAMEUCIQDzHrEHrGNtoNfB8jSJrGJU1qcxhse74wmDgIdoGjvfTwIgabRJ\\nJNvZKRpa/VyfYi3TXa5nhHRIn91ioF1dQroHQFc=\\n-----END CERTIFICATE-----"
				certificate, err := utils.ParseCert([]byte(member))
				if err != nil {
					t.Log(err)
					return nil
				}
				pubKeyStr, err := certificate.PublicKey.String()
				if err != nil {
					t.Log(err)
					return nil
				}

				return []byte(pubKeyStr)
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := publicKeyFromCert(tt.args.member)
			if (err != nil) != tt.wantErr {
				t.Errorf("publicKeyFromCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("publicKeyFromCert() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wholeCertInfo(t *testing.T) {
	type args struct {
		txSimContext protocol.TxSimContext
		certHash     string
	}

	tests := []struct {
		name    string
		args    args
		want    *commonPb.CertInfo
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				txSimContext: func() protocol.TxSimContext {
					ctrl := gomock.NewController(t)
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().Get(syscontract.SystemContract_CERT_MANAGE.String(), []byte("123456")).Return([]byte("123456"), nil)
					return txSimContext
				}(),
				certHash: "123456",
			},
			want: &commonPb.CertInfo{
				Hash: "123456",
				Cert: []byte("123456"),
			},
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				txSimContext: func() protocol.TxSimContext {
					ctrl := gomock.NewController(t)
					txSimContext := mock.NewMockTxSimContext(ctrl)
					txSimContext.EXPECT().Get(syscontract.SystemContract_CERT_MANAGE.String(), []byte("123456")).Return(nil, errors.New("txSimContext get is nil"))
					return txSimContext
				}(),
				certHash: "123456",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := wholeCertInfo(tt.args.txSimContext, tt.args.certHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("wholeCertInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wholeCertInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSenderGroup(t *testing.T) {
	type args struct {
		txBatch []*commonPb.Transaction
	}

	_, _, _, _, _, contractId, _ := prepare(t, false, false, 2, true)

	parameters := make(map[string]string, 8)

	txBatch := make([]*commonPb.Transaction, 0)

	txBatch = append(txBatch, newTx("a0000000000000000000000000000001", contractId, parameters))
	tests := []struct {
		name string
		args args
		want *SenderGroup
	}{
		{
			name: "test0",
			args: args{
				txBatch: txBatch,
			},
			want: &SenderGroup{
				txsMap:     getSenderTxsMap(txBatch),
				doneTxKeyC: make(chan [32]byte, len(txBatch)),
			},
		},
		{
			name: "test1",
			args: args{
				txBatch: []*commonPb.Transaction{
					newTx("a0000000000000000000000000000001", contractId, parameters),
					newTx("a0000000000000000000000000000002", contractId, parameters),
				},
			},
			want: &SenderGroup{
				txsMap: getSenderTxsMap([]*commonPb.Transaction{
					newTx("a0000000000000000000000000000001", contractId, parameters),
					newTx("a0000000000000000000000000000002", contractId, parameters),
				}),
				doneTxKeyC: make(chan [32]byte, 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSenderGroup(tt.args.txBatch); !reflect.DeepEqual(got.txsMap, tt.want.txsMap) {
				t.Errorf("NewSenderGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSenderTxsMap(t *testing.T) {
	type args struct {
		txBatch []*commonPb.Transaction
	}

	_, _, _, _, _, contractId, _ := prepare(t, false, false, 2, true)

	parameters := make(map[string]string, 8)
	tests := []struct {
		name string
		args args
		want map[[32]byte][]*commonPb.Transaction
	}{
		{
			name: "test0",
			args: args{
				txBatch: []*commonPb.Transaction{
					newTx("a0000000000000000000000000000001", contractId, parameters),
				},
			},
			want: func() map[[32]byte][]*commonPb.Transaction {
				senderTxsMap := make(map[[32]byte][]*commonPb.Transaction)
				tx := newTx("a0000000000000000000000000000001", contractId, parameters)
				hashKey, _ := getSenderHashKey(tx)
				senderTxsMap[hashKey] = append(senderTxsMap[hashKey], tx)
				return senderTxsMap
			}(),
		},
		{
			name: "test1",
			args: args{
				txBatch: []*commonPb.Transaction{
					newTx("a0000000000000000000000000000001", contractId, parameters),
					newTx("a0000000000000000000000000000002", contractId, parameters),
				},
			},
			want: func() map[[32]byte][]*commonPb.Transaction {
				senderTxsMap := make(map[[32]byte][]*commonPb.Transaction)
				tx1 := newTx("a0000000000000000000000000000001", contractId, parameters)
				hashKey1, _ := getSenderHashKey(tx1)
				senderTxsMap[hashKey1] = append(senderTxsMap[hashKey1], tx1)
				tx2 := newTx("a0000000000000000000000000000002", contractId, parameters)
				hashKey2, _ := getSenderHashKey(tx2)
				senderTxsMap[hashKey2] = append(senderTxsMap[hashKey2], tx2)
				return senderTxsMap
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSenderTxsMap(tt.args.txBatch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSenderTxsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSenderHashKey(t *testing.T) {
	type args struct {
		tx *commonPb.Transaction
	}

	_, _, _, _, _, contractId, _ := prepare(t, false, false, 2, true)

	parameters := make(map[string]string, 8)
	tests := []struct {
		name    string
		args    args
		want    [32]byte
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				tx: newTx("a0000000000000000000000000000001", contractId, parameters),
			},
			want: func() [32]byte {
				keyBytes, _ := newTx("a0000000000000000000000000000001", contractId, parameters).Sender.GetSigner().Marshal()
				return sha256.Sum256(keyBytes)
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSenderHashKey(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSenderHashKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSenderHashKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCycleExists(t *testing.T) {
	applyTxSimContext := func(txSimContext protocol.TxSimContext, specialTxType protocol.ExecOrderTxType,
		runVmSuccess bool, applySpecialTx bool) (bool, int) {
		i, _ := strconv.Atoi(txSimContext.GetTx().Payload.TxId)
		return true, i
	}
	dag0 := &commonPb.DAG{
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
			{
				Neighbors: []uint32{2, 6},
			},
			{
				Neighbors: []uint32{3},
			},
			{
				Neighbors: []uint32{4},
			},
			{
				Neighbors: []uint32{5},
			},
		},
	}
	dag1 := &commonPb.DAG{
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
			{
				Neighbors: []uint32{2, 6},
			},
			{
				Neighbors: []uint32{3},
			},
			{
				Neighbors: []uint32{4},
			},
			{
				Neighbors: []uint32{5, 7},
			},
			{
				Neighbors: []uint32{8},
			},
			{
				Neighbors: []uint32{3},
			},
		},
	}
	dag2 := &commonPb.DAG{
		Vertexes: []*commonPb.DAG_Neighbor{
			{
				Neighbors: []uint32{2, 5, 28},
			},
			{
				Neighbors: []uint32{4, 0},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{7},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{10},
			},
			{
				Neighbors: []uint32{14},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{18, 4},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{19, 2},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{26},
			},
			{
				Neighbors: []uint32{2},
			},
			{
				Neighbors: []uint32{26, 1, 3, 5, 14},
			},
			{
				Neighbors: []uint32{16},
			},
			{
				Neighbors: []uint32{23},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{21},
			},
			{
				Neighbors: []uint32{},
			},
			{
				Neighbors: []uint32{12},
			},
			{
				Neighbors: []uint32{5},
			},
			{
				Neighbors: []uint32{29, 2, 3},
			},
			{
				Neighbors: []uint32{20, 23, 26},
			},
			{
				Neighbors: []uint32{12},
			},
			{
				Neighbors: []uint32{15},
			},
		},
	}
	tests := []struct {
		name    string
		dag     *commonPb.DAG
		wantErr bool
	}{
		{
			name:    "test0",
			dag:     dag0,
			wantErr: true,
		},
		{
			name:    "test1",
			dag:     dag1,
			wantErr: true,
		},
		{
			name:    "test2",
			dag:     dag2,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, snapshot, scheduler, contractId, block := prepare(t, false, false, 2, true)

			parameters := make(map[string]string, 8)
			txs := make([]*commonPb.Transaction, len(tt.dag.Vertexes))
			for i := 1; i <= len(tt.dag.Vertexes); i++ {
				tx := newTx(fmt.Sprintf("%016d", i), contractId, parameters)
				txs[i-1] = tx
			}

			block.Txs = txs
			block.Dag = tt.dag

			snapshot.EXPECT().IsSealed().AnyTimes().Return(false)
			snapshot.EXPECT().Seal().Return().AnyTimes()
			snapshot.EXPECT().ApplyTxSimContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(applyTxSimContext)
			txRWSets := make(map[string]*commonPb.Result, len(block.Txs))
			snapshot.EXPECT().GetTxResultMap().AnyTimes().Return(txRWSets)

			txRwSet, result, err := scheduler.SimulateWithDag(block, snapshot)
			if tt.wantErr {
				require.NotNil(t, err)
				fmt.Println("err: ", err)
			} else {
				require.Nil(t, err)
				require.NotNil(t, txRwSet)
				require.NotNil(t, result)
				fmt.Println("txRWSet: ", txRwSet)
				fmt.Println("result: ", result)
			}
		})
	}
}

func TestTxScheduler_verifyExecOrderTxType(t *testing.T) {
	type fields struct {
		EnableOptimizeChargeGas  bool
		EnableConflictsBitWindow bool
		EnableGas                bool
	}
	type args struct {
		txExecOrderTypeMap map[string]protocol.ExecOrderTxType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint32
		want1   uint32
		want2   uint32
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				EnableOptimizeChargeGas:  false,
				EnableConflictsBitWindow: true,
				EnableGas:                false,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeIterator,
					txId2: protocol.ExecOrderTxTypeIterator,
				},
			},
			want:    1,
			want1:   2,
			want2:   0,
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				EnableOptimizeChargeGas:  false,
				EnableConflictsBitWindow: true,
				EnableGas:                false,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeIterator,
					txId2: protocol.ExecOrderTxTypeChargeGas,
				},
			},
			want:    1,
			want1:   1,
			want2:   1,
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				EnableOptimizeChargeGas:  true,
				EnableConflictsBitWindow: true,
				EnableGas:                true,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeIterator,
					txId2: protocol.ExecOrderTxTypeChargeGas,
				},
			},
			want:    1,
			want1:   1,
			want2:   1,
			wantErr: false,
		},
		{
			name: "test3",
			fields: fields{
				EnableOptimizeChargeGas:  true,
				EnableConflictsBitWindow: true,
				EnableGas:                true,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeChargeGas,
					txId2: protocol.ExecOrderTxTypeIterator,
				},
			},
			want:    1,
			want1:   1,
			want2:   1,
			wantErr: true,
		},
		{
			name: "test4",
			fields: fields{
				EnableOptimizeChargeGas:  true,
				EnableConflictsBitWindow: true,
				EnableGas:                true,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeIterator,
					txId1: protocol.ExecOrderTxTypeNormal,
					txId2: protocol.ExecOrderTxTypeChargeGas,
				},
			},
			want:    1,
			want1:   1,
			want2:   1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			chainConf := mock.NewMockChainConf(ctl)
			chainConfig := &configpb.ChainConfig{
				Core: &configpb.CoreConfig{
					EnableOptimizeChargeGas:  tt.fields.EnableOptimizeChargeGas,
					EnableConflictsBitWindow: tt.fields.EnableConflictsBitWindow,
				},
				AccountConfig: &configpb.GasAccountConfig{
					EnableGas: tt.fields.EnableGas,
				},
			}
			chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
			ts := &TxScheduler{
				chainConf: chainConf,
			}
			contractId := &commonPb.Contract{
				Name:        "",
				Version:     "",
				RuntimeType: 0,
				Status:      0,
				Creator:     nil,
				Address:     "",
			}
			parameters := make(map[string]string, 8)
			tx0 := newTx(txId0, contractId, parameters)
			tx1 := newTx(txId1, contractId, parameters)
			tx2 := newTx(txId2, contractId, parameters)

			block := &commonPb.Block{}
			block.Txs = []*commonPb.Transaction{tx0, tx1, tx2}
			got, got1, got2, err := ts.verifyExecOrderTxType(block, tt.args.txExecOrderTypeMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyExecOrderTxType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("verifyExecOrderTxType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("verifyExecOrderTxType() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("verifyExecOrderTxType() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestTxScheduler_compareDag(t *testing.T) {
	type fields struct {
		EnableOptimizeChargeGas  bool
		EnableConflictsBitWindow bool
		EnableGas                bool
	}
	type args struct {
		txExecOrderTypeMap map[string]protocol.ExecOrderTxType
		dag                *commonPb.DAG
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
				EnableOptimizeChargeGas:  false,
				EnableConflictsBitWindow: true,
				EnableGas:                false,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeNormal,
					txId2: protocol.ExecOrderTxTypeIterator,
					txId3: protocol.ExecOrderTxTypeIterator,
					txId4: protocol.ExecOrderTxTypeIterator,
				},
				dag: &commonPb.DAG{
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
						{
							Neighbors: []uint32{2},
						},
						{
							Neighbors: []uint32{3},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				EnableOptimizeChargeGas:  true,
				EnableConflictsBitWindow: true,
				EnableGas:                true,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeNormal,
					txId2: protocol.ExecOrderTxTypeIterator,
					txId3: protocol.ExecOrderTxTypeIterator,
					txId4: protocol.ExecOrderTxTypeIterator,
				},
				dag: &commonPb.DAG{
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
						{
							Neighbors: []uint32{2},
						},
						{
							Neighbors: []uint32{3},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				EnableOptimizeChargeGas:  true,
				EnableConflictsBitWindow: true,
				EnableGas:                true,
			},
			args: args{
				txExecOrderTypeMap: map[string]protocol.ExecOrderTxType{
					txId0: protocol.ExecOrderTxTypeNormal,
					txId1: protocol.ExecOrderTxTypeNormal,
					txId2: protocol.ExecOrderTxTypeIterator,
					txId3: protocol.ExecOrderTxTypeIterator,
					txId4: protocol.ExecOrderTxTypeChargeGas,
				},
				dag: &commonPb.DAG{
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
						{
							Neighbors: []uint32{2},
						},
						{
							Neighbors: []uint32{0, 1, 2, 3},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			snapshot := mock.NewMockSnapshot(ctl)
			chainConf := mock.NewMockChainConf(ctl)
			chainConfig := &configpb.ChainConfig{
				Core: &configpb.CoreConfig{
					EnableOptimizeChargeGas:  tt.fields.EnableOptimizeChargeGas,
					EnableConflictsBitWindow: tt.fields.EnableConflictsBitWindow,
				},
				AccountConfig: &configpb.GasAccountConfig{
					EnableGas: tt.fields.EnableGas,
				},
				Contract: &configpb.ContractConfig{
					EnableSqlSupport: false,
				},
			}
			dagCopy := &commonPb.DAG{
				Vertexes: []*commonPb.DAG_Neighbor{
					{
						Neighbors: []uint32{},
					},
					{
						Neighbors: []uint32{},
					},
				},
			}
			snapshot.EXPECT().BuildDAG(gomock.Any(), gomock.Any()).AnyTimes().Return(dagCopy)
			snapshot.EXPECT().GetSnapshotSize().AnyTimes().Return(5) // only the 5th tx will call it
			chainConf.EXPECT().ChainConfig().AnyTimes().Return(chainConfig)
			logger := mock.NewMockLogger(ctl)
			logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
			ts := &TxScheduler{
				chainConf: chainConf,
				log:       logger,
			}
			contractId := &commonPb.Contract{
				Name:        "",
				Version:     "",
				RuntimeType: 0,
				Status:      0,
				Creator:     nil,
				Address:     "",
			}
			parameters := make(map[string]string, 8)
			tx0 := newTx(txId0, contractId, parameters)
			tx1 := newTx(txId1, contractId, parameters)
			tx2 := newTx(txId2, contractId, parameters)
			tx3 := newTx(txId3, contractId, parameters)
			tx4 := newTx(txId4, contractId, parameters)

			block := &commonPb.Block{Header: &commonPb.BlockHeader{BlockVersion: blockVersion2300}}
			block.Txs = []*commonPb.Transaction{tx0, tx1, tx2, tx3, tx4}
			block.Dag = tt.args.dag
			txRWSetMap := make(map[string]*commonPb.TxRWSet)
			for i := 0; i < 5; i++ {
				txRWSetMap[fmt.Sprintf("a000000000000000000000000000%04d", i)] = &commonPb.TxRWSet{
					TxId:     fmt.Sprintf("a000000000000000000000000000%04d", i),
					TxReads:  nil,
					TxWrites: nil,
				}
			}
			err := ts.compareDag(block, snapshot, txRWSetMap, tt.args.txExecOrderTypeMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareDag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestTxScheduler_getPayerPk(t *testing.T) {
	type fields struct {
		VmManager       protocol.VmManager
		scheduleFinishC chan bool
		log             protocol.Logger
		chainConf       protocol.ChainConf
		metricVMRunTime *prometheus.HistogramVec
		StoreHelper     conf.StoreHelper
		keyReg          *regexp.Regexp
	}
	type args struct {
		txSimContext protocol.TxSimContext
		tx           *commonPb.Transaction
	}

	ctrl := gomock.NewController(t)
	vmM := mock.NewMockVmManager(ctrl)
	chainConf := mock.NewMockChainConf(ctrl)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: &prometheus.HistogramVec{},
				StoreHelper:     storeHelper,
				keyReg:          &regexp.Regexp{},
			},
			args: args{
				txSimContext: func(ctrl *gomock.Controller) protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_CERT,
					//	MemberInfo: []byte(cert),
					//})
					return txSimContext
				}(ctrl),
				tx: &commonPb.Transaction{
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_CERT,
							MemberInfo: []byte(cert),
						},
					},
				},
			},
			want: func() []byte {

				member := cert
				pubKeyBytes, err := publicKeyFromCert([]byte(member))
				if err != nil {
					t.Log(err)
					return nil
				}

				return pubKeyBytes
			}(),
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: &prometheus.HistogramVec{},
				StoreHelper:     storeHelper,
				keyReg:          &regexp.Regexp{},
			},
			args: args{
				txSimContext: func(ctrl *gomock.Controller) protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)

					txSimContext.EXPECT().Get(syscontract.SystemContract_CERT_MANAGE.String(), []byte("2d2d2d2d2d424547494e2043455254494649434154452d2d2d2d2d0a4d4949436e544343416b53674177494241674944424d58784d416f4743437147534d343942414d434d49474b4d517377435159445651514745774a44546a45510a4d4134474131554543424d48516d5670616d6c755a7a45514d4134474131554542784d48516d5670616d6c755a7a45664d4230474131554543684d57643367740a62334a6e4d53356a61474670626d3168613256794c6d39795a7a45534d4241474131554543784d4a636d39766443316a5a584a304d53497749415944565151440a45786c6a5953353365433176636d63784c6d4e6f59576c75625746725a58497562334a6e4d423458445449794d444d774d5445794d4449794e6c6f5844544d790a4d4449794e7a45794d4449794e6c6f7767596f78437a414a42674e5642415954416b4e4f4d5241774467594456515149457764435a576c716157356e4d5241770a4467594456515148457764435a576c716157356e4d523877485159445651514b45785a3365433176636d63784c6d4e6f59576c75625746725a58497562334a6e0a4d524977454159445651514c45776c79623239304c574e6c636e5178496a416742674e5642414d5447574e684c6e64344c5739795a7a457559326868615735740a5957746c63693576636d63775754415442676371686b6a4f5051494242676771686b6a4f50514d4242774e434141526347456e544441635666316475495477490a53493253355a43306a64514f7968554435694132567631586e47304749455a4e744a4d7a4c4a59756e5a4348673071774646394856445474675557777a6458380a633856426f3447574d4947544d41344741315564447745422f77514541774942426a415042674e5648524d4241663845425441444151482f4d436b47413155640a446751694243427a7958766f326f50683168304b49426570666f7071322f526864396238663545684b654a6255556e734c7a424642674e5648524545506a41380a6767356a61474670626d3168613256794c6d39795a34494a6247396a5957786f62334e3067686c6a5953353365433176636d63784c6d4e6f59576c75625746720a5a58497562334a6e6877522f414141424d416f4743437147534d343942414d43413063414d455143494346764749767868647a6b754d736a6b6756524e504d350a6679344b484c473870444c7a6a38626e32644771416942305a424131642f754242504e4a4166337331667942345233502f67644b4269754441765a39347a6e330a5a673d3d0a2d2d2d2d2d454e442043455254494649434154452d2d2d2d2d0a")).Return([]byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"), nil)
					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_CERT_HASH,
					//	MemberInfo: []byte(cert),
					//})
					return txSimContext
				}(ctrl),
				tx: &commonPb.Transaction{
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_CERT_HASH,
							MemberInfo: []byte(cert),
						},
					},
				},
			},
			want: func() []byte {

				certificate, err := utils.ParseCert([]byte(cert))
				if err != nil {
					t.Log(err)
					return nil
				}
				pubKeyStr, err := certificate.PublicKey.String()
				if err != nil {
					t.Log(err)
					return nil
				}

				return []byte(pubKeyStr)
			}(),
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: &prometheus.HistogramVec{},
				StoreHelper:     storeHelper,
				keyReg:          &regexp.Regexp{},
			},
			args: args{
				txSimContext: func(ctrl *gomock.Controller) protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)

					txSimContext.EXPECT().Get(
						syscontract.SystemContract_CERT_MANAGE.String(),
						[]byte("2d2d2d2d2d424547494e2043455254494649434154452d2d2d2d2d0a4d4949436e544343416b53674177494241674944424d58784d416f4743437147534d343942414d434d49474b4d517377435159445651514745774a44546a45510a4d4134474131554543424d48516d5670616d6c755a7a45514d4134474131554542784d48516d5670616d6c755a7a45664d4230474131554543684d57643367740a62334a6e4d53356a61474670626d3168613256794c6d39795a7a45534d4241474131554543784d4a636d39766443316a5a584a304d53497749415944565151440a45786c6a5953353365433176636d63784c6d4e6f59576c75625746725a58497562334a6e4d423458445449794d444d774d5445794d4449794e6c6f5844544d790a4d4449794e7a45794d4449794e6c6f7767596f78437a414a42674e5642415954416b4e4f4d5241774467594456515149457764435a576c716157356e4d5241770a4467594456515148457764435a576c716157356e4d523877485159445651514b45785a3365433176636d63784c6d4e6f59576c75625746725a58497562334a6e0a4d524977454159445651514c45776c79623239304c574e6c636e5178496a416742674e5642414d5447574e684c6e64344c5739795a7a457559326868615735740a5957746c63693576636d63775754415442676371686b6a4f5051494242676771686b6a4f50514d4242774e434141526347456e544441635666316475495477490a53493253355a43306a64514f7968554435694132567631586e47304749455a4e744a4d7a4c4a59756e5a4348673071774646394856445474675557777a6458380a633856426f3447574d4947544d41344741315564447745422f77514541774942426a415042674e5648524d4241663845425441444151482f4d436b47413155640a446751694243427a7958766f326f50683168304b49426570666f7071322f526864396238663545684b654a6255556e734c7a424642674e5648524545506a41380a6767356a61474670626d3168613256794c6d39795a34494a6247396a5957786f62334e3067686c6a5953353365433176636d63784c6d4e6f59576c75625746720a5a58497562334a6e6877522f414141424d416f4743437147534d343942414d43413063414d455143494346764749767868647a6b754d736a6b6756524e504d350a6679344b484c473870444c7a6a38626e32644771416942305a424131642f754242504e4a4166337331667942345233502f67644b4269754441765a39347a6e330a5a673d3d0a2d2d2d2d2d454e442043455254494649434154452d2d2d2d2d0a")).
						Return([]byte("123456"), nil)

					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_CERT_HASH,
					//	MemberInfo: []byte(cert),
					//})
					return txSimContext
				}(ctrl),
				tx: &commonPb.Transaction{
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_CERT_HASH,
							MemberInfo: []byte(cert),
						},
					},
				},
			},
			want: func() []byte {
				return nil
			}(),
			wantErr: true,
		},
		{
			name: "test3",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: &prometheus.HistogramVec{},
				StoreHelper:     storeHelper,
				keyReg:          &regexp.Regexp{},
			},
			args: args{
				txSimContext: func(ctrl *gomock.Controller) protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_PUBLIC_KEY,
					//	MemberInfo: []byte(cert),
					//})
					return txSimContext
				}(ctrl),
				tx: &commonPb.Transaction{
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_PUBLIC_KEY,
							MemberInfo: []byte(cert),
						},
					},
				},
			},
			want:    []byte(cert),
			wantErr: false,
		},
		{
			name: "test4",
			fields: fields{
				VmManager:       vmM,
				scheduleFinishC: make(chan bool),
				log:             logger,
				chainConf:       chainConf,
				metricVMRunTime: &prometheus.HistogramVec{},
				StoreHelper:     storeHelper,
				keyReg:          &regexp.Regexp{},
			},
			args: args{
				txSimContext: func(ctrl *gomock.Controller) protocol.TxSimContext {
					txSimContext := mock.NewMockTxSimContext(ctrl)
					//txSimContext.EXPECT().GetSender().Return(&acPb.Member{
					//	OrgId:      "org1",
					//	MemberType: acPb.MemberType_DID,
					//	MemberInfo: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
					//})
					return txSimContext
				}(ctrl),
				tx: &commonPb.Transaction{
					Payer: &commonPb.EndorsementEntry{
						Signer: &acPb.Member{
							OrgId:      "org1",
							MemberType: acPb.MemberType_DID,
							MemberInfo: []byte("-----BEGIN CERTIFICATE-----\nMIICnTCCAkSgAwIBAgIDBMXxMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIyMDMwMTEyMDIyNloXDTMy\nMDIyNzEyMDIyNlowgYoxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwlyb290LWNlcnQxIjAgBgNVBAMTGWNhLnd4LW9yZzEuY2hhaW5t\nYWtlci5vcmcwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARcGEnTDAcVf1duITwI\nSI2S5ZC0jdQOyhUD5iA2Vv1XnG0GIEZNtJMzLJYunZCHg0qwFF9HVDTtgUWwzdX8\nc8VBo4GWMIGTMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud\nDgQiBCBzyXvo2oPh1h0KIBepfopq2/Rhd9b8f5EhKeJbUUnsLzBFBgNVHREEPjA8\ngg5jaGFpbm1ha2VyLm9yZ4IJbG9jYWxob3N0ghljYS53eC1vcmcxLmNoYWlubWFr\nZXIub3JnhwR/AAABMAoGCCqGSM49BAMCA0cAMEQCICFvGIvxhdzkuMsjkgVRNPM5\nfy4KHLG8pDLzj8bn2dGqAiB0ZBA1d/uBBPNJAf3s1fyB4R3P/gdKBiuDAvZ94zn3\nZg==\n-----END CERTIFICATE-----\n"),
						},
					},
				},
			},
			want: func() []byte {
				return nil
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TxScheduler{
				VmManager:       tt.fields.VmManager,
				scheduleFinishC: tt.fields.scheduleFinishC,
				log:             tt.fields.log,
				chainConf:       tt.fields.chainConf,
				metricVMRunTime: tt.fields.metricVMRunTime,
				StoreHelper:     tt.fields.StoreHelper,
				keyReg:          tt.fields.keyReg,
				contractCache:   &sync.Map{},
			}

			got, err := ts.getPayerPk(tt.args.txSimContext, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPayerPk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPayerPk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
