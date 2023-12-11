/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/hex"
	"fmt"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/evmutils"
	"github.com/spf13/cobra"
)

// contractCMD contract command
// @return *cobra.Command
func contractCMD() *cobra.Command {
	contractCmd := &cobra.Command{
		Use:   "contract",
		Short: "contract command",
		Long:  "contract command",
	}

	contractCmd.AddCommand(userContractCMD())
	contractCmd.AddCommand(systemContractCMD())
	contractCmd.AddCommand(contractNameToAddrCMD())

	return contractCmd
}

// contractNameToAddrCMD calculate contract name to address
// @return *cobra.Command
func contractNameToAddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "name-to-addr [contract name]",
		Short: "contract name to address command",
		Long:  "contract name to address command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var addr string
			var err error
			var contractName = args[0]
			switch addressType {
			case 0:
				addrInt, err := evmutils.MakeAddressFromString(contractName)
				if err != nil {
					return err
				}
				addrObj := evmutils.BigToAddress(addrInt)
				addr = hex.EncodeToString(addrObj[:])
			case 1:
				addr, err = evmutils.ZXAddress([]byte(contractName))
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported address type %v", addressType)
			}

			util.PrintPrettyJson(addr)
			return nil
		},
	}
	attachFlags(cmd, []string{flagAddressType})
	return cmd
}
