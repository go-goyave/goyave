package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMultiValuesHeader(t *testing.T) {
	cases := []struct {
		desc   string
		header string
		want   []HeaderValue
	}{
		{
			desc:   "equal_quality_values",
			header: "text/html;q=0.8,text/*;q=0.8,*/*;q=0.8",
			want: []HeaderValue{
				{Value: "text/html", Priority: 0.8},
				{Value: "text/*", Priority: 0.8},
				{Value: "*/*", Priority: 0.8},
			},
		},
		{
			desc:   "equal_quality_values_reverse_order",
			header: "*/*;q=0.8,text/*;q=0.8,text/html;q=0.8",
			want: []HeaderValue{
				{Value: "text/html", Priority: 0.8},
				{Value: "text/*", Priority: 0.8},
				{Value: "*/*", Priority: 0.8},
			},
		},
		{
			desc:   "default_priority",
			header: "text/html,text/*;q=0.5,*/*;q=0.7",
			want: []HeaderValue{
				{Value: "text/html", Priority: 1},
				{Value: "*/*", Priority: 0.7},
				{Value: "text/*", Priority: 0.5},
			},
		},
		{
			desc:   "whitespace_around_values",
			header: "fr , fr-FR;q=0.8, en-US ;q=0.5, *;q=0.3, en-*;q=0.3, en;q=0.3",
			want: []HeaderValue{
				{Value: "fr", Priority: 1},
				{Value: "fr-FR", Priority: 0.8},
				{Value: "en-US", Priority: 0.5},
				{Value: "en-*", Priority: 0.3},
				{Value: "en", Priority: 0.3},
				{Value: "*", Priority: 0.3},
			},
		},
		{
			desc:   "single_value",
			header: "fr",
			want:   []HeaderValue{{Value: "fr", Priority: 1}},
		},
		{
			desc:   "single_value_with_quality_value",
			header: "fr;q=0.3",
			want:   []HeaderValue{{Value: "fr", Priority: 0.3}},
		},
		{
			desc:   "empty",
			header: "",
			want:   []HeaderValue{},
		},
		{
			desc:   "whitespace_only",
			header: "   ",
			want:   []HeaderValue{},
		},
		{
			desc:   "whitespace_before_quality_value",
			header: "gzip; q=0.9, br;q=0.8",
			want:   []HeaderValue{{Value: "gzip", Priority: 0.9}, {Value: "br", Priority: 0.8}},
		},
		{
			desc:   "whitespace_around_quality_value",
			header: "gzip;q= 0.9, br;q=0.8 , bz;q = 0.7 ",
			want: []HeaderValue{
				{Value: "gzip", Priority: 0.9},
				{Value: "br", Priority: 0.8},
				{Value: "bz", Priority: 0.7},
			},
		},
		{
			desc:   "integer_quality_value",
			header: "gzip;q=1, br;q=0.9",
			want:   []HeaderValue{{Value: "gzip", Priority: 1}, {Value: "br", Priority: 0.9}},
		},
		{
			desc:   "zero_quality_value",
			header: "gzip;q=0, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "quality_value_after_another_parameter",
			header: "text/html;charset=utf-8;q=0.9, text/plain;q=0.5",
			want:   []HeaderValue{{Value: "text/html", Priority: 0.9}, {Value: "text/plain", Priority: 0.5}},
		},
		{
			desc:   "no_quality_value_parameter",
			header: "text/html;charset=utf-8",
			want:   []HeaderValue{{Value: "text/html", Priority: 1}},
		},
		{
			desc:   "uppercase_parameter_name",
			header: "gzip;Q=0.5, br;q=0.1",
			want:   []HeaderValue{{Value: "gzip", Priority: 0.5}, {Value: "br", Priority: 0.1}},
		},
		{
			desc:   "empty_parameter_list",
			header: "text/html;",
			want:   []HeaderValue{{Value: "text/html", Priority: 1}},
		},
		{
			desc:   "empty_parameter_list_two_values",
			header: "text/html;,application/json",
			want:   []HeaderValue{{Value: "text/html", Priority: 1}, {Value: "application/json", Priority: 1}},
		},
		{
			desc:   "not_a_number",
			header: "gzip;q=abc, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "out_of_range",
			header: "gzip;q=2, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "negative_quality_value",
			header: "gzip;q=-0.5, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "scientific_notation",
			header: "gzip;q=1e-2, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "NaN",
			header: "gzip;q=NaN, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "plus_sign",
			header: "gzip;q=+0.5, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "two_dots",
			header: "gzip;q=0.5.5, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
		{
			desc:   "too_many_decimals",
			header: "gzip;q=0.55555, br;q=0.1",
			want:   []HeaderValue{{Value: "br", Priority: 0.1}, {Value: "gzip", Priority: 0}},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, ParseMultiValuesHeader(c.header))
		})
	}
}
