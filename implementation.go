package main

import (
	"golang.org/x/net/html"	
)

type Implementation interface {
	seed()
	prep()	
    processUri(uri string) string
    shouldFollowLink(link Link) bool
    parseHTML(node *html.Node)
}