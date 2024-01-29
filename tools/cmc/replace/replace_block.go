package replace

// replace_block.go

import (
	"chainmaker.org/chainmaker-go/tools/cmc/util"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"fmt"
	"github.com/spf13/cobra"
	"math"
	"strconv"
)

// newQueryBlockByHeightOnChainCMD `query block by block height` command implementation
func newReplaceBlockByHeightCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-by-height [height]",
		Short: "replace on-chain block by height, get last block if [height] not set",
		Long:  "replace on-chain block by height, get last block if [height] not set",
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

			// 2.Query block on-chain.

			_, err = cc.ReplaceBlockByHeight(height, withRWSet)
			//if err != nil {
			//	return err
			//}

			//output, err := prettyjson.Marshal(blkWithRWSetOnChain)
			//if err != nil {
			//	return err
			//}
			fmt.Println("replace block: ", height)
			return nil
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagEnableCertHash, flagTruncateValue, flagWithRWSet, flagChainId, flagSdkConfPath,
	})
	return cmd
}
