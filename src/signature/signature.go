package signature

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"
)

// Signature is a generic representation of the fingerprint of a blob of data.
type Signature string

// New creates a new signature from a string.
func New(signatureData string) Signature {
	return Signature(strings.ToLower(signatureData))
}

// Normalize will normalize a `Signature`. For most use cases, `Signature`s are
// normalized by calling `New`; the reason this method exists is to normalize
// signatures that are deserialized from strings.
func (s *Signature) Normalize() {
	*s = New(fmt.Sprintf("%s", *s))
}

func fromHash(signatureHash hash.Hash) Signature {
	return New(fmt.Sprintf("%x", signatureHash.Sum(nil)))
}

func dataSignature(data []byte) (Signature, error) {
	tarballSignature := sha512.New()

	_, writeErr := tarballSignature.Write(data)
	if writeErr != nil {
		return "", fmt.Errorf("Failed to generate signature for corpus "+
			"tarball:\n%v", writeErr)
	}

	return fromHash(tarballSignature), nil
}
