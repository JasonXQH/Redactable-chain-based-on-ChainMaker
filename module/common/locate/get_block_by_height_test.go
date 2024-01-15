package locate

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"testing"
)

func TestGetBlockByHeight(t *testing.T) {
	blockInfo := GetBlockByHeight(1)
	block := blockInfo.Block
	printTxIds(block)
}

// printTxIds 遍历block中的所有交易并打印tx_id
func printTxIds(block *commonPb.Block) {
	if block == nil {
		fmt.Println("Block is nil")
		return
	}

	for _, tx := range block.Txs {
		fmt.Println("Transaction ID:", tx.Payload.TxId)
	}
}
