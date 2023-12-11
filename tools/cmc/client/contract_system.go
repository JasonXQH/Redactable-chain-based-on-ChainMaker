/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/json"
	"fmt"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
)

// DEFAULT_TIMEOUT define default timeout in ms
const DEFAULT_TIMEOUT = 5000 // ms
// systemContractCMD system contract command
// @return *cobra.Command
func systemContractCMD() *cobra.Command {
	systemContractCmd := &cobra.Command{
		Use:   "system",
		Short: "system contract command",
		Long:  "system contract command",
	}

	systemContractCmd.AddCommand(getChainInfoCMD())
	systemContractCmd.AddCommand(getBlockByHeightCMD())
	systemContractCmd.AddCommand(getTxByTxIdCMD())

	// DPoS-erc20 contract
	systemContractCmd.AddCommand(erc20Mint())
	systemContractCmd.AddCommand(erc20Transfer())
	systemContractCmd.AddCommand(erc20BalanceOf())
	systemContractCmd.AddCommand(erc20Owner())
	systemContractCmd.AddCommand(erc20Decimals())
	systemContractCmd.AddCommand(erc20Total())
	//
	////DPoS.Stake
	systemContractCmd.AddCommand(stakeGetAllCandidates())
	systemContractCmd.AddCommand(stakeGetValidatorByAddress())
	systemContractCmd.AddCommand(stakeDelegate())
	systemContractCmd.AddCommand(stakeGetDelegationsByAddress())
	systemContractCmd.AddCommand(stakeGetUserDelegationByValidator())
	systemContractCmd.AddCommand(stakeUnDelegate())
	systemContractCmd.AddCommand(stakeReadEpochByID())
	systemContractCmd.AddCommand(stakeReadLatestEpoch())
	systemContractCmd.AddCommand(stakeSetNodeID())
	systemContractCmd.AddCommand(stakeGetNodeID())
	systemContractCmd.AddCommand(stakeReadMinSelfDelegation())
	systemContractCmd.AddCommand(stakeReadEpochValidatorNumber())
	systemContractCmd.AddCommand(stakeReadEpochBlockNumber())
	systemContractCmd.AddCommand(stakeReadSystemContractAddr())
	systemContractCmd.AddCommand(stakeReadCompleteUnBoundingEpochNumber())

	// system contract manage
	systemContractCmd.AddCommand(systemContractManageCMD())

	// system contract multi sign
	systemContractCmd.AddCommand(systemContractMultiSignCMD())

	// DPoS-stake contract
	return systemContractCmd
}

// getChainInfoCMD get chain info
// @return *cobra.Command include latest block height and server connected peers info
func getChainInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getchaininfo",
		Short: "get chain info",
		Long:  "get chain info",
		RunE: func(_ *cobra.Command, _ []string) error {
			return getChainInfo()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagParams, flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash,
	})

	return cmd
}

// getBlockByHeightCMD get block by height
// @return *cobra.Command
func getBlockByHeightCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "get block by height",
		Long:  "get block by height",
		RunE: func(_ *cobra.Command, _ []string) error {
			return getBlockByHeight()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagChainId, flagBlockHeight, flagWithRWSet, flagTruncateValue,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagBlockHeight)
	cmd.MarkFlagRequired(flagWithRWSet)

	return cmd
}

// getTxByTxIdCMD get tx by tx id
// @return *cobra.Command
func getTxByTxIdCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "get tx by tx id",
		Long:  "get tx by tx id",
		RunE: func(_ *cobra.Command, _ []string) error {
			return getTxByTxId()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagChainId, flagTxId,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

func getChainInfo() error {
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

	resp, err := client.GetChainInfo()
	if err != nil {
		return fmt.Errorf("get chain info failed, %s", err.Error())
	}

	fmt.Printf("get chain info resp: %+v\n", resp)

	return nil
}

func getBlockByHeight() error {
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
	var resp *common.BlockInfo
	if truncateValue {
		resp, err = client.GetBlockByHeightTruncate(blockHeight, withRWSet, 1000, "truncate")
	} else {
		resp, err = client.GetBlockByHeight(blockHeight, withRWSet)
	}
	if err != nil {
		return fmt.Errorf("get block by height failed, %s", err.Error())
	}
	blockJson, _ := prettyjson.Marshal(resp)
	fmt.Println(string(blockJson))
	return nil
}

func getTxByTxId() error {
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
	var resp *common.TransactionInfoWithRWSet
	if truncateValue {
		resp, err = client.GetTxByTxIdTruncate(txId, withRWSet, 1000, "truncate")
	} else {
		if withRWSet {
			resp, err = client.GetTxWithRWSetByTxId(txId)
		} else {
			tx, err := client.GetTxByTxId(txId)
			if err != nil {
				return fmt.Errorf("get tx by txid failed, %s", err.Error())
			}
			resp = &common.TransactionInfoWithRWSet{
				Transaction:    tx.Transaction,
				BlockHeight:    tx.BlockHeight,
				BlockHash:      tx.BlockHash,
				TxIndex:        tx.TxIndex,
				BlockTimestamp: tx.BlockTimestamp,
			}
		}
	}
	if err != nil {
		return fmt.Errorf("get tx by txid failed, %s", err.Error())
	}

	fmt.Printf("get tx by txid resp: %+v\n", resp)

	return nil
}
