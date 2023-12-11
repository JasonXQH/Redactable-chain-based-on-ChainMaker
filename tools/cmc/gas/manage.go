// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package gas

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	configpb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// newSetGasAdminCMD set gas admin, set self as an admin if [address] not set
// @return *cobra.Command
func newSetGasAdminCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-admin [address]",
		Short: "set gas admin, set self as an admin if [address] not set",
		Long:  "set gas admin, set self as an admin if [address] not set",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Set gas admin
			adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(cc, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
			if err != nil {
				return err
			}

			var adminAddr string
			if len(args) == 0 {
				adminAddr, err = getSelfAddress(cc)
				if err != nil {
					return err
				}
			} else {
				adminAddr = args[0]
			}

			payload, err := cc.CreateSetGasAdminPayload(adminAddr)
			if err != nil {
				return err
			}
			endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, cc, payload)
			if err != nil {
				return err
			}

			resp, err := cc.SendGasManageRequest(payload, endorsers, -1, syncResult)
			if err != nil {
				return err
			}

			util.PrintPrettyJson(struct {
				*common.TxResponse
				GasAdminAddress string `json:"gas_admin_address"`
			}{
				resp,
				adminAddr,
			})
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult, flagSdkConfPath,
	})
	return cmd
}

// newGetGasAdminCMD get gas admin
// @return *cobra.Command
func newGetGasAdminCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-admin",
		Short: "get gas admin",
		Long:  "get gas admin",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Get gas admin
			addr, err := cc.GetGasAdmin()
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(addr)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

// newRechargeGasCMD recharge gas for account
// @return *cobra.Command
func newRechargeGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recharge",
		Short: "recharge gas for account",
		Long:  "recharge gas for account",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Recharge gas
			rechargeGasList := []*syscontract.RechargeGas{
				{
					Address:   address,
					GasAmount: amount,
				},
			}
			payload, err := cc.CreateRechargeGasPayload(rechargeGasList)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAddress, flagAmount,
	})
	return cmd
}

// newGetGasBalanceCMD get gas balance of account, get self balance if --address not set
// @return *cobra.Command
func newGetGasBalanceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-balance",
		Short: "get gas balance of account, get self balance if --address not set",
		Long:  "get gas balance of account, get self balance if --address not set",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Get gas balance
			if address == "" {
				address, err = getSelfAddress(cc)
				if err != nil {
					return err
				}
			}

			balance, err := cc.GetGasBalance(address)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(balance)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAddress, flagSdkConfPath,
	})
	return cmd
}

// newRefundGasCMD refund gas for account
// @return *cobra.Command
func newRefundGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund",
		Short: "refund gas for account",
		Long:  "refund gas for account",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Refund gas
			payload, err := cc.CreateRefundGasPayload(address, amount)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAddress, flagAmount,
	})
	return cmd
}

// newFrozenGasAccountCMD frozen gas account
// @return *cobra.Command
func newFrozenGasAccountCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen",
		Short: "frozen gas account",
		Long:  "frozen gas account",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Frozen gas account
			payload, err := cc.CreateFrozenGasAccountPayload(address)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAddress,
	})
	return cmd
}

// newUnfrozenGasAccountCMD unfrozen gas account
// @return *cobra.Command
func newUnfrozenGasAccountCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfrozen",
		Short: "unfrozen gas account",
		Long:  "unfrozen gas account",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Unfrozen gas account
			payload, err := cc.CreateUnfrozenGasAccountPayload(address)
			if err != nil {
				return err
			}
			resp, err := cc.SendGasManageRequest(payload, nil, -1, syncResult)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAddress,
	})
	return cmd
}

// newGetGasAccountStatusCMD get gas account status, get self gas account status if --address not set
// @return *cobra.Command
func newGetGasAccountStatusCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-status",
		Short: "get gas account status, get self gas account status if --address not set",
		Long:  "get gas account status, get self gas account status if --address not set",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Get gas account status
			if address == "" {
				address, err = getSelfAddress(cc)
				if err != nil {
					return err
				}
			}

			status, err := cc.GetGasAccountStatus(address)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(status)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAddress, flagSdkConfPath,
	})
	return cmd
}

func getSelfAddress(cc *sdk.ChainClient) (string, error) {
	chainconf, err := cc.GetChainConfig()
	if err != nil {
		return "", err
	}
	pk, err := cc.GetPublicKey().String()
	if err != nil {
		return "", err
	}
	switch chainconf.Vm.AddrType {
	case configpb.AddrType_CHAINMAKER:
		address, err = sdk.GetCMAddressFromPKPEM(pk, cc.GetHashType())
		if err != nil {
			return "", err
		}
	case configpb.AddrType_ZXL:
		address, err = sdk.GetZXAddressFromPKPEM(pk, cc.GetHashType())
		if err != nil {
			return "", err
		}
	case configpb.AddrType_ETHEREUM:
		address, err = sdk.GetEVMAddressFromPKPEM(pk, cc.GetHashType())
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("unknown address type")
	}
	return address, nil
}

// newSetInvokeBaseGasCMD set VM base gas of invoke tx
// @return *cobra.Command
func newSetInvokeBaseGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-base-gas",
		Short: "set VM base gas of invoke tx",
		Long:  "set VM base gas of invoke tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Set base gas
			adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(cc, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
			if err != nil {
				return err
			}

			payload, err := cc.CreateSetInvokeBaseGasPayload(amount)
			if err != nil {
				return err
			}
			endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, cc, payload)
			if err != nil {
				return err
			}

			resp, err := cc.SendGasManageRequest(payload, endorsers, -1, syncResult)
			if err != nil {
				return err
			}

			util.PrintPrettyJson(resp)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAmount,
	})
	return cmd
}

// newSetInvokeGasPriceCMD set VM base gas of invoke tx
// @return *cobra.Command
func newSetInvokeGasPriceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-invoke-gas-price",
		Short: "set VM base gas price of invoke tx",
		Long:  "set VM base gas price of invoke tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Set base gas
			adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(cc, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
			if err != nil {
				return err
			}

			payload, err := cc.CreateSetInvokeGasPricePayload(price)
			if err != nil {
				return err
			}
			endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, cc, payload)
			if err != nil {
				return err
			}

			resp, err := cc.SendGasManageRequest(payload, endorsers, -1, syncResult)
			if err != nil {
				return err
			}

			util.PrintPrettyJson(resp)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagPrice,
	})
	return cmd
}

// newSetInstallBaseGasCMD enable or disable gas feature
// @return *cobra.Command
func newSetInstallBaseGasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-install-base-gas",
		Short: "set install base gas",
		Long:  "set install base gas",
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

			if !isGasEnabled {
				return errors.New("`enable_gas` flag is not set")
			}

			payload, err := cc.CreateSetInstallBaseGasPayload(amount)
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

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAmount,
	})
	return cmd
}

// newSetInstallGasPriceCMD set VM base gas of invoke tx
// @return *cobra.Command
func newSetInstallGasPriceCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-install-gas-price",
		Short: "set VM base gas price of install|upgrade tx",
		Long:  "set VM base gas price of install|upgrade tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			//// 2.Set base gas
			adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(cc, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
			if err != nil {
				return err
			}

			payload, err := cc.CreateSetInstallGasPricePayload(price)
			if err != nil {
				return err
			}
			endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, cc, payload)
			if err != nil {
				return err
			}

			resp, err := cc.SendGasManageRequest(payload, endorsers, -1, syncResult)
			if err != nil {
				return err
			}

			util.PrintPrettyJson(resp)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagAdminKeyFilePaths, flagAdminCrtFilePaths, flagAdminOrgIds, flagSyncResult, flagSdkConfPath,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagPrice,
	})
	return cmd
}
