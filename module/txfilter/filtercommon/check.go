/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package filtercommon transaction filter parameter check
package filtercommon

import (
	"fmt"
	"strconv"

	"chainmaker.org/chainmaker/common/v2/birdsnest"

	"chainmaker.org/chainmaker/localconf/v2"
	"github.com/linvon/cuckoo-filter"
)

var (
	// StringNil If the error string is not equal to StringNil then you need to returnup
	StringNil = ""
	// ErrStrSharding Error string prefix "sharding."
	ErrStrSharding = "sharding."
	// ErrStrBirdsNest Error string prefix "brids_nest."
	ErrStrBirdsNest = "brids_nest."
	// ErrStrCuckoo Error string prefix "cuckoo."
	ErrStrCuckoo = "cuckoo."
	// ErrStrSnapshot snapshot. Error string prefix snapshot.
	ErrStrSnapshot = "snapshot."

	// ErrStrInvalidShardingTimeoutMustBeGreaterThan1 Invalid sharding timeout must be greater than 1
	ErrStrInvalidShardingTimeoutMustBeGreaterThan1 = ErrStrSharding + "timeout must be greater than 1"
	// ErrStrInvalidShardingLengthRange2_50 Invalid sharding size range: 2_50
	ErrStrInvalidShardingLengthRange2_50 = ErrStrSharding + "length range 2 ~ 50"
	// ErrStrBirdsNestLength bird's nest length must range from 2 to 100
	ErrStrBirdsNestLength = ErrStrBirdsNest + "length must range from 2 to 100"

	//ErrStrInvalidChainIdCannotBeNil = "chain_id cannot be nil"

	// ErrStrSnapshotPathCannotBeNil snapshot path cannot be nil
	ErrStrSnapshotPathCannotBeNil = ErrStrSnapshot + "path cannot be nil"
	// ErrStrSnapshotSerializeIntervalMustBeGreaterThan0 snapshot serialize_interval must be greater than 0
	ErrStrSnapshotSerializeIntervalMustBeGreaterThan0 = ErrStrSnapshot + "serialize_interval must be greater than 0"

	// ErrStrRulesAbsoluteExpireTimeMustBeGreaterThan0 absolute expire time must be greater than 0
	ErrStrRulesAbsoluteExpireTimeMustBeGreaterThan0 = "rules.absolute_expire_time must be greater than 0"

	// ErrStrCuckooKeyTypeNotSupport cuckoo key type not support
	ErrStrCuckooKeyTypeNotSupport = ErrStrCuckoo + "key_type %v not support"
	// ErrStrInvalidCuckooTableTypeNotSupport cuckoo table_type not support
	ErrStrInvalidCuckooTableTypeNotSupport = ErrStrCuckoo + "table_type not support"
	// ErrStrInvalidCuckooBitsPerItemMustBeGreaterThan0 cuckoo bits_per_item must be greater than 0
	ErrStrInvalidCuckooBitsPerItemMustBeGreaterThan0 = ErrStrCuckoo + "bits_per_item must be greater than 0"
	// ErrStrInvalidCuckooMaxNumKeysMustBeGreaterThan0 cuckoo max_num_keys must be greater than 0
	ErrStrInvalidCuckooMaxNumKeysMustBeGreaterThan0 = ErrStrCuckoo + "max_num_keys must be greater than 0"
	// ErrStrInvalidCuckooTagsPerBucketMustBeGreaterThan0 cuckoo tags_per_bucket must be greater than 0
	ErrStrInvalidCuckooTagsPerBucketMustBeGreaterThan0 = ErrStrCuckoo + "tags_per_bucket must be greater than 0"
)

// CheckShardingBNConfig check sharding bird's nest configuration
func CheckShardingBNConfig(c localconf.ShardingBirdsNestConfig) string {
	// range 2 ~ 50
	if c.Length < 2 || c.Length > 50 {
		return ErrStrInvalidShardingLengthRange2_50
	}
	if c.Timeout < 1 {
		return ErrStrInvalidShardingTimeoutMustBeGreaterThan1
	}
	if err := CheckBNConfig(c.BirdsNest, false); err != StringNil {
		return ErrStrSharding + err
	}
	if err := CheckSnapshot(c.Snapshot); err != StringNil {
		return ErrStrSharding + err
	}
	return StringNil
}

// CheckBNConfig check bird's nest configuration
func CheckBNConfig(c localconf.BirdsNestConfig, isSnapshot bool) string {
	if c.Length < 2 || c.Length > 100 {
		return ErrStrBirdsNestLength
	}
	if err := checkRulesConfig(c.Rules); err != StringNil {
		return ErrStrBirdsNest + err
	}
	if isSnapshot {
		if err := CheckSnapshot(c.Snapshot); err != StringNil {
			return ErrStrBirdsNest + err
		}
	}
	if err := checkCuckooConfig(c.Cuckoo); err != StringNil {
		return ErrStrBirdsNest + err
	}
	return StringNil
}

// checkRulesConfig check rules configuration
func checkRulesConfig(c localconf.RulesConfig) string {
	if c.AbsoluteExpireTime < 0 {
		return ErrStrRulesAbsoluteExpireTimeMustBeGreaterThan0
	}
	return StringNil
}

// CheckSnapshot check snapshot configuration
func CheckSnapshot(c localconf.SnapshotSerializerConfig) string {
	if len(c.Path) == 0 {
		return ErrStrSnapshotPathCannotBeNil
	}
	switch c.Type {
	case int(birdsnest.SerializeIntervalType_Height):
		if c.BlockHeight.Interval <= 0 {
			return ErrStrSnapshotSerializeIntervalMustBeGreaterThan0
		}
	case int(birdsnest.SerializeIntervalType_Timed):
		if c.Timed.Interval <= 0 {
			return ErrStrSnapshotSerializeIntervalMustBeGreaterThan0
		}
	default:
		return "serialize interval type" + strconv.Itoa(c.Type) + " not support"
	}
	return StringNil
}

// checkCuckooConfig check cuckoo configuration
func checkCuckooConfig(c localconf.CuckooConfig) string {
	if _, ok := birdsnest.KeyType_name[birdsnest.KeyType(c.KeyType)]; !ok {
		return fmt.Sprintf(ErrStrCuckooKeyTypeNotSupport, c.KeyType)
	}
	if !(c.TableType == cuckoo.TableTypeSingle || c.TableType == cuckoo.TableTypePacked) {
		return ErrStrInvalidCuckooTableTypeNotSupport
	}
	if c.BitsPerItem <= 0 {
		return ErrStrInvalidCuckooBitsPerItemMustBeGreaterThan0
	}
	if c.MaxNumKeys <= 0 {
		return ErrStrInvalidCuckooMaxNumKeysMustBeGreaterThan0
	}
	if c.TagsPerBucket <= 0 {
		return ErrStrInvalidCuckooTagsPerBucketMustBeGreaterThan0
	}
	return StringNil
}
