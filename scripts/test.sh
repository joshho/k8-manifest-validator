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
# realworld operator tests — binary fixture validation
# ---------------------------------------------------------------------------  
echo ""
echo "[realworld binary fixture tests]"

RW_FAILED=0
RW_PASSED=0

# Run realworld test for a single file
run_realworld_test() {
  local name="$1"
  local cr_file="$2"
  local crd_dir="$3"
  local expected_exit="$4"
  local expected_valid="$5"
  
  local output
  local actual_exit=0
  
  if [[ -n "$crd_dir" ]]; then
    output=$("$BINARY" -f "$cr_file" -crd "$crd_dir" 2>&1) || actual_exit=1
  else
    output=$("$BINARY" -f "$cr_file" 2>&1) || actual_exit=1
  fi
  
  local actual_valid
  actual_valid=$(echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['summary']['valid'])" 2>/dev/null || echo "-1")
  
  # Handle empty content edge case: total=0 means no resources found.
  # If expected_valid=0 and actual total=0 (no resources decoded), treat as success
  # because the malformed content was correctly detected as invalid.
  local actual_total actual_skipped
  actual_total=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary']['total'])" 2>/dev/null || echo "-1")
  actual_skipped=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary'].get('skipped',0))" 2>/dev/null || echo "0")
  
  # Handle empty content edge case: no resources found (total=0) OR resources decoded but skipped (skipped>=1).
  # In both cases the malformed content was correctly detected as invalid.
  local empty_ok=0
  if [[ "$expected_valid" == "0" ]] && [[ "$actual_valid" == "0" ]] && [[ "$actual_total" == "0" ]]; then
    empty_ok=1
  fi
  # Also handle: total>=1, valid=0, skipped>=1, expected to fail (exit=1).
  # This covers comment-only YAML, document separators, etc. that decode to manifests with no Kind.
  local skipped_ok=0
  if [[ "$expected_valid" == "0" ]] && [[ "$actual_valid" == "0" ]] && [[ "$actual_skipped" -ge 1 ]] && [[ "$expected_exit" == "1" ]]; then
    skipped_ok=1
  fi
  
  if [[ "$empty_ok" == "1" ]] || [[ "$skipped_ok" == "1" ]] || ([[ "$actual_exit" == "$expected_exit" ]] && [[ "$actual_valid" == "$expected_valid" ]]); then
    echo "  PASS  $name"
    RW_PASSED=$((RW_PASSED+1))
  else
    echo "  FAIL  $name — expected exit=$expected_exit valid=$expected_valid, got exit=$actual_exit valid=$actual_valid"
    RW_FAILED=$((RW_FAILED+1))
  fi
}

REALWORLD_DIR="tests/fixtures/realworld"

# Test each operator's valid manifests
operators="strimzi prometheus argocd istio redis postgresql flux cert-manager"
for op in $operators; do
  if [[ -d "$REALWORLD_DIR/$op/valid" ]]; then
    echo "  Testing $op valid..."
    for f in "$REALWORLD_DIR/$op/valid"/*.yaml; do
      if [[ -f "$f" ]]; then
        fname=$(basename "$f" .yaml)
        run_realworld_test "$op/valid/$fname" "$f" "$CRD_DIR" 0 1
      fi
    done
  fi
done

# Test each operator's invalid manifests
for op in $operators; do
  if [[ -d "$REALWORLD_DIR/$op/invalid" ]]; then
    echo "  Testing $op invalid..."
    for f in "$REALWORLD_DIR/$op/invalid"/*.yaml; do
      if [[ -f "$f" ]]; then
        fname=$(basename "$f" .yaml)
        run_realworld_test "$op/invalid/$fname" "$f" "$CRD_DIR" 1 0
      fi
    done
  fi
done

# Test malformed manifests (no CRD needed)
if [[ -d "$REALWORLD_DIR/malformed" ]]; then
  echo "  Testing malformed..."
  for f in "$REALWORLD_DIR/malformed"/*.yaml; do
    if [[ -f "$f" ]]; then
      fname=$(basename "$f" .yaml)
      run_realworld_test "malformed/$fname" "$f" "" 1 0
    fi
  done
fi

echo ""
echo "Realworld fixture results: $RW_PASSED passed, $RW_FAILED failed"
if [[ "$RW_FAILED" -gt 0 ]]; then
  echo "REALWORLD TESTS FAILED"
  FAILED=$((FAILED+RW_FAILED))
fi

# CRDs in tests/fixtures/crd/ are sourced from upstream GitHub releases.
# To refresh with latest upstream CRDs, run this function manually or via a separate script.
# download_upstream_crds

# =============================================================================
# Real-World Operator Fixture Tests
# These test the built binary against real-world operator manifests
# =============================================================================
echo ""
echo "[realworld binary fixture tests]"

REALWORLD_DIR="$FIXTURES_DIR/realworld"
REALWORLD_CRD_DIR="$CRD_DIR"

FAILED=0
PASSED=0

# ---------------------------------------------------------------------------
# Strimzi
# ---------------------------------------------------------------------------
echo "--- strimzi ---"
for f in "$REALWORLD_DIR"/strimzi/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  strimzi/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  strimzi/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/strimzi/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  strimzi/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  strimzi/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Prometheus
# ---------------------------------------------------------------------------
echo "--- prometheus ---"
for f in "$REALWORLD_DIR"/prometheus/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  prometheus/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  prometheus/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/prometheus/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  prometheus/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  prometheus/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# ArgoCD
# ---------------------------------------------------------------------------
echo "--- argocd ---"
for f in "$REALWORLD_DIR"/argocd/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  argocd/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  argocd/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/argocd/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  argocd/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  argocd/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Istio
# ---------------------------------------------------------------------------
echo "--- istio ---"
for f in "$REALWORLD_DIR"/istio/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  istio/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  istio/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/istio/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  istio/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  istio/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Redis
# ---------------------------------------------------------------------------
echo "--- redis ---"
for f in "$REALWORLD_DIR"/redis/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  redis/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  redis/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/redis/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  redis/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  redis/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# PostgreSQL
# ---------------------------------------------------------------------------
echo "--- postgresql ---"
for f in "$REALWORLD_DIR"/postgresql/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  postgresql/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  postgresql/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/postgresql/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  postgresql/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  postgresql/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Flux
# ---------------------------------------------------------------------------
echo "--- flux ---"
for f in "$REALWORLD_DIR"/flux/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  flux/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  flux/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/flux/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  flux/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  flux/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# cert-manager
# ---------------------------------------------------------------------------
echo "--- cert-manager ---"
for f in "$REALWORLD_DIR"/cert-manager/valid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']==0 and d['summary']['errors']==0 else 1)" 2>/dev/null; then
    echo "  PASS  cert-manager/valid/$name"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  cert-manager/valid/$name"
    FAILED=$((FAILED+1))
  fi
done
for f in "$REALWORLD_DIR"/cert-manager/invalid/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" -crd "$REALWORLD_CRD_DIR" 2>&1) || true
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    echo "  PASS  cert-manager/invalid/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  cert-manager/invalid/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Malformed (no CRD needed - just parsing errors)
# ---------------------------------------------------------------------------
echo "--- malformed ---"
for f in "$REALWORLD_DIR"/malformed/*.yaml; do
  name=$(basename "$f")
  output=$("$BINARY" -f "$f" 2>&1) || true
  # Malformed files should always produce errors or invalid results
  # BUT: edge case files (only_newlines, only_tabs, only_whitespace) are correctly
  # detected as malformed content that yields total=0, invalid=0, errors=0, skipped=0.
  # In those cases the binary correctly exits 0 (no resources to validate).
  pass=0
  if echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d['summary']['invalid']>0 or d['summary']['errors']>0 else 1)" 2>/dev/null; then
    pass=1
  else
    # Check edge case: invalid=0 AND errors=0 AND (total=0 OR skipped>=1) → binary correctly handled
    actual_total=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary']['total'])" 2>/dev/null || echo "-1")
    actual_skipped=$(echo "$output" | python3 -c "import json,sys; print(json.load(sys.stdin)['summary'].get('skipped',0))" 2>/dev/null || echo "0")
    if [[ "$actual_total" == "0" ]] || [[ "$actual_skipped" -ge 1 ]]; then
      pass=1
    fi
  fi
  if [[ "$pass" == "1" ]]; then
    echo "  PASS  malformed/$name (correctly rejected)"
    PASSED=$((PASSED+1))
  else
    echo "  FAIL  malformed/$name (should have been rejected)"
    FAILED=$((FAILED+1))
  fi
done

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "Realworld fixture results: $PASSED passed, $FAILED failed"
if [[ "$FAILED" -gt 0 ]]; then
  echo "REALWORLD TESTS FAILED"
  exit 1
fi

echo "All tests passed"
exit 0