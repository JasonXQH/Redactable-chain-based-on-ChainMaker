// Copyright (C) BABEC. All rights reserved.
// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// newQueryBlockByHeightOnChainCMD `query block by block height` command implementation
func newQueryBlockByHeightOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-height [height]",
		Short: "query on-chain block by height, get last block if [height] not set",
		Long:  "query on-chain block by height, get last block if [height] not set",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var height uint64
			var err error
			if len(args) == 0 {
				height = math.MaxUint64
			} else {
				height, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return err
				}
			}
			//// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
				sdk.WithChainClientChainId(chainId),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			//// 2.Query block on-chain.
			truncateLength := 0
			if truncateValue {
				truncateLength = 1000
			}
			blkWithRWSetOnChain, err := cc.GetBlockByHeightTruncate(height, withRWSet, truncateLength, "truncate")
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(blkWithRWSetOnChain)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagTruncateValue, flagWithRWSet, flagChainId, flagSdkConfPath,
	})
	return cmd
}

// newQueryBlockByHashOnChainCMD `query block by block hash` command implementation
func newQueryBlockByHashOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-hash [block hash in base64/hex]",
		Short: "query on-chain block by hash",
		Long:  "query on-chain block by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
				sdk.WithChainClientChainId(chainId),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			//// 2.Query block on-chain.
			var blkHashStr = args[0]
			var blkHashBz []byte
			if blkHashBz, err = hex.DecodeString(blkHashStr); err != nil {
				if blkHashBz, err = base64.StdEncoding.DecodeString(blkHashStr); err != nil {
					return errors.New("invalid block hash, block hash must be base64 or hex encoding")
				}
			}
			height, err := cc.GetBlockHeightByHash(hex.EncodeToString(blkHashBz))
			if err != nil {
				return err
			}
			blkWithRWSetOnChain, err := cc.GetBlockByHeight(height, true)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(blkWithRWSetOnChain)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagChainId, flagSdkConfPath,
	})
	return cmd
}

// newQueryBlockByTxIdOnChainCMD `query block by txid` command implementation
func newQueryBlockByTxIdOnChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-txid [txid]",
		Short: "query on-chain block by txid",
		Long:  "query on-chain block by txid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			//// 1.Chain Client
			cc, err := sdk.NewChainClient(
				sdk.WithConfPath(sdkConfPath),
				sdk.WithChainClientChainId(chainId),
			)
			if err != nil {
				return err
			}
			defer cc.Stop()
			if err := util.DealChainClientCertHash(cc, enableCertHash); err != nil {
				return err
			}

			//// 2.Query block on-chain.
			height, err := cc.GetBlockHeightByTxId(args[0])
			if err != nil {
				return err
			}
			blkWithRWSetOnChain, err := cc.GetFullBlockByHeight(height)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(blkWithRWSetOnChain)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagChainId, flagSdkConfPath,
	})
	return cmd
}
