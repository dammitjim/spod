package main

import (	
	"fmt"
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"
	"database/sql"
)

var db *sql.DB
var err error
var implementations []Implementation

func main() {

	// The maximum depth it will crawl until, or zero for infinite depth
	maxDepth := 0

	// Spin up the and add it to the implementations slice
	//basicImplementation := new(BasicImplementation)
	//exampleImplementation := new(ExampleImplementation)
	gcloudImplementation := new(GcloudImplementation)
	implementations = append(implementations, gcloudImplementation)

	fmt.Print("Formatting environment\n")
	os.Remove("./data.sqlite")
	fmt.Print("Opening database\n")

	// Open the database
	db, err = sql.Open("sqlite3", "file:data.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Make sure the tables exist
	sqlStmt := `CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		parent INTEGER,
		uri TEXT,
		depth INTEGER DEFAULT 0,		
		added DATETIME DEFAULT CURRENT_TIMESTAMP,
		failures INTEGER DEFAULT 0,
		last_crawl DATETIME DEFAULT 0,
		next_crawl DATETIME DEFAULT 0);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	

	//tests()
	 for n, _ := range(implementations) {
		implementations[n].prep()
	}

	// Allow an URL to be added into the system
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		log.Print("Adding seed")
		link := *NewLink(args[0]) 		
		addLink(link)
	}

	// Seed it based on the implementation
	 for n, _ := range(implementations) {
		implementations[n].seed()
	}

	// Spawn 3 spiders
	var spiders [3]Spider
	spiders[0] = *NewSpider("Bob")
	spiders[1] = *NewSpider("Ren")
	spiders[2] = *NewSpider("Pat")  	

	completed := false

	for completed == false {

		var wg sync.WaitGroup

		// Iterate across the spiders and keep them busy
		for _,spider := range spiders {

			link := Link{}
			if (link.loadDue(maxDepth)) {

				wg.Add(1)
				go func(spider Spider, link Link) {
					spider.link = link
					if (spider.link.uri != "") {
						//fmt.Printf("%s going to %s\n", spider.name, spider.link.uri)
						spider.crawl()
					} else {
						//fmt.Printf("%s has nowhere to go\n", spider.name)
					}
					wg.Done()
				}(spider, link)

			}

		}

		wg.Wait()

		// Check to see if it's completed yet
		remainingLinks := countLinks(maxDepth)

		clear()
		fmt.Printf("%d Links remaining\n", remainingLinks)
		if (remainingLinks == 0) {
			completed = true
		}

	}

	fmt.Printf("All done, go home and be a family man")
}

func clear() {

	c := exec.Command("cmd", "/c", "cls")
	c.Stdout = os.Stdout
	c.Run()

}

func countLinks(depth int) (count int) {

	count = 0
	if (depth > 0) {
		err := db.QueryRow("SELECT COUNT(*) as count FROM links WHERE next_crawl <= CURRENT_TIMESTAMP AND depth < ?", depth).Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := db.QueryRow("SELECT COUNT(*) as count FROM links WHERE next_crawl <= CURRENT_TIMESTAMP", depth).Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	}

	return count

}


func crawling_completed(spider * Spider) {

}

// Add a link into the system
func addLink(link Link) {

	// Check the URI isnt already assigned to another link
	existingLink := Link{}
	if (existingLink.load(link.uri)) {
		if (existingLink.depth > link.depth) {
			existingLink.depth = link.depth
			_ = existingLink.save()		
		}
		return
	}

	_ = link.save()	

}
