package schema

import "errors"

// Sample represents a sample of the corpus. This sample can incorporate size
// constraints on the documents, a random component (i.e., we can choose a
// random subset of the corpus), and a document bound (e.g., we can choose the
// first 5 documents).
//
// NOTE: We implement fields below as pointers to easily test whether they
// exist. This is because the zero values of these parameters are valid
// configuration settings.
type Sample struct {
	Name             string            `yaml:"name"`
	GramSize         *uint             `yaml:"gram-size"`
	MaxDocuments     *uint             `yaml:"max-documents"`
	RandomnessConfig *RandomnessConfig `yaml:"random-sample"`
	SizeConstraint   *SizeConstraint   `yaml:"size-limits"`
}

func (sample *Sample) validate() error {
	if sample.Name == "" {
		return errors.New("Experiment schema specifies a sample without a " +
			"`name` field, but a name is required.")
	}

	return nil
}

// RandomnessConfig encodes instructions for selecting a random subset of a
// corpus. `Seed` seeds the PRNG, and `Fraction` is a number [0,1] that tells
// us how big to make the sample.
type RandomnessConfig struct {
	Seed     int     `yaml:"seed"`
	Fraction float64 `yaml:"Fraction"`
}

// SizeConstraint eoncodes instructions for selecting all documents in the
// corpus with `MinPostings` or more, and `MaxPostings` or fewer (i.e., it's an
// inclusive range).
type SizeConstraint struct {
	MinPostings uint `yaml:"min-postings"`
	MaxPostings uint `yaml:"max-postings"`
}
