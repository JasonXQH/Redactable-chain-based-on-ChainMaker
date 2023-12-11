/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0

This file is for version compatibility
*/

package accesscontrol

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	pbac "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"

	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/gogo/protobuf/proto"
)

// ModuleNameAccessControl  "Access Control"
const ModuleNameAccessControl = "Access Control"

var _ protocol.Watcher = (*pkACProvider)(nil)

var _ protocol.Watcher = (*permissionedPkACProvider)(nil)
var _ protocol.VmWatcher = (*permissionedPkACProvider)(nil)

var _ protocol.Watcher = (*certACProvider)(nil)
var _ protocol.VmWatcher = (*certACProvider)(nil)

func (cp *certACProvider) Module() string {
	return ModuleNameAccessControl
}

func (cp *certACProvider) Watch(chainConfig *config.ChainConfig) error {
	cp.acService.hashType = chainConfig.GetCrypto().GetHash()
	err := cp.initTrustRootsForUpdatingChainConfig(chainConfig, cp.localOrg.id)
	if err != nil {
		return err
	}

	cp.acService.initResourcePolicy(chainConfig.ResourcePolicies, cp.localOrg.id)
	cp.acService.initResourcePolicy_220(chainConfig.ResourcePolicies, cp.localOrg.id)

	cp.opts.KeyUsages = make([]x509.ExtKeyUsage, 1)
	cp.opts.KeyUsages[0] = x509.ExtKeyUsageAny

	cp.acService.memberCache.Clear()
	cp.certCache.Clear()
	err = cp.initTrustMembers(chainConfig.TrustMembers)
	if err != nil {
		return err
	}
	return nil
}

func (cp *certACProvider) ContractNames() []string {
	return []string{syscontract.SystemContract_CERT_MANAGE.String()}
}

func (cp *certACProvider) Callback(contractName string, payloadBytes []byte) error {
	switch contractName {
	case syscontract.SystemContract_CERT_MANAGE.String():
		cp.acService.log.Infof("[AC] callback msg, contract name: %s", contractName)
		return cp.systemContractCallbackCertManagementCase(payloadBytes)
	default:
		cp.acService.log.Debugf("unwatched smart contract [%s]", contractName)
		return nil
	}
}

func (cp *certACProvider) systemContractCallbackCertManagementCase(payloadBytes []byte) error {
	var payload common.Payload
	err := proto.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("resolve payload failed: %v", err)
	}
	switch payload.Method {
	case syscontract.CertManageFunction_CERTS_FREEZE.String():
		return cp.systemContractCallbackCertManagementCertFreezeCase(&payload)
	case syscontract.CertManageFunction_CERTS_UNFREEZE.String():
		return cp.systemContractCallbackCertManagementCertUnfreezeCase(&payload)
	case syscontract.CertManageFunction_CERTS_REVOKE.String():
		return cp.systemContractCallbackCertManagementCertRevokeCase(&payload)
	case syscontract.CertManageFunction_CERTS_DELETE.String():
		return cp.systemContractCallbackCertManagementCertsDeleteCase(&payload)
	case syscontract.CertManageFunction_CERTS_ALIAS_DELETE.String():
		return cp.systemContractCallbackCertManagementAliasDeleteCase(&payload)
	case syscontract.CertManageFunction_CERT_ALIAS_UPDATE.String():
		return cp.systemContractCallbackCertManagementAliasUpdateCase(&payload)

	default:
		cp.acService.log.Debugf("unwatched method [%s]", payload.Method)
		return nil
	}
}
func (cp *certACProvider) systemContractCallbackCertManagementCertFreezeCase(payload *common.Payload) error {
	for _, param := range payload.Parameters {
		if param.Key == PARAM_CERTS {
			certList := strings.Replace(string(param.Value), ",", "\n", -1)
			certBlock, rest := pem.Decode([]byte(certList))
			for certBlock != nil {
				cp.frozenList.Store(string(certBlock.Bytes), true)

				certBlock, rest = pem.Decode(rest)
			}
			return nil
		}
	}
	return nil
}

func (cp *certACProvider) systemContractCallbackCertManagementCertUnfreezeCase(payload *common.Payload) error {
	for _, param := range payload.Parameters {
		if param.Key == PARAM_CERTS {
			certList := strings.Replace(string(param.Value), ",", "\n", -1)
			certBlock, rest := pem.Decode([]byte(certList))
			for certBlock != nil {
				_, ok := cp.frozenList.Load(string(certBlock.Bytes))
				if ok {
					cp.frozenList.Delete(string(certBlock.Bytes))
				}
				certBlock, rest = pem.Decode(rest)
			}
			return nil
		}
	}
	return nil
}

func (cp *certACProvider) systemContractCallbackCertManagementCertRevokeCase(payload *common.Payload) error {
	for _, param := range payload.Parameters {
		if param.Key == "cert_crl" {
			crl := strings.Replace(string(param.Value), ",", "\n", -1)
			crls, err := cp.ValidateCRL([]byte(crl))
			if err != nil {
				return fmt.Errorf("update CRL failed, invalid CRLS: %v", err)
			}
			for _, crl := range crls {
				aki, _, err := bcx509.GetAKIFromExtensions(crl.TBSCertList.Extensions)
				if err != nil {
					return fmt.Errorf("update CRL failed: %v", err)
				}
				cp.crl.Store(string(aki), crl)
			}
			return nil
		}
	}
	return nil
}

func (cp *certACProvider) systemContractCallbackCertManagementCertsDeleteCase(payload *common.Payload) error {
	cp.acService.log.Debugf("callback for certsdelete")
	for _, param := range payload.Parameters {
		if param.Key == PARAM_CERTHASHES {
			certHashStr := strings.TrimSpace(string(param.Value))
			certHashes := strings.Split(certHashStr, ",")
			for _, hash := range certHashes {
				cp.acService.log.Debugf("certHashes in certsdelete = [%s]", hash)
				bin, err := hex.DecodeString(string(hash))
				if err != nil {
					cp.acService.log.Warnf("decode error for certhash: %s", string(hash))
					return nil
				}
				_, ok := cp.certCache.Get(string(bin))
				if ok {
					cp.acService.log.Infof("remove certhash from certcache: %s", string(hash))
					cp.certCache.Remove(string(bin))
				}
			}

			return nil
		}
	}
	return nil
}

func (cp *certACProvider) systemContractCallbackCertManagementAliasDeleteCase(payload *common.Payload) error {
	cp.acService.log.Debugf("callback for aliasdelete")
	for _, param := range payload.Parameters {
		if param.Key == PARAM_ALIASES {
			names := strings.TrimSpace(string(param.Value))
			nameList := strings.Split(names, ",")
			cp.acService.log.Debugf("names in aliasdelete = [%s]", nameList)
			for _, name := range nameList {
				_, ok := cp.certCache.Get(string(name))
				if ok {
					cp.acService.log.Infof("remove alias from certcache: %s", string(name))
					cp.certCache.Remove(string(name))
				}
			}
			return nil
		}
	}
	return nil
}

func (cp *certACProvider) systemContractCallbackCertManagementAliasUpdateCase(payload *common.Payload) error {
	cp.acService.log.Debugf("callback for aliasupdate")
	for _, param := range payload.Parameters {
		if param.Key == PARAM_ALIAS {
			name := strings.TrimSpace(string(param.Value))
			cp.acService.log.Infof("name in aliasupdate = [%s]", name)
			_, ok := cp.certCache.Get(string(name))
			if ok {
				cp.acService.log.Infof("remove alias from certcache: %s", string(name))
				cp.certCache.Remove(string(name))
			}
			return nil
		}
	}
	return nil
}

func (p *pkACProvider) Module() string {
	return ModuleNameAccessControl
}

func (p *pkACProvider) Watch(chainConfig *config.ChainConfig) error {

	p.hashType = chainConfig.GetCrypto().GetHash()
	err := p.initAdminMembers(chainConfig.TrustRoots)
	if err != nil {
		return fmt.Errorf("new public AC provider failed: %s", err.Error())
	}

	err = p.initConsensusMember(chainConfig)
	if err != nil {
		return fmt.Errorf("new public AC provider failed: %s", err.Error())
	}
	p.memberCache.Clear()
	return nil
}

func (pp *permissionedPkACProvider) Module() string {
	return ModuleNameAccessControl
}

func (pp *permissionedPkACProvider) Watch(chainConfig *config.ChainConfig) error {
	pp.acService.hashType = chainConfig.GetCrypto().GetHash()

	err := pp.initAdminMembers(chainConfig.TrustRoots)
	if err != nil {
		return fmt.Errorf("update chainconfig error: %s", err.Error())
	}

	err = pp.initConsensusMember(chainConfig.Consensus.Nodes)
	if err != nil {
		return fmt.Errorf("update chainconfig error: %s", err.Error())
	}

	pp.acService.initResourcePolicy(chainConfig.ResourcePolicies, pp.localOrg)
	pp.acService.initResourcePolicy_220(chainConfig.ResourcePolicies, pp.localOrg)

	pp.acService.memberCache.Clear()

	return nil
}

func (pp *permissionedPkACProvider) ContractNames() []string {
	return []string{syscontract.SystemContract_PUBKEY_MANAGE.String()}
}

func (pp *permissionedPkACProvider) Callback(contractName string, payloadBytes []byte) error {
	switch contractName {
	case syscontract.SystemContract_PUBKEY_MANAGE.String():
		return pp.systemContractCallbackPublicKeyManagementCase(payloadBytes)
	default:
		pp.acService.log.Debugf("unwatched smart contract [%s]", contractName)
		return nil
	}
}

func (pp *permissionedPkACProvider) systemContractCallbackPublicKeyManagementCase(payloadBytes []byte) error {
	var payload common.Payload
	err := proto.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return fmt.Errorf("resolve payload failed: %v", err)
	}
	switch payload.Method {
	case syscontract.PubkeyManageFunction_PUBKEY_DELETE.String():
		return pp.systemContractCallbackPublicKeyManagementDeleteCase(&payload)
	default:
		pp.acService.log.Debugf("unwatched method [%s]", payload.Method)
		return nil
	}
}

func (pp *permissionedPkACProvider) systemContractCallbackPublicKeyManagementDeleteCase(
	payload *common.Payload) error {
	for _, param := range payload.Parameters {
		if param.Key == PUBLIC_KEYS {
			pk, err := asym.PublicKeyFromPEM(param.Value)
			if err != nil {
				return fmt.Errorf("delete member cache failed, [%v]", err.Error())
			}
			pkStr, err := pk.String()
			if err != nil {
				return fmt.Errorf("delete member cache failed, [%v]", err.Error())
			}
			pp.acService.memberCache.Remove(pkStr)
			pp.acService.log.Debugf("The public key was removed from the cache,[%v]", pkStr)
		}
	}
	return nil
}

func (acs *accessControlService) initResourcePolicy_220(resourcePolicies []*config.ResourcePolicy,
	localOrgId string) {
	authType := strings.ToLower(acs.authType)
	switch authType {
	case protocol.PermissionedWithCert, protocol.Identity:
		acs.createDefaultResourcePolicy_220(localOrgId)
	case protocol.PermissionedWithKey:
		acs.createDefaultResourcePolicyForPK_220(localOrgId)
	}
	for _, resourcePolicy := range resourcePolicies {
		if acs.validateResourcePolicy(resourcePolicy) {
			policy := newPolicyFromPb(resourcePolicy.Policy)
			acs.resourceNamePolicyMap220.Store(resourcePolicy.ResourceName, policy)
		}
	}
}

func (acs *accessControlService) lookUpPolicy220(resourceName string) (*pbac.Policy, error) {
	if p, ok := acs.resourceNamePolicyMap220.Load(resourceName); ok {
		return p.(*policy).GetPbPolicy(), nil
	}
	return nil, fmt.Errorf("policy not found for resource %s", resourceName)
}

func (acs *accessControlService) lookUpExceptionalPolicy220(resourceName string) (*pbac.Policy, error) {
	if p, ok := acs.exceptionalPolicyMap220.Load(resourceName); ok {
		return p.(*policy).GetPbPolicy(), nil

	}
	return nil, fmt.Errorf("exceptional policy not found for resource %s", resourceName)
}

func (acs *accessControlService) lookUpPolicyByResourceName220(resourceName string) (*policy, error) {
	p, ok := acs.resourceNamePolicyMap220.Load(resourceName)
	if !ok {
		if p, ok = acs.exceptionalPolicyMap220.Load(resourceName); !ok {
			return nil, fmt.Errorf("look up access policy failed, did not configure access policy "+
				"for resource %s", resourceName)
		}
	}
	return p.(*policy), nil
}

func (acs *pkACProvider) lookUpPolicy220(resourceName string) (*pbac.Policy, error) {
	if p, ok := acs.resourceNamePolicyMap220.Load(resourceName); ok {
		return p.(*policy).GetPbPolicy(), nil
	}
	return nil, fmt.Errorf("policy not found for resource %s", resourceName)
}

func (acs *pkACProvider) lookUpExceptionalPolicy220(resourceName string) (*pbac.Policy, error) {
	if p, ok := acs.exceptionalPolicyMap220.Load(resourceName); ok {
		return p.(*policy).GetPbPolicy(), nil

	}
	return nil, fmt.Errorf("exceptional policy not found for resource %s", resourceName)
}

func (acs *pkACProvider) lookUpPolicyByResourceName220(resourceName string) (*policy, error) {
	p, ok := acs.resourceNamePolicyMap220.Load(resourceName)
	if !ok {
		if p, ok = acs.exceptionalPolicyMap220.Load(resourceName); !ok {
			return nil, fmt.Errorf("look up access policy failed, did not configure access policy "+
				"for resource %s", resourceName)
		}
	}
	return p.(*policy), nil
}
