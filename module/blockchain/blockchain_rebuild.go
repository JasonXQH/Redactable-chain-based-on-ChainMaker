/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"fmt"
	"os"

	"chainmaker.org/chainmaker/localconf/v2"

	commonErrors "chainmaker.org/chainmaker/common/v2/errors"
)

// Start all the modules.
func (bc *Blockchain) RebuildDbs(needVerify bool) {
	fmt.Printf("###########################")
	fmt.Printf("###start rebuild-dbs....###")
	fmt.Printf("###########################")
	bc.log.Infof("###########################")
	bc.log.Infof("###start rebuild-dbs....###")
	bc.log.Infof("###########################")
	lastBlock, err1 := bc.oldStore.GetLastBlock()
	if err1 != nil {
		bc.log.Errorf("get lastblockerr(%s)", err1.Error())
	} else {
		bc.log.Infof("lastBlock=%d", lastBlock.Header.BlockHeight)
	}
	var i, height uint64
	var preHash []byte
	bHeight, _ := localconf.ChainMakerConfig.StorageConfig["rebuild_block_height"].(int)
	if bHeight <= 0 {
		bc.log.Warnf("error block_height!")
		height = lastBlock.GetHeader().BlockHeight
	} else {
		if uint64(bHeight) <= lastBlock.GetHeader().BlockHeight {
			height = uint64(bHeight)
		} else {
			height = lastBlock.GetHeader().BlockHeight
		}
	}
	for i = 1; i <= height; i++ {
		block, err2 := bc.oldStore.GetBlock(i)
		if err2 != nil {
			bc.log.Errorf("get block %d err(%s)", i, err2.Error())
		}
		bc.log.Debugf("block %d hash is %x", i, block.GetHeader().BlockHash)
		bc.log.Debugf("block %d prehash is %x", i, block.GetHeader().PreBlockHash)
		if preHash != nil && string(preHash) != string(block.GetHeader().PreBlockHash) {
			bc.log.Errorf("\npreHash=%x\nprehash=%x", []byte(preHash), block.GetHeader().PreBlockHash)
			bc.log.Errorf("\nError!!!!\n")
		} else {
			bc.log.Debugf("\npreHash=%x\nprehash=%x", []byte(preHash), block.GetHeader().PreBlockHash)
		}
		preHash = block.GetHeader().BlockHash

		var err3 error
		if needVerify {
			err3 = bc.coreEngine.GetBlockVerifier().VerifyBlock(block, -1)
		} else {
			blockRwSets, err31 := bc.oldStore.GetBlockWithRWSets(block.Header.BlockHeight)
			if err31 == nil {
				err3 = bc.coreEngine.GetBlockVerifier().VerifyBlockWithRwSets(
					blockRwSets.GetBlock(), blockRwSets.GetTxRWSets(), -1)
			}
		}

		if err3 != nil {
			if err3 == commonErrors.ErrBlockHadBeenCommited {
				bc.log.Errorf("the block: %d has been committed in the blockChainStore ", block.Header.BlockHeight)
			} else {
				fmt.Printf("block[%d] verify success.", block.Header.BlockHeight)
				bc.log.Infof("block[%d] verify success.", block.Header.BlockHeight)
			}
		} else {
			fmt.Printf("block[%d] verify success.", block.Header.BlockHeight)
			bc.log.Infof("block[%d] verify success.", block.Header.BlockHeight)
		}

		//time.Sleep(500*time.Millisecond)
		if err4 := bc.coreEngine.GetBlockCommitter().AddBlock(block); err4 != nil {
			if err4 == commonErrors.ErrBlockHadBeenCommited {
				bc.log.Errorf("the block: %d has been committed in the blockChainStore ", block.Header.BlockHeight)
			} else {
				fmt.Printf("block[%d] rebuild success.", block.Header.BlockHeight)
				bc.log.Infof("block[%d] rebuild success.", block.Header.BlockHeight)
			}

		} else {
			bc.log.Infof("block[%d] rebuild success.", block.Header.BlockHeight)
			fmt.Printf("block[%d] rebuild success.", block.Header.BlockHeight)
		}
		//time.Sleep(500 * time.Millisecond)

	}
	fmt.Printf("###########################")
	fmt.Printf("###rebuild-dbs finished!###")
	fmt.Printf("###########################")
	bc.log.Infof("###########################")
	bc.log.Infof("###rebuild-dbs finished!###")
	bc.log.Infof("###########################")
	bc.Stop()
	os.Exit(0)
}
