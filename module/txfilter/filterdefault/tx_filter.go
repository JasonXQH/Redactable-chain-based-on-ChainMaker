/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filterdefault transaction filter implementation
package filterdefault

import (
	"time"

	"chainmaker.org/chainmaker-go/module/txfilter/filtercommon"
	"chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/txfilter"
	"chainmaker.org/chainmaker/protocol/v2"
)

// TxFilter protocol.BlockchainStore transaction filter
type TxFilter struct {
	// store block store
	store protocol.BlockchainStore
}

// ValidateRule validate transaction rules
func (f TxFilter) ValidateRule(_ string, _ ...birdsnest.RuleType) error {
	return nil
}

// New transaction filter init
func New(store protocol.BlockchainStore) *TxFilter {
	return &TxFilter{store: store}
}

// GetHeight get height from transaction filter
func (f TxFilter) GetHeight() uint64 {
	block, err := f.store.GetLastBlock()
	if err != nil {
		return 0
	}
	return block.Header.BlockHeight
}

// SetHeight set height from transaction filter
func (f TxFilter) SetHeight(_ uint64) {
}

// IsExistsAndReturnHeight is exists and return height
func (f TxFilter) IsExistsAndReturnHeight(txId string, _ ...birdsnest.RuleType) (bool, uint64, *txfilter.Stat, error) {
	start := time.Now()
	exists, height, err := f.store.TxExistsInFullDB(txId)
	costs := time.Since(start)
	return exists, height, filtercommon.NewStat1(0, costs), err
}

// Add txId to transaction filter
func (f TxFilter) Add(_ string) error {
	return nil
}

// Adds batch Add txId
func (f TxFilter) Adds(_ []string) error {
	return nil
}

// AddsAndSetHeight batch add tx id and set height
func (f TxFilter) AddsAndSetHeight(_ []string, _ uint64) error {
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f TxFilter) IsExists(txId string, _ ...birdsnest.RuleType) (bool, *txfilter.Stat, error) {
	start := time.Now()
	exists, err := f.store.TxExists(txId)
	costs := time.Since(start)
	return exists, filtercommon.NewStat1(0, costs), err
}

// Close transaction filter
func (f TxFilter) Close() {
}
