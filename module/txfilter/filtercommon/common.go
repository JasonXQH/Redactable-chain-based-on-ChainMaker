/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filtercommon transaction filter common tools
package filtercommon

import (
	"errors"
	"fmt"
	"time"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/common"

	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
)

// ChaseBlockHeight Chase high block
func ChaseBlockHeight(store protocol.BlockchainStore, filter protocol.TxFilter, log protocol.Logger) error {
	cost := time.Now()
	// get last block
	lastBlock, err := store.GetLastBlock()
	if err != nil {
		log.Errorf("query last block from db fail, error: %v", err)
		return err
	}
	log.Infof("chase block start,filter height: %v, block height: %v", filter.GetHeight(),
		lastBlock.Header.BlockHeight)
	// Loop query filter block height to the last block height added to the transaction filter
	for height := filter.GetHeight() + 1; height <= lastBlock.Header.BlockHeight; height++ {
		var block *common.Block
		if height != lastBlock.Header.BlockHeight {
			// Query a block if the current height is not equal to the last block height
			block, err = store.GetBlock(height)
			if err != nil {
				log.Errorf("query block from db fail, height: %v, error: %v", height, err)
				return err
			}
		} else {
			// The last block is assigned if the current height is equal to the last block height
			block = lastBlock
		}
		ids := utils.GetTxIds(block.Txs)
		// Add to the transaction filter
		err = filter.AddsAndSetHeight(ids, block.Header.BlockHeight)
		if err != nil {
			log.Errorf("chase block add fail, height: %v, keys: %v, error: %v", block.Header.BlockHeight, len(ids), err)
			return err
		}
		log.Infof("chasing block, height: %d", block.Header.BlockHeight)
	}
	log.Infof("chase block finish, height: %d, block height: %d, cost: %d", filter.GetHeight(),
		lastBlock.Header.BlockHeight, time.Since(cost))

	return nil
}

// GetConf Get the configuration of the transaction filter
func GetConf(chainId string) (*TxFilterConfig, error) {
	return ToPbConfig(localconf.ChainMakerConfig.TxFilter, chainId)
}

// ToPbConfig Convert localconf.TxFilterConfig to config.TxFilterConfig
func ToPbConfig(conf localconf.TxFilterConfig, chainId string) (*TxFilterConfig, error) {
	c := &TxFilterConfig{
		Type: TxFilterType(conf.Type),
	}
	switch TxFilterType(conf.Type) {
	case TxFilterTypeDefault:
		// Returns if the default transaction filter is specified in the configuration file
		return c, nil
	case TxFilterTypeBirdsNest:
		// Returns the Bird's Nest transaction filter if specified in the configuration file
		// Check the Bird's Nest configuration
		err := CheckBNConfig(conf.BirdsNest, true)
		if err != StringNil {
			return nil, errors.New(err)
		}
		cuckoo := &bn.CuckooConfig{
			KeyType:       bn.KeyType(conf.BirdsNest.Cuckoo.KeyType),
			TagsPerBucket: conf.BirdsNest.Cuckoo.TagsPerBucket,
			BitsPerItem:   conf.BirdsNest.Cuckoo.BitsPerItem,
			MaxNumKeys:    conf.BirdsNest.Cuckoo.MaxNumKeys,
			TableType:     conf.BirdsNest.Cuckoo.TableType,
		}
		rules := &bn.RulesConfig{
			AbsoluteExpireTime: conf.BirdsNest.Rules.AbsoluteExpireTime,
		}
		snapshot := &bn.SnapshotSerializerConfig{
			Type: bn.SerializeIntervalType(conf.BirdsNest.Snapshot.Type),
			Timed: &bn.TimedSerializeIntervalConfig{
				Interval: int64(conf.BirdsNest.Snapshot.Timed.Interval),
			},
			BlockHeight: &bn.BlockHeightSerializeIntervalConfig{
				Interval: uint64(conf.BirdsNest.Snapshot.BlockHeight.Interval),
			},
			Path: conf.BirdsNest.Snapshot.Path,
		}
		c.BirdsNest = &bn.BirdsNestConfig{
			Length:   conf.BirdsNest.Length,
			ChainId:  chainId,
			Rules:    rules,
			Cuckoo:   cuckoo,
			Snapshot: snapshot,
		}
		return c, nil
	case TxFilterTypeMap:
		// Returns the map transaction filter if specified in the configuration file
		return c, nil
	case TxFilterTypeShardingBirdsNest:
		// Returns the Sharding Bird's Nest transaction filter if specified in the configuration file
		// Check the Bird's Nest configuration
		err := CheckShardingBNConfig(conf.ShardingBirdsNest)
		if err != StringNil {
			return nil, errors.New(err)
		}
		snapshot := &bn.SnapshotSerializerConfig{
			Type: bn.SerializeIntervalType(conf.BirdsNest.Snapshot.Type),
			Timed: &bn.TimedSerializeIntervalConfig{
				Interval: int64(conf.ShardingBirdsNest.Snapshot.Timed.Interval),
			},
			BlockHeight: &bn.BlockHeightSerializeIntervalConfig{
				Interval: uint64(conf.ShardingBirdsNest.Snapshot.BlockHeight.Interval),
			},
			Path: conf.ShardingBirdsNest.Snapshot.Path,
		}
		rules := &bn.RulesConfig{
			AbsoluteExpireTime: conf.ShardingBirdsNest.BirdsNest.Rules.AbsoluteExpireTime,
		}
		cuckoo := &bn.CuckooConfig{
			KeyType:       bn.KeyType(conf.ShardingBirdsNest.BirdsNest.Cuckoo.KeyType),
			TagsPerBucket: conf.ShardingBirdsNest.BirdsNest.Cuckoo.TagsPerBucket,
			BitsPerItem:   conf.ShardingBirdsNest.BirdsNest.Cuckoo.BitsPerItem,
			MaxNumKeys:    conf.ShardingBirdsNest.BirdsNest.Cuckoo.MaxNumKeys,
			TableType:     conf.ShardingBirdsNest.BirdsNest.Cuckoo.TableType,
		}
		bidsNest := &bn.BirdsNestConfig{
			Length:   conf.ShardingBirdsNest.BirdsNest.Length,
			ChainId:  chainId,
			Rules:    rules,
			Cuckoo:   cuckoo,
			Snapshot: snapshot,
		}
		c.ShardingBirdsNest = &sbn.ShardingBirdsNestConfig{
			ChainId:   chainId,
			Length:    conf.ShardingBirdsNest.Length,
			Timeout:   conf.ShardingBirdsNest.Timeout,
			Birdsnest: bidsNest,
			Snapshot:  snapshot,
		}
		return c, nil
	default:
		return c, nil
	}
}

// TxFilterLogger protocol.Logger wrapper
type TxFilterLogger struct {
	// log
	log protocol.Logger
}

// Debugf debug sprintf
func (t TxFilterLogger) Debugf(format string, args ...interface{}) {
	t.log.DebugDynamic(LoggingFixLengthFunc(format, args...))
}

// Errorf error sprintf
func (t TxFilterLogger) Errorf(format string, args ...interface{}) {
	t.log.Errorf(format, args...)
}

// Infof info sprintf
func (t TxFilterLogger) Infof(format string, args ...interface{}) {
	t.log.Infof(format, args...)
}

// Warnf warning sprintf
func (t TxFilterLogger) Warnf(format string, args ...interface{}) {
	t.log.Warnf(format, args...)
}

// NewLogger new logger by protocol.Logger
func NewLogger(log protocol.Logger) *TxFilterLogger {
	return &TxFilterLogger{log: log}
}

// LoggingFixLengthFunc The string length is fixed to avoid outputting a large number of logs
func LoggingFixLengthFunc(format string, args ...interface{}) func() string {
	return func() string {
		return LoggingFixLength(format, args...)
	}
}

// LoggingFixLength The string length is fixed to avoid outputting a large number of logs
func LoggingFixLength(format string, args ...interface{}) string {
	str := fmt.Sprintf(format, args...)
	if len(str) > 1024 {
		str = str[:1024] + "..."
	}
	return str
}
