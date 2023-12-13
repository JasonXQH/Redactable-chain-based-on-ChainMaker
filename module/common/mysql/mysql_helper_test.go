package mysql

import (
	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"testing"
)

func createBlock(height uint64) *commonpb.Block {
	var hash = []byte("0123456789")
	var version = uint32(1)
	var block = &commonpb.Block{
		Header: &commonpb.BlockHeader{
			ChainId:        "Chain1",
			BlockHeight:    height,
			PreBlockHash:   hash,
			BlockHash:      hash,
			PreConfHeight:  0,
			BlockVersion:   version,
			DagHash:        hash,
			RwSetRoot:      hash,
			TxRoot:         hash,
			BlockTimestamp: 0,
			Proposer:       &accesscontrol.Member{MemberInfo: hash},
			ConsensusArgs:  nil,
			TxCount:        1,
			Signature:      []byte(""),
		},
		Dag: &commonpb.DAG{
			Vertexes: nil,
		},
		Txs: nil,
	}

	return block
}

var info = &BlockInfo{
	RandomSalt:  []byte{186, 46, 13, 213},
	BlockHeight: 3,
	IsModified:  true,
}

func TestPersistence(t *testing.T) {
	var testBlock = createBlock(3)
	salt := Persistence(testBlock)
	fmt.Println(salt)
}

func TestGetBlockInfoFromMysql(t *testing.T) {
	_, err := GetBlockInfoFromMysql(3)
	if err != nil {
		return
	}
}

func TestUpdateSalt(t *testing.T) {
	UpdateSalt(info)
}
