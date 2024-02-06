package parser

import (
	"fmt"
	"golang.org/x/net/html"
	"time"
)

type Entries []Entry

type Entry struct {
	Title     string     `json:"title"`
	Url       string     `json:"url"`
	Updated   *time.Time `json:"update"`
	Published *time.Time `json:"published"`
	Content   string     `json:"content"`
}

func parseEntries(entries Entries, n *html.Node) Entries {

	var f func(*html.Node)

	entry := Entry{}

	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "entry" {
			for cl := n.FirstChild; cl != nil; cl = cl.NextSibling {

				if cl.Type == html.ElementNode && cl.Data == "title" {
					entry.Title = getInnerText(cl)
				}

				if cl.Type == html.ElementNode && cl.Data == "id" {
					entry.Url = getInnerText(cl)
				}

				if cl.Type == html.ElementNode && cl.Data == "updated" {

					t, err := time.Parse("2006-01-02T15:04:05-07:00", getInnerText(cl))
					if err != nil {
						fmt.Println(err)
						return
					}
					entry.Updated = &t
				}

				if cl.Type == html.ElementNode && cl.Data == "published" {
					t, err := time.Parse("2006-01-02T15:04:05-07:00", getInnerText(cl))
					if err != nil {
						fmt.Println(err)
						return
					}
					entry.Published = &t
				}

				if cl.Type == html.ElementNode && cl.Data == "content" {
					entry.Content = getInnerText(cl)
				}

			}

			entries = append(entries, entry) //fmt.Println(entry)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return entries
}
