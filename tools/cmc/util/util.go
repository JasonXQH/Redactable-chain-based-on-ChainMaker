// Copyright (C) BABEC. All rights reserved.
// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0s

package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/types"
	"chainmaker.org/chainmaker/common/v2/evmutils"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	prettyjson "github.com/hokaccha/go-prettyjson"
)

// MaxInt returns the greater one between two integer
func MaxInt(i, j int) int {
	if j > i {
		return j
	}
	return i
}

// ConvertParameters convert params map to []*common.KeyValuePair
func ConvertParameters(pars map[string]interface{}) []*common.KeyValuePair {
	var kvp []*common.KeyValuePair
	for k, v := range pars {
		var value string
		switch v := v.(type) {
		case string:
			value = v
		default:
			bz, err := json.Marshal(v)
			if err != nil {
				return nil
			}
			value = string(bz)
		}
		kvp = append(kvp, &common.KeyValuePair{
			Key:   k,
			Value: []byte(value),
		})
	}
	return kvp
}

// CalcEvmContractName calc contract name of EVM kind
func CalcEvmContractName(contractName string) string {
	return hex.EncodeToString(evmutils.Keccak256([]byte(contractName)))[24:]
}

func RespResultToString(resp *common.TxResponse) interface{} {
	if resp == nil || resp.ContractResult == nil {
		return resp
	}
	return types.TxResponse{
		TxResponse: resp,
		ContractResult: &types.ContractResult{
			ContractResult: resp.ContractResult,
			Result:         string(resp.ContractResult.Result),
		},
	}
}

// PrintPrettyJson print pretty json of data
func PrintPrettyJson(data interface{}) {
	output, err := prettyjson.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
}
