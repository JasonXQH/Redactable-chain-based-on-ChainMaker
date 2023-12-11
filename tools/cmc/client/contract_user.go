/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"chainmaker.org/chainmaker-go/tools/cmc/types"
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/evmutils/abi"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/spf13/cobra"
)

// CHECK_PROPOSAL_RESPONSE_FAILED_FORMAT define CHECK_PROPOSAL_RESPONSE_FAILED_FORMAT error fmt
const CHECK_PROPOSAL_RESPONSE_FAILED_FORMAT = "checkProposalRequestResp failed, %s"

// SEND_CONTRACT_MANAGE_REQUEST_FAILED_FORMAT define SEND_CONTRACT_MANAGE_REQUEST_FAILED_FORMAT error fmt
const SEND_CONTRACT_MANAGE_REQUEST_FAILED_FORMAT = "SendContractManageRequest failed, %s"

// ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT define ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT error fmt
const ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT = "admin orgId & key & cert list length not equal, " +
	"[keys len: %d]/[certs len:%d]"

// ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT define ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT error fmt
const ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT = "admin orgId & key list length not equal, " +
	"[keys len: %d]/[org-ids len:%d]"

var (
	//errAdminOrgIdKeyCertIsEmpty = errors.New("admin orgId or key or cert list is empty")
	noConstructorErrMsg = "contract does not have a constructor"
)

// UserContract define user contract
type UserContract struct {
	ContractName string
	Method       string
	Params       map[string]string
}

// userContractCMD user contract command
// @return *cobra.Command
func userContractCMD() *cobra.Command {
	userContractCmd := &cobra.Command{
		Use:   "user",
		Short: "user contract command",
		Long:  "user contract command",
	}

	userContractCmd.AddCommand(createUserContractCMD())
	userContractCmd.AddCommand(invokeContractTimesCMD())
	userContractCmd.AddCommand(invokeUserContractCMD())
	userContractCmd.AddCommand(upgradeUserContractCMD())
	userContractCmd.AddCommand(freezeUserContractCMD())
	userContractCmd.AddCommand(unfreezeUserContractCMD())
	userContractCmd.AddCommand(revokeUserContractCMD())
	userContractCmd.AddCommand(getUserContractCMD())

	return userContractCmd
}

// createUserContractCMD create user contract command
// @return *cobra.Command
// use as:
// ./cmc client contract user create \
// --contract-name=fact \
// --runtime-type=WASMER \
// --byte-code-path=./testdata/claim-wasm-demo/rust-fact-2.0.0.wasm \
// --version=1.0 \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key
// --admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt
// --sync-result=true \
// --params="{}"
func createUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create user contract command",
		Long:  "create user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return createUpgradeUserContract(createContractOp)
		},
	}

	attachFlags(cmd, []string{
		flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagContractName, flagVersion, flagByteCodePath, flagOrgId, flagChainId, flagSendTimes,
		flagRuntimeType, flagTimeout, flagParams, flagSyncResult, flagEnableCertHash, flagAbiFilePath,
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagGasLimit, flagPayerKeyFilePath,
		flagPayerCrtFilePath, flagPayerOrgId,
	})

	cmd.MarkFlagRequired(flagContractName)
	cmd.MarkFlagRequired(flagVersion)
	cmd.MarkFlagRequired(flagByteCodePath)
	cmd.MarkFlagRequired(flagRuntimeType)

	return cmd
}

// invokeUserContractCMD invoke user contract command
// @return *cobra.Command
// use as:
// ./cmc client contract user invoke \
// --contract-name=fact \
// --method=save \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --params="{\"file_name\":\"name007\",\"file_hash\":\"ab3456df5799b87c77e7f88\",\"time\":\"6543234\"}" \
// --sync-result=true
func invokeUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "invoke user contract command",
		Long:  "invoke user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return invokeUserContract()
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagOrgId, flagChainId, flagSendTimes,
		flagEnableCertHash, flagContractName, flagMethod, flagParams, flagTimeout, flagSyncResult, flagAbiFilePath,
		flagGasLimit, flagTxId, flagContractAddress, flagRespResultToString,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagSdkConfPath, flagPayerKeyFilePath,
		flagPayerCrtFilePath, flagPayerOrgId,
	})
	return cmd
}

// invokeContractTimesCMD invoke contract and set invoke times
// 多次的并发调用指定合约方法
// @return *cobra.Command
func invokeContractTimesCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke-times",
		Short: "invoke contract times command",
		Long:  "invoke contract times command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return invokeContractTimes()
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagEnableCertHash, flagConcurrency, flagTotalCountPerGoroutine, flagOrgId, flagChainId,
		flagSendTimes, flagContractName, flagMethod, flagParams, flagTimeout, flagSyncResult, flagAbiFilePath,
		flagContractAddress, flagGasLimit, flagRespResultToString,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagSdkConfPath, flagPayerKeyFilePath,
		flagPayerCrtFilePath, flagPayerOrgId,
	})
	return cmd
}

// getUserContractCMD query user contract command
// @return *cobra.Command
func getUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get user contract command",
		Long:  "get user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return getUserContract()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagEnableCertHash, flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagSendTimes, flagContractName, flagMethod, flagParams, flagTimeout, flagContractAddress, flagAbiFilePath,
		flagRespResultToString,
	})

	cmd.MarkFlagRequired(flagMethod)

	return cmd
}

// upgradeUserContractCMD upgrade user contract command
// @return *cobra.Command
// use as:
// ./cmc client contract user upgrade \
// --contract-name=fact \
// --runtime-type=WASMER \
// --byte-code-path=./testdata/claim-wasm-demo/rust-fact-2.0.0.wasm \
// --version=2.0 \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key
// --admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt
// --org-id=wx-org1.chainmaker.org \
// --sync-result=true \
// --params="{}"
func upgradeUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade user contract command",
		Long:  "upgrade user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return createUpgradeUserContract(upgradeContractOp)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagSdkConfPath, flagContractName, flagVersion, flagByteCodePath, flagOrgId, flagChainId, flagSendTimes,
		flagRuntimeType, flagTimeout, flagParams, flagSyncResult, flagEnableCertHash, flagContractAddress,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagGasLimit, flagAbiFilePath,
		flagPayerKeyFilePath, flagPayerCrtFilePath, flagPayerOrgId,
	})

	cmd.MarkFlagRequired(flagVersion)
	cmd.MarkFlagRequired(flagByteCodePath)
	cmd.MarkFlagRequired(flagRuntimeType)

	return cmd
}

// freezeUserContractCMD freeze user contract command
// @return *cobra.Command
// use as:
// ./cmc client contract user freeze \
// --contract-name=fact \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key
// --admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt
// --org-id=wx-org1.chainmaker.org \
// --sync-result=true
func freezeUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze",
		Short: "freeze user contract command",
		Long:  "freeze user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return freezeOrUnfreezeOrRevokeUserContract(1)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagSdkConfPath, flagContractName, flagOrgId, flagChainId, flagSendTimes, flagTimeout, flagParams,
		flagSyncResult, flagEnableCertHash, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
		flagContractAddress,
	})

	return cmd
}

// unfreezeUserContractCMD unfreeze user contract command
// @return *cobra.Command
// use as：
// ./cmc client contract user unfreeze \
// --contract-name=fact \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key
// --admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt
// --org-id=wx-org1.chainmaker.org \
// --sync-result=true
func unfreezeUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze",
		Short: "unfreeze user contract command",
		Long:  "unfreeze user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return freezeOrUnfreezeOrRevokeUserContract(2)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagSdkConfPath, flagContractName, flagOrgId, flagChainId, flagSendTimes, flagTimeout, flagParams,
		flagSyncResult, flagEnableCertHash, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
		flagContractAddress,
	})

	return cmd
}

// revokeUserContractCMD revoke user contract command
// @return *cobra.Command
// use as:
// ./cmc client contract user revoke \
// --contract-name=fact \
// --sdk-conf-path=./testdata/sdk_config.yml \
// --admin-key-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.key
// --admin-crt-file-paths=./testdata/crypto-config/wx-org1.chainmaker.org/user/admin1/admin1.sign.crt
// --org-id=wx-org1.chainmaker.org \
// --sync-result=true
func revokeUserContractCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "revoke user contract command",
		Long:  "revoke user contract command",
		RunE: func(_ *cobra.Command, _ []string) error {
			return freezeOrUnfreezeOrRevokeUserContract(3)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath,
		flagSdkConfPath, flagContractName, flagOrgId, flagChainId, flagSendTimes, flagTimeout, flagParams,
		flagSyncResult, flagEnableCertHash, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
		flagContractAddress,
	})

	return cmd
}

type createUpgradeContractOp int

const (
	createContractOp createUpgradeContractOp = iota + 1
	upgradeContractOp
)

// nolint
func createUpgradeUserContract(op createUpgradeContractOp) error {
	client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(client, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
	if err != nil {
		return err
	}

	rt, ok := common.RuntimeType_value[runtimeType]
	if !ok {
		return fmt.Errorf("unknown runtime type [%s]", runtimeType)
	}

	var kvs []*common.KeyValuePair

	if runtimeType != "EVM" {
		if params != "" {
			kvsMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(params), &kvsMap)
			if err != nil {
				return err
			}
			kvs = util.ConvertParameters(kvsMap)
		}
	} else { // EVM contract deploy
		if abiFilePath == "" {
			return errors.New("required abi file path when deploy EVM contract")
		}
		abiBytes, err := ioutil.ReadFile(abiFilePath)
		if err != nil {
			return err
		}

		contractAbi, err := abi.JSON(bytes.NewReader(abiBytes))
		if err != nil {
			return err
		}

		inputData, err := util.Pack(contractAbi, "", params)
		if err != nil {
			if err.Error() != noConstructorErrMsg {
				return err
			}
		}

		inputDataHexStr := hex.EncodeToString(inputData)
		kvs = []*common.KeyValuePair{
			{
				Key:   "data",
				Value: []byte(inputDataHexStr),
			},
		}

		byteCode, err := ioutil.ReadFile(byteCodePath)
		if err != nil {
			return err
		}
		byteCodePath = string(byteCode)
	}

	var payload *common.Payload
	switch op {
	case createContractOp:
		payload, err = client.CreateContractCreatePayload(contractName, version,
			byteCodePath, common.RuntimeType(rt), kvs)
		if err != nil {
			return err
		}
	case upgradeContractOp:
		if contractAddress != "" {
			contractName = contractAddress
		}
		if contractName == "" {
			return errors.New("either contract-name or contract-address must be set")
		}
		payload, err = client.CreateContractUpgradePayload(contractName, version,
			byteCodePath, common.RuntimeType(rt), kvs)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown operation")
	}

	if gasLimit > 0 {
		var limit = &common.Limit{GasLimit: gasLimit}
		payload = client.AttachGasLimit(payload, limit)
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	var payer []*common.EndorsementEntry
	if len(payerKeyFilePath) > 0 {
		payer, err = util.MakeEndorsement([]string{payerKeyFilePath}, []string{payerCrtFilePath}, []string{payerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return err
		}
	}
	var resp *common.TxResponse
	if len(payer) == 0 {
		resp, err = client.SendContractManageRequest(payload, endorsementEntrys, timeout, syncResult)
	} else {
		resp, err = client.SendContractManageRequestWithPayer(payload, endorsementEntrys, payer[0], timeout, syncResult)
	}
	if err != nil {
		return err
	}
	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}
	return createUpgradeUserContractOutput(resp)
}

func createUpgradeUserContractOutput(resp *common.TxResponse) error {
	if resp.ContractResult != nil && resp.ContractResult.Result != nil {
		var contract common.Contract
		err := contract.Unmarshal(resp.ContractResult.Result)
		if err != nil {
			return err
		}
		util.PrintPrettyJson(types.CreateUpgradeContractTxResponse{
			TxResponse: resp,
			ContractResult: &types.CreateUpgradeContractContractResult{
				ContractResult: resp.ContractResult,
				Result:         &contract,
			},
		})
	} else {
		util.PrintPrettyJson(resp)
	}
	return nil
}

func invokeUserContract() error {
	cc, err := sdk.NewChainClient(
		sdk.WithConfPath(sdkConfPath),
		sdk.WithChainClientChainId(chainId),
		sdk.WithChainClientOrgId(orgId),
		sdk.WithUserCrtFilePath(userTlsCrtFilePath),
		sdk.WithUserKeyFilePath(userTlsKeyFilePath),
		sdk.WithUserSignCrtFilePath(userSignCrtFilePath),
		sdk.WithUserSignKeyFilePath(userSignKeyFilePath),
	)
	if err != nil {
		return err
	}
	defer cc.Stop()
	if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
		return err
	}

	if contractAddress != "" {
		contractName = contractAddress
	}
	if contractName == "" {
		return errors.New("either contract-name or contract-address must be set")
	}

	var kvs []*common.KeyValuePair
	var contractAbi *abi.ABI

	if abiFilePath != "" { // abi file path 非空 意味着调用的是EVM合约
		abiBytes, err := ioutil.ReadFile(abiFilePath)
		if err != nil {
			return err
		}

		contractAbi, err = abi.JSON(bytes.NewReader(abiBytes))
		if err != nil {
			return err
		}

		inputData, err := util.Pack(contractAbi, method, params)
		if err != nil {
			return err
		}

		inputDataHexStr := hex.EncodeToString(inputData)

		kvs = []*common.KeyValuePair{
			{
				Key:   "data",
				Value: []byte(inputDataHexStr),
			},
		}
	} else {
		if params != "" {
			kvsMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(params), &kvsMap)
			if err != nil {
				return err
			}
			kvs = util.ConvertParameters(kvsMap)
		}
	}

	var limit *common.Limit
	if gasLimit > 0 {
		limit = &common.Limit{GasLimit: gasLimit}
	}

	if txId != "" {
		invokeContract(cc, contractName, method, txId, kvs, contractAbi, limit)
	} else {
		Dispatch(cc, contractName, method, kvs, contractAbi, limit)
	}
	return nil
}

func invokeContractTimes() error {
	client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	if contractAddress != "" {
		contractName = contractAddress
	}
	if contractName == "" {
		return errors.New("either contract-name or contract-address must be set")
	}

	var kvs []*common.KeyValuePair
	var evmMethod *abi.ABI

	if abiFilePath != "" { // abi file path 非空 意味着调用的是EVM合约
		abiBytes, err := ioutil.ReadFile(abiFilePath)
		if err != nil {
			return err
		}

		contractAbi, err := abi.JSON(bytes.NewReader(abiBytes))
		if err != nil {
			return err
		}

		//m, exist := contractAbi.Methods[method]
		//if !exist {
		//	return fmt.Errorf("method '%s' not found", method)
		//}
		//evmMethod = &m
		evmMethod = contractAbi

		inputData, err := util.Pack(contractAbi, method, params)
		if err != nil {
			return err
		}

		inputDataHexStr := hex.EncodeToString(inputData)
		method = inputDataHexStr[0:8]

		kvs = []*common.KeyValuePair{
			{
				Key:   "data",
				Value: []byte(inputDataHexStr),
			},
		}
	} else {
		if params != "" {
			kvsMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(params), &kvsMap)
			if err != nil {
				return err
			}
			kvs = util.ConvertParameters(kvsMap)
		}
	}

	var limit *common.Limit
	if gasLimit > 0 {
		limit = &common.Limit{GasLimit: gasLimit}
	}

	DispatchTimes(client, contractName, method, kvs, evmMethod, limit)
	return nil
}

func getUserContract() error {
	client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	if contractAddress != "" {
		contractName = contractAddress
	}
	if contractName == "" {
		return errors.New("either contract-name or contract-address must be set")
	}

	kvs, contractAbi, err := generateKVs()
	if err != nil {
		return err
	}

	resp, err := client.QueryContract(contractName, method, kvs, -1)
	if err != nil {
		return fmt.Errorf("query contract failed, %s", err.Error())
	}

	if resp.Code != common.TxStatusCode_SUCCESS {
		util.PrintPrettyJson(resp)
		return nil
	}

	var output interface{}
	if contractAbi != nil && resp.ContractResult != nil && resp.ContractResult.Result != nil {
		unpackedData, err := contractAbi.Unpack(method, resp.ContractResult.Result)
		if err != nil {
			fmt.Println(err)
			return nil
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
	return nil
}

func generateKVs() (kvs []*common.KeyValuePair, contractAbi *abi.ABI, err error) {
	// handle evm params
	if abiFilePath != "" { // abi file path 非空 意味着调用的是EVM合约
		abiBytes, err := ioutil.ReadFile(abiFilePath)
		if err != nil {
			return nil, nil, err
		}

		contractAbi, err = abi.JSON(bytes.NewReader(abiBytes))
		if err != nil {
			return nil, nil, err
		}

		inputData, err := util.Pack(contractAbi, method, params)
		if err != nil {
			return nil, nil, err
		}

		inputDataHexStr := hex.EncodeToString(inputData)

		kvs = []*common.KeyValuePair{
			{
				Key:   "data",
				Value: []byte(inputDataHexStr),
			},
		}

		return kvs, contractAbi, nil

	}

	// handle common params
	if params != "" {
		kvsMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(params), &kvsMap)
		if err == nil {
			kvs = util.ConvertParameters(kvsMap)
		} else {
			var pms []*Param
			err = json.Unmarshal([]byte(params), &pms)
			if err != nil {
				return nil, nil, err
			}
			for _, pm := range pms {
				if pm.IsFile {
					byteCode, err := ioutil.ReadFile(pm.Value)
					if err != nil {
						return nil, nil, err
					}
					kvs = append(kvs, &common.KeyValuePair{
						Key:   pm.Key,
						Value: byteCode,
					})
				} else {
					kvs = append(kvs, &common.KeyValuePair{
						Key:   pm.Key,
						Value: []byte(pm.Value),
					})
				}
			}
		}

		return kvs, nil, nil
	}

	return []*common.KeyValuePair{}, nil, nil
}

func freezeOrUnfreezeOrRevokeUserContract(which int) error {
	client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	if contractAddress != "" {
		contractName = contractAddress
	}
	if contractName == "" {
		return errors.New("either contract-name or contract-address must be set")
	}

	adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(client, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
	if err != nil {
		return err
	}

	var (
		payload        *common.Payload
		whichOperation string
	)

	switch which {
	case 1:
		payload, err = client.CreateContractFreezePayload(contractName)
		whichOperation = "freeze"
	case 2:
		payload, err = client.CreateContractUnfreezePayload(contractName)
		whichOperation = "unfreeze"
	case 3:
		payload, err = client.CreateContractRevokePayload(contractName)
		whichOperation = "revoke"
	default:
		err = fmt.Errorf("wrong which param")
	}
	if err != nil {
		return fmt.Errorf("create cert manage %s payload failed, %s", whichOperation, err.Error())
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}
	// 发送创建合约请求
	resp, err := client.SendContractManageRequest(payload, endorsementEntrys, timeout, syncResult)
	if err != nil {
		return fmt.Errorf(SEND_CONTRACT_MANAGE_REQUEST_FAILED_FORMAT, err.Error())
	}

	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return fmt.Errorf(CHECK_PROPOSAL_RESPONSE_FAILED_FORMAT, err.Error())
	}
	util.PrintPrettyJson(resp)
	return nil
}
