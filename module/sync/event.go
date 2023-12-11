/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	syncPb "chainmaker.org/chainmaker/pb-go/v2/sync"
	"github.com/Workiva/go-datastructures/queue"
)

const (
	priorityTop    = 3
	priorityMiddle = 2
	priorityLow    = 1
	//priorityBase   = 0
)

//Priority can return the priority level of one implementation
type Priority interface {
	Level() int
}

//Compare sort the items put into the queue in ascending order
//in routine, smaller level will be processed first
func Compare(parent, other queue.Item) int {
	//doProcessBlockTk 	-> ... priority first
	//doScheduleTk 		-> ... priority second
	//doNodeStatusTk 	-> ... priority third
	var (
		ok   bool
		p, o Priority
	)
	if p, ok = parent.(Priority); !ok {
		return 0
	}
	if o, ok = other.(Priority); !ok {
		return 0
	}
	if p.Level() < o.Level() {
		return 1
	} else if p.Level() == o.Level() {
		return 0
	}
	return -1
}

//SyncedBlockMsg indicates that a synchronized block response was received
type SyncedBlockMsg struct {
	msg  []byte
	from string
}

//Level get SyncedBlockMsg priority level
func (m *SyncedBlockMsg) Level() int {
	return priorityMiddle
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *SyncedBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//NodeStatusMsg indicates that received status information from other nodes
type NodeStatusMsg struct {
	msg  syncPb.BlockHeightBCM
	from string
}

//Level get priority level
func (m *NodeStatusMsg) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *NodeStatusMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//SchedulerMsg notify scheduler to send a synchronous block request
type SchedulerMsg struct{}

//Level get priority level
func (m *SchedulerMsg) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *SchedulerMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//LivenessMsg to check whether the response of synchronous block request is timed out if received this message
type LivenessMsg struct{}

//Level get priority level
func (m *LivenessMsg) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *LivenessMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//ReceivedBlockInfos notify processor that it has received some blocks that needs to be processed
type ReceivedBlockInfos struct {
	*syncPb.SyncBlockBatch
	from string
}

//Level get priority level
func (m *ReceivedBlockInfos) Level() int {
	return priorityTop
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *ReceivedBlockInfos) Compare(other queue.Item) int {
	return Compare(m, other)
}

//ProcessBlockMsg notify processor to do a block processing operation
type ProcessBlockMsg struct{}

//Level get priority level
func (m *ProcessBlockMsg) Level() int {
	return priorityTop
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *ProcessBlockMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//DataDetection notify to do a data check
type DataDetection struct{}

//Level get priority level
func (m *DataDetection) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *DataDetection) Compare(other queue.Item) int {
	return Compare(m, other)
}

type processedBlockStatus int64

const (
	ok processedBlockStatus = iota
	dbErr
	addErr
	hasProcessed
	validateFailed
)

//ProcessedBlockResp the result of the block being processed by the processor
type ProcessedBlockResp struct {
	height uint64
	status processedBlockStatus
	from   string
}

//Level get priority level
func (m *ProcessedBlockResp) Level() int {
	return priorityTop
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *ProcessedBlockResp) Compare(other queue.Item) int {
	return Compare(m, other)
}

//StopSyncMsg notify sync server to stop syncing block data from other nodes
type StopSyncMsg struct {
}

//Level get priority level
func (m *StopSyncMsg) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *StopSyncMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}

//StartSyncMsg notify sync server to resume to block data from other nodes
type StartSyncMsg struct {
}

//Level get priority level
func (m *StartSyncMsg) Level() int {
	return priorityLow
}

//Compare invoke by queue.PriorityQueue to sort queue.Item
func (m *StartSyncMsg) Compare(other queue.Item) int {
	return Compare(m, other)
}
