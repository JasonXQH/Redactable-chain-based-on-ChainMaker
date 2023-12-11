/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"
	"strconv"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

// alterAddrTypeCMD alter addr type, eg. ChainMaker type, ZXL type etc.
// @return *cobra.Command
func alterAddrTypeCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alter-addr-type",
		Short: "alter addr type, eg. ChainMaker type, ZXL type etc.",
		Long:  "alter addr type, eg. ChainMaker type, ZXL type etc.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return alterAddrType()
		},
	}

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagUserTlsKeyFilePath, flagUserTlsCrtFilePath, flagOrgId,
		flagChainId, flagTimeout, flagSyncResult, flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagAdminOrgIds,
	})

	util.AttachFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagAddressType,
	})

	return cmd
}

func alterAddrType() error {

	client, err := util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	adminKeys, adminCrts, adminOrgs, err := util.MakeAdminInfo(client, adminKeyFilePaths, adminCrtFilePaths, adminOrgIds)
	if err != nil {
		return err
	}

	payload, err := client.CreateChainConfigAlterAddrTypePayload(strconv.FormatInt(int64(addressType), 10))
	if err != nil {
		return err
	}

	endorsementEntrys, err := util.MakeEndorsement(adminKeys, adminCrts, adminOrgs, client, payload)
	if err != nil {
		return err
	}

	// send
	resp, err := client.SendChainConfigUpdateRequest(payload, endorsementEntrys, timeout, syncResult)
	if err != nil {
		return err
	}

	err = util.CheckProposalRequestResp(resp, false)
	if err != nil {
		return err
	}

	output, err := prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
