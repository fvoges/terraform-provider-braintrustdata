package provider

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestGNUmakefile_TestAccTargetRunsOnlyAcceptanceTests(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve current file path")
	}

	makefilePath := filepath.Join(filepath.Dir(thisFile), "..", "..", "GNUmakefile")

	// #nosec G304 -- Test reads repo-local GNUmakefile path derived from runtime.Caller.
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read GNUmakefile (%s): %v", makefilePath, err)
	}

	recipe, ok := findMakeTargetRecipe(string(content), "testacc")
	if !ok {
		t.Fatalf("failed to locate testacc target command in GNUmakefile")
	}

	runValueRe := regexp.MustCompile(`-run(?:=|\s+)(?:"([^"]+)"|'([^']+)'|([^\s]+))`)
	runMatch := runValueRe.FindStringSubmatch(recipe)
	if len(runMatch) == 0 {
		t.Fatalf("GNUmakefile testacc target must include a -run filter targeting TestAcc; recipe: %q", recipe)
	}

	runValue := firstNonEmpty(runMatch[1:]...)
	if !regexp.MustCompile(`^\^?TestAcc`).MatchString(runValue) {
		t.Fatalf("GNUmakefile testacc target -run value must target TestAcc tests; got %q", runValue)
	}
}

func findMakeTargetRecipe(content string, target string) (string, bool) {
	lines := strings.Split(content, "\n")
	header := target + ":"
	inTarget := false
	var recipeLines []string

	for _, line := range lines {
		if !inTarget {
			if strings.HasPrefix(line, header) {
				inTarget = true
			}
			continue
		}

		if line == "" {
			if len(recipeLines) > 0 {
				break
			}
			continue
		}

		isRecipe := strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ")
		if isRecipe {
			recipeLines = append(recipeLines, strings.TrimSpace(line))
			continue
		}

		// Stop at the next top-level target.
		if strings.Contains(line, ":") {
			break
		}
	}

	if len(recipeLines) == 0 {
		return "", false
	}

	return strings.Join(recipeLines, " "), true
}

func firstNonEmpty(values ...string) string {
	idx := slices.IndexFunc(values, func(v string) bool { return v != "" })
	if idx == -1 {
		return ""
	}

	return values[idx]
}
