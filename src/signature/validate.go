package signature

// ValidateData will hash a stream of bytes using a canonical SHA512
// configuration, and validate that it matches a given SHA512 hash.
func ValidateData(stream []byte, targetSignature Signature) bool {
	// TODO: This actually returns an error. We should handle it correctly.
	actualSignature, _ := dataSignature(stream)
	return actualSignature == targetSignature
}
