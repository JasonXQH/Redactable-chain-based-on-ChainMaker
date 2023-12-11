/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0

This file is for version compatibility
*/

package blockchain

import (
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
)

var _ protocol.Watcher = (*Blockchain)(nil)

// Module
func (bc *Blockchain) Module() string {
	return "BlockChain"
}

// Watch
func (bc *Blockchain) Watch(_ *configPb.ChainConfig) error {
	bc.log.Infof("[Blockchain] callback msg")
	if err := bc.Init(); err != nil {
		bc.log.Errorf("blockchain init failed when the configuration of blockchain updating, %s", err)
		return err
	}
	bc.StopOnRequirements()
	if err := bc.Start(); err != nil {
		bc.log.Errorf("blockchain start failed when the configuration of blockchain updating, %s", err)
		return err
	}
	return nil
}
