package schema

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/BitFunnel/LabBook/src/corpus"
	"github.com/BitFunnel/LabBook/src/signature"
	"github.com/go-yaml/yaml"
)

// Experiment is the concrete type that represents an experiment schema
// that is declared in a YAML file.
//
// NOTE: We implement fields below as pointers to easily test whether they
// exist. This is because the zero values of these parameters are valid
// configuration settings.
type Experiment struct {
	BitFunnelSha     string                `yaml:"bitfunnel-commit-hash"`
	LabBookVersion   string                `yaml:"lab-book-version"`
	QueryLog         *QueryLog             `yaml:"query-log"`
	Corpus           []*corpus.ArchiveFile `yaml:"corpus"`
	Samples          []*Sample             `yaml:"samples"`
	StatisticsConfig *StatisticsConfig     `yaml:"statistics-config"`
	RuntimeConfig    *RuntimeConfig        `yaml:"runtime-config"`
}

// DeserializeExperimentSchema deserializes an `Schema` from an `io.Reader`.
func DeserializeExperimentSchema(reader io.Reader) (Experiment, error) {
	schemaData, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return Experiment{}, readErr
	}

	schema := Experiment{}
	deserializeErr := yaml.Unmarshal(schemaData, &schema)
	if deserializeErr != nil {
		return Experiment{}, deserializeErr
	}

	validationErr := schema.validate()
	if validationErr != nil {
		return Experiment{}, validationErr
	}

	return schema, nil
}

func (experiment *Experiment) validate() error {
	if experiment.BitFunnelSha == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `bitfunnel-commit-hash`, specifying the commit hash " +
			"to run experiment on")
	} else if experiment.LabBookVersion == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `lab-book-version`, which is required to specify the " +
			"version of LabBook the experiment is compatible with")
	} else if experiment.StatisticsConfig == nil {
		return errors.New("Experiment schema did not contain a " +
			"`statistics-config` field, but this is required.")
	} else if experiment.RuntimeConfig == nil {
		return errors.New("Experiment schema did not contain a " +
			"`runtime-config` field, but this is required.")
	} else if len(experiment.Corpus) < 1 {
		return errors.New("Experiment schema did not contain required " +
			"field `corpus`, specifying corpus to ingest when we run " +
			"the experiment")
	}

	queryLogValidationErr := experiment.QueryLog.validateAndDefault()
	if queryLogValidationErr != nil {
		return queryLogValidationErr
	}

	statisticsConfigValidationErr := experiment.StatisticsConfig.validate()
	if statisticsConfigValidationErr != nil {
		return statisticsConfigValidationErr
	}

	runtimeConfigValidationErr := experiment.RuntimeConfig.validate()
	if runtimeConfigValidationErr != nil {
		return runtimeConfigValidationErr
	}

	sampleNameErr := verifySampleNames(experiment)
	if sampleNameErr != nil {
		return sampleNameErr
	}

	for _, chunk := range experiment.Corpus {
		if chunk.Name == "" {
			return errors.New("Experiment schema contained a corpus without " +
				"the mandatory field `name`")
		} else if chunk.SHA512 == "" {
			return errors.New("Experiment schema contained a corpus without " +
				"the mandatory field `sha512`")
		}

		chunk.SHA512 = signature.Normalize(chunk.SHA512)
	}

	return nil
}

func verifySampleNames(experiment *Experiment) error {
	sampleNames := map[string]bool{}
	for _, sample := range experiment.Samples {
		if _, containsName := sampleNames[sample.Name]; containsName {
			return errors.New("Experiment schema requires that all sample " +
				"names are unique")
		}
		sampleNames[sample.Name] = true
	}

	if _, containsName := sampleNames[experiment.StatisticsConfig.SampleName]; !containsName {
		return fmt.Errorf("Experiment schema contains a statistics "+
			"configuration with name '%s', which is not defined in the schema",
			experiment.StatisticsConfig.SampleName)
	}

	if _, containsName := sampleNames[experiment.RuntimeConfig.SampleName]; !containsName {
		return fmt.Errorf("Experiment schema contains a runtime "+
			"configuration with name '%s', which is not defined in the schema",
			experiment.RuntimeConfig.SampleName)
	}

	return nil
}
