/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sync

import (
	"fmt"
	"time"
)

// BlockSyncServerConf sync service configurable options
type BlockSyncServerConf struct {
	timeOut time.Duration // Timeout of request, unit nanosecond
	// When the difference between the height of the node and the latest height of peers is 1,
	//the time interval for requesting
	reqTimeThreshold time.Duration
	// The ticker to process of the block, unit nanosecond
	processBlockTick time.Duration
	// The ticker to liveness checking, unit nanosecond
	livenessTick time.Duration
	// The ticker to request block from the peer, unit nanosecond
	schedulerTick time.Duration
	// The ticker to request node status from other peers, unit nanosecond
	nodeStatusTick time.Duration
	// The ticker to check data in processor
	dataDetectionTick time.Duration
	// Block incoming request ignore duration
	blockRequestTime time.Duration
	// Maximum number of blocks to be processed in scheduler
	blockPoolSize uint64
	// The number of blocks received from each node in a request
	batchSizeFromOneNode uint64
	//The minimum difference between the local block height and the maximum height
	// of other peers, if reachable, maybe should stop syncing blocks from peers
	minLagThreshold uint64
	// the config to receive more node status
	minLagThresholdTime time.Duration
}

// NewBlockSyncServerConf create a new BlockSyncServerConf instance with default values
func NewBlockSyncServerConf() *BlockSyncServerConf {
	return &BlockSyncServerConf{
		timeOut:              30 * time.Second,
		blockPoolSize:        bufferSize,
		batchSizeFromOneNode: 1,
		processBlockTick:     20 * time.Millisecond,
		livenessTick:         1 * time.Second,
		nodeStatusTick:       2 * time.Second,
		schedulerTick:        20 * time.Millisecond,
		dataDetectionTick:    time.Minute,
		reqTimeThreshold:     1 * time.Second,
		blockRequestTime:     5 * time.Second,
		minLagThreshold:      5,
		minLagThresholdTime:  3 * time.Second,
	}
}

// SetBlockPoolSize set block pool size
// this value will affect maximum number of cache blocks sync server has received and waiting to be processed
func (c *BlockSyncServerConf) SetBlockPoolSize(n uint64) *BlockSyncServerConf {
	c.blockPoolSize = n
	return c
}

// SetWaitTimeOfBlockRequestMsg set sync block request timeout
func (c *BlockSyncServerConf) SetWaitTimeOfBlockRequestMsg(n int64) *BlockSyncServerConf {
	c.timeOut = time.Duration(n) * time.Second
	return c
}

// SetBatchSizeFromOneNode set the number of blocks that can be fetched in one request
func (c *BlockSyncServerConf) SetBatchSizeFromOneNode(n uint64) *BlockSyncServerConf {
	c.batchSizeFromOneNode = n
	return c
}

// SetProcessBlockTicker set time interval for processing blocks
func (c *BlockSyncServerConf) SetProcessBlockTicker(n float64) *BlockSyncServerConf {
	c.processBlockTick = time.Duration(n * float64(time.Millisecond))
	return c
}

// SetSchedulerTicker set time interval for scheduling block request
func (c *BlockSyncServerConf) SetSchedulerTicker(n float64) *BlockSyncServerConf {
	c.schedulerTick = time.Duration(n * float64(time.Millisecond))
	return c
}

// SetLivenessTicker set time interval for doing a liveness check
func (c *BlockSyncServerConf) SetLivenessTicker(n float64) *BlockSyncServerConf {
	c.livenessTick = time.Duration(n * float64(time.Second))
	return c
}

// SetNodeStatusTicker set time interval for broadcasting to get other peers status
func (c *BlockSyncServerConf) SetNodeStatusTicker(n float64) *BlockSyncServerConf {
	c.nodeStatusTick = time.Duration(n * float64(time.Second))
	return c
}

// SetDataDetectionTicker set time interval for checking data in processor
func (c *BlockSyncServerConf) SetDataDetectionTicker(n float64) *BlockSyncServerConf {
	c.dataDetectionTick = time.Duration(n * float64(time.Second))
	return c
}

// SetReqTimeThreshold set request time limit
// if the difference between own block height and the highest block height is 1
// the time difference between two requests must be greater than reqTimeThreshold
func (c *BlockSyncServerConf) SetReqTimeThreshold(n float64) *BlockSyncServerConf {
	c.reqTimeThreshold = time.Duration(n * float64(time.Second))
	return c
}

// SetBlockRequestTime set expiration time of the received request
// node will ignore the request sent multiple times by others for the same block within this time
func (c *BlockSyncServerConf) SetBlockRequestTime(n float64) *BlockSyncServerConf {
	c.blockRequestTime = time.Duration(n * float64(time.Second))
	return c
}

// SetMinLagThreshold set the the minimum value local block height lags behind maximum block height
// If this value is reached, notify anyone who cares about this event
func (c *BlockSyncServerConf) SetMinLagThreshold(n uint64) *BlockSyncServerConf {
	c.minLagThreshold = n
	return c
}

// SetMinLagThresholdTime set time threshold of minimum lags
func (c *BlockSyncServerConf) SetMinLagThresholdTime(n float64) *BlockSyncServerConf {
	c.minLagThresholdTime = time.Duration(n * float64(time.Second))
	return c
}
func (c *BlockSyncServerConf) print() string {
	return fmt.Sprintf("blockPoolSize: %d, request timeout: %d, batchSizeFromOneNode: %d"+
		", processBlockTick: %v, schedulerTick: %v, livenessTick: %v, nodeStatusTick: %v\n",
		c.blockPoolSize, c.timeOut, c.batchSizeFromOneNode, c.processBlockTick, c.schedulerTick, c.livenessTick,
		c.nodeStatusTick)
}
