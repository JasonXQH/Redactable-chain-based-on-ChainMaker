/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker/pb-go/v2/consensus"

	bcx509 "chainmaker.org/chainmaker/common/v2/crypto/x509"

	"chainmaker.org/chainmaker/common/v2/msgbus"
	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"github.com/gogo/protobuf/proto"
)

var _ msgbus.Subscriber = (*certACProvider)(nil)

// OnMessage contract event data is a []string, hexToString(proto.Marshal(data))
func (cp *certACProvider) OnMessage(msg *msgbus.Message) {
	cp.acService.log.Infof("[AC] receive msg, topic: %s", msg.Topic.String())
	switch msg.Topic {
	case msgbus.ChainConfig:
		cp.onMessageChainConfig(msg)
	case msgbus.CertManageCertsFreeze:
		cp.onMessageCertFreeze(msg)
	case msgbus.CertManageCertsUnfreeze:
		cp.onMessageCertUnFreeze(msg)
	case msgbus.CertManageCertsRevoke:
		cp.onMessageCertRevoke(msg)
	case msgbus.CertManageCertsDelete:
		cp.onMessageCertDelete(msg)
	case msgbus.CertManageCertsAliasDelete:
		cp.onMessageCertAliasDelete(msg)
	case msgbus.CertManageCertsAliasUpdate:
		cp.onMessageCertAliasUpdate(msg)
	case msgbus.MaxbftEpochConf:
		cp.onMessageMaxbftChainconfigInEpoch(msg)
	}

}

func (cp *certACProvider) OnQuit() {
	// nothing
}

func (cp *certACProvider) onMessageChainConfig(msg *msgbus.Message) {
	dataStr, _ := msg.Payload.([]string)
	dataBytes, err := hex.DecodeString(dataStr[0])
	if err != nil {
		cp.acService.log.Error(err)
		return
	}
	chainConfig := &config.ChainConfig{}
	_ = proto.Unmarshal(dataBytes, chainConfig)

	cp.messageChainConfig(chainConfig, false)
}

func (cp *certACProvider) onMessageCertFreeze(msg *msgbus.Message) {
	data, _ := msg.Payload.([]string)
	certs := data[0]

	certList := strings.Replace(certs, ",", "\n", -1)
	cp.acService.log.Debugf("freeze certs: %s", certList)
	certBlock, rest := pem.Decode([]byte(certList))
	for certBlock != nil {
		if cp.consensusType == consensus.ConsensusType_MAXBFT && isConsensusCert(certBlock.Bytes) {
			cp.acService.log.Debugf("freeze certs delay for maxbft in epoch: %s")
			continue
		}
		cp.frozenList.Store(string(certBlock.Bytes), true)
		certBlock, rest = pem.Decode(rest)
	}
}

func (cp *certACProvider) onMessageCertUnFreeze(msg *msgbus.Message) {
	// full or hash cert
	data, _ := msg.Payload.([]string)
	certs := data[0]
	hashes := data[1]

	certList := strings.Replace(certs, ",", "\n", -1)
	cp.acService.log.Debugf("unfreeze cert hashes: %s, certs: %s", hashes, certList)
	certBlock, rest := pem.Decode([]byte(certList))
	for certBlock != nil {
		if cp.consensusType == consensus.ConsensusType_MAXBFT && isConsensusCert(certBlock.Bytes) {
			cp.acService.log.Debugf("unfreeze cert delay for maxbft in epoch: %s")
			continue
		}
		_, ok := cp.frozenList.Load(string(certBlock.Bytes))
		if ok {
			cp.frozenList.Delete(string(certBlock.Bytes))
		}
		certBlock, rest = pem.Decode(rest)
	}

	if hashes != "" {
		certHashes := strings.Split(hashes, ",")
		for _, hash := range certHashes {
			cert, err := cp.acService.dataStore.ReadObject(syscontract.SystemContract_CERT_MANAGE.String(), []byte(hash))
			if err != nil {
				cp.acService.log.Errorf("fail to load compressed certificate from local storage [%s]", hash)
				continue
			}
			if cert == nil {
				cp.acService.log.Warnf("cert id [%s] does not exist in local storage", hash)
				continue
			}
			_, ok := cp.frozenList.Load(string(cert))
			if ok {
				cp.frozenList.Delete(string(cert))
			}
		}
	}

}

func (cp *certACProvider) onMessageCertRevoke(msg *msgbus.Message) {
	crl := msg.Payload.([]string)[0]
	crl = strings.Replace(crl, ",", "\n", -1)
	cp.acService.log.Debugf("revoke cert crl: %s", crl)
	crls, err := cp.ValidateCRL([]byte(crl))
	if err != nil {
		err = fmt.Errorf("update CRL failed, invalid CRLS: %v", err)
		cp.acService.log.Error(err)
	}
	for _, crl := range crls {
		aki, _, err := bcx509.GetAKIFromExtensions(crl.TBSCertList.Extensions)
		if err != nil {
			err = fmt.Errorf("update CRL failed: %v", err)
			cp.acService.log.Error(err)
		}
		cp.crl.Store(string(aki), crl)
	}
}

func (cp *certACProvider) onMessageCertDelete(msg *msgbus.Message) {
	hashes := msg.Payload.([]string)[0]

	certHashStr := strings.TrimSpace(hashes)
	cp.acService.log.Debugf("delete cert hashes: %s", certHashStr)
	certHashes := strings.Split(certHashStr, ",")
	for _, hash := range certHashes {
		cp.acService.log.Debugf("certHashes in certsdelete = [%s]", hash)
		bin, err := hex.DecodeString(string(hash))
		if err != nil {
			cp.acService.log.Warnf("decode error for certhash: %s", string(hash))
		}
		_, ok := cp.certCache.Get(string(bin))
		if ok {
			cp.acService.log.Infof("remove certhash from certcache: %s", string(hash))
			cp.certCache.Remove(string(bin))
		}
	}
}

func (cp *certACProvider) onMessageCertAliasDelete(msg *msgbus.Message) {
	aliases := msg.Payload.([]string)[0]

	names := strings.TrimSpace(aliases)
	nameList := strings.Split(names, ",")
	cp.acService.log.Debugf("names in alias delete = [%s]", nameList)
	for _, name := range nameList {
		_, ok := cp.certCache.Get(string(name))
		if ok {
			cp.acService.log.Infof("remove alias from certcache: %s", string(name))
			cp.certCache.Remove(string(name))
		}
	}
}

func (cp *certACProvider) onMessageCertAliasUpdate(msg *msgbus.Message) {
	alias := msg.Payload.([]string)[0]

	name := strings.TrimSpace(alias)
	cp.acService.log.Infof("name in alias update = [%s]", name)
	_, ok := cp.certCache.Get(string(name))
	if ok {
		cp.acService.log.Infof("remove alias from certcache: %s", string(name))
		cp.certCache.Remove(string(name))
	}
}
