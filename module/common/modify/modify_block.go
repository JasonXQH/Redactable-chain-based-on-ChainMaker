package modify

import (
	"chainmaker.org/chainmaker-go/module/common/chameleon"
	"chainmaker.org/chainmaker-go/module/common/locate"
	"chainmaker.org/chainmaker-go/module/common/mysql"
	"chainmaker.org/chainmaker-go/module/core/common"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/utils/v2"
	"fmt"
	"sync"
)

func ModifyBlockByHeight(height uint64) (*commonPb.Block, []byte, map[string]*commonPb.TxRWSet) {
	oldBlock := locate.GetBlockByHeight(height).Block
	//oldBlockHash, _ := chameleon.ConvertToHashType(oldBlock.Header.BlockHash)

	txs, txRWSetMap := GetTxsAndRWSetMap()
	oldBlock.Txs = txs
	txHashes, _ := modifyBlockTransactions(oldBlock, txRWSetMap, "SHA256")
	//calc tx root
	oldMerkleTreeRoot, err := mysql.GetOldMerkleTreeRoot(height)
	oldMerkleTreeRootHash, err := chameleon.ConvertToHashType(oldMerkleTreeRoot)

	newMerkleTree, err := hash.BuildMerkleTree("SHA256", txHashes)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newMerkleTreeRoot := newMerkleTree[len(newMerkleTree)-1]
	newMerkleTreeRootHash, err := chameleon.ConvertToHashType(newMerkleTreeRoot)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newSalt, err := chameleon.ForgeMerkleRootSalt(oldMerkleTreeRootHash, "SHA256", newMerkleTreeRoot, oldBlock)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newHash := chameleon.Hash(newMerkleTreeRootHash, newSalt)
	newHashByte := chameleon.ConvertToBytesType(newHash)
	oldBlock.Header.BlockHash = newHashByte
	//fmt.Println("newHash: ", newHash.String(), "oldBlockHash: ", oldBlockHash.String())
	newBlock := common.CopyBlock(oldBlock)
	return newBlock, newSalt, txRWSetMap
}

func ModifyBlockByHash(blockhash string) (*commonPb.Block, []byte) {
	oldBlock := locate.GetBlockByHash(blockhash).Block
	oldBlockHash, _ := chameleon.ConvertToHashType(oldBlock.Header.BlockHash)

	txs, txRWSetMap := GetTxsAndRWSetMap()
	oldBlock.Txs = txs
	txHashes, _ := modifyBlockTransactions(oldBlock, txRWSetMap, "SHA256")
	//calc tx root
	oldMerkleTreeRoot, err := mysql.GetOldMerkleTreeRoot(oldBlock.Header.BlockHeight)
	oldMerkleTreeRootHash, err := chameleon.ConvertToHashType(oldMerkleTreeRoot)

	newMerkleTree, err := hash.BuildMerkleTree("SHA256", txHashes)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newMerkleTreeRoot := newMerkleTree[len(newMerkleTree)-1]
	newMerkleTreeRootHash, err := chameleon.ConvertToHashType(newMerkleTreeRoot)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newSalt, err := chameleon.ForgeMerkleRootSalt(oldMerkleTreeRootHash, "SHA256", newMerkleTreeRoot, oldBlock)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newHash := chameleon.Hash(newMerkleTreeRootHash, newSalt)
	newHashByte := chameleon.ConvertToBytesType(newHash)
	oldBlock.Header.BlockHash = newHashByte
	fmt.Println("newHash: ", newHash.String(), "oldBlockHash: ", oldBlockHash.String())
	newBlock := common.CopyBlock(oldBlock)
	return newBlock, newSalt
}

func ModifyBlockByTxID(txID string) (*commonPb.Block, []byte) {
	oldBlock := locate.GetBlockByTxID(txID).Block
	oldBlockHash, _ := chameleon.ConvertToHashType(oldBlock.Header.BlockHash)

	txs, txRWSetMap := GetTxsAndRWSetMap()
	oldBlock.Txs = txs
	txHashes, _ := modifyBlockTransactions(oldBlock, txRWSetMap, "SHA256")
	//calc tx root
	oldMerkleTreeRoot, err := mysql.GetOldMerkleTreeRoot(oldBlock.Header.BlockHeight)
	oldMerkleTreeRootHash, err := chameleon.ConvertToHashType(oldMerkleTreeRoot)

	newMerkleTree, err := hash.BuildMerkleTree("SHA256", txHashes)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newMerkleTreeRoot := newMerkleTree[len(newMerkleTree)-1]
	newMerkleTreeRootHash, err := chameleon.ConvertToHashType(newMerkleTreeRoot)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newSalt, err := chameleon.ForgeMerkleRootSalt(oldMerkleTreeRootHash, "SHA256", newMerkleTreeRoot, oldBlock)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newHash := chameleon.Hash(newMerkleTreeRootHash, newSalt)
	newHashByte := chameleon.ConvertToBytesType(newHash)
	oldBlock.Header.BlockHash = newHashByte
	fmt.Println("newHash: ", newHash.String(), "oldBlockHash: ", oldBlockHash.String())
	newBlock := common.CopyBlock(oldBlock)
	return newBlock, newSalt
}

func modifyBlockTransactions(oldBlock *commonPb.Block, txRWSetMap map[string]*commonPb.TxRWSet, hashType string) ([][]byte, error) {
	txCount := len(oldBlock.Txs)
	oldBlock.Header.TxCount = uint32(txCount)

	errsC := make(chan error, txCount+3) // txCount+3 possible errors
	txHashes := make([][]byte, txCount)
	wg := &sync.WaitGroup{}
	wg.Add(txCount)

	for i, tx := range oldBlock.Txs {
		// 获取或创建交易的读写集
		rwSet := txRWSetMap[tx.Payload.TxId]
		if rwSet == nil {
			rwSet = &commonPb.TxRWSet{
				TxId:     tx.Payload.TxId,
				TxReads:  nil,
				TxWrites: nil,
			}
		}

		// 并行计算交易哈希
		go func(tx *commonPb.Transaction, rwSet *commonPb.TxRWSet, x int) {
			defer wg.Done()
			var err error
			txHashes[x], err = getTxHash(tx, rwSet, hashType, oldBlock.Header)
			if err != nil {
				errsC <- err
			}
		}(tx, rwSet, i)
	}

	wg.Wait()

	// 检查是否有错误
	if len(errsC) > 0 {
		return nil, <-errsC
	}
	return txHashes, nil
}
func GetTxsAndRWSetMap() ([]*commonPb.Transaction, map[string]*commonPb.TxRWSet) {
	txs := make([]*commonPb.Transaction, 0)
	txRWSetMap := make(map[string]*commonPb.TxRWSet)
	for i := 0; i < 100; i++ {
		txId := "0x123456789" + fmt.Sprint(i)
		tx := createNewTestTx(txId)
		txs = append(txs, tx)
		txRWSetMap[txId] = &commonPb.TxRWSet{
			TxId:    txId,
			TxReads: nil,
			TxWrites: []*commonPb.TxWrite{{
				Key:          []byte(fmt.Sprintf("key%d", i)),
				Value:        []byte(fmt.Sprintf("value[%d]", i)),
				ContractName: "TestContract",
			}},
		}
	}
	return txs, txRWSetMap
}

func createNewTestTx(txID string) *commonPb.Transaction {
	//var hash = []byte("0123456789")
	return &commonPb.Transaction{
		Payload: &commonPb.Payload{
			ChainId:        "Chain1",
			TxType:         0,
			TxId:           txID,
			Timestamp:      utils.CurrentTimeMillisSeconds(),
			ExpirationTime: 0,
		},
		Result: &commonPb.Result{
			Code:           commonPb.TxStatusCode_SUCCESS,
			ContractResult: nil,
			RwSetHash:      nil,
		},
	}
}
func getTxHash(tx *commonPb.Transaction,
	rwSet *commonPb.TxRWSet,
	hashType string,
	blockHeader *commonPb.BlockHeader,
) (
	[]byte, error) {
	var rwSetHash []byte
	rwSetHash, err := utils.CalcRWSetHash(hashType, rwSet)
	if err != nil {
		return nil, err
	}
	if tx.Result == nil {
		// in case tx.Result is nil, avoid panic
		e := fmt.Errorf("tx(%s) result == nil", tx.Payload.TxId)
		return nil, e
	}
	tx.Result.RwSetHash = rwSetHash
	// calculate complete tx hash, include tx.Header, tx.Payload, tx.Result
	var txHash []byte
	txHash, err = utils.CalcTxHashWithVersion(
		hashType, tx, int(blockHeader.BlockVersion))
	if err != nil {
		return nil, err
	}
	return txHash, nil
}
