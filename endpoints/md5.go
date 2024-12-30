package endpoints

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
)

func GetMD5Hash(text ...string) string {
	buh := bytes.NewBuffer(nil)
	for i := range text {
		buh.WriteString(text[i])
	}
	hash := md5.Sum(buh.Bytes())
	return hex.EncodeToString(hash[:])
}
