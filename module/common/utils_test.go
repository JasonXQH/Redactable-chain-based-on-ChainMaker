package common

import (
	"fmt"
	"testing"
)

func TestBase64ToHex(t *testing.T) {
	base64Str := "vUtLKmLUei9h7Absot0r93KFx5YgWLaKS9M8iYvoE2c="
	hexstr, _ := Base64ToHex(base64Str)
	fmt.Println(hexstr)
}
