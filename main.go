package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Sitemapindex ...
type Sitemapindex struct {
	Locations []string `xml:"sitemap>loc"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var s Sitemapindex

	resp, err := http.Get("https://www.ebay-kleinanzeigen.de/sitemap_index.xml")
	check(err)
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)

	for _, Location := range s.Locations {
		resp, err := http.Get(Location)
		check(err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			m, c := WordCount(bodyString)

			f, err := os.OpenFile("ebay.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			check(err)
			defer f.Close()
			w := bufio.NewWriter(f)
			_, err = fmt.Fprintf(w, "{URL: '%v', Words: '%v', WordsCount: %v}", Location, m, len(c))
			check(err)
			w.Flush()
		}
	}

}

// WordCount ...
func WordCount(s string) (map[string]int, []string) {
	words := s

	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsSymbol(c)
	}
	// Convert HTML tags to lower case.
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	words = re.ReplaceAllStringFunc(words, strings.ToLower)

	// Remove STYLE.
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	words = re.ReplaceAllString(words, "")

	// Remove SCRIPT.
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	words = re.ReplaceAllString(words, "")

	// Remove all HTML code in angle brackets, and replace with newline.
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	words = re.ReplaceAllString(words, "\n")

	re, _ = regexp.Compile("\\s{2,}")
	words = re.ReplaceAllString(words, "\n")

	word := strings.FieldsFunc(words, f)

	m := make(map[string]int)

	for _, w := range word {
		m[w]++
	}
	return m, word
}
