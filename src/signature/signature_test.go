package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SignatureEmpty(t *testing.T) {
	signatureAccumulator := NewAccumulator()
	signature, sigErr := signatureAccumulator.AccumulatedSignature()
	assert.Error(t, sigErr)
	assert.EqualValues(t, "", signature)
}

func Test_SignatureSimple(t *testing.T) {
	signatureAccumulator := NewAccumulator()
	tarballSignature1, sigErr :=
		signatureAccumulator.AddData([]byte{1, 2, 3, 4})
	assert.NoError(t, sigErr)
	assert.NotEqual(t, "", tarballSignature1)

	tarballSignature2, sigErr :=
		signatureAccumulator.AddData([]byte{5, 6, 7, 8})
	assert.NoError(t, sigErr)
	assert.NotEqual(t, "", tarballSignature2)

	assert.NotEqual(t, tarballSignature1, tarballSignature2)

	signature, sigErr := signatureAccumulator.AccumulatedSignature()
	assert.NoError(t, sigErr)
	assert.NotEqual(t, "", signature)
	assert.NotEqual(t, tarballSignature1, signature)
	assert.NotEqual(t, tarballSignature2, signature)
}
