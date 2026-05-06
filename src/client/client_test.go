package client

import "testing"

func TestBinaryNameAndUserAgent(t *testing.T) {
	if BinaryName != "caspbx-cli" {
		t.Fatalf("unexpected binary name %q", BinaryName)
	}
	if UserAgent("dev") != "caspbx-cli/dev" {
		t.Fatalf("unexpected user agent %q", UserAgent("dev"))
	}
}
