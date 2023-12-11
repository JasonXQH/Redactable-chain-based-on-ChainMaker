/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package archive

import (
	"errors"
	"fmt"
	"strconv"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

func newRestoreCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore [restore end height]",
		Short: "restore blockchain data",
		Long:  "restore archive center data to chain storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(sdkConfPath) == 0 {
				return fmt.Errorf(" %s must not be empty",
					flagSdkConfPath)
			}
			restoreHeight, parseErr := strconv.ParseUint(args[0], 10, 64)
			if parseErr != nil {
				msg := fmt.Sprintf("args should be integer ,now is [%s], parse error [%s]",
					args[0], parseErr.Error())
				return errors.New(msg)
			}
			return runRestoreCMD(restoreHeight)
		},
	}
	util.AttachAndRequiredFlags(cmd, flags,
		[]string{
			flagSdkConfPath,
			flagTimeOut,
		})
	util.AttachFlags(cmd, flags, []string{flagBatch})
	return cmd
}

func runRestoreCMD(restoreEndHeight uint64) error {
	chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
		"", "", "", "", "")
	if chainClientErr != nil {
		return chainClientErr
	}
	defer chainClient.Stop()
	progress := uiprogress.New()
	var bar *uiprogress.Bar
	init := false
	err := chainClient.RestoreBlocks(restoreEndHeight, "", func(msg sdk.ProcessMessage) error {
		if msg.Error != nil {
			fmt.Println(msg.Error.Error())
			return msg.Error
		}
		if !init {
			bar = progress.AddBar(int(msg.Total)).AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Restoring Blocks (%d/%d)\n", b.Current(), msg.Total)
			})
			progress.Start()
			init = true
		}
		bar.Incr()
		return nil
	})
	progress.Stop()
	if err != nil {
		return err
	}
	return nil
}
