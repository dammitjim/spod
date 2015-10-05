package main

import (
	"log"	
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"fmt"	
	"path"
	_ "github.com/mattn/go-sqlite3"	
)

type Link struct {
	uri 	 string	// Resource URI the link goes to
	depth 	 int	// The lowest depth it took a spider to reach this link
	id		 int	// Auto increment
	parent	 int	// id for the link which led to this (shortest route)
	failures int	// number of times the crawl failed sequentially
	new 	 bool
}

func NewLink(uri string) *Link {
	l := new(Link)
	l.uri = uri
	l.new = true
	l.depth = 0
	l.id = 0
	l.parent = 0
	l.failures = 0
	return l
}


func (l *Link) saveData(data []byte) (bool) {


	hasher := md5.New()
    hasher.Write([]byte(l.uri))
    filename := hex.EncodeToString(hasher.Sum(nil))
    extension := path.Ext(l.uri)

    if (extension == "") {
    	extension = ".html"
    }

	// Write it as a binary blob
	filepath := fmt.Sprintf("data/%s%s", filename, extension)	// md5.Sum([]byte(l.uri))
   	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		log.Printf("%q: %s\n", err)
		return false
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

	_, err = stmt.Exec(l.uri, filepath)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()

	//@TODO - write to the 'data' table to keep the link

	return true

}

func (l *Link) loadDue(maxDepth int) (bool) {

	// Try to find a link due.
	var uri string
	err := db.QueryRow("SELECT uri FROM links WHERE next_crawl <= CURRENT_TIMESTAMP AND depth < ? ORDER BY next_crawl ASC LIMIT 1", maxDepth).Scan(&uri)
	if err != nil {
		return false
	} else {

		l.load(uri)

		//@TODO - Refactor to use native go code rather than sql
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}	

		stmt, err := tx.Prepare("UPDATE links SET last_crawl=CURRENT_TIMESTAMP, next_crawl=(datetime('now', '+7 days')) WHERE id=?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(l.id)
		if err != nil {
			log.Fatal(err)
		}	

		tx.Commit()	

		return true

	}

}

func (l *Link) save() (bool) {

	// If its a new link..
	if (l.new) {

		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err := tx.Prepare("insert into links(uri, depth, failures, parent) values(?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(l.uri, l.depth, l.failures, l.parent)
		if err != nil {
			log.Fatal(err)
		}

		tx.Commit()

		//@TODO - temporary until I get LastInsertId() working
		l.load(l.uri);		
		//l.id = stmt.LastInsertId()		

	} else {
		
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}	

		stmt, err := tx.Prepare("UPDATE links SET uri=?, depth=?, failures=?, parent=? WHERE id=?")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(l.uri, l.depth, l.failures, l.parent, l.id)
		if err != nil {
			log.Fatal(err)
		}	

		tx.Commit()			

	}

	return true

}


func (l *Link) load(uri string) (bool) {

	// Reset some defaults
	l.id = 0
	l.uri = ""
	l.depth = 0
	l.failures = 0
	l.parent = 0
	l.new = false

	err := db.QueryRow("SELECT id, uri, depth, failures FROM links WHERE uri = ?", uri).Scan(&l.id, &l.uri, &l.depth, &l.failures);
	if err != nil {
		return false
	}	

	return true

}

