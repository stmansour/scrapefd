package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// reCSZ is the regular expression to match City, State, and Zip in the scraped data
var reCSZ = string("^([^,]+),\\s*([A-Z]{2})\\s*([0-9]+)")

// CSVColumn defines a column of the CSV file
type CSVColumn struct {
	Name  string
	Index int
}

// CsvErrLoose et al, are constants used to control whether an error on a single line causes
// the entire CSV process to terminate or continue.   If LOOSE, then it will skip the error line
// and continue to process the remaining lines.  If STRICT, then the entire CSV loading process
// will terminate if any error is encountered
const (
	CsvErrLoose  = 0
	CsvErrStrict = 1
)

// CsvErrorSensitivity is the error return value used by all the loadXYZcsv.go routines. We
// initialize to LOOSE as it is best for testing and should be OK for normal use as well.
var CsvErrorSensitivity = int(CsvErrLoose)

// Stripchars returns a string with the characters from chars removed
func Stripchars(str, chars string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chars, r) < 0 {
			return r
		}
		return -1
	}, str)
}

// csvQuote checks the supplied string to see if it includes a comma (,) if it does
// the return value is a quoted (%q) version. Otherwise it just returns the input string.
// Furthermore, it removes any quotes it finds in the data before adding quotes.  That
// works for this data.
func csvQuote(s string) string {
	i := strings.LastIndex(Stripchars(s, "\""), ",")
	if i >= 0 {
		return "\"" + s + "\""
	}
	return s
}

// LoadCSV loads a comma-separated-value file into an array of strings and returns the array of strings
func LoadCSV(fname string) [][]string {
	t := [][]string{}
	f, err := os.Open(fname)
	if nil == err {
		defer f.Close()
		reader := csv.NewReader(f)
		reader.FieldsPerRecord = -1
		rawCSVdata, err := reader.ReadAll()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, sa := range rawCSVdata {
			t = append(t, sa)
		}
	} else {
		fmt.Printf("LoadCSV: could not open CSV file. err = %v\n", err)
		os.Exit(1)
	}
	return t
}

// ValidateCSVColumnsErr verifies the column titles with the supplied, expected titles.
// Returns:
//   bool --> false = everything is OK,  true = at least 1 column is wrong, error message already printed
//   err  --> nil if no problems
func ValidateCSVColumnsErr(csvCols []CSVColumn, sa []string, funcname string, lineno int) (bool, error) {
	required := len(csvCols)
	if len(sa) < required {
		l := len(sa)
		for i := 0; i < len(csvCols); i++ {
			if i < l {
				s := Stripchars(strings.ToLower(strings.TrimSpace(sa[i])), " ")
				if s != strings.ToLower(csvCols[i].Name) {
					return true, fmt.Errorf("%s: line %d - Error at column heading %d, expected %s, found %s\n", funcname, lineno, i, csvCols[i].Name, sa[i])
				}
			}
		}
		return true, fmt.Errorf("%s: line %d - found %d values, there must be at least %d\n", funcname, lineno, len(sa), required)
	}

	if lineno == 1 {
		for i := 0; i < len(csvCols); i++ {
			s := Stripchars(strings.ToLower(strings.TrimSpace(sa[i])), " ")
			if s != strings.ToLower(csvCols[i].Name) {
				return true, fmt.Errorf("%s: line %d - Error at column heading %d, expected %s, found %s\n", funcname, lineno, i, csvCols[i].Name, sa[i])
			}
		}
	}
	return false, nil
}
