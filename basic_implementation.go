package main

import (	
	"golang.org/x/net/html"		
	"regexp"	
)

type BasicImplementation struct {

}

func (i BasicImplementation) processUri(uri string) string {

	// Strip out anchors from the URL
	reg, _ := regexp.Compile("#(.+)")
	uri = reg.ReplaceAllString(uri, "")
	return uri;

}

func (i BasicImplementation) seed()  {

	/*
	link := *NewLink("http://www.google.com/")
	addLink(link)
	*/

}


func (i BasicImplementation) prep()  {

}


func (i BasicImplementation) shouldFollowLink(link Link) bool  {
	return true
}

func (i BasicImplementation) parseHTML(node *html.Node) {

}

