package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

// =============================================================================
// RW-6 — Prometheus Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — prometheus-operator group
// Owner file: pkg/validator/phase4_realworld_prometheus_test.go
// CRDs covered: Prometheus, ServiceMonitor, PodMonitor, PrometheusRule, Alertmanager
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// Prometheus Operator CRD definitions (inline for test isolation)
// Based on prometheus-operator v0.70+ CRD schemas
// -----------------------------------------------------------------------

const prometheusCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: prometheuses.monitoring.coreos.com
spec:
  group: monitoring.coreos.com
  names:
    kind: Prometheus
    plural: prometheuses
    singular: prometheus
    listKind: PrometheusList
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
              replicas:
                type: integer
                minimum: 1
              version:
                type: string
              image:
                type: string
              resources:
                type: object
              serviceAccountName:
                type: string
              serviceMonitorSelector:
                type: object
              podMonitorSelector:
                type: object
              ruleSelector:
                type: object
              alertmanagerConfigSelector:
                type: object
              retention:
                type: string
              retentionSize:
                type: string
              securityContext:
                type: object
              containers:
                type: array
              volumes:
                type: array
            required:
            - replicas
            - serviceAccountName
`

const serviceMonitorCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: servicemonitors.monitoring.coreos.com
spec:
  group: monitoring.coreos.com
  names:
    kind: ServiceMonitor
    plural: servicemonitors
    singular: servicemonitor
    listKind: ServiceMonitorList
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
              jobName:
                type: string
              endpoints:
                type: array
              selector:
                type: object
              namespaceSelector:
                type: object
              podTargetLabels:
                type: array
                items:
                  type: string
            required:
            - jobName
            - endpoints
            - selector
`

const podMonitorCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: podmonitors.monitoring.coreos.com
spec:
  group: monitoring.coreos.com
  names:
    kind: PodMonitor
    plural: podmonitors
    singular: podmonitor
    listKind: PodMonitorList
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
              jobName:
                type: string
              podMetricsEndpoints:
                type: array
              selector:
                type: object
              namespaceSelector:
                type: object
              podTargetLabels:
                type: array
                items:
                  type: string
            required:
            - jobName
            - podMetricsEndpoints
            - selector
`

const prometheusRuleCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: prometheusrules.monitoring.coreos.com
spec:
  group: monitoring.coreos.com
  names:
    kind: PrometheusRule
    plural: prometheusrules
    singular: prometheusrule
    listKind: PrometheusRuleList
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
              groups:
                type: array
            required:
            - groups
`

const alertmanagerCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: alertmanagers.monitoring.coreos.com
spec:
  group: monitoring.coreos.com
  names:
    kind: Alertmanager
    plural: alertmanagers
    singular: alertmanager
    listKind: AlertmanagerList
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
              replicas:
                type: integer
                minimum: 1
              version:
                type: string
              image:
                type: string
              serviceAccountName:
                type: string
              securityContext:
                type: object
              containers:
                type: array
              volumes:
                type: array
            required:
            - replicas
            - serviceAccountName
`

// -----------------------------------------------------------------------
// Prometheus Operator Valid Tests (RW-6.1 – RW-6.11)
// -----------------------------------------------------------------------

func TestPrometheusOperator_Prometheus_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.1 Prometheus valid minimal with replicas and serviceAccountName",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
`),
			wantErr: false,
		},
		{
			name: "RW-6.2 Prometheus valid with image and version",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  image: quay.io/prometheus/prometheus:v2.50.0
  version: v2.50.0
`),
			wantErr: false,
		},
		{
			name: "RW-6.3 Prometheus valid with resources",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  serviceAccountName: prometheus
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
`),
			wantErr: false,
		},
		{
			name: "RW-6.4 Prometheus valid with retention",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  retention: 15d
  retentionSize: 10GB
`),
			wantErr: false,
		},
		{
			name: "RW-6.5 Prometheus valid with serviceMonitorSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      tier: frontend
`),
			wantErr: false,
		},
		{
			name: "RW-6.6 Prometheus valid with podMonitorSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  podMonitorSelector:
    matchLabels:
      tier: backend
`),
			wantErr: false,
		},
		{
			name: "RW-6.7 Prometheus valid with ruleSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  ruleSelector:
    matchLabels:
      role: alert-rules
`),
			wantErr: false,
		},
		{
			name: "RW-6.8 Prometheus valid with alertmanagerConfigSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  alertmanagerConfigSelector:
    matchLabels:
      alertmanager: main
`),
			wantErr: false,
		},
		{
			name: "RW-6.9 Prometheus valid with securityContext",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
`),
			wantErr: false,
		},
		{
			name: "RW-6.10 Prometheus valid with containers and volumes",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
  containers:
  - name: prometheus
    image: quay.io/prometheus/prometheus:v2.50.0
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
  volumes:
  - name: storage
    emptyDir: {}
`),
			wantErr: false,
		},
		{
			name: "RW-6.11 Prometheus valid high availability 3 replicas",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus-ha
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: prometheus
  version: v2.50.0
  retention: 30d
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing Prometheus valid: %s", tc.name)
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
// ServiceMonitor Valid Tests (RW-6.12 – RW-6.22)
// -----------------------------------------------------------------------

func TestPrometheusOperator_ServiceMonitor_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.12 ServiceMonitor valid minimal",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.13 ServiceMonitor valid with interval",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
    interval: 30s
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.14 ServiceMonitor valid with bearerTokenFile",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: https-metrics
  endpoints:
  - port: web
    scheme: https
    bearerTokenFile: /var/run/secrets/tokens/token
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.15 ServiceMonitor valid with tlsConfig",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: https-metrics
  endpoints:
  - port: web
    scheme: https
    tlsConfig:
      insecureSkipVerify: false
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.16 ServiceMonitor valid with namespaceSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
  namespaceSelector:
    matchNames:
    - production
`),
			wantErr: false,
		},
		{
			name: "RW-6.17 ServiceMonitor valid with podTargetLabels",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
  podTargetLabels:
  - tier
  - app
`),
			wantErr: false,
		},
		{
			name: "RW-6.18 ServiceMonitor valid with multiple endpoints",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: multi-metrics
  endpoints:
  - port: web
    interval: 30s
  - port: metrics
    interval: 15s
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.19 ServiceMonitor valid path override",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
    path: /custom/metrics
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.20 ServiceMonitor valid honorLabels true",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
    honorLabels: true
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.21 ServiceMonitor valid with scrapeTimeout",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
    scrapeTimeout: 10s
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.22 ServiceMonitor valid with scheme http",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
    scheme: http
  selector:
    matchLabels:
      tier: api
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing ServiceMonitor valid: %s", tc.name)
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
// PodMonitor Valid Tests (RW-6.23 – RW-6.33)
// -----------------------------------------------------------------------

func TestPrometheusOperator_PodMonitor_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.23 PodMonitor valid minimal",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.24 PodMonitor valid with interval",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    interval: 30s
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.25 PodMonitor valid with path",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    path: /custom/path
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.26 PodMonitor valid with scheme https",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    scheme: https
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.27 PodMonitor valid with bearerTokenFile",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    bearerTokenFile: /var/run/secrets/tokens/token
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.28 PodMonitor valid with tlsConfig",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    tlsConfig:
      insecureSkipVerify: false
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.29 PodMonitor valid with namespaceSelector",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
  namespaceSelector:
    matchNames:
    - production
`),
			wantErr: false,
		},
		{
			name: "RW-6.30 PodMonitor valid with podTargetLabels",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
  podTargetLabels:
  - tier
  - app
`),
			wantErr: false,
		},
		{
			name: "RW-6.31 PodMonitor valid honorLabels true",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    honorLabels: true
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.32 PodMonitor valid with scrapeTimeout",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
    scrapeTimeout: 10s
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
		{
			name: "RW-6.33 PodMonitor valid with multiple endpoints",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-multi-metrics
  podMetricsEndpoints:
  - port: metrics
    interval: 30s
  - port: web
    interval: 15s
  selector:
    matchLabels:
      app: myapp
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing PodMonitor valid: %s", tc.name)
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
// PrometheusRule Valid Tests (RW-6.34 – RW-6.44)
// -----------------------------------------------------------------------

func TestPrometheusOperator_PrometheusRule_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.34 PrometheusRule valid minimal with groups",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: ExampleAlert
      expr: up == 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Instance down"
`),
			wantErr: false,
		},
		{
			name: "RW-6.35 PrometheusRule valid with recording rule",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: recordings
    rules:
    - record: job:http_requests:rate5m
      expr: rate(http_requests_total[5m])
`),
			wantErr: false,
		},
		{
			name: "RW-6.36 PrometheusRule valid with multiple groups",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: group1
    rules:
    - alert: Alert1
      expr: up == 0
      for: 5m
  - name: group2
    rules:
    - alert: Alert2
      expr: rate(errors_total[5m]) > 0.1
      for: 10m
`),
			wantErr: false,
		},
		{
			name: "RW-6.37 PrometheusRule valid with multiple rules in group",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: multi
    rules:
    - alert: HighMemory
      expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
      for: 5m
      labels:
        severity: warning
    - alert: HighCPU
      expr: rate(process_cpu_seconds_total[5m]) > 0.8
      for: 5m
      labels:
        severity: warning
`),
			wantErr: false,
		},
		{
			name: "RW-6.38 PrometheusRule valid with for duration",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: InstanceDown
      expr: up == 0
      for: 10m
      labels:
        severity: critical
`),
			wantErr: false,
		},
		{
			name: "RW-6.39 PrometheusRule valid with labels",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: InstanceDown
      expr: up == 0
      for: 5m
      labels:
        severity: critical
        team: ops
        service: api
`),
			wantErr: false,
		},
		{
			name: "RW-6.40 PrometheusRule valid with annotations",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: InstanceDown
      expr: up == 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Instance {{ $labels.instance }} is down"
        description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 5 minutes"
`),
			wantErr: false,
		},
		{
			name: "RW-6.41 PrometheusRule valid with interval",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    interval: 30s
    rules:
    - alert: InstanceDown
      expr: up == 0
      for: 5m
`),
			wantErr: false,
		},
		{
			name: "RW-6.42 PrometheusRule valid with complex expr",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: HighErrorRate
      expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
      for: 5m
      labels:
        severity: critical
`),
			wantErr: false,
		},
		{
			name: "RW-6.43 PrometheusRule valid with expr using functions",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: InstanceMemoryUsage
      expr: (node_memory_MemTotal_bytes - node_memory_MemFree_bytes) / node_memory_MemTotal_bytes > 0.85
      for: 10m
      labels:
        severity: warning
`),
			wantErr: false,
		},
		{
			name: "RW-6.44 PrometheusRule valid with empty annotations dict",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: InstanceDown
      expr: up == 0
      for: 5m
      labels:
        severity: critical
      annotations: {}
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing PrometheusRule valid: %s", tc.name)
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
// Alertmanager Valid Tests (RW-6.45 – RW-6.55)
// -----------------------------------------------------------------------

func TestPrometheusOperator_Alertmanager_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.45 Alertmanager valid minimal",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
`),
			wantErr: false,
		},
		{
			name: "RW-6.46 Alertmanager valid with version and image",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  image: quay.io/prometheus/alertmanager:v0.26.0
  version: v0.26.0
`),
			wantErr: false,
		},
		{
			name: "RW-6.47 Alertmanager valid single replica",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 1
  serviceAccountName: alertmanager
`),
			wantErr: false,
		},
		{
			name: "RW-6.48 Alertmanager valid with securityContext",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
`),
			wantErr: false,
		},
		{
			name: "RW-6.49 Alertmanager valid with containers",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  containers:
  - name: alertmanager
    image: quay.io/prometheus/alertmanager:v0.26.0
    resources:
      requests:
        cpu: 100m
        memory: 64Mi
      limits:
        cpu: 200m
        memory: 128Mi
`),
			wantErr: false,
		},
		{
			name: "RW-6.50 Alertmanager valid with volumes",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  volumes:
  - name: config
    secret:
      secretName: alertmanager-config
`),
			wantErr: false,
		},
		{
			name: "RW-6.51 Alertmanager valid 5 replicas HA",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager-ha
  namespace: monitoring
spec:
  replicas: 5
  serviceAccountName: alertmanager
  version: v0.26.0
`),
			wantErr: false,
		},
		{
			name: "RW-6.52 Alertmanager valid with resources and limits",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  containers:
  - name: alertmanager
    image: quay.io/prometheus/alertmanager:v0.26.0
    resources:
      requests:
        cpu: 50m
        memory: 32Mi
      limits:
        cpu: 500m
        memory: 256Mi
`),
			wantErr: false,
		},
		{
			name: "RW-6.53 Alertmanager valid empty containers array",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  containers: []
`),
			wantErr: false,
		},
		{
			name: "RW-6.54 Alertmanager valid empty volumes array",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  volumes: []
`),
			wantErr: false,
		},
		{
			name: "RW-6.55 Alertmanager valid with image pull policy",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  containers:
  - name: alertmanager
    image: quay.io/prometheus/alertmanager:v0.26.0
    imagePullPolicy: IfNotPresent
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing Alertmanager valid: %s", tc.name)
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
// Prometheus Operator Invalid Tests (RW-6.56 – RW-6.110)
// -----------------------------------------------------------------------

func TestPrometheusOperator_Prometheus_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.56 Prometheus missing replicas fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-6.57 Prometheus missing serviceAccountName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
`),
			wantErr:   true,
			errSubstr: "serviceAccountName",
		},
		{
			name: "RW-6.58 Prometheus replicas string fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: "two"
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-6.59 Prometheus replicas zero fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 0
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-6.60 Prometheus replicas negative fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: -1
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-6.61 Prometheus empty spec fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-6.62 Prometheus empty serviceAccountName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: ""
`),
			wantErr:   true,
			errSubstr: "serviceAccountName",
		},
		{
			name: "RW-6.63 Prometheus replicas float fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2.5
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-6.64 Prometheus wrong kind fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusBad
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.65 Prometheus empty metadata name fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: ""
  namespace: monitoring
spec:
  replicas: 2
  serviceAccountName: prometheus
`),
			wantErr:   true,
			errSubstr: "name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing Prometheus invalid: %s", tc.name)
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

func TestPrometheusOperator_ServiceMonitor_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.66 ServiceMonitor missing jobName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.67 ServiceMonitor missing endpoints fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "endpoints",
		},
		{
			name: "RW-6.68 ServiceMonitor missing selector fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
`),
			wantErr:   true,
			errSubstr: "selector",
		},
		{
			name: "RW-6.69 ServiceMonitor empty jobName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: ""
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.70 ServiceMonitor endpoints wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints: "web"
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.71 ServiceMonitor empty endpoints array fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints: []
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "endpoints",
		},
		{
			name: "RW-6.72 ServiceMonitor podTargetLabels wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
  podTargetLabels: "tier"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.73 ServiceMonitor empty spec fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.74 ServiceMonitor wrong apiVersion fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1beta1
kind: ServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.75 ServiceMonitor wrong kind fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: BadServiceMonitor
metadata:
  name: my-service-monitor
  namespace: monitoring
spec:
  jobName: http-metrics
  endpoints:
  - port: web
  selector:
    matchLabels:
      tier: api
`),
			wantErr:   true,
			errSubstr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing ServiceMonitor invalid: %s", tc.name)
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

func TestPrometheusOperator_PodMonitor_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.76 PodMonitor missing jobName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.77 PodMonitor missing podMetricsEndpoints fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "podMetricsEndpoints",
		},
		{
			name: "RW-6.78 PodMonitor missing selector fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
`),
			wantErr:   true,
			errSubstr: "selector",
		},
		{
			name: "RW-6.79 PodMonitor empty jobName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: ""
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.80 PodMonitor podMetricsEndpoints wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints: "metrics"
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.81 PodMonitor empty podMetricsEndpoints array fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints: []
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "podMetricsEndpoints",
		},
		{
			name: "RW-6.82 PodMonitor empty spec fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "jobName",
		},
		{
			name: "RW-6.83 PodMonitor wrong apiVersion fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1beta1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.84 PodMonitor wrong kind fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: BadPodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.85 PodMonitor podTargetLabels wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: my-pod-monitor
  namespace: monitoring
spec:
  jobName: pod-metrics
  podMetricsEndpoints:
  - port: metrics
  selector:
    matchLabels:
      app: myapp
  podTargetLabels: "tier"
`),
			wantErr:   true,
			errSubstr: "array",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing PodMonitor invalid: %s", tc.name)
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

func TestPrometheusOperator_PrometheusRule_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.86 PrometheusRule missing groups fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "groups",
		},
		{
			name: "RW-6.87 PrometheusRule empty groups array fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups: []
`),
			wantErr:   true,
			errSubstr: "groups",
		},
		{
			name: "RW-6.88 PrometheusRule groups wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups: "default"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.89 PrometheusRule empty spec fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "groups",
		},
		{
			name: "RW-6.90 PrometheusRule wrong apiVersion fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1beta1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: ExampleAlert
      expr: up == 0
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.91 PrometheusRule wrong kind fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: BadPrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: ExampleAlert
      expr: up == 0
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.92 PrometheusRule empty group name fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: ""
    rules:
    - alert: ExampleAlert
      expr: up == 0
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.93 PrometheusRule group missing rules fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
`),
			wantErr:   true,
			errSubstr: "rules",
		},
		{
			name: "RW-6.94 PrometheusRule group empty rules fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules: []
`),
			wantErr:   true,
			errSubstr: "rules",
		},
		{
			name: "RW-6.95 PrometheusRule alert missing expr fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-prometheus-rule
  namespace: monitoring
spec:
  groups:
  - name: example
    rules:
    - alert: ExampleAlert
`),
			wantErr:   true,
			errSubstr: "expr",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing PrometheusRule invalid: %s", tc.name)
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

func TestPrometheusOperator_Alertmanager_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(prometheusCRD))
	_ = engine.RegisterCRD([]byte(serviceMonitorCRD))
	_ = engine.RegisterCRD([]byte(podMonitorCRD))
	_ = engine.RegisterCRD([]byte(prometheusRuleCRD))
	_ = engine.RegisterCRD([]byte(alertmanagerCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		{
			name: "RW-6.96 Alertmanager missing replicas fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-6.97 Alertmanager missing serviceAccountName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
`),
			wantErr:   true,
			errSubstr: "serviceAccountName",
		},
		{
			name: "RW-6.98 Alertmanager replicas string fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: "three"
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-6.99 Alertmanager replicas zero fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 0
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-6.100 Alertmanager replicas negative fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: -1
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "minimum",
		},
		{
			name: "RW-6.101 Alertmanager empty spec fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec: {}
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-6.102 Alertmanager empty serviceAccountName fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: ""
`),
			wantErr:   true,
			errSubstr: "serviceAccountName",
		},
		{
			name: "RW-6.103 Alertmanager replicas float fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 2.5
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-6.104 Alertmanager wrong apiVersion fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1beta1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.105 Alertmanager wrong kind fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: BadAlertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-6.106 Alertmanager empty metadata name fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: ""
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-6.107 Alertmanager containers wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  containers: "alertmanager"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.108 Alertmanager volumes wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  volumes: "config"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-6.109 Alertmanager securityContext wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  securityContext: "nonroot"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-6.110 Alertmanager image wrong type fails",
			crYAML: []byte(`apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: alertmanager
  namespace: monitoring
spec:
  replicas: 3
  serviceAccountName: alertmanager
  image:
    repository: quay.io/prometheus/alertmanager
`),
			wantErr:   true,
			errSubstr: "string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-6] testing Alertmanager invalid: %s", tc.name)
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
// Helper — check if any error in the list contains the given substring
// -----------------------------------------------------------------------

func hasCRError(errs []ErrorItem, substr string) bool {
	for _, e := range errs {
		if contains(e.Field, substr) || contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}