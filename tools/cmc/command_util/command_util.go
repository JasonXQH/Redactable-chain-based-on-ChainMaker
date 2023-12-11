package commandutil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// 与业务无关的util命令
var (
	flags *pflag.FlagSet
)

func NewUtilCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "util",
		Short: "util",
		Long:  "util commands ",
	}
	cmd.AddCommand(newBase64ToHexCMD())

	return cmd
}

func init() {
	flags = &pflag.FlagSet{}
}
