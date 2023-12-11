/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package archive

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/store"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

var (
	errorNoBlock = errors.New("query got no block")
)

func newQueryCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query off-chain blockchain data",
		Long:  "query off-chain blockchain data",
	}
	cmd.AddCommand(newQueryBlockByTxIdCMD())
	cmd.AddCommand(newQueryTxOffChainCMD())
	cmd.AddCommand(newQueryBlockByHashCMD())
	cmd.AddCommand(newQueryArchiveStatus())
	cmd.AddCommand(newQueryBlockByHeightCMD())
	cmd.AddCommand(newQueryArchiveStatusOnChain())
	return cmd
}

func newQueryTxOffChainCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [txid]",
		Short: "query off-chain tx by txid",
		Long:  "query off-chain tx by txid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			txInfo, err := chainClient.GetArchiveService().GetTxByTxId(args[0])

			if err != nil {
				fmt.Println("query off-chain data got no block")
				return errorNoBlock
			}
			output, err := prettyjson.Marshal(txInfo)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func newQueryBlockByHeightCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-height [height]",
		Short: "query off-chain block by height",
		Long:  "query off-chain block by height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			height, transferErr := strconv.ParseUint(args[0], 10, 64)
			if transferErr != nil {
				return fmt.Errorf("height must be positive integer , but got %s", args[0])
			}
			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			blockInfo, err := chainClient.GetArchiveService().GetBlockByHeight(height, true)

			if err != nil {
				fmt.Println("query off-chain data got no block")
				return errorNoBlock
			}
			output, err := prettyjson.Marshal(blockInfo)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func newQueryBlockByHashCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-hash [block hash in hex]",
		Short: "query off-chain block by hash",
		Long:  "query off-chain block by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hashStr, hashErr := transferHashToHexString(args[0])
			if hashErr != nil {
				return fmt.Errorf("transfer block hash %s got error %s",
					args[0], hashErr.Error())
			}
			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			blockInfo, err := chainClient.GetArchiveService().GetBlockByHash(hashStr, true)
			if err != nil {
				fmt.Println("query off-chain data got no block")
				return errorNoBlock
			}
			output, err := prettyjson.Marshal(blockInfo)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func transferHashToHexString(hashStr string) (string, error) {
	if strings.Contains(hashStr, "+") ||
		strings.Contains(hashStr, "=") ||
		strings.Contains(hashStr, "/") {
		hashBytes, decodeErr := base64.StdEncoding.DecodeString(hashStr)
		if decodeErr != nil {
			return hashStr, decodeErr
		}
		return hex.EncodeToString(hashBytes), nil
	}
	return hashStr, nil
}

func newQueryBlockByTxIdCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-txid [txid]",
		Short: "query off-chain block by txid",
		Long:  "query off-chain block by txid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			blockInfo, err := chainClient.GetArchiveService().GetBlockByTxId(args[0], true)
			if err != nil {
				fmt.Println("query off-chain data got no block")
				return errorNoBlock
			}
			output, err := prettyjson.Marshal(blockInfo)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

type OutPutBlockHeader struct {
	BlockHeader *common.BlockHeader `json:"block_header"`
	BlockHash   string              `json:"block_hash"`
}
type OutPutBlock struct {
	Block  *common.Block      `json:"block"`
	Header *OutPutBlockHeader `json:"header"`
}

type OutPutBlockWithRWSet struct {
	BlockWithRWSet *store.BlockWithRWSet
	Block          *OutPutBlock
}

func newQueryArchiveStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archived-status",
		Short: "query off-chain archived status",
		Long:  "query off-chain archived status",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			archivedHeight, inArchive, code, err := chainClient.GetArchiveService().GetArchivedStatus()
			if err != nil {
				fmt.Println("query off-chain data got no block")
				return err
			}
			if code != 0 {
				return fmt.Errorf("archived-status rpc code %d", code)
			}
			fmt.Printf("archived-status height %d , inArchive %t \n",
				archivedHeight, inArchive)
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func newQueryArchiveStatusOnChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-archived-status",
		Short: "query on-chain archived status",
		Long:  "query on-chain archived status",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {

			chainClient, chainClientErr := util.CreateChainClient(sdkConfPath, "",
				"", "", "", "", "")
			if chainClientErr != nil {
				return chainClientErr
			}
			defer chainClient.Stop()
			archiveStatus, archiveStatusErr := chainClient.GetArchiveStatus()
			if archiveStatusErr != nil {
				return fmt.Errorf("chainClient.GetArchiveStatus got error: %s", archiveStatusErr.Error())
			}
			if archiveStatus == nil {
				return fmt.Errorf("chainClient.GetArchiveStatus got nothing")
			}
			var infos strings.Builder
			infos.WriteString(
				fmt.Sprintf("ArchivePivot: %d, MaxAllowArchiveHeight: %d . %#v",
					archiveStatus.ArchivePivot, archiveStatus.MaxAllowArchiveHeight, archiveStatus))
			for i := 0; i < len(archiveStatus.FileRanges); i++ {
				infos.WriteString(
					fmt.Sprintf("fileName: %s, start: %d, end: %d ; ",
						archiveStatus.FileRanges[i].FileName, archiveStatus.FileRanges[i].Start,
						archiveStatus.FileRanges[i].End))
			}
			fmt.Println(infos.String())
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}
