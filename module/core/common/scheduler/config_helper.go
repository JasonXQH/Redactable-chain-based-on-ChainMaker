package scheduler

import (
	"fmt"
	"strconv"

	configPb "chainmaker.org/chainmaker/pb-go/v2/config"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
)

func IsOptimizeChargeGasEnabled(chainConf protocol.ChainConf) bool {
	enableGas := false
	enableOptimizeChargeGas := false
	if chainConf.ChainConfig() != nil && chainConf.ChainConfig().AccountConfig != nil {
		enableGas = chainConf.ChainConfig().AccountConfig.EnableGas
	}

	if chainConf.ChainConfig() != nil && chainConf.ChainConfig().Core != nil {
		enableOptimizeChargeGas = chainConf.ChainConfig().Core.EnableOptimizeChargeGas
	}
	return enableGas && enableOptimizeChargeGas
}

func VerifyOptimizeChargeGasTx(block *commonPb.Block, snapshot protocol.Snapshot) error {
	// gas to charge from validator
	gasCalc := make(map[string]uint64, 24)
	// gas to charge from proposer
	gasNeedToCharge := make(map[string]uint64, 24)
	chainCfg, err := snapshot.GetBlockchainStore().GetLastChainConfig()
	if err != nil {
		return fmt.Errorf("GetLastChainConfig error: %v", err)
	}

	contractName := syscontract.SystemContract_ACCOUNT_MANAGER.String()
	methodName := syscontract.GasAccountFunction_CHARGE_GAS_FOR_MULTI_ACCOUNT.String()
	found := false
	for _, tx := range block.Txs {
		if tx.Payload.ContractName == contractName && tx.Payload.Method == methodName {
			found = true
			for _, kv := range tx.Payload.Parameters {
				total, err2 := strconv.ParseUint(string(kv.Value), 10, 64)
				if err2 != nil {
					return fmt.Errorf("ParseUint error: %v", err2)
				}
				gasNeedToCharge[kv.Key] = total
			}
		} else {
			gasUsed := tx.Result.ContractResult.GasUsed
			pk, err2 := getPayerPkFromTx(tx, snapshot)
			if err2 != nil {
				return fmt.Errorf("getPayerPkFromTx error: %v", err2)
			}

			// convert the public key to `ZX` or `CM` or `EVM` address
			address, err2 := publicKeyToAddress(pk, chainCfg)
			if err2 != nil {
				return fmt.Errorf("publicKeyToAddress failed: err = %v", err)
			}
			if totalGas, exists := gasCalc[address]; exists {
				gasCalc[address] = totalGas + gasUsed
			} else {
				gasCalc[address] = gasUsed
			}
		}
	}

	if !found {
		return fmt.Errorf("charge gas tx is missing")
	}
	// compare gasCalc and gasNeedToCharge
	if len(gasCalc) != len(gasNeedToCharge) {
		return fmt.Errorf("gas need to charging is not correct, expect %v account, got %v account",
			len(gasCalc), len(gasNeedToCharge))
	}

	for addr, totalGasCalc := range gasCalc {
		if totalGasNeedToCharge, exists := gasNeedToCharge[addr]; !exists {
			return fmt.Errorf("missing some account to charge gas => `%v`", addr)
		} else if totalGasCalc != totalGasNeedToCharge {
			return fmt.Errorf("gas to charge error for address `%v`, expect %v, got %v",
				addr, totalGasCalc, totalGasNeedToCharge)
		}
	}

	return nil
}

func getMultiSignEnableManualRun(chainConfig *configPb.ChainConfig) bool {
	if chainConfig.Vm == nil {
		return false
	} else if chainConfig.Vm.Native == nil {
		return false
	} else if chainConfig.Vm.Native.Multisign == nil {
		return false
	}

	return chainConfig.Vm.Native.Multisign.EnableManualRun
}
