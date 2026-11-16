package cmd

import (
	"reflect"
	"testing"
)

func TestParseKeyValueOptions(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "no values",
			values:   nil,
			expected: nil,
		},
		{
			name:   "string values",
			values: []string{"bucket_name=my-bucket", "blobstore_path=./.blobs"},
			expected: map[string]interface{}{
				"bucket_name":    "my-bucket",
				"blobstore_path": "./.blobs",
			},
		},
		{
			name:   "typed values",
			values: []string{"use_ssl=true", "insecure=false", "port=443", "ratio=1.5", "scale=1e3"},
			expected: map[string]interface{}{
				"use_ssl":  true,
				"insecure": false,
				"port":     int64(443),
				"ratio":    1.5,
				"scale":    float64(1000),
			},
		},
		{
			name:   "leading zeros stay strings",
			values: []string{"id=007", "zero=0"},
			expected: map[string]interface{}{
				"id":   "007",
				"zero": int64(0),
			},
		},
		{
			name:   "value may contain equals signs",
			values: []string{"key=a=b"},
			expected: map[string]interface{}{
				"key": "a=b",
			},
		},
		{
			name:   "quoted values stay strings",
			values: []string{`bucket_name="123"`, `access_key_id='true'`},
			expected: map[string]interface{}{
				"bucket_name":   "123",
				"access_key_id": "true",
			},
		},
		{
			name:    "missing separator",
			values:  []string{"missing"},
			wantErr: true,
		},
		{
			name:    "empty key",
			values:  []string{"=value"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := parseKeyValueOptions(test.values)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("expected %#v, got %#v", test.expected, actual)
			}
		})
	}
}
