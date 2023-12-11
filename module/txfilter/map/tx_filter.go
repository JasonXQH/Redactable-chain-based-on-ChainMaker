/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package mapimpl transaction filter implementation
package mapimpl

import (
	"sync"

	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	"chainmaker.org/chainmaker/pb-go/v2/txfilter"
)

// TxFilter sync.Map transaction filter
type TxFilter struct {
	// blcok height
	height uint64
	// m Thread-safe map
	m sync.Map
}

// ValidateRule validate rules
func (f *TxFilter) ValidateRule(_ string, _ ...bn.RuleType) error {
	return nil
}

// New transaction filter init
func New() *TxFilter {
	return &TxFilter{m: sync.Map{}}
}

// GetHeight get height from transaction filter
func (f *TxFilter) GetHeight() uint64 {
	return f.height
}

// SetHeight set height from transaction filter
func (f *TxFilter) SetHeight(height uint64) {
	f.height = height
}

// IsExistsAndReturnHeight is exists and return height
func (f *TxFilter) IsExistsAndReturnHeight(txId string, _ ...bn.RuleType) (bool, uint64, *txfilter.Stat, error) {
	exists, stat, err := f.IsExists(txId)
	if err != nil {
		return false, 0, nil, err
	}
	return exists, f.height, stat, nil
}

// Add txId to transaction filter
func (f *TxFilter) Add(txId string) error {
	f.m.Store(txId, struct{}{})
	return nil
}

// Adds batch Add txId
func (f *TxFilter) Adds(txIds []string) error {
	for _, txId := range txIds {
		go f.m.Store(txId, struct{}{})
	}
	return nil
}

// AddsAndSetHeight batch add tx id and set height
func (f *TxFilter) AddsAndSetHeight(txId []string, height uint64) error {
	err := f.Adds(txId)
	if err != nil {
		return err
	}
	f.SetHeight(height)
	return nil
}

// IsExists Check whether TxId exists in the transaction filter
func (f *TxFilter) IsExists(txId string, _ ...bn.RuleType) (bool, *txfilter.Stat, error) {
	_, ok := f.m.Load(txId)
	return ok, nil, nil
}

// Close transaction filter
func (f *TxFilter) Close() {
}
