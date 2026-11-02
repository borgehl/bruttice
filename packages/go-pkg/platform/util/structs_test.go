package util

import "testing"

type structTestFooBar struct{}
type structTestMyStruct struct{}
type structTestHTMLParser struct{}

func TestStructToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "two-word camel case",
			input: structTestFooBar{},
			want:  "struct_test_foo_bar",
		},
		{
			name:  "three-word camel case",
			input: structTestMyStruct{},
			want:  "struct_test_my_struct",
		},
		{
			name:  "pointer is dereferenced",
			input: &structTestFooBar{},
			want:  "struct_test_foo_bar",
		},
		{
			// "HTML" is treated as one word — consecutive caps are not split
			name:  "consecutive caps followed by lowercase",
			input: structTestHTMLParser{},
			want:  "struct_test_html_parser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StructToSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
