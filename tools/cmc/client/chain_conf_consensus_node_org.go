/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/spf13/cobra"
)

const (
	addNodeOrg = iota
	removeNodeOrg
	updateNodeOrg
)

// configConsensueNodeOrgCMD consensus node org management
//用于管理Consensus下的Org
// @return *cobra.Command
func configConsensueNodeOrgCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensusnodeorg",
		Short: "consensus node org management",
		Long:  "consensus node org management",
	}
	cmd.AddCommand(addConsensusNodeOrgCMD())
	cmd.AddCommand(removeConsensusNodeOrgCMD())
	cmd.AddCommand(updateConsensusNodeOrgCMD())

	return cmd
}

// addConsensusNodeOrgCMD add consensus node org
// @return *cobra.Command
func addConsensusNodeOrgCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add consensus node org cmd",
		Long:  "add consensus node org cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusNodeOrg(addNodeOrg)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagNodeOrgId, flagNodeIds,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagNodeOrgId)
	cmd.MarkFlagRequired(flagNodeIds)

	return cmd
}

// removeConsensusNodeOrgCMD remove consensus node org
// @return *cobra.Command
func removeConsensusNodeOrgCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "remove consensus node org cmd",
		Long:  "remove consensus node org cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusNodeOrg(removeNodeOrg)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagNodeOrgId,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagNodeOrgId)

	return cmd
}

// updateConsensusNodeOrgCMD update consensus node org
// @return *cobra.Command
func updateConsensusNodeOrgCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update consensus node org cmd",
		Long:  "update consensus node org cmd",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configConsensusNodeOrg(updateNodeOrg)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagNodeOrgId, flagNodeIds,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagNodeOrgId)
	cmd.MarkFlagRequired(flagNodeIds)

	return cmd
}

func configConsensusNodeOrg(op int) error {
	nodeIdSlice := strings.Split(nodeIds, ",")

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
	var payload *common.Payload
	switch op {
	case addNodeOrg:
		payload, err = client.CreateChainConfigConsensusNodeOrgAddPayload(nodeOrgId, nodeIdSlice)
	case removeNodeOrg:
		payload, err = client.CreateChainConfigConsensusNodeOrgDeletePayload(nodeOrgId)
	case updateNodeOrg:
		payload, err = client.CreateChainConfigConsensusNodeOrgUpdatePayload(nodeOrgId, nodeIdSlice)
	default:
		err = errors.New("invalid node address operation")
	}
	if err != nil {
		return err
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}
	resp, err := client.SendChainConfigUpdateRequest(payload, endorsementEntrys, timeout, syncResult)
	if err != nil {
		return err
	}
	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}
	fmt.Printf("consensusnodeorg response %+v\n", resp)
	return nil
}
