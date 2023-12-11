package commandutil

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/spf13/cobra"
)

func newBase64ToHexCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base64tohex [base64Str]",
		Short: "base64tohex transfer base64-encode-data to hex-encode-data",
		Long:  "base64tohex transfer base64-encode-data to hex-encode-data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			base64Str := args[0]
			hashBytes, decodeErr := base64.StdEncoding.DecodeString(base64Str)
			if decodeErr != nil {
				return fmt.Errorf("decode base64 [%s] got error [%s]", base64Str, decodeErr.Error())
			}
			outStr := hex.EncodeToString(hashBytes)
			fmt.Println(outStr)
			return nil
		},
	}
	util.AttachAndRequiredFlags(cmd, flags, []string{})
	return cmd
}
