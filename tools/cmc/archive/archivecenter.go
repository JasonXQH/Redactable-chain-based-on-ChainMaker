/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package archive

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sdkConfPath string // sdk-go的配置信息
	//archiveHeight uint64 // 归档到区块
	archiveMode string // 是否使用快速归档(是的话为quick默认为否)
	flags       *pflag.FlagSet
	timeOut     uint64 // rpc time out in seconds
	// 如果设置为y,就不校验高度参数.
	// 否则只有传递的高度参数小于归档中心的已经归档高度时候,才可以执行清理数据的操作
	// 默认为string
	ignoreCleanArchivedHeight string
	batchCount                uint64 // restore时候设置批量值
)

const (
	flagSdkConfPath = "sdk-conf-path"
	//flagArchiveHeight = "archive-height"
	flagArchiveMode               = "mode"
	flagTimeOut                   = "timeout"
	flagCleanIgnoreArchivedHeight = "ignore"
	flagBatch                     = "batch"
)

func NewArchiveCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "archive blockchain data",
		Long:  "archive blockchain data ",
	}
	cmd.AddCommand(newDumpCMD())
	cmd.AddCommand(newQueryCMD())
	cmd.AddCommand(newArchiveCMD())
	cmd.AddCommand(newRestoreCMD())
	return cmd
}

func init() {
	flags = &pflag.FlagSet{}
	flags.StringVar(&sdkConfPath,
		flagSdkConfPath, "", "specify sdk config path")
	//flags.Uint64Var(&archiveHeight, flagArchiveHeight,
	//	0, "specify archive end block height")
	flags.StringVar(&archiveMode, flagArchiveMode, "", "specify archive mode ,can be quick or normal ")
	flags.Uint64Var(&timeOut,
		flagTimeOut, 10, "specify clean bfdb timeout")
	flags.StringVar(&ignoreCleanArchivedHeight, flagCleanIgnoreArchivedHeight, "n",
		"specify clean chain data whether clean-height's data has been stored in archivecenter")
	flags.Uint64Var(&batchCount, flagBatch, 1,
		"specify restore batch block counts")
}
