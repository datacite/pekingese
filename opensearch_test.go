package main

import (
	"testing"
)

func TestTransformQueryFieldNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty query",
			input:    "",
			expected: "",
		},
		{
			name:     "fundingReferences wildcard query",
			input:    "fundingReferences:*",
			expected: "funding_references:*",
		},
		{
			name:     "publicationYear query",
			input:    "publicationYear:2024",
			expected: "publication_year:2024",
		},
		{
			name:     "relatedIdentifiers query",
			input:    "relatedIdentifiers:*",
			expected: "related_identifiers:*",
		},
		{
			name:     "relatedItems query",
			input:    "relatedItems:*",
			expected: "related_items:*",
		},
		{
			name:     "rightsList query",
			input:    "rightsList:*",
			expected: "rights_list:*",
		},
		{
			name:     "geoLocations query",
			input:    "geoLocations:*",
			expected: "geo_locations:*",
		},
		{
			name:     "version field",
			input:    "version:1.0",
			expected: "version_info:1.0",
		},
		{
			name:     "landingPage query",
			input:    "landingPage:*",
			expected: "landing_page:*",
		},
		{
			name:     "contentUrl query",
			input:    "contentUrl:*",
			expected: "content_url:*",
		},
		{
			name:     "citationCount query",
			input:    "citationCount:[1 TO *]",
			expected: "citation_count:[1 TO *]",
		},
		{
			name:     "viewCount query",
			input:    "viewCount:[100 TO *]",
			expected: "view_count:[100 TO *]",
		},
		{
			name:     "downloadCount query",
			input:    "downloadCount:[50 TO *]",
			expected: "download_count:[50 TO *]",
		},
		{
			name:     "schemaVersion query",
			input:    "schemaVersion:4.0",
			expected: "schema_version:4.0",
		},
		{
			name:     "publisher.name query",
			input:    "publisher.name:\"Test Publisher\"",
			expected: "publisher_obj.name:\"Test Publisher\"",
		},
		{
			name:     "publisher.publisherIdentifier query",
			input:    "publisher.publisherIdentifier:*",
			expected: "publisher_obj.publisherIdentifier:*",
		},
		{
			name:     "publisher.publisherIdentifierScheme query",
			input:    "publisher.publisherIdentifierScheme:ROR",
			expected: "publisher_obj.publisherIdentifierScheme:ROR",
		},
		{
			name:     "publisher.schemeUri query",
			input:    "publisher.schemeUri:*",
			expected: "publisher_obj.schemeUri:*",
		},
		{
			name:     "publisher.lang query",
			input:    "publisher.lang:en",
			expected: "publisher_obj.lang:en",
		},
		{
			name:     "forward slash escaping",
			input:    "doi:10.1234/test",
			expected: "doi:10.1234\\/test",
		},
		{
			name:     "complex query with multiple fields",
			input:    "fundingReferences:* AND publicationYear:2024",
			expected: "funding_references:* AND publication_year:2024",
		},
		{
			name:     "nested field query that should not be transformed",
			input:    "creators.affiliation.affiliationIdentifier:*",
			expected: "creators.affiliation.affiliationIdentifier:*",
		},
		{
			name:     "query with forward slashes and field transformations",
			input:    "fundingReferences:* AND doi:10.5281/zenodo.1234",
			expected: "funding_references:* AND doi:10.5281\\/zenodo.1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformQueryFieldNames(tt.input)
			if result != tt.expected {
				t.Errorf("transformQueryFieldNames(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
