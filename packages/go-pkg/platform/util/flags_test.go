package util

import "testing"

func TestParseLocationFromStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "parses location value",
			input: "bruttice-loc:tulsa",
			want:  "tulsa",
		},
		{
			name:  "parses location embedded in longer string",
			input: "some text bruttice-loc:downtown more text",
			want:  "downtown",
		},
		{
			name:  "returns empty string when flag absent",
			input: "no flag here",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLocationFromStr(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAge(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "parses age value",
			input: "bruttice-ages:18+",
			want:  "18+",
		},
		{
			name:  "parses 21+ age",
			input: "bruttice-ages:21+",
			want:  "21+",
		},
		{
			name:  "returns empty string when flag absent",
			input: "no flag here",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAge(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTagsFromStr(t *testing.T) {
	testCases := []string{
		// last line of file
		"other chars\n" + flagTags + ":tag1:tag2",
		"other chars\n" + flagTags + ":tag1:tag2:",
		// newline after tags
		flagTags + ":tag1:tag2\n",
		flagTags + ":tag1:tag2:\n",
		// space after tags
		flagTags + ":tag1:tag2 ",
		flagTags + ":tag1:tag2: ",
	}

	var testOutput []string
	output := []string{"tag1", "tag2"}

	for _, tagStr := range testCases {
		testOutput = ParseTagsFromStr(tagStr)

		if len(testOutput) != 2 {
			t.Errorf("%s: invalid len of %d instead of 2", tagStr, len(testOutput))
		}

		if testOutput[0] != output[0] ||
			testOutput[1] != output[1] {
			t.Errorf("%s: unequal tags: %+v != %+v (expected)", tagStr, testOutput, output)
		}
	}
}
