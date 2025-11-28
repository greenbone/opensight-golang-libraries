// SPDX-FileCopyrightText: 2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package httpassert

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"github.com/yalp/jsonpath"
)

const IgnoreJsonValue = "<IGNORE>"

// nolint:interfacebloat
// Response interface provides fluent response assertions.
type Response interface {
	StatusCode(int) Response

	JsonPath(string, any) Response
	JsonPathJson(path string, expectedJson string) Response

	ContentType(contentType string) Response

	NoContent() Response

	Json(json string) Response
	JsonTemplate(json string, values map[string]any) Response
	JsonTemplateFile(path string, values map[string]any) Response
	JsonFile(path string) Response

	Body(body string) Response
	GetJsonBodyObject(target any) Response
	GetBody() string

	Log() Response
}

func (r *responseImpl) StatusCode(expected int) Response {
	require.Equal(r.t, expected, r.response.Code)
	return r
}

func (r *responseImpl) JsonPath(path string, expected any) Response {
	var tmp any
	if err := jsoniter.Unmarshal(r.response.Body.Bytes(), &tmp); err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}

	out, err := jsonpath.Read(tmp, path)
	if err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}

	switch v := expected.(type) {
	case Extractor:
		v(r.t, out)
		return r
	case Matcher:
		v(r.t, out)
		return r

	default:
		assert.Equal(r.t, expected, out)
		return r
	}
}

func (r *responseImpl) ContentType(ct string) Response {
	assert.Equal(r.t, ct, r.response.Header().Get("Content-Type"))
	return r
}

func (r *responseImpl) JsonPathJson(path string, expectedJson string) Response {
	var tmp any
	if err := jsoniter.Unmarshal(r.response.Body.Bytes(), &tmp); err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}
	out, err := jsonpath.Read(tmp, path)
	if err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}
	pathJson, err := jsoniter.Marshal(out)
	if err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}
	assert.JSONEq(r.t, expectedJson, string(pathJson))
	return r
}

func (r *responseImpl) NoContent() Response {
	assert.Equal(r.t, "", strings.TrimSpace(r.response.Body.String()))
	return r
}

func (r *responseImpl) Json(expectedJson string) Response {
	assert.JSONEq(r.t, expectedJson, r.response.Body.String())
	return r
}

func (r *responseImpl) JsonTemplate(expectedJsonTemplate string, values map[string]any) Response {
	expectedJson := expectedJsonTemplate

	// apply provided values into the template
	for k, v := range values {
		// normalize JSONPath-like keys (convert $.a[0].b to a.0.b)
		key := strings.TrimPrefix(k, "$.")
		key = strings.ReplaceAll(key, "[", ".")
		key = strings.ReplaceAll(key, "]", "")

		tmp, err := sjson.Set(expectedJson, key, v)
		if err != nil {
			assert.Fail(r.t, "JsonTemplate set value failed: "+err.Error())
			return r
		}
		expectedJson = tmp
	}

	// handle <IGNORE> values: replace the actual body with <IGNORE> in same paths
	actual := r.response.Body.String()
	for k, v := range values {
		if v != IgnoreJsonValue {
			continue
		}

		key := strings.TrimPrefix(k, "$.")
		key = strings.ReplaceAll(key, "[", ".")
		key = strings.ReplaceAll(key, "]", "")

		tmp, err := sjson.Set(actual, key, v)
		if err != nil {
			assert.Fail(r.t, "JsonTemplate ignore replacement failed: "+err.Error())
			return r
		}
		actual = tmp
	}

	// compare the resulting JSONs ignoring order
	valid := assert.JSONEq(r.t, expectedJson, actual)
	if !valid {
		r.Log()
	}

	return r
}

func (r *responseImpl) JsonTemplateFile(path string, values map[string]any) Response {
	content, err := os.ReadFile(path)
	if err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}
	return r.JsonTemplate(string(content), values)
}

func (r *responseImpl) Body(expected string) Response {
	assert.Equal(r.t, expected, r.response.Body.String())
	return r
}

func (r *responseImpl) JsonFile(path string) Response {
	content, err := os.ReadFile(path)
	if err != nil {
		assert.Fail(r.t, err.Error())
		return r
	}
	assert.JSONEq(r.t, string(content), r.response.Body.String())
	return r
}

func (r *responseImpl) GetBody() string {
	return r.response.Body.String()
}

func (r *responseImpl) GetJsonBodyObject(target any) Response {
	err := json.Unmarshal(r.response.Body.Bytes(), &target)
	require.NoError(r.t, err)
	return r
}

func (r *responseImpl) Log() Response {
	fmt.Printf("Response: %d\nHeaders: %v\nBody: %s\n",
		r.response.Code,
		r.response.Header(),
		r.response.Body.String())
	return r
}
