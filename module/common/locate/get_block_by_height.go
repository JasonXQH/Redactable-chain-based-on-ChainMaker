package locate

import (
	"bytes"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

func GetBlockByHeight(blockHeight uint64) *commonpb.BlockInfo {
	// 创建一个命令对象
	cmd := exec.Command("/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/tools/cmc/cmc", "query", "block-by-height", strconv.FormatUint(blockHeight, 10),
		"--chain-id=chain1",
		"--sdk-conf-path=/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/module/common/locate/sdk_config.yml")

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
	var blockInfo *commonpb.BlockInfo
	err = json.Unmarshal([]byte(stdoutResult), &blockInfo)
	//fmt.Println(stdoutResult)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}
	return blockInfo
}
