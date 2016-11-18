package signature

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"
)

// ValidateData will hash a stream of bytes using a canonical SHA512
// configuration, and validate that it matches a given SHA512 hash.
func ValidateData(stream []byte, targetSignature string) bool {
	// TODO: This actually returns an error. We should handle it correctly.
	actualSignature, _ := signature(stream)
	return actualSignature == Normalize(targetSignature)
}

// NormalizeSignature puts a signature string into canonical form.
func Normalize(signature string) string {
	return strings.ToLower(signature)
}

//
// PRIVATE FUNCTIONS.
//

func signature(data []byte) (string, error) {
	tarballSignature := sha512.New()

	_, writeErr := tarballSignature.Write(data)
	if writeErr != nil {
		return "", fmt.Errorf("Failed to generate signature for corpus "+
			"tarball:\n%v", writeErr)
	}

	return signatureString(tarballSignature), nil
}

func signatureString(signature hash.Hash) string {
	return Normalize(fmt.Sprintf("%x", signature.Sum(nil)))
}
