// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package openSearchClient

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type CountReq struct {
	Indices []string
	Body    io.Reader
	Header  http.Header
	Params  map[string]string
}

// GetRequest returns the *http.Request that gets executed by the client
func (r CountReq) GetRequest() (*http.Request, error) {
	if r.Params == nil {
		r.Params = make(map[string]string)
	}
	var path string
	if len(r.Indices) > 0 {
		path = fmt.Sprintf("/%s/_count", strings.Join(r.Indices, ","))
	} else {
		path = "/_count"
	}

	return opensearch.BuildRequest(
		"GET",
		path,
		r.Body,
		r.Params,
		r.Header,
	)
}

type CountResp struct {
	Count  int64      `json:"count"`
	Shards ShardStats `json:"_shards"`
}

type ShardStats struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

func (r CountResp) Inspect() opensearchapi.Inspect {
	return opensearchapi.Inspect{}
}
