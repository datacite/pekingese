package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	OSAddress = os.Getenv("OPENSEARCH_ADDRESS")
	IndexName = os.Getenv("OPENSEARCH_INDEX")
	osClient  OSClient // FOR INITIAL DEV ONLY - DO NOT USE GLOBAL IN PRODUCTION (probably)
)

func main() {
	osClient = InitOpenSearch()

	router := gin.Default()

	router.GET("/", getData)

	router.Run(":8080")
}

func getData(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	clientId := c.DefaultQuery("client_id", "")
	providerId := c.DefaultQuery("provider_id", "")

	fieldsPresent := strings.Split(c.DefaultQuery("present", ""), ",")
	fieldsDistribution := strings.Split(c.DefaultQuery("distribution", ""), ",")

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

	c.IndentedJSON(http.StatusOK, results)
}
