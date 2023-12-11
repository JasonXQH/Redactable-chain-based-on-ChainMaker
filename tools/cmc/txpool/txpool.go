// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package txpool

import (
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/txpool"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sdkConfPath string
	txType      int32
	txStage     int32
	txIds       []string
)

const (
	flagSdkConfPath = "sdk-conf-path"
	flagType        = "type"
	flagStage       = "stage"
	flagTxIds       = "tx-ids"
)

// NewTxPoolCMD new txpool command
func NewTxPoolCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txpool",
		Short: "txpool command",
		Long:  "txpool command",
	}

	cmd.AddCommand(newGetPoolStatusCMD())
	cmd.AddCommand(newGetTxIdsByTypeAndStageCMD())
	cmd.AddCommand(newGetTxsInPoolByTxIdsCMD())
	return cmd
}

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}

	flags.StringVar(&sdkConfPath, flagSdkConfPath, "", "specify sdk config path")
	flags.Int32Var(&txType, flagType, 3, "tx type, config tx type:1, common tx type:2, all tx type:3")
	flags.Int32Var(&txStage, flagStage, 3, "tx stage, in queue stage:1, in pending stage:2, all stage:3")
	flags.StringSliceVar(&txIds, flagTxIds, []string{}, "tx id list. --tx-ids=\"abc,xyz\"")

	if sdkConfPath == "" {
		sdkConfPath = util.EnvSdkConfPath
	}
}

// newGetPoolStatusCMD get tx pool status
// @return *cobra.Command
func newGetPoolStatusCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "get tx pool status",
		Long:  "get tx pool status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			poolStatus, err := cc.GetPoolStatus()
			if err != nil {
				return err
			}

			util.PrintPrettyJson(poolStatus)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

// newGetTxIdsByTypeAndStageCMD get txids by type and stage
// @return *cobra.Command
func newGetTxIdsByTypeAndStageCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txids",
		Short: "get txids by type and stage",
		Long:  "get txids by type and stage",
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			txIds, err := cc.GetTxIdsByTypeAndStage(txpool.TxType(txType), txpool.TxStage(txStage))
			if err != nil {
				return err
			}

			util.PrintPrettyJson(txIds)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagType, flagStage, flagSdkConfPath,
	})
	return cmd
}

// newGetTxsInPoolByTxIdsCMD get txs in pool by tx ids
// @return *cobra.Command
func newGetTxsInPoolByTxIdsCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "get txs in pool by tx ids",
		Long:  "get txs in pool by tx ids",
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			txs, txIdsNotInPool, err := cc.GetTxsInPoolByTxIds(txIds)
			if err != nil {
				return err
			}

			util.PrintPrettyJson(struct {
				Txs   []*common.Transaction `json:"txs"`
				TxIds []string              `json:"tx_ids"`
			}{
				txs,
				txIdsNotInPool,
			})
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagTxIds, flagSdkConfPath,
	})
	return cmd
}
