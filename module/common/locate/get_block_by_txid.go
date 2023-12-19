package locate

import (
	"bytes"
	"fmt"
	"os/exec"
)

func GetBlockByTxID(txID string) {
	cmd := exec.Command("/Users/jasonxu/projects/shproj/chainMaker/chainmaker-go/tools/cmc/cmc", "query", "block-by-txid", txID,
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
