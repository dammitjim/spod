package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var db *sql.DB
var err error
var implementations []Implementation
var names []string
var namesInUse []bool

func main() {

	// The maximum depth it will crawl until, or zero for infinite depth
	maxDepth := 0

	// Add names here for your spiders :)
	// Note that if you limit the concurrency to 3, only the first 3 will be used
	names = append(names, "Spike", "Jet", "Fay")

	// Building usage array for later
	for i := 0; i < len(names); i++ {
		namesInUse = append(namesInUse, false)
	}

	// Spin up the and add it to the implementations slice
	//basicImplementation := new(BasicImplementation)
	//exampleImplementation := new(ExampleImplementation)
	gcloudImplementation := new(GcloudImplementation)
	//implementations = append(implementations, exampleImplementation)
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
	for n, _ := range implementations {
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
	for n, _ := range implementations {
		implementations[n].seed()
	}

	completed := false

	// Define how many concurrent spiders
	concurrency := 3

	// Channel used to throttle the amount of goroutines
	tasks := make(chan bool, concurrency)

	for completed == false {

		// Get initial link
		link := Link{}
		if link.loadDue(maxDepth) {

			// Occupy a slot in the channel
			tasks <- true

			name, e := getAvailableName()
			if e != nil {
				fmt.Print(e)
			}

			// Our spider needs a unique name!
			var spider = *NewSpider(name)

			// Send our spider off to crawl in a goroutine
			go func(spider Spider, link Link) {

				spider.link = link

				if spider.link.uri != "" {

					fmt.Printf("%s going to %s\n", spider.name, spider.link.uri)
					// Returns the crawl results as an array to be saved
					linksToSave := spider.crawl()

					// After this crawl has finished
					defer func(links []Link) {

						// Save links in main thread
						for _, link := range links {
							addLink(link)
						}

						// Calculate how much is remaining
						remainingLinks := countLinks(maxDepth)
						fmt.Printf("%d Links remaining\n", remainingLinks)

						if remainingLinks == 0 {
							completed = true
						}

						// Free up the name for use
						if key, e := getArrayIndexByValue(names, spider.name); e == nil {
							namesInUse[key] = false
						} else {
							fmt.Print(e)
						}

						// Free up the task
						<-tasks
					}(linksToSave)
				} else {
					// TODO return error
					if key, e := getArrayIndexByValue(names, spider.name); e == nil {
						namesInUse[key] = false
					} else {
						fmt.Print(e)
					}
					<-tasks
				}
			}(spider, link)
		}
	}

	fmt.Printf("All done, go home and be a family man")
}

func getAvailableName() (string, error) {
	for key, inUse := range namesInUse {
		if inUse == false {
			namesInUse[key] = true
			return names[key], nil
		}
	}
	return "", errors.New("Nobody is currently available")
}

func getArrayIndexByValue(array []string, value string) (int, error) {
	for k, v := range array {
		if v == value {
			return k, nil
		}
	}
	return 0, errors.New("Value not found in array")
}

func clear() {

	c := exec.Command("cmd", "/c", "cls")
	c.Stdout = os.Stdout
	c.Run()

}

func countLinks(depth int) (count int) {

	count = 0
	if depth > 0 {
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

func crawling_completed(spider *Spider) {

}

// Add a link into the system
func addLink(link Link) {

	// Check the URI isnt already assigned to another link
	existingLink := Link{}
	if existingLink.load(link.uri) {
		if existingLink.depth > link.depth {
			existingLink.depth = link.depth
			_ = existingLink.save()
		}
		return
	}

	_ = link.save()

}
