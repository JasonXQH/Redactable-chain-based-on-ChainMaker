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

func newDumpCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump [archive height]",
		Short: "dump blockchain data",
		Long:  "dump blockchain data to archive center storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(sdkConfPath) == 0 {
				return fmt.Errorf(" %s must not be empty",
					flagSdkConfPath)
			}
			archiveHeight, parseErr := strconv.ParseUint(args[0], 10, 64)
			if parseErr != nil {
				msg := fmt.Sprintf("args should be integer ,now is [%s], parse error [%s]",
					args[0], parseErr.Error())
				return errors.New(msg)
			}
			return runDumpCMD(archiveHeight)
		},
	}
	util.AttachAndRequiredFlags(cmd, flags,
		[]string{
			flagSdkConfPath,
			flagArchiveMode,
		})
	return cmd
}

func runDumpCMD(archiveHeight uint64) error {
	// create chain client
	cc, err := util.CreateChainClient(sdkConfPath, "",
		"", "", "", "", "")
	if err != nil {
		return err
	}
	defer cc.Stop()
	progress := uiprogress.New()
	var bar *uiprogress.Bar
	init := false
	err = cc.ArchiveBlocks(archiveHeight, archiveMode, func(msg sdk.ProcessMessage) error {
		if msg.Error != nil {
			fmt.Printf("Archive to height[%d] get error:%s\n", msg.CurrentHeight, msg.Error.Error())
		}
		//fmt.Printf("Archiving block[%d], total %d need archive)\n", msg.CurrentHeight, msg.Total)
		if !init {
			bar = progress.AddBar(int(msg.Total)).AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("Archiving Blocks (%d/%d)\n", b.Current(), msg.Total)
			})
			progress.Start()
			init = true
		}
		bar.Incr()
		return nil
	})
	if err != nil {
		return err
	}
	//time.Sleep(time.Second)
	progress.Stop()

	return nil
}
