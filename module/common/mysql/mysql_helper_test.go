package mysql

import (
	"fmt"
	"testing"
)

func TestPersistence(t *testing.T) {
	salt := persistence(3, []byte{186, 46, 13, 213, 161, 141, 229, 250, 87, 127, 53, 76, 195, 132, 126, 2, 19, 113, 26, 154, 195, 58, 12, 201, 232, 168, 191, 40, 122, 234, 217, 208}, true)
	fmt.Println(salt)
}
func TestGetBlockInfoFromMysql(t *testing.T) {

	_, err := getBlockInfoFromMysql(2)
	if err != nil {
		return
	}
}
