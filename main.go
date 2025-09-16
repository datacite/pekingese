package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

var (
	OSAddress = os.Getenv("OPENSEARCH_ADDRESS")
	IndexName = os.Getenv("OPENSEARCH_INDEX")
	osClient  OSClient // FOR INITIAL DEV ONLY - DO NOT USE GLOBAL IN PRODUCTION (probably)
)

func main() {
	osClient = InitOpenSearch()

	http.HandleFunc("/", getData)
	http.ListenAndServe(":8080", nil)
}

func getData(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("query")
	clientId := request.URL.Query().Get("client_id")
	providerId := request.URL.Query().Get("provider_id")

	fieldsPresent := strings.Split(request.URL.Query().Get("present"), ",")
	fieldsDistribution := strings.Split(request.URL.Query().Get("distribution"), ",")

	presentAggs := make([]OSAggregation, 0)
	for _, field := range fieldsPresent {
		if field != "" {
			presentAggs = append(presentAggs, buildPresentAggregation(field))
		}
	}

	distributionAggs := make([]OSAggregation, 0)
	for _, field := range fieldsDistribution {
		if field != "" {
			distributionAggs = append(distributionAggs, buildDistributionAggregation(field, 10))
		}
	}

	search := BuildBaseQuery(clientId, providerId, query)
	search = search.
		Aggs(presentAggs...).
		Aggs(distributionAggs...)

	results := Run(search)

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(results)
}

