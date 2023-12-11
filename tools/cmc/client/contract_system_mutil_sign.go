package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// ParamMultiSign define multi sign param
type ParamMultiSign struct {
	Key    string
	Value  string
	IsFile bool
}

// systemContractMultiSignCMD system contract multi sign command
// @return *cobra.Command
func systemContractMultiSignCMD() *cobra.Command {
	systemContractMultiSignCmd := &cobra.Command{
		Use:   "multi-sign",
		Short: "system contract multi sign command",
		Long:  "system contract multi sign command",
	}

	systemContractMultiSignCmd.AddCommand(multiSignReqCMD())
	systemContractMultiSignCmd.AddCommand(multiSignVoteCMD())
	systemContractMultiSignCmd.AddCommand(multiSignQueryCMD())
	systemContractMultiSignCmd.AddCommand(multiSignTrigCMD())

	return systemContractMultiSignCmd
}

// multiSignReqCMD multi sign req
// @return *cobra.Command
func multiSignReqCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "req",
		Short: "multi sign req",
		Long:  "multi sign req",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignReq()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagAdminKeyFilePaths, flagAdminCrtFilePaths,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagParams, flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagSyncResult,
		flagGasLimit, flagPayerKeyFilePath, flagPayerCrtFilePath, flagPayerOrgId,
	})

	cmd.MarkFlagRequired(flagParams)

	return cmd
}

// multiSignVoteCMD multi sign vote
// @return *cobra.Command
func multiSignVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "multi sign vote",
		Long:  "multi sign vote",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignVote()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId, flagTxId,
		flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagIsAgree,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagSyncResult, flagAdminOrgIds, flagGasLimit,
		flagPayerKeyFilePath, flagPayerCrtFilePath, flagPayerOrgId,
	})

	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

// multiSignQueryCMD multi sign query
// @return *cobra.Command
func multiSignQueryCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "multi sign query",
		Long:  "multi sign query",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignQuery()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagTxId,
		flagTruncateModel, flagTruncateValueLen,
	})

	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

// multiSignTrigCMD multi sign trig
// @return *cobra.Command
func multiSignTrigCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trig",
		Short: "multi sign trig",
		Long:  "multi sign trig",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignTrig()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagTxId, flagSyncResult,
		flagGasLimit, flagPayerKeyFilePath, flagPayerCrtFilePath, flagPayerOrgId,
	})

	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

func multiSignReq() error {
	var (
		err     error
		output  []byte
		payload *common.Payload
		client  *sdk.ChainClient
		resp    *common.TxResponse
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()
	var pms []*ParamMultiSign
	var pairs []*common.KeyValuePair
	if params != "" {
		err = json.Unmarshal([]byte(params), &pms)
		if err != nil {
			return err
		}
	}
	for _, pm := range pms {
		if pm.IsFile {
			byteCode, err := ioutil.ReadFile(pm.Value)
			if err != nil {
				panic(err)
			}
			pairs = append(pairs, &common.KeyValuePair{
				Key:   pm.Key,
				Value: byteCode,
			})
		} else {
			pairs = append(pairs, &common.KeyValuePair{
				Key:   pm.Key,
				Value: []byte(pm.Value),
			})
		}

	}

	payload = client.CreateMultiSignReqPayloadWithGasLimit(pairs, gasLimit)

	adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(client, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
	if err != nil {
		return err
	}
	endorsers, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	var payer []*common.EndorsementEntry
	if len(payerKeyFilePath) > 0 {
		payer, err = util.MakeEndorsement([]string{payerKeyFilePath}, []string{payerCrtFilePath}, []string{payerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return err
		}
	}
	if len(payer) == 0 {
		resp, err = client.MultiSignContractReq(payload, endorsers, timeout, syncResult)
	} else {
		resp, err = client.MultiSignContractReqWithPayer(payload, endorsers, payer[0], timeout, syncResult)
	}
	if err != nil {
		return fmt.Errorf("multi sign req failed, %s", err.Error())
	}
	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func multiSignVote() error {
	var (
		adminCrt string
		adminKey string
		adminOrg string
		output   []byte
		err      error
		payload  *common.Payload
		endorser *common.EndorsementEntry
		client   *sdk.ChainClient
		resp     *common.TxResponse
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	adminCrt, adminKey, adminOrg, err = getCertOrKeyFromParams(client)
	if err != nil {
		return err
	}
	payload, err = getMultiSignReqInfo(client, txId)
	if err != nil {
		return err
	}

	if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
		endorser, err = sdkutils.MakeEndorserWithPath(adminKey, adminCrt, payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}
	} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
		endorser, err = sdkutils.MakePkEndorserWithPath(adminKey, client.GetHashType(),
			adminOrg, payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}
	} else {
		endorser, err = sdkutils.MakePkEndorserWithPath(adminKey, client.GetHashType(),
			"", payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}
	}

	var payer []*common.EndorsementEntry
	if len(payerKeyFilePath) > 0 {
		payer, err = util.MakeEndorsement([]string{payerKeyFilePath}, []string{payerCrtFilePath}, []string{payerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return err
		}
	}
	if len(payer) == 0 {
		resp, err = client.MultiSignContractVoteWithGasLimit(payload, endorser, isAgree,
			timeout, gasLimit, syncResult)
	} else {
		resp, err = client.MultiSignContractVoteWithGasLimitAndPayer(payload, endorser, payer[0], isAgree,
			timeout, gasLimit, syncResult)
	}
	if err != nil {
		return fmt.Errorf("multi sign vote failed, %s", err.Error())
	}
	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

func getCertOrKeyFromParams(client *sdk.ChainClient) (string, string, string, error) {
	var (
		adminCrt  string
		adminKey  string
		adminOrg  string
		adminKeys []string
		adminCrts []string
		adminOrgs []string
	)
	if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminCrtFilePaths != "" {
			adminCrts = strings.Split(adminCrtFilePaths, ",")
		}
		if len(adminKeys) != len(adminCrts) {
			return "", "", "", fmt.Errorf(ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminCrts))
		}
		adminKey = adminKeys[0]
		adminCrt = adminCrts[0]
	} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminOrgIds != "" {
			adminOrgs = strings.Split(adminOrgIds, ",")
		}
		if len(adminKeys) != len(adminOrgs) {
			return "", "", "", fmt.Errorf(ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminOrgs))
		}
		adminKey = adminKeys[0]
		adminOrg = adminOrgs[0]
	} else {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if len(adminKeys) == 0 {
			return "", "", "", fmt.Errorf(ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminOrgs))
		}
		adminKey = adminKeys[0]
	}
	return adminCrt, adminKey, adminOrg, nil
}

func multiSignQuery() error {
	var (
		err    error
		resp   *common.TxResponse
		client *sdk.ChainClient
		output []byte
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	params := []*common.KeyValuePair{
		{
			Key:   "truncateModel",
			Value: []byte(truncateModel),
		},
		{
			Key:   "truncateValueLen",
			Value: []byte(truncateValueLen),
		},
	}
	resp, err = client.MultiSignContractQueryWithParams(txId, params)
	if err != nil {
		return fmt.Errorf("multi sign query failed, %s", err.Error())
	}
	if resp.ContractResult.Result == nil {
		return fmt.Errorf("multi sign req does not exist, req id = %v", txId)
	}

	if resp.Code == 0 && resp.ContractResult.Code == 0 {
		multiSignInfo := &syscontract.MultiSignInfo{}
		err = proto.Unmarshal(resp.ContractResult.Result, multiSignInfo)
		if err != nil {
			return err
		}
		output, err = prettyjson.Marshal(multiSignInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func multiSignTrig() error {
	var (
		err    error
		resp   *common.TxResponse
		client *sdk.ChainClient
		output []byte

		payload *common.Payload
		limit   *common.Limit
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	payload, err = getMultiSignReqInfo(client, txId)
	if err != nil {
		return err
	}

	if gasLimit > 0 {
		limit = &common.Limit{
			GasLimit: gasLimit,
		}
	}

	var payer []*common.EndorsementEntry
	if len(payerKeyFilePath) > 0 {
		payer, err = util.MakeEndorsement([]string{payerKeyFilePath}, []string{payerCrtFilePath}, []string{payerOrgId},
			client, payload)
		if err != nil {
			fmt.Printf("MakePayerEndorsement failed, %s", err)
			return err
		}
	}
	if len(payer) == 0 {
		resp, err = client.MultiSignContractTrig(payload, timeout, limit, syncResult)
	} else {
		resp, err = client.MultiSignContractTrigWithPayer(payload, payer[0], timeout, limit, syncResult)
	}
	if err != nil {
		return fmt.Errorf("multi sign trig failed, %s", err.Error())
	}
	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

func getMultiSignReqInfo(client *sdk.ChainClient, txId string) (
	*common.Payload, error) {
	resp, err := client.MultiSignContractQuery(txId)
	if err != nil {
		return nil, fmt.Errorf("get tx by txid failed, %s", err.Error())
	}
	if resp.ContractResult.Result == nil {
		return nil, fmt.Errorf("multi sign req does not exist, req id = %v", txId)
	}
	multiSignInfo := &syscontract.MultiSignInfo{}
	proto.Unmarshal(resp.ContractResult.Result, multiSignInfo)
	payload := multiSignInfo.Payload
	if payload == nil {
		return nil, fmt.Errorf("multi sign req info has not 'payload' field, req id = %v", txId)
	}

	return payload, nil
}
