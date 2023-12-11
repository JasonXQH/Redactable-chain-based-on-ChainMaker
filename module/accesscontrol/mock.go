/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"sync"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/protocol/v2/test"

	"chainmaker.org/chainmaker/common/v2/concurrentlru"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

var mockAcLogger = &test.GoLogger{}

// MockAccessControl Mock一个AccessControlProvider
// @return protocol.AccessControlProvider
func MockAccessControl() protocol.AccessControlProvider {
	certAc := &certACProvider{
		acService: &accessControlService{
			orgList:               &sync.Map{},
			orgNum:                0,
			resourceNamePolicyMap: &sync.Map{},
			hashType:              "",
			dataStore:             nil,
			memberCache:           concurrentlru.New(0),
			log:                   mockAcLogger,
		},
		certCache:  concurrentlru.New(0),
		crl:        sync.Map{},
		frozenList: sync.Map{},
		opts: bcx509.VerifyOptions{
			Intermediates: bcx509.NewCertPool(),
			Roots:         bcx509.NewCertPool(),
		},
		localOrg: nil,
	}
	return certAc
}

// MockAccessControlWithHash Mock一个AccessControlProvider
// @param hashAlg
// @return protocol.AccessControlProvider
func MockAccessControlWithHash(hashAlg string) protocol.AccessControlProvider {
	certAc := &certACProvider{
		acService: &accessControlService{
			orgList:               &sync.Map{},
			orgNum:                0,
			resourceNamePolicyMap: &sync.Map{},
			hashType:              hashAlg,
			dataStore:             nil,
			memberCache:           concurrentlru.New(0),
			log:                   mockAcLogger,
		},
		certCache:  concurrentlru.New(0),
		crl:        sync.Map{},
		frozenList: sync.Map{},
		opts: bcx509.VerifyOptions{
			Intermediates: bcx509.NewCertPool(),
			Roots:         bcx509.NewCertPool(),
		},
		localOrg: nil,
	}
	return certAc
}

// MockSignWithMultipleNodes Mock一个多用户的背书签名
// @param msg
// @param signers
// @param hashType
// @return []*commonPb.EndorsementEntry
// @return error
func MockSignWithMultipleNodes(msg []byte, signers []protocol.SigningMember, hashType string) (
	[]*commonPb.EndorsementEntry, error) {
	var ret []*commonPb.EndorsementEntry
	for _, signer := range signers {
		sig, err := signer.Sign(hashType, msg)
		if err != nil {
			return nil, err
		}
		signerSerial, err := signer.GetMember()
		if err != nil {
			return nil, err
		}
		ret = append(ret, &commonPb.EndorsementEntry{
			Signer:    signerSerial,
			Signature: sig,
		})
	}
	return ret, nil
}

// NewAccessControlWithChainConfig 构造一个AccessControlProvider
// @param chainConfig
// @param localOrgId
// @param store
// @param log
// @param msgBus
// @return protocol.AccessControlProvider
// @return error
func NewAccessControlWithChainConfig(chainConfig protocol.ChainConf, localOrgId string,
	store protocol.BlockchainStore, log protocol.Logger, msgBus msgbus.MessageBus) (
	protocol.AccessControlProvider, error) {
	conf := chainConfig.ChainConfig()
	acp, err := newCertACProvider(conf, localOrgId, store, log)
	if err != nil {
		return nil, err
	}
	msgBus.Register(msgbus.ChainConfig, acp)
	msgBus.Register(msgbus.CertManageCertsRevoke, acp)
	msgBus.Register(msgbus.CertManageCertsUnfreeze, acp)
	msgBus.Register(msgbus.CertManageCertsFreeze, acp)
	chainConfig.AddWatch(acp)   //nolint: staticcheck
	chainConfig.AddVmWatch(acp) //nolint: staticcheck
	//InitCertSigningMember(testChainConfig, localOrgId, localPrivKeyFile, localPrivKeyPwd, localCertFile)
	return acp, err
}
