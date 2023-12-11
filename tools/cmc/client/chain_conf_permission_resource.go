/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

const (
	addPermissionResourceEnum = iota + 1
	updatePermissionResourceEnum
	deletePermissionResourceEnum
	listPermissionResourceEnum
)

// permissionResourceCMD chain config permission resource operation
// @return *cobra.Command
func permissionResourceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permission",
		Short: "chain config permission resource operation",
		Long:  "chain config permission resource operation",
	}
	cmd.AddCommand(addPermissionResourceCMD())
	cmd.AddCommand(updatePermissionResourceCMD())
	cmd.AddCommand(deletePermissionResourceCMD())
	cmd.AddCommand(listPermissionResourceCMD())
	return cmd
}

// addPermissionResourceCMD add chain config permission resource
// @return *cobra.Command
func addPermissionResourceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add chain config permission resource",
		Long:  "add chain config permission resource",
		RunE: func(_ *cobra.Command, _ []string) error {
			return doPermissionResourceOperation(addPermissionResourceEnum)
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagOrgId,
		flagChainId, flagTimeout, flagSyncResult, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
		flagPermissionResourcePolicyRule, flagPermissionResourcePolicyOrgList, flagPermissionResourcePolicyRoleList,
	})

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagPermissionResourceName,
	})

	return cmd
}

// updatePermissionResourceCMD update chain config permission resource
// @return *cobra.Command
func updatePermissionResourceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update chain config permission resource",
		Long:  "update chain config permission resource",
		RunE: func(_ *cobra.Command, _ []string) error {
			return doPermissionResourceOperation(updatePermissionResourceEnum)
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagOrgId,
		flagChainId, flagTimeout, flagSyncResult, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
		flagPermissionResourcePolicyRule, flagPermissionResourcePolicyOrgList, flagPermissionResourcePolicyRoleList,
	})

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagPermissionResourceName,
	})

	return cmd
}

// deletePermissionResourceCMD delete chain config permission resource
// @return *cobra.Command
func deletePermissionResourceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete chain config permission resource",
		Long:  "delete chain config permission resource",
		RunE: func(_ *cobra.Command, _ []string) error {
			return doPermissionResourceOperation(deletePermissionResourceEnum)
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagOrgId,
		flagChainId, flagTimeout, flagSyncResult, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
	})

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagPermissionResourceName,
	})

	return cmd
}

// listPermissionResourceCMD query chain config permission resource list
// @return *cobra.Command
func listPermissionResourceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list chain config permission resource",
		Long:  "list chain config permission resource",
		RunE: func(_ *cobra.Command, _ []string) error {
			return doPermissionResourceOperation(listPermissionResourceEnum)
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagOrgId,
		flagChainId, flagTimeout, flagSyncResult, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})

	return cmd
}

func doPermissionResourceOperation(crud int) error {
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
	switch crud {
	case addPermissionResourceEnum:
		policy := &accesscontrol.Policy{
			Rule:     permissionResourcePolicyRule,
			OrgList:  permissionResourcePolicyOrgList,
			RoleList: permissionResourcePolicyRoleList,
		}
		payload, err = client.CreateChainConfigPermissionAddPayload(permissionResourceName, policy)
		if err != nil {
			return err
		}
	case updatePermissionResourceEnum:
		policy := &accesscontrol.Policy{
			Rule:     permissionResourcePolicyRule,
			OrgList:  permissionResourcePolicyOrgList,
			RoleList: permissionResourcePolicyRoleList,
		}
		payload, err = client.CreateChainConfigPermissionUpdatePayload(permissionResourceName, policy)
		if err != nil {
			return err
		}
	case deletePermissionResourceEnum:
		payload, err = client.CreateChainConfigPermissionDeletePayload(permissionResourceName)
		if err != nil {
			return err
		}
	case listPermissionResourceEnum:
		permissionList, err := client.GetChainConfigPermissionList()
		if err != nil {
			return fmt.Errorf("get chain config failed, %s", err.Error())
		}

		output, err := prettyjson.Marshal(permissionList)
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	default:
		return errors.New("invalid permission resource operation")
	}

	endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}
	// send
	resp, err := client.SendChainConfigUpdateRequest(payload, endorsers, timeout, syncResult)
	if err != nil {
		return err
	}

	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}

	output, err := prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
