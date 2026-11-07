package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func versionOf(t *testing.T, changelog string) string {
	t.Helper()
	out, err := exec.Command("sh", "../../scripts/version.sh", changelog).Output()
	if err != nil {
		t.Fatalf("scripts/version.sh %s: %v", changelog, err)
	}
	return strings.TrimSpace(string(out))
}

func TestVersionSkipsUnreleasedAndTakesTheNewestRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	body := "# Changelog\n\n## [Unreleased]\n\n- next\n\n## [1.2.3] - 2026-11-07\n\n## [1.2.2] - 2026-11-06\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := versionOf(t, path); got != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", got)
	}
}

func TestVersionOfTheRepoIsARelease(t *testing.T) {
	if got := versionOf(t, "../../CHANGELOG.md"); !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(got) {
		t.Fatalf("version = %q, want a bare semver like 1.16.0", got)
	}
}
