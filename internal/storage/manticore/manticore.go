package manticore

import (
	"context"
	"encoding/json"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"github.com/terratensor/kremlin-parser/internal/entities/entry"
	"log"
	"os"
)

var _ entry.StorageInterface = &Client{}

type DBEntry struct {
	Language  string `json:"language"`
	Title     string `json:"title"`
	Url       string `json:"url"`
	Updated   int64  `json:"updated"`
	Published int64  `json:"published"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
}

type Client struct {
	apiClient *openapiclient.APIClient
}

func New(tbl string) (*Client, error) {
	// Initialize ApiClient
	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	query := fmt.Sprintf(`show tables like '%v'`, tbl)

	// Проверяем существует ли таблица tbl, если нет, то создаем
	resp, _, _ := apiClient.UtilsAPI.Sql(context.Background()).Body(query).Execute()
	data := resp[0]["data"].([]interface{})

	if len(data) > 0 {
		myMap := data[0].(map[string]interface{})
		indexValue := myMap["Index"]

		if indexValue != tbl {
			err := createTable(apiClient, tbl)
			if err != nil {
				return nil, err
			}
		}
	} else {
		err := createTable(apiClient, tbl)
		if err != nil {
			return nil, err
		}
	}

	return &Client{apiClient: apiClient}, nil
}

func createTable(apiClient *openapiclient.APIClient, tbl string) error {
	query := fmt.Sprintf(`create table %v(language string, url string, title text, summary text, content text, published timestamp, updated timestamp) engine='columnar' morphology='stem_en,stem_ru,libstemmer_de,libstemmer_fr,libstemmer_es,libstemmer_pt'`, tbl)

	sqlreq := apiClient.UtilsAPI.Sql(context.Background()).Body(query)
	_, _, err := apiClient.UtilsAPI.SqlExecute(sqlreq)
	if err != nil {
		return err
	}

	//log.Println(resp)
	return nil
}

func (c *Client) Insert(ctx context.Context, entry *entry.Entry) error {

	dbe := &DBEntry{
		Language:  entry.Language,
		Title:     entry.Title,
		Url:       entry.Url,
		Updated:   entry.Updated.Unix(),
		Published: entry.Published.Unix(),
		Summary:   entry.Summary,
		Content:   entry.Content,
	}

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

	idr := openapiclient.InsertDocumentRequest{
		Index: "events",
		Doc:   doc,
	}

	_, r, err := c.apiClient.IndexAPI.Insert(ctx).InsertDocumentRequest(idr).Execute()

	//resp, r, err := apiClient.IndexAPI.Insert(context.Background()).InsertDocumentRequest(insertDocumentRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return fmt.Errorf("Error when calling `IndexAPI.Insert``: %v\n", err)
	}
	// response from `Insert`: SuccessResponse
	//fmt.Fprintf(os.Stdout, "Success Response from `IndexAPI.Insert`: %v\n", r)

	return nil
}

func (c *Client) Bulk(ctx context.Context, entries *[]entry.Entry) error {

	//entries map[string]interface{}
	log.Println(entries)
	buffer, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v\n", err)
	}
	//panic("stop")

	_, r, err := c.apiClient.IndexAPI.Bulk(ctx).Body(string(buffer)).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return fmt.Errorf("Error when calling `IndexAPI.Insert``: %v\n", err)
	}
	// response from `Insert`: SuccessResponse
	fmt.Fprintf(os.Stdout, "Success Response from `IndexAPI.Insert`: %v\n", r)

	return nil
}
