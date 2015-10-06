package main

import (
	"golang.org/x/net/html"	
)

type Implementation interface {
	seed()
	prep()	
	processUri(uri string) string
	shouldFollowLink(link Link) bool
	parseHTML(link Link, node *html.Node)
	parseRaw(link Link, data []byte)
}