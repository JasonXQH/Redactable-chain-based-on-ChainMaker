/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0

This file is for version compatibility
*/

package verifier

import (
	chainConfConfig "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
)

var _ protocol.Watcher = (*BlockVerifierImpl)(nil)

func (v *BlockVerifierImpl) Module() string {
	return ModuleNameCore
}

func (v *BlockVerifierImpl) Watch(chainConfig *chainConfConfig.ChainConfig) error {
	v.chainConf.ChainConfig().Block = chainConfig.Block
	v.log.Infof("update chainconf,blockverify[%v]", v.chainConf.ChainConfig().Block)
	return nil
}
