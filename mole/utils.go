package mole

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
)

func randNumber(length int) string {
	key := make([]byte, length)
	rand.Read(key)
	for i := range key {
		key[i] = key[i]%10 + '0'
	}
	return string(key)
}

// utils to convert int between bytes
func int2bytes(n int) []byte { // 4 bytes
	x := int32(n)
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, x)
	return buf.Bytes()
}

func bytes2int(bs []byte) int {
	var x int32
	buf := bytes.NewBuffer(bs)
	binary.Read(buf, binary.BigEndian, &x)
	return int(x)
}
