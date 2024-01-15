package locate

import (
	"bytes"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"encoding/json"
	"fmt"
	"os/exec"
)

func GetBlockByHash(blockHash string) *commonpb.BlockInfo {
	cmd := exec.Command("/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/tools/cmc/cmc", "query", "block-by-hash", blockHash,
		"--chain-id=chain1",
		"--sdk-conf-path=./sdk_config.yml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		fmt.Println("执行命令时发生错误:", err)
		fmt.Println("标准错误输出:", stderrBuf.String())
		return nil
	}

	stdoutResult := stdoutBuf.String()
	//stderrResult := stderrBuf.String()

	var blockInfo *commonpb.BlockInfo
	err = json.Unmarshal([]byte(stdoutResult), &blockInfo)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}
	//fmt.Println(blockInfo.Block.Txs)
	//fmt.Println("标准输出:", stdoutResult)
	//if stderrResult != "" {
	//	fmt.Println("标准错误输出:", stderrResult)
	//}
	return blockInfo

}
