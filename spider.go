package main

import (
	"bytes"
	"github.com/jackdanger/collectlinks"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func (spider *Spider) crawl() []Link {

	toStore := []Link{}
	resp, err := http.Get(spider.link.uri)

	if err != nil {
		spider.link.failures++
		spider.link.save()
	} else {

		defer resp.Body.Close()

		// Read the content
		var bodyBytes []byte

		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		for n, _ := range implementations {
			implementations[n].parseRaw(spider.link, bodyBytes)
		}

		// Restore the io.ReadCloser to its original state
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		//@TODO - detect the type of file from the headers, don't try to parse links on binary etc

		// Find all the links
		links := collectlinks.All(resp.Body)
		for _, link := range links {
			absolute := fixUrl(link, spider.link.uri)
			if absolute != "" {

				for n, _ := range implementations {
					absolute = implementations[n].processUri(absolute)
				}

				childLink := *NewLink(absolute)
				childLink.depth = spider.link.depth + 1
				childLink.parent = spider.link.id

				shouldFollow := false
				for n, _ := range implementations {
					if implementations[n].shouldFollowLink(childLink) {
						shouldFollow = true
					}
				}

				if shouldFollow {
					toStore = append(toStore, childLink)
					//addLink(childLink) //@TODO - store these somewhere in the local thread then sync at the end
				}

			}
		}

		// Restore the io.ReadCloser to its original state
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the HTML itself
		doc, _ := html.Parse(resp.Body)
		for n, _ := range implementations {
			implementations[n].parseHTML(spider.link, doc)
		}

	}

	crawling_completed(spider)
	return toStore

}

func fixUrl(href, base string) string {
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
