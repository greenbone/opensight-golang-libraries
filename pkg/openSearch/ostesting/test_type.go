// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package ostesting

// TestType can be used as generic document object for testing
type TestType struct {
	ID  string `json:"id"`
	Val string `json:"val"`
}

var (
	// testTypeMapping is an index mapping for testType
	testTypeMapping string = `{
    	"mappings": {
            "properties": {
            "id": {
                "type": "keyword"
            	},
			"val": {
                "type": "keyword"
            	}
        	}
    	}
	}`
)
