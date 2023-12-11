/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifier

import (
	"encoding/hex"
	"fmt"

	"chainmaker.org/chainmaker/protocol/v2"

	"chainmaker.org/chainmaker-go/module/core/common/scheduler"

	batch "chainmaker.org/chainmaker/txpool-batch/v2"

	"github.com/gogo/protobuf/proto"

	"chainmaker.org/chainmaker-go/module/consensus"
	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker-go/module/core/provider/conf"
	commonErrors "chainmaker.org/chainmaker/common/v2/errors"
	"chainmaker.org/chainmaker/common/v2/monitor"
	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/localconf/v2"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	chainConfConfig "chainmaker.org/chainmaker/pb-go/v2/config"
	consensuspb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/utils/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var ModuleNameCore = "Core"

// BlockVerifierImpl implements BlockVerifier interface.
// Verify block and transactions.
// nolint: structcheck,unused
type BlockVerifierImpl struct {
	chainId         string                   // chain id, to identity this chain
	msgBus          msgbus.MessageBus        // message bus
	txScheduler     protocol.TxScheduler     // scheduler orders tx batch into DAG form and returns a block
	snapshotManager protocol.SnapshotManager // snapshot manager
	ledgerCache     protocol.LedgerCache     // ledger cache
	blockchainStore protocol.BlockchainStore // blockchain store

	reentrantLocks *common.ReentrantLocks         // reentrant lock for avoid concurrent verify block
	proposalCache  protocol.ProposalCache         // proposal cache
	chainConf      protocol.ChainConf             // chain config
	ac             protocol.AccessControlProvider // access control manager
	log            protocol.Logger                // logger
	txPool         protocol.TxPool                // tx pool to check if tx is duplicate
	txFilter       protocol.TxFilter              // tx pool to check if tx is duplicate
	//mu             sync.Mutex                     // to avoid concurrent map modify
	verifierBlock *common.VerifierBlock
	storeHelper   conf.StoreHelper

	metricBlockVerifyTime *prometheus.HistogramVec // metrics monitor
	netService            protocol.NetService
}

type BlockVerifierConfig struct {
	ChainId         string
	MsgBus          msgbus.MessageBus
	SnapshotManager protocol.SnapshotManager
	BlockchainStore protocol.BlockchainStore
	LedgerCache     protocol.LedgerCache
	TxScheduler     protocol.TxScheduler
	ProposedCache   protocol.ProposalCache
	ChainConf       protocol.ChainConf
	AC              protocol.AccessControlProvider
	TxPool          protocol.TxPool
	VmMgr           protocol.VmManager
	StoreHelper     conf.StoreHelper
	NetService      protocol.NetService
	TxFilter        protocol.TxFilter
}

func NewBlockVerifier(config BlockVerifierConfig, log protocol.Logger) (protocol.BlockVerifier, error) {
	v := &BlockVerifierImpl{
		chainId:         config.ChainId,
		msgBus:          config.MsgBus,
		txScheduler:     config.TxScheduler,
		snapshotManager: config.SnapshotManager,
		ledgerCache:     config.LedgerCache,
		blockchainStore: config.BlockchainStore,
		reentrantLocks: &common.ReentrantLocks{
			ReentrantLocks: make(map[string]interface{}),
		},
		proposalCache: config.ProposedCache,
		chainConf:     config.ChainConf,
		ac:            config.AC,
		log:           log,
		txPool:        config.TxPool,
		storeHelper:   config.StoreHelper,
		netService:    config.NetService,
		txFilter:      config.TxFilter,
	}

	verifyConf := &common.VerifierBlockConf{
		ChainConf:       config.ChainConf,
		Log:             log,
		LedgerCache:     config.LedgerCache,
		Ac:              config.AC,
		SnapshotManager: config.SnapshotManager,
		VmMgr:           config.VmMgr,
		TxPool:          config.TxPool,
		BlockchainStore: config.BlockchainStore,
		ProposalCache:   config.ProposedCache,
		StoreHelper:     config.StoreHelper,
		TxScheduler:     config.TxScheduler,
		TxFilter:        config.TxFilter,
	}
	v.verifierBlock = common.NewVerifierBlock(verifyConf)

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		v.metricBlockVerifyTime = monitor.NewHistogramVec(monitor.SUBSYSTEM_CORE_VERIFIER, "metric_block_verify_time",
			"block verify time metric", []float64{0.005, 0.01, 0.015, 0.05, 0.1, 1, 2, 5, 10}, "chainId")
	}
	// v220_compat Deprecated
	config.ChainConf.AddWatch(v) //nolint: staticcheck
	config.MsgBus.Register(msgbus.ChainConfig, v)

	return v, nil
}

// VerifyBlockSync only maxbft use this method
func (v *BlockVerifierImpl) VerifyBlockSync(block *commonpb.Block,
	mode protocol.VerifyMode) (*consensuspb.VerifyResult, error) {
	//TODO implement me
	panic("only maxbft use this method")
}

// VerifyBlock to check if block is valid
func (v *BlockVerifierImpl) VerifyBlock(block *commonpb.Block, mode protocol.VerifyMode) (err error) {

	startTick := utils.CurrentTimeMillisSeconds()
	//检查是否为空
	if err = utils.IsEmptyBlock(block); err != nil {
		v.log.Error(err)
		v.log.Debugf("empty block. height:%+v, hash:%+v, chainId:%+v, preHash:%+v, signature:%+v",
			block.Header.BlockHeight, block.Header.BlockHash,
			block.Header.ChainId, block.Header.PreBlockHash, block.Header.Signature)
		return err
	}

	v.log.Debugf("verify receive [%d](%x,%d,%d), from sync %d",
		block.Header.BlockHeight, block.Header.BlockHash, block.Header.TxCount, len(block.Txs), mode)
	// avoid concurrent verify, only one block hash can be verified at the same time
	//避免并发验证
	if !v.reentrantLocks.Lock(string(block.Header.BlockHash)) {
		v.log.Warnf("block(%d,%x) concurrent verify, yield", block.Header.BlockHeight, block.Header.BlockHash)
		return commonErrors.ErrConcurrentVerify
	}
	defer v.reentrantLocks.Unlock(string(block.Header.BlockHash))

	// No duplicate verify
	//检查当前区块是否已经被重复验证过
	isRepeat := v.verifyRepeat(block, startTick, mode)
	if isRepeat {
		return nil
	}

	var contractEventMap map[string][]*commonpb.ContractEvent

	// avoid to recover the committed block.
	lastBlock, err := v.verifierBlock.FetchLastBlock(block)
	if err != nil {
		return err
	}

	startPoolTick := utils.CurrentTimeMillisSeconds()
	// 尝试恢复区块
	newBlock, batchIds, err := common.RecoverBlock(block, mode, v.chainConf, v.txPool, v.ac, v.netService, v.log)
	if err != nil {
		v.log.Errorf("RecoverBlock failed, err:%v", err)
		return err
	}
	lastPool := utils.CurrentTimeMillisSeconds() - startPoolTick
	//使用 validateBlock 方法对新区块进行验证，验证包括交易的合法性和相关的读写集。
	txRWSetMap, contractEventMap, timeLasts, rwSetVerifyFailTx, err := v.validateBlock(newBlock, lastBlock, mode)
	if err != nil {
		v.log.Warnf("verify failed [%d](%x),preBlockHash:%x, %s",
			newBlock.Header.BlockHeight, newBlock.Header.BlockHash, newBlock.Header.PreBlockHash, err.Error())
		if protocol.CONSENSUS_VERIFY == mode {
			v.log.DebugDynamic(func() string {
				return fmt.Sprintf("publish verfiy failed rw set txs, block height:%d, err: %s",
					newBlock.Header.BlockHeight, err.Error())
			})
			v.msgBus.Publish(msgbus.VerifyResult, parseVerifyResult(newBlock, false, txRWSetMap, rwSetVerifyFailTx))
		}

		// rollback sql
		if sqlErr := v.storeHelper.RollBack(newBlock, v.blockchainStore); sqlErr != nil {
			v.log.Errorf("block [%d] rollback sql failed: %s", newBlock.Header.BlockHeight, sqlErr)
		}

		// clear snapshot when verify fail
		if snapErr := v.snapshotManager.ClearSnapshot(block); snapErr != nil {
			snapErr = fmt.Errorf("clear snapshot fail[%d](hash:%x), err: %s",
				block.Header.BlockHeight, block.Header.BlockHash, snapErr.Error())
			v.log.Error(snapErr)
		}

		return err
	}

	snapshot := v.snapshotManager.GetSnapshot(lastBlock, block)
	if scheduler.IsOptimizeChargeGasEnabled(v.chainConf) {
		if err = scheduler.VerifyOptimizeChargeGasTx(block, snapshot); err != nil {
			return err
		}
	}

	// sync mode, need to verify consensus vote signature
	// 如果是同步模式，则需要验证共识票的签名。
	beginConsensCheck := utils.CurrentTimeMillisSeconds()
	if protocol.SYNC_VERIFY == mode {
		if err = v.verifyVoteSig(newBlock); err != nil {
			v.log.Warnf("verify failed [%d](%x), votesig %s",
				newBlock.Header.BlockHeight, newBlock.Header.BlockHash, err.Error())
			return err
		}
	}
	consensusCheckUsed := utils.CurrentTimeMillisSeconds() - beginConsensCheck

	// verify success, cache block and read write set
	// solo need this，too！！！
	v.log.Debugf("set proposed block(%d,%x)", newBlock.Header.BlockHeight, newBlock.Header.BlockHash)
	if err = v.proposalCache.SetProposedBlock(newBlock, txRWSetMap, contractEventMap, false); err != nil {
		return err
	}

	// mark transactions in block as pending status in txpool
	if common.TxPoolType == batch.TxPoolType {
		v.txPool.AddTxBatchesToPendingCache(batchIds, newBlock.Header.BlockHeight)
	} else {
		v.txPool.AddTxsToPendingCache(newBlock.Txs, newBlock.Header.BlockHeight)
	}
	//根据验证模式，发布验证结果到消息总线。
	if protocol.CONSENSUS_VERIFY == mode {
		v.msgBus.Publish(msgbus.VerifyResult, parseVerifyResult(newBlock, true, txRWSetMap, nil))
	}
	elapsed := utils.CurrentTimeMillisSeconds() - startTick
	v.log.Infof("verify success [%d,%x]"+
		"(blockSig:%d,vm:%d,txVerify:%d,txRoot:%d,pool:%d,consensusCheckUsed:%d,total:%d)",
		newBlock.Header.BlockHeight, newBlock.Header.BlockHash, timeLasts[common.BlockSig], timeLasts[common.VM],
		timeLasts[common.TxVerify], timeLasts[common.TxRoot], lastPool, consensusCheckUsed, elapsed)

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		v.metricBlockVerifyTime.WithLabelValues(v.chainId).Observe(float64(elapsed) / 1000)
	}

	return nil
}

// VerifyBlockWithRwSets to check if block is valid
func (v *BlockVerifierImpl) VerifyBlockWithRwSets(block *commonpb.Block,
	rwsets []*commonpb.TxRWSet, mode protocol.VerifyMode) (err error) {

	startTick := utils.CurrentTimeMillisSeconds()
	if err = utils.IsEmptyBlock(block); err != nil {
		v.log.Error(err)
		v.log.Debugf("empty block. height:%+v, hash:%+v, chainId:%+v, preHash:%+v, signature:%+v",
			block.Header.BlockHeight, block.Header.BlockHash,
			block.Header.ChainId, block.Header.PreBlockHash, block.Header.Signature)
		return err
	}

	v.log.Debugf("verify receive [%d](%x,%d,%d), from sync %d",
		block.Header.BlockHeight, block.Header.BlockHash, block.Header.TxCount, len(block.Txs), mode)
	// avoid concurrent verify, only one block hash can be verified at the same time
	if !v.reentrantLocks.Lock(string(block.Header.BlockHash)) {
		v.log.Warnf("block(%d,%x) concurrent verify, yield", block.Header.BlockHeight, block.Header.BlockHash)
		return commonErrors.ErrConcurrentVerify
	}
	defer v.reentrantLocks.Unlock(string(block.Header.BlockHash))

	// No duplicate verify
	isRepeat := v.verifyRepeat(block, startTick, mode)
	if isRepeat {
		return nil
	}

	var contractEventMap map[string][]*commonpb.ContractEvent
	txRWSetMap := make(map[string]*commonpb.TxRWSet)
	for _, txRWSet := range rwsets {
		if txRWSet != nil {
			txRWSetMap[txRWSet.TxId] = txRWSet
		}
	}

	// avoid to recover the committed block.
	lastBlock, err := v.verifierBlock.FetchLastBlock(block)
	if err != nil {
		return err
	}

	startPoolTick := utils.CurrentTimeMillisSeconds()
	newBlock, batchIds, err := common.RecoverBlock(block, mode, v.chainConf, v.txPool, v.ac, v.netService, v.log)
	if err != nil {
		return err
	}
	lastPool := utils.CurrentTimeMillisSeconds() - startPoolTick
	contractEventMap, timeLasts, err := v.validateBlockWithRWSets(newBlock, lastBlock, mode, txRWSetMap)
	if err != nil {
		v.log.Warnf("verify failed [%d](%x),preBlockHash:%x, %s",
			newBlock.Header.BlockHeight, newBlock.Header.BlockHash, newBlock.Header.PreBlockHash, err.Error())

		// rollback sql
		if sqlErr := v.storeHelper.RollBack(newBlock, v.blockchainStore); sqlErr != nil {
			v.log.Errorf("block [%d] rollback sql failed: %s", newBlock.Header.BlockHeight, sqlErr)
		}
		return err
	}

	// sync mode, need to verify consensus vote signature
	beginConsensCheck := utils.CurrentTimeMillisSeconds()
	if mode == protocol.SYNC_VERIFY {
		if err = v.verifyVoteSig(newBlock); err != nil {
			v.log.Warnf("verify failed [%d](%x), votesig %s",
				newBlock.Header.BlockHeight, newBlock.Header.BlockHash, err.Error())
			return err
		}
	}
	consensusCheckUsed := utils.CurrentTimeMillisSeconds() - beginConsensCheck

	// verify success, cache block and read write set
	// solo need this，too！！！
	v.log.Debugf("set proposed block(%d,%x)", newBlock.Header.BlockHeight, newBlock.Header.BlockHash)
	if err = v.proposalCache.SetProposedBlock(newBlock, txRWSetMap, contractEventMap, false); err != nil {
		return err
	}

	// mark transactions in block as pending status in txpool
	if common.TxPoolType == batch.TxPoolType {
		v.txPool.AddTxBatchesToPendingCache(batchIds, newBlock.Header.BlockHeight)
	} else {
		v.txPool.AddTxsToPendingCache(newBlock.Txs, newBlock.Header.BlockHeight)
	}

	//if protocol.CONSENSUS_VERIFY == mode {
	//	v.msgBus.Publish(msgbus.VerifyResult, parseVerifyResult(newBlock, true, txRWSetMap, nil))
	//}
	elapsed := utils.CurrentTimeMillisSeconds() - startTick
	v.log.Infof("verify success [%d,%x]"+
		"(blockSig:%d,vm:%d,txVerify:%d,txRoot:%d,pool:%d,consensusCheckUsed:%d,total:%d)",
		newBlock.Header.BlockHeight, newBlock.Header.BlockHash, timeLasts[common.BlockSig], timeLasts[common.VM],
		timeLasts[common.TxVerify], timeLasts[common.TxRoot], lastPool, consensusCheckUsed, elapsed)

	if localconf.ChainMakerConfig.MonitorConfig.Enabled {
		v.metricBlockVerifyTime.WithLabelValues(v.chainId).Observe(float64(elapsed) / 1000)
	}
	return nil
}

var _ msgbus.Subscriber = (*BlockVerifierImpl)(nil)

// OnMessage contract event data is a []string, hexToString(proto.Marshal(data))
func (v *BlockVerifierImpl) OnMessage(msg *msgbus.Message) {
	switch msg.Topic {
	case msgbus.ChainConfig:
		dataStr, ok := msg.Payload.([]string)
		if !ok {
			return
		}
		dataBytes, err := hex.DecodeString(dataStr[0])
		if err != nil {
			v.log.Warn(err)
			return
		}
		chainConfig := &chainConfConfig.ChainConfig{}
		err = proto.Unmarshal(dataBytes, chainConfig)
		if err != nil {
			v.log.Warn(err)
			return
		}
		v.chainConf.ChainConfig().Block = chainConfig.Block
		protocol.ParametersValueMaxLength = chainConfig.Block.TxParameterSize * 1024 * 1024
		if chainConfig.Block.TxParameterSize <= 0 {
			protocol.ParametersValueMaxLength = protocol.DefaultParametersValueMaxSize * 1024 * 1024
		}
		v.log.Infof("[BlockVerifierImpl] receive msg, topic: %s, blockverify[%v]",
			msg.Topic.String(), v.chainConf.ChainConfig().Block)
	default:

	}
}

func (v *BlockVerifierImpl) OnQuit() {
	// nothing, implement Subscriber interface
}

func (v *BlockVerifierImpl) validateBlock(block, lastBlock *commonpb.Block, mode protocol.VerifyMode) (
	map[string]*commonpb.TxRWSet,
	map[string][]*commonpb.ContractEvent,
	map[string]int64,
	*common.RwSetVerifyFailTx, error) {
	hashType := v.chainConf.ChainConfig().Crypto.Hash
	timeLasts := make(map[string]int64)
	var err error
	var txCapacity uint32
	if scheduler.IsOptimizeChargeGasEnabled(v.chainConf) {
		txCapacity = v.chainConf.ChainConfig().Block.BlockTxCapacity + 1
	} else {
		txCapacity = v.chainConf.ChainConfig().Block.BlockTxCapacity
	}
	if block.Header.TxCount > txCapacity {
		return nil, nil, timeLasts, nil, fmt.Errorf("txcapacity expect <= %d, got %d)", txCapacity, block.Header.TxCount)
	}

	if err = common.IsTxCountValid(block); err != nil {
		return nil, nil, timeLasts, nil, err
	}

	// proposed height == proposing height - 1
	proposedHeight := lastBlock.Header.BlockHeight
	// check if this block height is 1 bigger than last block height
	lastBlockHash := lastBlock.Header.BlockHash
	err = common.CheckPreBlock(block, lastBlockHash, proposedHeight)
	if err != nil {
		return nil, nil, timeLasts, nil, err
	}

	return v.verifierBlock.ValidateBlock(block, lastBlock, hashType, timeLasts, mode)
}

func (v *BlockVerifierImpl) validateBlockWithRWSets(block, lastBlock *commonpb.Block, mode protocol.VerifyMode,
	txRWSetMap map[string]*commonpb.TxRWSet) (
	map[string][]*commonpb.ContractEvent, map[string]int64, error) {
	hashType := v.chainConf.ChainConfig().Crypto.Hash
	timeLasts := make(map[string]int64)

	// proposed height == proposing height - 1
	proposedHeight := lastBlock.Header.BlockHeight
	// check if this block height is 1 bigger than last block height
	lastBlockHash := lastBlock.Header.BlockHash
	err := common.CheckPreBlock(block, lastBlockHash, proposedHeight)
	if err != nil {
		return nil, timeLasts, err
	}

	return v.verifierBlock.ValidateBlockWithRWSets(block, hashType, timeLasts, txRWSetMap, mode)
}

func (v *BlockVerifierImpl) verifyVoteSig(block *commonpb.Block) error {
	return consensus.VerifyBlockSignatures(v.chainConf, v.ac, v.blockchainStore, block, v.ledgerCache)
}

func parseVerifyResult(block *commonpb.Block, isValid bool,
	txsRwSet map[string]*commonpb.TxRWSet, rwSetVerifyFailTxs *common.RwSetVerifyFailTx) *consensuspb.VerifyResult {
	verifyResult := &consensuspb.VerifyResult{
		VerifiedBlock: block,
		TxsRwSet:      txsRwSet,
	}
	if isValid {
		verifyResult.Code = consensuspb.VerifyResult_SUCCESS
		verifyResult.Msg = "OK"
	} else {
		verifyResult.Msg = "FAIL"
		verifyResult.Code = consensuspb.VerifyResult_FAIL
		if rwSetVerifyFailTxs != nil {
			verifyResult.RwSetVerifyFailTxs = &consensuspb.RwSetVerifyFailTxs{
				TxIds:       rwSetVerifyFailTxs.TxIds,
				BlockHeight: rwSetVerifyFailTxs.BlockHeight,
			}
		}
	}
	return verifyResult
}

func (v *BlockVerifierImpl) cutBlocks(blocksToCut []*commonpb.Block, blockToKeep *commonpb.Block) {
	if common.TxPoolType == batch.TxPoolType {
		err := v.cutBlocksForBatchPool(blocksToCut, blockToKeep)
		if err != nil {
			v.log.Warnf(fmt.Sprintf("cut block[%d] failed, err:%v", blockToKeep.Header.BlockHeight, err))
		}
		return
	}

	cutTxs := make([]*commonpb.Transaction, 0)
	txMap := make(map[string]interface{})
	for _, tx := range blockToKeep.Txs {
		txMap[tx.Payload.TxId] = struct{}{}
	}
	for _, blockToCut := range blocksToCut {
		v.log.Infof("cut block hash: %x, height: %v", blockToCut.Header.BlockHash, blockToCut.Header.BlockHeight)
		for _, txToCut := range blockToCut.Txs {
			if _, ok := txMap[txToCut.Payload.TxId]; ok {
				// this transaction is kept, do NOT cut it.
				continue
			}
			v.log.Debugf("cut tx hash: %s", txToCut.Payload.TxId)
			cutTxs = append(cutTxs, txToCut)
		}
	}
	if len(cutTxs) > 0 {
		common.RetryAndRemoveTxs(v.txPool, cutTxs, nil, v.log)
	}
}

func (v *BlockVerifierImpl) cutBlocksForBatchPool(blocksToCut []*commonpb.Block, blockToKeep *commonpb.Block) error {

	keepBatchIdsMap := make(map[string]interface{})
	batchIds, _, err := common.GetBatchIds(blockToKeep)
	if err != nil {
		v.log.Errorf("get batch ids from keep block[%d,%x] failed, err:%v",
			blockToKeep.Header.BlockHeight, blockToKeep.Header.BlockHash, err)
		return err
	}
	for _, batchId := range batchIds {
		keepBatchIdsMap[batchId] = struct{}{}
	}

	finalCutBatchIds := make([]string, 0)
	for _, blockToCut := range blocksToCut {
		v.log.Infof("cut block hash: %x, height: %v", blockToCut.Header.BlockHash, blockToCut.Header.BlockHeight)
		cutBatchIds, _, err := common.GetBatchIds(blockToCut)
		if err != nil {
			v.log.Warnf("get batch ids from removed block[%d,%x] failed, err:%v",
				blockToCut.Header.BlockHeight, blockToCut.Header.BlockHash, err)
			continue
		}
		for _, cutBatchId := range cutBatchIds {
			if _, ok := keepBatchIdsMap[cutBatchId]; ok {
				// this transaction is kept, do NOT cut it.
				continue
			}
			v.log.Debugf("cut tx batchId: %s", cutBatchId)
			finalCutBatchIds = append(finalCutBatchIds, cutBatchId)
		}
	}

	if len(finalCutBatchIds) > 0 {
		v.txPool.RetryAndRemoveTxBatches(finalCutBatchIds, nil)
	}

	return nil
}

// verifyRepeat to check if the block has verified before
func (v *BlockVerifierImpl) verifyRepeat(block *commonpb.Block, startTick int64,
	mode protocol.VerifyMode) (isRepeat bool) {
	b, txRwSet, _ := v.proposalCache.GetProposedBlock(block)
	// Return not repeat if SQL is not enabled or if it is not solo
	if b == nil {
		return false
	}
	isSqlDb := v.chainConf.ChainConfig().Contract.EnableSqlSupport
	if consensuspb.ConsensusType_SOLO != v.chainConf.ChainConfig().Consensus.Type || isSqlDb {
		elapsed := utils.CurrentTimeMillisSeconds() - startTick
		// the block has verified before
		v.log.Infof("verify success repeat [%d](%x), total: %d", block.Header.BlockHeight, block.Header.BlockHash, elapsed)
		if protocol.CONSENSUS_VERIFY == mode {
			// consensus mode, publish verify result to message bus
			v.msgBus.Publish(msgbus.VerifyResult, parseVerifyResult(block, true, txRwSet, nil))
		}
		lastBlock, _ := v.proposalCache.GetProposedBlockByHashAndHeight(
			block.Header.PreBlockHash, block.Header.BlockHeight-1)
		if lastBlock == nil {
			v.log.Debugf(
				"no pre-block be found, preHeight:%d, preBlockHash:%x",
				block.Header.BlockHeight-1,
				block.Header.PreBlockHash,
			)
			return true
		}
		cutBlocks := v.proposalCache.KeepProposedBlock(lastBlock.Header.BlockHash, lastBlock.Header.BlockHeight)
		if len(cutBlocks) > 0 {
			v.log.Infof(
				"received block hash: %s, height: %v",
				hex.EncodeToString(lastBlock.Header.BlockHash),
				lastBlock.Header.BlockHeight,
			)
			v.cutBlocks(cutBlocks, lastBlock)
		}

		return true
	}
	return false
}
