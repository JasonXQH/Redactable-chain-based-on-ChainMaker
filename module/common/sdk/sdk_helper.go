package sdk

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/sdk-go/v2/examples"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hokaccha/go-prettyjson"
	"log"
)

const (
	sdkConfigOrg1Admin1Path  = "../sdk_configs/sdk_config_org1_admin1.yml"
	sdkConfigOrg1Client1Path = "../sdk_configs/sdk_config_org1_client1.yml"
)

func main() {
	testArchive()

	// 确保链外存储中已经有已归档的区块数据
	testRestore()
	testGetFromArchiveStore()
}

func testArchive() {
	admin1, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Admin1Path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 数据归档 ======================")
	var targetBlockHeight uint64 = 20
	testArchiveBlock(admin1, targetBlockHeight)
}

func testRestore() {
	admin1, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Admin1Path)
	if err != nil {
		log.Fatalln(err)
	}

	client1, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 归档恢复 ======================")
	var blockHeight uint64 = 20

	fullBlock, err := client1.GetArchivedBlockByHeight(blockHeight, true)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedFullBlockByHeight fullBlock", fullBlock)

	fullBlockBytes, err := proto.Marshal(fullBlock)
	if err != nil {
		log.Fatalln(err)
	}

	testRestoreBlock(admin1, fullBlockBytes)
}

func testGetFromArchiveStore() {
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 归档查询 ======================")
	var blockHeight uint64 = 8
	fullBlockInfo, err := client.GetArchivedBlockByHeight(blockHeight, false)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetFromArchiveStore fullBlockInfo", fullBlockInfo)

	fullBlockInfo, err = client.GetArchivedBlockByHeight(blockHeight, true)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedFullBlockByHeight fullBlockInfo", fullBlockInfo)

	blockInfo, err := client.GetArchivedBlockByHeight(blockHeight, true)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedBlockByHeight with rwset", blockInfo)

	blockInfo, err = client.GetArchivedBlockByHeight(blockHeight, false)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedBlockByHeight without rwset", blockInfo)

	blockInfo, err = client.GetArchivedBlockByHash(hex.EncodeToString(blockInfo.Block.Header.BlockHash), true)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedBlockByHash with rwset", blockInfo)

	blockInfo, err = client.GetArchivedBlockByHash(hex.EncodeToString(blockInfo.Block.Header.BlockHash), false)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedBlockByHash without rwset", blockInfo)

	txId := blockInfo.Block.Txs[0].Payload.TxId
	txInfo, err := client.GetArchivedTxByTxId(txId)
	if err != nil {
		log.Fatalln(err)
	}
	prettyJsonShow("GetArchivedTxByTxId", txInfo)
}

func prettyJsonShow(name string, v interface{}) {
	marshal, err := prettyjson.Marshal(v)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\n\n\n====== %s ======\n%s\n==========================\n", name, marshal)
}

func testArchiveBlock(admin1 *sdk.ChainClient, targetBlockHeight uint64) {
	var (
		err     error
		payload *common.Payload
		resp    *common.TxResponse
	)

	payload, err = admin1.CreateArchiveBlockPayload(targetBlockHeight)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err = admin1.SendArchiveBlockRequest(payload, -1)
	if err != nil {
		log.Fatalln(err)
	}

	err = examples.CheckProposalRequestResp(resp, false)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("resp: %+v\n", resp)
}

func testRestoreBlock(admin1 *sdk.ChainClient, fullBlock []byte) {
	var (
		err     error
		payload *common.Payload
		resp    *common.TxResponse
	)

	payload, err = admin1.CreateRestoreBlockPayload(fullBlock)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err = admin1.SendRestoreBlockRequest(payload, -1)
	if err != nil {
		log.Fatalln(err)
	}

	err = examples.CheckProposalRequestResp(resp, false)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("resp: %+v\n", resp)
}
