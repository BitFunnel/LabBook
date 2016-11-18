package lock

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/BitFunnel/LabBook/src/signature"
	"github.com/stretchr/testify/assert"
)

func Test_SimpleLockRoundTrip(t *testing.T) {
	lockFileData := bytes.NewBufferString(simpleCorpusLockFileData)
	lockFile, deserializeErr := DeserializeLockFile(
		lockFileData,
		"simpleCorpusLockFileData")
	assert.NoError(t, deserializeErr)

	assert.Equal(t, corpusLockFile.Signature(), lockFile.Signature())
	assert.Equal(t, 1, len(lockFile.DependencySignatures()))
	assert.Equal(
		t,
		corpusLockFile.DependencySignatures()[simpleCorpusTarball],
		lockFile.DependencySignatures()[simpleCorpusTarball])

	var serializedBuffer bytes.Buffer
	serializeErr := SerializeLockFile(lockFile, &serializedBuffer)
	assert.NoError(t, serializeErr)
	assert.EqualValues(t, serializedBuffer.String(), simpleCorpusLockFileData)
}

func Test_SimpleValidate(t *testing.T) {
	validationErr :=
		ValidateSampleLockFile(&corpusLockFile, &sampleLockFile)
	assert.NoError(t, validationErr)

	validationErr =
		ValidateConfigLockFile(&sampleLockFile, &configLockFile)
	assert.NoError(t, validationErr)

	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&configLockFile,
		&experimentLockFile)
	assert.NoError(t, validationErr)
}

func Test_DirectValidation(t *testing.T) {
	validationErr := validateDependency(
		&sampleLockFileEmptyDep,
		&corpusLockFileEmptySig,
		corpusKey)
	assert.Error(t, validationErr)
}

func Test_SimpleValidateFail(t *testing.T) {
	validationErr :=
		ValidateSampleLockFile(&File{}, &File{})
	assert.Error(t, validationErr)
	validationErr =
		ValidateSampleLockFile(&corpusLockFile, &File{})
	assert.Error(t, validationErr)
	validationErr =
		ValidateSampleLockFile(&File{}, &sampleLockFile)
	assert.Error(t, validationErr)
	validationErr =
		ValidateSampleLockFile(&sampleLockFile, &corpusLockFile)
	assert.Error(t, validationErr)

	validationErr =
		ValidateConfigLockFile(&File{}, &File{})
	assert.Error(t, validationErr)
	validationErr =
		ValidateConfigLockFile(&sampleLockFile, &File{})
	assert.Error(t, validationErr)
	validationErr =
		ValidateConfigLockFile(&File{}, &configLockFile)
	assert.Error(t, validationErr)
	validationErr =
		ValidateConfigLockFile(&configLockFile, &sampleLockFile)
	assert.Error(t, validationErr)

	validationErr = ValidateExperimentLockFile(&File{}, &File{}, &File{})
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&File{},
		&configLockFile,
		&experimentLockFile)
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&File{},
		&experimentLockFile)
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&configLockFile,
		&File{})
	assert.Error(t, validationErr)
}

func Test_ComplexValidateFail(t *testing.T) {
	validationErr :=
		ValidateSampleLockFile(&corpusLockFile, &sampleLockFileBrokenDep)
	assert.Error(t, validationErr)

	validationErr =
		ValidateConfigLockFile(&sampleLockFile, &configLockFile)
	assert.NoError(t, validationErr)
	validationErr =
		ValidateConfigLockFile(&sampleLockFileBrokenSig, &configLockFile)
	assert.Error(t, validationErr)
	validationErr =
		ValidateConfigLockFile(&sampleLockFile, &configLockFileBrokenDep)
	assert.Error(t, validationErr)

	validationErr = ValidateExperimentLockFile(
		&sampleLockFileBrokenDep,
		&configLockFileBrokenDep,
		&experimentLockFile)
	assert.NoError(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFileBrokenSig,
		&configLockFile,
		&experimentLockFile)
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&configLockFileBrokenSig,
		&experimentLockFile)
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&configLockFile,
		&experimentLockFileBrokenDep1)
	assert.Error(t, validationErr)
	validationErr = ValidateExperimentLockFile(
		&sampleLockFile,
		&configLockFile,
		&experimentLockFileBrokenDep2)
	assert.Error(t, validationErr)
}

//
// Test data.
//
// A series of `lock.File` objects for testing. Some of them are purposefully
// broken (specifically the ones with names following the pattern `*BrokenX`),
// and some are correct.
//
const simpleCorpusTarball = "enwiki-20161020-chunked1.tar.gz"
const simpleCorpusTarballSignature = "1a3be37650cbb6708c2c4385f6ebcf944d1cda933f2dd327de20acb6c72cf687737540f0108bcdcd4b6fc1e5014824bf1cdcb3304e87bfe6a82e0c7642b28e3f"
const simpleCorpusSignature = "7377a37246eb5472ce6103c2a292a1383e58b1e33d98692dc9bbe05e4a580bf52274aa041334cd1c12ebd12febca0c3c5370ab2ba12cbc2f3568d5b1d12f1201"
const simpleConfigSignature = "3a5df945fb20b1675a5bcf0e7b882c6e22e0262f67a96e684bf7329b8f912eaa2b20f6e6132fe7bc4d55d2bca135946dec04ae7cba128cd47cba4f9868a4fc9f"
const simpleSampleSignature = "b8244d028981d693af7b456af8efa4cad63d282e19ff14942c246e50d9351d22704a802a71c3580b6370de4ceb293c324a8423342557d4e5c38438f0e36910ee"

var corpusLockFile = File{
	Signature_: simpleCorpusSignature,
	DependencySignatures_: map[string]signature.Signature{
		simpleCorpusTarball: simpleCorpusTarballSignature,
	},
}
var corpusLockFileEmptySig = File{
	Signature_: "",
	DependencySignatures_: map[string]signature.Signature{
		simpleCorpusTarball: simpleCorpusTarballSignature,
	},
}

var sampleLockFile = File{
	Signature_: simpleSampleSignature,
	DependencySignatures_: map[string]signature.Signature{
		corpusKey: simpleCorpusSignature,
	},
}
var sampleLockFileBrokenDep = File{
	Signature_: simpleSampleSignature,
	DependencySignatures_: map[string]signature.Signature{
		corpusKey: signature.New(simpleCorpusSignature[5:]),
	},
}
var sampleLockFileBrokenSig = File{
	Signature_: signature.New(simpleSampleSignature[5:]),
	DependencySignatures_: map[string]signature.Signature{
		corpusKey: simpleCorpusSignature,
	},
}
var sampleLockFileEmptyDep = File{
	Signature_: simpleSampleSignature,
	DependencySignatures_: map[string]signature.Signature{
		corpusKey: "",
	},
}

var configLockFile = File{
	Signature_: simpleConfigSignature,
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: simpleSampleSignature,
	},
}
var configLockFileBrokenDep = File{
	Signature_: simpleConfigSignature,
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: signature.New(simpleSampleSignature[5:]),
	},
}
var configLockFileBrokenSig = File{
	Signature_: signature.New(simpleConfigSignature[5:]),
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: simpleSampleSignature,
	},
}

var experimentLockFile = File{
	Signature_: "",
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: simpleSampleSignature,
		configKey: simpleConfigSignature,
	},
}
var experimentLockFileBrokenDep1 = File{
	Signature_: "",
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: signature.New(simpleSampleSignature[5:]),
		configKey: simpleConfigSignature,
	},
}
var experimentLockFileBrokenDep2 = File{
	Signature_: "",
	DependencySignatures_: map[string]signature.Signature{
		sampleKey: simpleSampleSignature,
		configKey: signature.New(simpleConfigSignature[5:]),
	},
}

var simpleCorpusLockFileData = fmt.Sprintf(`dependency-signatures:
  %s: %s
signature: %s
`,
	simpleCorpusTarball,
	corpusLockFile.DependencySignatures()[simpleCorpusTarball],
	corpusLockFile.Signature())
