package requestid

import (
	"crypto/rand"
	"encoding/hex"
)

func Next() string {
	value := make([]byte, 24)
	_, err := rand.Read(value)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
