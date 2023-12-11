package scheduler

import (
	"fmt"
	"strconv"

	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/vm-native/v2/accountmgr"

	"chainmaker.org/chainmaker/common/v2/crypto"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// SenderCollection contains:
// key: address
// value: tx collection will address's other data
type SenderCollection struct {
	txsMap map[string]*TxCollection
}

type TxCollection struct {
	// public key to generate address
	publicKey crypto.PublicKey
	// balance of the address saved at SenderCollection
	accountBalance int64
	// total gas added each tx
	totalGasUsed int64
	txs          []*commonPb.Transaction
}

func (g *TxCollection) String() string {
	pubKeyStr, _ := g.publicKey.String()
	return fmt.Sprintf(
		"\nTxsGroup{ \n\tpublicKey: %s, \n\taccountBalance: %v, \n\ttotalGasUsed: %v, \n\ttxs: [%d items] }",
		pubKeyStr, g.accountBalance, g.totalGasUsed, len(g.txs))
}

func NewSenderCollection(
	txBatch []*commonPb.Transaction,
	snapshot protocol.Snapshot,
	log protocol.Logger) *SenderCollection {
	return &SenderCollection{
		txsMap: getSenderTxCollection(txBatch, snapshot, log),
	}
}

// getSenderTxCollection split txs in txBatch by sender account
func getSenderTxCollection(
	txBatch []*commonPb.Transaction,
	snapshot protocol.Snapshot,
	log protocol.Logger) map[string]*TxCollection {
	txCollectionMap := make(map[string]*TxCollection, len(txBatch))

	var err error
	chainCfg := snapshot.GetLastChainConfig()

	for _, tx := range txBatch {
		// get the public key from tx
		pk, err2 := getPayerPkFromTx(tx, snapshot)
		if err2 != nil {
			log.Errorf("getPayerPkFromTx failed: err = %v", err)
			continue
		}

		// convert the public key to `ZX` or `CM` or `EVM` address
		address, err2 := publicKeyToAddress(pk, chainCfg)
		if err2 != nil {
			log.Error("publicKeyToAddress failed: err = %v", err)
			continue
		}

		txCollection, exists := txCollectionMap[address]
		if !exists {
			txCollection = &TxCollection{
				publicKey:      pk,
				accountBalance: int64(0),
				totalGasUsed:   int64(0),
				txs:            make([]*commonPb.Transaction, 0),
			}
			txCollectionMap[address] = txCollection
		}
		txCollection.txs = append(txCollection.txs, tx)
	}

	if chainCfg.GetBlockVersion() < blockVersion2312 {
		preHandleTxCollectionMap2310(txCollectionMap, snapshot, log)
	} else {
		preHandleTxCollectionMap2312(txCollectionMap, snapshot, log)
	}

	return txCollectionMap
}

func preHandleTxCollectionMap2310(
	txCollectionMap map[string]*TxCollection,
	snapshot protocol.Snapshot,
	log protocol.Logger) {

	var err error
	for senderAddress, txCollection := range txCollectionMap {
		// get the account balance from snapshot
		txCollection.accountBalance, err = getAccountBalanceFromSnapshotLt2312(senderAddress, snapshot, log)
		if err != nil {
			errMsg := fmt.Sprintf("get account balance failed: err = %v", err)
			log.Error(errMsg)
			for _, tx := range txCollection.txs {
				tx.Result = &commonPb.Result{
					Code: commonPb.TxStatusCode_CONTRACT_FAIL,
					ContractResult: &commonPb.ContractResult{
						Code:    uint32(1),
						Result:  nil,
						Message: "",
						GasUsed: uint64(0),
					},
					RwSetHash: nil,
					Message:   errMsg,
				}

			}
		}
	}
}

func preHandleTxCollectionMap2312(
	txCollectionMap map[string]*TxCollection,
	snapshot protocol.Snapshot,
	log protocol.Logger) {

	for senderAddress, txCollection := range txCollectionMap {
		// get the account balance from snapshot
		var errCode commonPb.TxStatusCode
		txCollection.accountBalance, errCode = getAccountBalanceFromSnapshot2312(senderAddress, snapshot, log)
		if errCode != commonPb.TxStatusCode_SUCCESS {
			errMsg := fmt.Sprintf("get account balance failed: errCode = %v", errCode)
			log.Error(errMsg)
			for _, tx := range txCollection.txs {
				tx.Result = &commonPb.Result{
					Code: errCode,
					ContractResult: &commonPb.ContractResult{
						Code:    uint32(1),
						Result:  nil,
						Message: errMsg,
						GasUsed: uint64(0),
					},
					RwSetHash: nil,
					Message:   errMsg,
				}

			}
		}
	}
}

func (s SenderCollection) Clear() {
	for addr := range s.txsMap {
		delete(s.txsMap, addr)
	}
}

func getAccountBalanceFromSnapshotLt2312(
	address string, snapshot protocol.Snapshot, log protocol.Logger) (int64, error) {
	chainConfig := snapshot.GetLastChainConfig()
	blockVersion := chainConfig.GetBlockVersion()
	log.Debugf("address = %v, blockVersion = %v", address, blockVersion)

	if blockVersion < blockVersion2310 {
		return getAccountBalanceFromSnapshot2300(address, snapshot, log)
	}

	return getAccountBalanceFromSnapshot2310(address, snapshot, log)
}

func getAccountBalanceFromSnapshot2300(
	address string, snapshot protocol.Snapshot, log protocol.Logger) (int64, error) {

	var err error
	var balance int64
	balanceData, err := snapshot.GetKey(-1,
		syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		[]byte(accountmgr.AccountPrefix+address))
	if err != nil {
		return -1, err
	}

	if len(balanceData) == 0 {
		balance = int64(0)
	} else {
		balance, err = strconv.ParseInt(string(balanceData), 10, 64)
		if err != nil {
			return 0, err
		}
	}

	return balance, nil
}

func getAccountBalanceFromSnapshot2310(
	address string, snapshot protocol.Snapshot, log protocol.Logger) (int64, error) {
	var err error
	var balance int64
	var frozen bool

	// 查询账户的余额
	balanceData, err := snapshot.GetKey(-1,
		syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		[]byte(accountmgr.AccountPrefix+address))
	if err != nil {
		return -1, err
	}

	if len(balanceData) == 0 {
		balance = int64(0)
	} else {
		balance, err = strconv.ParseInt(string(balanceData), 10, 64)
		if err != nil {
			return 0, err
		}
	}

	// 查询账户的状态
	frozenData, err := snapshot.GetKey(-1,
		syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		[]byte(accountmgr.FrozenPrefix+address))
	if err != nil {
		return -1, err
	}

	if len(frozenData) == 0 {
		frozen = false
	} else {
		if string(frozenData) == "0" {
			frozen = false
		} else if string(frozenData) == "1" {
			frozen = true
		}
	}
	log.Debugf("balance = %v, freeze = %v", balance, frozen)

	if frozen {
		return 0, fmt.Errorf("account `%s` has been locked", address)
	}

	return balance, nil
}

func getAccountBalanceFromSnapshot2312(
	address string, snapshot protocol.Snapshot, log protocol.Logger) (int64, commonPb.TxStatusCode) {

	var err error
	var balance int64
	var frozen bool

	// 查询账户的余额
	balanceData, err := snapshot.GetKey(-1,
		syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		[]byte(accountmgr.AccountPrefix+address))
	if err != nil {
		return -1, commonPb.TxStatusCode_GET_ACCOUNT_BALANCE_FAILED
	}

	if len(balanceData) == 0 {
		balance = int64(0)
	} else {
		balance, err = strconv.ParseInt(string(balanceData), 10, 64)
		if err != nil {
			return 0, commonPb.TxStatusCode_PARSE_ACCOUNT_BALANCE_FAILED
		}
	}

	// 查询账户的状态
	frozenData, err := snapshot.GetKey(-1,
		syscontract.SystemContract_ACCOUNT_MANAGER.String(),
		[]byte(accountmgr.FrozenPrefix+address))
	if err != nil {
		return -1, commonPb.TxStatusCode_GET_ACCOUNT_STATUS_FAILED
	}

	if len(frozenData) == 0 {
		frozen = false
	} else {
		if string(frozenData) == "0" {
			frozen = false
		} else if string(frozenData) == "1" {
			frozen = true
		}
	}
	log.Debugf("balance = %v, freeze = %v", balance, frozen)

	if frozen {
		return 0, commonPb.TxStatusCode_ACCOUNT_STATUS_FROZEN
	}

	return balance, commonPb.TxStatusCode_SUCCESS
}
