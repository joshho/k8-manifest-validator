#!/usr/bin/env bash
set -euo pipefail

# test.sh — Run unit tests and binary smoke test.
#
# Exit codes:
#   0 = all tests passed
#   1 = test failed

SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$SCRIPT_DIR"

# Find go — use explicit path (works locally and in CI with setup-go having fixed PATH)
if [[ -x /home/node/.local/go/go/bin/go ]]; then
  GO=/home/node/.local/go/go/bin/go
elif command -v go &>/dev/null; then
  GO=$(command -v go)
else
  echo "ERROR: go not found" >&2
  exit 1
fi

echo "=== test.sh ==="
echo ""
echo "[unit tests] go test ./..."
if ! "$GO" test ./...; then
  echo "UNIT TESTS FAILED"
  exit 1
fi
echo "Unit tests passed"

echo ""
echo "[binary smoke test]"
BINARY="dist/k8-manifest-validator"
if [[ ! -f "$BINARY" ]]; then
  echo "ERROR: Binary not found at $BINARY (run build.sh first)"
  exit 1
fi

# Smoke test: --version
VERSION_OUTPUT=$("$BINARY" --version 2>&1 || true)
if [[ -z "$VERSION_OUTPUT" ]]; then
  echo "ERROR: Binary --version produced no output"
  exit 1
fi
echo "Binary version: $VERSION_OUTPUT"

# Smoke test: --help
HELP_OUTPUT=$("$BINARY" --help 2>&1 || true)
if [[ -z "$HELP_OUTPUT" ]]; then
  echo "ERROR: Binary --help produced no output"
  exit 1
fi
echo "Binary help: $HELP_OUTPUT"

# =============================================================================
# Integration tests: binary vs CR/CRD fixture pairs
# =============================================================================
echo ""
echo "[binary CR/CRD fixture integration tests]"

FAILED=0
PASSED=0

run_fixture_test() {
  local name="$1"
  local cr_file="$2"
  local crd_file="$3"
  local expected_exit="$4"
  local expected_valid="$5"

  local output
  output=$("$BINARY" -f "$cr_file" -crd "$crd_file" 2>&1) || true

  local actual_exit=0
  if ! echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if (d['summary']['invalid'] + d['summary']['errors']) == 0 else 1)" 2>/dev/null; then
    actual_exit=1
  fi

  local actual_valid
  actual_valid=$(echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['summary']['valid'])" 2>/dev/null || echo "-1")

  if [[ "$actual_exit" == "$expected_exit" ]] && [[ "$actual_valid" == "$expected_valid" ]]; then
    echo "  PASS  $name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  $name — expected exit=$expected_exit valid=$expected_valid, got exit=$actual_exit valid=$actual_valid"
    FAILED=$((FAILED+1))
  fi
}

FIXTURES_DIR="tests/fixtures"
CRD_DIR="$FIXTURES_DIR/crd"

# ---------------------------------------------------------------------------
# Comprehensive CRD (example.com/v1) — string validations
# ---------------------------------------------------------------------------
run_fixture_test "valid-string-all-formats" \
  "$FIXTURES_DIR/cr/valid-string-all-formats.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-string-min-length" \
  "$FIXTURES_DIR/cr/invalid-string-min-length.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-string-max-length" \
  "$FIXTURES_DIR/cr/invalid-string-max-length.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-string-pattern" \
  "$FIXTURES_DIR/cr/invalid-string-pattern.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Integer / number validations
# ---------------------------------------------------------------------------
run_fixture_test "valid-int-in-range" \
  "$FIXTURES_DIR/cr/valid-int-in-range.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-int-too-low" \
  "$FIXTURES_DIR/cr/invalid-int-too-low.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-int-too-high" \
  "$FIXTURES_DIR/cr/invalid-int-too-high.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Boolean
# ---------------------------------------------------------------------------
run_fixture_test "valid-bool-true" \
  "$FIXTURES_DIR/cr/valid-bool-true.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-bool-false" \
  "$FIXTURES_DIR/cr/valid-bool-false.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-bool-as-string" \
  "$FIXTURES_DIR/cr/invalid-bool-as-string.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Enum
# ---------------------------------------------------------------------------
run_fixture_test "valid-enum-value" \
  "$FIXTURES_DIR/cr/valid-enum-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-enum-value" \
  "$FIXTURES_DIR/cr/invalid-enum-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Required fields
# ---------------------------------------------------------------------------
run_fixture_test "valid-required-field" \
  "$FIXTURES_DIR/cr/valid-required-field.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-required-missing" \
  "$FIXTURES_DIR/cr/invalid-required-missing.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Array constraints
# ---------------------------------------------------------------------------
run_fixture_test "valid-array-min-items" \
  "$FIXTURES_DIR/cr/valid-array-min-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-array-min-items" \
  "$FIXTURES_DIR/cr/invalid-array-min-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-array-max-items" \
  "$FIXTURES_DIR/cr/valid-array-max-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-array-max-items" \
  "$FIXTURES_DIR/cr/invalid-array-max-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Object constraints
# ---------------------------------------------------------------------------
run_fixture_test "valid-obj-min-props" \
  "$FIXTURES_DIR/cr/valid-obj-min-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-obj-min-props" \
  "$FIXTURES_DIR/cr/invalid-obj-min-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-obj-max-props" \
  "$FIXTURES_DIR/cr/valid-obj-max-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-obj-max-props" \
  "$FIXTURES_DIR/cr/invalid-obj-max-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# oneOf
# ---------------------------------------------------------------------------
run_fixture_test "valid-oneof-choice" \
  "$FIXTURES_DIR/cr/valid-oneof-choice.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-oneof-none" \
  "$FIXTURES_DIR/cr/invalid-oneof-none.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# allOf / anyOf / not
# ---------------------------------------------------------------------------
run_fixture_test "valid-logical-allof" \
  "$FIXTURES_DIR/cr/valid-logical-allof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-logical-allof" \
  "$FIXTURES_DIR/cr/invalid-logical-allof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-logical-anyof" \
  "$FIXTURES_DIR/cr/valid-logical-anyof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-logical-anyof" \
  "$FIXTURES_DIR/cr/invalid-logical-anyof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-logical-not" \
  "$FIXTURES_DIR/cr/valid-logical-not.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-logical-not" \
  "$FIXTURES_DIR/cr/invalid-logical-not.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# x-kubernetes-int-or-string
# ---------------------------------------------------------------------------
run_fixture_test "valid-int-or-string-int" \
  "$FIXTURES_DIR/cr/valid-int-or-string-int.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-int-or-string-string" \
  "$FIXTURES_DIR/cr/valid-int-or-string-string.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

# ---------------------------------------------------------------------------
# x-kubernetes-embedded-resource
# ---------------------------------------------------------------------------
run_fixture_test "valid-embedded-resource" \
  "$FIXTURES_DIR/cr/valid-embedded-resource.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

# ---------------------------------------------------------------------------
# nullable
# ---------------------------------------------------------------------------
run_fixture_test "valid-nullable-null" \
  "$FIXTURES_DIR/cr/valid-nullable-null.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-nullable-value" \
  "$FIXTURES_DIR/cr/valid-nullable-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

# ---------------------------------------------------------------------------
# preserve-unknown-fields
# ---------------------------------------------------------------------------
run_fixture_test "valid-preserve-unknown" \
  "$FIXTURES_DIR/cr/valid-preserve-unknown.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

# ---------------------------------------------------------------------------
# Database CRD (existing tests)
# ---------------------------------------------------------------------------
run_fixture_test "valid-database" \
  "$FIXTURES_DIR/cr/valid-database.yaml" \
  "$CRD_DIR/database-crd.yaml" \
  0 1

run_fixture_test "invalid-database-missing-version" \
  "$FIXTURES_DIR/cr/invalid-database-missing-version.yaml" \
  "$CRD_DIR/database-crd.yaml" \
  1 0

run_fixture_test "invalid-database-wrong-type" \
  "$FIXTURES_DIR/cr/invalid-database-wrong-type.yaml" \
  "$CRD_DIR/database-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Versioned CRD
# ---------------------------------------------------------------------------
run_fixture_test "valid-versioned-v1" \
  "$FIXTURES_DIR/cr/valid-versioned-v1.yaml" \
  "$CRD_DIR/versioned-crd.yaml" \
  0 1

run_fixture_test "valid-versioned-v1alpha1" \
  "$FIXTURES_DIR/cr/valid-versioned-v1alpha1.yaml" \
  "$CRD_DIR/versioned-crd.yaml" \
  0 1

run_fixture_test "invalid-versioned-v1-missing-v1field" \
  "$FIXTURES_DIR/cr/invalid-versioned-v1-missing-v1field.yaml" \
  "$CRD_DIR/versioned-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "Fixture test results: $PASSED passed, $FAILED failed"
if [[ "$FAILED" -gt 0 ]]; then
  echo "INTEGRATION TESTS FAILED"
  exit 1
fi

echo ""
echo "All tests passed"
exit 0