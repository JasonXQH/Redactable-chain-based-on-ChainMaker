/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/spf13/cobra"
)

// updateBlockConfigCMD update block config sub command
// @return *cobra.Command
func updateBlockConfigCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "update block command",
		Long:  "update block command",
	}
	cmd.AddCommand(updateBlockIntervalCMD())
	cmd.AddCommand(updateTxParameterSizeCMD())

	return cmd
}

// updateBlockIntervalCMD update block interval
// @return *cobra.Command
func updateBlockIntervalCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "updateblockinterval",
		Short: "update block interval",
		Long:  "update block interval",
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateBlockInterval()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagChainId,
		flagSdkConfPath, flagOrgId, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagBlockInterval,
	})

	cmd.MarkFlagRequired(flagBlockInterval)

	return cmd
}

func updateBlockInterval() error {

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
	chainConfig, err := client.GetChainConfig()
	if err != nil {
		return fmt.Errorf("get chain config failed, %s", err.Error())
	}
	txTimestampVerify := chainConfig.Block.TxTimestampVerify
	txTimeout := chainConfig.Block.TxTimeout
	blockTxCap := chainConfig.Block.BlockTxCapacity
	blockSize := chainConfig.Block.BlockSize
	txParameterSize = chainConfig.Block.TxParameterSize

	payload, err := client.CreateChainConfigBlockUpdatePayload(txTimestampVerify, txTimeout, blockTxCap,
		blockSize, blockInterval, txParameterSize)
	if err != nil {
		return fmt.Errorf("create chain config block update payload failed, %s", err.Error())
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	resp, err := client.SendChainConfigUpdateRequest(payload, endorsementEntrys, -1, true)
	if err != nil {
		return fmt.Errorf("send chain config update request failed, %s", err.Error())
	}
	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return fmt.Errorf("check proposal request resp failed, %s", err.Error())
	}
	fmt.Printf("response %+v\n", resp)
	return nil
}

// updateTxParameterSizeCMD update txparameter size
// @return *cobra.Command
func updateTxParameterSizeCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "updatetxparametersize",
		Short: "update txparameter size",
		Long:  "update txparameter size",
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateTxParameterSize()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagChainId,
		flagSdkConfPath, flagOrgId, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagTxParameterSize,
	})

	cmd.MarkFlagRequired(flagTxParameterSize)

	return cmd
}

func updateTxParameterSize() error {

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
	chainConfig, err := client.GetChainConfig()
	if err != nil {
		return fmt.Errorf("get chain config failed, %s", err.Error())
	}
	txTimestampVerify := chainConfig.Block.TxTimestampVerify
	txTimeout := chainConfig.Block.TxTimeout
	blockTxCap := chainConfig.Block.BlockTxCapacity
	blockSize := chainConfig.Block.BlockSize
	blockInterval = chainConfig.Block.BlockInterval

	payload, err := client.CreateChainConfigBlockUpdatePayload(txTimestampVerify, txTimeout, blockTxCap,
		blockSize, blockInterval, txParameterSize)
	if err != nil {
		return fmt.Errorf("create chain config block update payload failed, %s", err.Error())
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	resp, err := client.SendChainConfigUpdateRequest(payload, endorsementEntrys, -1, true)
	if err != nil {
		return fmt.Errorf("send chain config update request failed, %s", err.Error())
	}
	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return fmt.Errorf("check proposal request resp failed, %s", err.Error())
	}
	fmt.Printf("response %+v\n", resp)
	return nil
}
