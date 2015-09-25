## Synopsis

Spod is a simple but expandable webcrawler written in Go

## Code Example

  go run *.go "http://example.com/"

## Motivation

My primary reason for this project is to **learn** golang, and so the quality of the code will initially be very poor.

## Reference

Bespoke functionality and behavior is abstracted out from the core crawling functionality by extending the Implementation protocol. 

### seed()

Called upon startup to add new links into the database, you can also seed by passing an URL as an argument run running the app

### prep()	

Called after seeding, allows you to do a one time inization of any local scope variables used

### processUri(uri string) string

Allows you to clean new Uri before they get added in to the index, e.g. remove anchor links.

### shouldFollowLink(link Link) bool

Allows you to decide if a link should get added to the queue or not

### parseHTML(node *html.Node)

Gives you the raw HTML For each URI crawled

## License

Use it, hack it, fork it - just include a link back to this repo
