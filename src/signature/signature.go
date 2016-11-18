package signature

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"

	"github.com/BitFunnel/LabBook/src/util"
)

// Accumulator accumulates a signature for a corpus by
// repeatedly taking byte blobs of data and incorporating them into the
// signature.
type Accumulator interface {
	AddData(data []byte) (dataSignature string, err error)
	AccumulatedSignature() (accumulatedSignature string, err error)
}

type accumulatorContext struct {
	signatureAccumulator hash.Hash
	hasData              bool
	err                  error
}

// NewAccumulator returns a corpus signature accumulator.
func NewAccumulator() Accumulator {
	return &accumulatorContext{
		signatureAccumulator: sha512.New(),
		hasData:              false,
		err:                  nil,
	}
}

func (ctx *accumulatorContext) AddData(data []byte) (string, error) {
	// TODO: What if `data` is empty? Do we still set this to true? This will
	// break tests because when we call `createSignature`, if we haven't put
	// data into the sig, it will fail.
	ctx.hasData = true

	_, writeErr := ctx.signatureAccumulator.Write(data)
	if writeErr != nil {
		ctx.err = fmt.Errorf("Failed to generate signature for corpus "+
			"tarball:\n%v", writeErr)
		return "", ctx.err
	}

	return signature(data)
}

func (ctx *accumulatorContext) AccumulatedSignature() (string, error) {
	// TODO: Test this function when we have given it no data.
	if ctx.err != nil {
		return "", ctx.err
	} else if ctx.hasData == false {
		return "", errors.New("No data accumulated in signature accumulator")
	}

	return signatureString(ctx.signatureAccumulator), nil
}

//
// PRIVATE FUNCTIONS.
//

func signature(tarballData []byte) (string, error) {
	tarballSignature := sha512.New()

	_, writeErr := tarballSignature.Write(tarballData)
	if writeErr != nil {
		return "", fmt.Errorf("Failed to generate signature for corpus "+
			"tarball:\n%v", writeErr)
	}

	return signatureString(tarballSignature), nil
}

func signatureString(signature hash.Hash) string {
	return util.NormalizeSignature(fmt.Sprintf("%x", signature.Sum(nil)))
}
