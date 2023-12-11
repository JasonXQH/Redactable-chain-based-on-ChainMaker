// Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

// Package address 关于用户地址的相关命令
package address

import (
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	commonCert "chainmaker.org/chainmaker/common/v2/cert"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/sdk-go/v2/utils"
	cmutils "chainmaker.org/chainmaker/utils/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagHashType     = "hash-type"
	flagBlockVersion = "block-version"

	// address types
	addressTypeZXL = "zhixinchain"
	addressTypeCM  = "chainmaker"
	addressTypeEVM = "ethereum"
)

var (
	hashType    int
	hashAlgoMap = map[int]crypto.HashType{
		0: crypto.HASH_TYPE_SHA256,
		1: crypto.HASH_TYPE_SHA3_256,
		2: crypto.HASH_TYPE_SM3,
	}
	blockVersion uint32
)

var flags *pflag.FlagSet

func init() {
	flags = &pflag.FlagSet{}
	flags.IntVar(&hashType, flagHashType, 0,
		"The type of hash algo obtained. 0: SAH256 (default), 1: SHA3_256, 2: SM3 eg. --hash-type=1")
	flags.Uint32Var(&blockVersion, flagBlockVersion, 2300, "Block version")
}

type addrSki struct {
	Address string `json:"address"`
	Ski     string `json:"ski"`
}

// NewAddressCMD new address parse command
func NewAddressCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "address parse command",
		Long:  "address parse command",
	}

	cmd.AddCommand(newPK2AddrCMD())
	cmd.AddCommand(newHex2AddrCMD())
	cmd.AddCommand(newCert2AddrCMD())
	cmd.AddCommand(newName2AddrCMD())

	return cmd
}

// newPK2AddrCMD get address from public key file or pem string
// @return *cobra.Command
func newPK2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pk-to-addr [public key file path / pem string]",
		Short: "get address from public key file or pem string",
		Long:  "get address from public key file or pem string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var keyPemStr string
			var isFile = utils.Exists(args[0])
			if isFile {
				keyPem, err := ioutil.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read key file failed, %s", err)
				}
				keyPemStr = string(keyPem)
			} else {
				keyPemStr = args[0]
			}

			hash, ok := hashAlgoMap[hashType]
			if !ok {
				return fmt.Errorf("unsupported hash type %d", hashType)
			}
			pemBlock, _ := pem.Decode([]byte(keyPemStr))
			if pemBlock == nil {
				return fmt.Errorf("fail to resolve public key, key file not exists or PEM string invalid")
			}
			pkDER := pemBlock.Bytes
			pk, err := asym.PublicKeyFromDER(pkDER)
			if err != nil {
				return fmt.Errorf("fail to resolve public key from DER format: %v", err)
			}
			ski, err := commonCert.ComputeSKI(hash, pk.ToStandardKey())
			if err != nil {
				return err
			}
			skiHex := hex.EncodeToString(ski)

			addrZxl, err := sdk.GetZXAddressFromPKPEM(keyPemStr, hash)
			if err != nil {
				return err
			}
			addrCm, err := sdk.GetCMAddressFromPKPEM(keyPemStr, hash)
			if err != nil {
				return err
			}
			addrEvm, err := sdk.GetEVMAddressFromPKPEM(keyPemStr, hash)
			if err != nil {
				return err
			}

			var addrSkis = map[string]addrSki{
				addressTypeEVM: {addrEvm, skiHex},
				addressTypeCM:  {addrCm, skiHex},
				addressTypeZXL: {addrZxl, skiHex},
			}

			util.PrintPrettyJson(addrSkis)
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagHashType})
	return cmd
}

// newHex2AddrCMD get address from hex string
// @return *cobra.Command
func newHex2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hex-to-addr [hex string]",
		Short: "get address from public key DER hex string",
		Long:  "get address from public key DER hex string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, ok := hashAlgoMap[hashType]
			if !ok {
				return fmt.Errorf("unsupported hash type %d", hashType)
			}
			pkDER, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			pk, err := asym.PublicKeyFromDER(pkDER)
			if err != nil {
				return fmt.Errorf("fail to resolve public key from DER format: %v", err)
			}
			ski, err := commonCert.ComputeSKI(hash, pk.ToStandardKey())
			if err != nil {
				return err
			}
			skiHex := hex.EncodeToString(ski)

			addrZxl, err := sdk.GetZXAddressFromPKHex(args[0], hash)
			if err != nil {
				return err
			}
			addrCm, err := sdk.GetCMAddressFromPKHex(args[0], hash)
			if err != nil {
				return err
			}
			addrEvm, err := sdk.GetEVMAddressFromPKHex(args[0], hash)
			if err != nil {
				return err
			}

			var addrSkis = map[string]addrSki{
				addressTypeEVM: {addrEvm, skiHex},
				addressTypeCM:  {addrCm, skiHex},
				addressTypeZXL: {addrZxl, skiHex},
			}

			util.PrintPrettyJson(addrSkis)
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagHashType})
	return cmd
}

// newCert2AddrCMD get address from cert file or pem string
// @return *cobra.Command
func newCert2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert-to-addr [cert file path / pem string]",
		Short: "get address from cert file or pem string",
		Long:  "get address from cert file or pem string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var certPemStr string
			var isFile = utils.Exists(args[0])
			if isFile {
				keyPem, err := ioutil.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read key file failed, %s", err)
				}
				certPemStr = string(keyPem)
			} else {
				certPemStr = args[0]
			}

			block, _ := pem.Decode([]byte(certPemStr))
			if block == nil {
				return fmt.Errorf("fail to resolve cert, cert file not exists or PEM string invalid")
			}
			cert, err := bcx509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("ParseCertificate failed, %s", err)
			}
			skiHex := hex.EncodeToString(cert.SubjectKeyId)

			addrZxl, err := sdk.GetZXAddressFromCertPEM(certPemStr)
			if err != nil {
				return err
			}
			addrCm, err := sdk.GetCMAddressFromCertPEM(certPemStr)
			if err != nil {
				return err
			}
			addrEvm, err := sdk.GetEVMAddressFromCertBytes([]byte(certPemStr))
			if err != nil {
				return err
			}

			var addrSkis = map[string]addrSki{
				addressTypeEVM: {addrEvm, skiHex},
				addressTypeCM:  {addrCm, skiHex},
				addressTypeZXL: {addrZxl, skiHex},
			}

			util.PrintPrettyJson(addrSkis)
			return nil
		},
	}
	return cmd
}

// newName2AddrCMD get address from contract name string
// @return *cobra.Command
func newName2AddrCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "name-to-addr [contract name string]",
		Short: "get address from contract name string",
		Long:  "get address from contract name string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrCm, err := cmutils.NameToAddrStr(args[0], configPb.AddrType_CHAINMAKER, blockVersion)
			if err != nil {
				return err
			}
			addrZxl, err := cmutils.NameToAddrStr(args[0], configPb.AddrType_ZXL, blockVersion)
			if err != nil {
				return err
			}
			addrEvm, err := cmutils.NameToAddrStr(args[0], configPb.AddrType_ETHEREUM, blockVersion)
			if err != nil {
				return err
			}

			var addrSkis = map[string]addrSki{
				addressTypeEVM: {addrEvm, ""},
				addressTypeCM:  {addrCm, ""},
				addressTypeZXL: {"ZX" + addrZxl, ""},
			}

			util.PrintPrettyJson(addrSkis)
			return nil
		},
	}
	util.AttachFlags(cmd, flags, []string{flagBlockVersion})
	return cmd
}
