package main

import (	
	"golang.org/x/net/html"		
	"regexp"	
)

type BasicImplementation struct {

}


// Modify a new URI for clean up and other logic
func (i BasicImplementation) processUri(uri string) string {

	// Strip out anchors from the URL
	reg, _ := regexp.Compile("#(.+)")
	uri = reg.ReplaceAllString(uri, "")
	return uri;

}

// One time initiation code for this implementation
func (i BasicImplementation) prep() {}

// Seed the starting links in the system
func (i BasicImplementation) seed() {}

// Parse raw bytecode for a link
func (i BasicImplementation) parseRaw(link Link, data []byte) {}

// Parse HTML text for a link
func (i BasicImplementation) parseHTML(link Link, node *html.Node) {}

// Logic to decide if a link should be followed or not
func (i BasicImplementation) shouldFollowLink(link Link) bool  {
	return true
}