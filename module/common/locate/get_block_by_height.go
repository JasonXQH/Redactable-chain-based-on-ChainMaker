package locate

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
)

func GetBlockByHeight(blockHeight uint64) {
	// 创建一个命令对象
	cmd := exec.Command("/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/tools/cmc/cmc", "query", "block-by-height", strconv.FormatUint(blockHeight, 10),
		"--chain-id=chain1",
		"--sdk-conf-path=./sdk_config.yml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		fmt.Println("执行命令时发生错误:", err)
		fmt.Println("标准错误输出:", stderrBuf.String())
		return
	}

	stdoutResult := stdoutBuf.String()
	stderrResult := stderrBuf.String()

	fmt.Println("标准输出:", stdoutResult)
	if stderrResult != "" {
		fmt.Println("标准错误输出:", stderrResult)
	}
}
