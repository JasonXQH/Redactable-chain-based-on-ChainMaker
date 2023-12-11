/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package accesscontrol

import (
	"reflect"

	"chainmaker.org/chainmaker/common/v2/msgbus"

	"chainmaker.org/chainmaker/protocol/v2"
)

func init() {
	RegisterACProvider(protocol.PermissionedWithCert, nilCertACProvider)
	RegisterACProvider(protocol.Identity, nilCertACProvider)
	RegisterACProvider(protocol.PermissionedWithKey, nilPermissionedPkACProvider)
	RegisterACProvider(protocol.Public, nilPkACProvider)
}

var acProviderRegistry = map[string]reflect.Type{}

// ACProvider is an interface of ac initialize for different ac implementation
type ACProvider interface {
	NewACProvider(chainConf protocol.ChainConf, localOrgId string,
		store protocol.BlockchainStore, log protocol.Logger, msgBus msgbus.MessageBus) (protocol.AccessControlProvider, error)
}

// RegisterACProvider registers a ACProvider to global ac registry
func RegisterACProvider(authType string, acp ACProvider) {
	_, found := acProviderRegistry[authType]
	if found {
		panic("accesscontrol provider[" + authType + "] already registered!")
	}
	acProviderRegistry[authType] = reflect.TypeOf(acp)
}

// NewACProviderByMemberType returns a ACProvider by authType
func NewACProviderByMemberType(authType string) ACProvider {
	t, found := acProviderRegistry[authType]
	if !found {
		panic("accesscontrol provider[" + authType + "] not found!")
	}
	return reflect.New(t).Elem().Interface().(ACProvider)
}
