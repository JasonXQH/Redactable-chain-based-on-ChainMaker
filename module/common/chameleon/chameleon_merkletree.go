package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"sync"
)

// 声明全局map变量
var blockHeightToSalt map[int][]byte

// 使用sync.Once确保线程安全的初始化
var initMapOnce sync.Once

// var blockHeightToSaltMap map[int][]byte
func initMap() {
	blockHeightToSalt = make(map[int][]byte)
}

func getBlockHeightToHashMap() map[int][]byte {
	initMapOnce.Do(initMap)
	return blockHeightToSalt
}

func GetMerkleRoot(hashType string, txHases [][]byte, block *commonPb.Block) ([]byte, error) {
	blockHeight := block.Header.BlockHeight
	blockMap := getBlockHeightToHashMap()
	blockMap[int(blockHeight)] = []byte{1, 2, 3}

	merkleTree, err := hash.BuildMerkleTree(hashType, txHases)
	if err != nil {
		return nil, err
	}
	merkleTreeRoot, _ := ConvertToHashType(merkleTree[len(merkleTree)-1])
	chameleonMerkleRoot := Hash(merkleTreeRoot, blockMap[int(blockHeight)])
	fmt.Println("chameleonMerkleRoot: ", chameleonMerkleRoot.String())
	if err != nil {
		return nil, err
	}
	return ConvertToBytesType(chameleonMerkleRoot), nil
}

func ForgeMerkleRootSalt(oldTreeHash common.Hash, hashType string, txHases [][]byte, block *commonPb.Block) ([]byte, error) {
	blockHeight := block.Header.BlockHeight
	blockMap := getBlockHeightToHashMap()
	var merkleTree, err = hash.BuildMerkleTree(hashType, txHases)

	if err != nil {
		return nil, err
	}
	new_merkleTreeRoot, _ := ConvertToHashType(merkleTree[len(merkleTree)-1])
	new_salt := UForge(oldTreeHash, new_merkleTreeRoot, blockMap[int(blockHeight)]).Bytes()

	//TODO 替换旧的salt
	blockMap[int(blockHeight)] = new_salt

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
