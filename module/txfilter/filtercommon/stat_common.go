/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filtercommon stat common
package filtercommon

import (
	"time"

	"chainmaker.org/chainmaker/pb-go/v2/txfilter"
)

// NewStat txfilter stat
func NewStat(fpCount uint32, filterCosts, dbCosts time.Duration) *txfilter.Stat {
	return &txfilter.Stat{
		FpCount:     fpCount,
		FilterCosts: filterCosts.Milliseconds(),
		DbCosts:     dbCosts.Milliseconds(),
	}
}

// NewStat0 txfilter stat FP count 0
func NewStat0(filterCosts, dbCosts time.Duration) *txfilter.Stat {
	return NewStat(0, filterCosts, dbCosts)
}

// NewStat1 txfilter stat FP count 1
func NewStat1(filterCosts, dbCosts time.Duration) *txfilter.Stat {
	return NewStat(1, filterCosts, dbCosts)
}
