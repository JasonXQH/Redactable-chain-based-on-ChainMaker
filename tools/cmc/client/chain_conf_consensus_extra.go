/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/spf13/cobra"
)

const (
	addExtraConfig = iota
	deleteExtraConfig
	updateExtraConfig
)

// configConsensueExtraCMD consensus extra config management sub command
//用于管理Consensus下的ext_config
// @return *cobra.Command
func configConsensueExtraCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensusextra",
		Short: "consensus extra config management",
		Long:  "consensus extra config management",
	}
	cmd.AddCommand(addConsensusExtraConfigCMD())
	cmd.AddCommand(deleteConsensExtraConfigCMD())
	cmd.AddCommand(updateConsensusExtraConfigCMD())

	return cmd
}

// addConsensusExtraConfigCMD add consensus extra config command
// @return *cobra.Command
func addConsensusExtraConfigCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add consensus extra config cmd",
		Long:  "add consensus extra config cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusExtra(addExtraConfig)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagChainId,
		flagSdkConfPath, flagOrgId, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagExtraConfigKey,
		flagExtraConfigValue,
	})

	cmd.MarkFlagRequired(flagExtraConfigKey)
	cmd.MarkFlagRequired(flagExtraConfigValue)

	return cmd
}

// deleteConsensExtraConfigCMD delete consensus extra config command
// @return *cobra.Command
func deleteConsensExtraConfigCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete consensus extra config cmd",
		Long:  "delete consensus extra config cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusExtra(deleteExtraConfig)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagChainId,
		flagSdkConfPath, flagOrgId, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagExtraConfigKey,
	})

	cmd.MarkFlagRequired(flagExtraConfigKey)

	return cmd
}

// updateConsensusNodeIdCMD update consensus extra config command
// @return *cobra.Command
func updateConsensusExtraConfigCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update consensus extra config cmd",
		Long:  "update consensus extra config cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusExtra(updateExtraConfig)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagChainId,
		flagSdkConfPath, flagOrgId, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagExtraConfigKey,
		flagExtraConfigValue,
	})

	cmd.MarkFlagRequired(flagExtraConfigKey)
	cmd.MarkFlagRequired(flagExtraConfigValue)

	return cmd
}

// configConsensusExtra update ChainConfig.Consensus.ext_config
// @param op 0:add;1:delete;2:update
// @return error
func configConsensusExtra(op int) error {
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

	kvs := []*common.KeyValuePair{
		{
			Key:   extraConfigKey,
			Value: []byte(extraConfigValue),
		},
	}

	var payload *common.Payload
	switch op {
	case addExtraConfig:
		payload, err = client.CreateChainConfigConsensusExtAddPayload(kvs)
	case deleteExtraConfig:
		keys := []string{extraConfigKey}
		payload, err = client.CreateChainConfigConsensusExtDeletePayload(keys)
	case updateExtraConfig:
		payload, err = client.CreateChainConfigConsensusExtUpdatePayload(kvs)
	default:
		err = errors.New("invalid operation")
	}
	if err != nil {
		return err
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
