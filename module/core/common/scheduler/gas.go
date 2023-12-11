package scheduler

import (
	"strings"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/utils/v2"
	gasutils "chainmaker.org/chainmaker/utils/v2/gas"
)

// calcTxGasUsed for block version greater than 2030102
func calcTxGasUsed(txSimContext protocol.TxSimContext, log protocol.Logger) (uint64, error) {

	gasUsed := uint64(0)
	blockVersion := txSimContext.GetBlockVersion()
	if blockVersion < blockVersion2312 {
		return gasUsed, nil
	} // for block version < 2030102
	gasConfig := gasutils.NewGasConfig(txSimContext.GetLastChainConfig().AccountConfig)
	if gasConfig == nil {
		return gasUsed, nil
	} // for enable_gas == false

	tx := txSimContext.GetTx()
	contractName := tx.Payload.ContractName
	method := tx.Payload.Method
	parameters := tx.Payload.Parameters

	// 用户合约，按 Invoke 扣费
	if !utils.IsNativeContract(contractName) {
		return calcInvokeTxGasUsed(tx.Payload, gasConfig, log)
	}

	if contractName == syscontract.SystemContract_CONTRACT_MANAGE.String() {
		if method == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
			method == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
			// Native 合约中的 install & upgrade, 按 Install 扣费
			return calcInstallTxGasUsed(tx.Payload, gasConfig, log)
		}

	} else if contractName == syscontract.SystemContract_MULTI_SIGN.String() {
		if method == syscontract.MultiSignFunction_REQ.String() {
			var sysContractName string
			var sysMethod string
			for _, kvPair := range parameters {
				if kvPair.Key == syscontract.MultiReq_SYS_CONTRACT_NAME.String() {
					sysContractName = string(kvPair.Value)
				} else if kvPair.Key == syscontract.MultiReq_SYS_METHOD.String() {
					sysMethod = string(kvPair.Value)
				}
			}

			if sysContractName == syscontract.SystemContract_CONTRACT_MANAGE.String() {
				if sysMethod == syscontract.ContractManageFunction_INIT_CONTRACT.String() ||
					sysMethod == syscontract.ContractManageFunction_UPGRADE_CONTRACT.String() {
					// 针对 install & upgrade 的 multi-sign req, 按 Install 扣费
					return calcInstallTxGasUsed(tx.Payload, gasConfig, log)
				}
			}
		}
		// 除了 install & upgrade 的 multi-sign, 按 Invoke 扣费
		return calcInvokeTxGasUsed(tx.Payload, gasConfig, log)
	}

	// 普通 Native 合约，除了 install & upgrade，都不扣费
	return uint64(0), nil
}

func calcInstallTxGasUsed(payload *commonPb.Payload,
	gasConfig *gasutils.GasConfig, log protocol.Logger) (uint64, error) {
	if gasConfig == nil {
		return 0, nil
	}

	parameters := payload.Parameters
	installBaseGas := gasConfig.GetBaseGasForInstall()
	dataSize := len(payload.ContractName) + len(payload.Method) + len(payload.TxId)

	for _, kvPair := range parameters {
		log.Debugf("【gas calc】%v, key = %v, value size = %v", payload.TxId, kvPair.Key, len(kvPair.Value))
		dataSize += len(kvPair.Key) + len(kvPair.Value)
	}

	dataGas, err := gasutils.MultiplyGasPrice(dataSize, gasConfig.GetGasPriceForInstall())
	if err != nil {
		return 0, err
	}

	return installBaseGas + dataGas, nil
}

func calcInvokeTxGasUsed(payload *commonPb.Payload,
	gasConfig *gasutils.GasConfig, log protocol.Logger) (uint64, error) {
	if gasConfig == nil {
		return 0, nil
	}

	parameters := payload.Parameters
	invokeBaseGas := gasConfig.GetBaseGasForInvoke()
	dataSize := len(payload.ContractName) + len(payload.Method) + len(payload.TxId)

	for _, kvPair := range parameters {
		log.Debugf("【gas calc】%v, key = %v, value size = %v", payload.TxId, kvPair.Key, len(kvPair.Value))
		dataSize += len(kvPair.Key) + len(kvPair.Value)
	}

	dataGas, err := gasutils.MultiplyGasPrice(dataSize, gasConfig.GetGasPriceForInvoke())
	if err != nil {
		return 0, err
	}

	return invokeBaseGas + dataGas, nil
}

func calcTxRWSetGasUsed(txSimContext protocol.TxSimContext,
	isTxSuccess bool,
	log protocol.Logger) (uint64, error) {

	gasRWSet := uint64(0)
	blockVersion := txSimContext.GetBlockVersion()
	if blockVersion < blockVersion2312 {
		return gasRWSet, nil
	} // for block version < 2030102

	gasConfig := gasutils.NewGasConfig(txSimContext.GetLastChainConfig().AccountConfig)
	if gasConfig == nil {
		return gasRWSet, nil
	}

	dataSize := 0
	rwSet := txSimContext.GetTxRWSet(isTxSuccess)
	for _, txRead := range rwSet.TxReads {
		if !utils.IsNativeContract(txRead.ContractName) {
			dataSize += calcReadSetItemSize(txRead)
		}
		log.Debugf("【gas calc】%v, read key = %v # %v, value size = %v, dataSize = %v",
			txSimContext.GetTx().Payload.TxId, txRead.ContractName, string(txRead.Key), len(txRead.Value), dataSize)
	}
	for _, txWrite := range rwSet.TxWrites {
		if !utils.IsNativeContract(txWrite.ContractName) {
			dataSize += calcWriteSetItemSize(txWrite)
		}
		log.Debugf("【gas calc】%v, write key = %v # %v, value size = %v, dataSize = %v",
			txSimContext.GetTx().Payload.TxId, txWrite.ContractName, string(txWrite.Key), len(txWrite.Value), dataSize)
	}

	log.Debugf("【gas calc】%v, calcTxRWSetGasUsed, dataSize = %v, gas_price = %v, read_set(%v), write_set(%v)",
		txSimContext.GetTx().Payload.TxId,
		dataSize, gasConfig.GetGasPriceForInvoke(),
		len(rwSet.TxReads), len(rwSet.TxWrites))
	gasRWSet, err := gasutils.MultiplyGasPrice(dataSize, gasConfig.GetGasPriceForInvoke())
	if err != nil {
		return 0, err
	}

	return gasRWSet, nil
}

func calcReadSetItemSize(txRead *commonPb.TxRead) int {
	dataSize := len(txRead.ContractName)

	keyLabels := strings.Split(string(txRead.Key), protocol.ContractStoreSeparator)
	for _, keyLabel := range keyLabels {
		dataSize += len(keyLabel)
	}

	dataSize += len(txRead.Value)

	return dataSize
}

func calcWriteSetItemSize(txWrite *commonPb.TxWrite) int {
	dataSize := len(txWrite.ContractName)

	keyLabels := strings.Split(string(txWrite.Key), protocol.ContractStoreSeparator)
	for _, keyLabel := range keyLabels {
		dataSize += len(keyLabel)
	}

	dataSize += len(txWrite.Value)

	return dataSize
}

func calcTxEventGasUsed(
	txSimContext protocol.TxSimContext,
	events []*commonPb.ContractEvent, log protocol.Logger) (uint64, error) {

	gasEvents := uint64(0)
	blockVersion := txSimContext.GetBlockVersion()
	if blockVersion < blockVersion2312 {
		return gasEvents, nil
	} // for block version < 2030102

	gasConfig := gasutils.NewGasConfig(txSimContext.GetLastChainConfig().AccountConfig)
	if gasConfig == nil {
		return gasEvents, nil
	}

	dataSize := 0
	for _, event := range events {
		if !utils.IsNativeContract(event.ContractName) {
			dataSize += len(event.ContractName) + len(event.ContractVersion) + len(event.Topic)
			for _, dataItem := range event.EventData {
				dataSize += len(dataItem)
			}
		}
		log.Debugf("【gas calc】%v, event contractName = %v, topic = %v, dataSize = %v",
			txSimContext.GetTx().Payload.TxId, event.ContractName, string(event.Topic), dataSize)
	}

	log.Debugf("【gas calc】%v, calcTxEventGasUsed, dataSize = %v, gas_price = %v",
		txSimContext.GetTx().Payload.TxId,
		dataSize, gasConfig.GetGasPriceForInvoke())
	gasEvents, err := gasutils.MultiplyGasPrice(dataSize, gasConfig.GetGasPriceForInvoke())
	if err != nil {
		return 0, err
	}

	return gasEvents, nil
}
