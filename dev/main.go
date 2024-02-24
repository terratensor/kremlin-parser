package main

import (
	"context"
	"fmt"
	openapiclient "github.com/manticoresoftware/manticoresearch-go"
	"os"
)

func main() {
	//body := "SHOW TABLES" // string | A query parameter string.
	//rawResponse := true   // bool | Optional parameter, defines a format of response. Can be set to `False` for Select only queries and set to `True` or omitted for any type of queries:  (optional) (default to true)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)

	searchRequest := *openapiclient.NewSearchRequest("events")

	// Пример для запроса фильтра по url
	filter := map[string]interface{}{"url": "http://kremlin.ru/events/president/news/73519"}
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

	fmt.Fprintf(os.Stdout, "resp.Hits: %v\n", tot)
	// response from `Sql`: []map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `UtilsAPI.Sql`: %v\n", resp)
	//fmt.Fprintf(os.Stdout, "Response from `UtilsAPI.Sql`: %v\n", r)
}
