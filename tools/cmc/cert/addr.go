/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cert

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"

	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	hashAlo "chainmaker.org/chainmaker/common/v2/crypto/hash"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
)

// certToUserAddrInStake get user addr feature of the DPoS from cert
// @return *cobra.Command
func certToUserAddrInStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userAddr",
		Short: "get user addr feature of the DPoS from cert",
		RunE: func(_ *cobra.Command, _ []string) error {
			if len(pubkeyOrCertPath) == 0 {
				return fmt.Errorf("cert or pubkey path is null")
			}
			chainClient, err := sdk.NewChainClient(sdk.WithConfPath(sdkConfPath))
			if err != nil {
				return err
			}

			var (
				pubkey   []byte
				hashBz   []byte
				authType = chainClient.GetAuthType()
				hashType = chainClient.GetHashType()
			)
			content, err := ioutil.ReadFile(pubkeyOrCertPath)
			if err != nil {
				return fmt.Errorf("read cert content failed, reason: %s", err)
			}
			if authType == sdk.PermissionedWithCert {
				if pubkey, err = getPubkeyFromCert(content); err != nil {
					return err
				}
			} else if authType == sdk.PermissionedWithKey || authType == sdk.Public {
				pk, err := asym.PublicKeyFromPEM(content)
				if err != nil {
					return err
				}
				pkStr, err := pk.String()
				if err != nil {
					return err
				}
				pubkey = []byte(pkStr)
			}
			if hashType == crypto.HASH_TYPE_SM3 || hashType == crypto.HASH_TYPE_SHA256 {
				if hashBz, err = hashAlo.Get(hashType, pubkey); err != nil {
					return err
				}
			}
			addr := base58.Encode(hashBz[:])
			fmt.Printf("address: %s \n\nfrom cert: %s\n", addr, pubkeyOrCertPath)
			return nil
		},
	}

	attachFlags(cmd, []string{
		flagSdkConfPath,
		flagCertOrPubkeyPath,
	})

	cmd.MarkFlagRequired(flagCertOrPubkeyPath)
	return cmd
}

// getPubkeyFromCert get public key from a pem format certificate
// @param certContent
// @return []byte
// @return error
func getPubkeyFromCert(certContent []byte) ([]byte, error) {
	block, _ := pem.Decode(certContent)
	if block == nil {
		return nil, errors.New("pem.Decode failed, invalid cert")
	}
	cert, err := bcx509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert failed, reason: %s", err)
	}
	pubkey, err := cert.PublicKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("get pubkey failed from cert, reason: %s", err)
	}
	return pubkey, nil
}
