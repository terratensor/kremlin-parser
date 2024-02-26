package main

import (
	"context"
	"encoding/json"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/kremlin-parser/internal/entities/entry"
	"log"
	"os"
	"strconv"
	"time"
)

type DBEntry struct {
	Language  string `json:"language"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	Updated   int64  `json:"updated"`
	Published int64  `json:"published"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
}

func main() {
	//body := "SHOW TABLES" // string | A query parameter string.
	//rawResponse := true   // bool | Optional parameter, defines a format of response. Can be set to `False` for Select only queries and set to `True` or omitted for any type of queries:  (optional) (default to true)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	searchRequest := *openapiclient.NewSearchRequest("events")

	// Пример для запроса фильтра по url
	filter := map[string]interface{}{"url": "http://kremlin.ru/events/president/news/73538"}
	query := map[string]interface{}{"equals": filter}

	//log.Printf("%v", query)

	searchRequest.SetQuery(query)
	resp, r, err := apiClient.SearchAPI.Search(context.Background()).SearchRequest(searchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UtilsAPI.Sql``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	// показывает общее количество найденных результатов, предполагается тчо должно быть больше нуля
	total := resp.GetHits()
	tot := total.GetTotal()

	//log.Println(resp.GetHits().Hits)

	id, _ := getEntryID(resp)
	// response превращаем в объект entry
	ent := makeDBEntry(resp)
	log.Printf("ID: %d\n", *id)
	//ent.ID = id
	//log.Println(ent)

	ent.Summary = "test1"
	Update(id, ent)

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

func makeDBEntry(resp *openapiclient.SearchResponse) *entry.Entry {
	var hits []map[string]interface{}
	hits = resp.Hits.Hits

	hit := hits[0]

	sr := hit["_source"]
	jsonData, err := json.Marshal(sr)

	var dbe DBEntry
	err = json.Unmarshal(jsonData, &dbe)
	if err != nil {
		log.Println("json.Unmarshal(jsonData, &ent)")
		log.Fatal(err)
	}

	updated := time.Unix(dbe.Updated, 0)
	published := time.Unix(dbe.Published, 0)

	ent := entry.Entry{
		Language:  dbe.Language,
		Title:     dbe.Title,
		Url:       dbe.Url,
		Updated:   &updated,
		Published: &published,
		Summary:   dbe.Summary,
		Content:   dbe.Content,
	}

	//fmt.Printf("Title: %s, Summary: %v\n", entry.Title, entry.Summary)
	//log.Println(string(jsonData))

	return &ent
}

func Update(id *int64, entry *entry.Entry) error {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	log.Printf("Entry update: %#v:\n", entry)

	//marshal into JSON buffer
	buffer, err := json.Marshal(entry)
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
		Id:    id,
		Doc:   doc,
	}

	//log.Println(udr)
	_, r, err := apiClient.IndexAPI.Replace(context.Background()).InsertDocumentRequest(idr).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return fmt.Errorf("Error when calling `IndexAPI.Update``: %v\n", err)
	}

	log.Println(r)

	return nil
}
