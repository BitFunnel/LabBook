package schema

import "errors"

// StatisticsConfig maintains a configuration for building corpus statistics
// the and termtable.
//
// NOTE: We implement fields below as pointers to easily test whether they
// exist. This is because the zero values of these parameters are valid
// configuration settings.
type StatisticsConfig struct {
	SampleName string `yaml:"sample-name"`
	GramSize   *uint  `yaml:"gram-size"`
}

func (config *StatisticsConfig) validate() error {
	if config.SampleName == "" {
		return errors.New("Runtime configuration requires the `sample-name` " +
			"field to be populated")
	}

	return nil
}

// RuntimeConfig maintains a configuration and running BitFunnel.
//
// NOTE: We implement fields below as pointers to easily test whether they
// exist. This is because the zero values of these parameters are valid
// configuration settings.
type RuntimeConfig struct {
	SampleName    string `yaml:"sample-name"`
	GramSize      *uint  `yaml:"gram-size"`
	IngestThreads *uint  `yaml:"ingest-threads"`
}

func (config *RuntimeConfig) validate() error {
	if config.SampleName == "" {
		return errors.New("Runtime configuration requires the `sample-name` " +
			"field to be populated")
	}

	return nil
}
