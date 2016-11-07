package experiment

import (
	"errors"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/bitfunnel/LabBook/src/corpus"
	"github.com/go-yaml/yaml"
)

// Schema is the concrete type that represents an experiment schema
// that is declared in a YAML file.
type Schema struct {
	BitFunnelSha   string          `yaml:"bitfunnel-commit-hash"`
	LabBookVersion string          `yaml:"lab-book-version"`
	QueryLog       *QueryLog       `yaml:"query-log"`
	ManifestFile   string          `yaml:"manifest-file"`
	ScriptFile     string          `yaml:"script-file"`
	Corpus         []*corpus.Chunk `yaml:"corpus"`
}

type QueryLog struct {
	RawURL string `yaml:"raw-url"`
	URL    *url.URL
	SHA512 string `yaml:"sha512"`
}

// DeserializeSchema deserializes an `Schema` from an `io.Reader`.
func DeserializeSchema(reader io.Reader) (Schema, error) {
	schemaData, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return Schema{}, readErr
	}

	schema := Schema{}
	deserializeErr := yaml.Unmarshal(schemaData, &schema)
	if deserializeErr != nil {
		return Schema{}, deserializeErr
	}

	validationErr := validateSchema(&schema)
	if validationErr != nil {
		return Schema{}, validationErr
	}

	// Parse and populate the URL.
	queryLogURL, parseErr := url.Parse(schema.QueryLog.RawURL)
	if parseErr != nil {
		return Schema{}, parseErr
	}
	schema.QueryLog.URL = queryLogURL

	return schema, nil
}

func validateSchema(schema *Schema) error {
	if schema.BitFunnelSha == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `bitfunnel-commit-hash`, specifying the commit hash " +
			"to run experiment on")
	} else if schema.LabBookVersion == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `lab-book-version`, which is required to specify the " +
			"version of LabBook the experiment is compatible with")
	} else if schema.QueryLog.RawURL == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `raw-url` inside the `query-log` field, specifying URL " +
			"of the query log to retrieve")
	} else if schema.QueryLog.SHA512 == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `sha512` inside the `query-log` field, specifying " +
			"SHA512 hash of the query log to retrieve")
	} else if len(schema.Corpus) < 1 {
		return errors.New("Experiment schema did not contain required " +
			"field `corpus`, specifying corpus to ingest when we run " +
			"the experiment")
	}

	for _, chunk := range schema.Corpus {
		if chunk.Name == "" {
			return errors.New("Experiment schema contained a corpus without " +
				"the mandatory field `name`")
		} else if chunk.SHA512 == "" {
			return errors.New("Experiment schema contained a corpus without " +
				"the mandatory field `sha512`")
		}
	}

	return nil
}
