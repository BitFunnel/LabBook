package lock

import "github.com/BitFunnel/LabBook/src/signature"

// TODO: Replace uses of `File` with `Manager` instead.

// Manager is a general interface for managing all the information we need to
// verify that a cached dataset is the dataset we ran an experiment on. In
// particular, it contains a signature for any dependency steps, as well as a
// signature for the current step. For example, the BitFunnel runtime
// configuration step depends on a specific version of a corpus sample; that
// sample will have a signature, and the signature in that lock.File must fit
// the signature listed here.
type Manager interface {
	DependencySignatures() map[string]signature.Signature
	Signature() signature.Signature
	UpdateSignature(signature signature.Signature)
	Name() string
	IsLocked() bool
}

// File (i.e., lock.File) is an specific implementation of the `lock.Manager`
// interface, that centers around a YAML specification that can be serialized
// deserialized to and from disk.
type File struct {
	DependencySignatures_ map[string]signature.Signature `yaml:"dependency-signatures"`
	Signature_            signature.Signature            `yaml:"signature"`
	name                  string
	isLocked              bool
}

// DependencySignatures returns a map containing all dependencies of the
// resource being locked, and their signatures. What the keys are depends on
// the context, and should be largly opaque, as it's not intended to be
// manipulated. You can see the schema by looking at the `Validate*` functions.
func (lockFile *File) DependencySignatures() map[string]signature.Signature {
	return lockFile.DependencySignatures_
}

// Signature returns the signature of the resource being locked.
func (lockFile *File) Signature() signature.Signature {
	return lockFile.Signature_
}

// UpdateSignature updates the signature of a resource being locked.
func (lockFile *File) UpdateSignature(signature signature.Signature) {
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

func (lockFile *File) validateAndDefault(name string) error {
	lockFile.name = name

	// We deserialize right to a string without calling `New`, so we need
	// to normalize the signature.
	lockFile.Signature_.Normalize()
	for _, dependencySignature := range lockFile.DependencySignatures_ {
		dependencySignature.Normalize()
	}

	return nil
}
