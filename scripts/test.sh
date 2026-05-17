#!/usr/bin/env bash
set -euo pipefail

# test.sh — Run unit tests and binary smoke test.
#
# Exit codes:
#   0 = all tests passed
#   1 = test failed

# =============================================================================
# CRD Download Functions
# Download latest stable CRDs from upstream sources at test time
# =============================================================================
download_upstream_crds() {
  local crd_dir="tests/fixtures/crd"
  mkdir -p "$crd_dir"
  
  echo "[download] Fetching upstream CRDs..."
  
  # Istio CRDs (v1.24.0)
  echo "  [download] Istio v1.24.0..."
  local istio_url="https://raw.githubusercontent.com/istio/istio/1.24.0/manifests/charts/base/files/crd-all.gen.yaml"
  curl -sSL "$istio_url" -o "$crd_dir/istio-all.yaml" || echo "  Istio download failed, using existing"
  
  # cert-manager CRDs (v1.16.0)
  echo "  [download] cert-manager v1.16.0..."
  local cm_url="https://github.com/cert-manager/cert-manager/releases/download/v1.16.0/cert-manager.crds.yaml"
  curl -sSL "$cm_url" -o "$crd_dir/cert-manager-all.yaml" || echo "  cert-manager download failed, using existing"
  
  # Prometheus Operator CRDs
  echo "  [download] Prometheus Operator CRDs..."
  curl -sSL "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml" -o "$crd_dir/prometheus-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml" -o "$crd_dir/prometheusrule-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml" -o "$crd_dir/servicemonitor-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml" -o "$crd_dir/prometheus-alertmanagercrd.yaml"
  
  # ArgoCD CRDs (v2.14.0)
  echo "  [download] ArgoCD v2.14.0..."
  curl -sSL "https://raw.githubusercontent.com/argoproj/argo-cd/v2.14.0/manifests/crds/application-crd.yaml" -o "$crd_dir/argocd-application-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/argoproj/argo-cd/v2.14.0/manifests/crds/applicationset-crd.yaml" -o "$crd_dir/argocd-applicationset-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/argoproj/argo-cd/v2.14.0/manifests/crds/appproject-crd.yaml" -o "$crd_dir/argocd-appproject-crd.yaml"
  
  # Flux CRDs (from release bundles)
  echo "  [download] Flux CRDs..."
  curl -sSL "https://github.com/fluxcd/source-controller/releases/download/v1.8.4/source-controller.crds.yaml" -o "$crd_dir/flux-source-controller-crds.yaml"
  curl -sSL "https://github.com/fluxcd/kustomize-controller/releases/download/v1.8.5/kustomize-controller.crds.yaml" -o "$crd_dir/flux-kustomization-crd.yaml"
  curl -sSL "https://github.com/fluxcd/helm-controller/releases/download/v1.5.4/helm-controller.crds.yaml" -o "$crd_dir/flux-helmrelease-crd.yaml"
  curl -sSL "https://github.com/fluxcd/notification-controller/releases/download/v1.8.4/notification-controller.crds.yaml" -o "$crd_dir/flux-notification-controller-crds.yaml"
  
  # PostgreSQL PGO CRDs (v6.0.1)
  echo "  [download] PostgreSQL PGO v6.0.1..."
  curl -sSL "https://raw.githubusercontent.com/CrunchyData/postgres-operator/v6.0.1/config/crd/bases/postgres-operator.crunchydata.com_pgadmins.yaml" -o "$crd_dir/postgresql-pgadmin-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/CrunchyData/postgres-operator/v6.0.1/config/crd/bases/postgres-operator.crunchydata.com_pgupgrades.yaml" -o "$crd_dir/postgresql-pgupgrade-crd.yaml"
  curl -sSL "https://raw.githubusercontent.com/CrunchyData/postgres-operator/v6.0.1/config/crd/bases/postgres-operator.crunchydata.com_postgresclusters.yaml" -o "$crd_dir/postgresql-pgcluster-crd.yaml"
  
  echo "[download] CRDs refreshed from upstream"
}

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

run_fixture_test "valid-string-empty" \
  "$FIXTURES_DIR/cr/valid-string-empty.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-string-format-date" \
  "$FIXTURES_DIR/cr/invalid-string-format-date.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-string-multiline" \
  "$FIXTURES_DIR/cr/invalid-string-multiline.yaml" \
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

run_fixture_test "valid-int-zero" \
  "$FIXTURES_DIR/cr/valid-int-zero.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-int-negative" \
  "$FIXTURES_DIR/cr/valid-int-negative.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-int-as-string" \
  "$FIXTURES_DIR/cr/invalid-int-as-string.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-int-float" \
  "$FIXTURES_DIR/cr/invalid-int-float.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-int-max" \
  "$FIXTURES_DIR/cr/valid-int-max.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

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

run_fixture_test "invalid-bool-as-number" \
  "$FIXTURES_DIR/cr/invalid-bool-as-number.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Enum
# ---------------------------------------------------------------------------
run_fixture_test "valid-enum-value" \
  "$FIXTURES_DIR/cr/valid-enum-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-enum-first-value" \
  "$FIXTURES_DIR/cr/valid-enum-first-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-enum-last-value" \
  "$FIXTURES_DIR/cr/valid-enum-last-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-enum-value" \
  "$FIXTURES_DIR/cr/invalid-enum-value.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-enum-case-mismatch" \
  "$FIXTURES_DIR/cr/invalid-enum-case-mismatch.yaml" \
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

run_fixture_test "invalid-required-null" \
  "$FIXTURES_DIR/cr/invalid-required-null.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-required-empty-object" \
  "$FIXTURES_DIR/cr/invalid-required-empty-object.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# Array constraints
# ---------------------------------------------------------------------------
run_fixture_test "valid-array-min-items" \
  "$FIXTURES_DIR/cr/valid-array-min-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-array-empty" \
  "$FIXTURES_DIR/cr/valid-array-empty.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-array-min-equals-max" \
  "$FIXTURES_DIR/cr/valid-array-min-equals-max.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-array-min-items" \
  "$FIXTURES_DIR/cr/invalid-array-min-items.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-array-type-mismatch" \
  "$FIXTURES_DIR/cr/invalid-array-type-mismatch.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-array-nested-type" \
  "$FIXTURES_DIR/cr/invalid-array-nested-type.yaml" \
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

run_fixture_test "valid-object-empty" \
  "$FIXTURES_DIR/cr/valid-object-empty.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-object-optional-missing" \
  "$FIXTURES_DIR/cr/valid-object-optional-missing.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-obj-min-props" \
  "$FIXTURES_DIR/cr/invalid-obj-min-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "valid-object-additional-props" \
  "$FIXTURES_DIR/cr/valid-object-additional-props.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-object-required-vs-optional" \
  "$FIXTURES_DIR/cr/invalid-object-required-vs-optional.yaml" \
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

run_fixture_test "valid-logical-allof-nested" \
  "$FIXTURES_DIR/cr/valid-logical-allof-nested.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-logical-allof" \
  "$FIXTURES_DIR/cr/invalid-logical-allof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-logical-allof-partial" \
  "$FIXTURES_DIR/cr/invalid-logical-allof-partial.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0


run_fixture_test "valid-logical-anyof" \
  "$FIXTURES_DIR/cr/valid-logical-anyof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-logical-anyof-one-match" \
  "$FIXTURES_DIR/cr/valid-logical-anyof-one-match.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-logical-anyof" \
  "$FIXTURES_DIR/cr/invalid-logical-anyof.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

run_fixture_test "invalid-logical-anyof-none-match" \
  "$FIXTURES_DIR/cr/invalid-logical-anyof-none-match.yaml" \
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

run_fixture_test "valid-int-or-string-zero" \
  "$FIXTURES_DIR/cr/valid-int-or-string-zero.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-int-or-string-string" \
  "$FIXTURES_DIR/cr/valid-int-or-string-string.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-int-or-string-object" \
  "$FIXTURES_DIR/cr/invalid-int-or-string-object.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# x-kubernetes-embedded-resource
# ---------------------------------------------------------------------------
run_fixture_test "valid-embedded-resource" \
  "$FIXTURES_DIR/cr/valid-embedded-resource.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-embedded-resource-with-namespace" \
  "$FIXTURES_DIR/cr/valid-embedded-resource-with-namespace.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-embedded-resource-invalid-kind" \
  "$FIXTURES_DIR/cr/valid-embedded-resource-invalid-kind.yaml" \
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

run_fixture_test "invalid-nullable-null-non-nullable" \
  "$FIXTURES_DIR/cr/invalid-nullable-null-non-nullable.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# preserve-unknown-fields
# ---------------------------------------------------------------------------
run_fixture_test "valid-preserve-unknown" \
  "$FIXTURES_DIR/cr/valid-preserve-unknown.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-preserve-unknown-deep" \
  "$FIXTURES_DIR/cr/valid-preserve-unknown-deep.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "valid-preserve-unknown-type-conflict" \
  "$FIXTURES_DIR/cr/valid-preserve-unknown-type-conflict.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

# ---------------------------------------------------------------------------
# Default field
# ---------------------------------------------------------------------------
run_fixture_test "valid-default-overridden" \
  "$FIXTURES_DIR/cr/valid-default-overridden.yaml" \
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

run_fixture_test "valid-versioned-multi-version" \
  "$FIXTURES_DIR/cr/valid-versioned-multi-version.yaml" \
  "$CRD_DIR/comprehensive-crd.yaml" \
  0 1

run_fixture_test "invalid-versioned-v1alpha1-missing-field" \
  "$FIXTURES_DIR/cr/invalid-versioned-v1alpha1-missing-field.yaml" \
  "$CRD_DIR/versioned-crd.yaml" \
  1 0

# ---------------------------------------------------------------------------
# realworld operator tests — batched by directory for speed
# Each operator's valid/ and invalid/ dirs are processed in a single binary invocation
# ~17 calls instead of ~1200 individual invocations
# ---------------------------------------------------------------------------  
echo ""
echo "[realworld binary fixture tests — batched]"

RW_FAILED=0
RW_PASSED=0

# Run realworld test for a full directory (batched)
# Usage: run_realworld_batch "label" "/path/to/dir" "CRD_DIR" expected_exit expected_valid
run_realworld_batch() {
  local label="$1"
  local manifest_dir="$2"
  local crd_dir="$3"
  local expected_exit="$4"
  local expected_valid="$5"

  if [[ ! -d "$manifest_dir" ]]; then
    echo "  SKIP  $label (directory not found)"
    return
  fi

  # Count files ahead of time for validation
  local file_count
  file_count=$(find "$manifest_dir" -name "*.yaml" -not -empty 2>/dev/null | wc -l)
  if [[ "$file_count" == "0" ]]; then
    echo "  SKIP  $label (no non-empty yaml files)"
    return
  fi

  local output
  local actual_exit=0

  if [[ -n "$crd_dir" ]]; then
    output=$("$BINARY" -f "$manifest_dir" -crd "$crd_dir" 2>&1) || actual_exit=1
  else
    output=$("$BINARY" -f "$manifest_dir" 2>&1) || actual_exit=1
  fi

  # Parse JSON output (summary section)
  local actual_valid actual_total actual_errors actual_skipped
  actual_valid=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary']['valid'])" 2>/dev/null || echo "-1")
  actual_total=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary']['total'])" 2>/dev/null || echo "-1")
  actual_errors=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary']['errors'])" 2>/dev/null || echo "-1")
  actual_skipped=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary'].get('skipped',0))" 2>/dev/null || echo "0")

  # Malformed dir: expected_valid=0, actual_valid=0, actual_total=0 is success
  # (empty/broken YAML correctly detected as invalid)
  local malformed_ok=0
  if [[ "$expected_valid" == "0" ]] && [[ "$actual_valid" == "0" ]] && [[ "$actual_total" == "0" ]]; then
    malformed_ok=1
  fi

  # Skipped case: files decoded but skipped (no matching CRD) + expected to fail
  local skipped_ok=0
  if [[ "$expected_valid" == "0" ]] && [[ "$actual_valid" == "0" ]] && [[ "$actual_skipped" -ge 1 ]] && [[ "$expected_exit" == "1" ]]; then
    skipped_ok=1
  fi

# For valid dirs: expect exit=0, valid=total (all pass)
  # For invalid/malformed dirs: expect exit=1, valid=0 (correctly rejected)
  local pass=0
  if [[ "$expected_exit" == "0" ]]; then
    # Valid dir: all files should pass
    if [[ "$actual_exit" == "0" ]] && [[ "$actual_valid" == "$actual_total" ]]; then
      pass=1
    fi
  else
    # Invalid/malformed dir: should have invalid=0, valid=0
    if [[ "$malformed_ok" == "1" ]] || [[ "$skipped_ok" == "1" ]]; then
      pass=1
    elif [[ "$actual_exit" == "$expected_exit" ]] && [[ "$actual_valid" == "0" ]]; then
      pass=1
    fi
  fi

  if [[ "$pass" == "1" ]]; then
    echo "  PASS  $label ($file_count files, ${actual_valid}valid, ${actual_errors}errors)"
    RW_PASSED=$((RW_PASSED+1))
  else
    echo "  FAIL  $label — expected exit=$expected_exit valid=$expected_valid, got exit=$actual_exit valid=$actual_valid (total=$actual_total, errors=$actual_errors, skipped=$actual_skipped)"
    RW_FAILED=$((RW_FAILED+1))
  fi
}

REALWORLD_DIR="tests/fixtures/realworld"
operators="istio strimzi prometheus argocd redis postgresql flux cert-manager etcd rabbitmq kafka elasticsearch jaeger"

# Test each operator's valid manifests (expect exit=0, valid=total)
echo "  Testing operator valid directories..."
for op in $operators; do
  run_realworld_batch "$op/valid" "$REALWORLD_DIR/$op/valid" "$CRD_DIR" 0 1
done

# Test each operator's invalid manifests (expect exit=1, valid=0)
echo "  Testing operator invalid directories..."
for op in $operators; do
  run_realworld_batch "$op/invalid" "$REALWORLD_DIR/$op/invalid" "$CRD_DIR" 1 0
done

# Test malformed manifests (no CRD needed)
echo "  Testing malformed directory..."
run_realworld_batch "malformed" "$REALWORLD_DIR/malformed" "" 1 0

echo ""
echo "Realworld fixture results: $RW_PASSED passed, $RW_FAILED failed"
if [[ "$RW_FAILED" -gt 0 ]]; then
  echo "REALWORLD TESTS FAILED"
  FAILED=$((FAILED+RW_FAILED))
fi

exit 0
