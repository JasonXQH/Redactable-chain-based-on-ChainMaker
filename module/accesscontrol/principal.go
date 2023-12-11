/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

var _ protocol.Principal = (*principal)(nil)

type principal struct {
	resourceName string
	endorsement  []*common.EndorsementEntry
	message      []byte

	targetOrg string
}

// GetResourceName returns principal resource name
func (p *principal) GetResourceName() string {
	return p.resourceName
}

// GetEndorsement returns principal endorsement
func (p *principal) GetEndorsement() []*common.EndorsementEntry {
	return p.endorsement
}

// GetMessage returns principal message
func (p *principal) GetMessage() []byte {
	return p.message
}

// GetTargetOrgId returns principal target orgId
func (p *principal) GetTargetOrgId() string {
	return p.targetOrg
}
