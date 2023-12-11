// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package gas

import (
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sdkConfPath       string
	chainId           string
	syncResult        bool
	adminKeyFilePaths string
	adminCrtFilePaths string
	adminOrgIds       string
	address           string
	amount            int64
	price             string
)

const (
	flagSdkConfPath       = "sdk-conf-path"
	flagChainId           = "chain-id"
	flagSyncResult        = "sync-result"
	flagAdminKeyFilePaths = "admin-key-file-paths"
	flagAdminCrtFilePaths = "admin-crt-file-paths"
	flagAdminOrgIds       = "admin-org-ids"
	flagAddress           = "address"
	flagAmount            = "amount"
	flagPrice             = "price"
)

// NewGasManageCMD new gas management command
func NewGasManageCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas",
		Short: "gas management",
		Long:  "gas management",
	}

	cmd.AddCommand(newSetGasAdminCMD())
	cmd.AddCommand(newGetGasAdminCMD())
	cmd.AddCommand(newRechargeGasCMD())
	cmd.AddCommand(newGetGasBalanceCMD())
	cmd.AddCommand(newRefundGasCMD())
	cmd.AddCommand(newFrozenGasAccountCMD())
	cmd.AddCommand(newUnfrozenGasAccountCMD())
	cmd.AddCommand(newGetGasAccountStatusCMD())
	cmd.AddCommand(newSetInvokeBaseGasCMD())
	cmd.AddCommand(newSetInvokeGasPriceCMD())
	cmd.AddCommand(newSetInstallBaseGasCMD())
	cmd.AddCommand(newSetInstallGasPriceCMD())

	return cmd
}

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}

	flags.StringVar(&chainId, flagChainId, "", "Chain ID")
	flags.StringVar(&sdkConfPath, flagSdkConfPath, "", "specify sdk config path")
	flags.BoolVar(&syncResult, flagSyncResult, false, "whether wait the result of the transaction, default false")
	flags.StringVar(&adminKeyFilePaths, flagAdminKeyFilePaths, "", "specify admin key file paths, use ',' to separate")
	flags.StringVar(&adminCrtFilePaths, flagAdminCrtFilePaths, "", "specify admin cert file paths, use ',' to separate")
	flags.StringVar(&adminOrgIds, flagAdminOrgIds, "", "specify admin org-ids, use ',' to separate")
	flags.StringVar(&address, flagAddress, "", "address of account")
	flags.Int64Var(&amount, flagAmount, 0, "amount of gas")
	flags.StringVar(&price, flagPrice, "0", "price of one byte")

	if sdkConfPath == "" {
		sdkConfPath = util.EnvSdkConfPath
	}
}
