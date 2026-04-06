package protocol

import (
	"crypto/rand"
	"encoding/hex"
)

// generateMsgID 生成消息 ID
func generateMsgID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}