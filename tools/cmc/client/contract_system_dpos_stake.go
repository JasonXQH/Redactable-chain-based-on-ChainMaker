/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/json"
	"fmt"

	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
)

// stakeGetAllCandidates all-candidates feature of the DPoS stake
// @return *cobra.Command
func stakeGetAllCandidates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-candidates",
		Short: "all-candidates feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := getAllCandidates(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("all-candidates failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeGetValidatorByAddress get-validator feature of the DPoS stake
// @return *cobra.Command
func stakeGetValidatorByAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-validator",
		Short: "get-validator feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := getValidatorByAddress(client, address, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("get-validator failed, %s", err.Error())
			}
			val := &syscontract.Validator{}
			if err := proto.Unmarshal(resp.ContractResult.Result, val); err != nil {
				fmt.Println("unmarshal validatorInfo failed")
				return nil
			}
			resp.ContractResult.Result = []byte(val.String())
			fmt.Printf("resp: %+v\n", resp)
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAddress,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	cmd.MarkFlagRequired(flagAddress)

	return cmd
}

// stakeDelegate delegate feature of the stake
// @return *cobra.Command
func stakeDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				if err := json.Unmarshal([]byte(params), &pairs); err != nil {
					return err
				}
			}
			txId = sdkutils.GetTimestampTxId()
			resp, err := delegate(client, address, amount, txId, DEFAULT_TIMEOUT, syncResult)
			if err != nil {
				return fmt.Errorf("delegate failed, %s", err.Error())
			}
			if resp.ContractResult == nil {
				fmt.Printf("resp: %+v\n", resp)
				return nil
			}
			info := &syscontract.Delegation{}
			if err := proto.Unmarshal(resp.ContractResult.Result, info); err != nil {
				return fmt.Errorf("unmarshal delegate info failed, %v", err)
			}
			resp.ContractResult.Result = []byte(info.String())
			fmt.Printf("resp: %+v\n", resp)
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAddress, flagAmount,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
		flagUserSignCrtFilePath, flagUserSignKeyFilePath,
		flagSyncResult,
	})

	cmd.MarkFlagRequired(flagAddress)
	cmd.MarkFlagRequired(flagAmount)

	return cmd
}

// stakeGetDelegationsByAddress get-delegations-by-address feature of the stake
// @return *cobra.Command
func stakeGetDelegationsByAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-delegations-by-address",
		Short: "get-delegations-by-address feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := getDelegationsByAddress(client, address, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("get-delegations-by-address failed, %s", err.Error())
			}
			delegateInfo := &syscontract.DelegationInfo{}
			if err := proto.Unmarshal(resp.ContractResult.Result, delegateInfo); err != nil {
				fmt.Println("unmarshal delegateInfo failed: ", err)
				return nil
			}
			resp.ContractResult.Result = []byte(delegateInfo.String())
			fmt.Printf("resp: %+v\n", resp)
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAddress,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	cmd.MarkFlagRequired(flagAddress)

	return cmd
}

// stakeGetUserDelegationByValidator get-delegations-by-address feature of the stake
// @return *cobra.Command
func stakeGetUserDelegationByValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-user-delegation-by-validator",
		Short: "get-delegations-by-address feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := getUserDelegationByValidator(client, delegator, validator, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("get-user-delegation-by-validator failed, %s", err.Error())
			}
			info := &syscontract.Delegation{}
			if err := proto.Unmarshal(resp.ContractResult.Result, info); err != nil {
				return fmt.Errorf("unmarshal delegate info failed, %v", err)
			}
			resp.ContractResult.Result = []byte(info.String())
			fmt.Printf("resp: %+v\n", resp)
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagDelegator, flagValidator,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	cmd.MarkFlagRequired(flagDelegator)
	cmd.MarkFlagRequired(flagValidator)

	return cmd
}

// stakeUnDelegate undo delegate of the stake
// @return *cobra.Command
func stakeUnDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate",
		Short: "undelegate feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}
			txId = sdkutils.GetTimestampTxId()
			resp, err := unDelegate(client, address, amount, txId, DEFAULT_TIMEOUT, syncResult)
			if err != nil {
				return fmt.Errorf("undelegate failed, %s", err.Error())
			}
			if resp.ContractResult == nil {
				fmt.Printf("resp: %+v\n", resp)
				return nil
			}
			info := &syscontract.UnbondingDelegation{}
			if err := proto.Unmarshal(resp.ContractResult.Result, info); err != nil {
				return fmt.Errorf("unmarshal UnbondingDelegation info failed, %v", err)
			}
			resp.ContractResult.Result = []byte(info.String())
			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAddress, flagAmount,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
		flagUserSignCrtFilePath, flagUserSignKeyFilePath,
		flagSyncResult,
	})

	cmd.MarkFlagRequired(flagAddress)
	cmd.MarkFlagRequired(flagAmount)

	return cmd
}

// stakeReadEpochByID read-epoch-by-id feature of the stake
// @return *cobra.Command
func stakeReadEpochByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read-epoch-by-id",
		Short: "read-epoch-by-id feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readEpochByID(client, epochID, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("read-epoch-by-id failed, %s", err.Error())
			}
			info := &syscontract.Epoch{}
			if err := proto.Unmarshal(resp.ContractResult.Result, info); err != nil {
				return fmt.Errorf("unmarshal epoch err: %v", err)
			}
			resp.ContractResult.Result = []byte(info.String())
			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagEpochID,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	cmd.MarkFlagRequired(flagEpochID)

	return cmd
}

// stakeReadLatestEpoch read-latest-epoch feature of the stake
// @return *cobra.Command
func stakeReadLatestEpoch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read-latest-epoch",
		Short: "read-latest-epoch feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readLatestEpoch(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("read-latest-epoch failed, %s", err.Error())
			}
			info := &syscontract.Epoch{}
			if err := proto.Unmarshal(resp.ContractResult.Result, info); err != nil {
				return fmt.Errorf("unmarshal epoch err: %v", err)
			}
			resp.ContractResult.Result = []byte(info.String())
			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeSetNodeID set-node-id feature of the stake
// @return *cobra.Command
func stakeSetNodeID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-node-id",
		Short: "set-node-id feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := setNodeID(client, nodeId, DEFAULT_TIMEOUT, syncResult)
			if err != nil {
				return fmt.Errorf("set-node-id failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagNodeId,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
		flagUserSignCrtFilePath, flagUserSignKeyFilePath,
		flagSyncResult,
	})

	cmd.MarkFlagRequired(flagNodeId)

	return cmd
}

// stakeGetNodeID get-node-id feature of the stake
// @return *cobra.Command
func stakeGetNodeID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-node-id",
		Short: "get-node-id feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := getNodeID(client, address, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("get-node-id failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAddress,
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	cmd.MarkFlagRequired(flagAddress)

	return cmd
}

// stakeReadMinSelfDelegation min-self-delegation feature of the stake
// @return *cobra.Command
func stakeReadMinSelfDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-self-delegation",
		Short: "min-self-delegation feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readMinSelfDelegation(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("min-self-delegation failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeReadEpochValidatorNumber validator-number feature of the stake
// @return *cobra.Command
func stakeReadEpochValidatorNumber() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-number",
		Short: "validator-number feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readEpochValidatorNumber(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("validator-number failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeReadEpochBlockNumber epoch-block-number feature of the stake
// @return *cobra.Command
func stakeReadEpochBlockNumber() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-block-number",
		Short: "epoch-block-number feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readEpochBlockNumber(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("epoch-block-number failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeReadSystemContractAddr system-address feature of the stake
// @return *cobra.Command
func stakeReadSystemContractAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system-address",
		Short: "system-address feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readSystemContractAddr(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("system-address failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

// stakeReadCompleteUnBoundingEpochNumber unbonding-epoch-number feature of the stake
// @return *cobra.Command
func stakeReadCompleteUnBoundingEpochNumber() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-epoch-number",
		Short: "unbonding-epoch-number feature of the stake",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err error
			)

			client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
				userSignCrtFilePath, userSignKeyFilePath)
			if err != nil {
				return err
			}
			defer client.Stop()
			pairs := make(map[string]string)
			if params != "" {
				err := json.Unmarshal([]byte(params), &pairs)
				if err != nil {
					return err
				}
			}

			resp, err := readCompleteUnBoundingEpochNumber(client, DEFAULT_TIMEOUT)
			if err != nil {
				return fmt.Errorf("unbonding-epoch-number failed, %s", err.Error())
			}

			fmt.Printf("resp: %+v\n", resp)

			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagOrgId, flagChainId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagUserSignCrtFilePath, flagUserSignKeyFilePath,
	})

	return cmd
}

func getAllCandidates(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_GET_ALL_CANDIDATES.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func getValidatorByAddress(cc *sdk.ChainClient, address string, timeout int64) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"address": address,
	}
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_GET_VALIDATOR_BY_ADDRESS.String(),
		util.ConvertParameters(params),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func delegate(cc *sdk.ChainClient, address, amount string, txId string, timeout int64,
	withSyncResult bool) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"to":     address,
		"amount": amount,
	}
	if txId == "" {
		txId = sdkutils.GetTimestampTxId()
	}
	resp, err := cc.InvokeSystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_DELEGATE.String(),
		txId,
		util.ConvertParameters(params),
		timeout,
		withSyncResult,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_INVOKE_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func getDelegationsByAddress(cc *sdk.ChainClient, address string, timeout int64) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"address": address,
	}
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_GET_DELEGATIONS_BY_ADDRESS.String(),
		util.ConvertParameters(params),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func getUserDelegationByValidator(cc *sdk.ChainClient, delegatorAddress, validatorAddress string,
	timeout int64) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"delegator_address": delegatorAddress,
		"validator_address": validatorAddress,
	}
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_GET_USER_DELEGATION_BY_VALIDATOR.String(),
		util.ConvertParameters(params),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func unDelegate(cc *sdk.ChainClient, address, amount string, txId string, timeout int64,
	withSyncResult bool) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"from":   address,
		"amount": amount,
	}
	if txId == "" {
		txId = sdkutils.GetTimestampTxId()
	}
	resp, err := cc.InvokeSystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_UNDELEGATE.String(),
		txId,
		util.ConvertParameters(params),
		timeout,
		withSyncResult,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_INVOKE_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readEpochByID(cc *sdk.ChainClient, epochID string, timeout int64) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"epoch_id": epochID,
	}
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_EPOCH_BY_ID.String(),
		util.ConvertParameters(params),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readLatestEpoch(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_LATEST_EPOCH.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func setNodeID(cc *sdk.ChainClient, nodeID string, timeout int64, withSyncResult bool) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"node_id": nodeID,
	}
	if txId == "" {
		txId = sdkutils.GetTimestampTxId()
	}
	resp, err := cc.InvokeSystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_SET_NODE_ID.String(),
		txId,
		util.ConvertParameters(params),
		timeout,
		withSyncResult,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_INVOKE_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func getNodeID(cc *sdk.ChainClient, address string, timeout int64) (*common.TxResponse, error) {
	params := map[string]interface{}{
		"address": address,
	}
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_GET_NODE_ID.String(),
		util.ConvertParameters(params),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readMinSelfDelegation(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_MIN_SELF_DELEGATION.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readEpochValidatorNumber(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_EPOCH_VALIDATOR_NUMBER.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readEpochBlockNumber(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_EPOCH_BLOCK_NUMBER.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readSystemContractAddr(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_SYSTEM_CONTRACT_ADDR.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}

func readCompleteUnBoundingEpochNumber(cc *sdk.ChainClient, timeout int64) (*common.TxResponse, error) {
	resp, err := cc.QuerySystemContract(
		syscontract.SystemContract_DPOS_STAKE.String(),
		syscontract.DPoSStakeFunction_READ_COMPLETE_UNBOUNDING_EPOCH_NUMBER.String(),
		nil,
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%s failed, %s", common.TxType_QUERY_CONTRACT.String(), err.Error())
	}

	return resp, nil
}
