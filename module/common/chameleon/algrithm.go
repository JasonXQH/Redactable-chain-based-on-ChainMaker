package chameleon

import (
	"chainmaker.org/chainmaker-go/module/common"
	"chainmaker.org/chainmaker/common/v2/crypto"
	chainmakerHash "chainmaker.org/chainmaker/common/v2/crypto/hash"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	goHash "hash"
	"math/big"
)

var (
	Tao                = big.NewInt(256) // Parameter of chameleon hash.
	Kappa              = big.NewInt(256) // Parameter of chameleon hash.
	logger             protocol.Logger
	customizedIdentity = []byte("Hello world!")
	j                  = new(big.Int)
	p                  = new(big.Int)
	q                  = new(big.Int)
	n                  = new(big.Int)
	e                  = new(big.Int)
	d                  = new(big.Int)
)

func init() {
	//hashType := crypto.HASH_TYPE_SHA256
	h, _ := chainmakerHash.GetHashAlgorithm(crypto.HASH_TYPE_SHA256)
	h.Write(customizedIdentity)
	jHash := h.Sum(nil)
	h.Reset()
	jBytes, err := emsaPSSEncode(jHash, 360, []byte{1}, h)

	if err != nil {
		logger.Warn("Emsa pass encode err %s", err)
		return
	}
	j.SetBytes(jBytes)
	p.SetString("d0c1baff1b227fb6dc35150c217467aeede5e30babbdff7407bba941b64a4669", 16)
	q.SetString("ce66f004358b85619abae98c5ad95bf21e0b0a5aa5f0c37f65aca1e58314fc67", 16)
	n.SetString("a84fd562d77a899c63311b03dd83ec4636096bfbf5edc97b29b1c67d03575e1aa28bebeb0e4f66fd99809b83bcf3bafe61e7d837bd3b0dc432cd0da3b065b03f", 16)
	e.SetString("01e34970639f9a14dbe7386be418345c7743ab116200e78b11502f84a8d83e6a2f", 16)
	d.SetString("9dd1141a3f48a5f2d210af79427229a5fb1a59b6b2fb2b5a7ef91da9682af0e2b6fb2beebce4b3f2625a098703793828c99bc0bb0f8a50ff77ae96c242ec4b8f", 16)
	//keys := generateKey()
	//// 如果确保返回的切片长度是 5，可以直接按索引取值并分配给对应的变量
	//if len(keys) == 5 {
	//	p, q, n, e, d := keys[0], keys[1], keys[2], keys[3], keys[4]
	//	// 在这里使用得到的 p、q、n、e、d 值
	//	// 例如：
	//	fmt.Println("p:", hex.EncodeToString(p.Bytes()))
	//	fmt.Println("q:", hex.EncodeToString(q.Bytes()))
	//	fmt.Println("n:", hex.EncodeToString(n.Bytes()))
	//	fmt.Println("e:", hex.EncodeToString(e.Bytes()))
	//	fmt.Println("d:", hex.EncodeToString(d.Bytes()))
	//}

}

// From Go/src/crypto/rsa/pss.go.
func emsaPSSEncode(mHash []byte, emBits int, salt []byte, hash goHash.Hash) ([]byte, error) {
	// See [1], section 9.1.1
	hLen := hash.Size()
	sLen := len(salt)
	emLen := (emBits + 7) / 8

	// 1.  If the length of M is greater than the input limitation for the
	//     hash function (2^61 - 1 octets for SHA-1), output "message too
	//     long" and stop.
	//
	// 2.  Let mHash = Hash(M), an octet string of length hLen.

	if len(mHash) != hLen {
		return nil, errors.New("crypto/rsa: input must be hashed message")
	}

	// 3.  If emLen < hLen + sLen + 2, output "encoding error" and stop.

	if emLen < hLen+sLen+2 {
		return nil, errors.New("crypto/rsa: key size too small for PSS signature")
	}

	em := make([]byte, emLen)
	db := em[:emLen-sLen-hLen-2+1+sLen]
	h := em[emLen-sLen-hLen-2+1+sLen : emLen-1]

	// 4.  Generate a random octet string salt of length sLen; if sLen = 0,
	//     then salt is the empty string.
	//
	// 5.  Let
	//       M' = (0x)00 00 00 00 00 00 00 00 || mHash || salt;
	//
	//     M' is an octet string of length 8 + hLen + sLen with eight
	//     initial zero octets.
	//
	// 6.  Let H = Hash(M'), an octet string of length hLen.

	var prefix [8]byte

	hash.Write(prefix[:])
	hash.Write(mHash)
	hash.Write(salt)

	h = hash.Sum(h[:0])
	hash.Reset()

	// 7.  Generate an octet string PS consisting of emLen - sLen - hLen - 2
	//     zero octets. The length of PS may be 0.
	//
	// 8.  Let DB = PS || 0x01 || salt; DB is an octet string of length
	//     emLen - hLen - 1.

	db[emLen-sLen-hLen-2] = 0x01
	copy(db[emLen-sLen-hLen-1:], salt)

	// 9.  Let dbMask = MGF(H, emLen - hLen - 1).
	//
	// 10. Let maskedDB = DB \xor dbMask.

	mgf1XOR(db, hash, h)

	// 11. Set the leftmost 8 * emLen - emBits bits of the leftmost octet in
	//     maskedDB to zero.

	db[0] &= 0xFF >> uint(8*emLen-emBits)

	// 12. Let EM = maskedDB || H || 0xbc.
	em[emLen-1] = 0xBC

	// 13. Output EM.
	return em, nil
}

// From Go/src/crypto/rsa/pss.go.
func mgf1XOR(out []byte, hash goHash.Hash, seed []byte) {
	var counter [4]byte
	var digest []byte

	done := 0
	for done < len(out) {
		hash.Write(seed)
		hash.Write(counter[0:4])
		digest = hash.Sum(digest[:0])
		hash.Reset()

		for i := 0; i < len(digest) && done < len(out); i++ {
			out[done] ^= digest[i]
			done++
		}
		incCounter(&counter)
	}
}

// From Go/src/crypto/rsa/pss.go.
func incCounter(c *[4]byte) {
	if c[3]++; c[3] != 0 {
		return
	}
	if c[2]++; c[2] != 0 {
		return
	}
	if c[1]++; c[1] != 0 {
		return
	}
	c[0]++
}

func generateKey() []*big.Int {
	//初始化两个大整数one和two
	one := big.NewInt(1)
	two := big.NewInt(2)
	//生成素数 p
	var p *big.Int
	pLimit := new(big.Int).Set(two)
	pLimit.Exp(pLimit, Kappa.Sub(Kappa, one), nil)
	for {
		p, _ = rand.Prime(rand.Reader, int(Kappa.Int64())+1)
		if p.Cmp(pLimit) == 1 {
			break
		}
	}
	//生成素数 q
	var q *big.Int
	qLimit := new(big.Int).Set(two)
	qLimit.Exp(qLimit, Kappa.Add(Kappa, one), nil)
	qLimit.Sub(qLimit, one)
	for {
		q, _ = rand.Prime(rand.Reader, int(Kappa.Int64()))
		if q.Cmp(qLimit) == -1 {
			break
		}
	}
	//计算n，它是p和q的乘积，这将成为RSA公钥的一部分
	n := new(big.Int)
	n.Mul(p, q)
	/*
		欧拉函数 ϕ(n):
		计算pSub为p-1，qSub为q-1。
		ϕ(n) = (p-1)*(q-1)
	*/
	euler := new(big.Int)
	qSub := new(big.Int)
	qSub.SetBytes(q.Bytes())
	qSub.Sub(qSub, one)
	pSub := new(big.Int)
	pSub.SetBytes(p.Bytes())
	pSub.Sub(pSub, one)
	euler.Mul(pSub, qSub)
	/*
		生成素数 e:
		设置一个限制eLimit为 2^τ
		在循环中随机生成一个大素数e，其位大小为τ+1。
		计算e和euler的最大公约数（gcd）。
		检查e是否大于eLimit并且gcd等于1。如果是，则跳出循环。
	*/
	var e *big.Int
	eLimit := new(big.Int).Set(two)
	eLimit.Exp(eLimit, Tao, nil)
	gcd := new(big.Int)
	for {
		e, _ = rand.Prime(rand.Reader, int(Tao.Int64())+1)
		gcd = gcd.GCD(nil, nil, e, euler)
		if e.Cmp(eLimit) == 1 && gcd.Cmp(one) == 0 {
			break
		}
	}
	//计算私钥d，它是e模ϕ(n)的逆元
	d := new(big.Int)
	d.ModInverse(e, euler)
	fmt.Println(hex.EncodeToString(p.Bytes()), "p")
	fmt.Println(hex.EncodeToString(q.Bytes()), "q")
	fmt.Println(hex.EncodeToString(n.Bytes()), "n")
	fmt.Println(hex.EncodeToString(e.Bytes()), "e")
	fmt.Println(hex.EncodeToString(d.Bytes()), "d")
	return []*big.Int{p, q, n, e, d}
}

func Hash(msg common.Hash, salt []byte) common.Hash {
	//Hash(L,m,r)=J^(H(m)) r^e mod n
	reverseMsg := new(big.Int).SetBytes(msg.Bytes())
	// 这里的salt就代表随机数r
	reverseSalt := new(big.Int).SetBytes(salt)
	//计算jHm：计算J^(H(m))，这是通过对j（应该是预先计算好的C(L)）进行模n的指数运算来完成的
	jHm := new(big.Int)
	jHm.Exp(j, reverseMsg, n)
	//计算r^e，即对salt（即r）进行模n的指数运算
	rE := new(big.Int)
	rE.Exp(reverseSalt, e, n)
	//将jHm和rE相乘，然后对结果取模n，得到最终的哈希值。
	re := new(big.Int)
	re.Mul(jHm, rE)
	re.Mod(re, n)
	return common.BytesToHash(re.Bytes())
}

func CreateChameleonHash(tx *commonPb.Transaction, rwSet *commonPb.TxRWSet) ([]byte, error) {

	return nil, nil
}

func UForge(oldMsg, newMsg common.Hash, oldSalt []byte) *big.Int {
	//消息和盐的转换：将oldMsg、newMsg和oldSalt转换成大整数。
	reverseOldMsg := new(big.Int).SetBytes(oldMsg.Bytes())
	reverseNewMsg := new(big.Int).SetBytes(newMsg.Bytes())
	reverseOldSalt := new(big.Int).SetBytes(oldSalt)
	//计算b：计算B=J^d mod n，即用私钥d对j进行模n指数运算。
	b := new(big.Int)
	b.Exp(j, d, n)
	//计算哈希差值：计算H(m)-H(m')。
	hM := new(big.Int).Sub(reverseOldMsg, reverseNewMsg)
	//根据论文中的方程， B^{H(m)-H(m')}
	bHmMod := new(big.Int).Exp(b, hM, n)
	//计算原式的 r mod n
	rMod := new(big.Int).Mod(reverseOldSalt, n)
	//计算新的r' = r * B^{H(m)-H(m')}
	re := new(big.Int)
	re.Mul(bHmMod, rMod)
	re.Mod(re, n)
	//返回新的r'，再次调用 Hash(m',r') 即可得到和原式一致的哈希函数
	return re
}
