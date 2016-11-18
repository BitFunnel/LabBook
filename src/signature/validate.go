package signature

import (
	"github.com/BitFunnel/LabBook/src/systems"
)

// ValidateData will hash a stream of bytes using a canonical SHA512
// configuration, and validate that it matches a given SHA512 hash.
func ValidateData(stream []byte, targetSignature Signature) bool {
	// NOTE: A dry run never needs to validate anything.
	if systems.IsDryRun() {
		return true
	}

	// TODO: This actually returns an error. We should handle it correctly.
	actualSignature, _ := dataSignature(stream)
	return actualSignature == targetSignature
}

// NormalizeAndValidate will normalize and ensure two signatures are equal.
func NormalizeAndValidate(signature1 Signature, signature2 Signature) bool {
	// NOTE: A dry run never needs to validate anything.
	if systems.IsDryRun() {
		return true
	}

	signature1.Normalize()
	signature2.Normalize()
	return signature1 == signature2
}
