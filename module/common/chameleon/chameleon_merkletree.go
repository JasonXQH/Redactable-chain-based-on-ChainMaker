package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	"chainmaker.org/chainmaker-go/module/common/mysql"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/utils/v2"
	"fmt"
)

//TODO 思考清楚，可以问问
func GetMerkleRoot(hashType string, txHases [][]byte, block *commonPb.Block) ([]byte, error) {
	salt := mysql.Persistence(block)
	if block.Header.TxCount == 0 {
		if utils.CanProposeEmptyBlock(1) {
			// for consensus that allows empty block, skip txs verify
			return nil, nil
		}
		//TODO 思考如果是空区块该如何？如果是创世区块？那么需要创建一个随机salt，并更具随机salt创建一个区块哈希
		return salt, fmt.Errorf("tx must not empty")
	}
	//检查通过以后，创建merkleTree
	merkleTree, err := hash.BuildMerkleTree(hashType, txHases)
	if err != nil {
		return nil, err
	}

	merkleTreeRoot, _ := ConvertToHashType(merkleTree[len(merkleTree)-1])
	chameleonMerkleRoot := Hash(merkleTreeRoot, salt)
	fmt.Println("chameleonMerkleRoot: ", chameleonMerkleRoot.String())
	if err != nil {
		return nil, err
	}
	return ConvertToBytesType(chameleonMerkleRoot), nil
}

func ForgeMerkleRootSalt(oldTreeHash common.Hash, hashType string, txHases [][]byte, block *commonPb.Block) ([]byte, error) {
	blockHeight := block.Header.BlockHeight
	blockInfo, err2 := mysql.GetBlockInfoFromMysql(uint(blockHeight))

	if err2 != nil {
		return nil, err2
	}
	//TODO 需要做是否没有交易的判断，如果没有，该...
	var merkleTree, err = hash.BuildMerkleTree(hashType, txHases)

	if err != nil {
		return nil, err
	}
	new_merkleTreeRoot, _ := ConvertToHashType(merkleTree[len(merkleTree)-1])
	new_salt := UForge(oldTreeHash, new_merkleTreeRoot, blockInfo.RandomSalt).Bytes()

	//替换旧的salt
	blockInfo.RandomSalt = new_salt
	err3 := mysql.UpdateSalt(blockInfo)
	if err3 != nil {
		return nil, err3
	}
	new_Hash := Hash(new_merkleTreeRoot, new_salt)
	fmt.Println("newForgeHash: ", new_Hash.String())
	return ConvertToBytesType(new_Hash), nil
}
func ConvertToHashType(merkleRoot []byte) (common.Hash, error) {
	var hash common.Hash
	if len(merkleRoot) == 32 {
		copy(hash[:], merkleRoot)
	} else {
		// 处理长度不匹配的情况
		return hash, fmt.Errorf("merkle root length mismatch")
	}
	return hash, nil
}
func ConvertToBytesType(hash common.Hash) []byte {
	return hash[:]
}
