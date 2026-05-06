package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--help"}, "renamed-server", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "renamed-server") {
		t.Fatalf("expected help output to contain binary name, got %q", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--version"}, "renamed-server", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "renamed-server") {
		t.Fatalf("expected version output to contain binary name, got %q", stdout.String())
	}
}

func TestRunStatus(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--status"}, "caspbx", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "bootstrap status: ready") {
		t.Fatalf("unexpected status output %q", stdout.String())
	}
}

func TestRunDefault(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, "caspbx", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "bootstrap scaffold is active") {
		t.Fatalf("unexpected startup output %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Running in mode: production") {
		t.Fatalf("expected startup mode output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "HTTP runtime scaffold is ready") {
		t.Fatalf("expected runtime scaffold output, got %q", stdout.String())
	}
}

func TestRunDefaultWithOfficialSite(t *testing.T) {
	oldOfficialSite := OfficialSite
	OfficialSite = "https://example.invalid"
	defer func() {
		OfficialSite = oldOfficialSite
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, "caspbx", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Official site: https://example.invalid") {
		t.Fatalf("expected official site output, got %q", stdout.String())
	}
}

func TestRunDefaultWithEnvironmentMode(t *testing.T) {
	oldMode := os.Getenv("MODE")
	oldDebug := os.Getenv("DEBUG")
	t.Setenv("MODE", "development")
	t.Setenv("DEBUG", "enabled")
	defer func() {
		if oldMode != "" {
			_ = os.Setenv("MODE", oldMode)
		}
		if oldDebug != "" {
			_ = os.Setenv("DEBUG", oldDebug)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(nil, "caspbx", &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Running in mode: development [debugging]") {
		t.Fatalf("expected environment mode output, got %q", stdout.String())
	}
}

func TestRunInvalidFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--nope"}, "caspbx", &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output on invalid flag, got %q", stdout.String())
	}
}

func TestMain(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	os.Args = []string{"renamed-server", "--version"}
	os.Stdout = w

	main()
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if !strings.Contains(buf.String(), "renamed-server") {
		t.Fatalf("expected main output to contain binary name, got %q", buf.String())
	}
}
