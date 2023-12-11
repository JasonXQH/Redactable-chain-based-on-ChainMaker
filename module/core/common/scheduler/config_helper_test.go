package scheduler

import (
	"fmt"
	"testing"

	crypto2 "chainmaker.org/chainmaker/common/v2/crypto"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	configPb "chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	mock2 "chainmaker.org/chainmaker/protocol/v2/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	MemberInfo_Node1_Org1 = "-----BEGIN CERTIFICATE-----\nMIICwzCCAmigAwIBAgIDB+wHMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZcxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMRIwEAYDVQQLEwljb25zZW5zdXMxLzAtBgNVBAMTJmNvbnNlbnN1czEuc2lnbi53\neC1vcmcxLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\nCM+X9KLu/23Vv54PE0WFC2bFz3tuvxjWx6NWCXfzRVSYLuKlrz0aqF8FudMPZlxH\nIy8mYi8ikgpTLdv4+JH9QqOBrTCBqjAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgw\nBgYEVR0lADApBgNVHQ4EIgQgBfC3rh1kbAb01LjWSkl/p9PWm+6RSLwQYV9mDskB\nBI8wKwYDVR0jBCQwIoAgNSQ/cRy5t8Q1LpMfcMVzMfl0CcLZ4Pvf7BxQX9sQiWcw\nLwYLgSdYj2QLHo9kCwQEIDAwMTY0NmU2NzgwZjRiMGQ4YmVhMzIzZWU4YzI0OTE1\nMAoGCCqGSM49BAMCA0kAMEYCIQD4zscd08B/9QW/ywYC3u2OfXAzm+ajqHB75eAX\nOQYIlgIhAIbAZ95LCP+QWIg2UuCpSXLVvDOlwRb4jq01Pfm30N14\n-----END CERTIFICATE-----"

	MemberInfo_Client1_Org1 = "-----BEGIN CERTIFICATE-----\nMIICijCCAi+gAwIBAgIDBS9vMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcxLmNoYWlubWFrZXIub3Jn\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcx\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE56xayRx0\n/a8KEXPxRfiSzYgJ/sE4tVeI/ZbjpiUX9m0TCJX7W/VHdm6WeJLOdCDuLLNvjGTy\nt8LLyqyubJI5AKN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\nKQYDVR0OBCIEIMjAiM2eMzlQ9HzV9ePW69rfUiRZVT2pDBOMqM4WVJSAMCsGA1Ud\nIwQkMCKAIDUkP3EcubfENS6TH3DFczH5dAnC2eD73+wcUF/bEIlnMAoGCCqGSM49\nBAMCA0kAMEYCIQCWUHL0xisjQoW+o6VV12pBXIRJgdeUeAu2EIjptSg2GAIhAIxK\nLXpHIBFxIkmWlxUaanCojPSZhzEbd+8LRrmhEO8n\n-----END CERTIFICATE-----"
	MemberInfo_Client1_Org2 = "-----BEGIN CERTIFICATE-----\nMIICiTCCAi+gAwIBAgIDA+zYMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmcyLmNoYWlubWFrZXIub3Jn\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmcy\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEZd92CJez\nCiOMzLSTrJfX5vIUArCycg05uKru2qFaX0uvZUCwNxbfSuNvkHRXE8qIBUhTbg1Q\nR9rOlfDY1WfgMaN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\nKQYDVR0OBCIEICfLatSyyebzRsLbnkNKZJULB2bZOtG+88NqvAHCsXa3MCsGA1Ud\nIwQkMCKAIPGP1bPT4/Lns2PnYudZ9/qHscm0pGL6Kfy+1CAFWG0hMAoGCCqGSM49\nBAMCA0gAMEUCIQDzHrEHrGNtoNfB8jSJrGJU1qcxhse74wmDgIdoGjvfTwIgabRJ\nJNvZKRpa/VyfYi3TXa5nhHRIn91ioF1dQroHQFc=\n-----END CERTIFICATE-----"
	MemberInfo_Client1_Org3 = "-----BEGIN CERTIFICATE-----\nMIIChzCCAi2gAwIBAgIDAzqjMAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnMy5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmczLmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgY8xCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmczLmNoYWlubWFrZXIub3Jn\nMQ4wDAYDVQQLEwVhZG1pbjErMCkGA1UEAxMiYWRtaW4xLnNpZ24ud3gtb3JnMy5j\naGFpbm1ha2VyLm9yZzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABPSS/PaGKoB3\nPbPb1bCAw9naF1U85niyPEUJ9toVHixIPMrARcVHx8reuttIicr3IilPV65c56aZ\nhQvJEiKftL6jezB5MA4GA1UdDwEB/wQEAwIBpjAPBgNVHSUECDAGBgRVHSUAMCkG\nA1UdDgQiBCCBB9bwHSK8/Pbjksr03+sAo2c4TKexLH2hh79C67eYJTArBgNVHSME\nJDAigCDRj2UdLFcK72LRZ3kw+hlMgUH5cKU5idKgrIL3RYCJ/TAKBggqhkjOPQQD\nAgNIADBFAiEA3S65egfTOPxqLmxvMJAjx6+G32Q7FhnJzBCx+SRt4SACICFvjRRn\nU+jpCYT2nE5cpL4KHVf3DTvum/+2tebm5SxY\n-----END CERTIFICATE-----\n"
	MemberInfo_Client1_Org4 = "-----BEGIN CERTIFICATE-----\nMIICiTCCAi+gAwIBAgIDC8x8MAoGCCqGSM49BAMCMIGKMQswCQYDVQQGEwJDTjEQ\nMA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEfMB0GA1UEChMWd3gt\nb3JnNC5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MSIwIAYDVQQD\nExljYS53eC1vcmc0LmNoYWlubWFrZXIub3JnMB4XDTIwMTIwODA2NTM0M1oXDTI1\nMTIwNzA2NTM0M1owgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAw\nDgYDVQQHEwdCZWlqaW5nMR8wHQYDVQQKExZ3eC1vcmc0LmNoYWlubWFrZXIub3Jn\nMQ8wDQYDVQQLEwZjbGllbnQxLDAqBgNVBAMTI2NsaWVudDEuc2lnbi53eC1vcmc0\nLmNoYWlubWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE3+arrmoR\nyRZ8SSeeu+omEjsQJFEZuKdpZT+DWWESt31r/xCoOWzWVUUL/OG1IULiPlmHESCD\nlViLtkes//1KzqN7MHkwDgYDVR0PAQH/BAQDAgGmMA8GA1UdJQQIMAYGBFUdJQAw\nKQYDVR0OBCIEIFmwGrL6Uk7/iARcZM5nNhpLqce/UL2eGhXQ2WMPWML/MCsGA1Ud\nIwQkMCKAILnKrWgqcf7PtgtMJFynmIqdFt0oVKmKzxD+NfPshcHEMAoGCCqGSM49\nBAMCA0gAMEUCIQCBXzILOOrLIRHzD2Y3cvocutlnFNEf1hTiKec85J81uwIgHLvg\nOgJ0Jauz2EebjDLNBf94omiCO2WasmUxqLSFTCM=\n-----END CERTIFICATE-----"

	Address_Client1_Org1_ZXL = "ZXd980152d2e47b8c4bb9ec4bb773cfbc4538a92f3"
	Address_Client1_Org2_ZXL = "ZXe5a01ff94403a0ccd18b5974c46334f26fb1f0b2"
	Address_Client1_Org3_ZXL = "ZX8fee833329840b773103ce26bc557784ba667ee7"
	Address_Client1_Org4_ZXL = "ZXc7ccc0a30c84eb777e47a93782b1e911e69fda1a"
)

var (
	memberInfoMapping map[string][]byte
)

func initMemberInfoMapping() {
	memberInfoMapping = make(map[string][]byte)
	memberInfoMapping["org1"] = []byte(MemberInfo_Client1_Org1)
	memberInfoMapping["org2"] = []byte(MemberInfo_Client1_Org2)
	memberInfoMapping["org3"] = []byte(MemberInfo_Client1_Org3)
	memberInfoMapping["org4"] = []byte(MemberInfo_Client1_Org4)
}

func getMemberInfo(i int) (string, []byte) {
	idx := i % len(memberInfoMapping)
	orgName := fmt.Sprintf("org%d", idx+1)
	return orgName, memberInfoMapping[orgName]
}

func TestVerifyOptimizeChargeGasTx_OK(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   Address_Client1_Org1_ZXL,
					Value: []byte("8000"),
				},
				{
					Key:   Address_Client1_Org2_ZXL,
					Value: []byte("10000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err, nil)
}

func TestVerifyOptimizeChargeGasTx_NoChargeGasTx(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err, fmt.Errorf("charge gas tx is missing"))
}

func TestVerifyOptimizeChargeGasTx_WithWrongAccountAddress(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   "bad_account_address",
					Value: []byte("8000"),
				},
				{
					Key:   Address_Client1_Org2_ZXL,
					Value: []byte("10000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err, fmt.Errorf("missing some account to charge gas => `%v`", Address_Client1_Org1_ZXL))
}

func TestVerifyOptimizeChargeGasTx_WithLessAddress(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   Address_Client1_Org2_ZXL,
					Value: []byte("10000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err, fmt.Errorf("gas need to charging is not correct, expect 4 account, got 3 account"))
}

func TestVerifyOptimizeChargeGasTx_WithMoreAddress(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   "additional_address_to_charge",
					Value: []byte("8000"),
				},
				{
					Key:   Address_Client1_Org1_ZXL,
					Value: []byte("8000"),
				},
				{
					Key:   Address_Client1_Org2_ZXL,
					Value: []byte("10000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err, fmt.Errorf("gas need to charging is not correct, expect 4 account, got 5 account"))
}

func TestVerifyOptimizeChargeGasTx_WithWrongGas(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   Address_Client1_Org1_ZXL,
					Value: []byte("9000"),
				},
				{
					Key:   Address_Client1_Org2_ZXL,
					Value: []byte("10000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err,
		fmt.Errorf("gas to charge error for address `%v`, expect %v, got %v",
			Address_Client1_Org1_ZXL, 8000, 9000))
}

func TestVerifyOptimizeChargeGasTx_WithNotExistAddress(t *testing.T) {
	// construct test data
	initMemberInfoMapping()

	// contract tx collection for block
	txs := make([]*commonPb.Transaction, 0, 10)
	for i := 0; i < 10; i++ {
		org, memberInfo := getMemberInfo(i)
		tx := &commonPb.Transaction{
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      org,
					MemberType: acPb.MemberType_CERT,
					MemberInfo: memberInfo,
				},
			},
			Payload: &commonPb.Payload{
				ContractName: "TestContract",
				Method:       "TestMethod",
			},
			Result: &commonPb.Result{
				ContractResult: &commonPb.ContractResult{
					GasUsed: uint64(1000 * i),
				},
			},
		}
		txs = append(txs, tx)
	}

	// construct charge_gas tx for block
	tx := &commonPb.Transaction{
		Sender: &commonPb.EndorsementEntry{
			Signer: &acPb.Member{
				OrgId:      "org1",
				MemberType: acPb.MemberType_CERT,
				MemberInfo: []byte(MemberInfo_Node1_Org1),
			},
		},
		Payload: &commonPb.Payload{
			ContractName: syscontract.SystemContract_ACCOUNT_MANAGER.String(),
			Method:       syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String(),
			Parameters: []*commonPb.KeyValuePair{
				{
					Key:   Address_Client1_Org1_ZXL,
					Value: []byte("8000"),
				},
				{
					Key:   "address_not_exists",
					Value: []byte("9000"),
				},
				{
					Key:   Address_Client1_Org3_ZXL,
					Value: []byte("12000"),
				},
				{
					Key:   Address_Client1_Org4_ZXL,
					Value: []byte("15000"),
				},
			},
		},
		Result: &commonPb.Result{},
	}
	txs = append(txs, tx)

	// construct block
	block := &commonPb.Block{
		Txs: txs,
	}

	// init mock data
	ctl := gomock.NewController(t)
	mockChainConfig := &configPb.ChainConfig{
		Vm: &configPb.Vm{
			AddrType: configPb.AddrType_ZXL,
		},
		Crypto: &configPb.CryptoConfig{
			Hash: crypto2.CRYPTO_ALGO_SHA256,
		},
	}
	mockBlockchainStore := mock2.NewMockBlockchainStore(ctl)
	mockBlockchainStore.EXPECT().GetLastChainConfig().Return(mockChainConfig, nil).AnyTimes()
	mockSnapshot := mock2.NewMockSnapshot(ctl)
	mockSnapshot.EXPECT().GetBlockchainStore().Return(mockBlockchainStore).AnyTimes()

	err := VerifyOptimizeChargeGasTx(block, mockSnapshot)
	assert.Equal(t, err,
		fmt.Errorf("missing some account to charge gas => `%v`", Address_Client1_Org2_ZXL))
}
