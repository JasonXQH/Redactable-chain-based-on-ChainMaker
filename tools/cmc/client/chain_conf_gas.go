/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// enableOrDisableGasCMD enable or disable gas feature
// @return *cobra.Command
func enableOrDisableGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas",
		Short: "enable or disable gas feature",
		Long:  "enable or disable gas feature",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(cc, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
			if err != nil {
				return err
			}

			//// 2.Enable or disable gas feature
			chainConfig, err := cc.GetChainConfig()
			if err != nil {
				return err
			}
			isGasEnabled := false
			if chainConfig.AccountConfig != nil {
				isGasEnabled = chainConfig.AccountConfig.EnableGas
			}

			if (gasEnable && !isGasEnabled) || (!gasEnable && isGasEnabled) {
				payload, err := cc.CreateChainConfigEnableOrDisableGasPayload()
				if err != nil {
					return err
				}
				endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, cc, payload)
				if err != nil {
					return err
				}

				resp, err := cc.SendChainConfigUpdateRequest(payload, endorsers, -1, syncResult)
				if err != nil {
					return err
				}

				err = util.CheckProposalRequestResp(resp, false)
				if err != nil {
					return err
				}
			}

			output, err := prettyjson.Marshal("OK")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult,
	})

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagGasEnable,
	})
	return cmd
}
