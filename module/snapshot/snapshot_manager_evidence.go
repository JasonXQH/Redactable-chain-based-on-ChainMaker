/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
)

type ManagerEvidence struct {
	delegate  *ManagerDelegate
	snapshots map[utils.BlockFingerPrint]*SnapshotEvidence
	log       protocol.Logger
}

// NewSnapshot when generating blocks, generate a Snapshot for each block, which is used as read-write set cache
func (m *ManagerEvidence) NewSnapshot(prevBlock *commonPb.Block, block *commonPb.Block) protocol.Snapshot {
	m.delegate.lock.Lock()
	defer m.delegate.lock.Unlock()
	blockHeight := block.Header.BlockHeight
	snapshotImpl := m.delegate.makeSnapshotImpl(block)
	evidenceSnapshot := &SnapshotEvidence{
		delegate: snapshotImpl,
		log:      m.log,
	}

	// 计算前序指纹, 和当前指纹
	prevFingerPrint := utils.CalcBlockFingerPrint(prevBlock)
	fingerPrint := utils.CalcBlockFingerPrint(block)

	// 存储当前指纹的snapshot
	m.snapshots[fingerPrint] = evidenceSnapshot

	// 如果前序指纹对应的snapshot存在, 就建立snapshot的对应关系
	if prevSnapshot, ok := m.snapshots[prevFingerPrint]; ok {
		evidenceSnapshot.SetPreSnapshot(prevSnapshot)
	}

	evidenceSnapshot.SetBlockFingerprint(fingerPrint)

	m.log.Infof(
		"create snapshot at height %d, fingerPrint[%v] -> prevFingerPrint[%v]",
		blockHeight,
		fingerPrint,
		prevFingerPrint,
	)
	return evidenceSnapshot
}

// GetSnapshot return a Snapshot from SnapshotManager for read, don't modify any data.
func (m *ManagerEvidence) GetSnapshot(prevBlock *commonPb.Block, block *commonPb.Block) protocol.Snapshot {
	fingerPrint := utils.CalcBlockFingerPrint(block)
	snapshot, exist := m.snapshots[fingerPrint]
	if !exist {
		return m.NewSnapshot(prevBlock, block)
	}
	return snapshot
}

// NotifyBlockCommitted notify block committed info
func (m *ManagerEvidence) NotifyBlockCommitted(block *commonPb.Block) error {
	m.delegate.lock.Lock()
	defer m.delegate.lock.Unlock()

	m.log.Infof("commit snapshot at height %d", block.Header.BlockHeight)

	// 计算刚落块的区块指纹
	deleteFp := utils.CalcBlockFingerPrint(block)
	// 如果有snapshot对应的前序snapshot的指纹, 等于刚落块的区块指纹
	for _, snapshot := range m.snapshots {
		if snapshot == nil || snapshot.GetPreSnapshot() == nil {
			continue
		}
		prevFp := m.delegate.calcSnapshotFingerPrint(snapshot.GetPreSnapshot().(*SnapshotEvidence).delegate)
		if deleteFp == prevFp {
			snapshot.SetPreSnapshot(nil)
		}
	}

	m.log.Infof("delete snapshot %v at height %d", deleteFp, block.Header.BlockHeight)
	delete(m.snapshots, deleteFp)

	// in case of switch-fork, gc too old snapshot
	for _, snapshot := range m.snapshots {
		if snapshot == nil || snapshot.GetPreSnapshot() == nil {
			continue
		}
		preSnapshot, _ := snapshot.GetPreSnapshot().(*SnapshotEvidence)
		if block.Header.BlockHeight-preSnapshot.GetBlockHeight() > 8 {
			deleteOldFp := m.delegate.calcSnapshotFingerPrint(preSnapshot.delegate)
			delete(m.snapshots, deleteOldFp)
			m.log.Infof("delete snapshot %v at height %d while gc", deleteFp, preSnapshot.GetBlockHeight())
			snapshot.SetPreSnapshot(nil)
		}
	}
	return nil
}

// ClearSnapshot remove block and unused snapshots
func (m *ManagerEvidence) ClearSnapshot(block *commonPb.Block) error {
	m.delegate.lock.Lock()
	defer m.delegate.lock.Unlock()

	m.log.Infof("clear snapshot@%s at height %d", block.Header.ChainId, block.Header.BlockHeight)

	// 计算需要删除的区块指纹
	deleteFp := utils.CalcBlockFingerPrintWithoutTx(block)
	deleteFpEx := calcNotConsensusFingerPrint(block)

	delete(m.snapshots, deleteFp)

	// 删除未共识的区块指纹
	if _, ok := m.snapshots[deleteFpEx]; ok {
		delete(m.snapshots, deleteFpEx)
		m.log.Infof("delete snapshot@%s %v & %v at height %d",
			block.Header.ChainId, deleteFp, deleteFpEx, block.Header.BlockHeight)
	} else {
		m.log.Infof("delete snapshot@%s %v at height %d",
			block.Header.ChainId, deleteFp, block.Header.BlockHeight)
	}

	return nil
}
