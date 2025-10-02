# Metadata Completenesss API

API for the Metadata Completeness Dashboard

## Running Locally

### Set up OpenSearch Tunnel

The API requires a tunnel into OpenSearch. This can be set up using the instructions here: https://datacite.atlassian.net/wiki/spaces/DAT/pages/1038811137/How+to+SSH+tunnel+to+ElasticSearch.

You may need to prepend `0.0.0.0` to your local forward in the `.ssh/config` to allow Docker to connect to OpenSearch:

```
Host es-stage
    ...
    LocalForward 0.0.0.0:9202 ...
```

### Start the API

1. `ssh es-stage`
2. `docker-compose --profile dev up` (in a separate terminal)

### Using the API

The API is served on `http://localhost:8080/`<br>
_Currently there is only one endpoint at the root_

Example URL that fetches `present` and `distribution` aggregations for DataCite, along with a test query:<br>
http://localhost:8080/?client_id=datacite.datacite&present=creators,creators.name&distribution=types.resourceTypeGeneral&query=test

**_Supported Query Parameters_**

- `client_id`: string
- `provider_id`: string
- `query`: string
- `present`: []string - _comma separated list of fields for which to fetch the present/absent counts_
- `distribution`: []string - _comma separated list of fields for which to fetch the distribution values_
- `distribution_size`: number - _specifies the number of top results to return for each distribution field_

---

## Design Choices

### [net/http](https://pkg.go.dev/net/http) (Standard library for API routing)

I chose to use the standard `net/http` library because the needs for thes project are minimal. If it needs more features in the future - such as middleware or auth - I would switch to the [Chi](https://go-chi.io/#/) framework

### [defensestation/osquery](https://github.com/defensestation/osquery) (An idiomatic Go query builder for OpenSearch)

Without this package, writing OpenSearch queries in Go is quite tedious, especially if you need to do anything conditionally.
