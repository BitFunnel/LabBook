package schema

import (
	"errors"
	"net/url"
)

// QueryLog contains information about the query log; where to obtain it, and a
// SHA of the contents.
type QueryLog struct {
	RawURL string `yaml:"raw-url"`
	URL    *url.URL
	SHA512 string `yaml:"sha512"`
}

func (queryLog *QueryLog) validateAndDefault() error {
	if queryLog.RawURL == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `raw-url` inside the `query-log` field, specifying URL " +
			"of the query log to retrieve")
	} else if queryLog.SHA512 == "" {
		return errors.New("Experiment schema did not contain required " +
			"field `sha512` inside the `query-log` field, specifying " +
			"SHA512 hash of the query log to retrieve")
	}

	// Parse and populate the URL.
	queryLogURL, parseErr := url.Parse(queryLog.RawURL)
	if parseErr != nil {
		return parseErr
	}
	queryLog.URL = queryLogURL

	return nil
}