package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Structures returned by OpenSearch for specific aggregations
type PresentAggregation struct {
	Buckets struct {
		Present struct {
			Count int `json:"doc_count"`
		} `json:"present"`
		Absent struct {
			Count int `json:"doc_count"`
		} `json:"absent"`
	} `json:"buckets"`
}

type DistributionAggregation struct {
	DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
	OtherCount              int `json:"sum_other_doc_count"`
	Buckets                 []struct {
		Value string `json:"key"`
		Count int    `json:"doc_count"`
	} `json:"buckets"`
}

// Structures for API response
type PresentResponse struct {
	Field   string `json:"field"`
	Percent int    `json:"percent"`
}

type DistributionResponse struct {
	Field  string `json:"field"`
	Values []ValuePercent
}

type ValuePercent struct {
	Value   string `json:"value"`
	Percent int    `json:"percent"`
}

type APIResponse struct {
	Present      []PresentResponse      `json:"present"`
	Distribution []DistributionResponse `json:"distribution"`
}

func ParseSearchResp(resp OSResponse) (*APIResponse, error) {
	var aggs map[string]json.RawMessage
	if err := json.Unmarshal(resp.Aggregations, &aggs); err != nil {
		return nil, fmt.Errorf("unmarshal aggregations: %w", err)
	}

	var response APIResponse

	for key, value := range aggs {
		isPresentAgg := strings.HasPrefix(key, "present_")
		isDistributionAgg := strings.HasPrefix(key, "distribution_")

		if !isPresentAgg && !isDistributionAgg {
			continue
		}

		if isPresentAgg {
			present, err := ParsePresentAgg(key, value)
			if err != nil {
				return nil, err
			}

			response.Present = append(response.Present, *present)
		}

		if isDistributionAgg {
			distribution, err := ParseDistributionAgg(key, value)
			if err != nil {
				return nil, err
			}

			response.Distribution = append(response.Distribution, *distribution)
		}
	}

	return &response, nil
}

func ParsePresentAgg(key string, value json.RawMessage) (*PresentResponse, error) {
	var present PresentAggregation
	if err := json.Unmarshal(value, &present); err != nil {
		return nil, fmt.Errorf("unmarshal present aggregation %s: %w", key, err)
	}

	presentCount, totalCount := present.Buckets.Present.Count, present.Buckets.Present.Count+present.Buckets.Absent.Count

	response := &PresentResponse{
		Field:   strings.TrimPrefix(key, "present_"),
		Percent: calcPercent(presentCount, totalCount),
	}

	return response, nil
}

func ParseDistributionAgg(key string, value json.RawMessage) (*DistributionResponse, error) {
	var distribution DistributionAggregation
	if err := json.Unmarshal(value, &distribution); err != nil {
		return nil, fmt.Errorf("unmarshal distribution aggregation %s: %w", key, err)
	}

	totalCount := distribution.OtherCount

	for _, value := range distribution.Buckets {
		totalCount += value.Count
	}

	response := &DistributionResponse{
		Field:  strings.TrimPrefix(key, "distribution_"),
		Values: make([]ValuePercent, 0, len(distribution.Buckets)),
	}

	for _, value := range distribution.Buckets {
		fieldPercent := ValuePercent{
			Value:   value.Value,
			Percent: calcPercent(value.Count, totalCount),
		}
		response.Values = append(response.Values, fieldPercent)
	}

	return response, nil
}

func calcPercent(part, total int) int {
	if total == 0 {
		return 0
	}
	return part * 100 / total
}
