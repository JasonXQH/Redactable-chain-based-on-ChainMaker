package common

import (
	"chainmaker.org/chainmaker/localconf/v2"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	"fmt"
)

type ReplaceBlock struct {
	store           protocol.BlockchainStore
	log             protocol.Logger
	snapshotManager protocol.SnapshotManager
	ledgerCache     protocol.LedgerCache
	chainConf       protocol.ChainConf
	txFilter        protocol.TxFilter
}

func (rb *ReplaceBlock) ReplaceBlock(
	block *commonpb.Block,
	rwSetMap map[string]*commonpb.TxRWSet,
	// conEventMap map[string][]*commonpb.ContractEvent
) (
	dbLasts, snapshotLasts, confLasts, otherLasts, pubEventLasts, filterLasts int64, blockInfo *commonpb.BlockInfo,
	err error) {

	//更新数据库中的salt

	//新的区块更新上链

	// record block
	rwSet := utils.RearrangeRWSet(block, rwSetMap)
	// record contract event
	//events := rearrangeContractEvent(block, conEventMap)
	fmt.Println("rwSet: ", rwSet)
	// put block
	startDBTick := utils.CurrentTimeMillisSeconds()
	if err = rb.store.PutBlock(block, rwSet); err != nil {
		// if put db error, then panic
		rb.log.Error(err)
		fmt.Println("err in putblock : ", err)
		panic(err)
	}
	rb.ledgerCache.SetLastCommittedBlock(block)
	dbLasts = utils.CurrentTimeMillisSeconds() - startDBTick

	// TxFilter adds
	filterLasts = utils.CurrentTimeMillisSeconds()
	// The default filter type does not run AddsAndSetHeight
	if localconf.ChainMakerConfig.TxFilter.Type != int32(config.TxFilterType_None) {
		err = rb.txFilter.AddsAndSetHeight(utils.GetTxIds(block.Txs), block.Header.GetBlockHeight())
		if err != nil {
			// if add filter error, then panic
			rb.log.Error(err)
			fmt.Println("err in AddsAndSetHeight : ", err)
			panic(err)
		}
	}
	filterLasts = utils.CurrentTimeMillisSeconds() - filterLasts

	// clear snapshot
	startSnapshotTick := utils.CurrentTimeMillisSeconds()
	if err = rb.snapshotManager.NotifyBlockCommitted(block); err != nil {
		err = fmt.Errorf("notify snapshot error [%d](hash:%x)",
			block.Header.BlockHeight, block.Header.BlockHash)
		rb.log.Error(err)
		return 0, 0, 0, 0, 0, 0, nil, err
	}
	snapshotLasts = utils.CurrentTimeMillisSeconds() - startSnapshotTick
	//// v220_compat Deprecated
	//if block.Header.BlockVersion < blockVersion230 {
	//	// notify chainConf to update config when config block committed
	//	startConfTick := utils.CurrentTimeMillisSeconds()
	//	if err = NotifyChainConf(block, cb.chainConf); err != nil {
	//		return 0, 0, 0, 0, 0, 0, nil, err
	//	}
	//	confLasts = utils.CurrentTimeMillisSeconds() - startConfTick
	//}
	// contract event
	//pubEventLasts = cb.publishContractEvent(block, events)
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
func (rb *ReplaceBlock) publishContractEvent(block *commonpb.Block, events []*commonpb.ContractEvent) int64 {
	if len(events) == 0 {
		return 0
	}

	startPublishContractEventTick := utils.CurrentTimeMillisSeconds()
	rb.log.DebugDynamic(func() string {
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
	//cb.msgBus.Publish(msgbus.ContractEventInfo, &commonpb.ContractEventInfoList{ContractEvents: eventsInfo})
	return utils.CurrentTimeMillisSeconds() - startPublishContractEventTick
}

//
//func rearrangeContractEvent(block *commonpb.Block,
//	conEventMap map[string][]*commonpb.ContractEvent) []*commonpb.ContractEvent {
//	conEvent := make([]*commonpb.ContractEvent, 0, len(block.Txs))
//	if conEventMap == nil {
//		return conEvent
//	}
//	for _, tx := range block.Txs {
//		if event, ok := conEventMap[tx.Payload.TxId]; ok {
//			conEvent = append(conEvent, event...)
//		}
//	}
//	return conEvent
//}
