package main

import (
	"context"
	"encoding/json"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"log"
	"os"
	"strconv"
	"time"
)

type DBEntry struct {
	Language   string `json:"language"`
	Title      string `json:"title"`
	Url        string `json:"url"`
	Updated    int64  `json:"updated"`
	Published  int64  `json:"published"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	Author     string `json:"author"`
	Number     string `json:"number"`
	ResourceID int    `json:"resource_id"`
}

type Response struct {
	Items []struct {
		Replace struct {
			Index   string `json:"_index"`
			Id      int    `json:"_id"`
			Created bool   `json:"created"`
			Result  string `json:"result"`
			Status  int    `json:"status"`
		} `json:"replace"`
	} `json:"items"`
	Errors bool `json:"errors"`
}

func NewDBEntry(entry *feed.Entry) *DBEntry {
	dbe := &DBEntry{
		Language:   entry.Language,
		Title:      entry.Title,
		Url:        entry.Url,
		Updated:    entry.Updated.Unix(),
		Published:  entry.Published.Unix(),
		Summary:    entry.Summary,
		Content:    entry.Content,
		Author:     entry.Author,
		Number:     entry.Number,
		ResourceID: entry.ResourceID,
	}

	return dbe
}

func main() {
	//body := "SHOW TABLES" // string | A query parameter string.
	//rawResponse := true   // bool | Optional parameter, defines a format of response. Can be set to `False` for Select only queries and set to `True` or omitted for any type of queries:  (optional) (default to true)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	searchRequest := *openapiclient.NewSearchRequest("events")

	// Пример для запроса фильтра по url
	filter := map[string]interface{}{"url": "http://kremlin.ru/events/president/news/73568"}
	query := map[string]interface{}{"equals": filter}

	//log.Printf("%v", query)

	searchRequest.SetQuery(query)
	resp, r, err := apiClient.SearchAPI.Search(context.Background()).SearchRequest(searchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UtilsAPI.Sql``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	defer r.Body.Close()

	log.Printf("%#v\n", r)
	//panic("stop")

	// показывает общее количество найденных результатов, предполагается тчо должно быть больше нуля
	total := resp.GetHits()
	tot := total.GetTotal()

	//log.Println(resp.GetHits().Hits)

	id, _ := getEntryID(resp)
	// response превращаем в объект entry
	dbe := makeDBEntry(resp)
	log.Printf("ID: %d\n", *id)
	//ent.ID = id
	//log.Println(ent)
	updated := time.Unix(dbe.Updated, 0)
	published := time.Unix(dbe.Published, 0)

	ent := &feed.Entry{
		ID:         id,
		Language:   dbe.Language,
		Title:      dbe.Title,
		Url:        dbe.Url,
		Updated:    &updated,
		Published:  &published,
		Summary:    dbe.Summary,
		Content:    dbe.Content,
		Author:     dbe.Author,
		Number:     dbe.Number,
		ResourceID: 1,
	}

	ent.Summary = "test1"
	Update(ent)

	fmt.Fprintf(os.Stdout, "resp.Hits: %v\n", tot)
	// response from `Sql`: []map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `UtilsAPI.Sql`: %v\n", resp)
}

func getEntryID(resp *openapiclient.SearchResponse) (*int64, error) {
	var hits []map[string]interface{}
	var _id interface{}

	hits = resp.Hits.Hits

	// Если слайс Hits пустой (0) значит нет совпадений
	if len(hits) == 0 {
		return nil, nil
	}

	hit := hits[0]

	_id = hit["_id"]
	id, err := strconv.ParseInt(_id.(string), 10, 64)
	//log.Printf("id %d\n", id)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse ID to int64: %v\n", resp)
	}

	return &id, nil
}

func makeDBEntry(resp *openapiclient.SearchResponse) *DBEntry {
	var hits []map[string]interface{}
	hits = resp.Hits.Hits

	// Если слайс Hits пустой (0) значит нет совпадений
	if len(hits) == 0 {
		return nil
	}

	hit := hits[0]

	sr := hit["_source"]
	jsonData, err := json.Marshal(sr)

	var dbe DBEntry
	err = json.Unmarshal(jsonData, &dbe)
	if err != nil {
		log.Fatal(err)
	}

	return &dbe
}

func Update(entry *feed.Entry) error {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	//log.Printf("Entry update: %#v:\n", entry)

	dbe := NewDBEntry(entry)

	//marshal into JSON buffer
	buffer, err := json.Marshal(dbe)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v\n", err)
	}

	var doc map[string]interface{}
	err = json.Unmarshal(buffer, &doc)
	if err != nil {
		return fmt.Errorf("error unmarshaling buffer: %v\n", err)
	}

	//updateDocumentRequest := *openapiclient.NewUpdateDocumentRequest(
	//	"events",
	//	doc,
	//) // UpdateDocumentRequest |
	idr := openapiclient.InsertDocumentRequest{
		Index: "events",
		Id:    entry.ID,
		Doc:   doc,
	}

	//log.Println(udr)
	_, r, err := apiClient.IndexAPI.Replace(context.Background()).InsertDocumentRequest(idr).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return fmt.Errorf("Error when calling `IndexAPI.Update``: %v\n", err)
	}

	log.Printf("%#v", r)
	//
	//defer r.Body.Close()
	//body, err := io.ReadAll(r.Body)
	//// snippet only
	//var result Response
	//if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
	//	fmt.Println("Can not unmarshal JSON")
	//}
	//fmt.Println(PrettyPrint(result))

	return nil
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
