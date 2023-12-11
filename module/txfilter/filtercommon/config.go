/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filtercommon transaction filter common config
package filtercommon

import (
	bn "chainmaker.org/chainmaker/common/v2/birdsnest"
	sbn "chainmaker.org/chainmaker/common/v2/shardingbirdsnest"
)

// TxFilterType Transaction filter type
type TxFilterType int32

const (
	// TxFilterTypeDefault Default transaction filter type
	TxFilterTypeDefault TxFilterType = 0
	// TxFilterTypeBirdsNest Bird's Nest transaction filter type
	TxFilterTypeBirdsNest TxFilterType = 1
	// TxFilterTypeMap Map transaction filter type
	TxFilterTypeMap TxFilterType = 2
	// TxFilterTypeShardingBirdsNest Sharding Bird's Nest transaction filter type
	TxFilterTypeShardingBirdsNest TxFilterType = 3
)

// TxFilterConfig transaction filter config
type TxFilterConfig struct {
	// Transaction filter type
	Type TxFilterType `json:"type,omitempty"`
	// Bird's nest configuration
	BirdsNest *bn.BirdsNestConfig `json:"birds_nest,omitempty"`
	// Sharding bird's nest configuration
	ShardingBirdsNest *sbn.ShardingBirdsNestConfig `json:"sharding_birds_nest,omitempty"`
}
