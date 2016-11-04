package experiment

import (
	"io"
	"io/ioutil"

	"github.com/bitfunnel/bf-lab-notebook/src/corpus"
	"github.com/go-yaml/yaml"
)

// Schema is the concrete type that represents an experiment schema
// that is declared in a YAML file.
type Schema struct {
	BfSha        string         `yaml:"bf-sha"`
	QueryLog     string         `yaml:"query-log-file"`
	ManifestFile string         `yaml:"manifest-file"`
	ScriptFile   string         `yaml:"script-file"`
	Corpus       []corpus.Chunk `yaml:"corpus"`
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

	// TODO: Validate schema.

	return schema, nil
}
