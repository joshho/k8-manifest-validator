package validator

import (
	"testing"
)

// =============================================================================
// RW-7 — ArgoCD Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — argo-cd operator group
// Owner file: pkg/validator/phase4_realworld_argocd_test.go
// CRDs covered: Application, AppProject, ArgoCD, ApplicationSet
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// ArgoCD Operator CRD definitions (inline for test isolation)
// Based on argo-cd v2.x CRD schemas
// -----------------------------------------------------------------------

const applicationCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: applications.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: Application
    plural: applications
    singular: application
    listKind: ApplicationList
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              project:
                type: string
              source:
                type: object
                properties:
                  repoURL:
                    type: string
                  targetRevision:
                    type: string
                  path:
                    type: string
                  chart:
                    type: string
                  helm:
                    type: object
                  kustomize:
                    type: object
                required:
                - repoURL
              destination:
                type: object
                properties:
                  server:
                    type: string
                  namespace:
                    type: string
                  name:
                    type: string
                required:
                - namespace
              syncPolicy:
                type: object
                properties:
                  automated:
                    type: object
                  syncOptions:
                    type: array
                    items:
                      type: string
              ignoreDifferences:
                type: array
              info:
                type: array
            required:
            - project
            - source
            - destination
          status:
            type: object
`

const appProjectCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: appprojects.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: AppProject
    plural: appprojects
    singular: appproject
    listKind: AppProjectList
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              description:
                type: string
              sourceRepos:
                type: array
                items:
                  type: string
              destinations:
                type: array
                properties:
                  server:
                    type: string
                  namespace:
                    type: string
                  name:
                    type: string
                required:
                - namespace
              clusterResourceBlacklist:
                type: array
              namespaceResourceBlacklist:
                type: array
              permittedOnly:
                type: boolean
              syncWindows:
                type: array
              roles:
                type: array
              signatureKeys:
                type: array
            required:
            - sourceRepos
            - destinations
`

const argocdCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: argocds.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: ArgoCD
    plural: argocds
    singular: argocd
    listKind: ArgoCDList
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              server:
                type: object
                properties:
                  replicas:
                    type: integer
                    minimum: 1
                  insecure:
                    type: boolean
                  host:
                    type: string
              repo:
                type: object
                properties:
                  replicas:
                    type: integer
                    minimum: 1
                  resources:
                    type: object
              controller:
                type: object
                properties:
                  replicas:
                    type: integer
                    minimum: 1
              image:
                type: string
              version:
                type: string
              namespace:
                type: string
              rbac:
                type: object
              redis:
                type: object
              metrics:
                type: object
`

const applicationSetCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: applicationsets.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: ApplicationSet
    plural: applicationsets
    singular: applicationset
    listKind: ApplicationSetList
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              generators:
                type: array
              template:
                type: object
                properties:
                  metadata:
                    type: object
                  spec:
                    type: object
                required:
                - metadata
                - spec
              syncPolicy:
                type: object
              strategy:
                type: array
            required:
            - generators
            - template
`

// -----------------------------------------------------------------------
// ArgoCD Operator Valid Tests (RW-7.1 – RW-7.55)
// -----------------------------------------------------------------------

func TestArgoCD_Application_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.1 Application valid minimal with project source destination",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.2 Application valid with helm chart",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-helm-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 9.0.0
  destination:
    server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.3 Application valid with helm values",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-helm-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 9.0.0
    helm:
      values: |
        replicaCount: 2
        service:
          type: LoadBalancer
  destination:
    server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.4 Application valid with kustomize",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-kustomize-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    path: kustomize-guestbook
    targetRevision: HEAD
    kustomize:
      nameSuffix: -prod
  destination:
    server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.5 Application valid with syncPolicy automated",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.6 Application valid with syncOptions",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    syncOptions:
    - CreateNamespace=true
    - PrunePropagation=foreground
`),
			wantErr: false,
		},
		{
			name: "RW-7.7 Application valid with ignoreDifferences",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  ignoreDifferences:
  - group: apps
    jsonPointers:
    - /spec/replicas
`),
			wantErr: false,
		},
		{
			name: "RW-7.8 Application valid with info entries",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  info:
  - name: team
    value: platform
  - name: tier
    value: frontend
`),
			wantErr: false,
		},
		{
			name: "RW-7.9 Application valid with destination by name",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    name: in-cluster
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.10 Application valid with empty syncPolicy",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy: {}
`),
			wantErr: false,
		},
		{
			name: "RW-7.11 Application valid with multiple ignoreDifferences",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  ignoreDifferences:
  - group: apps
    jsonPointers:
    - /spec/replicas
  - group: ""
    jsonPointers:
    - /metadata/annotations
`),
			wantErr: false,
		},
		{
			name: "RW-7.12 Application valid with revision history limit",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.13 Application valid using cluster server address",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://10.0.0.1:6443
    namespace: staging
`),
			wantErr: false,
		},
		{
			name: "RW-7.14 Application valid with labels and annotations",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
  labels:
    environment: production
    team: platform
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.15 Application valid with namespace scoped to cluster",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: dev
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing Application valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_AppProject_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.16 AppProject valid minimal with sourceRepos and destinations",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.17 AppProject valid with description",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  description: Production deployment project
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.18 AppProject valid with multiple sourceRepos",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  - https://github.com/argoproj/argo-cd
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.19 AppProject valid with multiple destinations",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  - server: https://kubernetes.default.svc
    namespace: staging
`),
			wantErr: false,
		},
		{
			name: "RW-7.20 AppProject valid with permittedOnly",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: restricted-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  permittedOnly: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.21 AppProject valid with roles",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  roles:
  - name: project-admin
    description: Admin role for the project
`),
			wantErr: false,
		},
		{
			name: "RW-7.22 AppProject valid with clusterResourceBlacklist",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  clusterResourceBlacklist:
  - group: ""
    kind: ConfigMap
`),
			wantErr: false,
		},
		{
			name: "RW-7.23 AppProject valid with namespaceResourceBlacklist",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  namespaceResourceBlacklist:
  - group: ""
    kind: Secret
`),
			wantErr: false,
		},
		{
			name: "RW-7.24 AppProject valid with syncWindows",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  syncWindows:
  - kind: deny
    schedule: "* * * * *"
    duration: 1h
`),
			wantErr: false,
		},
		{
			name: "RW-7.25 AppProject valid with empty description",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  description: ""
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.26 AppProject valid destination by name",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - name: in-cluster
    namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.27 AppProject valid with signatureKeys",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  signatureKeys:
  - keyID: 1234567890
`),
			wantErr: false,
		},
		{
			name: "RW-7.28 AppProject valid with empty roles array",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  roles: []
`),
			wantErr: false,
		},
		{
			name: "RW-7.29 AppProject valid permittedOnly false",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: my-project
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  permittedOnly: false
`),
			wantErr: false,
		},
		{
			name: "RW-7.30 AppProject valid with all fields populated",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: full-project
  namespace: argocd
spec:
  description: Full featured project
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  - https://github.com/argoproj/argo-cd
  destinations:
  - server: https://kubernetes.default.svc
    namespace: production
  clusterResourceBlacklist:
  - group: ""
    kind: ConfigMap
  namespaceResourceBlacklist:
  - group: ""
    kind: Secret
  permittedOnly: false
  syncWindows:
  - kind: allow
    schedule: "0 8-17 * * 1-5"
    duration: 8h
    applications:
    - "*"
  roles:
  - name: admin
    description: Project administrator
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing AppProject valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_ArgoCD_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.31 ArgoCD valid minimal with image and version",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  image: quay.io/argoproj/argocd
  version: v2.8.0
`),
			wantErr: false,
		},
		{
			name: "RW-7.32 ArgoCD valid with server replicas",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 3
`),
			wantErr: false,
		},
		{
			name: "RW-7.33 ArgoCD valid with repo replicas",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  repo:
    replicas: 2
`),
			wantErr: false,
		},
		{
			name: "RW-7.34 ArgoCD valid with controller replicas",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  controller:
    replicas: 2
`),
			wantErr: false,
		},
		{
			name: "RW-7.35 ArgoCD valid with server insecure",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    insecure: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.36 ArgoCD valid with server host",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    host: argocd.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-7.37 ArgoCD valid with repo resources",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  repo:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
`),
			wantErr: false,
		},
		{
			name: "RW-7.38 ArgoCD valid with all components",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 2
    insecure: false
    host: argocd.example.com
  repo:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
  controller:
    replicas: 1
`),
			wantErr: false,
		},
		{
			name: "RW-7.39 ArgoCD valid with redis config",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  redis: {}
`),
			wantErr: false,
		},
		{
			name: "RW-7.40 ArgoCD valid with metrics config",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  metrics:
    enabled: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.41 ArgoCD valid with rbac config",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  rbac: {}
`),
			wantErr: false,
		},
		{
			name: "RW-7.42 ArgoCD valid with server replicas and insecure",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 3
    insecure: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.43 ArgoCD valid server replicas 1",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 1
`),
			wantErr: false,
		},
		{
			name: "RW-7.44 ArgoCD valid with namespace field",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  namespace: argocd
`),
			wantErr: false,
		},
		{
			name: "RW-7.45 ArgoCD valid full HA config",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd-ha
  namespace: argocd
spec:
  image: quay.io/argoproj/argocd
  version: v2.8.0
  server:
    replicas: 5
    host: argocd-ha.example.com
  repo:
    replicas: 3
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  controller:
    replicas: 3
  redis: {}
  metrics:
    enabled: true
  rbac: {}
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing ArgoCD valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_ApplicationSet_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.46 ApplicationSet valid with matrix generator",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - matrix:
      generators:
      - git:
          repoURL: https://github.com/argoproj/argocd-example-apps
          revision: HEAD
      - clusters:
          values:
            namespace: production
  template:
    metadata:
      name: '{{path.basename}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: '{{path.basename}}'
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.47 ApplicationSet valid with list generator",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
      - cluster: production
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.48 ApplicationSet valid with cluster generator",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - clusters:
      values:
        environment: production
  template:
    metadata:
      name: 'cluster-{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{server}}'
        namespace: production
`),
			wantErr: false,
		},
		{
			name: "RW-7.49 ApplicationSet valid with git generator",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - git:
      repoURL: https://github.com/argoproj/argocd-example-apps
      revision: HEAD
      directories:
      - path: "*"
  template:
    metadata:
      name: '{{path.basename}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: '{{path}}/guestbook'
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.50 ApplicationSet valid with syncPolicy",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
  syncPolicy:
    preserveResourcesOnDeletion: true
`),
			wantErr: false,
		},
		{
			name: "RW-7.51 ApplicationSet valid with strategy",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
  strategy:
  - type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
`),
			wantErr: false,
		},
		{
			name: "RW-7.52 ApplicationSet valid with multiple generators",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - env: prod
        cluster: production
  template:
    metadata:
      name: '{{env}}-{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.53 ApplicationSet valid with labels in metadata",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
      labels:
        environment: production
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
`),
			wantErr: false,
		},
		{
			name: "RW-7.54 ApplicationSet valid with merge strategy type",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
  strategy:
  - type: Replace
`),
			wantErr: false,
		},
		{
			name: "RW-7.55 ApplicationSet valid empty strategy array",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: '{{url}}'
        namespace: default
  strategy: []
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing ApplicationSet valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// ArgoCD Operator Invalid Tests (RW-7.56 – RW-7.110)
// -----------------------------------------------------------------------

func TestArgoCD_Application_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.56 Application missing project fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "project",
		},
		{
			name: "RW-7.57 Application missing source fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "source",
		},
		{
			name: "RW-7.58 Application missing destination fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
`),
			wantErr:   true,
			errSubstr: "destination",
		},
		{
			name: "RW-7.59 Application empty project fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: ""
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "project",
		},
		{
			name: "RW-7.60 Application source missing repoURL fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "repoURL",
		},
		{
			name: "RW-7.61 Application destination missing namespace fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
`),
			wantErr:   true,
			errSubstr: "namespace",
		},
		{
			name: "RW-7.62 Application empty spec fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec: {}
`),
			wantErr:   true,
			errSubstr: "project",
		},
		{
			name: "RW-7.63 Application wrong kind fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: BadApplication
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.64 Application wrong apiVersion fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1beta1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.65 Application empty metadata name fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ""
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-7.66 Application source repoURL wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: 12345
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.67 Application destination server wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: 12345
    namespace: default
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.68 Application destination namespace wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: 12345
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.69 Application source targetRevision wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: 12345
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.70 Application syncPolicy wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy: "automated"
`),
			wantErr:   true,
			errSubstr: "object",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing Application invalid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_AppProject_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.71 AppProject missing sourceRepos fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "sourceRepos",
		},
		{
			name: "RW-7.72 AppProject missing destinations fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
`),
			wantErr:   true,
			errSubstr: "destinations",
		},
		{
			name: "RW-7.73 AppProject empty sourceRepos fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos: []
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "sourceRepos",
		},
		{
			name: "RW-7.74 AppProject empty destinations fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations: []
`),
			wantErr:   true,
			errSubstr: "destinations",
		},
		{
			name: "RW-7.75 AppProject sourceRepos wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos: "https://github.com/argoproj/argocd-example-apps"
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-7.76 AppProject destinations wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations: {}
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-7.77 AppProject destination missing namespace fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
`),
			wantErr:   true,
			errSubstr: "namespace",
		},
		{
			name: "RW-7.78 AppProject empty spec fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec: {}
`),
			wantErr:   true,
			errSubstr: "sourceRepos",
		},
		{
			name: "RW-7.79 AppProject wrong kind fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: BadAppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.80 AppProject wrong apiVersion fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1beta1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.81 AppProject empty metadata name fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: ""
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-7.82 AppProject permittedOnly wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
  permittedOnly: "true"
`),
			wantErr:   true,
			errSubstr: "boolean",
		},
		{
			name: "RW-7.83 AppProject roles wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
  roles: "admin"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-7.84 AppProject description wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  description: 12345
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.85 AppProject syncWindows wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: default
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/argoproj/argocd-example-apps
  destinations:
  - server: https://kubernetes.default.svc
    namespace: default
  syncWindows: "always"
`),
			wantErr:   true,
			errSubstr: "array",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing AppProject invalid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_ArgoCD_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.86 ArgoCD server replicas string fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: "three"
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-7.87 ArgoCD server replicas zero fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 0
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-7.88 ArgoCD repo replicas zero fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  repo:
    replicas: 0
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-7.89 ArgoCD controller replicas zero fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  controller:
    replicas: 0
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-7.90 ArgoCD server replicas negative fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: -1
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-7.91 ArgoCD image wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  image: 12345
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.92 ArgoCD version wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  version: 12345
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.93 ArgoCD wrong kind fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: BadArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  image: quay.io/argoproj/argocd
  version: v2.8.0
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.94 ArgoCD wrong apiVersion fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1beta1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  image: quay.io/argoproj/argocd
  version: v2.8.0
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.95 ArgoCD empty metadata name fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: ""
  namespace: argocd
spec:
  image: quay.io/argoproj/argocd
  version: v2.8.0
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-7.96 ArgoCD server replicas float fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    replicas: 2.5
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-7.97 ArgoCD server host wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    host: 12345
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-7.98 ArgoCD server insecure wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  server:
    insecure: "true"
`),
			wantErr:   true,
			errSubstr: "boolean",
		},
		{
			name: "RW-7.99 ArgoCD repo resources wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  repo:
    resources: "cpu: 100m"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-7.100 ArgoCD rbac wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ArgoCD
metadata:
  name: argocd
  namespace: argocd
spec:
  rbac: "default"
`),
			wantErr:   true,
			errSubstr: "object",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing ArgoCD invalid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestArgoCD_ApplicationSet_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(applicationCRD))
	_ = engine.RegisterCRD([]byte(appProjectCRD))
	_ = engine.RegisterCRD([]byte(argocdCRD))
	_ = engine.RegisterCRD([]byte(applicationSetCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-7.101 ApplicationSet missing generators fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr:   true,
			errSubstr: "generators",
		},
		{
			name: "RW-7.102 ApplicationSet missing template fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
`),
			wantErr:   true,
			errSubstr: "template",
		},
		{
			name: "RW-7.103 ApplicationSet empty generators fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators: []
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr:   true,
			errSubstr: "generators",
		},
		{
			name: "RW-7.104 ApplicationSet generators wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators: "list"
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-7.105 ApplicationSet template wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template: "metadata"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-7.106 ApplicationSet template missing metadata fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr:   true,
			errSubstr: "metadata",
		},
		{
			name: "RW-7.107 ApplicationSet template missing spec fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-7.108 ApplicationSet empty spec fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec: {}
`),
			wantErr:   true,
			errSubstr: "generators",
		},
		{
			name: "RW-7.109 ApplicationSet wrong kind fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: BadApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-7.110 ApplicationSet syncPolicy wrong type fails",
			crYAML: []byte(`apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-appset
  namespace: argocd
spec:
  generators:
  - list:
      elements:
      - cluster: staging
        url: https://kubernetes.default.svc
  template:
    metadata:
      name: '{{name}}-app'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        path: guestbook
        targetRevision: HEAD
      destination:
        server: https://kubernetes.default.svc
        namespace: default
  syncPolicy: "delete"
`),
			wantErr:   true,
			errSubstr: "object",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-7] testing ApplicationSet invalid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}