package scanner

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		s := New()

		if s.config.Pattern != "./..." {
			t.Errorf("expected pattern './...', got '%s'", s.config.Pattern)
		}
		if s.config.Dir != "." {
			t.Errorf("expected dir '.', got '%s'", s.config.Dir)
		}
		if len(s.config.IgnorePaths) != 0 {
			t.Errorf("expected empty ignore paths, got %v", s.config.IgnorePaths)
		}
	})

	t.Run("with options", func(t *testing.T) {
		s := New(
			WithPattern("./api/..."),
			WithDir("/project"),
			WithIgnorePaths("vendor/**", "test/**"),
		)

		if s.config.Pattern != "./api/..." {
			t.Errorf("expected pattern './api/...', got '%s'", s.config.Pattern)
		}
		if s.config.Dir != "/project" {
			t.Errorf("expected dir '/project', got '%s'", s.config.Dir)
		}
		if len(s.config.IgnorePaths) != 2 {
			t.Errorf("expected 2 ignore paths, got %d", len(s.config.IgnorePaths))
		}
	})
}

func TestNewWithConfig(t *testing.T) {
	t.Run("with explicit config", func(t *testing.T) {
		config := &Config{
			Pattern:     "./handlers/...",
			Dir:         "/app",
			IgnorePaths: []string{"legacy/**"},
		}

		s := NewWithConfig(config)

		if s.config.Pattern != "./handlers/..." {
			t.Errorf("expected pattern './handlers/...', got '%s'", s.config.Pattern)
		}
	})

	t.Run("fills defaults for empty config", func(t *testing.T) {
		config := &Config{}

		s := NewWithConfig(config)

		if s.config.Pattern != "./..." {
			t.Errorf("expected default pattern './...', got '%s'", s.config.Pattern)
		}
		if s.config.Dir != "." {
			t.Errorf("expected default dir '.', got '%s'", s.config.Dir)
		}
	})
}

func TestScanner_shouldIgnorePath(t *testing.T) {
	tests := []struct {
		name        string
		ignorePaths []string
		path        string
		want        bool
	}{
		{
			name:        "no ignore paths",
			ignorePaths: []string{},
			path:        "api/handlers/user.go",
			want:        false,
		},
		{
			name:        "exact directory match",
			ignorePaths: []string{"vendor"},
			path:        "vendor/github.com/pkg",
			want:        true,
		},
		{
			name:        "glob pattern with **",
			ignorePaths: []string{"vendor/**"},
			path:        "vendor/github.com/pkg/errors/errors.go",
			want:        true,
		},
		{
			name:        "glob pattern with /*",
			ignorePaths: []string{"test/*"},
			path:        "test/handler_test.go",
			want:        true,
		},
		{
			name:        "partial match in path",
			ignorePaths: []string{"legacy"},
			path:        "api/legacy/old_handler.go",
			want:        true,
		},
		{
			name:        "no match",
			ignorePaths: []string{"vendor/**", "test/**"},
			path:        "api/handlers/user.go",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(WithIgnorePaths(tt.ignorePaths...))

			got := s.shouldIgnorePath(tt.path)
			if got != tt.want {
				t.Errorf("shouldIgnorePath(%s) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestScanner_Scan(t *testing.T) {
	t.Run("scan current package", func(t *testing.T) {
		s := New(
			WithPattern("."),
			WithDir("."),
		)

		result, err := s.Scan()
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if result.Fset == nil {
			t.Error("expected FileSet, got nil")
		}
		if len(result.Files) == 0 {
			t.Error("expected at least one file")
		}

		// Verify we found scanner.go
		found := false
		for path := range result.Files {
			if path == "scanner.go" || containsSuffix(path, "scanner.go") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find scanner.go in scanned files")
		}
	})
}

func TestScanner_ScanFiles(t *testing.T) {
	t.Run("returns files and fset", func(t *testing.T) {
		s := New(
			WithPattern("."),
			WithDir("."),
		)

		files, fset, err := s.ScanFiles()
		if err != nil {
			t.Fatalf("ScanFiles() error = %v", err)
		}

		if fset == nil {
			t.Error("expected FileSet, got nil")
		}
		if len(files) == 0 {
			t.Error("expected at least one file")
		}
	})
}

func containsSuffix(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}
