package accesscontrol

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
)

func (acs *accessControlService) createDefaultResourcePolicy_220(localOrgId string) {

	policyArchive.orgList = []string{localOrgId}

	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameReadData, policyRead)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameWriteData, policyWrite)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameUpdateSelfConfig, policySelfConfig)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameUpdateConfig, policyConfig)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameConsensusNode, policyConsensus)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameP2p, policyP2P)

	// only used for test
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameAllTest, policyAllTest)
	acs.resourceNamePolicyMap220.Store("test_2", policyLimitTestAny)
	acs.resourceNamePolicyMap220.Store("test_2_admin", policyLimitTestAdmin)
	acs.resourceNamePolicyMap220.Store("test_3/4", policyPortionTestAny)
	acs.resourceNamePolicyMap220.Store("test_3/4_admin", policyPortionTestAnyAdmin)

	// for txtype
	acs.resourceNamePolicyMap220.Store(common.TxType_QUERY_CONTRACT.String(), policyRead)
	acs.resourceNamePolicyMap220.Store(common.TxType_INVOKE_CONTRACT.String(), policyWrite)
	acs.resourceNamePolicyMap220.Store(common.TxType_SUBSCRIBE.String(), policySubscribe)
	acs.resourceNamePolicyMap220.Store(common.TxType_ARCHIVE.String(), policyArchive)

	// exceptional resourceName opened for light user
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_WITH_TXRWSETS_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_WITH_TXRWSETS_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_TX_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_LAST_CONFIG_BLOCK.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_LAST_BLOCK.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_FULL_BLOCK_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEIGHT_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEIGHT_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEADER_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_ARCHIVED_BLOCK_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_GET_CHAIN_CONFIG.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_QUERY.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ADD.String(), policySpecialWrite)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_ALIAS_QUERY.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_ADD.String(), policySpecialWrite)

	// Disable pubkey management for cert mode
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_ADD.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_DELETE.String(), policyForbidden)

	//for private compute
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNamePrivateCompute, policyWrite)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_CA_CERT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_ENCLAVE_REPORT.String(), policyConfig)

	// system contract interface resource definitions
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CORE_UPDATE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_BLOCK_UPDATE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_DELETE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_UPDATE.String(), policySelfConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_DELETE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_UPDATE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_DELETE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_UPDATE.String(), policySelfConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_INIT_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UPGRADE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_FREEZE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UNFREEZE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_GRANT_CONTRACT_ACCESS.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT_ACCESS.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_VERIFY_CONTRACT_ACCESS.String(), policyConfig)

	// certificate management
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_FREEZE.String(), policyAdmin)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_UNFREEZE.String(), policyAdmin)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_DELETE.String(), policyAdmin)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_REVOKE.String(), policyAdmin)
	// for cert_alias
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_UPDATE.String(), policyAdmin)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_ALIAS_DELETE.String(), policyAdmin)

}

func (acs *accessControlService) createDefaultResourcePolicyForPK_220(localOrgId string) {

	policyArchive.orgList = []string{localOrgId}

	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameReadData, policyRead)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameWriteData, policyWrite)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameUpdateSelfConfig, policySelfConfig)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameUpdateConfig, policyConfig)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameConsensusNode, policyConsensus)
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameP2p, policyP2P)

	// only used for test
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNameAllTest, policyAllTest)
	acs.resourceNamePolicyMap220.Store("test_2", policyLimitTestAny)
	acs.resourceNamePolicyMap220.Store("test_2_admin", policyLimitTestAdmin)
	acs.resourceNamePolicyMap220.Store("test_3/4", policyPortionTestAny)
	acs.resourceNamePolicyMap220.Store("test_3/4_admin", policyPortionTestAnyAdmin)

	// for txtype
	acs.resourceNamePolicyMap220.Store(common.TxType_QUERY_CONTRACT.String(), policyRead)
	acs.resourceNamePolicyMap220.Store(common.TxType_INVOKE_CONTRACT.String(), policyWrite)
	acs.resourceNamePolicyMap220.Store(common.TxType_SUBSCRIBE.String(), policySubscribe)
	acs.resourceNamePolicyMap220.Store(common.TxType_ARCHIVE.String(), policyArchive)

	// exceptional resourceName opened for light user
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_WITH_TXRWSETS_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_WITH_TXRWSETS_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_TX_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_LAST_CONFIG_BLOCK.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_LAST_BLOCK.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_FULL_BLOCK_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEIGHT_BY_TX_ID.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEIGHT_BY_HASH.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_BLOCK_HEADER_BY_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_QUERY.String()+"-"+
		syscontract.ChainQueryFunction_GET_ARCHIVED_BLOCK_HEIGHT.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_GET_CHAIN_CONFIG.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_QUERY.String(), policySpecialRead)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ADD.String(), policySpecialWrite)

	// Disable certificate management for pk mode
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ADD.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_FREEZE.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_UNFREEZE.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_DELETE.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_REVOKE.String(), policyForbidden)

	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_ADD.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_UPDATE.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_ALIAS_DELETE.String(), policyForbidden)

	// Disable trust member management for pk mode
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_ADD.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_DELETE.String(), policyForbidden)
	acs.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_UPDATE.String(), policyForbidden)

	//for private compute
	acs.resourceNamePolicyMap220.Store(protocol.ResourceNamePrivateCompute, policyWrite)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_CA_CERT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_ENCLAVE_REPORT.String(), policyConfig)

	// system contract interface resource definitions
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CORE_UPDATE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_BLOCK_UPDATE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_DELETE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_UPDATE.String(), policySelfConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_DELETE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_UPDATE.String(), policySelfConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_ADD.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_UPDATE.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_DELETE.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_INIT_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UPGRADE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_FREEZE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UNFREEZE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_GRANT_CONTRACT_ACCESS.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT_ACCESS.String(), policyConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_VERIFY_CONTRACT_ACCESS.String(), policyConfig)

	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_ADD.String(), policySelfConfig)
	acs.resourceNamePolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_DELETE.String(), policySelfConfig)
}

// need to consistent with 2.1.0 for dpos
func (p *pkACProvider) createDefaultResourcePolicyForDPoS_220() {

	p.resourceNamePolicyMap220.Store(protocol.ResourceNameConsensusNode, pubPolicyConsensus)
	// for txtype
	p.resourceNamePolicyMap220.Store(common.TxType_QUERY_CONTRACT.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_INVOKE_CONTRACT.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_SUBSCRIBE.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_ARCHIVE.String(), pubPolicyManage)

	// exceptional resourceName
	p.exceptionalPolicyMap220.Store(protocol.ResourceNamePrivateCompute, pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_CA_CERT.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_ENCLAVE_REPORT.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_DELETE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_UPDATE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_DELETE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_UPDATE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_FREEZE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_UNFREEZE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_DELETE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_REVOKE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_ALIAS_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_DELETE.String(), pubPolicyForbidden)

	// disable trust root add & delete for public mode
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_DELETE.String(), pubPolicyForbidden)

	// disable multisign for public mode
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_MULTI_SIGN.String()+"-"+
		syscontract.MultiSignFunction_REQ.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_MULTI_SIGN.String()+"-"+
		syscontract.MultiSignFunction_VOTE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_MULTI_SIGN.String()+"-"+
		syscontract.MultiSignFunction_QUERY.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CORE_UPDATE.String(), pubPolicyManage)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_BLOCK_UPDATE.String(), pubPolicyManage)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UPGRADE_CONTRACT.String(), pubPolicyManage)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_FREEZE_CONTRACT.String(), pubPolicyManage)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UNFREEZE_CONTRACT.String(), pubPolicyManage)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT.String(), pubPolicyManage)
	// disable contract access for public mode
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_GRANT_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_VERIFY_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractQueryFunction_GET_DISABLED_CONTRACT_LIST.String(), pubPolicyForbidden)

	// disable gas related native contract
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_CHARGE_GAS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_REFUND_GAS_VM.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_SET_ADMIN.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_ENABLE_OR_DISABLE_GAS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_ALTER_ADDR_TYPE.String(), pubPolicyForbidden)

	// for admin management
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_UPDATE.String(), pubPolicyMajorityAdmin)
}

func (p *pkACProvider) createDefaultResourcePolicy_220() {

	p.resourceNamePolicyMap220.Store(protocol.ResourceNameConsensusNode, pubPolicyConsensus)
	// for txtype
	p.resourceNamePolicyMap220.Store(common.TxType_QUERY_CONTRACT.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_INVOKE_CONTRACT.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_SUBSCRIBE.String(), pubPolicyTransaction)
	p.resourceNamePolicyMap220.Store(common.TxType_ARCHIVE.String(), pubPolicyManage)

	// exceptional resourceName
	p.exceptionalPolicyMap220.Store(protocol.ResourceNamePrivateCompute, pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_CA_CERT.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PRIVATE_COMPUTE.String()+"-"+
		syscontract.PrivateComputeFunction_SAVE_ENCLAVE_REPORT.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_DELETE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_MEMBER_UPDATE.String(), pubPolicyForbidden)

	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_ADD.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_DELETE.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ID_UPDATE.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_UPDATE.String(), pubPolicyMajorityAdmin)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_NODE_ORG_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CONSENSUS_EXT_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_PERMISSION_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_FREEZE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_UNFREEZE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_DELETE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_REVOKE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERT_ALIAS_UPDATE.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CERT_MANAGE.String()+"-"+
		syscontract.CertManageFunction_CERTS_ALIAS_DELETE.String(), pubPolicyForbidden)

	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_PUBKEY_MANAGE.String()+"-"+
		syscontract.PubkeyManageFunction_PUBKEY_DELETE.String(), pubPolicyForbidden)

	// disable trust root add & delete for public mode
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_ADD.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_DELETE.String(), pubPolicyForbidden)

	// disable contract access for public mode
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_GRANT_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_VERIFY_CONTRACT_ACCESS.String(), pubPolicyForbidden)
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractQueryFunction_GET_DISABLED_CONTRACT_LIST.String(), pubPolicyForbidden)

	// forbidden charge gas by go sdk
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_CHARGE_GAS.String(), pubPolicyForbidden)

	// forbidden refund gas vm by go sdk
	p.exceptionalPolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_REFUND_GAS_VM.String(), pubPolicyForbidden)

	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_INIT_CONTRACT.String(), pubPolicyManage)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UPGRADE_CONTRACT.String(), pubPolicyManage)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_FREEZE_CONTRACT.String(), pubPolicyManage)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_UNFREEZE_CONTRACT.String(), pubPolicyManage)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CONTRACT_MANAGE.String()+"-"+
		syscontract.ContractManageFunction_REVOKE_CONTRACT.String(), pubPolicyManage)

	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_CORE_UPDATE.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_BLOCK_UPDATE.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_ENABLE_OR_DISABLE_GAS.String(), pubPolicyMajorityAdmin)
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_ALTER_ADDR_TYPE.String(), pubPolicyMajorityAdmin)

	// for admin management
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_CHAIN_CONFIG.String()+"-"+
		syscontract.ChainConfigFunction_TRUST_ROOT_UPDATE.String(), pubPolicyMajorityAdmin)

	// for gas admin
	p.resourceNamePolicyMap220.Store(syscontract.SystemContract_ACCOUNT_MANAGER.String()+"-"+
		syscontract.GasAccountFunction_SET_ADMIN.String(), pubPolicyMajorityAdmin)
}
