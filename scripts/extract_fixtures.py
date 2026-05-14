#!/usr/bin/env python3
"""Parse Go test files and extract CRD definitions and test cases into YAML fixtures."""

import re
import os
import sys
from pathlib import Path

WORKSPACE = Path("/home/node/.openclaw/workspace/github/SWE/swe-k8-manifest-validator/workspace")
FIXTURES = WORKSPACE / "tests/fixtures/realworld"
CRD_DIR = WORKSPACE / "tests/fixtures/crd"

OPERATORS = ["strimzi", "prometheus", "argocd", "istio", "redis", "postgresql", "flux", "cert-manager"]

GO_FILES = {
    "strimzi": WORKSPACE / "pkg/validator/phase4_realworld_strimzi_test.go",
    "prometheus": WORKSPACE / "pkg/validator/phase4_realworld_prometheus_test.go",
    "argocd": WORKSPACE / "pkg/validator/phase4_realworld_argocd_test.go",
    "istio": WORKSPACE / "pkg/validator/phase4_realworld_istio_test.go",
    "redis": WORKSPACE / "pkg/validator/phase4_realworld_redis_test.go",
    "postgresql": WORKSPACE / "pkg/validator/phase4_realworld_postgresql_test.go",
    "flux": WORKSPACE / "pkg/validator/phase4_realworld_flux_test.go",
    "cert-manager": WORKSPACE / "pkg/validator/phase4_realworld_certmanager_test.go",
}

def extract_crd_definitions(content, operator):
    crds = {}
    pattern = r'const\s+(\w+CRD)\s*=\s*`([^`]+)`'
    for match in re.finditer(pattern, content, re.DOTALL):
        name = match.group(1)
        yaml_content = match.group(2)
        crds[name] = yaml_content
    return crds

def extract_test_cases(content, operator):
    cases = []
    case_pattern = r'\{\s*name:\s*"([^"]+)"\s*,\s*crYAML:\s*\[\]byte\(`([^`]*)`\)(?:\s*,)?\s*(?:wantErr:\s*(true|false))?(?:\s*,)?'
    for match in re.finditer(case_pattern, content, re.DOTALL):
        name = match.group(1)
        yaml_content = match.group(2)
        want_err = match.group(3) == "true" if match.group(3) else False
        cases.append({"name": name, "yaml": yaml_content, "wantErr": want_err})
    return cases

def extract_malformed_cases(content):
    cases = []
    
    # Pattern for string form with double quotes
    string_pattern = r'\{\s*name:\s*"([^"]+)"\s*,\s*(yaml|doc|json):\s*\[\]byte\("((?:[^"\\]|\\.)*)"\)\s*,\s*wantErr:\s*(true|false)'
    for match in re.finditer(string_pattern, content, re.DOTALL):
        name = match.group(1)
        yaml_content = match.group(3).replace('\\n', '\n').replace('\\t', '\t').replace('\\"', '"').replace('\\\\', '\\')
        want_err = match.group(4) == "true"
        cases.append({"name": name, "yaml": yaml_content, "wantErr": want_err})
    
    # Handle backtick raw strings with string manipulation
    # Find all json: []byte(`...`) blocks manually
    json_backtick_start = 'json:    []byte(' + chr(96)  # backtick
    json_backtick_end = '`)' + chr(44) + chr(32) + chr(119) + 'antErr: '  # `), wantErr:
    
    idx = 0
    while True:
        start = content.find(json_backtick_start, idx)
        if start < 0:
            break
        # Find the name before this
        name_start = content.rfind('name:    "', 0, start)
        if name_start < 0:
            break
        name_end = content.find('"', name_start + 9)
        if name_end < 0 or name_end > start:
            break
        name = content[name_start+9:name_end]
        
        # Find the backtick content
        content_start = start + len(json_backtick_start)
        content_end = content.find(chr(96), content_start)
        if content_end < 0:
            break
        yaml_content = content[content_start:content_end]
        
        # Find wantErr after the closing `
        after_backtick = content[content_end+2:content_end+50]
        wanterr_match = re.search(r'wantErr:\s*(true|false)', after_backtick)
        if wanterr_match:
            want_err = wanterr_match.group(1) == "true"
            cases.append({"name": name, "yaml": yaml_content, "wantErr": want_err})
        
        idx = content_end + 1
    
    # Pattern for empty byte slice: []byte{}
    empty_pattern = r'\{\s*name:\s*"([^"]+)"\s*,\s*(yaml|doc|json):\s*\[\]byte\{\}\s*,\s*wantErr:\s*(true|false)'
    for match in re.finditer(empty_pattern, content):
        name = match.group(1)
        want_err = match.group(3) == "true"
        cases.append({"name": name, "yaml": "", "wantErr": want_err})
    
    return cases

def save_crd_files(crds, operator):
    crd_dir = CRD_DIR
    crd_dir.mkdir(parents=True, exist_ok=True)
    
    mapping = {
        "strimzi": {
            "strimziKafkaCRD": "strimzi-kafka-crd.yaml",
            "strimziKafkaTopicCRD": "strimzi-topics-crd.yaml",
            "strimziKafkaUserCRD": "strimzi-users-crd.yaml",
            "strimziKafkaConnectCRD": "strimzi-connect-crd.yaml",
            "strimziKafkaMirrorMaker2CRD": "strimzi-mirrormaker2-crd.yaml",
        },
        "prometheus": {
            "prometheusCRD": "prometheus-crd.yaml",
            "prometheusRuleCRD": "prometheusrule-crd.yaml",
            "serviceMonitorCRD": "servicemonitor-crd.yaml",
            "podMonitorCRD": "podmonitor-crd.yaml",
            "probeCRD": "probe-crd.yaml",
        },
        "argocd": {
            "argocdCRD": "argocd-crd.yaml",
            "applicationCRD": "argocd-application-crd.yaml",
            "appProjectCRD": "argocd-appproject-crd.yaml",
            "argocdNotificationCRD": "argocd-notification-crd.yaml",
        },
        "istio": {
            "virtualServiceCRD": "istio-virtualservice-crd.yaml",
            "destinationRuleCRD": "istio-destinationrule-crd.yaml",
            "gatewayCRD": "istio-gateway-crd.yaml",
            "serviceEntryCRD": "istio-serviceentry-crd.yaml",
            "sidecarCRD": "istio-sidecar-crd.yaml",
        },
        "redis": {
            "redisCRD": "redis-crd.yaml",
            "redisClusterCRD": "redis-cluster-crd.yaml",
            "redisSentinelCRD": "redis-sentinel-crd.yaml",
        },
        "postgresql": {
            "pgclusterCRD": "postgresql-crd.yaml",
            "pgreplicaCRD": "postgresql-replica-crd.yaml",
            "pgbackupCRD": "postgresql-backup-crd.yaml",
            "pgtaskCRD": "postgresql-task-crd.yaml",
        },
        "flux": {
            "gitRepositoryCRD": "flux-gitrepository-crd.yaml",
            "helmRepositoryCRD": "flux-helmrepository-crd.yaml",
            "helmReleaseCRD": "flux-helmrelease-crd.yaml",
            "kustomizationCRD": "flux-kustomization-crd.yaml",
            "fluxInstallCRD": "flux-install-crd.yaml",
        },
        "cert-manager": {
            "certificateCRD": "cert-manager-crd.yaml",
            "issuerCRD": "cert-manager-issuer-crd.yaml",
            "clusterIssuerCRD": "cert-manager-clusterissuer-crd.yaml",
            "certificateRequestCRD": "cert-manager-certificaterequest-crd.yaml",
        },
    }
    
    name_mapping = mapping.get(operator, {})
    for crd_name, crd_content in crds.items():
        filename = name_mapping.get(crd_name, f"{operator}-{crd_name.lower()}.yaml")
        file_path = crd_dir / filename
        file_path.write_text(crd_content.strip())
        print(f"    Wrote CRD: {filename}")

def sanitize_filename(name):
    safe_name = name.replace(" ", "-").replace(".", "-").replace(",", "").replace("/", "-").replace("\n", "")
    safe_name = re.sub(r'[^a-zA-Z0-9_\-.>]', '-', safe_name)
    safe_name = safe_name[:200]
    return safe_name

def main():
    CRD_DIR.mkdir(parents=True, exist_ok=True)
    
    for operator, go_file in GO_FILES.items():
        print(f"Processing {operator}...")
        content = go_file.read_text()
        
        crds = extract_crd_definitions(content, operator)
        print(f"  Found {len(crds)} CRD definitions")
        save_crd_files(crds, operator)
        
        cases = extract_test_cases(content, operator)
        print(f"  Found {len(cases)} test cases")
        
        valid_cases = [c for c in cases if not c["wantErr"]]
        invalid_cases = [c for c in cases if c["wantErr"]]
        print(f"    Valid: {len(valid_cases)}, Invalid: {len(invalid_cases)}")
        
        valid_dir = FIXTURES / operator / "valid"
        valid_dir.mkdir(parents=True, exist_ok=True)
        for i, case in enumerate(valid_cases, 1):
            filename = f"case-{i:03d}-{sanitize_filename(case['name'])}.yaml"
            (valid_dir / filename).write_text(case["yaml"].strip())
        
        invalid_dir = FIXTURES / operator / "invalid"
        invalid_dir.mkdir(parents=True, exist_ok=True)
        for i, case in enumerate(invalid_cases, 1):
            filename = f"case-{i:03d}-{sanitize_filename(case['name'])}.yaml"
            (invalid_dir / filename).write_text(case["yaml"].strip())
        
        print(f"  Wrote {len(valid_cases)} valid + {len(invalid_cases)} invalid YAML files")
    
    print("\nProcessing malformed...")
    malformed_content = (WORKSPACE / "pkg/validator/phase4_realworld_malformed_test.go").read_text()
    malformed_cases = extract_malformed_cases(malformed_content)
    print(f"  Found {len(malformed_cases)} malformed test cases")
    
    malformed_dir = FIXTURES / "malformed"
    malformed_dir.mkdir(parents=True, exist_ok=True)
    
    for i, case in enumerate(malformed_cases, 1):
        filename = f"case-{i:03d}-{sanitize_filename(case['name'])}.yaml"
        (malformed_dir / filename).write_text(case["yaml"])
    
    print(f"  Wrote {len(malformed_cases)} malformed YAML files")
    print("\nDone! All fixtures extracted.")

if __name__ == "__main__":
    main()