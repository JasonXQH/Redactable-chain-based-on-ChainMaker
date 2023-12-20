package common

import (
	"fmt"
	"testing"
)

func TestBase64ToHex(t *testing.T) {
	base64Str := "GMYjZPgGdhhiV3yaMbGHNFA2oafb18t3aQ0kB596EX4="
	hexstr, _ := Base64ToHex(base64Str)
	fmt.Println(hexstr)
}
