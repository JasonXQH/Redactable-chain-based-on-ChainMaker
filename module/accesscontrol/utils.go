/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"

	"chainmaker.org/chainmaker/common/v2/cert"
	bccrypto "chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	"chainmaker.org/chainmaker/common/v2/crypto/pkcs11"
	"chainmaker.org/chainmaker/common/v2/crypto/sdf"
	"chainmaker.org/chainmaker/localconf/v2"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/mr-tron/base58"
)

func getHSMHandleId() string {
	p11Config := localconf.ChainMakerConfig.NodeConfig.P11Config
	return p11Config.Library + p11Config.Label
}

func getHSMHandle() (interface{}, error) {
	var err error
	cfg := localconf.ChainMakerConfig.NodeConfig.P11Config
	hsmKey := getHSMHandleId()
	handle, ok := hsmHandleMap[hsmKey]
	if !ok {
		if strings.EqualFold(cfg.Type, "pkcs11") {
			handle, err = pkcs11.New(cfg.Library, cfg.Label, cfg.Password, cfg.SessionCacheSize,
				cfg.Hash)
		} else if strings.EqualFold(cfg.Type, "sdf") {
			handle, err = sdf.New(cfg.Library, cfg.SessionCacheSize)
		} else {
			err = fmt.Errorf("invalid hsm type, want pkcs11 | sdf, got %s", cfg.Type)
		}
		if err != nil {
			return nil, fmt.Errorf("fail to initialize organization with HSM: [%v]", err)
		}
		hsmHandleMap[hsmKey] = handle
	}
	return handle, nil
}

func pubkeyHash(pubkey []byte) string {
	pkHash := sha256.Sum256(pubkey)
	strPkHash := base58.Encode(pkHash[:])
	return strPkHash
}

// InitCertSigningMember 初始化一个证书签名的用户
// @param chainConfig
// @param localOrgId
// @param localPrivKeyFile
// @param localPrivKeyPwd
// @param localCertFile
// @return protocol.SigningMember
// @return error
func InitCertSigningMember(chainConfig *config.ChainConfig, localOrgId,
	localPrivKeyFile, localPrivKeyPwd, localCertFile string) (
	protocol.SigningMember, error) {

	var certMember *certificateMember

	if localPrivKeyFile != "" && localCertFile != "" {
		certPEM, err := ioutil.ReadFile(localCertFile)
		if err != nil {
			return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
		}

		isTrustMember := false
		for _, v := range chainConfig.TrustMembers {
			certBlock, _ := pem.Decode([]byte(v.MemberInfo))
			if certBlock == nil {
				return nil, fmt.Errorf("new member failed, the trsut member cert is not PEM")
			}
			if v.MemberInfo == string(certPEM) {
				certMember, err = newCertMemberFromParam(v.OrgId, v.Role,
					chainConfig.Crypto.Hash, false, certPEM)
				if err != nil {
					return nil, fmt.Errorf("init signing member failed, init trust member failed: [%s]", err.Error())
				}
				isTrustMember = true
				break
			}
		}

		if !isTrustMember {
			certMember, err = newMemberFromCertPem(localOrgId, chainConfig.Crypto.Hash, certPEM, false)
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
			}
		}

		skPEM, err := ioutil.ReadFile(localPrivKeyFile)
		if err != nil {
			return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
		}

		var sk bccrypto.PrivateKey
		cfg := localconf.ChainMakerConfig.NodeConfig.P11Config
		if cfg.Enabled {
			var handle interface{}
			handle, err = getHSMHandle()
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
			}
			sk, err = cert.ParseP11PrivKey(handle, skPEM)
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
			}
		} else {
			sk, err = asym.PrivateKeyFromPEM(skPEM, []byte(localPrivKeyPwd))
			if err != nil {
				return nil, err
			}
		}

		return &signingCertMember{
			*certMember,
			sk,
		}, nil
	}
	return nil, nil
}

// InitPKSigningMember 初始化一个公钥模式的用户
// @param hashType
// @param localOrgId
// @param localPrivKeyFile
// @param localPrivKeyPwd
// @return protocol.SigningMember
// @return error
func InitPKSigningMember(hashType,
	localOrgId, localPrivKeyFile, localPrivKeyPwd string) (protocol.SigningMember, error) {

	if localPrivKeyFile != "" {
		skPEM, err := ioutil.ReadFile(localPrivKeyFile)
		if err != nil {
			return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
		}

		var sk bccrypto.PrivateKey
		p11Config := localconf.ChainMakerConfig.NodeConfig.P11Config
		if p11Config.Enabled {
			var handle interface{}
			handle, err = getHSMHandle()
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%v]", err)
			}
			sk, err = cert.ParseP11PrivKey(handle, skPEM)
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%v]", err)
			}
		} else {
			sk, err = asym.PrivateKeyFromPEM(skPEM, []byte(localPrivKeyPwd))
			if err != nil {
				return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
			}
		}

		publicKeyBytes, err := sk.PublicKey().Bytes()
		if err != nil {
			return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
		}

		member, err := newPkMemberFromParam(localOrgId, publicKeyBytes, protocol.Role(""), hashType)
		if err != nil {
			return nil, fmt.Errorf("fail to initialize identity management service: [%s]", err.Error())
		}

		return &signingPKMember{
			*member,
			sk,
		}, nil
	}
	return nil, nil
}

//cryptoEngineOption parse public key by CryptoEngine
func cryptoEngineOption(cert *bcx509.Certificate) error {
	pkPem, err := cert.PublicKey.String()
	if err != nil {
		return fmt.Errorf("failed to get public key pem, err = %s", err)
	}
	cert.PublicKey, err = asym.PublicKeyFromPEM([]byte(pkPem))
	if err != nil {
		return fmt.Errorf("failed to parse public key, err = %s", err.Error())
	}
	return nil
}

// getBlockVersionAndResourceName return blockVersion and resourceName
func getBlockVersionAndResourceName(resourceNameWithPrefix string) (blockVersion uint32, resourceName string) {
	blockVersionAndResourceName := strings.Split(resourceNameWithPrefix, ":")
	if len(blockVersionAndResourceName) == 2 {
		version, err := strconv.ParseUint(blockVersionAndResourceName[0], 10, 32)
		if err != nil {
			blockVersion = 0
		}
		blockVersion = uint32(version)
		resourceName = blockVersionAndResourceName[1]
	} else if len(blockVersionAndResourceName) == 1 {
		resourceName = blockVersionAndResourceName[0]
	}

	return blockVersion, resourceName
}
