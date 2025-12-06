package conf

import (
	"encoding/hex"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	ciphertext, err := EncryptAESGCM([]byte("你好啊"))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(hex.EncodeToString(ciphertext))

	cipher, _ := hex.DecodeString(hex.EncodeToString(ciphertext))

	paintext, err := DecryptAESGCM(cipher)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(paintext))
}
