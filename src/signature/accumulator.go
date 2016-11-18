package signature

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
)

// Accumulator accumulates a signature for a corpus by
// repeatedly taking byte blobs of data and incorporating them into the
// signature.
type Accumulator interface {
	AddData(data []byte) (dataSignature Signature, err error)
	AccumulatedSignature() (accumulatedSignature Signature, err error)
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

func (ctx *accumulatorContext) AddData(data []byte) (Signature, error) {
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

	return dataSignature(data)
}

func (ctx *accumulatorContext) AccumulatedSignature() (Signature, error) {
	// TODO: Test this function when we have given it no data.
	if ctx.err != nil {
		return "", ctx.err
	} else if ctx.hasData == false {
		return "", errors.New("No data accumulated in signature accumulator")
	}

	return fromHash(ctx.signatureAccumulator), nil
}
