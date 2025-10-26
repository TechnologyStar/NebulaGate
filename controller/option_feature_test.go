package controller

import "testing"

func TestCanonicalFeatureSection(t *testing.T) {
	tests := map[string]string{
		"billing":      "billing",
		"Billing":      "billing",
		" finance ":    "billing",
		"fea_mance":    "billing",
		"Finance":      "billing",
		"governance":   "governance",
		"Public Logs":  "public_logs",
		"public-log":   "public_logs",
		"public\tlogs": "public_logs",
		"public_logs":  "public_logs",
		"publiclogs":   "public_logs",
		"public log":   "public_logs",
		"PUBLIC_LOG":   "public_logs",
	}

	for input, expected := range tests {
		got, ok := canonicalFeatureSection(input)
		if !ok {
			t.Fatalf("expected %s to be normalised", input)
		}

		if got != expected {
			t.Errorf("expected %s -> %s, got %s", input, expected, got)
		}
	}
}

func TestParseFeatureSectionsFromQueryFallback(t *testing.T) {
	sections := parseFeatureSections("billing, public logs ,unknown")

	if _, ok := sections["billing"]; !ok {
		t.Fatalf("expected billing to be included")
	}

	if _, ok := sections["public_logs"]; !ok {
		t.Fatalf("expected public_logs to be included")
	}

	if _, ok := sections["unknown"]; !ok {
		t.Fatalf("expected unknown to be included for forward compatibility")
	}
}
