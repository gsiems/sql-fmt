package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestSQLFiles(t *testing.T) {

	baseDir := path.Join("..", "testdata")

	//dialects := []string{"mariadb", "mssql", "mysql", "oracle", "postgresql", "sqlite", "standard"}

	dataDir := path.Join(baseDir, "input")

	rd, err := os.ReadDir(dataDir)
	if err != nil {
		t.Error(err)
		return
	}

	for _, f := range rd {

		if !f.IsDir() {
			continue
		}

		d := f.Name()

		inputDir := path.Join(dataDir, d)
		parsedDir := path.Join(baseDir, "parsed")

		files, err := os.ReadDir(inputDir)
		if err != nil {
			t.Error(err)
		}

		for _, file := range files {
			// Ensure that it is a *.sql file
			if !strings.HasSuffix(file.Name(), ".sql") {
				continue
			}

			inputFile := path.Join(inputDir, file.Name())

			inBytes, err := ioutil.ReadFile(inputFile)
			if err != nil {
				t.Errorf("%s (%s)", file.Name(), err)
			}
			input := string(inBytes)

			p := NewParser(d)

			////////////////////////////////////////////////////////////////////////
			var parsed []Token
			parsed, err = p.ParseStatements(input)
			if err != nil {
				t.Errorf("Error parsing input for %s (%s)", file.Name(), err)
			}
			var z []string

			for _, tc := range parsed {
				if tc.vSpace > 0 {
					z = append(z, strings.Repeat("\n", tc.vSpace))
				}
				if tc.hSpace != "" {
					z = append(z, tc.hSpace)
				}
				z = append(z, tc.Value())
			}

			reconsFile := path.Join(parsedDir, "actual", d, file.Name()+".reconstructed")

			err = writeReconstructed(reconsFile, strings.Join(z, ""))
			if err != nil {
				t.Errorf("Error writing reconstructed for %s: %s", file.Name(), err)
			}

			reconsBytes, err := ioutil.ReadFile(reconsFile)
			if err != nil {
				t.Errorf("Error reading reconstructed for %s: %s", file.Name(), err)
			} else {
				if strings.Compare(strings.TrimRight(string(inBytes), "\n\r\t "),
					strings.TrimRight(string(reconsBytes), "\n\r\t ")) != 0 {
					t.Errorf("Input vs reconstructed failed for %q", file.Name())
				}
			}

			err = writeParsed(parsedDir, d, file.Name(), parsed)
			if err != nil {
				t.Errorf("Error writing parsed for %s: %s", file.Name(), err)
				continue
			}

		}
	}
}

func compareFiles(dir, d, fName string) error {

	actFile := path.Join(dir, "actual", d, fName)
	expFile := path.Join(dir, "expected", d, fName)

	actBytes, err := ioutil.ReadFile(actFile)
	if err != nil {
		return err
	}

	expBytes, err := ioutil.ReadFile(expFile)
	if err != nil {
		return err
	}

	if strings.Compare(string(actBytes), string(expBytes)) != 0 {
		return fmt.Errorf("Actual vs expected failed for %q", fName)
	}

	return err
}

func writeReconstructed(reconsFile, reconstructed string) error {

	f, err := os.OpenFile(reconsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(reconstructed))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return err
}

func writeParsed(dir, d, fName string, parsed []Token) error {

	outFile := path.Join(dir, "actual", d, fName)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	var toks []string
	toks = append(toks, "Parsed")
	toks = append(toks, fmt.Sprintf("InputFile   %s", fName))
	toks = append(toks, fmt.Sprintf("Dialect     %s", d))
	toks = append(toks, "")

	for _, t := range parsed {
		toks = append(toks, t.String())
	}

	_, err = f.Write([]byte(strings.Join(toks, "\n") + "\n"))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return err
}
