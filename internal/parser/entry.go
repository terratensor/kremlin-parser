package parser

import (
	"fmt"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"golang.org/x/net/html"
	"time"
)

func (p *Parser) parseEntries(n *html.Node) []feed.Entry {

	var entries []feed.Entry
	var f func(*html.Node)

	e := feed.Entry{}

	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "entry" {
			for cl := n.FirstChild; cl != nil; cl = cl.NextSibling {

				if cl.Type == html.ElementNode && cl.Data == "title" {
					e.Title = getInnerText(cl)
				}

				if cl.Type == html.ElementNode && cl.Data == "id" {
					e.Url = getInnerText(cl)
				}

				if cl.Type == html.ElementNode && cl.Data == "updated" {

					t, err := time.Parse("2006-01-02T15:04:05-07:00", getInnerText(cl))
					if err != nil {
						fmt.Println(err)
						return
					}
					e.Updated = &t
				}

				if cl.Type == html.ElementNode && cl.Data == "published" {
					t, err := time.Parse("2006-01-02T15:04:05-07:00", getInnerText(cl))
					if err != nil {
						fmt.Println(err)
						return
					}
					e.Published = &t
				}

				if cl.Type == html.ElementNode && cl.Data == "summary" {
					e.Summary = getInnerText(cl)
				}

				if cl.Type == html.ElementNode && cl.Data == "content" {
					e.Content = getInnerText(cl)
				}

			}

			e.Language = p.Lang
			e.ResourceID = p.ResourceID
			entries = append(entries, e) //fmt.Println(entry)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return entries
}
