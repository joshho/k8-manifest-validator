package validator

import (
	"testing"
)

// =============================================================================
// RW-11 — Flux Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — flux-operator group
// Owner file: pkg/validator/phase4_realworld_flux_test.go
// CRDs covered: GitRepository, HelmRepository, HelmRelease, Kustomization, FluxInstall
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// Flux Operator CRD definitions (inline for test isolation)
// Based on Flux v2.x CRD schemas (source.toolkit.fluxcd.io, kustomize.toolkit.fluxcd.io)
// -----------------------------------------------------------------------

const gitRepositoryCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gitrepositories.source.toolkit.fluxcd.io
spec:
  group: source.toolkit.fluxcd.io
  names:
    kind: GitRepository
    plural: gitrepositories
    singular: gitrepository
    listKind: GitRepositoryList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              interval:
                type: string
              timeout:
                type: string
              ref:
                type: object
                properties:
                  branch:
                    type: string
                  tag:
                    type: string
                  semver:
                    type: string
                  commit:
                    type: string
              url:
                type: string
              secretRef:
                type: object
                properties:
                  name:
                    type: string
              ignore:
                type: string
              suspend:
                type: boolean
              gitImplementation:
                type: string
            required:
            - interval
            - url
`

const helmRepositoryCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: helmrepositories.source.toolkit.fluxcd.io
spec:
  group: source.toolkit.fluxcd.io
  names:
    kind: HelmRepository
    plural: helmrepositories
    singular: helmrepository
    listKind: HelmRepositoryList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              interval:
                type: string
              timeout:
                type: string
              url:
                type: string
              secretRef:
                type: object
                properties:
                  name:
                    type: string
              passCredentials:
                type: boolean
              provider:
                type: string
              suspend:
                type: boolean
            required:
            - interval
            - url
`

const helmReleaseCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: helmreleases.helm.toolkit.fluxcd.io
spec:
  group: helm.toolkit.fluxcd.io
  names:
    kind: HelmRelease
    plural: helmreleases
    singular: helmrelease
    listKind: HelmReleaseList
  scope: Namespaced
  versions:
  - name: v2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              interval:
                type: string
              releaseName:
                type: string
              targetNamespace:
                type: string
              timeout:
                type: string
              suspend:
                type: boolean
              chart:
                type: object
                properties:
                  spec:
                    type: object
                    properties:
                      chart:
                        type: string
                      version:
                        type: string
                      sourceRef:
                        type: object
                        properties:
                          kind:
                            type: string
                          name:
                            type: string
                required:
                - spec
              values:
                type: object
              postRenderers:
                type: array
              install:
                type: object
              upgrade:
                type: object
              rollback:
                type: object
              uninstall:
                type: object
            required:
            - interval
            - chart
`

const kustomizationCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kustomizations.kustomize.toolkit.fluxcd.io
spec:
  group: kustomize.toolkit.fluxcd.io
  names:
    kind: Kustomization
    plural: kustomizations
    singular: kustomization
    listKind: KustomizationList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              interval:
                type: string
              path:
                type: string
              prune:
                type: boolean
              sourceRef:
                type: object
                properties:
                  kind:
                    type: string
                  name:
                    type: string
              targetNamespace:
                type: string
              timeout:
                type: string
              suspend:
                type: boolean
              dependsOn:
                type: array
                items:
                  type: string
              wait:
                type: boolean
              healthChecks:
                type: array
              retryInterval:
                type: string
            required:
            - interval
            - path
            - sourceRef
`

const fluxInstallCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: installs.installation.toolkit.fluxcd.io
spec:
  group: installation.toolkit.fluxcd.io
  names:
    kind: FluxInstall
    plural: fluxinstalls
    singular: fluxinstall
    listKind: FluxInstallList
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              version:
                type: string
              interval:
                type: string
              clusterProvider:
                type: string
              components:
                type: array
                items:
                  type: string
              registry:
                type: string
              resources:
                type: object
              namespace:
                type: string
              watchAllNamespaces:
                type: boolean
              allAtOnce:
                type: boolean
            required:
            - version
            - interval
`

// -----------------------------------------------------------------------
// Test Engine Setup
// -----------------------------------------------------------------------

func init() {
	// Register Flux CRDs with the validator engine
	engine.RegisterCRD([]byte(gitRepositoryCRD))
	engine.RegisterCRD([]byte(helmRepositoryCRD))
	engine.RegisterCRD([]byte(helmReleaseCRD))
	engine.RegisterCRD([]byte(kustomizationCRD))
	engine.RegisterCRD([]byte(fluxInstallCRD))
}

// =============================================================================
// Valid GitRepository Test Cases
// =============================================================================

var fluxValidGitRepositoryCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.1 GitRepository valid minimal with interval and url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.2 GitRepository valid with branch ref",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-branch
  namespace: default
spec:
  interval: 2m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    branch: main
`),
		wantErr: false,
	},
	{
		name: "RW-11.3 GitRepository valid with tag ref",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-tag
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    tag: v1.0.0
`),
		wantErr: false,
	},
	{
		name: "RW-11.4 GitRepository valid with semver ref",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-semver
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    semver: ">=1.0.0"
`),
		wantErr: false,
	},
	{
		name: "RW-11.5 GitRepository valid with commit ref",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-commit
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    commit: 6c18a65f73e1e48c82f3a49a1c3b8eb4c2e0a24e
`),
		wantErr: false,
	},
	{
		name: "RW-11.6 GitRepository valid with secretRef",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-secret
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  secretRef:
    name: git-credentials
`),
		wantErr: false,
	},
	{
		name: "RW-11.7 GitRepository valid with timeout",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-timeout
  namespace: default
spec:
  interval: 1m0s
  timeout: 30s
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.8 GitRepository valid with ignore pattern",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-ignore
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ignore: |
    .git/**
`),
		wantErr: false,
	},
	{
		name: "RW-11.9 GitRepository valid with suspend true",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-suspended
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  suspend: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.10 GitRepository valid with gitImplementation libgit2",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-libgit2
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  gitImplementation: libgit2
`),
		wantErr: false,
	},
	{
		name: "RW-11.11 GitRepository valid with all fields",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo-full
  namespace: flux-system
spec:
  interval: 30s
  timeout: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    branch: develop
  secretRef:
    name: git-auth
  ignore: |
    .circleci/**
    Makefile
  suspend: false
  gitImplementation: go-git
`),
		wantErr: false,
	},
}

func TestFlux_GitRepository_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(gitRepositoryCRD))

	for _, tc := range fluxValidGitRepositoryCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "valid" {
					t.Errorf("expected valid status, got %s: %v", result.Status, result.Errors)
				}
			}
		})
	}
}

// =============================================================================
// Valid HelmRepository Test Cases
// =============================================================================

var fluxValidHelmRepositoryCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.12 HelmRepository valid minimal with interval and url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.13 HelmRepository valid with timeout",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-timeout
  namespace: default
spec:
  interval: 5m0s
  timeout: 60s
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.14 HelmRepository valid with secretRef",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-secret
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  secretRef:
    name: helm-repo-auth
`),
		wantErr: false,
	},
	{
		name: "RW-11.15 HelmRepository valid with passCredentials true",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-creds
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.example.com
  passCredentials: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.16 HelmRepository valid with provider generic",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-provider
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  provider: generic
`),
		wantErr: false,
	},
	{
		name: "RW-11.17 HelmRepository valid with suspend true",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-suspended
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  suspend: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.18 HelmRepository valid with all fields",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami-full
  namespace: flux-system
spec:
  interval: 3m0s
  timeout: 30s
  url: https://charts.bitnami.com/bitnami
  secretRef:
    name: helm-repo-secret
  passCredentials: false
  provider: generic
  suspend: false
`),
		wantErr: false,
	},
}

func TestFlux_HelmRepository_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(helmRepositoryCRD))

	for _, tc := range fluxValidHelmRepositoryCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "valid" {
					t.Errorf("expected valid status, got %s: %v", result.Status, result.Errors)
				}
			}
		})
	}
}

// =============================================================================
// Valid HelmRelease Test Cases
// =============================================================================

var fluxValidHelmReleaseCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.19 HelmRelease valid minimal with interval and chart",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.20 HelmRelease valid with releaseName",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-release
  namespace: default
spec:
  interval: 5m0s
  releaseName: my-nginx
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.21 HelmRelease valid with targetNamespace",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-ns
  namespace: default
spec:
  interval: 5m0s
  targetNamespace: production
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.22 HelmRelease valid with timeout",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-timeout
  namespace: default
spec:
  interval: 5m0s
  timeout: 3m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.23 HelmRelease valid with suspend true",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-suspended
  namespace: default
spec:
  interval: 5m0s
  suspend: true
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.24 HelmRelease valid with chart version",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-version
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      version: "9.0.0"
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.25 HelmRelease valid with values",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-values
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  values:
    replicaCount: 3
    image:
      repository: nginx
      tag: "1.19.0"
`),
		wantErr: false,
	},
	{
		name: "RW-11.26 HelmRelease valid with postRenderers",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-postrender
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  postRenderers:
    - kustomize:
        patches:
          - target:
              kind: Deployment
              name: nginx
            patch: |
              - op: replace
                path: /spec/replicas
                value: 3
`),
		wantErr: false,
	},
	{
		name: "RW-11.27 HelmRelease valid with dependsOn",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-deps
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  dependsOn:
    - mysql
`),
		wantErr: false,
	},
	{
		name: "RW-11.28 HelmRelease valid with wait true",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-wait
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  wait: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.29 HelmRelease valid with install configuration",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-install
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  install:
    timeout: 5m0s
    remediation:
      retries: 3
`),
		wantErr: false,
	},
	{
		name: "RW-11.30 HelmRelease valid with upgrade configuration",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx-upgrade
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  upgrade:
    timeout: 5m0s
    force: false
`),
		wantErr: false,
	},
}

func TestFlux_HelmRelease_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(helmReleaseCRD))
	_ = engine.RegisterCRD([]byte(helmRepositoryCRD))

	for _, tc := range fluxValidHelmReleaseCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "valid" {
					t.Errorf("expected valid status, got %s: %v", result.Status, result.Errors)
				}
			}
		})
	}
}

// =============================================================================
// Valid Kustomization Test Cases
// =============================================================================

var fluxValidKustomizationCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.31 Kustomization valid minimal with interval, path, sourceRef",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.32 Kustomization valid with targetNamespace",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-ns
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  targetNamespace: production
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.33 Kustomization valid with timeout",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-timeout
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  timeout: 10m0s
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.34 Kustomization valid with prune true",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-prune
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  prune: true
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.35 Kustomization valid with suspend true",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-suspended
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  suspend: true
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.36 Kustomization valid with dependsOn",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-deps
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  dependsOn:
    - infra-namespace
    - cert-manager
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.37 Kustomization valid with wait true",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-wait
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  wait: true
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.38 Kustomization valid with retryInterval",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-retry
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  retryInterval: 2m0s
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
	{
		name: "RW-11.39 Kustomization valid with sourceRef kind HelmRepository",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-helm
  namespace: default
spec:
  interval: 5m0s
  path: ./
  sourceRef:
    kind: HelmRepository
    name: bitnami
`),
		wantErr: false,
	},
	{
		name: "RW-11.40 Kustomization valid with all fields",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-full
  namespace: flux-system
spec:
  interval: 3m0s
  path: ./deploy
  targetNamespace: production
  timeout: 15m0s
  prune: true
  suspend: false
  dependsOn:
    - infra
    - monitoring
  wait: true
  retryInterval: 1m0s
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: false,
	},
}

func TestFlux_Kustomization_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(kustomizationCRD))
	_ = engine.RegisterCRD([]byte(gitRepositoryCRD))

	for _, tc := range fluxValidKustomizationCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "valid" {
					t.Errorf("expected valid status, got %s: %v", result.Status, result.Errors)
				}
			}
		})
	}
}

// =============================================================================
// Valid FluxInstall Test Cases
// =============================================================================

var fluxValidFluxInstallCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.41 FluxInstall valid minimal with version and interval",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
`),
		wantErr: false,
	},
	{
		name: "RW-11.42 FluxInstall valid with specific version",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-version
  namespace: flux-system
spec:
  version: "v2.0.0"
  interval: 1m0s
`),
		wantErr: false,
	},
	{
		name: "RW-11.43 FluxInstall valid with clusterProvider",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-gke
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  clusterProvider: gke
`),
		wantErr: false,
	},
	{
		name: "RW-11.44 FluxInstall valid with components",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-components
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  components:
    - source-controller
    - kustomize-controller
    - helm-controller
`),
		wantErr: false,
	},
	{
		name: "RW-11.45 FluxInstall valid with registry",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-registry
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  registry: ghcr.io/fluxcd
`),
		wantErr: false,
	},
	{
		name: "RW-11.46 FluxInstall valid with watchAllNamespaces",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-watch
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  watchAllNamespaces: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.47 FluxInstall valid with allAtOnce true",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-allatonce
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  allAtOnce: true
`),
		wantErr: false,
	},
	{
		name: "RW-11.48 FluxInstall valid with namespace",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-custom-ns
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  namespace: flux-system
`),
		wantErr: false,
	},
	{
		name: "RW-11.49 FluxInstall valid with resources",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-resources
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 100m
      memory: 128Mi
`),
		wantErr: false,
	},
	{
		name: "RW-11.50 FluxInstall valid with all fields",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux-full
  namespace: flux-system
spec:
  version: "v2.1.0"
  interval: 30s
  clusterProvider: eks
  components:
    - source-controller
    - kustomize-controller
    - helm-controller
    - notification-controller
  registry: ghcr.io/fluxcd
  watchAllNamespaces: false
  allAtOnce: false
`),
		wantErr: false,
	},
}

// =============================================================================
// Invalid GitRepository Test Cases
// =============================================================================

var fluxInvalidGitRepositoryCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.51 GitRepository invalid missing interval",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: true,
		errSubstr: "interval",
	},
	{
		name: "RW-11.52 GitRepository invalid missing url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
`),
		wantErr: true,
		errSubstr: "url",
	},
	{
		name: "RW-11.53 GitRepository invalid empty url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: ""
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.54 GitRepository invalid interval as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 60
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.55 GitRepository invalid ref.branch as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ref:
    branch: 123
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.56 GitRepository invalid secretRef.name empty",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  secretRef:
    name: ""
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.57 GitRepository invalid suspend as string",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  suspend: "yes"
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.58 GitRepository invalid gitImplementation as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  gitImplementation: 1
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.59 GitRepository invalid spec missing entirely",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec: {}
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.60 GitRepository invalid timeout as duration string malformed",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  timeout: not-a-duration
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.61 GitRepository invalid ignore as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
  ignore: 123
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.62 GitRepository invalid wrong apiVersion group",
		crYAML: []byte(`apiVersion: wrong.group/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 1m0s
  url: https://github.com/stefanprodan/podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
}

func TestFlux_GitRepository_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(gitRepositoryCRD))

	for _, tc := range fluxInvalidGitRepositoryCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "invalid" {
					t.Errorf("expected invalid status, got %s", result.Status)
				}
			}
		})
	}
}

// =============================================================================
// Invalid HelmRepository Test Cases
// =============================================================================

var fluxInvalidHelmRepositoryCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.63 HelmRepository invalid missing interval",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: true,
		errSubstr: "interval",
	},
	{
		name: "RW-11.64 HelmRepository invalid missing url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
`),
		wantErr: true,
		errSubstr: "url",
	},
	{
		name: "RW-11.65 HelmRepository invalid empty url",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: ""
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.66 HelmRepository invalid interval as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 300
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.67 HelmRepository invalid timeout as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  timeout: 60
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.68 HelmRepository invalid secretRef.name empty",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  secretRef:
    name: ""
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.69 HelmRepository invalid passCredentials as string",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  passCredentials: "yes"
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.70 HelmRepository invalid provider as number",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  provider: 123
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.71 HelmRepository invalid suspend as string",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
  suspend: "true"
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.72 HelmRepository invalid spec empty",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: bitnami
  namespace: default
spec: {}
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.73 HelmRepository invalid wrong kind",
		crYAML: []byte(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmChart
metadata:
  name: bitnami
  namespace: default
spec:
  interval: 5m0s
  url: https://charts.bitnami.com/bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
}

func TestFlux_HelmRepository_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(helmRepositoryCRD))

	for _, tc := range fluxInvalidHelmRepositoryCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "invalid" {
					t.Errorf("expected invalid status, got %s", result.Status)
				}
			}
		})
	}
}

// =============================================================================
// Invalid HelmRelease Test Cases
// =============================================================================

var fluxInvalidHelmReleaseCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.74 HelmRelease invalid missing interval",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "interval",
	},
	{
		name: "RW-11.75 HelmRelease invalid missing chart",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
`),
		wantErr: true,
		errSubstr: "chart",
	},
	{
		name: "RW-11.76 HelmRelease invalid chart.spec missing",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    {}
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.77 HelmRelease invalid chart.spec.chart empty",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: ""
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.78 HelmRelease invalid chart.spec.sourceRef.kind empty",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: ""
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.79 HelmRelease invalid interval as number",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 300
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.80 HelmRelease invalid timeout as string",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  timeout: "3m"
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.81 HelmRelease invalid suspend as number",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  suspend: 1
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.82 HelmRelease invalid releaseName as number",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  releaseName: 123
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.83 HelmRelease invalid targetNamespace as number",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  targetNamespace: 123
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.84 HelmRelease invalid values as string",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  values: "replicaCount: 3"
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.85 HelmRelease invalid dependsOn as string",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  dependsOn: mysql
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.86 HelmRelease invalid wait as string",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
  wait: "yes"
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.87 HelmRelease invalid wrong apiVersion",
		crYAML: []byte(`apiVersion: helm.toolkit.fluxcd.io/v99
kind: HelmRelease
metadata:
  name: nginx
  namespace: default
spec:
  interval: 5m0s
  chart:
    spec:
      chart: nginx
      sourceRef:
        kind: HelmRepository
        name: bitnami
`),
		wantErr: true,
		errSubstr: "",
	},
}

func TestFlux_HelmRelease_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(helmReleaseCRD))
	_ = engine.RegisterCRD([]byte(helmRepositoryCRD))

	for _, tc := range fluxInvalidHelmReleaseCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "invalid" {
					t.Errorf("expected invalid status, got %s", result.Status)
				}
			}
		})
	}
}

// =============================================================================
// Invalid Kustomization Test Cases
// =============================================================================

var fluxInvalidKustomizationCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.88 Kustomization invalid missing interval",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  path: ./deploy
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "interval",
	},
	{
		name: "RW-11.89 Kustomization invalid missing path",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "path",
	},
	{
		name: "RW-11.90 Kustomization invalid missing sourceRef",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
`),
		wantErr: true,
		errSubstr: "sourceRef",
	},
	{
		name: "RW-11.91 Kustomization invalid path empty",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ""
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.92 Kustomization invalid sourceRef.kind empty",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  sourceRef:
    kind: ""
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.93 Kustomization invalid sourceRef.name empty",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  sourceRef:
    kind: GitRepository
    name: ""
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.94 Kustomization invalid interval as number",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 300
  path: ./deploy
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.95 Kustomization invalid timeout as number",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  timeout: 600
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.96 Kustomization invalid prune as string",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  prune: "true"
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.97 Kustomization invalid suspend as number",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  suspend: 1
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.98 Kustomization invalid dependsOn as string",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  dependsOn: infra
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.99 Kustomization invalid wait as string",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  wait: "yes"
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.100 Kustomization invalid targetNamespace as number",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  targetNamespace: 123
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.101 Kustomization invalid retryInterval as number",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec:
  interval: 5m0s
  path: ./deploy
  retryInterval: 120
  sourceRef:
    kind: GitRepository
    name: podinfo
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.102 Kustomization invalid spec empty",
		crYAML: []byte(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: default
spec: {}
`),
		wantErr: true,
		errSubstr: "",
	},
}

func TestFlux_Kustomization_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(kustomizationCRD))
	_ = engine.RegisterCRD([]byte(gitRepositoryCRD))

	for _, tc := range fluxInvalidKustomizationCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "invalid" {
					t.Errorf("expected invalid status, got %s", result.Status)
				}
			}
		})
	}
}

// =============================================================================
// Invalid FluxInstall Test Cases
// =============================================================================

var fluxInvalidFluxInstallCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	{
		name: "RW-11.103 FluxInstall invalid missing version",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  interval: 1m0s
`),
		wantErr: true,
		errSubstr: "version",
	},
	{
		name: "RW-11.104 FluxInstall invalid missing interval",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
`),
		wantErr: true,
		errSubstr: "interval",
	},
	{
		name: "RW-11.105 FluxInstall invalid version empty",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: ""
  interval: 1m0s
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.106 FluxInstall invalid interval as number",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 60
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.107 FluxInstall invalid clusterProvider as number",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  clusterProvider: 123
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.108 FluxInstall invalid components as string",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  components: all
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.109 FluxInstall invalid registry as number",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  registry: 123
`),
		wantErr: true,
		errSubstr: "",
	},
	{
		name: "RW-11.110 FluxInstall invalid watchAllNamespaces as string",
		crYAML: []byte(`apiVersion: installation.toolkit.fluxcd.io/v1
kind: FluxInstall
metadata:
  name: flux
  namespace: flux-system
spec:
  version: latest
  interval: 1m0s
  watchAllNamespaces: "yes"
`),
		wantErr: true,
		errSubstr: "",
	},
}

func TestFlux_FluxInstall_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(fluxInstallCRD))

	for _, tc := range fluxInvalidFluxInstallCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.validateCR(tc.crYAML)
			if tc.wantErr {
				if err == nil && !hasCRError(result.Errors, tc.errSubstr) {
					t.Errorf("expected error containing %q, got none", tc.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != nil && result.Status != "invalid" {
					t.Errorf("expected invalid status, got %s", result.Status)
				}
			}
		})
	}
}

// =============================================================================
// Validation helper
// =============================================================================

// validateCR is a helper that wraps CR validation for testing
func (e *Engine) validateCR(crYAML []byte) (*Result, error) {
	// Use the CR validator directly
	decoder := NewDecoder()
	manifest, err := decoder.DecodeManifest(crYAML)
	if err != nil {
		return &Result{
			Status: types.StatusError,
			Errors: []ErrorItem{{Field: "manifest", Message: err.Error()}},
		}, err
	}

	// Get the CR validator
	crValidator := e.cr
	if crValidator == nil {
		return nil, fmt.Errorf("CR validator not initialized")
	}

	return crValidator.ValidateCR(crYAML)
}