/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0

This file is for version compatibility
*/

package common

import (
	"fmt"

	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
)

// NotifyChainConf Notify each module of callback before version v2.3.0
func NotifyChainConf(block *commonpb.Block, chainConf protocol.ChainConf) (err error) {
	if block != nil && block.GetTxs() != nil && len(block.GetTxs()) > 0 {
		if ok, _ := utils.IsNativeTx(block.GetTxs()[0]); ok || utils.HasDPosTxWritesInHeader(block, chainConf) {
			if err = chainConf.CompleteBlock(block); err != nil { //nolint: staticcheck
				return fmt.Errorf("chainconf block complete, %s", err)
			}
		}
	}
	return nil
}
