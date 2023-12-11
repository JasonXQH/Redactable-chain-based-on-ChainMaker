package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	cmCrypto "chainmaker.org/chainmaker/common/v2/crypto"

	//"golang.org/x/crypto/scrypt"
	//"golang.org/x/crypto/sha3"
	chainmakerHash "chainmaker.org/chainmaker/common/v2/crypto/hash"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	test := "Hello world!"
	//h.Write([]byte(test))
	h, _ := chainmakerHash.GetHashAlgorithm(cmCrypto.HASH_TYPE_SHA256)
	h.Write([]byte(test))
	testHash := common.BytesToHash(h.Sum(nil))
	//fmt.Println(testHash)
	testSalt := []byte{1, 2, 3}
	fmt.Println(Hash(testHash, testSalt).String())
	h.Reset()
	test1 := "It's a beautiful day!"
	h.Write([]byte(test1))
	test1Hash := common.BytesToHash(h.Sum(nil))
	test1Salt := UForge(testHash, test1Hash, testSalt).Bytes()
	//fmt.Println("新的test1Salt", test1Salt)
	//fmt.Println("旧的testSalt", testSalt)
	fmt.Println(Hash(test1Hash, test1Salt).String())
}
