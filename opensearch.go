package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"strings"

	"github.com/defensestation/osquery/v2"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type OSClient = *opensearch.Client
type OSAggregation = osquery.Aggregation
type OSResponse = *opensearchapi.SearchResp

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
	// Transform field names from camelCase to snake_case to match OpenSearch index
	transformedQuery := transformQueryFieldNames(query)

	queryString := map[string]any{
		"query_string": map[string]any{
			"query": transformedQuery,
		},
	}

	return osquery.CustomQuery(queryString)
}

// transformQueryFieldNames converts camelCase field names to snake_case
// to match the field names used in the OpenSearch index
func transformQueryFieldNames(query string) string {
	if query == "" {
		return query
	}

	result := query

	// Field name transformations (matching lupo implementation)
	result = strings.ReplaceAll(result, "publicationYear", "publication_year")
	result = strings.ReplaceAll(result, "relatedIdentifiers", "related_identifiers")
	result = strings.ReplaceAll(result, "relatedItems", "related_items")
	result = strings.ReplaceAll(result, "rightsList", "rights_list")
	result = strings.ReplaceAll(result, "fundingReferences", "funding_references")
	result = strings.ReplaceAll(result, "geoLocations", "geo_locations")
	// Note: version: includes colon to avoid transforming "version" in other contexts
	// as "version" is a reserved field in OpenSearch
	result = strings.ReplaceAll(result, "version:", "version_info:")
	result = strings.ReplaceAll(result, "landingPage", "landing_page")
	result = strings.ReplaceAll(result, "contentUrl", "content_url")
	result = strings.ReplaceAll(result, "citationCount", "citation_count")
	result = strings.ReplaceAll(result, "viewCount", "view_count")
	result = strings.ReplaceAll(result, "downloadCount", "download_count")
	result = strings.ReplaceAll(result, "schemaVersion", "schema_version")

	// Handle publisher nested fields
	// publisher.name -> publisher_obj.name
	// publisher.publisherIdentifier -> publisher_obj.publisherIdentifier
	// etc.
	publisherFields := []string{
		"publisher.name",
		"publisher.publisherIdentifier",
		"publisher.publisherIdentifierScheme",
		"publisher.schemeUri",
		"publisher.lang",
	}
	for _, field := range publisherFields {
		result = strings.ReplaceAll(result, field, strings.Replace(field, "publisher.", "publisher_obj.", 1))
	}

	// Escape forward slashes for OpenSearch query syntax
	// This is done last to ensure proper escaping of all forward slashes in the query
	result = strings.ReplaceAll(result, "/", "\\/")

	return result
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
