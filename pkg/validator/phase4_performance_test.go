package validator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// RW-15 — Phase 4 Performance Validation Test
// PRD ref: PRD Section 20 — Performance validation (<60s runtime)
// Owner file: pkg/validator/phase4_performance_test.go
// Target: verify all phase4 tests run within 60 seconds
// =============================================================================

// TestPhase4_Performance_RunAll verifies all phase4 tests complete in <60s.
// Uses testing.Short() to skip in short mode (e.g., CI -short flag).
func TestPhase4_Performance_RunAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Discover all phase4 test files
	phase4Files, err := discoverPhase4Files()
	if err != nil {
		t.Fatalf("failed to discover phase4 files: %v", err)
	}

	if len(phase4Files) == 0 {
		t.Fatal("no phase4 test files found")
	}

	t.Logf("Found %d phase4 test files", len(phase4Files))

	// Run all phase4 tests and measure time
	start := time.Now()

	passed, failed, timedOut := runPhase4Tests(t, phase4Files)

	elapsed := time.Since(start)

	// Report timing per group
	reportTimings(t, phase4Files, elapsed)

	// Assert total time < 60 seconds
	const maxDuration = 60 * time.Second
	if elapsed > maxDuration {
		t.Errorf("Phase4 performance regression: took %v, expected < %v (failed: %d, passed: %d, timedOut: %d)",
			elapsed, maxDuration, failed, passed, timedOut)
	} else {
		t.Logf("✓ Phase4 tests completed in %v (< %v)", elapsed, maxDuration)
	}

	if failed > 0 {
		t.Errorf("%d phase4 tests failed", failed)
	}
	if timedOut > 0 {
		t.Errorf("%d phase4 tests timed out", timedOut)
	}
}

// discoverPhase4Files finds all phase4 test files in pkg/validator.
func discoverPhase4Files() ([]string, error) {
	const phase4Pattern = "pkg/validator/phase4_"
	const testSuffix = "_test.go"

	entries, err := os.ReadDir("pkg/validator")
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, phase4Pattern) && strings.HasSuffix(name, testSuffix) {
			files = append(files, name)
		}
	}
	return files, nil
}

// runPhase4Tests executes all phase4 tests via `go test` and returns counts.
func runPhase4Tests(t *testing.T, files []string) (passed, failed, timedOut int) {
	// Build -run pattern for all phase4 test functions
	patternParts := make([]string, 0, len(files))
	for _, f := range files {
		// Extract test group name from filename: phase4_xxx_test.go -> xxx
		group := strings.TrimSuffix(strings.TrimPrefix(f, "phase4_"), "_test.go")
		patternParts = append(patternParts, "TestPhase4_"+group)
	}
	runPattern := strings.Join(patternParts, "|")

	args := []string{
		"test", "-v", "-timeout", "65s",
		"-run", runPattern,
		"./pkg/validator/...",
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	// Parse output for pass/fail counts
	outputStr := string(output)
	t.Logf("Phase4 test output (last 2000 chars):\n%s", trimLast(outputStr, 2000))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 124 {
				// Exit code 124 = timeout
				timedOut = -1 // unknown count, flag as timed out
				return
			}
			failed = parseFailCount(outputStr)
		} else {
			t.Logf("go test error (non-exit): %v", err)
			failed = len(files) // conservative estimate
		}
	}

	passed = parsePassCount(outputStr)
	failed = parseFailCount(outputStr)

	return
}

// reportTimings logs timing breakdown for each operator group.
func reportTimings(t *testing.T, files []string, total time.Duration) {
	t.Logf("=== Phase4 Performance Timings ===")
	t.Logf("Total: %v", total)
	for _, f := range files {
		group := strings.TrimSuffix(strings.TrimPrefix(f, "phase4_"), "_test.go")
		t.Logf("  [%s]", group)
	}
	t.Logf("==================================")
}

// trimLast returns the last n characters of s.
func trimLast(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// parsePassCount extracts "passed" count from go test output.
func parsePassCount(output string) int {
	// Look for patterns like "--- PASS: TestPhase4_..."
	lines := strings.Split(output, "\n")
	count := 0
	for _, l := range lines {
		if strings.Contains(l, "--- PASS:") {
			count++
		}
	}
	return count
}

// parseFailCount extracts "failed" count from go test output.
func parseFailCount(output string) int {
	// Look for patterns like "--- FAIL: TestPhase4_..."
	lines := strings.Split(output, "\n")
	count := 0
	for _, l := range lines {
		if strings.Contains(l, "--- FAIL:") {
			count++
		}
	}
	return count
}

// BenchmarkPhase4_RunAll is a benchmark variant for profiling phase4 tests.
// Run with: go test -bench=BenchmarkPhase4_RunAll ./pkg/validator/...
func BenchmarkPhase4_RunAll(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	files, err := discoverPhase4Files()
	if err != nil {
		b.Fatalf("failed to discover phase4 files: %v", err)
	}

	for i := 0; i < b.N; i++ {
		start := time.Now()

		patternParts := make([]string, 0, len(files))
		for _, f := range files {
			group := strings.TrimSuffix(strings.TrimPrefix(f, "phase4_"), "_test.go")
			patternParts = append(patternParts, "TestPhase4_"+group)
		}
		runPattern := strings.Join(patternParts, "|")

		args := []string{
			"test", "-timeout", "65s",
			"-run", runPattern,
			"./pkg/validator/...",
		}

		cmd := exec.Command("go", args...)
		cmd.Dir = "."
		_ = cmd.Run()

		b.SetBytes(time.Since(start).Milliseconds())
	}
}