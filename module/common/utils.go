package common

import (
	"encoding/base64"
	"encoding/hex"
)

// Base64ToHex 将Base64编码的字符串转换为16进制哈希值
func Base64ToHex(base64Str string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HexToBase64 将16进制哈希值转换为Base64编码的字符串
func HexToBase64(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
