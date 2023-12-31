package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	"chainmaker.org/chainmaker-go/module/common/mysql"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/utils/v2"
	"crypto/rand"
	"fmt"
)

func generateRandomSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %v", err)
	}
	return salt, nil
}

// TODO 思考清楚，可以问问
func GetMerkleRoot(hashType string, txHashes [][]byte, block *commonPb.Block) ([]byte, error) {
	var saltIsNew = false
	if block.Header.TxCount == 0 {
		if utils.CanProposeEmptyBlock(3) {
			return nil, nil
		}
		return nil, fmt.Errorf("tx must not empty")
	}
	// 从数据库查询salt
	salt, err := mysql.GetSalt(block.Header.BlockHeight)
	if err != nil || salt == nil {
		// 如果没有找到，生成新的salt
		fmt.Println("如果没有找到，生成新的salt")
		salt, err = generateRandomSalt()
		saltIsNew = true
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("找到，旧的salt: ", salt)
	}
	//检查通过以后，创建merkleTree
	merkleTree, err := hash.BuildMerkleTree(hashType, txHashes)
	if err != nil {
		return nil, err
	}

	merkleTreeRoot := merkleTree[len(merkleTree)-1]
	merkleTreeRootHash, _ := ConvertToHashType(merkleTreeRoot)
	chameleonMerkleRoot := Hash(merkleTreeRootHash, salt)
	//fmt.Println("chameleonMerkleRoot: ", chameleonMerkleRoot.String())
	// 只有在生成了新的salt时才调用mysql.Persistence
	if err == nil && saltIsNew {
		//mysql.Persistence(block.Header.BlockHeight, merkleTreeRootHash, salt)
	}
	if err != nil {
		return nil, err
	}
	return ConvertToBytesType(chameleonMerkleRoot), nil
}
func GetBlockHash(block *commonPb.Block) ([]byte, error) {
	var saltFlag = false
	salt, err := mysql.GetSalt(block.Header.BlockHeight)
	merkleTreeRoot := block.Header.TxRoot
	merkleTreeRootHash, _ := ConvertToHashType(merkleTreeRoot)
	if salt == nil {
		// 如果没有找到，生成新的salt
		fmt.Println("如果没有找到，生成新的salt")
		salt, err = generateRandomSalt()
		saltFlag = true
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("找到，旧的salt: ", salt)
	}
	blockHeaderHash := Hash(merkleTreeRootHash, salt)
	if block.Header.TxCount == 0 {
		if utils.CanProposeEmptyBlock(3) {
			return nil, nil
		}
		return nil, fmt.Errorf("tx must not empty")
	}
	if err == nil && saltFlag {
		mysql.Persistence(block.Header.BlockHeight, merkleTreeRootHash, salt, blockHeaderHash)
	}
	return ConvertToBytesType(blockHeaderHash), nil
}
func ForgeMerkleRootSalt(oldTreeHash common.Hash, hashType string, new_merkleTreeRoot []byte, block *commonPb.Block) ([]byte, error) {
	blockHeight := block.Header.BlockHeight
	blockInfo, err2 := mysql.GetBlockInfoFromMysql(uint(blockHeight))
	if err2 != nil {
		return nil, err2
	}
	//fmt.Println("new_merkleTreeRoot: ", new_merkleTreeRoot)
	//oldSaltHash, _ := ConvertToHashType(blockInfo.RandomSalt)
	//fmt.Println("blockInfo.oldSalt: ", oldSaltHash.String())
	newTreeRootHash, _ := ConvertToHashType(new_merkleTreeRoot)
	new_salt := UForge(oldTreeHash, newTreeRootHash, blockInfo.RandomSalt).Bytes()
	//fmt.Println("new_salt: ", new_salt)
	//替换旧的salt
	//blockInfo.RandomSalt = new_salt
	//err3 := mysql.UpdateSalt(blockInfo)
	//if err3 != nil {
	//	return nil, err3
	//}
	//new_Hash := Hash(new_merkleTreeRoot, new_salt)
	//fmt.Println("newForgeHash: ", new_Hash.String())
	return new_salt, nil
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
