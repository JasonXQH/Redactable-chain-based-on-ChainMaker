/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package snapshot

import (
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"chainmaker.org/chainmaker/utils/v2"

	"github.com/golang/mock/gomock"
)

var snapshotMgr = &ManagerEvidence{
	snapshots: make(map[utils.BlockFingerPrint]*SnapshotEvidence, 1024),
	delegate: &ManagerDelegate{
		blockchainStore: nil,
	},
	log: &test.GoLogger{},
}

func TestNewSnapshot(t *testing.T) {
	t.Log("TestNewSnapshot")
	snapshotList, _ := createNewBlockGroup(t)
	for _, snapshot := range snapshotList {
		t.Logf("%v\n", snapshot)
	}
}

func TestNotifyBlockCommitted(t *testing.T) {
	t.Log("TestNotifyBlockCommitted")

	_, blockList := createNewBlockGroup(t)

	for _, block := range blockList {
		err := snapshotMgr.NotifyBlockCommitted(block)

		if err != nil {
			t.Error(err)
		}
	}
}

func createNewBlockGroup(t *testing.T) ([]protocol.Snapshot, []*common.Block) {

	//snapshotMgr.delegate.
	blockchainStore := mock.NewMockBlockchainStore(gomock.NewController(t))
	blockchainStore.EXPECT().GetLastChainConfig().Return(&configPb.ChainConfig{}, nil).AnyTimes()
	snapshotMgr.delegate.blockchainStore = blockchainStore

	genesis := createNewBlock(0, 0)

	block1 := createNewBlock(1, 1)
	snapshot1 := snapshotMgr.NewSnapshot(genesis, block1)

	block2 := createNewBlock(2, 2)
	snapshot2 := snapshotMgr.NewSnapshot(block1, block2)

	block3 := createNewBlock(3, 3)
	snapshot3 := snapshotMgr.NewSnapshot(block2, block3)

	block3a := createNewBlock(3, 4)
	snapshot3a := snapshotMgr.NewSnapshot(block2, block3a)
	return []protocol.Snapshot{snapshot1, snapshot2, snapshot3, snapshot3a}, []*common.Block{block1, block2, block3, block3a}
}
