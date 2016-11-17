package signature

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"

	"github.com/BitFunnel/LabBook/src/util"
)

type accumulatorContext struct {
	signatureAccumulator hash.Hash
	hasData              bool
	err                  error
}

//
// CorpusSignatureAccumulator
//

// CorpusSignatureAccumulator accumulates a signature for a corpus by
// repeatedly taking tarballs from the corpus and incorporating them into the
// signature.
type CorpusSignatureAccumulator interface {
	AddCorpusTarball(tarballData []byte) (string, error)
	Signature() (string, error)
}

// NewCorpusSignatureAccumulator returns a corpus signature accumulator.
func NewCorpusSignatureAccumulator() CorpusSignatureAccumulator {
	return &accumulatorContext{
		signatureAccumulator: sha512.New(),
		hasData:              false,
		err:                  nil,
	}
}

func (ctx *accumulatorContext) Signature() (string, error) {
	// TODO: Test this function when we have given it no data.
	if ctx.err != nil {
		return "", ctx.err
	} else if ctx.hasData == false {
		return "", errors.New("No data accumulated in signature accumulator")
	}

	return signatureString(ctx.signatureAccumulator), nil
}

func (ctx *accumulatorContext) AddCorpusTarball(tarballData []byte) (string, error) {
	return ctx.addData(tarballData)
}

//
// SampleSignatureAccumulator
//

// SampleSignatureAccumulator accumulates a signature for a corpus sample by
// repeatedly taking data from the corpus and incorporating them into the
// signature.
type SampleSignatureAccumulator interface {
	AddSampleData(sampleData []byte) (string, error)
	Signature() (string, error)
}

// NewSampleSignatureAccumulator returns a sample signature accumulator.
func NewSampleSignatureAccumulator() SampleSignatureAccumulator {
	return &accumulatorContext{
		signatureAccumulator: sha512.New(),
		hasData:              false,
		err:                  nil,
	}
}

func (ctx *accumulatorContext) AddSampleData(sampleData []byte) (string, error) {
	return ctx.addData(sampleData)
}

//
// PRIVATE METHODS.
//

func (ctx *accumulatorContext) addData(data []byte) (string, error) {
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
