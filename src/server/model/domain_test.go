package model

import "testing"

func TestDefaultDomainConstraints(t *testing.T) {
	constraints := DefaultDomainConstraints()
	if constraints.MaxDomainsPerUser != 5 || constraints.MaxDomainsPerOrg != 20 {
		t.Fatalf("unexpected domain limits %+v", constraints)
	}
	if constraints.AllowWildcard {
		t.Fatalf("expected wildcards disabled by default")
	}
}

func TestNormalizeDomainName(t *testing.T) {
	if normalizedDomain := NormalizeDomainName(" Example.COM. "); normalizedDomain != "example.com" {
		t.Fatalf("unexpected normalized domain %q", normalizedDomain)
	}
}

func TestValidateDomainName(t *testing.T) {
	constraints := DefaultDomainConstraints()
	if validationError := ValidateDomainName("pbx.example.com", constraints); validationError != nil {
		t.Fatalf("expected subdomain to validate, got %v", validationError)
	}
	if validationError := ValidateDomainName("example.com", constraints); validationError != nil {
		t.Fatalf("expected apex domain to validate, got %v", validationError)
	}
	apexDisabledConstraints := DefaultDomainConstraints()
	apexDisabledConstraints.AllowApex = false
	if validationError := ValidateDomainName("example.com", apexDisabledConstraints); validationError == nil {
		t.Fatalf("expected apex domain to fail when disabled")
	}
	if validationError := ValidateDomainName("https://example.com", constraints); validationError == nil {
		t.Fatalf("expected scheme to fail validation")
	}
	if validationError := ValidateDomainName("", constraints); validationError == nil {
		t.Fatalf("expected empty domain to fail")
	}
	if validationError := ValidateDomainName("localhost", constraints); validationError == nil {
		t.Fatalf("expected reserved domain to fail")
	}
	if validationError := ValidateDomainName("pbx.local", constraints); validationError == nil {
		t.Fatalf("expected wildcard reserved domain pattern to fail")
	}
	if validationError := ValidateDomainName("site.example.edu", constraints); validationError == nil {
		t.Fatalf("expected blocked pattern to fail")
	}
	if validationError := ValidateDomainName("*.example.com", constraints); validationError == nil {
		t.Fatalf("expected wildcard to fail with default constraints")
	}

	wildcardConstraints := DefaultDomainConstraints()
	wildcardConstraints.AllowWildcard = true
	if validationError := ValidateDomainName("*.example.com", wildcardConstraints); validationError != nil {
		t.Fatalf("expected wildcard to validate when enabled, got %v", validationError)
	}

	subdomainDisabledConstraints := DefaultDomainConstraints()
	subdomainDisabledConstraints.AllowSubdomain = false
	if validationError := ValidateDomainName("pbx.example.com", subdomainDisabledConstraints); validationError == nil {
		t.Fatalf("expected subdomain to fail when disabled")
	}
	if validationError := ValidateDomainName("exa_mple.com", constraints); validationError == nil {
		t.Fatalf("expected invalid label to fail")
	}
	if validationError := ValidateDomainName("example.c", constraints); validationError == nil {
		t.Fatalf("expected short TLD to fail")
	}
}

func TestSelectSSLChallengeAndActive(t *testing.T) {
	if challengeType := SelectSSLChallenge(CustomDomain{SSLProvider: "cloudflare"}, true, true); challengeType != SSLChallengeDNS01 {
		t.Fatalf("expected dns challenge for explicit provider, got %s", challengeType)
	}
	if challengeType := SelectSSLChallenge(CustomDomain{IsWildcard: true}, true, true); challengeType != SSLChallengeDNS01 {
		t.Fatalf("expected dns challenge for wildcard, got %s", challengeType)
	}
	if challengeType := SelectSSLChallenge(CustomDomain{}, true, true); challengeType != SSLChallengeTLSALPN {
		t.Fatalf("expected tls-alpn challenge, got %s", challengeType)
	}
	if challengeType := SelectSSLChallenge(CustomDomain{}, false, true); challengeType != SSLChallengeHTTP01 {
		t.Fatalf("expected http challenge, got %s", challengeType)
	}
	if challengeType := SelectSSLChallenge(CustomDomain{}, false, false); challengeType != SSLChallengeDNS01 {
		t.Fatalf("expected dns fallback challenge, got %s", challengeType)
	}

	activeDomain := CustomDomain{Status: DomainStatusActive, VerificationStatus: VerificationStatusVerified}
	if !activeDomain.IsActive() {
		t.Fatalf("expected verified active domain to be active")
	}
	if (CustomDomain{Status: DomainStatusPending, VerificationStatus: VerificationStatusVerified}).IsActive() {
		t.Fatalf("expected pending domain to be inactive")
	}
}
