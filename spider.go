package main

import (
	"github.com/jackdanger/collectlinks"
	"log"
	"net/http"
	"net/url"
	"io/ioutil"
	"bytes"
	"golang.org/x/net/html"
)

type Spider struct {
	link Link
	name string
	busy bool
}

func NewSpider(name string) *Spider {
	s := new(Spider)
	s.name = name
	return s
}

func (spider *Spider) crawl() {
	
	resp, err := http.Get(spider.link.uri)
	if err != nil {
		spider.link.failures++;
		spider.link.save()
	} else {

		defer resp.Body.Close()

		// Read the content
		var bodyBytes []byte		

		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
				log.Fatal(err)
		}
	
		// Save the data
		spider.link.saveData(bodyBytes)

		// Restore the io.ReadCloser to its original state
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Find all the links
		links := collectlinks.All(resp.Body)
		for _, link := range(links) {
			absolute := fixUrl(link, spider.link.uri)
			if absolute != "" {

				for n, _ := range(implementations) {
					absolute = implementations[n].processUri(absolute)
				}				

				childLink := *NewLink(absolute)
				childLink.depth = spider.link.depth + 1
				childLink.parent = spider.link.id

				shouldFollow := false
				for n, _ := range(implementations) {
					if (implementations[n].shouldFollowLink(childLink)) {
						shouldFollow = true
					}
				}

				if (shouldFollow) {
					addLink(childLink) // @todo - store these somewhere and store on the main thread				
				}

			}
		}

		// Restore the io.ReadCloser to its original state
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the HTML itself
		doc, _ := html.Parse(resp.Body)  	
		for n, _ := range(implementations) {
			implementations[n].parseHTML(doc)
		}

	}

	crawling_completed(spider)

}

func fixUrl(href, base string) (string) {
  uri, err := url.Parse(href)
  if err != nil {
	return ""
  }
  baseUrl, err := url.Parse(base)
  if err != nil {
	return ""
  }
  uri = baseUrl.ResolveReference(uri)
  return uri.String()
}