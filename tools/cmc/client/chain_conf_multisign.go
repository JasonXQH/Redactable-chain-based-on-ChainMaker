package client

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// enableMultiSignManualRunCMD enable or disable manual run feature of multi-sign
// @return *cobra.Command
func enableMultiSignManualRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable-manual-run",
		Short: "enable or disable manual_run feature of multi-sign",
		Long:  "enable or disable manual_run feature of multi-sign",
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

			if chainConfig.Vm == nil {
				return errors.New("the Vm section of chain config is nil")

			} else if chainConfig.Vm.Native == nil {
				chainConfig.Vm.Native = &configPb.VmNative{
					Multisign: &configPb.MultiSign{
						EnableManualRun: multiSignEnableManualRun,
					},
				}
			} else if chainConfig.Vm.Native.Multisign == nil {
				chainConfig.Vm.Native.Multisign = &configPb.MultiSign{
					EnableManualRun: multiSignEnableManualRun,
				}
			} else {
				chainConfig.Vm.Native.Multisign.EnableManualRun = multiSignEnableManualRun
			}

			payload, err := cc.CreateChainConfigEnableMultiSignManualRunPayload(multiSignEnableManualRun)
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

			output, err := prettyjson.Marshal("OK")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult,
		flagSdkConfPath, flagMultiSignEnableManualRun,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagMultiSignEnableManualRun)

	return cmd
}
