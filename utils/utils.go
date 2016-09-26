package utils

import (
	"crypto/rand"
	"fmt"
)

const (
	UUID_SIZE = 16
)

func Uuid4() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", buf[0:4], buf[4:6], buf[6:8],
		buf[8:10], buf[10:16])
}
