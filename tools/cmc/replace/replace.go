// Copyright (C) BABEC. All rights reserved.
// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

// Package query query block chain
package replace

import (
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sdkConfPath    string
	chainId        string
	enableCertHash bool
	withRWSet      bool
	truncateValue  bool
)

const (
	flagSdkConfPath    = "sdk-conf-path"
	flagChainId        = "chain-id"
	flagEnableCertHash = "enable-cert-hash"
	flagWithRWSet      = "with-rw-set"
	flagTruncateValue  = "truncate-value"
)

// NewQueryOnChainCMD new query on-chain blockchain data command
func NewReplaceOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace",
		Short: "replace on-chain blockchain data",
		Long:  "replace on-chain blockchain data",
	}

	cmd.AddCommand(newReplaceBlockByHeightCMD())

	return cmd
}

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}

	flags.StringVar(&chainId, flagChainId, "", "Chain ID")
	flags.StringVar(&sdkConfPath, flagSdkConfPath, "", "specify sdk config path")
	flags.BoolVar(&enableCertHash, flagEnableCertHash, true, "whether enable cert hash")
	flags.BoolVar(&withRWSet, flagWithRWSet, true, "whether with RWSet")
	flags.BoolVar(&truncateValue, flagTruncateValue, true, "enable truncate value, default true")

	if sdkConfPath == "" {
		sdkConfPath = util.EnvSdkConfPath
	}
}
