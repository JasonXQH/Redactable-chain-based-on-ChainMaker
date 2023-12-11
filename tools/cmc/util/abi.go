/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"encoding/json"

	"chainmaker.org/chainmaker/common/v2/evmutils/abi"
)

// Param list
type Param map[string]interface{}

// loadFromJSON string into ABI data
func loadFromJSON(jString string) ([]Param, error) {
	if len(jString) == 0 {
		return nil, nil
	}
	data := []Param{}
	err := json.Unmarshal([]byte(jString), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Pack data into bytes
func Pack(a *abi.ABI, method string, paramsJson string) ([]byte, error) {
	param, err := loadFromJSON(paramsJson)
	if err != nil {
		return nil, err
	}
	args := []interface{}{}
	for _, p := range param {
		for _, v := range p {
			args = append(args, v)
		}
	}
	return a.Pack(method, args...)

}
