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

func ModifyBlockByHeight(height uint64) *commonPb.Block {
	blockInfo := locate.GetBlockByHeight(height)
	oldBlock := blockInfo.Block
	oldBlockHash, _ := chameleon.ConvertToHashType(oldBlock.Header.BlockHash)
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
	oldBlock.Txs = txs
	// TxRoot/RwSetRoot
	txCount := len(oldBlock.Txs)
	oldBlock.Header.TxCount = uint32(txCount)
	errsC := make(chan error, txCount+3) // txCount+3 possible errors
	txHashes := make([][]byte, txCount)
	wg := &sync.WaitGroup{}
	wg.Add(txCount)
	for i, tx := range oldBlock.Txs {
		// finalize tx, put rwsethash into tx.Result
		rwSet := txRWSetMap[tx.Payload.TxId]
		if rwSet == nil {
			rwSet = &commonPb.TxRWSet{
				TxId:     tx.Payload.TxId,
				TxReads:  nil,
				TxWrites: nil,
			}
		}
		go func(tx *commonPb.Transaction, rwSet *commonPb.TxRWSet, x int) {
			defer wg.Done()
			var err error
			txHashes[x], err = getTxHash(tx, rwSet, "SHA256", oldBlock.Header)
			if err != nil {
				errsC <- err
			}

		}(tx, rwSet, i)
	}
	wg.Wait()
	if len(errsC) > 0 {
		//err := <-errsC
		return nil
	}
	wg.Add(3)
	//calc tx root
	oldMerkleTreeRoot, err := mysql.GetOldMerkleTreeRoot(height)
	oldSalt, _ := mysql.GetSalt(height)

	//fmt.Println("oldMerkleTreeRoot: ", oldMerkleTreeRoot, " oldSalt: ", oldSalt)
	if err != nil {
		fmt.Println("err: ", err)
	}
	oldMerkleTreeRootHash, err := chameleon.ConvertToHashType(oldMerkleTreeRoot)
	if err != nil {
		fmt.Println("err: ", err)
	}
	oldHash := chameleon.Hash(oldMerkleTreeRootHash, oldSalt)
	fmt.Println("oldHash: ", oldHash.String())

	newMerkleTree, err := hash.BuildMerkleTree("SHA256", txHashes)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newMerkleTreeRoot := newMerkleTree[len(newMerkleTree)-1]
	newMerkleTreeRootHash, err := chameleon.ConvertToHashType(newMerkleTreeRoot)
	if err != nil {
		fmt.Println("err: ", err)
	}
	//fmt.Println("newMerkleTreeRoot: ", newMerkleTreeRootHash.String())
	newSalt, err := chameleon.ForgeMerkleRootSalt(oldMerkleTreeRootHash, "SHA256", newMerkleTreeRoot, oldBlock)
	if err != nil {
		fmt.Println("err: ", err)
	}
	newHash := chameleon.Hash(newMerkleTreeRootHash, newSalt)
	newHashByte := chameleon.ConvertToBytesType(newHash)
	oldBlock.Header.BlockHash = newHashByte
	//fmt.Println("oldHash: ", newSalt)
	fmt.Println("newHash: ", newHash.String(), "oldBlockHash: ", oldBlockHash.String())
	newBlock := common.CopyBlock(oldBlock)
	oldTxRootHash, _ := chameleon.ConvertToHashType(oldBlock.Header.TxRoot)
	fmt.Println("oldTxRoot: ", oldTxRootHash.String())
	return newBlock
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
