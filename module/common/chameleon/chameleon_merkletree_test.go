package chameleon

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"testing"
)

// 测试GetMerkleRoot函数
func TestGetMerkleRoot(t *testing.T) {
	// 初始化测试数据
	hashType := "SHA256" // 或者是其他哈希类型
	txHashes := [][]byte{
		// 这里添加一些示例交易哈希
		{0xe2, 0xa8, 0x33, 0xb2, 0x85, 0x29, 0xdf, 0xbc, 0x1a, 0xc6, 0x0e, 0xa8, 0x0e, 0x9f, 0x29, 0x9f, 0x97, 0xb4, 0xa5, 0x06, 0x6e, 0x57, 0x9f, 0x1b, 0x0b, 0xa0, 0x97, 0xeb, 0x64, 0xc2, 0x60, 0x92},
		{0xa5, 0xc1, 0xee, 0x8c, 0xbc, 0xb4, 0xb2, 0x85, 0x29, 0xdd, 0xe1, 0xa9, 0x83, 0x2e, 0x49, 0x2a, 0xd0, 0xb3, 0x74, 0xbe, 0xd1, 0xb4, 0xfc, 0xa3, 0x4f, 0x07, 0x3a, 0x1a, 0x14, 0xb9, 0x00, 0x74},
		{0x1f, 0x29, 0x69, 0xce, 0xbb, 0x9c, 0x64, 0x79, 0x2e, 0x07, 0x80, 0xc1, 0x8a, 0x4d, 0x91, 0x22, 0x96, 0x3f, 0x4c, 0xee, 0xca, 0x15, 0x7f, 0x29, 0xa8, 0xb3, 0xe4, 0x3a, 0xad, 0x2e, 0x7f, 0x1f},
		{0x61, 0xab, 0x28, 0x3d, 0x4b, 0x82, 0x19, 0x84, 0x22, 0x5b, 0x6c, 0x4e, 0x04, 0xeb, 0x08, 0x8f, 0x44, 0x79, 0xfd, 0x35, 0xe8, 0x61, 0x8b, 0x2b, 0x7d, 0x51, 0xbd, 0x50, 0xec, 0xf7, 0x26, 0x5b},
		{0xe6, 0x94, 0x28, 0x10, 0x92, 0x92, 0x3a, 0x1f, 0xae, 0x12, 0xe7, 0xe5, 0x5c, 0x19, 0x6b, 0x39, 0x7e, 0xa3, 0x40, 0xb1, 0x78, 0xff, 0x27, 0x1a, 0x78, 0x04, 0xdb, 0xcc, 0xf2, 0x26, 0x63, 0xf4},
	}
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			BlockHeight: 1, // 示例区块高度
		},
	}
	//var blockHeightToSaltMap map[int][]byte
	// 假设的salt值，正常情况下应从blockHeightToSaltMap获取
	//blockHeightToSaltMap[int(block.Header.BlockHeight)] = []byte{1, 2, 3}
	// 调用函数
	merkleroot, err := GetMerkleRoot(hashType, txHashes, block)
	if err != nil {
		t.Errorf("GetMerkleRoot returned an error: %v", err)
	}
	fmt.Print(merkleroot)
	// 添加其他断言，检查返回的Merkle根是否符合预期
}