package accesscontrol

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"chainmaker.org/chainmaker/common/v2/json"
	"chainmaker.org/chainmaker/common/v2/msgbus"

	"chainmaker.org/chainmaker/pb-go/v2/consensus/maxbft"
	"github.com/gogo/protobuf/proto"

	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/consensus"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
)

const CONSENSUS = "consensus"

func (cp *certACProvider) messageChainConfig(chainConfig *config.ChainConfig, fromMaxBFT bool) {
	cp.acService.hashType = chainConfig.GetCrypto().GetHash()
	cp.acService.initResourcePolicy(chainConfig.ResourcePolicies, cp.localOrg.id)

	updateTrustRootAndMemberFunc := func() {
		err := cp.initTrustRootsForUpdatingChainConfig(chainConfig, cp.localOrg.id)
		if err != nil {
			cp.acService.log.Error(err)
			return
		}

		cp.opts.KeyUsages = make([]x509.ExtKeyUsage, 1)
		cp.opts.KeyUsages[0] = x509.ExtKeyUsageAny

		cp.acService.memberCache.Clear()
		cp.certCache.Clear()
		err = cp.initTrustMembers(chainConfig.TrustMembers)
		if err != nil {
			cp.acService.log.Error(err)
			return
		}
	}

	//if consensus is maxbft, delay update
	if cp.consensusType != consensus.ConsensusType_MAXBFT {
		updateTrustRootAndMemberFunc()
	} else {
		if fromMaxBFT {
			updateTrustRootAndMemberFunc()
		}
	}
}

// nolint: unused
func (cp *certACProvider) isConsensusAKI(aki string) {
	//检查aki是否对应共识节点，需要获取共识节点证书列表
	//func (cp *certACProvider) checkCRL(certChain []*bcx509.Certificate) error {
	//	if len(certChain) < 1 {
	//	return fmt.Errorf("given certificate chain is empty")
	//}
	//
	//	for _, cert := range certChain {
	//	akiCert := cert.AuthorityKeyId
	//
	//	crl, ok := cp.crl.Load(string(akiCert))
	//	if ok {
	//	// we have ac CRL, check whether the serial number is revoked
	//	for _, rc := range crl.(*pkix.CertificateList).TBSCertList.RevokedCertificates {
	//	if rc.SerialNumber.Cmp(cert.SerialNumber) == 0 {
	//	return errors.New("certificate is revoked")
	//}

	//nodeList := cp.chainConfig.Consensus.Nodes
}

func isConsensusCert(raw interface{}) bool {
	switch certInfo := raw.(type) {
	case *bcx509.Certificate:
		if len(certInfo.Subject.OrganizationalUnit) != 0 &&
			certInfo.Subject.OrganizationalUnit[0] == CONSENSUS {
			return true
		}
	case []byte:
		cert, err := bcx509.ParseCertificate(certInfo)
		if err != nil {
			return false
		}
		if len(cert.Subject.OrganizationalUnit) != 0 &&
			cert.Subject.OrganizationalUnit[0] == CONSENSUS {
			return true
		}
	}
	return false
}

//loadChainConfigFromGovernance used to load config from system contract, only for maxbft
func (cp *certACProvider) loadChainConfigFromGovernance() (*maxbft.GovernanceContract, error) {
	contractName := syscontract.SystemContract_GOVERNANCE.String()
	bz, err := cp.store.ReadObject(contractName, []byte(contractName))
	if err != nil {
		return nil, fmt.Errorf("get contractName=%s from db failed, reason: %s", contractName, err)
	}
	if bz == nil {
		return nil, nil
	}
	cfg := &maxbft.GovernanceContract{}
	if err = proto.Unmarshal(bz, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal contractName=%s failed, reason: %s", contractName, err)
	}
	return cfg, nil
}

// onMessageMaxbftChainconfigInEpoch update ac for maxbft
/*
	1. if not maxbft, return
	2. update consensusType if change (not effective now, need consensus module support)
	3. refresh trustroots
	4. refresh trustmembers
	5. refresh freezeList
	6. refresh crlList
*/
func (cp *certACProvider) onMessageMaxbftChainconfigInEpoch(msg *msgbus.Message) {
	epochConfig, ok := msg.Payload.(*maxbft.GovernanceContract)
	if !ok {
		cp.acService.log.Error("payload is not *maxbft.GovernanceContract")
		return
	}

	//update chainconfig
	cp.messageChainConfig(epochConfig.ChainConfig, true)
	if err := cp.updateFrozenAndCRL(epochConfig); err != nil {
		cp.acService.log.Errorf("fail to updateFrozenAndCRL: %s", err.Error())
	}
}

func (cp *certACProvider) updateFrozenAndCRL(epochConfig *maxbft.GovernanceContract) error {
	//update frozenList
	if len(epochConfig.CertFrozenList) != 0 {
		var certIDs []string
		if err := json.Unmarshal(epochConfig.CertFrozenList, &certIDs); err != nil {
			return fmt.Errorf("unmarshal frozen certificate list failed: %v", err)
		}
		for _, certID := range certIDs {
			certBytes, err := cp.acService.dataStore.
				ReadObject(syscontract.SystemContract_CERT_MANAGE.String(), []byte(certID))
			if err != nil {
				return fmt.Errorf("load frozen certificate failed: %s", certID)
			}
			if certBytes == nil {
				return fmt.Errorf("load frozen certificate failed: empty certificate [%s]", certID)
			}

			certBlock, _ := pem.Decode(certBytes)
			cp.frozenList.Store(string(certBlock.Bytes), true)
		}
	}

	//update crl
	if len(epochConfig.CRL) != 0 {
		var crlAKIs []string
		if err := json.Unmarshal(epochConfig.CRL, &crlAKIs); err != nil {
			return fmt.Errorf("fail to Unmarshal CRL list: %v", err)
		}
		if err := cp.storeCrls(crlAKIs); err != nil {
			return fmt.Errorf("fail to update CRL list: %v", err)
		}
	}
	return nil
}
