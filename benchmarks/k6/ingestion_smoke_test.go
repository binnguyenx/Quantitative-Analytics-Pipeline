package k6

import (
	"os"
	"strings"
	"testing"
)

func TestIngestionBenchmarkScriptSmoke(t *testing.T) {
	data, err := os.ReadFile("ingestion.js")
	if err != nil {
		t.Fatalf("failed to read k6 script: %v", err)
	}
	content := string(data)

	required := []string{
		"export const options",
		"export function setup()",
		"export default function",
		"/api/v1/users",
		"/profile",
	}

	for _, marker := range required {
		if !strings.Contains(content, marker) {
			t.Fatalf("missing expected marker in k6 script: %s", marker)
		}
	}
}
