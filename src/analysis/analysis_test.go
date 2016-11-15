package analysis

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const resultsFilename = "/Users/alex/src/go/src/github.com/bitfunnel/LabBook/out.txt"
const resultsHeader = "query,rows,matches,quadwords,parse,plan,match"

// const rowFilename = "/tmp/expt1/no_verify_out/Frequencies.csv"
const rowFilename = "/tmp/expt1/no_verify_out/RowDensities-0.csv"
const rowHeader = "term,frequency"

const analysisFilename = "/tmp/expt1/analysis.csv"

func Test_SimpleAnalysis(t *testing.T) {
	resultsRecords, readErr := getResultsCsv(t, resultsFilename)
	assert.NoError(t, readErr)

	termData, readErr := getTermData(t, rowFilename)
	assert.NoError(t, readErr)

	analysis := project(t, resultsRecords, termData, 3)
	writeAnalysisCsv(t, analysisFilename, analysis)
}

func project(
	t *testing.T,
	records [][]string,
	termData map[string]*TermDatum,
	fields ...uint,
) [][]string {

	var projectedRecords [][]string
	for _, record := range records {
		numAdditionalFields := 3
		projectedRecord := make([]string, len(fields)+numAdditionalFields)
		projectedRecord[0] = record[0]

		termDatum, _ := termData[record[0]]
		// assert.True(t, ok)
		if termDatum == nil {
			continue
		}

		projectedRecord[1] = strconv.FormatFloat(termDatum.Frequency, 'f', 12, 64)
		projectedRecord[2] = strconv.FormatUint(termDatum.kRank0, 10)

		for i, index := range fields {
			projectedRecord[i+numAdditionalFields] = record[index]
		}
		projectedRecords = append(projectedRecords, projectedRecord)
		// fmt.Println(projectedRecord)
	}
	return projectedRecords
}

type TermDatum struct {
	Frequency float64
	kRank0    uint64
}

func writeAnalysisCsv(t *testing.T, filename string, analysis [][]string) error {
	analysisFile, openErr := os.Create(filename)
	assert.NoError(t, openErr)
	defer analysisFile.Close()

	fileWriter := csv.NewWriter(analysisFile)
	writeErr := fileWriter.WriteAll(analysis)
	assert.NoError(t, writeErr)
	fileWriter.Flush()

	return nil
}

func getTermData(t *testing.T, filename string) (termData map[string]*TermDatum, readErr error) {
	resultsFile, outErr := os.Open(filename)
	assert.NoError(t, outErr)
	defer resultsFile.Close()

	resultsData, readErr := ioutil.ReadAll(resultsFile)
	assert.NoError(t, readErr)

	// NOTE: We must parse each line individually as a separate CSV because the
	// row densities CSV schema has variable line length, and golang's `csv`
	// package can't handle such things.
	fileLines := strings.Split(string(resultsData), "\n")
	termData = make(map[string]*TermDatum)
	for _, line := range fileLines {
		// Parse a single (variable-length) line as a CSV.
		csvData := bytes.NewBufferString(line)
		csvReader := csv.NewReader(csvData)
		record, readErr := csvReader.ReadAll()
		assert.NoError(t, readErr)
		if len(record) == 0 {
			continue
		}
		assert.EqualValues(t, 1, len(record))

		// Count k, the number of hash functions allocated to query.
		var k uint64 = 0
		allocatedHighRankRows := false
		for i, element := range record[0][2:] {
			elementIsRankDescription := i%3 == 0
			if elementIsRankDescription && element == "r0" {
				k++
			} else if elementIsRankDescription && element != "r0" {
				allocatedHighRankRows = true
				break
			}
		}

		// Add if it's not allocated high-rank rows.
		if !allocatedHighRankRows {
			freq, parseErr := strconv.ParseFloat(record[0][1], 64)
			assert.NoError(t, parseErr)
			termData[record[0][0]] = &TermDatum{Frequency: freq, kRank0: k}
			// fmt.Printf("'%s'", record[0][0])
		}
	}

	// resultsText := rowHeader + "\n" + string(resultsData)

	// csvData := bytes.NewBufferString(resultsText)
	// csvReader := csv.NewReader(csvData)
	// records, readErr := csvReader.ReadAll()
	// assert.NoError(t, readErr)

	// frequencies = make(map[string]float64)
	// for _, record := range records[1:] {
	// 	freq, parseErr := strconv.ParseFloat(record[1], 64)
	// 	assert.NoError(t, parseErr)
	// 	frequencies[record[0]] = freq
	// }

	return
}

func getResultsCsv(t *testing.T, filename string) (records [][]string, readErr error) {
	resultsFile, outErr := os.Open(filename)
	assert.NoError(t, outErr)
	defer resultsFile.Close()

	resultsData, readErr := ioutil.ReadAll(resultsFile)
	assert.NoError(t, readErr)

	// TODO: This newline probably needs to be escaped.
	fileLines := strings.Split(string(resultsData), "\n")
	var csvStrings = []string{resultsHeader}
	for i, line := range fileLines {
		if strings.HasPrefix(line, "Results:") {
			// Add a column, "word", to the beginning of the record. This code
			// is somewhat opaque, so here is an example record we're parsing:
			//   Results:
			//   rows,matches,quadwords,parse,plan,match
			//   2,210,25,5.23e-07,7.457e-06,1.2709e-05
			//   741: query one also
			//   Processing query " also"
			queryLine := fileLines[i+4]
			query := queryLine[19 : len(queryLine)-1]

			csvStrings = append(csvStrings, query+","+fileLines[i+2])
		}
	}

	csvData := bytes.NewBufferString(strings.Join(csvStrings, "\n"))
	csvReader := csv.NewReader(csvData)
	records, readErr = csvReader.ReadAll()
	assert.NoError(t, readErr)
	return
}
