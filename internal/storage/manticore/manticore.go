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

type Entry struct {
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

func (c *Client) Insert(ctx context.Context, entry *entry.Entry) {
	//log.Println(entry)

	//configuration := openapiclient.NewConfiguration()
	//apiClient := openapiclient.NewAPIClient(configuration)

	//marshal into JSON buffer
	buffer, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("error marshaling JSON: %v\n", err)
	}

	//log.Println(string(buffer))

	var doc map[string]interface{}
	err = json.Unmarshal(buffer, &doc)
	if err != nil {
		// Handle error
	}

	idr := openapiclient.InsertDocumentRequest{
		Index: "events",
		Doc:   doc,
	}

	_, r, err := c.apiClient.IndexAPI.Insert(ctx).InsertDocumentRequest(idr).Execute()

	//resp, r, err := apiClient.IndexAPI.Insert(context.Background()).InsertDocumentRequest(insertDocumentRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IndexAPI.Insert``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Insert`: SuccessResponse
	fmt.Fprintf(os.Stdout, "Response from `IndexAPI.Insert1`: %v\n", r)

}

func (c *Client) Bulk(ctx context.Context, entries *[]entry.Entry) {

	//entries map[string]interface{}
	log.Println(entries)
	buffer, err := json.Marshal(entries)
	if err != nil {
		fmt.Printf("error marshaling JSON: %v\n", err)
	}
	//panic("stop")

	_, r, err := c.apiClient.IndexAPI.Bulk(ctx).Body(string(buffer)).Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `IndexAPI.Insert``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Insert`: SuccessResponse
	fmt.Fprintf(os.Stdout, "Response from `IndexAPI.Insert1`: %v\n", r)
}
