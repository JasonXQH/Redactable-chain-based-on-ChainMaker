package common

import (
	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/localconf/v2"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"fmt"
	"strconv"
)

/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

type ReplaceBlock struct {
	store           protocol.BlockchainStore
	log             protocol.Logger
	snapshotManager protocol.SnapshotManager
	ledgerCache     protocol.LedgerCache
	chainConf       protocol.ChainConf
	txFilter        protocol.TxFilter
	msgBus          msgbus.MessageBus
}

// CommitBlock the action that all consensus types do when a block is committed
func (pb *ReplaceBlock) ReplaceBlock(
	block *commonpb.Block,
	rwSetMap map[string]*commonpb.TxRWSet,
	conEventMap map[string][]*commonpb.ContractEvent) (
	dbLasts, snapshotLasts, confLasts, otherLasts, pubEventLasts, filterLasts int64, blockInfo *commonpb.BlockInfo,
	err error) {
	// record block
	rwSet := utils.RearrangeRWSet(block, rwSetMap)
	// record contract event
	events := rearrangeContractEvent(block, conEventMap)

	if block.Header.BlockVersion >= blockVersion230 {
		// notify chainConf to update config before put block
		startConfTick := utils.CurrentTimeMillisSeconds()
		if err = pb.NotifyMessage(block, events); err != nil {
			return 0, 0, 0, 0, 0, 0, nil, err
		}
		confLasts = utils.CurrentTimeMillisSeconds() - startConfTick
	}
	// put block
	startDBTick := utils.CurrentTimeMillisSeconds()
	if err = pb.store.ReplaceBlock(block, rwSet); err != nil {
		// if put db error, then panic
		pb.log.Error(err)
		panic(err)
	}
	pb.ledgerCache.SetLastCommittedBlock(block)
	dbLasts = utils.CurrentTimeMillisSeconds() - startDBTick

	// TxFilter adds
	filterLasts = utils.CurrentTimeMillisSeconds()
	// The default filter type does not run AddsAndSetHeight
	if localconf.ChainMakerConfig.TxFilter.Type != int32(config.TxFilterType_None) {
		err = pb.txFilter.AddsAndSetHeight(utils.GetTxIds(block.Txs), block.Header.GetBlockHeight())
		if err != nil {
			// if add filter error, then panic
			pb.log.Error(err)
			panic(err)
		}
	}
	filterLasts = utils.CurrentTimeMillisSeconds() - filterLasts

	// clear snapshot
	startSnapshotTick := utils.CurrentTimeMillisSeconds()
	if err = pb.snapshotManager.NotifyBlockCommitted(block); err != nil {
		err = fmt.Errorf("notify snapshot error [%d](hash:%x)",
			block.Header.BlockHeight, block.Header.BlockHash)
		pb.log.Error(err)
		return 0, 0, 0, 0, 0, 0, nil, err
	}
	snapshotLasts = utils.CurrentTimeMillisSeconds() - startSnapshotTick
	// v220_compat Deprecated
	if block.Header.BlockVersion < blockVersion230 {
		// notify chainConf to update config when config block committed
		startConfTick := utils.CurrentTimeMillisSeconds()
		if err = NotifyChainConf(block, pb.chainConf); err != nil {
			return 0, 0, 0, 0, 0, 0, nil, err
		}
		confLasts = utils.CurrentTimeMillisSeconds() - startConfTick
	}
	// contract event
	pubEventLasts = pb.publishContractEvent(block, events)

	// monitor
	startOtherTick := utils.CurrentTimeMillisSeconds()
	blockInfo = &commonpb.BlockInfo{
		Block:     block,
		RwsetList: rwSet,
	}
	otherLasts = utils.CurrentTimeMillisSeconds() - startOtherTick

	return
}

// publishContractEvent publish contract event, return time used
func (pb *ReplaceBlock) publishContractEvent(block *commonpb.Block, events []*commonpb.ContractEvent) int64 {
	if len(events) == 0 {
		return 0
	}

	startPublishContractEventTick := utils.CurrentTimeMillisSeconds()
	pb.log.DebugDynamic(func() string {
		return fmt.Sprintf("start publish contractEventsInfo: block[%d] ",
			block.Header.BlockHeight)
	})
	eventsInfo := make([]*commonpb.ContractEventInfo, 0, len(events))
	for _, t := range events {
		eventInfo := &commonpb.ContractEventInfo{
			BlockHeight:     block.Header.BlockHeight,
			ChainId:         block.Header.GetChainId(),
			Topic:           t.Topic,
			TxId:            t.TxId,
			ContractName:    t.ContractName,
			ContractVersion: t.ContractVersion,
			EventData:       t.EventData,
		}
		eventsInfo = append(eventsInfo, eventInfo)
	}
	pb.msgBus.Publish(msgbus.ContractEventInfo, &commonpb.ContractEventInfoList{ContractEvents: eventsInfo})
	return utils.CurrentTimeMillisSeconds() - startPublishContractEventTick
}

// NotifyMessage Notify other subscription modules of chain configuration and certificate management events
func (pb *ReplaceBlock) NotifyMessage(block *commonpb.Block, events []*commonpb.ContractEvent) (err error) {
	if block == nil || len(block.GetTxs()) == 0 {
		return nil
	}

	if native, _ := utils.IsNativeTx(block.Txs[0]); !native {
		return nil
	}

	for _, event := range events { // one by one
		data := event.EventData
		if len(data) == 0 {
			continue
		}
		topicEnum, err := strconv.Atoi(event.Topic)
		if err != nil {
			continue
		}
		topic := msgbus.Topic(topicEnum)
		pb.msgBus.PublishSync(topic, data) // data is a []string, hexToString(proto.Marshal(data))
	}
	return nil
}
