/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"chainmaker.org/chainmaker/pb-go/v2/consensus"

	configPb "chainmaker.org/chainmaker/pb-go/v2/config"

	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"chainmaker.org/chainmaker/vm/v2"

	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	"github.com/gogo/protobuf/proto"

	"github.com/hokaccha/go-prettyjson"

	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/vm-native/v2/accountmgr"
	"github.com/panjf2000/ants/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ScheduleTimeout        = 10
	ScheduleWithDagTimeout = 20
	blockVersion2300       = uint32(2300)
	blockVersion2310       = uint32(2030100)
	blockVersion2312       = uint32(2030102)
)

const (
	ErrMsgOfGasLimitNotSet = "field `GasLimit` must be set in payload."
)

// TxScheduler transaction scheduler structure
type TxScheduler struct {
	lock            sync.Mutex
	VmManager       protocol.VmManager
	scheduleFinishC chan bool
	log             protocol.Logger
	chainConf       protocol.ChainConf // chain config

	metricVMRunTime *prometheus.HistogramVec
	StoreHelper     conf.StoreHelper
	keyReg          *regexp.Regexp
	signer          protocol.SigningMember
	ledgerCache     protocol.LedgerCache
	contractCache   *sync.Map
}

// Transaction dependency in adjacency table representation
type dagNeighbors map[int]struct{}

type TxIdAndExecOrderType struct {
	string
	protocol.ExecOrderTxType
}

// Schedule according to a batch of transactions,
// and generating DAG according to the conflict relationship
func (ts *TxScheduler) Schedule(block *commonPb.Block, txBatch []*commonPb.Transaction,
	snapshot protocol.Snapshot) (map[string]*commonPb.TxRWSet, map[string][]*commonPb.ContractEvent, error) {

	ts.lock.Lock()
	defer ts.lock.Unlock()

	defer ts.releaseContractCache()

	var err error
	lastCommittedHeight, err := ts.ledgerCache.CurrentHeight()
	if err != nil {
		return nil, nil, err
	}

	if ts.chainConf.ChainConfig().Consensus.Type == consensus.ConsensusType_TBFT &&
		int64(block.Header.BlockHeight)-int64(lastCommittedHeight) < 1 {
		return nil, nil, fmt.Errorf("no need to schedule old block, ledger height: %d, block height: %d",
			lastCommittedHeight, block.Header.BlockHeight)
	}

	txBatchSize := len(txBatch)
	ts.log.Infof("schedule tx batch start, block %d, size = %d", block.Header.BlockHeight, txBatchSize)

	var goRoutinePool *ants.Pool
	poolCapacity := ts.StoreHelper.GetPoolCapacity()
	ts.log.Debugf("GetPoolCapacity() => %v", poolCapacity)
	if goRoutinePool, err = ants.NewPool(poolCapacity, ants.WithPreAlloc(false)); err != nil {
		return nil, nil, err
	}
	defer goRoutinePool.Release()

	timeoutC := time.After(ScheduleTimeout * time.Second)
	startTime := time.Now()

	runningTxC := make(chan *commonPb.Transaction, txBatchSize)
	finishC := make(chan bool)

	enableOptimizeChargeGas := IsOptimizeChargeGasEnabled(ts.chainConf)
	enableSenderGroup := ts.chainConf.ChainConfig().Core.EnableSenderGroup
	enableConflictsBitWindow, conflictsBitWindow := ts.initOptimizeTools(txBatch)
	var senderGroup *SenderGroup
	var senderCollection *SenderCollection
	if enableOptimizeChargeGas {
		ts.log.Debugf("before prepare `SenderCollection` ")
		senderCollection = NewSenderCollection(txBatch, snapshot, ts.log)
		ts.log.Debugf("end prepare `SenderCollection` ")
	} else if enableSenderGroup {
		ts.log.Debugf("before prepare `SenderGroup` ")
		senderGroup = NewSenderGroup(txBatch)
		ts.log.Debugf("end prepare `SenderGroup` ")
	}

	blockFingerPrint := string(utils.CalcBlockFingerPrintWithoutTx(block))
	ts.VmManager.BeforeSchedule(blockFingerPrint, block.Header.BlockHeight)
	defer ts.VmManager.AfterSchedule(blockFingerPrint, block.Header.BlockHeight)

	// launch the go routine to dispatch tx to runningTxC
	go func() {
		ts.log.Infof("before Schedule(...) dispatch txs of block(%v)", block.Header.BlockHeight)
		if len(txBatch) == 0 {
			finishC <- true
		} else {
			ts.dispatchTxs(
				block,
				txBatch,
				runningTxC,
				goRoutinePool,
				enableOptimizeChargeGas,
				senderCollection,
				enableSenderGroup,
				senderGroup,
				enableConflictsBitWindow,
				conflictsBitWindow,
				snapshot)
		}
		ts.log.Infof("end Schedule(...) dispatch txs of block(%v)", block.Header.BlockHeight)
	}()

	// Put the pending transaction into the running queue
	go func() {
		counter := 0
		for {
			select {
			case tx := <-runningTxC:
				ts.log.Debugf("prepare to submit running task for tx id:%s", tx.Payload.GetTxId())

				err := goRoutinePool.Submit(func() {
					handleTx(block, snapshot, ts, tx, runningTxC, finishC, goRoutinePool, txBatchSize,
						enableConflictsBitWindow, conflictsBitWindow, enableSenderGroup, senderGroup)
				})
				if err != nil {
					ts.log.Warnf("failed to submit running task, tx id:%s during schedule, %+v",
						tx.Payload.GetTxId(), err)
				}
			case <-timeoutC:
				ts.log.Debugf("Schedule(...) timeout ...")
				ts.scheduleFinishC <- true
				if !enableOptimizeChargeGas && enableSenderGroup {
					senderGroup.doneTxKeyC <- [32]byte{}
				}
				ts.log.Warnf("block [%d] schedule reached time limit", block.Header.BlockHeight)
				return
			case <-finishC:
				ts.log.Debugf("Schedule(...) finish ...")
				ts.scheduleFinishC <- true
				if !enableOptimizeChargeGas && enableSenderGroup {
					senderGroup.doneTxKeyC <- [32]byte{}
				}
				return
			}
			counter++
			ts.log.Debugf("schedule tx run %d times ... ", counter)
		}
	}()

	// Wait for schedule finish signal
	<-ts.scheduleFinishC
	// Build DAG from read-write table
	snapshot.Seal()
	timeCostA := time.Since(startTime)
	block.Dag = snapshot.BuildDAG(ts.chainConf.ChainConfig().Contract.EnableSqlSupport, nil)

	// Execute special tx sequentially, and add to dag
	if len(snapshot.GetSpecialTxTable()) > 0 {
		ts.simulateSpecialTxs(block.Dag, snapshot, block, txBatchSize)
	}

	// if the block is not empty, append the charging gas tx
	if enableOptimizeChargeGas && snapshot.GetSnapshotSize() > 0 {
		ts.log.Debug("append charge gas tx to block ...")
		ts.appendChargeGasTx(block, snapshot, senderCollection)
	}

	timeCostB := time.Since(startTime)
	ts.log.Infof("schedule tx batch finished, block %d, success %d, txs execution cost %v, "+
		"dag building cost %v, total used %v, tps %v", block.Header.BlockHeight, len(block.Dag.Vertexes), timeCostA,
		timeCostB-timeCostA, timeCostB, float64(len(block.Dag.Vertexes))/(float64(timeCostB)/1e9))

	txRWSetMap := ts.getTxRWSetTable(snapshot, block)
	contractEventMap := ts.getContractEventMap(block)

	return txRWSetMap, contractEventMap, nil
}

// handleTx: run tx and apply tx sim context to snapshot
func handleTx(block *commonPb.Block, snapshot protocol.Snapshot,
	ts *TxScheduler, tx *commonPb.Transaction,
	runningTxC chan *commonPb.Transaction, finishC chan bool,
	goRoutinePool *ants.Pool, txBatchSize int,
	enableConflictsBitWindow bool, conflictsBitWindow *ConflictsBitWindow,
	enableSenderGroup bool, senderGroup *SenderGroup) {

	// If snapshot is sealed, no more transaction will be added into snapshot
	if snapshot.IsSealed() {
		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("handleTx(`%v`) snapshot has already sealed.", tx.GetPayload().TxId)
		})
		return
	}
	var start time.Time
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		start = time.Now()
	}

	// execute tx, and get
	// 1) the read/write set
	// 2) the result that telling if the invoke success.
	txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
	tx.Result = txSimContext.GetTxResult()
	ts.log.DebugDynamic(func() string {
		return fmt.Sprintf("handleTx(`%v`) => executeTx(...) => runVmSuccess = %v", tx.GetPayload().TxId, runVmSuccess)
	})

	// Apply failed means this tx's read set conflict with other txs' write set
	applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType,
		runVmSuccess, false)
	ts.log.DebugDynamic(func() string {
		return fmt.Sprintf("handleTx(`%v`) => ApplyTxSimContext(...) => snapshot.txTable = %v, applySize = %v",
			tx.GetPayload().TxId, len(snapshot.GetTxTable()), applySize)
	})

	// reduce the conflictsBitWindow size to eliminate the read/write set conflict
	if !applyResult {
		if enableConflictsBitWindow {
			ts.adjustPoolSize(goRoutinePool, conflictsBitWindow, ConflictTx)
		}

		runningTxC <- tx

		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("apply to snapshot failed, tx id:%s, result:%+v, apply count:%d",
				tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize)
		})

	} else {
		ts.handleApplyResult(enableConflictsBitWindow, enableSenderGroup,
			conflictsBitWindow, senderGroup, goRoutinePool, tx, start)

		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("apply to snapshot success, tx id:%s, result:%+v, apply count:%d",
				tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize)
		})
	}
	// If all transactions have been successfully added to dag
	if applySize >= txBatchSize {
		finishC <- true
	}
}

func (ts *TxScheduler) initOptimizeTools(
	txBatch []*commonPb.Transaction) (bool, *ConflictsBitWindow) {
	txBatchSize := len(txBatch)
	var conflictsBitWindow *ConflictsBitWindow
	enableConflictsBitWindow := ts.chainConf.ChainConfig().Core.EnableConflictsBitWindow

	ts.log.Infof("enable conflicts bit window: [%t]\n", enableConflictsBitWindow)

	if AdjustWindowSize*MinAdjustTimes > txBatchSize {
		enableConflictsBitWindow = false
	}
	if enableConflictsBitWindow {
		conflictsBitWindow = NewConflictsBitWindow(txBatchSize)
	}

	return enableConflictsBitWindow, conflictsBitWindow
}

// send txs from sender group
func (ts *TxScheduler) sendTxBySenderGroup(conflictsBitWindow *ConflictsBitWindow, senderGroup *SenderGroup,
	runningTxC chan *commonPb.Transaction, enableConflictsBitWindow bool) {
	// first round
	for _, v := range senderGroup.txsMap {
		runningTxC <- v[0]
	}
	// solve done tx channel
	for {
		k := <-senderGroup.doneTxKeyC
		if k == [32]byte{} {
			return
		}
		senderGroup.txsMap[k] = senderGroup.txsMap[k][1:]
		if len(senderGroup.txsMap[k]) > 0 {
			runningTxC <- senderGroup.txsMap[k][0]
		} else {
			delete(senderGroup.txsMap, k)
			if enableConflictsBitWindow {
				conflictsBitWindow.setMaxPoolCapacity(len(senderGroup.txsMap))
			}
		}
	}
}

// apply the read/write set to txSimContext,
// and adjust the go routine size
func (ts *TxScheduler) handleApplyResult(enableConflictsBitWindow bool, enableSenderGroup bool,
	conflictsBitWindow *ConflictsBitWindow, senderGroup *SenderGroup, goRoutinePool *ants.Pool,
	tx *commonPb.Transaction, start time.Time) {
	if enableConflictsBitWindow {
		ts.adjustPoolSize(goRoutinePool, conflictsBitWindow, NormalTx)
	}
	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		elapsed := time.Since(start)
		ts.metricVMRunTime.WithLabelValues(tx.Payload.ChainId).Observe(elapsed.Seconds())
	}
	if enableSenderGroup {
		hashKey, _ := getSenderHashKey(tx)
		senderGroup.doneTxKeyC <- hashKey
	}
}

func (ts *TxScheduler) getTxRWSetTable(snapshot protocol.Snapshot, block *commonPb.Block) map[string]*commonPb.TxRWSet {
	block.Txs = snapshot.GetTxTable()
	txRWSetTable := snapshot.GetTxRWSetTable()
	txRWSetMap := make(map[string]*commonPb.TxRWSet, len(txRWSetTable))
	for _, txRWSet := range txRWSetTable {
		if txRWSet != nil {
			txRWSetMap[txRWSet.TxId] = txRWSet
		}
	}
	//ts.dumpDAG(block.Dag, block.Txs)
	if localconf.ChainMakerConfig.SchedulerConfig.RWSetLog {
		result, _ := prettyjson.Marshal(txRWSetMap)
		ts.log.Infof("schedule rwset :%s, dag:%+v", result, block.Dag)
	}
	return txRWSetMap
}

func (ts *TxScheduler) getContractEventMap(block *commonPb.Block) map[string][]*commonPb.ContractEvent {
	contractEventMap := make(map[string][]*commonPb.ContractEvent, len(block.Txs))
	for _, tx := range block.Txs {
		event := tx.Result.ContractResult.ContractEvent
		contractEventMap[tx.Payload.TxId] = event
	}
	return contractEventMap
}

// SimulateWithDag based on the dag in the block, perform scheduling and execution transactions
func (ts *TxScheduler) SimulateWithDag(block *commonPb.Block, snapshot protocol.Snapshot) (
	map[string]*commonPb.TxRWSet, map[string]*commonPb.Result, error) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	defer ts.releaseContractCache()

	var (
		startTime  = time.Now()
		txRWSetMap = make(map[string]*commonPb.TxRWSet, len(block.Txs))
	)
	if block.Header.BlockVersion >= blockVersion2300 && len(block.Txs) != len(block.Dag.Vertexes) {
		ts.log.Warnf("found dag size mismatch txs length in "+
			"block[%x] dag:%d, txs:%d", block.Header.BlockHash, len(block.Dag.Vertexes), len(block.Txs))
		return nil, nil, fmt.Errorf("found dag size mismatch txs length in "+
			"block[%x] dag:%d, txs:%d", block.Header.BlockHash, len(block.Dag.Vertexes), len(block.Txs))
	}
	if len(block.Txs) == 0 {
		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("no txs in block[%x] when simulate", block.Header.BlockHash)
		})
		return txRWSetMap, snapshot.GetTxResultMap(), nil
	}
	ts.log.Infof("simulate with dag start, size %d", len(block.Txs))
	txMapping := make(map[int]*commonPb.Transaction, len(block.Txs))
	for index, tx := range block.Txs {
		txMapping[index] = tx
	}

	// Construct the adjacency list of dag, which describes the subsequent adjacency transactions of all transactions
	dag := block.Dag
	txIndexBatch, dagRemain, reverseDagRemain, err := ts.initSimulateDag(dag)
	if err != nil {
		ts.log.Warnf("initialize simulate dag error:%s", err)
		return nil, nil, err
	}

	txBatchSize := len(dag.Vertexes)
	runningTxC := make(chan int, txBatchSize)
	doneTxC := make(chan int, txBatchSize)

	timeoutC := time.After(ScheduleWithDagTimeout * time.Second)
	finishC := make(chan bool)

	txExecOrderTypeC := make(chan TxIdAndExecOrderType, txBatchSize)

	var goRoutinePool *ants.Pool
	if goRoutinePool, err = ants.NewPool(len(block.Txs), ants.WithPreAlloc(true)); err != nil {
		return nil, nil, err
	}
	defer goRoutinePool.Release()

	ts.log.DebugDynamic(func() string {
		return fmt.Sprintf("block [%d] simulate with dag first batch size:%d, total batch size:%d",
			block.Header.BlockHeight, len(txIndexBatch), txBatchSize)
	})

	blockFingerPrint := string(utils.CalcBlockFingerPrintWithoutTx(block))
	ts.VmManager.BeforeSchedule(blockFingerPrint, block.Header.BlockHeight)
	defer ts.VmManager.AfterSchedule(blockFingerPrint, block.Header.BlockHeight)

	go func() {
		for _, tx := range txIndexBatch {
			runningTxC <- tx
		}
	}()
	go func() {
		for {
			select {
			case txIndex := <-runningTxC:
				tx := txMapping[txIndex]
				ts.log.Debugf("simulate with dag, prepare to submit running task for tx id:%s", tx.Payload.GetTxId())
				err = goRoutinePool.Submit(func() {
					handleTxInSimulateWithDag(block, snapshot, ts, tx, txIndex, doneTxC, finishC, txExecOrderTypeC, txBatchSize)
				})
				if err != nil {
					ts.log.Warnf("failed to submit tx id %s during simulate with dag, %+v",
						tx.Payload.GetTxId(), err)
				}
			case doneTxIndex := <-doneTxC:
				txIndexBatchAfterShrink := ts.shrinkDag(doneTxIndex, dagRemain, reverseDagRemain)
				ts.log.Debugf("block [%d] simulate with dag, pop next tx index batch size:%d, dagRemain size:%d",
					block.Header.BlockHeight, len(txIndexBatchAfterShrink), len(dagRemain))
				for _, tx := range txIndexBatchAfterShrink {
					runningTxC <- tx
				}
			case <-finishC:
				ts.log.Debugf("block [%d] simulate with dag finish", block.Header.BlockHeight)
				ts.scheduleFinishC <- true
				return
			case <-timeoutC:
				ts.log.Errorf("block [%d] simulate with dag timeout", block.Header.BlockHeight)
				ts.scheduleFinishC <- true
				return
			}
		}
	}()

	<-ts.scheduleFinishC
	snapshot.Seal()
	timeUsed := time.Since(startTime)
	ts.log.Infof("simulate with dag finished, block %d, size %d, time used %v, tps %v", block.Header.BlockHeight,
		len(block.Txs), timeUsed, float64(len(block.Txs))/(float64(timeUsed)/1e9))

	// Return the read and write set after the scheduled execution
	for _, txRWSet := range snapshot.GetTxRWSetTable() {
		if txRWSet != nil {
			txRWSetMap[txRWSet.TxId] = txRWSet
		}
	}
	txExecOrderTypeMap := make(map[string]protocol.ExecOrderTxType, len(block.Txs))
	// we only receive fixed number of elements from this channel since we process unreceived things
	// and return error in later parts
	length := len(txExecOrderTypeC)
	for i := 0; i < length; i++ {
		t := <-txExecOrderTypeC
		txExecOrderTypeMap[t.string] = t.ExecOrderTxType
	}
	err = ts.compareDag(block, snapshot, txRWSetMap, txExecOrderTypeMap)
	if err != nil {
		return nil, nil, err
	}
	if localconf.ChainMakerConfig.SchedulerConfig.RWSetLog {
		result, _ := prettyjson.Marshal(txRWSetMap)
		ts.log.Infof("simulate with dag rwset :%s, dag: %+v", result, block.Dag)
	}
	return txRWSetMap, snapshot.GetTxResultMap(), nil
}

func (ts *TxScheduler) initSimulateDag(dag *commonPb.DAG) (
	[]int, map[int]dagNeighbors, map[int]dagNeighbors, error) {
	dagRemain := make(map[int]dagNeighbors, len(dag.Vertexes))
	reverseDagRemain := make(map[int]dagNeighbors, len(dag.Vertexes)*4)
	var txIndexBatch []int
	for txIndex, neighbors := range dag.Vertexes {
		if neighbors == nil {
			return nil, nil, nil, fmt.Errorf("dag has nil neighbor")
		}
		if len(neighbors.Neighbors) == 0 {
			txIndexBatch = append(txIndexBatch, txIndex)
			continue
		}
		dn := make(dagNeighbors)
		for index, neighbor := range neighbors.Neighbors {
			if index > 0 {
				if neighbors.Neighbors[index-1] >= neighbor {
					return nil, nil, nil, fmt.Errorf("dag neighbors not strict increasing, neighbors: %v", neighbors.Neighbors)
				}
			}
			if int(neighbor) >= txIndex {
				return nil, nil, nil, fmt.Errorf("dag has neighbor >= txIndex, txIndex: %d, neighbor: %d", txIndex, neighbor)
			}
			dn[int(neighbor)] = struct{}{}
			if _, ok := reverseDagRemain[int(neighbor)]; !ok {
				reverseDagRemain[int(neighbor)] = make(dagNeighbors)
			}
			reverseDagRemain[int(neighbor)][txIndex] = struct{}{}
		}
		dagRemain[txIndex] = dn
	}
	return txIndexBatch, dagRemain, reverseDagRemain, nil
}

func handleTxInSimulateWithDag(
	block *commonPb.Block, snapshot protocol.Snapshot,
	ts *TxScheduler, tx *commonPb.Transaction, txIndex int,
	doneTxC chan int, finishC chan bool,
	txExecOrderTypeC chan TxIdAndExecOrderType, txBatchSize int) {
	txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
	// send specialTxType BEFORE snapshot.ApplyTxSimContext which has a lock, ensuring that all txs have it
	// and eliminating race condition
	txExecOrderTypeC <- TxIdAndExecOrderType{tx.Payload.GetTxId(), specialTxType}
	// if apply failed means this tx's read set conflict with other txs' write set
	applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType, runVmSuccess, true)
	if !applyResult {
		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("failed to apply snapshot for tx id:%s, shouldn't have its rwset", tx.Payload.TxId)
		})
		// apply fails in verification, make it done rather than retry it
		doneTxC <- txIndex
	} else {
		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("apply to snapshot for tx id:%s, result:%+v, apply count:%d, tx batch size:%d",
				tx.Payload.GetTxId(), txSimContext.GetTxResult(), applySize, txBatchSize)
		})
		doneTxC <- txIndex
	}
	// If all transactions in current batch have been successfully added to dag
	if applySize >= txBatchSize {
		ts.log.DebugDynamic(func() string {
			return fmt.Sprintf("finished 1 batch, apply size:%d, tx batch size:%d", applySize, txBatchSize)
		})
		finishC <- true
	}
}

func (ts *TxScheduler) adjustPoolSize(pool *ants.Pool, conflictsBitWindow *ConflictsBitWindow, txExecType TxExecType) {
	newPoolSize := conflictsBitWindow.Enqueue(txExecType, pool.Cap())
	if newPoolSize == -1 {
		return
	}
	pool.Tune(newPoolSize)
}

func (ts *TxScheduler) executeTx(
	tx *commonPb.Transaction, snapshot protocol.Snapshot, block *commonPb.Block) (
	protocol.TxSimContext, protocol.ExecOrderTxType, bool) {
	txSimContext := vm.NewTxSimContext(ts.VmManager, snapshot, tx, block.Header.BlockVersion, ts.log)
	ts.log.DebugDynamic(func() string {
		return fmt.Sprintf("NewTxSimContext finished for tx id:%s", tx.Payload.GetTxId())
	})
	//ts.log.DebugDynamic(func() string {
	//	return fmt.Sprintf("tx.Result = %v", tx.Result)
	//})

	enableGas := ts.checkGasEnable()
	enableOptimizeChargeGas := IsOptimizeChargeGasEnabled(ts.chainConf)
	blockVersion := block.GetHeader().BlockVersion

	if blockVersion >= 2300 {
		if !ts.guardForExecuteTx2300(tx, txSimContext, enableGas, enableOptimizeChargeGas, snapshot) {
			return txSimContext, protocol.ExecOrderTxTypeNormal, false
		}
	} else if blockVersion >= 2220 {
		if !ts.guardForExecuteTx2220(tx, txSimContext, enableGas, enableOptimizeChargeGas) {
			return txSimContext, protocol.ExecOrderTxTypeNormal, false
		}
	}

	runVmSuccess := true
	var txResult *commonPb.Result
	var err error
	var specialTxType protocol.ExecOrderTxType

	ts.log.Debugf("run vm start for tx:%s", tx.Payload.GetTxId())
	if blockVersion >= 2300 {
		if txResult, specialTxType, err = ts.runVM2300(tx, txSimContext, enableOptimizeChargeGas); err != nil {
			runVmSuccess = false
			ts.log.Errorf("failed to run vm for tx id:%s,contractName:%s, tx result:%+v, error:%+v",
				tx.Payload.GetTxId(), tx.Payload.ContractName, txResult, err)
		}
	} else if blockVersion >= 2220 {
		if txResult, specialTxType, err = ts.runVM2220(tx, txSimContext, enableOptimizeChargeGas); err != nil {
			runVmSuccess = false
			ts.log.Errorf("failed to run vm for tx id:%s,contractName:%s, tx result:%+v, error:%+v",
				tx.Payload.GetTxId(), tx.Payload.ContractName, txResult, err)
		}
	} else {
		if txResult, specialTxType, err = ts.runVM2210(tx, txSimContext); err != nil {
			runVmSuccess = false
			ts.log.Errorf("failed to run vm for tx id:%s,contractName:%s, tx result:%+v, error:%+v",
				tx.Payload.GetTxId(), tx.Payload.ContractName, txResult, err)
		}
	}
	ts.log.Debugf("run vm finished for tx:%s, runVmSuccess:%v, txResult = %v ", tx.Payload.TxId, runVmSuccess, txResult)
	txSimContext.SetTxResult(txResult)
	return txSimContext, specialTxType, runVmSuccess
}

func (ts *TxScheduler) simulateSpecialTxs(dag *commonPb.DAG, snapshot protocol.Snapshot, block *commonPb.Block,
	txBatchSize int) {
	specialTxs := snapshot.GetSpecialTxTable()
	specialTxsLen := len(specialTxs)
	var firstTx *commonPb.Transaction
	runningTxC := make(chan *commonPb.Transaction, specialTxsLen)
	scheduleFinishC := make(chan bool)
	timeoutC := time.After(ScheduleWithDagTimeout * time.Second)
	go func() {
		for _, tx := range specialTxs {
			runningTxC <- tx
		}
	}()

	go func() {
		for {
			select {
			case tx := <-runningTxC:
				// simulate tx
				txSimContext, specialTxType, runVmSuccess := ts.executeTx(tx, snapshot, block)
				tx.Result = txSimContext.GetTxResult()
				// apply tx
				applyResult, applySize := snapshot.ApplyTxSimContext(txSimContext, specialTxType, runVmSuccess, true)
				if !applyResult {
					ts.log.Debugf("failed to apply according to dag with tx %s ", tx.Payload.TxId)
					runningTxC <- tx
					continue
				}
				if firstTx == nil {
					firstTx = tx
					dagNeighbors := &commonPb.DAG_Neighbor{
						Neighbors: make([]uint32, 0, snapshot.GetSnapshotSize()-1),
					}
					for i := uint32(0); i < uint32(snapshot.GetSnapshotSize()-1); i++ {
						dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, i)
					}
					dag.Vertexes = append(dag.Vertexes, dagNeighbors)
				} else {
					dagNeighbors := &commonPb.DAG_Neighbor{
						Neighbors: make([]uint32, 0, 1),
					}
					dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, uint32(snapshot.GetSnapshotSize())-2)
					dag.Vertexes = append(dag.Vertexes, dagNeighbors)
				}
				if applySize >= txBatchSize {
					ts.log.Debugf("block [%d] schedule special txs finished, apply size:%d, len of txs:%d, "+
						"len of special txs:%d", block.Header.BlockHeight, applySize, txBatchSize, specialTxsLen)
					scheduleFinishC <- true
					return
				}
			case <-timeoutC:
				ts.log.Errorf("block [%d] schedule special txs timeout", block.Header.BlockHeight)
				scheduleFinishC <- true
				return
			}
		}
	}()
	<-scheduleFinishC
}

func (ts *TxScheduler) shrinkDag(txIndex int, dagRemain map[int]dagNeighbors,
	reverseDagRemain map[int]dagNeighbors) []int {
	var txIndexBatch []int
	for k := range reverseDagRemain[txIndex] {
		delete(dagRemain[k], txIndex)
		if len(dagRemain[k]) == 0 {
			txIndexBatch = append(txIndexBatch, k)
			delete(dagRemain, k)
		}
	}
	delete(reverseDagRemain, txIndex)
	return txIndexBatch
}

func (ts *TxScheduler) Halt() {
	ts.scheduleFinishC <- true
}

// nolint: unused
func (ts *TxScheduler) dumpDAG(dag *commonPb.DAG, txs []*commonPb.Transaction) {
	dagString := "digraph DAG {\n"
	for i, ns := range dag.Vertexes {
		if len(ns.Neighbors) == 0 {
			dagString += fmt.Sprintf("id_%s -> begin;\n", txs[i].Payload.TxId[:8])
			continue
		}
		for _, n := range ns.Neighbors {
			dagString += fmt.Sprintf("id_%s -> id_%s;\n", txs[i].Payload.TxId[:8], txs[n].Payload.TxId[:8])
		}
	}
	dagString += "}"
	ts.log.Infof("Dump Dag: %s", dagString)
}

func (ts *TxScheduler) chargeGasLimit(accountMangerContract *commonPb.Contract, tx *commonPb.Transaction,
	txSimContext protocol.TxSimContext, contractName, method string, pk []byte,
	result *commonPb.Result) (re *commonPb.Result, err error) {
	if ts.checkGasEnable() &&
		ts.checkNativeFilter(txSimContext.GetBlockVersion(), contractName, method, tx, txSimContext.GetSnapshot()) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		var code commonPb.TxStatusCode
		var runChargeGasContract *commonPb.ContractResult
		var limit uint64
		if tx.Payload.Limit == nil {
			err = errors.New("tx payload limit is nil")
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		limit = tx.Payload.Limit.GasLimit
		chargeParameters := map[string][]byte{
			accountmgr.ChargePublicKey: pk,
			accountmgr.ChargeGasAmount: []byte(strconv.FormatUint(limit, 10)),
		}
		ts.log.Debugf("【chargeGasLimit】%v, pk = %s, amount = %v", tx.Payload.TxId, pk, limit)
		runChargeGasContract, _, code = ts.VmManager.RunContract(
			accountMangerContract, syscontract.GasAccountFunction_CHARGE_GAS.String(),
			nil, chargeParameters, txSimContext, 0, commonPb.TxType_INVOKE_CONTRACT)
		if code != commonPb.TxStatusCode_SUCCESS {
			result.Code = code
			result.ContractResult = runChargeGasContract
			return result, errors.New(runChargeGasContract.Message)
		}
	} else {
		ts.log.Debugf("%s:%s no need to charge gas.", contractName, method)
	}
	return result, nil
}

func (ts *TxScheduler) checkRefundGas(accountMangerContract *commonPb.Contract, tx *commonPb.Transaction,
	txSimContext protocol.TxSimContext, contractName, method string, pk []byte,
	result *commonPb.Result, contractResultPayload *commonPb.ContractResult, enableOptimizeChargeGas bool) error {

	// get tx's gas limit
	limit, err := getTxGasLimit(tx)
	if err != nil {
		ts.log.Errorf("getTxGasLimit error: %v", err)
		result.Message = err.Error()
		return err
	}

	// compare the gas used with gas limit
	if limit < contractResultPayload.GasUsed {
		err = fmt.Errorf("gas limit is not enough, [limit:%d]/[gasUsed:%d]",
			limit, contractResultPayload.GasUsed)
		ts.log.Error(err.Error())
		result.ContractResult.Code = uint32(commonPb.TxStatusCode_CONTRACT_FAIL)
		result.ContractResult.Message = err.Error()
		result.ContractResult.GasUsed = limit
		result.ContractResult.Result = nil
		result.ContractResult.ContractEvent = nil
		return err
	}

	if !enableOptimizeChargeGas {
		if _, err = ts.refundGas(accountMangerContract, tx, txSimContext, contractName, method, pk, result,
			contractResultPayload); err != nil {
			ts.log.Errorf("refund gas err is %v", err)
			if txSimContext.GetBlockVersion() >= blockVersion2300 {
				result.Code = commonPb.TxStatusCode_INTERNAL_ERROR
				result.Message = err.Error()
				result.ContractResult.Code = uint32(1)
				result.ContractResult.Message = err.Error()
				result.ContractResult.ContractEvent = nil
				return err
			}
		}
	}

	return nil
}

func (ts *TxScheduler) refundGas(accountMangerContract *commonPb.Contract, tx *commonPb.Transaction,
	txSimContext protocol.TxSimContext, contractName, method string, pk []byte,
	result *commonPb.Result, contractResultPayload *commonPb.ContractResult) (re *commonPb.Result, err error) {
	if ts.checkGasEnable() &&
		ts.checkNativeFilter(txSimContext.GetBlockVersion(), contractName, method, tx, txSimContext.GetSnapshot()) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		var code commonPb.TxStatusCode
		var refundGasContract *commonPb.ContractResult
		var limit uint64
		if tx.Payload.Limit == nil {
			err = errors.New("tx payload limit is nil")
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		limit = tx.Payload.Limit.GasLimit
		if limit < contractResultPayload.GasUsed {
			err = fmt.Errorf("gas limit is not enough, [limit:%d]/[gasUsed:%d]", limit, contractResultPayload.GasUsed)
			ts.log.Error(err.Error())
			result.Message = err.Error()
			return result, err
		}

		refundGas := limit - contractResultPayload.GasUsed
		ts.log.Debugf("refund gas [%d], gas used [%d]", refundGas, contractResultPayload.GasUsed)

		if refundGas == 0 {
			return result, nil
		}

		refundGasParameters := map[string][]byte{
			accountmgr.RechargeKey:       pk,
			accountmgr.RechargeAmountKey: []byte(strconv.FormatUint(refundGas, 10)),
		}

		refundGasContract, _, code = ts.VmManager.RunContract(
			accountMangerContract, syscontract.GasAccountFunction_REFUND_GAS_VM.String(),
			nil, refundGasParameters, txSimContext, 0, commonPb.TxType_INVOKE_CONTRACT)
		if code != commonPb.TxStatusCode_SUCCESS {
			result.Code = code
			result.ContractResult = refundGasContract
			return result, errors.New(refundGasContract.Message)
		}
	}
	return result, nil
}

func (ts *TxScheduler) getAccountMgrContractAndPk(txSimContext protocol.TxSimContext, tx *commonPb.Transaction,
	contractName, method string) (accountMangerContract *commonPb.Contract, pk []byte, err error) {
	if ts.checkGasEnable() &&
		ts.checkNativeFilter(txSimContext.GetBlockVersion(), contractName, method, tx, txSimContext.GetSnapshot()) &&
		tx.Payload.TxType == commonPb.TxType_INVOKE_CONTRACT {
		ts.log.Debugf("getAccountMgrContractAndPk => txSimContext.GetContractByName(`%s`)",
			syscontract.SystemContract_ACCOUNT_MANAGER.String())
		accountMangerContract, err = txSimContext.GetContractByName(syscontract.SystemContract_ACCOUNT_MANAGER.String())
		if err != nil {
			ts.log.Error(err.Error())
			return nil, nil, err
		}

		pk, err = ts.getPayerPk(txSimContext, tx)
		if err != nil {
			ts.log.Error(err.Error())
			return accountMangerContract, nil, err
		}
		return accountMangerContract, pk, err
	}
	return nil, nil, nil
}

func (ts *TxScheduler) checkGasEnable() bool {
	if ts.chainConf.ChainConfig() != nil && ts.chainConf.ChainConfig().AccountConfig != nil {
		ts.log.Debugf("chain config account config enable gas is:%v", ts.chainConf.ChainConfig().AccountConfig.EnableGas)
		return ts.chainConf.ChainConfig().AccountConfig.EnableGas
	}
	return false
}

// checkNativeFilter use snapshot instead of blockchainStore
func (ts *TxScheduler) checkNativeFilter(blockVersion uint32, contractName, method string,
	tx *commonPb.Transaction, snapshot protocol.Snapshot) bool {
	ts.log.Debugf("checkNativeFilter => contractName = %s, method = %s", contractName, method)

	// 用户合约，扣费
	if !utils.IsNativeContract(contractName) {
		return true
	}

	// add by Cai.Zhihong for compatible with v2.3.1.2
	if blockVersion < blockVersion2312 {
		// install & upgrade 系统合约扣费
		if method == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
			method == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
			return true
		}

		return ts.checkMultiSignFilterOld(contractName, method, tx, snapshot)
	}

	// install & upgrade 系统合约扣费
	if contractName == syscontract.SystemContract_CONTRACT_MANAGE.String() {
		if method == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
			method == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
			return true
		}
	}

	return ts.checkMultiSignFilter2312(contractName, method, tx, snapshot)
}

func (ts *TxScheduler) checkMultiSignFilterOld(
	contractName string, method string, tx *commonPb.Transaction, snapshot protocol.Snapshot) bool {
	if contractName == syscontract.SystemContract_MULTI_SIGN.String() &&
		method == syscontract.MultiSignFunction_TRIG.String() {
		if getMultiSignEnableManualRun(ts.chainConf.ChainConfig()) {
			var multiSignReqId []byte
			for _, kvpair := range tx.Payload.Parameters {
				if kvpair.Key == syscontract.MultiVote_TX_ID.String() {
					multiSignReqId = kvpair.Value
				}
			}
			multiSignInfoBytes, err := snapshot.GetKey(-1, contractName, multiSignReqId)
			if err != nil {
				ts.log.Errorf("read multi-sign failed, multiSignReqId = %v, err = %v", multiSignReqId, err)
				return true
			}

			multiSignInfo := &syscontract.MultiSignInfo{}
			err = proto.Unmarshal(multiSignInfoBytes, multiSignInfo)
			if err != nil {
				ts.log.Errorf("unmarshal MultiSignInfo failed, multiSignReqId = %v, err = %v", multiSignReqId, err)
				return true
			}

			var calleeContractName string
			var calleeMethod string
			for _, kvpair := range multiSignInfo.Payload.Parameters {
				if kvpair.Key == syscontract.MultiReq_SYS_CONTRACT_NAME.String() {
					calleeContractName = string(kvpair.Value)
				}
				if kvpair.Key == syscontract.MultiReq_SYS_METHOD.String() {
					calleeMethod = string(kvpair.Value)
				}
			}
			if calleeContractName == syscontract.SystemContract_CONTRACT_MANAGE.String() {
				if calleeMethod == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
					calleeMethod == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
					ts.log.Debugf("need charging gas, multiSignReqId = %v", multiSignReqId)
					return true
				}
			}
		}
	}
	return false
}

func (ts *TxScheduler) checkMultiSignFilter2312(
	contractName string, method string, tx *commonPb.Transaction, snapshot protocol.Snapshot) bool {
	return contractName == syscontract.SystemContract_MULTI_SIGN.String()
}

// todo: merge with getPayerPk
func getPayerPkFromTx(tx *commonPb.Transaction, snapshot protocol.Snapshot) (crypto.PublicKey, error) {

	var err error
	var pk []byte
	var publicKey crypto.PublicKey
	signingMember := getTxPayerSigner(tx)
	if signingMember == nil {
		err = errors.New(" can not find sender from tx ")
		return nil, err
	}

	switch signingMember.MemberType {
	case accesscontrol.MemberType_CERT:
		pk, err = publicKeyFromCert(signingMember.MemberInfo)
		if err != nil {
			return nil, err
		}
		publicKey, err = asym.PublicKeyFromPEM(pk)
		if err != nil {
			return nil, err
		}

	case accesscontrol.MemberType_CERT_HASH:
		var certInfo *commonPb.CertInfo
		infoHex := hex.EncodeToString(signingMember.MemberInfo)
		if certInfo, err = wholeCertInfoFromSnapshot(snapshot, infoHex); err != nil {
			return nil, fmt.Errorf(" can not load the whole cert info,member[%s],reason: %s", infoHex, err)
		}

		pk, err = publicKeyFromCert(certInfo.Cert)
		if err != nil {
			return nil, err
		}

		publicKey, err = asym.PublicKeyFromPEM(pk)
		if err != nil {
			return nil, err
		}

	case accesscontrol.MemberType_PUBLIC_KEY:
		pk = signingMember.MemberInfo
		publicKey, err = asym.PublicKeyFromPEM(pk)
		if err != nil {
			return nil, err
		}

	default:
		err = fmt.Errorf("invalid member type: %s", signingMember.MemberType)
		return nil, err
	}

	return publicKey, nil
}

func (ts *TxScheduler) getPayerPk(txSimContext protocol.TxSimContext, tx *commonPb.Transaction) ([]byte, error) {

	var err error
	var pk []byte
	sender := getTxPayerSigner(tx)
	if sender == nil {
		err = errors.New(" can not find sender from tx ")
		ts.log.Error(err.Error())
		return nil, err
	}

	switch sender.MemberType {
	case accesscontrol.MemberType_CERT:
		pk, err = publicKeyFromCert(sender.MemberInfo)
		if err != nil {
			ts.log.Error(err.Error())
			return nil, err
		}
	case accesscontrol.MemberType_CERT_HASH:
		var certInfo *commonPb.CertInfo
		infoHex := hex.EncodeToString(sender.MemberInfo)
		if certInfo, err = wholeCertInfo(txSimContext, infoHex); err != nil {
			ts.log.Error(err.Error())
			return nil, fmt.Errorf(" can not load the whole cert info,member[%s],reason: %s", infoHex, err)
		}

		if pk, err = publicKeyFromCert(certInfo.Cert); err != nil {
			ts.log.Error(err.Error())
			return nil, err
		}

	case accesscontrol.MemberType_PUBLIC_KEY:
		pk = sender.MemberInfo
	default:
		err = fmt.Errorf("invalid member type: %s", sender.MemberType)
		ts.log.Error(err.Error())
		return nil, err
	}

	return pk, nil
}

// dispatchTxs dispatch txs from:
//  1. senderCollection when flag `enableOptimizeChargeGas` was set
//  2. senderGroup when flag `enableOptimizeChargeGas` was not set, and flag `enableSenderGroup` was set
//  3. txBatch directly where no flags was set
//     to runningTxC
func (ts *TxScheduler) dispatchTxs(
	block *commonPb.Block,
	txBatch []*commonPb.Transaction,
	runningTxC chan *commonPb.Transaction,
	goRoutinePool *ants.Pool,
	enableOptimizeChargeGas bool,
	senderCollection *SenderCollection,
	enableSenderGroup bool,
	senderGroup *SenderGroup,
	enableConflictsBitWindow bool,
	conflictsBitWindow *ConflictsBitWindow,
	snapshot protocol.Snapshot) {
	if enableOptimizeChargeGas {
		ts.log.Debugf("before `SenderCollection` dispatch => ")
		ts.dispatchTxsInSenderCollection(block, senderCollection, runningTxC, snapshot)
		ts.log.Debugf("end `SenderCollection` dispatch => ")

	} else if enableSenderGroup {
		ts.log.Debugf("before `SenderGroup` dispatch => ")
		if enableConflictsBitWindow {
			conflictsBitWindow.setMaxPoolCapacity(len(senderGroup.txsMap))
		}
		goRoutinePool.Tune(len(senderGroup.txsMap))
		ts.sendTxBySenderGroup(conflictsBitWindow, senderGroup, runningTxC, enableConflictsBitWindow)
		ts.log.Debugf("end `SenderGroup` dispatch => ")

	} else {
		ts.log.Debugf("before `Normal` dispatch => ")
		for _, tx := range txBatch {
			runningTxC <- tx
		}
		ts.log.Debugf("end `Normal` dispatch => ")
	}
}

// dispatchTxsInSenderCollection dispatch txs from senderCollection to runningTxC chan
// if the balance less than gas limit, set the result of tx and dispatch this tx.
// use snapshot for newest data
func (ts *TxScheduler) dispatchTxsInSenderCollection(
	block *commonPb.Block,
	senderCollection *SenderCollection,
	runningTxC chan *commonPb.Transaction,
	snapshot protocol.Snapshot) {
	ts.log.Debugf("begin dispatchTxsInSenderCollection(...)")
	for addr, txCollection := range senderCollection.txsMap {
		ts.log.Debugf("%v => {balance: %v, tx size: %v}",
			addr, txCollection.accountBalance, len(txCollection.txs))
	}

	for addr, txCollection := range senderCollection.txsMap {
		balance := txCollection.accountBalance
		for _, tx := range txCollection.txs {
			ts.log.Debugf("dispatch sender collection tx => %s", tx.Payload)
			var gasLimit int64
			limit := tx.Payload.Limit
			txNeedChargeGas := ts.checkNativeFilter(
				block.GetHeader().GetBlockVersion(),
				tx.GetPayload().ContractName,
				tx.GetPayload().Method,
				tx, snapshot)
			ts.log.Debugf("tx need charge gas => %v", txNeedChargeGas)
			if limit == nil && txNeedChargeGas {
				// tx需要扣费，但是limit没有设置
				tx.Result = &commonPb.Result{
					Code: commonPb.TxStatusCode_GAS_LIMIT_NOT_SET,
					ContractResult: &commonPb.ContractResult{
						Code:    uint32(1),
						Result:  nil,
						Message: ErrMsgOfGasLimitNotSet,
						GasUsed: uint64(0),
					},
					RwSetHash: nil,
					Message:   ErrMsgOfGasLimitNotSet,
				}

				runningTxC <- tx
				continue

			} else if txNeedChargeGas && tx.Result != nil {
				runningTxC <- tx
				continue

			} else if !txNeedChargeGas {
				// tx 不需要扣费
				gasLimit = int64(0)
			} else {
				// tx 需要扣费，limit 正常设置
				gasLimit = int64(limit.GasLimit)
			}

			// if the balance less than gas limit, set the result ahead, working goroutine will never runVM for it.
			if balance-gasLimit < 0 {
				pkStr, _ := txCollection.publicKey.String()
				ts.log.Debugf("balance is too low to execute tx. address = %v, public key = %s", addr, pkStr)
				errMsg := fmt.Sprintf("`%s` has no enough balance to execute tx.", addr)
				tx.Result = &commonPb.Result{
					Code: commonPb.TxStatusCode_GAS_BALANCE_NOT_ENOUGH_FAILED,
					ContractResult: &commonPb.ContractResult{
						Code:    uint32(1),
						Result:  nil,
						Message: errMsg,
						GasUsed: uint64(0),
					},
					RwSetHash: nil,
					Message:   errMsg,
				}
			} else {
				balance = balance - gasLimit
			}

			runningTxC <- tx
		}
	}
}

// appendChargeGasTx include 3 step:
// 1) create a new charging gas tx
// 2) execute tx by calling native contract
// 3) append tx to DAG struct
func (ts *TxScheduler) appendChargeGasTx(
	block *commonPb.Block,
	snapshot protocol.Snapshot,
	senderCollection *SenderCollection) {
	ts.log.Debug("TxScheduler => appendChargeGasTx() => createChargeGasTx() begin ")
	tx, err := ts.createChargeGasTx(senderCollection)
	if err != nil {
		return
	}

	ts.log.Debug("TxScheduler => appendChargeGasTx() => executeGhargeGasTx() begin ")
	txSimContext := ts.executeChargeGasTx(tx, block, snapshot)
	tx.Result = txSimContext.GetTxResult()

	ts.log.Debug("TxScheduler => appendChargeGasTx() => appendChargeGasTxToDAG() begin ")
	ts.appendChargeGasTxToDAG(block.Dag, snapshot)
}

// signTxPayload sign charging tx with node's private key
func (ts *TxScheduler) signTxPayload(
	payload *commonPb.Payload) ([]byte, error) {

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// using the default hash type of the chain
	hashType := ts.chainConf.ChainConfig().GetCrypto().Hash
	return ts.signer.Sign(hashType, payloadBytes)
}

func (ts *TxScheduler) createChargeGasTx(
	senderCollection *SenderCollection) (*commonPb.Transaction, error) {

	// 构造参数
	parameters := make([]*commonPb.KeyValuePair, 0, len(senderCollection.txsMap))
	for address, txCollection := range senderCollection.txsMap {
		totalGasUsed := int64(0)
		for _, tx := range txCollection.txs {
			if tx.Result != nil {
				totalGasUsed += int64(tx.Result.ContractResult.GasUsed)
			}
		}
		keyValuePair := commonPb.KeyValuePair{
			Key:   address,
			Value: []byte(fmt.Sprintf("%d", totalGasUsed)),
		}
		parameters = append(parameters, &keyValuePair)
	}

	// 构造 Payload
	payload := &commonPb.Payload{
		ChainId:        ts.chainConf.ChainConfig().ChainId,
		TxType:         commonPb.TxType_INVOKE_CONTRACT,
		TxId:           utils.GetRandTxId(),
		Timestamp:      time.Now().Unix(),
		ExpirationTime: time.Now().Add(time.Second * 1).Unix(),
		ContractName:   syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		Method:         syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
		Parameters:     parameters,
		Sequence:       uint64(0),
		Limit:          &commonPb.Limit{GasLimit: uint64(0)},
	}

	// 对 Payload 签名
	signature, err := ts.signTxPayload(payload)
	if err != nil {
		ts.log.Errorf("createChargeGasTx => signTxPayload() error: %v", err.Error())
		return nil, err
	}

	// 构造 Transaction
	signingMember, err := ts.signer.GetMember()
	if err != nil {
		ts.log.Errorf("createChargeGasTx => GetMember() error: %v", err.Error())
		return nil, err
	}

	endorser := &commonPb.EndorsementEntry{
		Signer: &accesscontrol.Member{
			OrgId:      signingMember.OrgId,
			MemberInfo: signingMember.MemberInfo,
			MemberType: signingMember.MemberType,
		},
		Signature: signature,
	}

	return &commonPb.Transaction{
		Payload: payload,
		Sender: &commonPb.EndorsementEntry{
			Signer:    signingMember,
			Signature: signature,
		},
		Endorsers: []*commonPb.EndorsementEntry{endorser},
		Result:    nil,
	}, nil
}

func (ts *TxScheduler) executeChargeGasTx(
	tx *commonPb.Transaction,
	block *commonPb.Block,
	snapshot protocol.Snapshot) protocol.TxSimContext {

	txSimContext := vm.NewTxSimContext(ts.VmManager, snapshot, tx, block.Header.BlockVersion, ts.log)
	ts.log.Debugf("new tx for charging gas, id = %s", tx.Payload.GetTxId())

	result := &commonPb.Result{
		Code: commonPb.TxStatusCode_SUCCESS,
		ContractResult: &commonPb.ContractResult{
			Code:    uint32(0),
			Result:  nil,
			Message: "",
		},
		RwSetHash: nil,
	}

	ts.log.Debugf("executeChargeGasTx => txSimContext.GetContractByName(`%s`)", tx.Payload.ContractName)
	contract, err := txSimContext.GetContractByName(tx.Payload.ContractName)
	if err != nil {
		ts.log.Errorf("Get contract info by name[%s] error:%s", tx.Payload.ContractName, err)
		result.ContractResult.Message = err.Error()
		result.Code = commonPb.TxStatusCode_INVALID_PARAMETER
		result.ContractResult.Code = 1
		txSimContext.SetTxResult(result)
		return txSimContext
	}

	params := make(map[string][]byte, len(tx.Payload.Parameters))
	for _, item := range tx.Payload.Parameters {
		address := item.Key
		data := item.Value
		params[address] = data
	}

	// this native contract call will never failed
	contractResultPayload, _, txStatusCode := ts.VmManager.RunContract(contract, tx.Payload.Method, nil,
		params, txSimContext, 0, tx.Payload.TxType)
	if txStatusCode != commonPb.TxStatusCode_SUCCESS {
		panic("running the tx of charging gas will never failed.")
	}
	result.Code = txStatusCode
	result.ContractResult = contractResultPayload
	ts.log.Debugf("finished tx for charging gas, id = :%s, txStatusCode = %v", tx.Payload.TxId, txStatusCode)

	txSimContext.SetTxResult(result)
	snapshot.ApplyTxSimContext(
		txSimContext,
		protocol.ExecOrderTxTypeChargeGas,
		true, true)

	return txSimContext
}

// appendChargeGasTxToDAG append the tx to the DAG with dependencies on all tx.
func (ts *TxScheduler) appendChargeGasTxToDAG(
	dag *commonPb.DAG,
	snapshot protocol.Snapshot) {

	dagNeighbors := &commonPb.DAG_Neighbor{
		Neighbors: make([]uint32, 0, snapshot.GetSnapshotSize()-1),
	}
	for i := uint32(0); i < uint32(snapshot.GetSnapshotSize()-1); i++ {
		dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, i)
	}
	dag.Vertexes = append(dag.Vertexes, dagNeighbors)
}

func errResult(result *commonPb.Result, err error) (*commonPb.Result, protocol.ExecOrderTxType, error) {
	result.ContractResult.Message = err.Error()
	result.Code = commonPb.TxStatusCode_INVALID_PARAMETER
	result.ContractResult.Code = 1
	return result, protocol.ExecOrderTxTypeNormal, err
}

// parseUserAddress
func publicKeyFromCert(member []byte) ([]byte, error) {
	certificate, err := utils.ParseCert(member)
	if err != nil {
		return nil, err
	}
	pubKeyStr, err := certificate.PublicKey.String()
	if err != nil {
		return nil, err
	}
	return []byte(pubKeyStr), nil
}

func wholeCertInfo(txSimContext protocol.TxSimContext, certHash string) (*commonPb.CertInfo, error) {
	certBytes, err := txSimContext.Get(syscontract.SystemContract_CERT_MANAGE.String(), []byte(certHash))
	if err != nil {
		return nil, err
	}

	return &commonPb.CertInfo{
		Hash: certHash,
		Cert: certBytes,
	}, nil
}

type SenderGroup struct {
	txsMap     map[[32]byte][]*commonPb.Transaction
	doneTxKeyC chan [32]byte
}

func NewSenderGroup(txBatch []*commonPb.Transaction) *SenderGroup {
	return &SenderGroup{
		txsMap:     getSenderTxsMap(txBatch),
		doneTxKeyC: make(chan [32]byte, len(txBatch)),
	}
}

func getSenderTxsMap(txBatch []*commonPb.Transaction) map[[32]byte][]*commonPb.Transaction {
	senderTxsMap := make(map[[32]byte][]*commonPb.Transaction, len(txBatch))
	for _, tx := range txBatch {
		hashKey, _ := getSenderHashKey(tx)
		senderTxsMap[hashKey] = append(senderTxsMap[hashKey], tx)
	}
	return senderTxsMap
}

func getSenderHashKey(tx *commonPb.Transaction) ([32]byte, error) {
	sender := getTxPayerSigner(tx)
	keyBytes, err := sender.Marshal()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(keyBytes), nil
}

// publicKeyToAddress: generate address from public key, according to chainconfig parameter
func publicKeyToAddress(pk crypto.PublicKey, chainCfg *configPb.ChainConfig) (string, error) {

	publicKeyString, err := utils.PkToAddrStr(pk, chainCfg.Vm.AddrType, crypto.HashAlgoMap[chainCfg.Crypto.Hash])
	if err != nil {
		return "", err
	}

	if chainCfg.Vm.AddrType == configPb.AddrType_ZXL {
		publicKeyString = "ZX" + publicKeyString
	}
	return publicKeyString, nil
}

func getTxPayerSigner(tx *commonPb.Transaction) *accesscontrol.Member {
	payer := tx.GetPayer()
	// don't need version compatibility
	if payer == nil {
		payer = tx.GetSender()
	}
	return payer.GetSigner()
}

func wholeCertInfoFromSnapshot(snapshot protocol.Snapshot, certHash string) (*commonPb.CertInfo, error) {
	certBytes, err := snapshot.GetKey(-1, syscontract.SystemContract_CERT_MANAGE.String(), []byte(certHash))
	if err != nil {
		return nil, err
	}

	return &commonPb.CertInfo{
		Hash: certHash,
		Cert: certBytes,
	}, nil
}

// getTxGasLimit get the gas limit field from tx, and will return err when the gas limit field is not set.
func getTxGasLimit(tx *commonPb.Transaction) (uint64, error) {
	var limit uint64

	if tx.Payload.Limit == nil {
		return limit, errors.New("tx payload limit is nil")
	}

	limit = tx.Payload.Limit.GasLimit
	return limit, nil
}

func (ts *TxScheduler) verifyExecOrderTxType(block *commonPb.Block,
	txExecOrderTypeMap map[string]protocol.ExecOrderTxType) (uint32, uint32, uint32, error) {

	var txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount uint32
	for _, v := range txExecOrderTypeMap {
		switch v {
		case protocol.ExecOrderTxTypeNormal:
			txExecOrderNormalCount++
		case protocol.ExecOrderTxTypeIterator:
			txExecOrderIteratorCount++
		case protocol.ExecOrderTxTypeChargeGas:
			txExecOrderChargeGasCount++
		}
	}
	if (IsOptimizeChargeGasEnabled(ts.chainConf) && txExecOrderChargeGasCount != 1) ||
		(!IsOptimizeChargeGasEnabled(ts.chainConf) && txExecOrderChargeGasCount != 0) {
		return txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount,
			fmt.Errorf("charge gas enabled but charge gas tx is not 1")
	}
	// check type are all correct
	for i, tx := range block.Txs {
		t, ok := txExecOrderTypeMap[tx.Payload.GetTxId()]
		if !ok {
			return txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount,
				fmt.Errorf("cannot get tx ExecOrderTxType, txId:%s", tx.Payload.GetTxId())
		}
		var typeShouldBe protocol.ExecOrderTxType
		if uint32(i) < txExecOrderNormalCount {
			typeShouldBe = protocol.ExecOrderTxTypeNormal
		} else {
			typeShouldBe = protocol.ExecOrderTxTypeIterator
		}
		if IsOptimizeChargeGasEnabled(ts.chainConf) && uint32(i+1) == uint32(len(block.Txs)) {
			typeShouldBe = protocol.ExecOrderTxTypeChargeGas
		}
		if t != typeShouldBe {
			return txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount,
				fmt.Errorf("tx type mismatch, txId:%s, index:%d", tx.Payload.GetTxId(), i)
		}
	}
	return txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount, nil
}

func (ts *TxScheduler) compareDag(block *commonPb.Block, snapshot protocol.Snapshot,
	txRWSetMap map[string]*commonPb.TxRWSet, txExecOrderTypeMap map[string]protocol.ExecOrderTxType) error {
	if block.Header.BlockVersion < blockVersion2300 {
		return nil
	}
	startTime := time.Now()
	txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount, err :=
		ts.verifyExecOrderTxType(block, txExecOrderTypeMap)
	if err != nil {
		ts.log.Errorf("verifyExecOrderTxType has err:%s, tx type count:%d,%d,%d, block tx count:%d", err,
			txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount, block.Header.TxCount)
		return err
	}
	// rebuild and verify dag
	txRWSetTable := utils.RearrangeRWSet(block, txRWSetMap)
	if uint32(len(txRWSetTable)) != txExecOrderNormalCount+txExecOrderIteratorCount+txExecOrderChargeGasCount {
		return fmt.Errorf("txRWSetTable:%d != txExecOrderTypeCount:%d+%d+%d", len(txRWSetTable),
			txExecOrderNormalCount, txExecOrderIteratorCount, txExecOrderChargeGasCount)
	}

	// first, only build dag for normal tx
	txRWSetTable = txRWSetTable[0:txExecOrderNormalCount]
	dag := snapshot.BuildDAG(ts.chainConf.ChainConfig().Contract.EnableSqlSupport, txRWSetTable)
	// then, append special tx into dag
	if txExecOrderIteratorCount > 0 {
		appendSpecialTxsToDag(dag, txExecOrderIteratorCount)
	}
	// snapshot.GetSnapshotSize() > 0 prevent snapshot.GetSnapshotSize() - 1 overflow
	if IsOptimizeChargeGasEnabled(ts.chainConf) && snapshot.GetSnapshotSize() > 0 {
		ts.appendChargeGasTxToDAG(dag, snapshot)
	}
	equal, err := utils.IsDagEqual(block.Dag, dag)
	if err != nil {
		return err
	}
	if !equal {
		ts.log.Warnf("compare block dag (vertex:%d) with simulate dag (vertex:%d)",
			len(block.Dag.Vertexes), len(dag.Vertexes))
		return fmt.Errorf("simulate dag not equal to block dag")
	}
	timeUsed := time.Since(startTime)
	ts.log.Infof("compare dag finished, time used %v", timeUsed)
	return nil
}

func (ts *TxScheduler) releaseContractCache() {
	ts.contractCache.Range(func(key interface{}, value interface{}) bool {
		ts.contractCache.Delete(key)
		return true
	})
}

// appendSpecialTxsToDag similar to ts.simulateSpecialTxs except do not execute tx, only handle dag
// txExecOrderSpecialCount must >0
func appendSpecialTxsToDag(dag *commonPb.DAG, txExecOrderSpecialCount uint32) {
	txExecOrderNormalCount := uint32(len(dag.Vertexes))
	// the first special tx
	dagNeighbors := &commonPb.DAG_Neighbor{
		Neighbors: make([]uint32, 0, txExecOrderNormalCount),
	}
	for i := uint32(0); i < txExecOrderNormalCount; i++ {
		dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, i)
	}
	dag.Vertexes = append(dag.Vertexes, dagNeighbors)
	// other special tx
	for i := uint32(1); i < txExecOrderSpecialCount; i++ {
		dagNeighbors := &commonPb.DAG_Neighbor{
			Neighbors: make([]uint32, 0, 1),
		}
		// this special tx (txExecOrderNormalCount+i) only depend on previous special tx (txExecOrderNormalCount+i-1)
		dagNeighbors.Neighbors = append(dagNeighbors.Neighbors, txExecOrderNormalCount+i-1)
		dag.Vertexes = append(dag.Vertexes, dagNeighbors)
	}
}
