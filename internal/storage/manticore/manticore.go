package manticore

import (
	"context"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"log"
)

func New(tbl string) {
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
			createTable(apiClient, tbl)
		}
	} else {
		createTable(apiClient, tbl)
	}
}

func createTable(apiClient *openapiclient.APIClient, tbl string) {
	query := fmt.Sprintf(`create table %v(language string, url string, title text, summary text, content text, created timestamp, updated timestamp) engine='columnar' morphology='stem_en,stem_ru,libstemmer_de,libstemmer_fr,libstemmer_es,libstemmer_pt'`, tbl)

	sqlreq := apiClient.UtilsAPI.Sql(context.Background()).Body(query)
	resp, _, _ := apiClient.UtilsAPI.SqlExecute(sqlreq)

	log.Println(resp)
}
