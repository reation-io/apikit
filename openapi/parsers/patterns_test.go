package parsers

import (
	"testing"
)

func TestGetSectionRegex(t *testing.T) {
	tests := []struct {
		name      string
		directive string
		input     string
		wantMatch bool
		wantValue string
	}{
		{
			name:      "Description section",
			directive: "Description:",
			input: `Description:
This is a multi-line
description text.
Title: Something`,
			wantMatch: true,
			wantValue: "\nThis is a multi-line\ndescription text.\n",
		},
		{
			name:      "Responses section",
			directive: "Responses:",
			input: `Responses:
- 200: Success
- 404: Not Found
Security:`,
			wantMatch: true,
			wantValue: "\n- 200: Success\n- 404: Not Found\n",
		},
		{
			name:      "No match",
			directive: "NotFound:",
			input:     "Title: Test",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx := GetSectionRegex(tt.directive)
			matches := rx.FindStringSubmatch(tt.input)

			if tt.wantMatch {
				if len(matches) < 2 {
					t.Errorf("expected match for directive %s", tt.directive)
					return
				}
				if matches[1] != tt.wantValue {
					t.Errorf("got %q, want %q", matches[1], tt.wantValue)
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("unexpected match: %v", matches)
				}
			}
		})
	}
}

func TestGetSingleLineRegex(t *testing.T) {
	tests := []struct {
		name      string
		directive string
		input     string
		wantMatch bool
		wantValue string
	}{
		{
			name:      "Title directive",
			directive: "Title:",
			input:     "Title: My API",
			wantMatch: true,
			wantValue: "My API",
		},
		{
			name:      "Version directive",
			directive: "Version:",
			input:     "Version: 1.0.0",
			wantMatch: true,
			wantValue: "1.0.0",
		},
		{
			name:      "Case insensitive",
			directive: "title:",
			input:     "TITLE: Uppercase",
			wantMatch: true,
			wantValue: "Uppercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx := GetSingleLineRegex(tt.directive)
			matches := rx.FindStringSubmatch(tt.input)

			if tt.wantMatch {
				if len(matches) < 2 {
					t.Errorf("expected match for directive %s", tt.directive)
					return
				}
				if matches[1] != tt.wantValue {
					t.Errorf("got %q, want %q", matches[1], tt.wantValue)
				}
			}
		})
	}
}

func TestGetMapTypeRegex(t *testing.T) {
	rx := GetMapTypeRegex()

	tests := []struct {
		input    string
		wantKey  string
		wantVal  string
		wantOK   bool
	}{
		{"map[string]int", "string", "int", true},
		{"map[string][]User", "string", "[]User", true},
		{"map[int64]*Response", "int64", "*Response", true},
		{"string", "", "", false},
		{"[]string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := rx.FindStringSubmatch(tt.input)
			if tt.wantOK {
				if len(matches) < 3 {
					t.Errorf("expected match for %s", tt.input)
					return
				}
				if matches[1] != tt.wantKey {
					t.Errorf("key: got %q, want %q", matches[1], tt.wantKey)
				}
				if matches[2] != tt.wantVal {
					t.Errorf("val: got %q, want %q", matches[2], tt.wantVal)
				}
			} else if len(matches) > 0 {
				t.Errorf("unexpected match for %s: %v", tt.input, matches)
			}
		})
	}
}

func TestRegexCaching(t *testing.T) {
	// First call should compile and cache
	rx1 := GetSectionRegex("Test:")
	// Second call should return cached version
	rx2 := GetSectionRegex("Test:")

	if rx1 != rx2 {
		t.Error("expected same regex instance from cache")
	}
}

