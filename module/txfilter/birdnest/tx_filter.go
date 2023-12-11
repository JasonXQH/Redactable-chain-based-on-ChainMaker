/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package birdnest transaction filter implementation
package birdnest

import (
	"fmt"
	"sync"
	"time"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/txfilter"
	"chainmaker.org/chainmaker/protocol/v2"
)

// TxFilter bn.BirdsNestImpl transaction filter
type TxFilter struct {
	// log Log output protocol.Logger
	log protocol.Logger
	// bn Bird's Nest implementation
	bn *bn.BirdsNestImpl
	// store block store protocol.BlockchainStore
	store protocol.BlockchainStore
	// exitC Exit channel
	exitC chan struct{}
	// l read write lock
	l sync.RWMutex
}

// ValidateRule validate rules
func (f *TxFilter) ValidateRule(txId string, ruleType ...bn.RuleType) error {
	// Convert the transaction ID to TimestampKey
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	err = f.bn.ValidateRule(key, ruleType...)
	if err != nil {
		return err
	}
	return err
}

// New transaction filter init
func New(config *bn.BirdsNestConfig, log protocol.Logger, store protocol.BlockchainStore) (
	protocol.TxFilter, error) {
	// Because it is compatible with Normal type, the transaction ID cannot be converted to time transaction ID, so the
	// database can be queried directly. Therefore, the transaction ID type is fixed as TimestampKey
	config.Cuckoo.KeyType = bn.KeyType_KTTimestampKey

	initLasts := time.Now()
	exitC := make(chan struct{})
	// New bird's nest
	birdsNest, err := bn.NewBirdsNest(config, exitC, bn.LruStrategy, filtercommon.NewLogger(log))
	if err != nil {
		if err != bn.ErrCannotModifyTheNestConfiguration {
			log.Errorf("new filter fail, error: %v", err)
			return nil, err
		}
		log.Warnf("new filter, %v", err)
	}
	txFilter := &TxFilter{
		log:   log,
		bn:    birdsNest,
		exitC: exitC,
	}
	// chase block height
	err = filtercommon.ChaseBlockHeight(store, txFilter, log)
	if err != nil {
		return nil, err
	}
	log.Infof("bird's nest filter init success, size: %v, max keys: %v, cost: %v",
		config.Length, config.Cuckoo.MaxNumKeys, time.Since(initLasts))
	birdsNest.Start()
	return txFilter, nil
}

// GetHeight get height from transaction filter
func (f *TxFilter) GetHeight() uint64 {
	return f.bn.GetHeight()
}

// SetHeight set height from transaction filter
func (f *TxFilter) SetHeight(height uint64) {
	f.bn.SetHeight(height)
}

// IsExistsAndReturnHeight is exists and return height
func (f *TxFilter) IsExistsAndReturnHeight(txId string, ruleType ...bn.RuleType) (bool, uint64, *txfilter.Stat, error) {
	exists, stat, err := f.IsExists(txId, ruleType...)
	if err != nil {
		return false, 0, stat, err
	}
	return exists, f.GetHeight(), stat, nil
}

// Add txId to transaction filter
func (f *TxFilter) Add(txId string) error {
	// Convert the transaction ID to TimestampKey
	timestampKey, err := bn.ToTimestampKey(txId)
	if err != nil {
		return nil
	}
	f.l.Lock()
	defer f.l.Unlock()
	return f.bn.Add(timestampKey)
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	start := time.Now()
	// Convert the transaction ID to TimestampKey
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) > 0 {
		f.l.Lock()
		err := f.bn.Adds(timestampKeys)
		f.l.Unlock()
		if err != nil {
			f.log.Warnf("filter adds fail, txid size: %v, error: %v", len(txIds), err)
		}
		f.addsPrintInfo(txIds, start)
		return nil
	}
	f.log.Warnf("no time-type transaction")
	return nil
}

// addsPrintInfo Output logs after adding transactions
// index 1 cuckoo size
// index 2 current index
// index 3 total cuckoo size
// index 4 total space occupied by cuckoo
func (f *TxFilter) addsPrintInfo(txIds []string, start time.Time) {
	info := f.bn.Info()
	f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc(
		"filter adds success, height: %v, txids: %v, size: %v, curr: %v, total keys: %v, bytes: %v, cost: %v",
		f.GetHeight(), len(txIds), info[1], info[2], info[3], info[4],
		time.Since(start),
	))
}

// AddsAndSetHeight batch add tx id and set height
func (f *TxFilter) AddsAndSetHeight(txIds []string, height uint64) error {
	start := time.Now()
	// Convert the transaction ID to TimestampKey
	timestampKeys, _ := bn.ToTimestampKeysAndNormalKeys(txIds)
	if len(timestampKeys) <= 0 {
		// Update block height if there is no time to transaction ID
		f.SetHeight(height)
		f.log.DebugDynamic(filtercommon.LoggingFixLengthFunc("adds and set height, no timestamp keys height: %d",
			height))
		return nil
	}
	f.l.Lock()
	// Add the time transaction ID into the Bird's Nest transaction filter and update the height
	err := f.bn.AddsAndSetHeight(timestampKeys, height)
	f.l.Unlock()
	if err != nil {
		return err
	}
	f.addsPrintInfo(txIds, start)
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f *TxFilter) IsExists(txId string, ruleType ...bn.RuleType) (exists bool, stat *txfilter.Stat, err error) {
	var costs time.Duration
	// Convert the transaction ID to TimestampKey
	key, err := bn.ToTimestampKey(txId)
	if err != nil {
		// If the transaction ID is not a time type, query whether the database exists
		exists, costs, err = f.findDb(txId)
		if err != nil {
			err = fmt.Errorf("%v, txid type: normal", err)
		}
		return exists, filtercommon.NewStat1(0, costs), err
	}
	f.l.RLock()
	defer f.l.RUnlock()
	// If the transaction ID is of the time type, the transaction filter exists
	start := time.Now()
	contains, err := f.bn.Contains(key, ruleType...)
	filterCosts := time.Since(start)
	if err != nil {
		// If not, query DB
		if err == bn.ErrKeyTimeIsNotInTheFilterRange {
			exists, costs, err = f.findDb(txId)
			if err != nil {
				err = fmt.Errorf("%v, key time is not in the filter range", err)
			}
			return exists, filtercommon.NewStat1(filterCosts, costs), err
		}
	}
	if contains {
		// False positive treatment
		// If not, query DB
		if err == bn.ErrKeyTimeIsNotInTheFilterRange {
			exists, costs, err = f.findDb(txId)
			if err != nil {
				err = fmt.Errorf("%v, %v positive", err, exists)
			}
			return exists, filtercommon.NewStat1(filterCosts, costs), err
		}
	}
	// True positive
	return contains, filtercommon.NewStat0(filterCosts, costs), nil
}

// Close transaction filter
func (f *TxFilter) Close() {
	close(f.exitC)
}

func (f *TxFilter) findDb(txId string) (bool, time.Duration, error) {
	start := time.Now()
	exists, err := f.store.TxExists(txId)
	costs := time.Since(start)
	if err != nil {
		f.log.Errorf("[%v] filter check exists, query from db fail, error:%v", txId, err)
		return false, costs, err
	}
	return exists, costs, err
}
