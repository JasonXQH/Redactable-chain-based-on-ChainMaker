/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package key

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	"github.com/spf13/cobra"
)

var (
	keyPath string
)

// exportPublickeyCMD export the public key of the specified private key.
// @return *cobra.Command
func exportPublickeyCMD() *cobra.Command {
	genCmd := &cobra.Command{
		Use:   "export_pub",
		Short: "public key export",
		Long: strings.TrimSpace(
			`Export the public key of the specified private key.
Example:
$ cmc key export_pub -p ./ -n ca.pem -key ca.key
`,
		),
		RunE: func(_ *cobra.Command, _ []string) error {
			return exportPublicKey()
		},
	}

	flags := genCmd.Flags()
	flags.StringVarP(&keyPath, "key", "k", "", "specify private key")
	flags.StringVarP(&path, "path", "p", "", "specify public key storage path")
	flags.StringVarP(&name, "name", "n", "", "specify public key storage file name")

	return genCmd
}

func exportPublicKey() error {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	pri, err := asym.PrivateKeyFromPEM(key, nil)
	if err != nil {
		return err
	}
	pubPem, err := pri.PublicKey().String()
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(filepath.Join(path, name),
		[]byte(pubPem), 0600); err != nil {
		return fmt.Errorf("failed to save public key, path = %s,  err = %s", name, err.Error())
	}
	return nil
}
