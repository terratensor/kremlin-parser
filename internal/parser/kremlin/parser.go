package kremlin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"golang.org/x/net/html"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Parser struct {
	ID             uuid.UUID
	ManticoreIndex string
	SaveToFile     bool
	ResourceID     int
	Lang           string
	URI            string
	PageCount      int
	OutputPath     string
	Delay          *time.Duration
	Meta           *Meta
	entries        *feed.Entries
}

func NewParser(uri config.StartURL, cfg *config.Config, entries *feed.Entries) Parser {
	parser := Parser{
		ID:             uuid.New(),
		ManticoreIndex: cfg.ManticoreIndex,
		SaveToFile:     cfg.SaveToFile,
		ResourceID:     cfg.Parser.ResourceID,
		Lang:           uri.Lang,
		URI:            uri.Url,
		PageCount:      cfg.Parser.PageCount,
		OutputPath:     cfg.Parser.OutputPath,
		Delay:          cfg.ParseDelay,
		Meta:           NewMeta(),
		entries:        entries,
	}
	return parser
}

// Parse парсит указанное количество страниц rss ленты сайта кремля.
// Сохраняет каждую страницу в отдельный json файл.
// При каждом успешном парсинге возвращает ссылку на следующую страницу rss ленты.
// Делает установленную в конфиге паузу между парсингами (5 сек по умолчанию).
// Используется logger для записи различных событий во время анализа.
func (p *Parser) Parse(ctx context.Context, log *slog.Logger) {
	const op = "parser.parse"
	log = log.With(
		slog.String("op", op),
		slog.String("pid", p.ID.String()),
		slog.String("lang", p.Lang),
	)

	count := 1
	// Парсит указанное количество страниц rss ленты сайта кремля.
	// Сохраняет каждую страницу в отдельный файл.
	// При каждом успешном парсинге возвращает ссылку на следующую страницу rss ленты.
	// Делает паузу 5 секунд между парсингами.
	for {

		url := p.getUrl()

		// Если url пустой, следующей достигнут конец RSS ленты,
		// следующей станицы не существует, заканчиваем парсинг
		if url == "" {
			break
		}

		path := p.NewFilepath(url)

		if count != p.PageCount || count != 1 {
			log.Info("waiting", slog.String("parse_delay", p.Delay.String()))
			time.Sleep(*p.Delay)
		}

		log.Debug("parsing url", slog.Any("url", url))

		node, err := getTopicBody(url)

		if os.IsTimeout(err) {
			log.Error("server timeout error", sl.Err(err))
			continue
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			continue
		}

		p.parseMeta(node)
		entries := p.parseEntries(node)

		// Итерируемся по слайсу спарсеных entries, ищем по url запись в мантикоре,
		// если записи нет nil, то делаем запись в мантикору
		// todo сделать проверку и логику для update, когда запись есть но поля updated не совпадают
		for _, e := range entries {
			dbe, err := p.entries.Storage.FindByUrl(ctx, e.Url)
			if err != nil {
				log.Error("failed find entry by url", sl.Err(err))
			}
			if dbe == nil {
				id, err := p.entries.Storage.Insert(ctx, &e)
				if err != nil {
					log.Error(
						"failed insert entry",
						slog.Int64("id", *id),
						slog.String("url", e.Url),
						sl.Err(err),
					)
				}
				log.Info(
					"entry successful inserted",
					slog.Int64("id", *id),
					slog.String("url", e.Url),
				)
			} else {
				if !matchTimes(dbe, e) {
					e.ID = dbe.ID
					err = p.entries.Storage.Update(ctx, &e)
					if err != nil {
						log.Error(
							"failed update entry",
							slog.Int64("id", *e.ID),
							slog.String("url", e.Url),
							sl.Err(err),
						)
					} else {
						log.Info(
							"entry successful updated",
							slog.Int64("id", *e.ID),
							slog.String("url", e.Url),
						)
					}
				}
			}
		}

		if p.SaveToFile {
			WriteJsonFile(log, entries, path)
		}

		if count == p.PageCount {
			break
		}
		count++
	}
}

func (p *Parser) getUrl() string {
	var url string
	// Если Meta только инициализирован, то Meta.Self и Meta.Next пусты,
	// устанавливаем то url равен начальному url,
	// иначе url равен ссылке на следующую страницу
	if p.Meta.Self == "" && p.Meta.Next == "" {
		url = p.URI
	} else {
		url = p.Meta.Next
	}
	return url
}

func (p *Parser) NewFilepath(url string) string {
	file := fmt.Sprintf("%v/%v.json", p.OutputPath, slug.Make(url))
	file = filepath.Clean(file)
	return file
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
		return nil, err
	}

	// Без user-agent kremlin.ru не отдает данные
	req.Header.Add("User-Agent", "TestProgram/0.01")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, err
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

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func WriteJsonFile(logger *slog.Logger, entries []feed.Entry, outputPath string) {
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
	logger.Debug("path was successful writing", slog.Any("path", outputPath))
}

func matchTimes(dbe *feed.Entry, e feed.Entry) bool {
	// Приводим время в обоих объектах к GMT+4, как на сайте Кремля
	loc, _ := time.LoadLocation("Etc/GMT-4")
	dbeTime := dbe.Updated.In(loc)
	eTime := e.Updated.In(loc)

	if dbeTime != eTime {
		log.Printf("`updated` fields do not match dbe updated %v", dbeTime)
		log.Printf("`updated` fields do not match prs updated %v", eTime)
		return false
	}
	return true
}
