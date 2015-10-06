package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
)

/**
 * Example implementation
 * This will crawl a list of page results and the result detail pages
 * Data is saved into a /data/ folder and indexed in a new database table called 'data'
 */

type ExampleImplementation struct {
}

func (i ExampleImplementation) processUri(uri string) string {

	// Strip out anchors from the URL
	reg, _ := regexp.Compile("#(.+)")
	uri = reg.ReplaceAllString(uri, "")
	return uri

}

func (i ExampleImplementation) seed() {

	// Seed the drupal search results
	link := *NewLink("https://www.example.com/search?q=drupal")
	addLink(link)

}

func (i ExampleImplementation) prep() {

	// Make sure the tables exist
	sqlStmt := `CREATE TABLE IF NOT EXISTS data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uri TEXT,	
		filename TEXT,	
		added DATETIME DEFAULT CURRENT_TIMESTAMP);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

}

func (i ExampleImplementation) shouldFollowLink(link Link) bool {

	// Allow the usage of pagers
	if strings.Contains(link.uri, "search?q=drupal&page=") {
		return true
	}

	// Allow service details page
	if strings.Contains(link.uri, "/result/") {
		return true
	}

	return false

}

func (i ExampleImplementation) parseHTML(link Link, node *html.Node) {

}

func (i ExampleImplementation) parseRaw(link Link, data []byte) {

	hasher := md5.New()
	hasher.Write([]byte(link.uri))
	filename := hex.EncodeToString(hasher.Sum(nil))
	extension := path.Ext(link.uri)

	if extension == "" {
		extension = ".html"
	}

	// Write it as a binary blob
	filepath := fmt.Sprintf("data/%s%s", filename, extension)
	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		log.Printf("%q: %s\n", err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into data(uri, filename) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(link.uri, filepath)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

}
