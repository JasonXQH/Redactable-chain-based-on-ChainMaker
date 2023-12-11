/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"
	"io/ioutil"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/spf13/cobra"
)

const (
	addTrustRoot = iota
	removeTrustRoot
	updateTrustRoot
)

// configTrustRootCMD trust root command
// @return *cobra.Command
func configTrustRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trustroot",
		Short: "trust root command",
		Long:  "trust root command",
	}
	cmd.AddCommand(addTrustRootCMD())
	cmd.AddCommand(removeTrustRootCMD())
	cmd.AddCommand(updateTrustRootCMD())

	return cmd
}

// addTrustRootCMD add trust root ca cert
// @return *cobra.Command
func addTrustRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add trust root ca cert",
		Long:  "add trust root ca cert",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configTrustRoot(addTrustRoot)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagTrustRootCrtPath, flagTrustRootOrgId,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagTrustRootOrgId)
	cmd.MarkFlagRequired(flagTrustRootCrtPath)

	return cmd
}

// removeTrustRootCMD remove trust root ca cert
// @return *cobra.Command
func removeTrustRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "remove trust root ca cert",
		Long:  "remove trust root ca cert",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configTrustRoot(removeTrustRoot)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagTrustRootOrgId,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagTrustRootOrgId)

	return cmd
}

// updateTrustRootCMD update trust root ca cert
// @return *cobra.Command
func updateTrustRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update trust root ca cert",
		Long:  "update trust root ca cert",
		RunE: func(_ *cobra.Command, _ []string) error {
			return configTrustRoot(updateTrustRoot)
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagSdkConfPath, flagOrgId, flagEnableCertHash, flagTrustRootCrtPath, flagTrustRootOrgId,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath,
	})

	cmd.MarkFlagRequired(flagTrustRootOrgId)
	cmd.MarkFlagRequired(flagTrustRootCrtPath)

	return cmd
}

func configTrustRoot(op int) error {
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

	var trustRootBytes []string
	if op == addTrustRoot || op == updateTrustRoot {

		if len(trustRootPaths) == 0 {
			return fmt.Errorf("please specify trust root path")
		}
		for _, trustRootPath := range trustRootPaths {
			trustRoot, err := ioutil.ReadFile(trustRootPath)
			if err != nil {
				return err
			}
			trustRootBytes = append(trustRootBytes, string(trustRoot))
		}
	}

	var payload *common.Payload
	switch op {
	case addTrustRoot:
		payload, err = client.CreateChainConfigTrustRootAddPayload(trustRootOrgId, trustRootBytes)
	case removeTrustRoot:
		payload, err = client.CreateChainConfigTrustRootDeletePayload(trustRootOrgId)
	case updateTrustRoot:
		payload, err = client.CreateChainConfigTrustRootUpdatePayload(trustRootOrgId, trustRootBytes)
	default:
		err = errors.New("invalid trust root operation")
	}
	if err != nil {
		return err
	}

	endorsementEntrys := make([]*common.EndorsementEntry, len(adminKeys))
	for i := range adminKeys {
		if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
			e, err := sdkutils.MakeEndorserWithPath(adminKeys[i], adminCrts[i], payload)
			if err != nil {
				return err
			}

			endorsementEntrys[i] = e
		} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
			e, err := sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				client.GetHashType(),
				adminOrgs[i],
				payload,
			)
			if err != nil {
				return err
			}

			endorsementEntrys[i] = e
		} else {
			e, err := sdkutils.MakePkEndorserWithPath(
				adminKeys[i],
				client.GetHashType(),
				"",
				payload,
			)
			if err != nil {
				return err
			}

			endorsementEntrys[i] = e
		}
	}

	resp, err := client.SendChainConfigUpdateRequest(payload, endorsementEntrys, -1, syncResult)
	if err != nil {
		return err
	}
	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}
	fmt.Printf("trustroot response %+v\n", resp)
	return nil
}
