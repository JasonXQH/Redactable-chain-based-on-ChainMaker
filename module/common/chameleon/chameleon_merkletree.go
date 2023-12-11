package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	"chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
)

var blockHeightToSaltMap map[int][]byte

func GetMerkleRoot(hashType string, txHases [][]byte, block *commonPb.Block) ([]byte, error) {
	var blockHeight = block.Header.BlockHeight
	fmt.Println(blockHeight)
	merkleTree, err := hash.BuildMerkleTree(hashType, txHases)
	if err != nil {
		return nil, err
	}
	salt := blockHeightToSaltMap[int(blockHeight)]
	merkleTreeRoot, _ := ConvertToHashType(merkleTree[len(merkleTree)-1])
	chameleonMerkleRoot := Hash(merkleTreeRoot, salt)
	if err != nil {
		return nil, err
	}

	return ConvertToBytesType(chameleonMerkleRoot), nil

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
