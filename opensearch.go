package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/defensestation/osquery/v2"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type OSClient = *opensearch.Client
type OSAggregation = osquery.Aggregation
type OSResponse = *opensearchapi.SearchResp

var camelCaseToSnakeCaseFields = []struct {
	camelCase string
	snakeCase string
}{
	{"publicationYear", "publication_year"},
	{"relatedIdentifiers", "related_identifiers"},
	{"relatedItems", "related_items"},
	{"rightsList", "rights_list"},
	{"fundingReferences", "funding_references"},
	{"geoLocations", "geo_locations"},
	{"version:", "version_info:"},
	{"landingPage", "landing_page"},
	{"contentUrl", "content_url"},
	{"citationCount", "citation_count"},
	{"viewCount", "view_count"},
	{"downloadCount", "download_count"},
	{"schemaVersion", "schema_version"},
}

var camelCaseToSnakeCaseFieldsRegex = []struct {
	regex     string
	snakeCase string
}{
	{`(publisher\.)(name|publisherIdentifierScheme|publisherIdentifier|schemeUri|lang)`, "publisher_obj.$2"},
}

func InitOpenSearch() OSClient {
	osClient, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{OSAddress},
	})
	if err != nil {
		log.Fatalf("Error initializing OpenSearch client: %s", err)
	}

	return osClient
}

func BuildBaseQuery(clientId string, providerId string, consortiumId string, query string) *osquery.SearchRequest {
	// set up base filters
	filters := []osquery.Mappable{
		osquery.Term("agency", "datacite"),
		osquery.Term("aasm_state", "findable"),
	}

	// apply conditional filters
	if clientId != "" {
		filters = append(filters, osquery.Term("client.id", clientId))
	}

	if providerId != "" {
		filters = append(filters, osquery.Term("provider.id", providerId))
	}

	if consortiumId != "" {
		filters = append(filters, osquery.Term("consortium_id", consortiumId))
	}

	if query != "" {
		filters = append(filters, buildQueryString(query))
	}

	return osquery.Search().Size(0).Query(
		osquery.Bool().Filter(filters...),
	)
}

func buildQueryString(query string) *osquery.CustomQueryMap {
	// support camel case for certain fields for consistency with lupo
	for _, field := range camelCaseToSnakeCaseFields {
		query = strings.ReplaceAll(query, field.camelCase, field.snakeCase)
	}

	for _, field := range camelCaseToSnakeCaseFieldsRegex {
		query = regexp.MustCompile(field.regex).ReplaceAllString(query, field.snakeCase)
	}

	query = strings.ReplaceAll(query, "/", "\\/")

	queryString := map[string]any{
		"query_string": map[string]any{
			"query": query,
		},
	}

	return osquery.CustomQuery(queryString)
}

func buildPresentAggregation(field string) OSAggregation {
	presentAgg := map[string]any{
		"filters": map[string]any{
			"filters": map[string]any{
				"present": map[string]any{
					"exists": map[string]any{
						"field": field,
					},
				},
				"absent": map[string]any{
					"bool": map[string]any{
						"must_not": map[string]any{
							"exists": map[string]any{
								"field": field,
							},
						},
					},
				},
			},
		},
	}

	return osquery.CustomAgg("present_"+field, presentAgg)
}

func buildDistributionAggregation(field string, size uint64) OSAggregation {
	return osquery.TermsAgg("distribution_"+field, field).Size(size)
}

func Run(query *osquery.SearchRequest) *opensearchapi.SearchResp {
	searchResponse, err := query.Run(
		context.TODO(),
		osClient,
		&osquery.Options{
			Indices: []string{IndexName},
		},
	)
	if err != nil {
		log.Fatalf("Failed searching for stuff: %s", err)
	}

	return searchResponse
}
