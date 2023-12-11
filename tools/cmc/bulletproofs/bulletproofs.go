/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package bulletproofs Bulletproofs中文意思为“防弹证明”，所以也称为防弹证明协议，防弹证明是由Bootle在2016年提出的，
//基于离散对数的更节省空间的零知识证明。
package bulletproofs

import (
	"github.com/spf13/cobra"
)

var (
	// genOpeningCMD flags
	// proveCMD flags
	openingStr string
	valueX     int64

	// proveMethodCMD flags
	commitmentMethod string
	valueY           int64
	commitmentXStr   string
	commitmentYStr   string
	openingXStr      string
	openingYStr      string

	// pedersenMethodCMD
	// openingMethodCMD
	// commitmentMethodCmd
	pedersenNegMethod string
)

// BulletproofsCMD new bullet proofs command
func BulletproofsCMD() *cobra.Command {
	bulletproofsCmd := &cobra.Command{
		Use:   "bulletproofs",
		Short: "ChainMaker bulletproofs command",
		Long:  "ChainMaker bulletproofs command",
	}
	// generate opening
	bulletproofsCmd.AddCommand(genOpeningCMD())

	// generate proof, commitment, opening
	bulletproofsCmd.AddCommand(proveCMD())

	// prove method
	bulletproofsCmd.AddCommand(proveMethodCMD())

	// Verify the validity of a commitment with respect to a value-opening pair
	bulletproofsCmd.AddCommand(pedersenVerifyCMD())

	bulletproofsCmd.AddCommand(pedersenNegCMD())

	return bulletproofsCmd
}
