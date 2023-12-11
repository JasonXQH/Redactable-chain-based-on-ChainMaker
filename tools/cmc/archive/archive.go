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
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/spf13/cobra"
)

func newArchiveCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [archive block height]",
		Short: "archive block file under height",
		Long:  "archive block file under height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cleanHeight, parseErr := strconv.ParseUint(args[0], 10, 64)
			if parseErr != nil {
				msg := fmt.Sprintf("args should be integer ,now is [%s], parse error [%s]",
					args[0], parseErr.Error())
				return errors.New(msg)
			}
			return runArchiveCMD(cleanHeight)

		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath, flagTimeOut,
	})
	util.AttachFlags(cmd, flags, []string{flagCleanIgnoreArchivedHeight})

	return cmd
}

func runArchiveCMD(cleanHeight uint64) error {
	// create chain client
	cc, err := util.CreateChainClient(sdkConfPath, "",
		"", "", "", "", "")
	if err != nil {
		return err
	}
	defer cc.Stop()
	// 在这里先先校验一下cleanheight是否小于归档中心的高度
	if strings.ToLower(strings.TrimSpace(ignoreCleanArchivedHeight)) != "y" {
		archiveClient := cc.GetArchiveService()
		archivedHeight, _, _, err1 := archiveClient.GetArchivedStatus()
		if err1 != nil {
			return err1
		}
		if cleanHeight > archivedHeight {
			return fmt.Errorf(
				"param height [%d] must not be greater than max height [%d] in archivecenter",
				cleanHeight, archivedHeight)
		}
	}
	// archivedBlkHeightOnChain, queryArchivedErr := cc.GetArchivedBlockHeight()
	// if queryArchivedErr != nil {
	// 	fmt.Printf("query GetArchivedBlockHeight error %s", queryArchivedErr.Error())
	// } else {
	// 	fmt.Printf("query GetArchivedBlockHeight  %d \n", archivedBlkHeightOnChain)
	// }
	payload, payloadErr := cc.CreateArchiveBlockPayload(cleanHeight)
	if payloadErr != nil {
		return payloadErr
	}

	resp, respErr := cc.SendArchiveBlockRequest(payload, int64(timeOut))
	if respErr != nil {
		return fmt.Errorf(
			"send archive request error [%s],call GetArchivedBlockHeight for archivedHeight on chain",
			respErr.Error())
	}

	checkResult := util.CheckProposalRequestResp(resp, false)
	if checkResult != nil {
		return fmt.Errorf("runArchiveCMD  failed error %s , resp message: %#v", checkResult.Error(), *resp)
	}
	fmt.Println("chain has got archive command!")
	return nil
}
