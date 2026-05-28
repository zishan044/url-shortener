package url

import (
	"crypto/rand"
	"math/big"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateShortCode(minLen, maxLen int) (string, error) {
	if minLen == 0 {
		minLen = 6
	}
	if maxLen == 0 {
		maxLen = 8
	}

	lengthBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxLen-minLen+1)))
	if err != nil {
		return "", err
	}
	length := int(lengthBig.Int64()) + minLen

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIdx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[randomIdx.Int64()]
	}

	return string(result), nil
}
