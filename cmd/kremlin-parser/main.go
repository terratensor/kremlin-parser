package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

type Meta struct {
	Updated *time.Time `json:"updated"`
	ID      string     `json:"id"`
	Self    string     `json:"self"`
	Prev    string     `json:"prev"`
	First   string     `json:"first"`
	Next    string     `json:"next"`
	Last    string     `json:"last"`
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

var pageCount, outputPath string

func main() {
	flag.StringVarP(&outputPath, "output", "o", ".", "путь сохранения файлов")
	flag.StringVarP(&pageCount, "page-count", "p", "1", "спарсить указанное количество страниц")
	flag.Parse()

	startUrl := "http://kremlin.ru/events/all/feed"

	var entries Entries
	file := fmt.Sprintf("%v/kremlin.json", outputPath)
	file = filepath.Clean(file)

	var meta Meta
	var url string

	count := 1
	pages := getPages(pageCount)

	for {
		if meta.Self == "" && meta.Next == "" {
			url = startUrl
		} else {
			url = meta.Next
		}

		if count != pages || count != 1 {
			log.Printf("waiting 5 seconds")
			time.Sleep(5 * time.Second)
		}

		log.Printf("parsing %v", url)

		node, err := getTopicBody(url)
		meta = getMeta(node)

		if os.IsTimeout(err) {
			log.Println("IsTimeoutError: true")
			log.Printf("Waiting 5 seconds")
			time.Sleep(5 * time.Second)
			//log.Printf("Attempt: %d\n", i+1)
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		entries = parseEntries(entries, node)

		if count == pages {
			break
		}
		count++
	}

	log.Printf("length of entries %v", len(entries))

	writeJsonFile(entries, file)
	log.Printf("file %v was successful writing\n", file)

}

func getPages(count string) int {
	pages, err := strconv.Atoi(count)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if pages <= 0 {
		pages = 1
	}
	return pages
}

func getTopicBody(url string) (*html.Node, error) {

	resp, err := call(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("status code error: %d %s\r\n", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := html.Parse(resp.Body)
	//body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err) // Handle error
	}
	return doc, nil
}

// call is a Go function that makes a GET request to the provided URL and returns the response and an error, if any.
//
// It takes a string 'url' as a parameter and returns a pointer to http.Response and an error.
func call(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Fatal(err)
	}

	// Без user-agent kremlin.ru не отдает данные
	req.Header.Add("User-Agent", "TestProgram/0.01")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	return resp, err
}

func getMeta(node *html.Node) Meta {
	var f func(*html.Node)

	meta := Meta{}

	f = func(node *html.Node) {

		// Если у ноды есть родитель и этот родитель — тэг feed
		if node.Parent != nil && node.Parent.Data == "feed" {

			if node.Type == html.ElementNode && node.Data == "updated" {
				t, err := time.Parse("2006-01-02T15:04:05-07:00", getInnerText(node))
				if err != nil {
					fmt.Println(err)
					return
				}
				meta.Updated = &t
			}
			if node.Type == html.ElementNode && node.Data == "id" {
				meta.ID = getInnerText(node)
			}
			//
			if node.Type == html.ElementNode && node.Data == "link" {

				if nodeHasRequiredRelAttr("self", node) {
					meta.Self = getRequiredDataAttr("href", node)
				}
				if nodeHasRequiredRelAttr("prev", node) {
					meta.Prev = getRequiredDataAttr("href", node)
				}
				if nodeHasRequiredRelAttr("first", node) {
					meta.First = getRequiredDataAttr("href", node)
				}
				if nodeHasRequiredRelAttr("next", node) {
					meta.Next = getRequiredDataAttr("href", node)
				}
				if nodeHasRequiredRelAttr("last", node) {
					meta.Last = getRequiredDataAttr("href", node)
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)

	return meta
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

// getInnerText returns the inner text of the HTML node.
//
// It takes a pointer to a html.Node as a parameter and returns a string.
func getInnerText(node *html.Node) string {
	for el := node.FirstChild; el != nil; el = el.NextSibling {
		if el.Type == html.TextNode {
			return el.Data
		}
	}
	return ""
}

// nodeHasRequiredRelAttr checks if the given html.Node has the required rel attribute.
//
// It takes a string rcc and a *html.Node n as parameters and returns a boolean.
func nodeHasRequiredRelAttr(rcc string, n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "rel" && attr.Val == rcc {
			return true
		}
	}
	return false
}

// getRequiredDataAttr returns the value of the specified attribute from the given html.Node.
//
// rda string - the attribute key to search for.
// n *html.Node - the html node to search within.
// string - the value of the specified attribute, or an empty string if not found.
func getRequiredDataAttr(rda string, n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == rda {
			return attr.Val
		}
	}
	return ""
}

func writeJsonFile(entries Entries, outputPath string) {

	// Create file
	file, err := os.Create(outputPath)
	checkError("Cannot create file", err)
	defer file.Close()

	aJson, err := json.MarshalIndent(entries, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(aJson)
	checkError("Cannot write to the file", err)
}
