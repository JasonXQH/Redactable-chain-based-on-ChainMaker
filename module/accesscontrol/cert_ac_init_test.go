/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chainmaker.org/chainmaker/pb-go/v2/config"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"github.com/stretchr/testify/require"
)

const (
	testChainId      = "chain1"
	testVersion      = "v1.0.0"
	testCertAuthType = "permissionedwithcert"
	testHashType     = "SM3"
	testPKHashType   = "SHA256"

	testOrg1 = "org1.chainmaker.org"
	testOrg2 = "org2.chainmaker.org"
	testOrg3 = "org3.chainmaker.org"
	testOrg4 = "org4.chainmaker.org"
	testOrg5 = "org5"

	tempOrg1KeyFileName  = "org1.key"
	tempOrg1CertFileName = "org1.crt"

	testConsensusRole = protocol.Role("CONSENSUS")

	testConsensusCN = "consensus1.sign.org1.chainmaker.org"

	testMsg = "Winter is coming."

	testCAOrg1 = `-----BEGIN CERTIFICATE-----
MIICjzCCAjWgAwIBAgIDB6ZbMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
MS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTMzMDcwMzAz
NTgxNlowgYQxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcxLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewlyb290LWNlcnQxHzAdBgNVBAMTFmNhLm9yZzEuY2hhaW5tYWtlci5vcmcwWTAT
BgcqhkjOPQIBBggqgRzPVQGCLQNCAAQbBh0kV51bDcM3UGHdNLI9pnIBEdkQdiAk
iGBdlBIFsIZDEDgs/pSz9eAClTOKkGwcATSL1Dpje4jLOw7aGaLIo4GTMIGQMA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCCVSqMelvIP
UItOEcnHz2vFG3f6e+NXn0MmsXxf6dk2wDBCBgNVHREEOzA5gg5jaGFpbm1ha2Vy
Lm9yZ4IJbG9jYWxob3N0ghZjYS5vcmcxLmNoYWlubWFrZXIub3JnhwR/AAABMAoG
CCqBHM9VAYN1A0gAMEUCIQC4By1zKqM+Cys8+LnW2i2RsL6hCUc5HymBDmFwwfMu
hgIgGszKP5Fg608iyLjhcqKpDeVdB5ExXAF/h0uPhFHzfTg=
-----END CERTIFICATE-----
`
	testCAOrg2 = `-----BEGIN CERTIFICATE-----
MIICjzCCAjWgAwIBAgIDCFXXMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
Mi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTMzMDcwMzAz
NTgxNlowgYQxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcyLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewlyb290LWNlcnQxHzAdBgNVBAMTFmNhLm9yZzIuY2hhaW5tYWtlci5vcmcwWTAT
BgcqhkjOPQIBBggqgRzPVQGCLQNCAARQdaAbha2vjort1n8/aekrhi+c/VS2SCyl
FMqTutL3VJ5Mfj07j7c4PkNaAjCjFVhq7YYUc+/Xxe1PKuL9edFKo4GTMIGQMA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCAmxn3WMSHu
R9vrQNFfJ07OZEiud1b73uwwPlwaGLC96jBCBgNVHREEOzA5gg5jaGFpbm1ha2Vy
Lm9yZ4IJbG9jYWxob3N0ghZjYS5vcmcyLmNoYWlubWFrZXIub3JnhwR/AAABMAoG
CCqBHM9VAYN1A0gAMEUCIQDy3RJBhjMEs8ayY+jNOmCzwlXaFeHdWAdhaVJ05SWQ
YwIgYc2txPMpPqP0VH4fk996V9HuJvUj9qp+BJ8U0l26xx4=
-----END CERTIFICATE-----
`
	testCAOrg3 = `-----BEGIN CERTIFICATE-----
MIICjzCCAjWgAwIBAgIDBYMhMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
My5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmczLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTMzMDcwMzAz
NTgxNlowgYQxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmczLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewlyb290LWNlcnQxHzAdBgNVBAMTFmNhLm9yZzMuY2hhaW5tYWtlci5vcmcwWTAT
BgcqhkjOPQIBBggqgRzPVQGCLQNCAAR+P5CYHJorHxlIfb8sIOQH5Nsrfk/JdY5W
dkVBH88ZNP0j030r6lTfuKtR/XXq89nHSXuRr+YMybRsDHGRdgHOo4GTMIGQMA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCBSIUJt5yN6
vrqkCGPoDbAgFyKfLpAgWKfwQmzzxM1tdjBCBgNVHREEOzA5gg5jaGFpbm1ha2Vy
Lm9yZ4IJbG9jYWxob3N0ghZjYS5vcmczLmNoYWlubWFrZXIub3JnhwR/AAABMAoG
CCqBHM9VAYN1A0gAMEUCIF131R4vgomTDJ3bVu+POeZ/omAO0CKJyEyTLSw276wp
AiEApn6aujIXCYh+5kFNxTmhvGr0fwrl5JHyL61v48fqo18=
-----END CERTIFICATE-----
`
	testCAOrg4 = `-----BEGIN CERTIFICATE-----
MIICkDCCAjWgAwIBAgIDA00UMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
NC5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmc0LmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTMzMDcwMzAz
NTgxNlowgYQxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmc0LmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewlyb290LWNlcnQxHzAdBgNVBAMTFmNhLm9yZzQuY2hhaW5tYWtlci5vcmcwWTAT
BgcqhkjOPQIBBggqgRzPVQGCLQNCAASGq7wWxrQK8VbBEiJ6JGKZjU5BRqcg+jVa
trKrwShTda1p7gjC0SsKX+bsmuJMHoDIXKtSM9ch5ObJuI50qRd5o4GTMIGQMA4G
A1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCDcyHU1iuYf
JZT+ooLW9+xg4JheiLetZXD5w4Jz273sTTBCBgNVHREEOzA5gg5jaGFpbm1ha2Vy
Lm9yZ4IJbG9jYWxob3N0ghZjYS5vcmc0LmNoYWlubWFrZXIub3JnhwR/AAABMAoG
CCqBHM9VAYN1A0kAMEYCIQCGo1Isyz5PmpdyGtqV3+ab5RD6+wVMcwvAPlbORAUL
QgIhAJmTYuIkrLqGWkPo3Pm2+0htkXqPIoGGKOr5Yf5wpvvR
-----END CERTIFICATE-----
`
	testTrustMember1 = `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIKVbkVBlA0XYwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQwNDda
Fw0yMzA4MDUwMzQwNDdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4xMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAELf71
DQTS9zpzs3nDDdt6ncocPHrlqdpZvobToTNPeYmrIFBuahrokQZ14CvxZP632KJk
ohAlGfAfoxsdciuIiaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCAuvneJ
0X1P7/K9yZRF+I0VEEWrFWTmqkq4In9l45GAFTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJUmhAHycQXCV68HnQvF761kE5157fXoQB6huFKBj1ySAiEA87/G
VF6kotuIP24ujAzANvkoZJeOhpk1hVS2xdIZ86s=
-----END CERTIFICATE-----
`
	testTrustMember2 = `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIILBJts5OBl+8wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQyMDVa
Fw0yMzA4MDUwMzQyMDVaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4yMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAExoxt
S//rEqnhj6/ZNxxmfFY767XyeZrbrxewTtYqLZJYwOik3CsVhSsrelgAdsBOG4Pe
o7eCet9lxpq2NM/XBKN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDbJtdv
0Krwa+vHyNB2urb8XC54OFy8oAwfvTbKc9l8iTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJZqy5zdbQ3ZfAJci7QcKLkMNXoqz2VHxH0QXz26uDvUAiEAgmIc
Ds20ILx7wy349jvs8s4Rc1P4hJZQdfkxdI2GhXU=
-----END CERTIFICATE-----
`
)

var testChainConfig = &config.ChainConfig{
	ChainId:  testChainId,
	Version:  testVersion,
	AuthType: testCertAuthType,
	Sequence: 0,
	Crypto: &config.CryptoConfig{
		Hash: testHashType,
	},
	Block: nil,
	Core:  nil,
	Consensus: &config.ConsensusConfig{
		Type: 0,
		Nodes: []*config.OrgConfig{{
			OrgId:  testOrg1,
			NodeId: nil,
		}, {
			OrgId:  testOrg2,
			NodeId: nil,
		}, {
			OrgId:  testOrg3,
			NodeId: nil,
		}, {
			OrgId:  testOrg4,
			NodeId: nil,
		},
		},
		ExtConfig: nil,
	},
	TrustRoots: []*config.TrustRootConfig{
		{
			OrgId: testOrg1,
			Root:  []string{testCAOrg1},
		},
		{
			OrgId: testOrg2,
			Root:  []string{testCAOrg2},
		},
		{
			OrgId: testOrg3,
			Root:  []string{testCAOrg3},
		},
		{
			OrgId: testOrg4,
			Root:  []string{testCAOrg4},
		},
	},
	TrustMembers: []*config.TrustMemberConfig{
		{
			OrgId:      testOrg5,
			Role:       "admin",
			MemberInfo: testTrustMember1,
		},
		{
			OrgId:      testOrg5,
			Role:       "admin",
			MemberInfo: testTrustMember2,
		},
	},
}

type testCertInfo struct {
	cert string
	sk   string
}

var testConsensusSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICcTCCAhigAwIBAgIDAzPiMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
MS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcxLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewljb25zZW5zdXMxLDAqBgNVBAMTI2NvbnNlbnN1czEuc2lnbi5vcmcxLmNoYWlu
bWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEZ0zzMkjELU9/6Pn4
cfRONS0bPtXdlcZcQP7C1etFlkjVUomb0FYqgZb9UjVaEy179mcAue/rplPysRTV
NNkRnaNqMGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCBZmKTqj8XcF8RQ0UWJ
MBkZPVZb2rjmC1FyyTLmDu1ojzArBgNVHSMEJDAigCCVSqMelvIPUItOEcnHz2vF
G3f6e+NXn0MmsXxf6dk2wDAKBggqgRzPVQGDdQNHADBEAiAeTKQwZgQ8ZBtEf7+V
uI8by7QMqz0ebs2gMG9P0pBQqgIgdWgbb5ftmSbxulV6cvJDaACIDsdR2U6hlgPR
lPcUrp4=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgvroSOy0LI2pGtalr
ER1nEig6tuLd5SlV+4qQLkF12G6gCgYIKoEcz1UBgi2hRANCAARnTPMySMQtT3/o
+fhx9E41LRs+1d2VxlxA/sLV60WWSNVSiZvQViqBlv1SNVoTLXv2ZwC57+umU/Kx
FNU02RGd
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICczCCAhigAwIBAgIDDNkLMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
Mi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcyLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewljb25zZW5zdXMxLDAqBgNVBAMTI2NvbnNlbnN1czEuc2lnbi5vcmcyLmNoYWlu
bWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEP1ihjBwbZWDand1L
DvXmGTQrvJD97ltyGDzK8ECIm6Hj9v7uMzWr9Ho8GsFx7iLj51eqTsvsr/21AmvJ
dMVrsKNqMGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCAwz6gszwZm/ki+fiuP
XfjEFdanJ7ui4Y2ODtgtv4woYDArBgNVHSMEJDAigCAmxn3WMSHuR9vrQNFfJ07O
ZEiud1b73uwwPlwaGLC96jAKBggqgRzPVQGDdQNJADBGAiEAhyBc+LcbkJW5cfkS
t6xCrhcQw6yYKJvyv4L1MXg7XdsCIQCkXp9Dpy8fYdVDkPyBzNhqPIBOL8mC12y8
ICSGmGTOmA==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgm0Y4b2FVjdoX5NYk
mk60O88r4mkypu02lipqzuqg2fGgCgYIKoEcz1UBgi2hRANCAAQ/WKGMHBtlYNqd
3UsO9eYZNCu8kP3uW3IYPMrwQIiboeP2/u4zNav0ejwawXHuIuPnV6pOy+yv/bUC
a8l0xWuw
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICcjCCAhigAwIBAgIDCYqOMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
My5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmczLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmczLmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewljb25zZW5zdXMxLDAqBgNVBAMTI2NvbnNlbnN1czEuc2lnbi5vcmczLmNoYWlu
bWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE+g7F+rqm35vbT+kh
FZHEl02WXDeKviydeW2N1P2n+z69OzfKGr8qsDdOzWKIF69S+FkkeONhqVmezKSt
fzFixaNqMGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCC5K2p1+C1SelhMEM+C
H9T2nqggtWSuXFVHosXPPCJE/jArBgNVHSMEJDAigCBSIUJt5yN6vrqkCGPoDbAg
FyKfLpAgWKfwQmzzxM1tdjAKBggqgRzPVQGDdQNIADBFAiBGBvI0+E2AxOeMuz2M
6ZVBdDwSpBlOIjbZdEeTthf/VQIhANbZMGQqNW3JVWlZFhCFHCYxLD2fPEqa3yS3
0EVf5MQ5
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg/Sz/T5s5he19NP9M
XnJpjUG2yuZHzM/dAg8ZHffj4UmgCgYIKoEcz1UBgi2hRANCAAT6DsX6uqbfm9tP
6SEVkcSXTZZcN4q+LJ15bY3U/af7Pr07N8oavyqwN07NYogXr1L4WSR442GpWZ7M
pK1/MWLF
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICczCCAhigAwIBAgIDCGoFMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
NC5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmc0LmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgZExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmc0LmNoYWlubWFrZXIub3JnMRIwEAYDVQQL
Ewljb25zZW5zdXMxLDAqBgNVBAMTI2NvbnNlbnN1czEuc2lnbi5vcmc0LmNoYWlu
bWFrZXIub3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE21DGMQenGKVslnTt
KwE+oHsy4GtaMMxkPCTMXhzRgAaKNrzeq3+Bq7ul6ArNjqDgtjLBaz8hnZcXcgA3
hYrQuaNqMGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCBsY3hItn5LNzHbHFpR
goey8ADWKqMrSF/KqTR+RHRLKDArBgNVHSMEJDAigCDcyHU1iuYfJZT+ooLW9+xg
4JheiLetZXD5w4Jz273sTTAKBggqgRzPVQGDdQNJADBGAiEAwzlU1cXA19MfAC2B
+AacD4QRMOLrv9cWhw5oXowJ7HsCIQDyc5Ik4XxTeQwjgwSyl87FvfC6Yp/ZWff7
3/Kggzj2Mg==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgy2jkmFbNwNx78niz
reu+yUXrA5ntYKJeLMZgHtqIkkKgCgYIKoEcz1UBgi2hRANCAATbUMYxB6cYpWyW
dO0rAT6gezLga1owzGQ8JMxeHNGABoo2vN6rf4Gru6XoCs2OoOC2MsFrPyGdlxdy
ADeFitC5
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICaTCCAhCgAwIBAgIDBfCEMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
MS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYkxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcxLmNoYWlubWFrZXIub3JnMQ4wDAYDVQQL
EwVhZG1pbjEoMCYGA1UEAxMfYWRtaW4xLnNpZ24ub3JnMS5jaGFpbm1ha2VyLm9y
ZzBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABIpohsbx8lCyRSiZatb4V0N1+jrZ
0y2tXrj3diYtYCScb50yM7Mva8Uj7H+6/lPUUBo2RghoN5FWmvobYgLofGKjajBo
MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4EIgQgIlFMy4ihwlYA0z+rVP8EXzoJav1c
PzPZVdQVMJodB0QwKwYDVR0jBCQwIoAglUqjHpbyD1CLThHJx89rxRt3+nvjV59D
JrF8X+nZNsAwCgYIKoEcz1UBg3UDRwAwRAIgZzJyKLf1hbFAub4AwCI8CLbQQMFY
8acUwkLjPeppUd8CIHgZcaBVRcVqRIJFNdX0TIggRF4qu6V4zjHiJ26uc2xV
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgYlI8IiM6+rzZLAxX
uemdEfEN5XztOz0lW7Xa5oDedqugCgYIKoEcz1UBgi2hRANCAASKaIbG8fJQskUo
mWrW+FdDdfo62dMtrV6493YmLWAknG+dMjOzL2vFI+x/uv5T1FAaNkYIaDeRVpr6
G2IC6Hxi
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICaTCCAhCgAwIBAgIDCPBBMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
Mi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYkxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcyLmNoYWlubWFrZXIub3JnMQ4wDAYDVQQL
EwVhZG1pbjEoMCYGA1UEAxMfYWRtaW4xLnNpZ24ub3JnMi5jaGFpbm1ha2VyLm9y
ZzBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABGnywdsfJfVfgcbkXkEpuHxrdzwQ
cYxvl1sPLuK3F2TGGk/PeyXrBUWDQPxZxTtlypbhAoaitgj0DEKHCJ/DCEujajBo
MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4EIgQgPVudmYwTB4RxEK6QBFqWIBbeb1fR
yx+B8U6ZBYZPXJ0wKwYDVR0jBCQwIoAgJsZ91jEh7kfb60DRXydOzmRIrndW+97s
MD5cGhiwveowCgYIKoEcz1UBg3UDRwAwRAIgVlrHlY+VlI4YEz4nx6RTSv2gDIa5
3QVNo0qFwRHtyy0CIGlNFnHrivMoECAu1kEOYYCwL05HjLt/7zZXuqxCiX3p
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgWEzV5ozbCiXEQ+b1
lxfwYXAiR/TVKWGlosT9KQ9pIHKgCgYIKoEcz1UBgi2hRANCAARp8sHbHyX1X4HG
5F5BKbh8a3c8EHGMb5dbDy7itxdkxhpPz3sl6wVFg0D8WcU7ZcqW4QKGorYI9AxC
hwifwwhL
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICaTCCAg+gAwIBAgICOiAwCgYIKoEcz1UBg3UwgYQxCzAJBgNVBAYTAkNOMRAw
DgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQHEwdCZWlqaW5nMRwwGgYDVQQKExNvcmcz
LmNoYWlubWFrZXIub3JnMRIwEAYDVQQLEwlyb290LWNlcnQxHzAdBgNVBAMTFmNh
Lm9yZzMuY2hhaW5tYWtlci5vcmcwHhcNMjMwNzA2MDM1ODE2WhcNMjgwNzA0MDM1
ODE2WjCBiTELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcT
B0JlaWppbmcxHDAaBgNVBAoTE29yZzMuY2hhaW5tYWtlci5vcmcxDjAMBgNVBAsT
BWFkbWluMSgwJgYDVQQDEx9hZG1pbjEuc2lnbi5vcmczLmNoYWlubWFrZXIub3Jn
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEWkdLeoBqDEg8D9GF9ZDsJOV0fYhb
krqZFDZFpWEU9hHn+I3r/WHx76fbgg4/QUW3Eocln0miwZ/BHiBYSbR6yKNqMGgw
DgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDgnoeF41S0w5HsHIoNhxzaWFd4oC3n
8HdzHaZEGhOlLDArBgNVHSMEJDAigCBSIUJt5yN6vrqkCGPoDbAgFyKfLpAgWKfw
QmzzxM1tdjAKBggqgRzPVQGDdQNIADBFAiEAgW3AztZKo27QRm93srjh4gEyJ9eK
6JcIAFh9f4yjDJgCIFFHfZo7PUBs4YMwZ55L8SiDWvYArKleTcsLNQHQUSk+
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgQLvl4CiHBGh1tgQw
3skHzhP1RA1zPcurBADI7sJh4zigCgYIKoEcz1UBgi2hRANCAARaR0t6gGoMSDwP
0YX1kOwk5XR9iFuSupkUNkWlYRT2Eef4jev9YfHvp9uCDj9BRbcShyWfSaLBn8Ee
IFhJtHrI
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICaTCCAhCgAwIBAgIDBWPgMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
NC5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmc0LmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYkxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmc0LmNoYWlubWFrZXIub3JnMQ4wDAYDVQQL
EwVhZG1pbjEoMCYGA1UEAxMfYWRtaW4xLnNpZ24ub3JnNC5jaGFpbm1ha2VyLm9y
ZzBZMBMGByqGSM49AgEGCCqBHM9VAYItA0IABEGE+2P7GTC8ltsvFhoDVRn6OiwA
traNc8i6oaoJ6XP4t2DhuDNtT658dnerwWnlaEO0nS9vVt2N9G8KWbzGK4WjajBo
MA4GA1UdDwEB/wQEAwIGwDApBgNVHQ4EIgQgL7TJCSrRX/HWI5t9jSJOH6HtUA+X
x4xIQxGTv7MyblYwKwYDVR0jBCQwIoAg3Mh1NYrmHyWU/qKC1vfsYOCYXoi3rWVw
+cOCc9u97E0wCgYIKoEcz1UBg3UDRwAwRAIgLoL9orIh8fu5NzzQKreOCXVTkALd
5p02UoGKn8guPFECIBvxUfBdZD9qLp2leNvWpYKIVYoSO0EvxQ1YBtk/NV1t
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgbB8khiponsJ9FCxA
/NzfuDBToyy7s+bjVm8NSn1afbOgCgYIKoEcz1UBgi2hRANCAARBhPtj+xkwvJbb
LxYaA1UZ+josALa2jXPIuqGqCelz+Ldg4bgzbU+ufHZ3q8Fp5WhDtJ0vb1bdjfRv
Clm8xiuF
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICbDCCAhKgAwIBAgIDA3/kMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
MS5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcxLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYsxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcxLmNoYWlubWFrZXIub3JnMQ8wDQYDVQQL
EwZjbGllbnQxKTAnBgNVBAMTIGNsaWVudDEuc2lnbi5vcmcxLmNoYWlubWFrZXIu
b3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAErNx3auvSIV6EiLT5zn7k+Gsd
5NfLWqjYc0/421yY7RPIJwwiu8Y9J8gpvjIAA6jDogIMFPFcI4iahuQnIFf44aNq
MGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCCV6NCXxdhg3J8l/P4QIBfo6RvI
FjOQo5u/9VN6KG2yczArBgNVHSMEJDAigCCVSqMelvIPUItOEcnHz2vFG3f6e+NX
n0MmsXxf6dk2wDAKBggqgRzPVQGDdQNIADBFAiB5ohBX5JBZDRnObxFf33ycBkqT
ClSzTnhA1xjTD0+jKAIhAONLFLnAsuZiRILzjcCV6DO1+n1XvyeoJcASGqJjAe1R
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg5zWMqi1JOQ3iwVCV
lP1t5HqSSIskzQw5yJFjaDrsUjOgCgYIKoEcz1UBgi2hRANCAASs3Hdq69IhXoSI
tPnOfuT4ax3k18taqNhzT/jbXJjtE8gnDCK7xj0nyCm+MgADqMOiAgwU8VwjiJqG
5CcgV/jh
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICazCCAhKgAwIBAgIDCE3LMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
Mi5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmcyLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYsxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmcyLmNoYWlubWFrZXIub3JnMQ8wDQYDVQQL
EwZjbGllbnQxKTAnBgNVBAMTIGNsaWVudDEuc2lnbi5vcmcyLmNoYWlubWFrZXIu
b3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE+xXAKN1912lp5r5f3rJfjTPN
Kjr1pe85jCQ0syWnHbT8QaPc6kqFp4xt4xFKegudext5kit80TXw0+lfV8XfLqNq
MGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCA3aFzE3M8TJBzfTJPrMbBbzIOZ
CwXk5CI7AQnXMs2MwTArBgNVHSMEJDAigCAmxn3WMSHuR9vrQNFfJ07OZEiud1b7
3uwwPlwaGLC96jAKBggqgRzPVQGDdQNHADBEAiAgJe55Ewi6hrn61Df0ctwW/+ED
3ElhxG2VKhVxEmhDbQIgbU2UwBz9PJ5u7uTVX7QjbGdsKM6bLCiYg2BvzTn7ffw=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg20J0Zr7G8YPdiqSV
COKB+sMrFKP8rm7Dp7katWd46/SgCgYIKoEcz1UBgi2hRANCAAT7FcAo3X3XaWnm
vl/esl+NM80qOvWl7zmMJDSzJacdtPxBo9zqSoWnjG3jEUp6C517G3mSK3zRNfDT
6V9Xxd8u
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg3 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICazCCAhKgAwIBAgIDDJPGMAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
My5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmczLmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYsxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmczLmNoYWlubWFrZXIub3JnMQ8wDQYDVQQL
EwZjbGllbnQxKTAnBgNVBAMTIGNsaWVudDEuc2lnbi5vcmczLmNoYWlubWFrZXIu
b3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEWInAsRStv88IKQ0DbgZaJlP/
P38FNrEz97jwJNzcEesa70CcGzr+gwxnic7eae83ltSKzRhZs6DofvvZTQe9BqNq
MGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCBNEtLUdZXdJZEa7Aw1xCJwB47H
P/6Yu0hqqHgU7iY9wTArBgNVHSMEJDAigCBSIUJt5yN6vrqkCGPoDbAgFyKfLpAg
WKfwQmzzxM1tdjAKBggqgRzPVQGDdQNHADBEAiBpC302bS2RH3rsT7r2UI2CZrOM
X1F9gc8nR6ZRPZv6aQIgFmSt49juUp+i5z3bhTN26/Xp/8wFDshMmgaQqF7w8qI=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgcgwMV6sF4Y82e8pE
ALDMNqnKALDw1ITkm4x27ZTZJyygCgYIKoEcz1UBgi2hRANCAARYicCxFK2/zwgp
DQNuBlomU/8/fwU2sTP3uPAk3NwR6xrvQJwbOv6DDGeJzt5p7zeW1IrNGFmzoOh+
+9lNB70G
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg4 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICbDCCAhKgAwIBAgIDC8v8MAoGCCqBHM9VAYN1MIGEMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpamluZzEQMA4GA1UEBxMHQmVpamluZzEcMBoGA1UEChMTb3Jn
NC5jaGFpbm1ha2VyLm9yZzESMBAGA1UECxMJcm9vdC1jZXJ0MR8wHQYDVQQDExZj
YS5vcmc0LmNoYWlubWFrZXIub3JnMB4XDTIzMDcwNjAzNTgxNloXDTI4MDcwNDAz
NTgxNlowgYsxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRAwDgYDVQQH
EwdCZWlqaW5nMRwwGgYDVQQKExNvcmc0LmNoYWlubWFrZXIub3JnMQ8wDQYDVQQL
EwZjbGllbnQxKTAnBgNVBAMTIGNsaWVudDEuc2lnbi5vcmc0LmNoYWlubWFrZXIu
b3JnMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEwrS8oY0oCRTJPnLPUMXx6+iN
l9RsJvZUBk1+7HdZRzXr1DsjXoNcow8IWksqvcddtSI2grf6tzDWy0RyQSGVnaNq
MGgwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDiZusN+FdclunoUeQMA3NO9FCv
+eYCZB3T/8L8IzCmVzArBgNVHSMEJDAigCDcyHU1iuYfJZT+ooLW9+xg4JheiLet
ZXD5w4Jz273sTTAKBggqgRzPVQGDdQNIADBFAiEAgrXIv+EpakysUZ0/qhhJ8tNG
VY67tYou9y2D5L81Ae0CIH6dqMvqDFzRSg7Gv22CsB7mRpTGJTKY9TeuyWdGv3ji
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgmCiuQgaafXOoBiMA
+1wdkqSsZxMKRjJ2Z6jGK5KDdXmgCgYIKoEcz1UBgi2hRANCAATCtLyhjSgJFMk+
cs9QxfHr6I2X1Gwm9lQGTX7sd1lHNevUOyNeg1yjDwhaSyq9x121IjaCt/q3MNbL
RHJBIZWd
-----END PRIVATE KEY-----
`,
}

var testConsensusSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICOjCCAeGgAwIBAgIIGN/iRBNkA0kwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjE1NTla
Fw0yMzA4MDUxMjE1NTlaMGkxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MRIwEAYDVQQLEwljb25z
ZW5zdXMxEzARBgNVBAMTCmNvbnNlbnN1czEwWTATBgcqhkjOPQIBBggqgRzPVQGC
LQNCAATrDM8W9PU/9idSEGLXbCneUqlrY5ExNWShWg+1Qy8p1rDtwpLFTEuDR6sf
kQV8T9i1zeXefyS066zJZnhBpyJWo4GBMH8wDgYDVR0PAQH/BAQDAgbAMCkGA1Ud
DgQiBCDmj9z0hrOaUVRwkG6YlPxXarHRD37KGLU4YdOYre0aATArBgNVHSMEJDAi
gCCrqUGCeuslhFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTAVBgNVHREEDjAMggpj
b25zZW5zdXMxMAoGCCqBHM9VAYN1A0cAMEQCIEKFF/F682Ok2SO1dMUsVpKWmIBa
DagEEDWacKJ/07bAAiBevhEXM+6cqblRcTqLPQMNG6+Xz/gwcvHww8k9GspwRA==
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgkycHiKroE0z0AQmT
zX18Q2ia2YpytlMxzHJ1hxKUn8igCgYIKoEcz1UBgi2hRANCAATrDM8W9PU/9idS
EGLXbCneUqlrY5ExNWShWg+1Qy8p1rDtwpLFTEuDR6sfkQV8T9i1zeXefyS066zJ
ZnhBpyJW
-----END PRIVATE KEY-----
`,
}

var testAdminSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIJb8IwGdGzDMwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjIyMjda
Fw0yMzA4MDUxMjIyMjdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4zMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEPuP1
YyIVkKYfHQa2iR6A2E5MnnQYftIu6UhKRZI4EDT/DDs4l+2ksfTf4YeJUQqailwe
QESUFyyhXPWWKU0yDaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCCE6BKh
MBVt9KFx7LmE3wXkILhbdTX07ZY8dKTDglFCIjArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAIudP9N2PbqWyOFrJKUwW5qO51hQciQsKKyLY8YTafsRAiEAmND5
BpWsfd537YspBgQRBDg5ztVRc68wp3C4AdqWc5Q=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgvCLtqX1Ser0F8H+2
1lCKQw54umUBz97DxcCdR6/ur2CgCgYIKoEcz1UBgi2hRANCAAQ+4/VjIhWQph8d
BraJHoDYTkyedBh+0i7pSEpFkjgQNP8MOziX7aSx9N/hh4lRCpqKXB5ARJQXLKFc
9ZYpTTIN
-----END PRIVATE KEY-----
`,
}

var testClientSignOrg5 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKjCCAdCgAwIBAgIIKi+Lqj7RJ50wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUxMjIwNDVa
Fw0yMzA4MDUxMjIwNDVaMGMxCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ8wDQYDVQQLEwZjbGll
bnQxEDAOBgNVBAMTB2NsaWVudDEwWTATBgcqhkjOPQIBBggqgRzPVQGCLQNCAARU
XOhCeTsKsPBYt2xEyxYGSBY6xuNwcj0ppqcMTnH9J6javljoxDpKpNF2tcFK6CA3
/Z9j/APE7s5vkZK2W7Czo3cwdTAOBgNVHQ8BAf8EBAMCBsAwKQYDVR0OBCIEIAcw
XrkZCq8G6XkdUKqiNeQfijHC+VLKXfmtdKk3ADp5MCsGA1UdIwQkMCKAIKupQYJ6
6yWEU2vDnoJeUjWTAvmtys/FwZDpkJdt6GgxMAsGA1UdEQQEMAKCADAKBggqgRzP
VQGDdQNIADBFAiAy+gAGzbZcGIP17iKzyYBpu2qIEs9CXaM45AUelLb4QwIhAKA3
uZFq5Yw+M+1RCZm1JWYEqICxws5LW4I5vxoEM7F/
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgzwNxhFP8qFtIv2WZ
n4X2TFihYUnESnfnQ7M2kPkI//CgCgYIKoEcz1UBgi2hRANCAARUXOhCeTsKsPBY
t2xEyxYGSBY6xuNwcj0ppqcMTnH9J6javljoxDpKpNF2tcFK6CA3/Z9j/APE7s5v
kZK2W7Cz
-----END PRIVATE KEY-----
`,
}

var testTrustMemberAdmin1 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIIKVbkVBlA0XYwCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQwNDda
Fw0yMzA4MDUwMzQwNDdaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4xMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAELf71
DQTS9zpzs3nDDdt6ncocPHrlqdpZvobToTNPeYmrIFBuahrokQZ14CvxZP632KJk
ohAlGfAfoxsdciuIiaN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCAuvneJ
0X1P7/K9yZRF+I0VEEWrFWTmqkq4In9l45GAFTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJUmhAHycQXCV68HnQvF761kE5157fXoQB6huFKBj1ySAiEA87/G
VF6kotuIP24ujAzANvkoZJeOhpk1hVS2xdIZ86s=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgoSPsIukE1OdHTtwd
AJm1b4QjzOVv+B6N+M9rCb6OELSgCgYIKoEcz1UBgi2hRANCAAQt/vUNBNL3OnOz
ecMN23qdyhw8euWp2lm+htOhM095iasgUG5qGuiRBnXgK/Fk/rfYomSiECUZ8B+j
Gx1yK4iJ
-----END PRIVATE KEY-----
`,
}

var testTrustMemberAdmin2 = &testCertInfo{
	cert: `-----BEGIN CERTIFICATE-----
MIICKTCCAc6gAwIBAgIILBJts5OBl+8wCgYIKoEcz1UBg3UwWjELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaWppbmcxEDAOBgNVBAcTB0JlaWppbmcxDTALBgNVBAoT
BG9yZzUxCzAJBgNVBAsTAmNhMQswCQYDVQQDEwJjYTAeFw0yMTA4MDUwMzQyMDVa
Fw0yMzA4MDUwMzQyMDVaMGExCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5n
MRAwDgYDVQQHEwdCZWlqaW5nMQ0wCwYDVQQKEwRvcmc1MQ4wDAYDVQQLEwVhZG1p
bjEPMA0GA1UEAxMGYWRtaW4yMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAExoxt
S//rEqnhj6/ZNxxmfFY767XyeZrbrxewTtYqLZJYwOik3CsVhSsrelgAdsBOG4Pe
o7eCet9lxpq2NM/XBKN3MHUwDgYDVR0PAQH/BAQDAgbAMCkGA1UdDgQiBCDbJtdv
0Krwa+vHyNB2urb8XC54OFy8oAwfvTbKc9l8iTArBgNVHSMEJDAigCCrqUGCeusl
hFNrw56CXlI1kwL5rcrPxcGQ6ZCXbehoMTALBgNVHREEBDACggAwCgYIKoEcz1UB
g3UDSQAwRgIhAJZqy5zdbQ3ZfAJci7QcKLkMNXoqz2VHxH0QXz26uDvUAiEAgmIc
Ds20ILx7wy349jvs8s4Rc1P4hJZQdfkxdI2GhXU=
-----END CERTIFICATE-----
`,
	sk: `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgReLJhNH+Hn8RNyYO
OnXvLVHEDt500mN060ym0gl52N6gCgYIKoEcz1UBgi2hRANCAATGjG1L/+sSqeGP
r9k3HGZ8VjvrtfJ5mtuvF7BO1iotkljA6KTcKxWFKyt6WAB2wE4bg96jt4J632XG
mrY0z9cE
-----END PRIVATE KEY-----
`,
}

var testCRL = `-----BEGIN CRL-----
MIIBNjCB3AIBATAKBggqhkjOPQQDAjBfMQswCQYDVQQGEwJDTjEQMA4GA1UECBMH
QmVpamluZzEQMA4GA1UEBxMHQmVpamluZzENMAsGA1UEChMEb3JnMTELMAkGA1UE
CxMCY2ExEDAOBgNVBAMTB2NhLm9yZzgXDTIxMDYxODA4NDEyOFoXDTIxMDYxODA5
NDEyOFowGzAZAggmTrsNBHY7GBcNMjMwNjE3MTI1MTU3WqAvMC0wKwYDVR0jBCQw
IoAgYmqb9hiWkAJKn8GnweOVQtp6C7q8FNl+8qmkiGcuoIowCgYIKoZIzj0EAwID
SQAwRgIhAPmTMGrCxxhq/dvwk1HJp4h17qfeHxv1T7ETg/ri23zoAiEAnz6XlUcn
7YojGXHlfoxpmV8ao6sZIm+j4ylNoBaawOI=
-----END CRL-----`

func createTempDirWithCleanFunc() (string, func(), error) {
	var td = filepath.Join("./temp")
	err := os.MkdirAll(td, os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	var cleanFunc = func() {
		_ = os.RemoveAll(td)
		_ = os.RemoveAll(filepath.Join("./default.log"))
		now := time.Now()
		_ = os.RemoveAll(filepath.Join("./default.log." + now.Format("2006010215")))
		now = now.Add(-2 * time.Hour)
		_ = os.RemoveAll(filepath.Join("./default.log." + now.Format("2006010215")))
	}
	return td, cleanFunc, nil
}

type orgMemberInfo struct {
	orgId        string
	consensus    *testCertInfo
	admin        *testCertInfo
	client       *testCertInfo
	trustMember1 *testCertInfo
	trustMember2 *testCertInfo
}

type orgMember struct {
	orgId        string
	acProvider   protocol.AccessControlProvider
	consensus    protocol.SigningMember
	admin        protocol.SigningMember
	client       protocol.SigningMember
	trustMember1 protocol.SigningMember
	trustMember2 protocol.SigningMember
}

var orgMemberInfoMap = map[string]*orgMemberInfo{
	testOrg1: {
		orgId:     testOrg1,
		consensus: testConsensusSignOrg1,
		admin:     testAdminSignOrg1,
		client:    testClientSignOrg1,
	},
	testOrg2: {
		orgId:     testOrg2,
		consensus: testConsensusSignOrg2,
		admin:     testAdminSignOrg2,
		client:    testClientSignOrg2,
	},
	testOrg3: {
		orgId:     testOrg3,
		consensus: testConsensusSignOrg3,
		admin:     testAdminSignOrg3,
		client:    testClientSignOrg3,
	},
	testOrg4: {
		orgId:     testOrg4,
		consensus: testConsensusSignOrg4,
		admin:     testAdminSignOrg4,
		client:    testClientSignOrg4,
	},
	testOrg5: {
		orgId:        testOrg5,
		consensus:    testConsensusSignOrg5,
		admin:        testAdminSignOrg5,
		client:       testClientSignOrg5,
		trustMember1: testTrustMemberAdmin1,
		trustMember2: testTrustMemberAdmin2,
	},
}

func initOrgMember(t *testing.T, info *orgMemberInfo) *orgMember {
	td, cleanFunc, err := createTempDirWithCleanFunc()
	require.Nil(t, err)
	defer cleanFunc()
	logger := &test.GoLogger{}
	certProvider, err := newCertACProvider(testChainConfig, info.orgId, nil, logger)
	require.Nil(t, err)
	require.NotNil(t, certProvider)

	localPrivKeyFile := filepath.Join(td, info.orgId+".key")
	localCertFile := filepath.Join(td, info.orgId+".crt")

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.consensus.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.consensus.cert), os.ModePerm)
	require.Nil(t, err)
	consensus, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.admin.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.admin.cert), os.ModePerm)
	require.Nil(t, err)
	admin, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	err = ioutil.WriteFile(localPrivKeyFile, []byte(info.client.sk), os.ModePerm)
	require.Nil(t, err)
	err = ioutil.WriteFile(localCertFile, []byte(info.client.cert), os.ModePerm)
	require.Nil(t, err)
	client, err := InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
	require.Nil(t, err)

	var (
		trustMember1 protocol.SigningMember
		trustMember2 protocol.SigningMember
	)
	if info.trustMember1 != nil && info.trustMember2 != nil {
		err = ioutil.WriteFile(localPrivKeyFile, []byte(info.trustMember1.sk), os.ModePerm)
		require.Nil(t, err)
		err = ioutil.WriteFile(localCertFile, []byte(info.trustMember1.cert), os.ModePerm)
		require.Nil(t, err)
		trustMember1, err = InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
		require.Nil(t, err)

		err = ioutil.WriteFile(localPrivKeyFile, []byte(info.trustMember2.sk), os.ModePerm)
		require.Nil(t, err)
		err = ioutil.WriteFile(localCertFile, []byte(info.trustMember2.cert), os.ModePerm)
		require.Nil(t, err)
		trustMember2, err = InitCertSigningMember(testChainConfig, info.orgId, localPrivKeyFile, "", localCertFile)
		require.Nil(t, err)
	}

	return &orgMember{
		orgId:        info.orgId,
		acProvider:   certProvider,
		consensus:    consensus,
		admin:        admin,
		client:       client,
		trustMember1: trustMember1,
		trustMember2: trustMember2,
	}
}
