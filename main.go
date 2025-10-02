package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	APIPort   = os.Getenv("API_PORT")
	OSAddress = os.Getenv("OPENSEARCH_HOST")
	IndexName = os.Getenv("OPENSEARCH_INDEX")
	osClient  OSClient // FOR INITIAL DEV ONLY - DO NOT USE GLOBAL IN PRODUCTION (probably)
)

func main() {
	log.Println("Initializing OpenSearch client...")
	osClient = InitOpenSearch()
	log.Println("OpenSearch client initialized successfully")

	log.Println("Starting API server...")
	http.HandleFunc("/", getData)
	err := http.ListenAndServe(":"+APIPort, nil)
	log.Fatalf("Error starting server: %s", err)
}

func getData(response http.ResponseWriter, request *http.Request) {
	// Parse query parameters
	query := request.URL.Query().Get("query")
	clientId := request.URL.Query().Get("client_id")
	providerId := request.URL.Query().Get("provider_id")
	numDistributionResults := GetURLQueryAsUInt(request, "distribution_size", 10)

	fieldsPresent := strings.Split(request.URL.Query().Get("present"), ",")
	fieldsDistribution := strings.Split(request.URL.Query().Get("distribution"), ",")

	// Build aggregation slices and query
	presentAggs := make([]OSAggregation, 0)
	for _, field := range fieldsPresent {
		if field != "" {
			presentAggs = append(presentAggs, buildPresentAggregation(field))
		}
	}

	distributionAggs := make([]OSAggregation, 0)
	for _, field := range fieldsDistribution {
		if field != "" {
			distributionAggs = append(distributionAggs, buildDistributionAggregation(field, numDistributionResults))
		}
	}

	search := BuildBaseQuery(clientId, providerId, query)
	search = search.
		Aggs(presentAggs...).
		Aggs(distributionAggs...)

	// Execute search and returned parsed openSearchResponse
	openSearchResponse := Run(search)
	apiResponse, err := ParseSearchResp(openSearchResponse)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(apiResponse)
}

func GetURLQueryAsUInt(request *http.Request, param string, defaultValue uint64) uint64 {
	valueStr := request.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
