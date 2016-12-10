// Scrape the OK Fire Department website for names, addresses, and email addresses.
// Create a CSV file with the results.  Much of this code was "borrowed" from the
// RentRoll csv library.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"rentroll/rlib"
	"strings"
)

// FDRecord encapsulates the full set of fields that we want for each fire department
type FDRecord struct {
	FDID       int
	Name       string
	County     string
	Chief      string
	Phone      string
	Fax        string
	Address    string
	Address2   string
	City       string
	State      string
	PostalCode string
}

// FDList is the map indexed by the fire department name and returns its associated FDRecord.
// It is useful in that the only link between the different pages on their website is the
// fire department name.
var FDList = map[string]FDRecord{}

// Search the HTML for the FD Name, the address, city, state, and zip.
// Update the internal map with this new info.
//
// returns: 0 - success
//			1 - could not find start string
//  		2 - could not find stop string
//			3 - could not bracket FD Name
//			4 - could not bracket FD address
//			5 - could not find FDList entry for FD Name
func scrapeHTML(src string, FDID int) int {
	i1 := strings.Index(src, "id=\"fdcontact\"")
	if i1 < 0 {
		return 1
	}
	i2 := strings.Index(src, "id=\"fdresources\"")
	if i2 < 0 {
		return 2
	}
	s := src[i1:i2]

	// The name is in <h3> </h3> brackets
	i1 = strings.Index(s, "<h3>")
	i2 = strings.Index(s, "</h3>")
	if i1 < 0 || i2 < 0 {
		return 3
	}
	name := strings.TrimSpace(s[i1+4 : i2])

	// The address is after <h4>Contact Information:</h4>
	// address lines are separated by <br />
	// address is terminated by Ph:
	target := "<h4>Contact Information:</h4>"
	i1 = strings.Index(s, target)
	i2 = strings.Index(s, "Ph:")
	if i1 < 0 || i2 < 0 {
		return 4
	}

	raw := s[i1+len(target) : i2]
	addr := strings.Split(raw, "<br />")
	for i := 0; i < len(addr); i++ {
		addr[i] = strings.TrimSpace(addr[i])
	}

	r, _ := regexp.Compile(reCSZ)
	m := r.FindAllStringSubmatchIndex(addr[1], -1)
	city := addr[1][m[0][2]:m[0][3]]
	state := addr[1][m[0][4]:m[0][5]]
	zip := addr[1][m[0][6]:m[0][7]]

	f, ok := FDList[name]
	if !ok {
		fmt.Printf("*** Unable to find FDList entry for %q\n", name)
		return 5
	}
	f.FDID = FDID
	f.Address = csvQuote(strings.TrimSpace(addr[0]))
	f.City = csvQuote(strings.TrimSpace(city))
	f.State = csvQuote(strings.TrimSpace(state))
	f.PostalCode = csvQuote(strings.TrimSpace(zip))
	FDList[name] = f
	return 0
}

func getHTML(url string, FDID int) error {
	r, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("getHTML(%s) error = %s\n", url, err.Error())
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	status := scrapeHTML(string(body), FDID)
	if status != 0 {
		return fmt.Errorf("Problem finding scrape markers for FDID=%d: %d\n", FDID, status)

	}
	return nil
}

func buildContactList() {
	baseURL := "http://www.firereporting.ok.gov/directory/detail.aspx?fdid=%d"
	for i := 1; i < len(FDList); i++ {
		url := fmt.Sprintf(baseURL, i)
		err := getHTML(url, i)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}
}

// Name, et al, are the fields we expect to find in the csv file
const (
	Name   = 0
	County = iota
	Chief  = iota
	Phone  = iota
	Fax    = iota
)

// csvCols is an array that defines all the columns that should be in this csv file
var csvCols = []CSVColumn{
	{"Name", Name},
	{"County", County},
	{"Chief", Chief},
	{"Phone", Phone},
	{"Fax", Fax},
}

// processfd uses the csv header line validation code used by RentRoll's csv importers.
// It may be overkill, but the code was already written.
func processfd(sa []string, lineno int) (int, error) {
	y, err := ValidateCSVColumnsErr(csvCols, sa, "processfd", lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	var f FDRecord
	name := strings.TrimSpace(sa[Name]) // no quotes on the FDList key value
	f.Name = csvQuote(name)
	f.County = csvQuote(strings.TrimSpace(sa[County]))
	f.Chief = csvQuote(strings.TrimSpace(sa[Chief]))
	f.Phone = csvQuote(strings.TrimSpace(sa[Phone]))
	f.Fax = csvQuote(strings.TrimSpace(sa[Fax]))
	FDList[name] = f
	return 0, nil
}

// ingestcsv reads in a file named "fd.csv" -- this is the excel spreadsheet that
// you can download
func ingestcsv() {
	t := rlib.LoadCSV("fd.csv")
	for i := 0; i < len(t); i++ {
		s, err := processfd(t[i], i+1)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
		if s > 0 { // if handler indicates that we need to stop...
			break //... then exit out of the loop
		}
	}
}

func generateCSV() {
	fmt.Printf("FDID,Name,County,Chief,Phone,Fax,Address,Address2,City,State,PostalCode\n")
	for _, v := range FDList {
		fmt.Printf("%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n", v.FDID, v.Name, v.County, v.Chief, v.Phone, v.Fax, v.Address, v.Address2, v.City, v.State, v.PostalCode)
	}
}

func main() {
	ingestcsv()
	buildContactList()
	generateCSV()
}
