/*
Copyright (C) The ChainMaker Authors

SPDX-License-Identifier: Apache-2.0

@Description 该文件中实现了 合约相关命令行查询功能具体包括：
	1.按合约名字查询合约对象
	2.查询所有的合约列表
	3.查询被禁用的系统合约
*/

package query

import (
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// newQueryContractOnChainCMD query on-chain contract data
// @return *cobra.Command
func newQueryContractOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "query on-chain contract data",
		Long:  "query on-chain contract data",
	}

	cmd.AddCommand(newQueryContractByNameOnChainCMD())
	cmd.AddCommand(newQueryContractListOnChainCMD())
	cmd.AddCommand(newQueryDisableNativeContractListOnChainCMD())

	return cmd
}

// newQueryContractByNameOnChainCMD `query contract info by contract name` command implementation
func newQueryContractByNameOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [contract name]",
		Short: "query on-chain contract info by contract name",
		Long:  "query on-chain contract info by contract name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			// 2.Query contracct info on-chain
			var contractInfo interface{}
			contractInfo, err = cc.GetContractInfo(args[0])
			if err != nil {
				return err
			}
			output, err := prettyjson.Marshal(contractInfo)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagSdkConfPath,
	})
	return cmd
}

// newQueryContractListOnChainCMD `query contract list` command implementation
func newQueryContractListOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "query on-chain contract list",
		Long:  "query on-chain contract list",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			// 2.Query contracct list on-chain
			var contractList interface{}
			contractList, err = cc.GetContractList()
			if err != nil {
				return err
			}
			output, err := prettyjson.Marshal(contractList)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagSdkConfPath,
	})
	return cmd
}

// newQueryDisableNativeContractListOnChainCMD `query disable native contract list` command implementation
func newQueryDisableNativeContractListOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "query on-chain disable native contract list",
		Long:  "query on-chain disable native contract list",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			// 2.Query contracct list on-chain
			contractList, err := cc.GetDisabledNativeContractList()
			if err != nil {
				return err
			}
			if contractList == nil {
				contractList = []string{}
			}
			output, err := prettyjson.Marshal(contractList)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagSdkConfPath,
	})
	return cmd
}
