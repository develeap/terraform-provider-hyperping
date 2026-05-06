// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package uptimerobot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFunc is an http.RoundTripper backed by a function. Letting tests
// inject a transport keeps the production NewClient surface unchanged while
// still allowing full request/response control.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newClientWithTransport(rt http.RoundTripper) *Client {
	c := NewClient("test-api-key")
	c.httpClient.Transport = rt
	return c
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestFlexibleInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FlexibleInt
		wantErr  bool
	}{
		{"number zero", "0", FlexibleInt(0), false},
		{"number positive", "42", FlexibleInt(42), false},
		{"number negative", "-1", FlexibleInt(-1), false},
		{"string number", `"42"`, FlexibleInt(42), false},
		{"string zero", `"0"`, FlexibleInt(0), false},
		{"string empty", `""`, FlexibleInt(0), false},
		{"boolean", "true", FlexibleInt(0), true},
		{"invalid string", `"abc"`, FlexibleInt(0), true},
		{"null", "null", FlexibleInt(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fi FlexibleInt
			err := json.Unmarshal([]byte(tt.input), &fi)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, fi)
			}
		})
	}
}

func TestFlexibleInt_InStruct(t *testing.T) {
	type testStruct struct {
		Port    *FlexibleInt `json:"port,omitempty"`
		SubType *FlexibleInt `json:"sub_type,omitempty"`
		Method  *FlexibleInt `json:"http_method,omitempty"`
	}

	tests := []struct {
		name     string
		json     string
		expected testStruct
	}{
		{
			name: "numeric fields",
			json: `{"port": 443, "sub_type": 3, "http_method": 1}`,
			expected: testStruct{
				Port:    flexIntPtr(443),
				SubType: flexIntPtr(3),
				Method:  flexIntPtr(1),
			},
		},
		{
			name: "string fields",
			json: `{"port": "8080", "sub_type": "2", "http_method": "6"}`,
			expected: testStruct{
				Port:    flexIntPtr(8080),
				SubType: flexIntPtr(2),
				Method:  flexIntPtr(6),
			},
		},
		{
			name: "empty string sub_type",
			json: `{"sub_type": ""}`,
			expected: testStruct{
				SubType: flexIntPtr(0),
			},
		},
		{
			name:     "missing optional fields",
			json:     `{}`,
			expected: testStruct{},
		},
		{
			name: "mixed types",
			json: `{"port": 443, "sub_type": "3"}`,
			expected: testStruct{
				Port:    flexIntPtr(443),
				SubType: flexIntPtr(3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testStruct
			err := json.Unmarshal([]byte(tt.json), &result)
			require.NoError(t, err)

			if tt.expected.Port != nil {
				require.NotNil(t, result.Port)
				assert.Equal(t, int(*tt.expected.Port), int(*result.Port))
			} else {
				assert.Nil(t, result.Port)
			}

			if tt.expected.SubType != nil {
				require.NotNil(t, result.SubType)
				assert.Equal(t, int(*tt.expected.SubType), int(*result.SubType))
			} else {
				assert.Nil(t, result.SubType)
			}

			if tt.expected.Method != nil {
				require.NotNil(t, result.Method)
				assert.Equal(t, int(*tt.expected.Method), int(*result.Method))
			} else {
				assert.Nil(t, result.Method)
			}
		})
	}
}

func flexIntPtr(n int) *FlexibleInt {
	fi := FlexibleInt(n)
	return &fi
}

// =============================================================================
// Client tests
// =============================================================================

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("k")
	assert.Equal(t, "k", c.apiKey)
	require.NotNil(t, c.httpClient)
	assert.NotZero(t, c.httpClient.Timeout, "expected a non-zero default timeout")
}

func TestGetMonitors_Success(t *testing.T) {
	var captured *http.Request
	var capturedBody []byte
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		captured = r
		capturedBody, _ = io.ReadAll(r.Body)
		return jsonResponse(200, `{"stat":"ok","monitors":[{"id":1,"friendly_name":"A","url":"https://a.example.com","type":1,"interval":60,"status":2}]}`), nil
	})
	c := newClientWithTransport(rt)

	got, err := c.GetMonitors(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "A", got[0].FriendlyName)

	require.NotNil(t, captured)
	assert.Equal(t, "POST", captured.Method)
	assert.Contains(t, captured.URL.String(), "/getMonitors")
	assert.Equal(t, "application/json", captured.Header.Get("Content-Type"))

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(capturedBody, &payload))
	assert.Equal(t, "test-api-key", payload["api_key"])
	assert.Equal(t, "json", payload["format"])
}

func TestGetMonitors_APIErrorStruct(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"stat":"fail","error":{"type":"invalid_parameter","message":"bad key"}}`), nil
	})
	_, err := newClientWithTransport(rt).GetMonitors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_parameter")
	assert.Contains(t, err.Error(), "bad key")
}

func TestGetMonitors_StatNotOK_NoErrorStruct(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"stat":"fail"}`), nil
	})
	_, err := newClientWithTransport(rt).GetMonitors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API returned status")
}

func TestGetMonitors_NonOKHTTPStatus(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `unauthorized`), nil
	})
	_, err := newClientWithTransport(rt).GetMonitors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 401")
}

func TestGetMonitors_NetworkError(t *testing.T) {
	netErr := errors.New("dial tcp: refused")
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, netErr
	})
	_, err := newClientWithTransport(rt).GetMonitors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "executing request")
}

func TestGetMonitors_BadJSON(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `not json`), nil
	})
	_, err := newClientWithTransport(rt).GetMonitors(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding response")
}

func TestGetAlertContacts_Success(t *testing.T) {
	var captured *http.Request
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		captured = r
		return jsonResponse(200, `{"stat":"ok","alert_contacts":[{"id":"1","friendly_name":"Ops","type":2,"value":"ops@example.com","status":2}]}`), nil
	})
	got, err := newClientWithTransport(rt).GetAlertContacts(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Ops", got[0].FriendlyName)
	assert.Equal(t, 2, got[0].Type)
	assert.Contains(t, captured.URL.String(), "/getAlertContacts")
}

func TestGetAlertContacts_StatFail(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"stat":"fail","error":{"type":"forbidden","message":"nope"}}`), nil
	})
	_, err := newClientWithTransport(rt).GetAlertContacts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestGetAlertContacts_BadJSON(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{`), nil
	})
	_, err := newClientWithTransport(rt).GetAlertContacts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding response")
}
