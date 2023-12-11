/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"
	"sync"

	"chainmaker.org/chainmaker-go/tools/cmc/types"
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/evmutils/abi"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
)

// Dispatch dispatch tx
func Dispatch(client *sdk.ChainClient, contractName, methodName string, kvs []*common.KeyValuePair,
	abi *abi.ABI, limit *common.Limit) {
	var (
		wgSendReq sync.WaitGroup
	)

	for i := 0; i < concurrency; i++ {
		wgSendReq.Add(1)
		go runInvokeContract(client, contractName, methodName, kvs, &wgSendReq, abi, limit)
	}

	wgSendReq.Wait()
}

// DispatchTimes dispatch tx in times
func DispatchTimes(client *sdk.ChainClient, contractName, method string, kvs []*common.KeyValuePair,
	evmMethod *abi.ABI, limit *common.Limit) {
	var (
		wgSendReq sync.WaitGroup
	)
	times := util.MaxInt(1, sendTimes)
	wgSendReq.Add(times)
	for i := 0; i < times; i++ {
		go runInvokeContractOnce(client, contractName, method, kvs, &wgSendReq, evmMethod, limit)
	}
	wgSendReq.Wait()
}

func runInvokeContract(client *sdk.ChainClient, contractName, methodName string,
	kvs []*common.KeyValuePair, wg *sync.WaitGroup, abi *abi.ABI, limit *common.Limit) {

	defer func() {
		wg.Done()
	}()

	for i := 0; i < totalCntPerGoroutine; i++ {
		invokeContract(client, contractName, methodName, "", kvs, abi, limit)
	}
}

func runInvokeContractOnce(client *sdk.ChainClient, contractName, method string, kvs []*common.KeyValuePair,
	wg *sync.WaitGroup, abi *abi.ABI, limit *common.Limit) {

	defer func() {
		wg.Done()
	}()

	invokeContract(client, contractName, method, "", kvs, abi, limit)
}

func invokeContract(client *sdk.ChainClient, contractName, methodName, txId string,
	kvs []*common.KeyValuePair, abi *abi.ABI, limit *common.Limit) {

	payload := client.CreatePayload(txId, common.TxType_INVOKE_CONTRACT, contractName, methodName, kvs, 0, limit)
	adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(client, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
	if err != nil {
		fmt.Printf("MakeAdminInfo failed, %s", err)
		return
	}
	endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		fmt.Printf("MakeEndorsement failed, %s", err)
		return
	}

	var payer []*common.EndorsementEntry
	if len(payerKeyFilePath) > 0 {
		payer, err = util.MakeEndorsement([]string{payerKeyFilePath}, []string{payerCrtFilePath}, []string{payerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return
		}
	}
	var req *common.TxRequest
	if len(payer) == 0 {
		req, err = client.GenerateTxRequest(payload, endorsers)
	} else {
		req, err = client.GenerateTxRequestWithPayer(payload, endorsers, payer[0])
	}

	if err != nil {
		fmt.Printf("GenerateTxRequest failed, %s", err)
		return
	}
	resp, err := client.SendTxRequest(req, timeout, syncResult)
	if err != nil {
		fmt.Printf("[ERROR] invoke contract failed, %s", err.Error())
		return
	}

	var output interface{}
	if abi != nil && resp.ContractResult != nil && resp.ContractResult.Result != nil {
		unpackedData, err := abi.Unpack(method, resp.ContractResult.Result)
		if err != nil {
			fmt.Println(err)
			return
		}
		output = types.EvmTxResponse{
			TxResponse: resp,
			ContractResult: &types.EvmContractResult{
				ContractResult: resp.ContractResult,
				Result:         fmt.Sprintf("%v", unpackedData),
			},
		}
	} else {
		if respResultToString {
			output = util.RespResultToString(resp)
		} else {
			output = resp
		}
	}
	util.PrintPrettyJson(output)
}
