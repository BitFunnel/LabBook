package lock

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/go-yaml/yaml"
)

const corpusKey = "corpus-signature"
const sampleKey = "sample-signature"
const configKey = "config-signature"

// TODO: Replace uses of `File` with `Manager` instead.

// Manager is a general interface for managing all the information we need to
// verify that a cached dataset is the dataset we ran an experiment on. In
// particular, it contains a signature for any dependency steps, as well as a
// signature for the current step. For example, the BitFunnel runtime
// configuration step depends on a specific version of a corpus sample; that
// sample will have a signature, and the signature in that lock.File must fit
// the signature listed here.
type Manager interface {
	DependencySignatures() map[string]string
	Signature() string
	UpdateSignature(signature string)
	Name() string
	IsLocked() bool
}

// LOCKING PROTOCOL:
// * Check to see if LOCKFILE is present for both the dependencies and the
//   current step before performing any operation on cached data. If it doesn't,
//   report that the data is not cached, and error out.
// * If they all exist, verify that the current step's dependency signatures
//   match the dependency signatures:
//     * Deserialize both the LOCKFILE for the current step, and the LOCKFILE
//       for all dependency steps. For example, if we're running an experiment,
//       we deserialize the LOCKFILE for the corpus and the configuration, as
//       well as the experiment.
//     * Check that the dependency signatures of the current step are the same
//       as the signatures contained in the dependency LOCKFILES. If they are
//       the same, proceed; if not, report they don't match, suggest how to
//       mitigate, and error out.
// * Delete current LOCKFILE. This is a safety measure.
// * Run the step; if the current run fails and we haven't overwritten any
//   files, re-write the old LOCKFILE; if we succeed, re-generate LOCKFILE and
//   write to disk.

// File (i.e., lock.File) is an specific implementation of the `lock.Manager`
// interface, that centers around a YAML specification that can be serialized
// deserialized to and from disk.
type File struct {
	DependencySignatures_ map[string]string `yaml:"dependency-signatures"`
	Signature_            string            `yaml:"signature"`
	name                  string
	isLocked              bool
}

// DependencySignatures returns a map containing all dependencies of the
// resource being locked, and their signatures. What the keys are depends on
// the context, and should be largly opaque, as it's not intended to be
// manipulated. You can see the schema by looking at the `Validate*` functions.
func (lockFile *File) DependencySignatures() map[string]string {
	return lockFile.DependencySignatures_
}

// Signature returns the signature of the resource being locked.
func (lockFile *File) Signature() string {
	return lockFile.Signature_
}

// UpdateSignature updates the signature of a resource being locked.
func (lockFile *File) UpdateSignature(signature string) {
	lockFile.Signature_ = signature
}

// Name returns the name of the resource being locked. This is primarily used
// for debugging; in the case of file locks, it's the path to the file, while
// in tests it's just a string.
func (lockFile *File) Name() string {
	return lockFile.name
}

// IsLocked indicates whether the resource being locked is allowed to be
// overwritten.
func (lockFile *File) IsLocked() bool {
	return lockFile.isLocked
}

// DeserializeLockFile takes an `io.Reader` and transforms that into a
// `lock.File`. No validation occurs. The `name` parameter is so that we can
// print out intelligible errors; sometimes `name` is a path, and other times
// (such as in tests) it is just a string.
func DeserializeLockFile(lockFileReader io.Reader, name string) (Manager, error) {
	lockFileData, deserializeErr := ioutil.ReadAll(lockFileReader)
	if deserializeErr != nil {
		return nil, deserializeErr
	}

	lockFile := File{}
	deserializeErr = yaml.Unmarshal(lockFileData, &lockFile)
	if deserializeErr != nil {
		return nil, deserializeErr
	}

	lockFile.name = name

	return &lockFile, nil
}

// SerializeLockFile takes a `lock.File` and writes it to an `io.Writer`.
func SerializeLockFile(
	lockFile Manager,
	lockFileWriter io.Writer,
) (serializeErr error) {
	serialized, serializeErr := yaml.Marshal(lockFile)
	if serializeErr != nil {
		return serializeErr
	}

	_, serializeErr = lockFileWriter.Write(serialized)

	return
}

// NewCorpusLockFile creates a lockfile representing a cached decompressed
// corpus.
func NewCorpusLockFile(name string, signature string) Manager {
	return &File{
		DependencySignatures_: make(map[string]string),
		Signature_:            signature,
		name:                  name,
		isLocked:              false,
	}
}

// ValidateCorpusLockFile will deserialize and validate a lockfile for a corpus
// dataset.
func ValidateCorpusLockFile(corpusLockFile Manager) error {
	// Corpus lock file:
	// * DependencySignatures is empty.
	// * Signature is the SHA512 of every datafile inside the corpus.

	if len(corpusLockFile.DependencySignatures()) != 0 {
		return errors.New("Corpus lock file contains dependencies, but " +
			"corpus lock file must have no dependencies")
	}

	if corpusLockFile.Signature() == "" {
		return errors.New("Corpus lock file does not contain a signature for " +
			"corpus, but this field is required")
	}

	return nil
}

// NewSampleLockFile creates a lockfile representing a cached corpus sample.
func NewSampleLockFile(
	name string,
	signature string,
	corpusSignature string,
) Manager {
	return &File{
		DependencySignatures_: map[string]string{
			corpusKey: corpusSignature,
		},
		Signature_: signature,
		name:       name,
		isLocked:   false,
	}
}

// ValidateSampleLockFile will deserialize and validate the lockfile for a
// sample, and validate it against the corpus lockfile it depends on.
func ValidateSampleLockFile(
	corpusLockFile Manager,
	sampleLockFile Manager,
) error {
	// SampleLock:
	// * DependencySignatures is a dictionary of 1 key, `corpus`, which maps to
	//   the signature of that corpus.
	// * Signature is the SHA512 of every datafile in the sample.

	validationErr := validateSingletonDependency(
		sampleLockFile,
		corpusLockFile,
		corpusKey)
	if validationErr != nil {
		return validationErr
	}

	if sampleLockFile.Signature() == "" {
		return errors.New("Sample lock file did not contain a valid " +
			"signature; the lock file may be corrupt, and you might need to " +
			"re-generate the corpus sample")
	}

	return nil
}

// ValidateConfigLockFile will deserialize and validate the lockfile for a
// BitFunnel runtime configuration, and validate it against the sample lockfile
// it depends on.
func ValidateConfigLockFile(
	sampleLockFile Manager,
	configLockFile Manager,
) error {
	// ConfigLock:
	// * DependencySignatures is a dictionary of 1 key, `sample`, which maps to
	//   the signature of a sample.
	// * Signature contains the SHA512 of every datafile generated by the
	//   config steps (e.g., the termtable, etc.).

	validationErr := validateSingletonDependency(
		configLockFile,
		sampleLockFile,
		sampleKey)
	if validationErr != nil {
		return validationErr
	}

	if configLockFile.Signature() == "" {
		return errors.New("Configuration lock file did not contain a valid " +
			"signature; the lock file may be corrupt, and you might need to " +
			"re-generate the BitFunnel runtime configuration")
	}

	return nil
}

// ValidateExperimentLockFile will deserialize and validate the lockfile for a
// BitFunnel experiment, and validate it against both the sample and
// configuration lockfiles it depends on.
func ValidateExperimentLockFile(
	sampleLockFile Manager,
	configLockFile Manager,
	experimentLockFile Manager,
) error {
	// ExperimentLock:
	// * DependencySignatures is a dictionary of 2 keys: `sample`, which maps
	//   to the signature of that sample, and `config`, which maps to the
	//   signature of that corpus.
	// * Signature is empty?

	if len(experimentLockFile.DependencySignatures()) != 2 {
		return fmt.Errorf("Lockfile for a data sample require "+
			"exactly 1 dependency, but %d were given",
			len(experimentLockFile.DependencySignatures()))
	}

	validateSampleErr := validateDependency(
		experimentLockFile,
		sampleLockFile,
		sampleKey)
	if validateSampleErr != nil {
		return validateSampleErr
	}

	configSampleErr := validateDependency(
		experimentLockFile,
		configLockFile,
		configKey)
	if configSampleErr != nil {
		return configSampleErr
	}

	// NOTE: Not necessary to validate experiment signature, as it is currently
	// unused.

	return nil
}

func normalizeSignature(signature string) string {
	return strings.ToLower(signature)
}

func validateDependency(
	currentLockFile Manager,
	dependencyLockFile Manager,
	key string,
) error {
	rawActualSignature, ok := currentLockFile.DependencySignatures()[key]
	if !ok {
		return fmt.Errorf("Lock file for data sample does not contain a key "+
			"'%s', which is expected to contain the signature for the"+
			"dependency data", corpusKey)
	}

	expectedSignature := normalizeSignature(dependencyLockFile.Signature())
	actualSignature := normalizeSignature(rawActualSignature)
	if expectedSignature == "" {
		return fmt.Errorf("Attempted to parse lock file '%s', but signature "+
			"was missing",
			dependencyLockFile.Name())
	} else if actualSignature == "" {
		return fmt.Errorf("Attempted to parse lock file '%s', but "+
			"dependency for key '%s' was missing",
			currentLockFile.Name(),
			key)
	} else if expectedSignature != actualSignature {
		return fmt.Errorf("Lockfile contains a key '%s', with expected "+
			"signature '%s', but signature of the dependency was actually "+
			"'%s'; you may need to re-generate this data",
			key,
			expectedSignature,
			actualSignature)
	}

	return nil
}

func validateSingletonDependency(
	currentLockFile Manager,
	dependencyLockFile Manager,
	key string,
) error {
	if len(currentLockFile.DependencySignatures()) != 1 {
		return fmt.Errorf("Lockfile requires exactly 1 dependency, but %d "+
			"were given", len(currentLockFile.DependencySignatures()))
	}

	validationErr := validateDependency(
		currentLockFile,
		dependencyLockFile,
		key)
	if validationErr != nil {
		return validationErr
	}

	return nil
}
