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

func TestGitHubWorkflow_AcceptanceStepQuarantinesFlakyTests(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("skipping CI-only workflow assertion test when CI is not set")
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve current file path")
	}

	workflowPath := filepath.Join(filepath.Dir(thisFile), "..", "..", ".github", "workflows", "test.yml")

	// #nosec G304 -- Test reads repo-local workflow path derived from runtime.Caller.
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read workflow file (%s): %v", workflowPath, err)
	}

	workflow := string(content)
	acceptanceGoTestCommandRe := regexp.MustCompile(`(?s)run:\s*go\s+test\s+\./internal/provider/\.\.\.`)
	if !acceptanceGoTestCommandRe.MatchString(workflow) {
		t.Fatalf("workflow must include acceptance go test command for internal/provider tests")
	}

	runAcceptanceRegexRe := regexp.MustCompile(`(?s)go\s+test\s+\./internal/provider/\.\.\..*?-run(?:=|\s+)(?:'\^TestAcc'|"\^TestAcc"|\^TestAcc)`)
	if !runAcceptanceRegexRe.MatchString(workflow) {
		t.Fatalf("acceptance workflow step must include -run '^TestAcc' to keep CI scope explicit")
	}

	flakySkipRegexRe := regexp.MustCompile(`(?s)go\s+test\s+\./internal/provider/\.\.\..*?-skip(?:=|\s+)(?:'TestAccACLResource_WithRole\|TestAccRoleResource'|"TestAccACLResource_WithRole\|TestAccRoleResource"|TestAccACLResource_WithRole\|TestAccRoleResource)`)
	if !flakySkipRegexRe.MatchString(workflow) {
		t.Fatalf("acceptance workflow step must quarantine known flaky tests with explicit -skip regex")
	}

	if !strings.Contains(workflow, "TODO(#65)") {
		t.Fatalf("workflow must document quarantine with TODO(#65)")
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
