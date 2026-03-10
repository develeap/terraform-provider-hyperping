// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package uptimerobot

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
