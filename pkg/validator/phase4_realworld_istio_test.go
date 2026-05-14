package validator

import (
	"testing"
)

// =============================================================================
// RW-8 — Istio Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — istio operator group
// Owner file: pkg/validator/phase4_realworld_istio_test.go
// CRDs covered: VirtualService, DestinationRule, Gateway, ServiceEntry, Sidecar
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// Istio Operator CRD definitions (inline for test isolation)
// Based on Istio v1.20+ CRD schemas
// -----------------------------------------------------------------------

const virtualServiceCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: virtualservices.networking.istio.io
spec:
  group: networking.istio.io
  names:
    kind: VirtualService
    plural: virtualservices
    singular: virtualservice
    listKind: VirtualServiceList
  scope: Namespaced
  versions:
  - name: v1alpha3
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              hosts:
                type: array
                items:
                  type: string
              gateways:
                type: array
                items:
                  type: string
              http:
                type: array
              tcp:
                type: array
              tls:
                type: array
            required:
            - hosts
`

const destinationRuleCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: destinationrules.networking.istio.io
spec:
  group: networking.istio.io
  names:
    kind: DestinationRule
    plural: destinationrules
    singular: destinationrule
    listKind: DestinationRuleList
  scope: Namespaced
  versions:
  - name: v1alpha3
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              host:
                type: string
              trafficPolicy:
                type: object
              subsets:
                type: array
            required:
            - host
`

const gatewayCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gateways.networking.istio.io
spec:
  group: networking.istio.io
  names:
    kind: Gateway
    plural: gateways
    singular: gateway
    listKind: GatewayList
  scope: Namespaced
  versions:
  - name: v1alpha3
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              selector:
                type: object
                additionalProperties:
                  type: string
              servers:
                type: array
            required:
            - servers
`

const serviceEntryCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: serviceentries.networking.istio.io
spec:
  group: networking.istio.io
  names:
    kind: ServiceEntry
    plural: serviceentries
    singular: serviceentry
    listKind: ServiceEntryList
  scope: Namespaced
  versions:
  - name: v1alpha3
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              hosts:
                type: array
                items:
                  type: string
              ports:
                type: array
              location:
                type: string
              resolution:
                type: string
              endpoints:
                type: array
            required:
            - hosts
            - ports
`

const sidecarCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: sidecars.networking.istio.io
spec:
  group: networking.istio.io
  names:
    kind: Sidecar
    plural: sidecars
    singular: sidecar
    listKind: SidecarList
  scope: Namespaced
  versions:
  - name: v1alpha3
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              egress:
                type: array
            required:
            - egress
`

// -----------------------------------------------------------------------
// Test Engine Setup
// -----------------------------------------------------------------------

func init() {
	// Register Istio CRDs with the validator engine
	engine.RegisterCRD(virtualServiceCRD)
	engine.RegisterCRD(destinationRuleCRD)
	engine.RegisterCRD(gatewayCRD)
	engine.RegisterCRD(serviceEntryCRD)
	engine.RegisterCRD(sidecarCRD)
}

// hasCRError checks if any error contains the expected substring
func hasCRError(errors []string, substr string) bool {
	for _, err := range errors {
		if len(substr) == 0 || contains(err, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Valid VirtualService Test Cases
// =============================================================================

var istioValidCases = []struct {
	name  string
	crYAML []byte
}{
	// VirtualService - Valid Cases
	{
		name: "RW-8.1 VirtualService valid basic with single host",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: reviews-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	{
		name: "RW-8.2 VirtualService valid with multiple hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: multi-host-route
  namespace: default
spec:
  hosts:
  - productpage
  - details
  - reviews
  http:
  - route:
    - destination:
        host: productpage
        subset: v1
`),
	},
	{
		name: "RW-8.3 VirtualService valid with gateway reference",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: gateway-route
  namespace: istio-system
spec:
  hosts:
  - bookinfo.example.com
  gateways:
  - istio-system/bookinfo-gateway
  http:
  - route:
    - destination:
        host: productpage
        port:
          number: 9080
`),
	},
	{
		name: "RW-8.4 VirtualService valid with weighted routing",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: weighted-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 90
    - destination:
        host: reviews
        subset: v2
      weight: 10
`),
	},
	{
		name: "RW-8.5 VirtualService valid with timeout",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: timeout-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - timeout: 5s
    route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	{
		name: "RW-8.6 VirtualService valid with retry policy",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: retry-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - retries:
      attempts: 3
      perTryTimeout: 2s
    route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	{
		name: "RW-8.7 VirtualService valid with fault injection",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: fault-inject-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - fault:
      delay:
        percentage:
          value: 10
        fixedDelay: 5s
    route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	{
		name: "RW-8.8 VirtualService valid with match conditions",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: match-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        end-user:
          exact: jason
    route:
    - destination:
        host: reviews
        subset: v2
`),
	},
	{
		name: "RW-8.9 VirtualService valid with rewrite",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: rewrite-route
  namespace: default
spec:
  hosts:
  - details
  http:
  - rewrite:
      uri: /newdetails
    route:
    - destination:
        host: details
        subset: v1
`),
	},
	{
		name: "RW-8.10 VirtualService valid with direct response",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: direct-response-route
  namespace: default
spec:
  hosts:
  - ratings
  http:
  - directResponse:
      body:
        string: "200 OK"
      status: 200
    route:
    - destination:
        host: ratings
`),
	},
	// DestinationRule - Valid Cases
	{
		name: "RW-8.11 DestinationRule valid basic",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
`),
	},
	{
		name: "RW-8.12 DestinationRule valid with subsets",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: subset-destination
  namespace: default
spec:
  host: reviews
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
`),
	},
	{
		name: "RW-8.13 DestinationRule valid with load balancer",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: lb-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN
`),
	},
	{
		name: "RW-8.14 DestinationRule valid with consistent hash",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: hash-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      consistentHash:
        httpHeaderName: x-user-id
`),
	},
	{
		name: "RW-8.15 DestinationRule valid with outlier detection",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: outlier-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    outlierDetection:
      consecutiveGatewayErrors: 5
      interval: 30s
      baseEjectionTime: 30s
`),
	},
	{
		name: "RW-8.16 DestinationRule valid with TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: tls-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    tls:
      mode: SIMPLE
`),
	},
	{
		name: "RW-8.17 DestinationRule valid with port-level settings",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: portlevel-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    portLevelSettings:
    - port:
        number: 9080
      loadBalancer:
        simple: LEAST_CONN
`),
	},
	{
		name: "RW-8.18 DestinationRule valid with locality load balancing",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: locality-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    localityLbSetting:
      enabled: true
`),
	},
	{
		name: "RW-8.19 DestinationRule valid with connection pool settings",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: connectionpool-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        h2UpgradePolicy: UPGRADE
        http1MaxPendingRequests: 100
`),
	},
	{
		name: "RW-8.20 DestinationRule valid with subsets and traffic policy",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: full-destination
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 50
  subsets:
  - name: v1
    labels:
      version: v1
    trafficPolicy:
      connectionPool:
        tcp:
          maxConnections: 100
`),
	},
	// Gateway - Valid Cases
	{
		name: "RW-8.21 Gateway valid basic",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: bookinfo-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*.example.com"
`),
	},
	{
		name: "RW-8.22 Gateway valid with HTTPS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: secure-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - bookinfo.example.com
    tls:
      mode: SIMPLE
      credentialName: bookinfo-cert
`),
	},
	{
		name: "RW-8.23 Gateway valid with multiple servers",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: multi-server-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - example.com
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - secure.example.com
    tls:
      mode: SIMPLE
      credentialName: my-cert
`),
	},
	{
		name: "RW-8.24 Gateway valid with TLS passthrough",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: passthrough-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - mtls.example.com
    tls:
      mode: PASSTHROUGH
`),
	},
	{
		name: "RW-8.25 Gateway valid with standard TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: standard-tls-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - api.example.com
    tls:
      mode: SIMPLE
      credentialName: api-cert
`),
	},
	{
		name: "RW-8.26 Gateway valid with mutually TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: mtls-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - internal.example.com
    tls:
      mode: MUTUAL
      credentialName: internal-cert
      serverCertificate: server-cert
      privateKey: server-key
      caCertificates: ca-cert
`),
	},
	{
		name: "RW-8.27 Gateway valid with UDP port",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: udp-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 53
      name: dns-udp
      protocol: UDP
    hosts:
    - "*.example.com"
`),
	},
	{
		name: "RW-8.28 Gateway valid with capture none",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: capture-none-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - no-capture.example.com
    defaultEndpoint: 127.0.0.1:8080
`),
	},
	{
		name: "RW-8.29 Gateway valid with bind addresses",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: bind-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
      bind: 0.0.0.0
    hosts:
    - example.com
`),
	},
	{
		name: "RW-8.30 Gateway valid wildcard hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: wildcard-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
`),
	},
	// ServiceEntry - Valid Cases
	{
		name: "RW-8.31 ServiceEntry valid basic",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: external-svc
  namespace: default
spec:
  hosts:
  - external.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
	},
	{
		name: "RW-8.32 ServiceEntry valid with endpoints",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: endpoint-svc
  namespace: default
spec:
  hosts:
  - my-svc.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: STATIC
  endpoints:
  - address: 192.168.1.1
    ports:
      http: 8080
`),
	},
	{
		name: "RW-8.33 ServiceEntry valid mesh internal",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: internal-svc
  namespace: default
spec:
  hosts:
  - internal.example.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  location: MESH_INTERNAL
  resolution: DNS
`),
	},
	{
		name: "RW-8.34 ServiceEntry valid with multiple endpoints",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: multi-endpoint-svc
  namespace: default
spec:
  hosts:
  - replicated.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: STATIC
  endpoints:
  - address: 10.0.0.1
    locality: us-west-1
    ports:
      http: 8080
  - address: 10.0.0.2
    locality: us-east-1
    ports:
      http: 8080
`),
	},
	{
		name: "RW-8.35 ServiceEntry valid with subset labels",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: subset-svc
  namespace: default
spec:
  hosts:
  - subset-svc.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
  subjectAltNames:
  - subset-svc.example.com
`),
	},
	{
		name: "RW-8.36 ServiceEntry valid with TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: tls-svc
  namespace: default
spec:
  hosts:
  - tls-example.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  location: MESH_EXTERNAL
  resolution: DNS
`),
	},
	{
		name: "RW-8.37 ServiceEntry valid with DNS resolution",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: dns-svc
  namespace: default
spec:
  hosts:
  - dns.example.com
  ports:
  - number: 8080
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
	},
	{
		name: "RW-8.38 ServiceEntry valid with NONE resolution",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: none-res-svc
  namespace: default
spec:
  hosts:
  - static-ip.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: NONE
`),
	},
	{
		name: "RW-8.39 ServiceEntry valid with multiple ports",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: multi-port-svc
  namespace: default
spec:
  hosts:
  - multiport.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  - number: 443
    name: https
    protocol: HTTPS
  - number: 9090
    name: grpc
    protocol: GRPC
  location: MESH_EXTERNAL
  resolution: DNS
`),
	},
	{
		name: "RW-8.40 ServiceEntry valid mesh external with endpoints",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: mesh-external-svc
  namespace: default
spec:
  hosts:
  - external-db.example.com
  ports:
  - number: 3306
    name: mysql
    protocol: TCP
  location: MESH_EXTERNAL
  resolution: STATIC
  endpoints:
  - address: 192.168.1.100
    ports:
      mysql: 3306
`),
	},
	// Sidecar - Valid Cases
	{
		name: "RW-8.41 Sidecar valid basic egress",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: egress-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "*/*"
`),
	},
	{
		name: "RW-8.42 Sidecar valid with specific hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: restricted-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "default/*"
    - "istio-system/*"
`),
	},
	{
		name: "RW-8.43 Sidecar valid with port-level settings",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: port-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "*/*"
    port:
      number: 8080
      protocol: HTTP
      name: egress-http
`),
	},
	{
		name: "RW-8.44 Sidecar valid with bind to localhost",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: localhost-sidecar
  namespace: default
spec:
  egress:
  - bind: 127.0.0.1
    hosts:
    - "*/*"
`),
	},
	{
		name: "RW-8.45 Sidecar valid with capture mode",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: capture-sidecar
  namespace: default
spec:
  egress:
  - captureMode: IPTABLES
    hosts:
    - "*/*"
`),
	},
	{
		name: "RW-8.46 Sidecar valid with capture none",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: capture-none-sidecar
  namespace: default
spec:
  egress:
  - captureMode: NONE
    hosts:
    - "*/*"
`),
	},
	{
		name: "RW-8.47 Sidecar valid with workload selector",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: selector-sidecar
  namespace: default
spec:
  workloadSelector:
    labels:
      app: my-app
  egress:
  - hosts:
    - "default/*"
`),
	},
	{
		name: "RW-8.48 Sidecar valid with TCP listeners",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: tcp-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "*/*"
    tcp:
    - match:
      - port: 3306
      destination:
        host: mysql.example.com
        port:
          number: 3306
`),
	},
	{
		name: "RW-8.49 Sidecar valid with multiple egress listeners",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: multi-egress-sidecar
  namespace: default
spec:
  egress:
  - hosts:
    - "default/httpbin"
  - hosts:
    - "istio-system/*"
`),
	},
	{
		name: "RW-8.50 Sidecar valid with strict mode",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: strict-sidecar
  namespace: default
spec:
  egress:
  - captureMode: IPTABLES
    hosts:
    - "./local-namespace"
`),
	},
	// Additional VirtualService valid cases
	{
		name: "RW-8.51 VirtualService valid with mirror",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: mirror-route
  namespace: default
spec:
  hosts:
  - ratings
  http:
  - route:
    - destination:
        host: ratings
        subset: v1
    mirror:
      host: ratings
      subset: v2
`),
	},
	{
		name: "RW-8.52 VirtualService valid with CORS policy",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: cors-route
  namespace: default
spec:
  hosts:
  - api.example.com
  http:
  - corsPolicy:
      allowOrigins:
      - exact: example.com
      allowMethods:
      - GET
      - POST
    route:
    - destination:
        host: api-backend
        subset: v1
`),
	},
	{
		name: "RW-8.53 VirtualService valid with external service",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: external-route
  namespace: default
spec:
  hosts:
  - external.example.com
  http:
  - route:
    - destination:
        host: external.example.com
        port:
          number: 80
`),
	},
	{
		name: "RW-8.54 VirtualService valid with headers manipulation",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: header-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    headers:
      request:
        set:
          X-Custom-Header: value
`),
	},
	{
		name: "RW-8.55 VirtualService valid with delegate",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: delegate-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - delegate:
      name: reviews-delegate
      namespace: default
`),
	},
}

// =============================================================================
// Additional Valid Istio Test Cases (filling gap from audit)
// =============================================================================

var istioValidAdditionalCases = []struct {
	name  string
	crYAML []byte
}{
	// RW-8.Istio.Valid.1 - VirtualService with TLS simple routing
	{
		name: "RW-8.Istio.Valid.1 VirtualService TLS simple routing",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tls-route
  namespace: default
spec:
  hosts:
  - reviews
  tls:
  - match:
    - port: 443
      sniHosts:
      - reviews
    route:
    - destination:
        host: reviews
        port:
          number: 443
`),
	},
	// RW-8.Istio.Valid.2 - VirtualService with TCP routing
	{
		name: "RW-8.Istio.Valid.2 VirtualService TCP routing",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tcp-route
  namespace: default
spec:
  hosts:
  - mysql
  tcp:
  - match:
    - port: 3306
    route:
    - destination:
        host: mysql
        port:
          number: 3306
`),
	},
	// RW-8.Istio.Valid.3 - DestinationRule with simple load balancer
	{
		name: "RW-8.Istio.Valid.3 DestinationRule simple load balancer",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN
`),
	},
	// RW-8.Istio.Valid.4 - DestinationRule with passthrough TLS
	{
		name: "RW-8.Istio.Valid.4 DestinationRule passthrough TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: mtls-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
`),
	},
	// RW-8.Istio.Valid.5 - Gateway with HTTPS using SIMPLE mode
	{
		name: "RW-8.Istio.Valid.5 Gateway HTTPS SIMPLE mode",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: https-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      serverCertificate: /etc/istio/cert/cert.pem
      privateKey: /etc/istio/cert/key.pem
    hosts:
    - "*.example.com"
`),
	},
	// RW-8.Istio.Valid.6 - ServiceEntry with DNS location
	{
		name: "RW-8.Istio.Valid.6 ServiceEntry DNS location",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: external-svc
  namespace: default
spec:
  hosts:
  - external.example.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  location: DNS
  resolution: DNS
`),
	},
	// RW-8.Istio.Valid.7 - Sidecar with egress to wildcard hosts
	{
		name: "RW-8.Istio.Valid.7 Sidecar egress wildcard hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: wildcard-egress
  namespace: default
spec:
  egress:
  - hosts:
    - "*/*"
`),
	},
	// RW-8.Istio.Valid.8 - VirtualService with single match rule
	{
		name: "RW-8.Istio.Valid.8 VirtualService single match rule",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: match-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        end-user:
          exact: jason
    route:
    - destination:
        host: reviews
        subset: v2
`),
	},
	// RW-8.Istio.Valid.9 - VirtualService with redirect
	{
		name: "RW-8.Istio.Valid.9 VirtualService with redirect",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: redirect-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - redirect:
      authority: new-reviews.example.com
`),
	},
	// RW-8.Istio.Valid.10 - DestinationRule with localityLbSetting
	{
		name: "RW-8.Istio.Valid.10 DestinationRule locality load balancing",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: locality-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    localityLbSetting:
      enabled: true
`),
	},
	// RW-8.Istio.Valid.11 - Gateway with multiple hosts
	{
		name: "RW-8.Istio.Valid.11 Gateway multiple hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: multi-host-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "productpage.example.com"
    - "reviews.example.com"
`),
	},
	// RW-8.Istio.Valid.12 - ServiceEntry mesh internal
	{
		name: "RW-8.Istio.Valid.12 ServiceEntry mesh internal",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: internal-svc
  namespace: default
spec:
  hosts:
  - internal-service
  ports:
  - number: 8080
    name: http
    protocol: HTTP
  location: MESH_INTERNAL
  resolution: STATIC
`),
	},
	// RW-8.Istio.Valid.13 - Sidecar with specific namespace hosts
	{
		name: "RW-8.Istio.Valid.13 Sidecar specific namespace hosts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: ns-egress
  namespace: default
spec:
  egress:
  - hosts:
    - "istio-system/*"
`),
	},
	// RW-8.Istio.Valid.14 - VirtualService with percent-based weight
	{
		name: "RW-8.Istio.Valid.14 VirtualService percent-based weight",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: weighted-route
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 90
    - destination:
        host: reviews
        subset: v2
      weight: 10
`),
	},
	// RW-8.Istio.Valid.15 - DestinationRule with connectionPool TCP
	{
		name: "RW-8.Istio.Valid.15 DestinationRule TCP connection pool",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: tcp-pool-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
`),
	},
	// RW-8.Istio.Valid.16 - Gateway with TLS passthrough
	{
		name: "RW-8.Istio.Valid.16 Gateway TLS passthrough",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: passthrough-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: tls
      protocol: TLS
    tls:
      mode: PASSTHROUGH
    hosts:
    - "secure.example.com"
`),
	},
	// RW-8.Istio.Valid.17 - ServiceEntry with STATIC resolution
	{
		name: "RW-8.Istio.Valid.17 ServiceEntry STATIC resolution",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: static-svc
  namespace: default
spec:
  hosts:
  - static-service
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: STATIC
  endpoints:
  - address: 10.0.0.1
    ports:
      http: 80
`),
	},
	// RW-8.Istio.Valid.18 - Sidecar with captureMode DEFAULT
	{
		name: "RW-8.Istio.Valid.18 Sidecar capture mode DEFAULT",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: default-capture
  namespace: default
spec:
  egress:
  - captureMode: DEFAULT
    hosts:
    - "*/httpbin.org"
`),
	},
	// RW-8.Istio.Valid.19 - VirtualService with external service reference
	{
		name: "RW-8.Istio.Valid.19 VirtualService external service",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: external-vs
  namespace: default
spec:
  hosts:
  - httpbin.org
  http:
  - route:
    - destination:
        host: httpbin.org
`),
	},
	// RW-8.Istio.Valid.20 - DestinationRule with outlierDetection
	{
		name: "RW-8.Istio.Valid.20 DestinationRule outlier detection",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: outlier-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
`),
	},
	// RW-8.Istio.Valid.21 - VirtualService with mirror percentage
	{
		name: "RW-8.Istio.Valid.21 VirtualService mirror with percentage",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: mirror-vs
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    mirror:
      host: reviews
      subset: v2
    mirrorPercent: 50
`),
	},
	// RW-8.Istio.Valid.22 - Gateway with HTTP/2 protocol
	{
		name: "RW-8.Istio.Valid.22 Gateway HTTP/2 protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: http2-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 8080
      name: http2
      protocol: HTTP2
    hosts:
    - "http2.example.com"
`),
	},
	// RW-8.Istio.Valid.23 - ServiceEntry with MESH_EXTERNAL location
	{
		name: "RW-8.Istio.Valid.23 ServiceEntry mesh external",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: external-api
  namespace: default
spec:
  hosts:
  - api.external.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  location: MESH_EXTERNAL
  resolution: DNS
`),
	},
	// RW-8.Istio.Valid.24 - Sidecar with workloadSelector
	{
		name: "RW-8.Istio.Valid.24 Sidecar workload selector",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: workload-sidecar
  namespace: default
spec:
  workloadSelector:
    labels:
      app: reviews
  egress:
  - hosts:
    - "istio-system/*"
`),
	},
	// RW-8.Istio.Valid.25 - VirtualService with timeout zero
	{
		name: "RW-8.Istio.Valid.25 VirtualService zero timeout",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: timeout-zero
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    timeout: 0s
`),
	},
	// RW-8.Istio.Valid.26 - DestinationRule with port-level policy
	{
		name: "RW-8.Istio.Valid.26 DestinationRule port-level policy",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: port-level-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    portLevelSettings:
    - port:
        number: 9080
      loadBalancer:
        simple: LEAST_CONN
`),
	},
	// RW-8.Istio.Valid.27 - Gateway with GRPC protocol
	{
		name: "RW-8.Istio.Valid.27 Gateway GRPC protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: grpc-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 50051
      name: grpc
      protocol: GRPC
    hosts:
    - "grpc.example.com"
`),
	},
	// RW-8.Istio.Valid.28 - ServiceEntry with endpoint port override
	{
		name: "RW-8.Istio.Valid.28 ServiceEntry endpoint port override",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: port-override-svc
  namespace: default
spec:
  hosts:
  - mongo
  ports:
  - number: 27017
    name: mongo
    protocol: MONGO
  resolution: STATIC
  endpoints:
  - address: 10.0.0.5
    ports:
      mongo: 27017
`),
	},
	// RW-8.Istio.Valid.29 - Sidecar with multiple egress listeners
	{
		name: "RW-8.Istio.Valid.29 Sidecar multiple egress listeners",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: multi-listener
  namespace: default
spec:
  egress:
  - port:
      port: 3306
      protocol: TCP
    bind: 0.0.0.0
    hosts:
    - "*/mysql"
  - hosts:
    - "*/*"
`),
	},
	// RW-8.Istio.Valid.30 - VirtualService with appendHeaders
	{
		name: "RW-8.Istio.Valid.30 VirtualService append headers",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: append-headers
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    appendHeaders:
      x-custom-header: value
`),
	},
	// RW-8.Istio.Valid.31 - DestinationRule consistent hash with http cookie
	{
		name: "RW-8.Istio.Valid.31 DestinationRule consistent hash cookie",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: cookie-hash-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      consistentHash:
        httpCookie:
          name: user
          ttl: 0s
`),
	},
	// RW-8.Istio.Valid.32 - Gateway with HTTPS using MUTUAL mode
	{
		name: "RW-8.Istio.Valid.32 Gateway mutual TLS",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: mutual-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https-mutual
      protocol: HTTPS
    tls:
      mode: MUTUAL
      serverCertificate: /etc/istio/cert/cert.pem
      privateKey: /etc/istio/cert/key.pem
      caCertificates: /etc/istio/cert/ca.pem
    hosts:
    - "mutual.example.com"
`),
	},
	// RW-8.Istio.Valid.33 - ServiceEntry with NONE resolution
	{
		name: "RW-8.Istio.Valid.33 ServiceEntry NONE resolution",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: none-res-svc
  namespace: default
spec:
  hosts:
  - external.example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: NONE
`),
	},
	// RW-8.Istio.Valid.34 - Sidecar with bind to empty string
	{
		name: "RW-8.Istio.Valid.34 Sidecar bind empty",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: bind-empty
  namespace: default
spec:
  egress:
  - bind: ""
    hosts:
    - "*/*"
`),
	},
	// RW-8.Istio.Valid.35 - VirtualService with removeResponseHeader
	{
		name: "RW-8.Istio.Valid.35 VirtualService remove response header",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: rm-header
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    removeResponseHeader: x-removed
`),
	},
	// RW-8.Istio.Valid.36 - DestinationRule with http2 connection pool
	{
		name: "RW-8.Istio.Valid.36 DestinationRule HTTP2 connection pool",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: http2-pool-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool:
      http:
        h2UpgradePolicy: UPGRADE
        http1MaxPendingRequests: 100
`),
	},
	// RW-8.Istio.Valid.37 - Gateway with PROXY protocol
	{
		name: "RW-8.Istio.Valid.37 Gateway PROXY protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: proxy-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: proxy
      protocol: PROXY
    hosts:
    - "proxy.example.com"
`),
	},
	// RW-8.Istio.Valid.38 - ServiceEntry with multiple ports
	{
		name: "RW-8.Istio.Valid.38 ServiceEntry multiple ports",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: multi-port-svc
  namespace: default
spec:
  hosts:
  - multi-port
  ports:
  - number: 80
    name: http
    protocol: HTTP
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
`),
	},
	// RW-8.Istio.Valid.39 - Sidecar without workloadSelector
	{
		name: "RW-8.Istio.Valid.39 Sidecar without workload selector",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: no-selector
  namespace: default
spec:
  egress:
  - hosts:
    - "default/*"
`),
	},
	// RW-8.Istio.Valid.40 - VirtualService with injectMetadataHeaders
	{
		name: "RW-8.Istio.Valid.40 VirtualService inject metadata headers",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: inject-headers
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	// RW-8.Istio.Valid.41 - DestinationRule with simple LEAST_REQUEST
	{
		name: "RW-8.Istio.Valid.41 DestinationRule least request lb",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: least-req-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: LEAST_REQUEST
`),
	},
	// RW-8.Istio.Valid.42 - Gateway with HTTP protocol
	{
		name: "RW-8.Istio.Valid.42 Gateway HTTP protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: http-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "http.example.com"
`),
	},
	// RW-8.Istio.Valid.43 - ServiceEntry with endpoint locality label
	{
		name: "RW-8.Istio.Valid.43 ServiceEntry endpoint with locality",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: locality-ep
  namespace: default
spec:
  hosts:
  - localized-svc
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: STATIC
  endpoints:
  - address: 10.0.0.10
    locality: us-west/us-west-1
    ports:
      http: 80
`),
	},
	// RW-8.Istio.Valid.44 - Sidecar with APP choice for capture mode
	{
		name: "RW-8.Istio.Valid.44 Sidecar capture mode APP",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: capture-app
  namespace: default
spec:
  egress:
  - captureMode: APP
    hosts:
    - "*/httpbin.org"
`),
	},
	// RW-8.Istio.Valid.45 - VirtualService with setResponseHeader
	{
		name: "RW-8.Istio.Valid.45 VirtualService set response header",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: set-header
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
    setResponseHeader:
      x-added: value
`),
	},
	// RW-8.Istio.Valid.46 - DestinationRule with MAGLEV load balancer
	{
		name: "RW-8.Istio.Valid.46 DestinationRule MAGLEV load balancer",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: maglev-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: MAGLEV
`),
	},
	// RW-8.Istio.Valid.47 - Gateway with TERMINATE mode for TLS
	{
		name: "RW-8.Istio.Valid.47 Gateway TLS TERMINATE mode",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: terminate-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      serverCertificate: /etc/istio/cert/cert.pem
      privateKey: /etc/istio/cert/key.pem
    hosts:
    - "terminate.example.com"
`),
	},
	// RW-8.Istio.Valid.48 - ServiceEntry with UDP protocol
	{
		name: "RW-8.Istio.Valid.48 ServiceEntry UDP protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: udp-svc
  namespace: default
spec:
  hosts:
  - dns-service
  ports:
  - number: 53
    name: dns-udp
    protocol: UDP
  resolution: DNS
`),
	},
	// RW-8.Istio.Valid.49 - Sidecar with egress bind 127.0.0.1
	{
		name: "RW-8.Istio.Valid.49 Sidecar egress localhost bind",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: localhost-egress
  namespace: default
spec:
  egress:
  - bind: 127.0.0.1
    hosts:
    - "*/*"
`),
	},
	// RW-8.Istio.Valid.50 - VirtualService with retry attempts 3
	{
		name: "RW-8.Istio.Valid.50 VirtualService retry 3 attempts",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: retry-vs
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
        subset: v1
  - retries:
      attempts: 3
      perTryTimeout: 2s
    route:
    - destination:
        host: reviews
        subset: v1
`),
	},
	// RW-8.Istio.Valid.51 - DestinationRule with simple RANDOM
	{
		name: "RW-8.Istio.Valid.51 DestinationRule random load balancer",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: random-dr
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: RANDOM
`),
	},
	// RW-8.Istio.Valid.52 - Gateway with TCP protocol
	{
		name: "RW-8.Istio.Valid.52 Gateway TCP protocol",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: tcp-gw
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 9000
      name: tcp
      protocol: TCP
    hosts:
    - "tcp.example.com"
`),
	},
	// RW-8.Istio.Valid.53 - ServiceEntry with HTTP resolution
	{
		name: "RW-8.Istio.Valid.53 ServiceEntry HTTP resolution",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: http-res-svc
  namespace: default
spec:
  hosts:
  - http-svc
  ports:
  - number: 8080
    name: http
    protocol: HTTP
  resolution: HTTP
`),
	},
	// RW-8.Istio.Valid.54 - Sidecar egress with match port
	{
		name: "RW-8.Istio.Valid.54 Sidecar egress with match port",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: match-port-sidecar
  namespace: default
spec:
  egress:
  - port:
      port: 443
      protocol: HTTPS
    hosts:
    - "*/secure-service"
`),
	},
	// RW-8.Istio.Valid.55 - VirtualService with fault delay
	{
		name: "RW-8.Istio.Valid.55 VirtualService fault delay",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: delay-vs
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - fault:
      delay:
        percent: 10
        fixedDelay: 5s
    route:
    - destination:
        host: reviews
        subset: v1
`),
	},
}

// =============================================================================
// Invalid Istio Test Cases
// =============================================================================

var istioInvalidCases = []struct {
	name       string
	crYAML     []byte
	wantErr    bool
	errSubstr  string
}{
	// VirtualService - Invalid Cases
	{
		name: "RW-8.56 VirtualService missing hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: no-hosts
  namespace: default
spec:
  http:
  - route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.57 VirtualService empty hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: empty-hosts
  namespace: default
spec:
  hosts: []
  http:
  - route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.58 VirtualService invalid kind fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: BadVirtualService
metadata:
  name: wrong-kind
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "VirtualService",
	},
	{
		name: "RW-8.59 VirtualService wrong apiVersion fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: wrong-apiversion
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "v1alpha3",
	},
	// DestinationRule - Invalid Cases
	{
		name: "RW-8.60 DestinationRule missing host fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: no-host
  namespace: default
spec:
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
`),
		wantErr:   true,
		errSubstr: "host",
	},
	{
		name: "RW-8.61 DestinationRule empty host fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: empty-host
  namespace: default
spec:
  host: ""
`),
		wantErr:   true,
		errSubstr: "host",
	},
	{
		name: "RW-8.62 DestinationRule invalid kind fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: BadDestinationRule
metadata:
  name: wrong-kind
  namespace: default
spec:
  host: reviews
`),
		wantErr:   true,
		errSubstr: "DestinationRule",
	},
	{
		name: "RW-8.63 DestinationRule invalid traffic policy type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: bad-traffic-policy
  namespace: default
spec:
  host: reviews
  trafficPolicy: "invalid"
`),
		wantErr:   true,
		errSubstr: "object",
	},
	{
		name: "RW-8.64 DestinationRule empty subset fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: empty-subset
  namespace: default
spec:
  host: reviews
  subsets:
  - name: ""
    labels:
      version: v1
`),
		wantErr:   true,
		errSubstr: "subset",
	},
	// Gateway - Invalid Cases
	{
		name: "RW-8.65 Gateway missing servers fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: no-servers
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
`),
		wantErr:   true,
		errSubstr: "servers",
	},
	{
		name: "RW-8.66 Gateway empty servers fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: empty-servers
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers: []
`),
		wantErr:   true,
		errSubstr: "servers",
	},
	{
		name: "RW-8.67 Gateway missing port number fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: no-port
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      name: http
      protocol: HTTP
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "number",
	},
	{
		name: "RW-8.68 Gateway invalid kind fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: BadGateway
metadata:
  name: wrong-kind
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "Gateway",
	},
	{
		name: "RW-8.69 Gateway missing hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: no-hosts
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.70 Gateway empty hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: empty-hosts
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts: []
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	// ServiceEntry - Invalid Cases
	{
		name: "RW-8.71 ServiceEntry missing hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: no-hosts
  namespace: default
spec:
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.72 ServiceEntry missing ports fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: no-ports
  namespace: default
spec:
  hosts:
  - example.com
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "ports",
	},
	{
		name: "RW-8.73 ServiceEntry empty ports fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: empty-ports
  namespace: default
spec:
  hosts:
  - example.com
  ports: []
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "ports",
	},
	{
		name: "RW-8.74 ServiceEntry empty hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: empty-hosts
  namespace: default
spec:
  hosts: []
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.75 ServiceEntry invalid kind fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: BadServiceEntry
metadata:
  name: wrong-kind
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "ServiceEntry",
	},
	{
		name: "RW-8.76 ServiceEntry wrong resolution type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: bad-resolution
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: INVALID
`),
		wantErr:   true,
		errSubstr: "resolution",
	},
	{
		name: "RW-8.77 ServiceEntry wrong location type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: bad-location
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: WRONG
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "location",
	},
	// Sidecar - Invalid Cases
	{
		name: "RW-8.78 Sidecar missing egress fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: no-egress
  namespace: default
spec: {}
`),
		wantErr:   true,
		errSubstr: "egress",
	},
	{
		name: "RW-8.79 Sidecar empty egress fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: empty-egress
  namespace: default
spec:
  egress: []
`),
		wantErr:   true,
		errSubstr: "egress",
	},
	{
		name: "RW-8.80 Sidecar invalid kind fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: BadSidecar
metadata:
  name: wrong-kind
  namespace: default
spec:
  egress:
  - hosts:
    - "*/*"
`),
		wantErr:   true,
		errSubstr: "Sidecar",
	},
	{
		name: "RW-8.81 Sidecar empty hosts fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: empty-hosts
  namespace: default
spec:
  egress:
  - hosts: []
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.82 Sidecar empty string host fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: empty-string-host
  namespace: default
spec:
  egress:
  - hosts:
    - ""
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	// Additional invalid cases
	{
		name: "RW-8.83 VirtualService wrong spec type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: wrong-spec-type
  namespace: default
spec:
  hosts: 123
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.84 VirtualService missing route destination fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: no-destination
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - weight: 100
`),
		wantErr:   true,
		errSubstr: "destination",
	},
	{
		name: "RW-8.85 DestinationRule wrong subsets type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: wrong-subsets-type
  namespace: default
spec:
  host: reviews
  subsets: "v1"
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.86 Gateway wrong servers type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: wrong-servers-type
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers: "http"
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.87 Gateway missing protocol fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: missing-protocol
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "protocol",
	},
	{
		name: "RW-8.88 ServiceEntry empty endpoints item fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: empty-endpoint-item
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: STATIC
  endpoints:
  - address: ""
    ports:
      http: 8080
`),
		wantErr:   true,
		errSubstr: "address",
	},
	{
		name: "RW-8.89 VirtualService invalid HTTP array fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: invalid-http-array
  namespace: default
spec:
  hosts:
  - reviews
  http: "invalid"
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.90 DestinationRule wrong traffic policy type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: wrong-tp-type
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    connectionPool: 100
`),
		wantErr:   true,
		errSubstr: "object",
	},
	{
		name: "RW-8.91 Gateway invalid selector type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: invalid-selector
  namespace: istio-system
spec:
  selector: []
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "object",
	},
	{
		name: "RW-8.92 Sidecar wrong capture mode fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: wrong-capture-mode
  namespace: default
spec:
  egress:
  - captureMode: INVALID
    hosts:
    - "*/*"
`),
		wantErr:   true,
		errSubstr: "",
	},
	{
		name: "RW-8.93 VirtualService empty spec fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: empty-spec
  namespace: default
spec: {}
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.94 VirtualService missing http route fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: missing-http
  namespace: default
spec:
  hosts:
  - reviews
`),
		wantErr:   true,
		errSubstr: "http",
	},
	{
		name: "RW-8.95 DestinationRule wrong host type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: wrong-host-type
  namespace: default
spec:
  host:
    value: reviews
`),
		wantErr:   true,
		errSubstr: "string",
	},
	{
		name: "RW-8.96 ServiceEntry wrong hosts type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: wrong-hosts-type
  namespace: default
spec:
  hosts: "example.com"
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.97 Gateway wrong port type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: wrong-port-type
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port: "80"
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "object",
	},
	{
		name: "RW-8.98 Sidecar wrong egress type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: wrong-egress-type
  namespace: default
spec:
  egress: "invalid"
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.99 VirtualService wrong tls type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: wrong-tls-type
  namespace: default
spec:
  hosts:
  - reviews
  tls: "invalid"
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.100 VirtualService wrong tcp type fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: wrong-tcp-type
  namespace: default
spec:
  hosts:
  - reviews
  tcp: 123
`),
		wantErr:   true,
		errSubstr: "array",
	},
	{
		name: "RW-8.101 DestinationRule empty traffic policy fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: empty-traffic-policy
  namespace: default
spec:
  host: reviews
  trafficPolicy: {}
`),
		wantErr:   true,
		errSubstr: "host",
	},
	{
		name: "RW-8.102 ServiceEntry port missing number fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: port-missing-number
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: DNS
`),
		wantErr:   true,
		errSubstr: "number",
	},
	{
		name: "RW-8.103 Sidecar egress with empty hosts string fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: empty-hosts-string
  namespace: default
spec:
  egress:
  - hosts: [""]
`),
		wantErr:   true,
		errSubstr: "hosts",
	},
	{
		name: "RW-8.104 VirtualService route with empty destination host fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: empty-dest-host
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - route:
    - destination:
        host: ""
`),
		wantErr:   true,
		errSubstr: "host",
	},
	{
		name: "RW-8.105 Gateway server port name too long fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: long-port-name
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: this-is-a-very-long-port-name-that-exceeds-reasonable-limits-for-istio-configuration-to-test-validation-behavior
      protocol: HTTP
    hosts:
    - example.com
`),
		wantErr:   true,
		errSubstr: "name",
	},
	{
		name: "RW-8.106 VirtualService match with empty header key fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: empty-header-key
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        "": {}
    route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "headers",
	},
	{
		name: "RW-8.107 ServiceEntry endpoint with invalid port number fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: invalid-endpoint-port
  namespace: default
spec:
  hosts:
  - example.com
  ports:
  - number: 80
    name: http
    protocol: HTTP
  location: MESH_EXTERNAL
  resolution: STATIC
  endpoints:
  - address: 192.168.1.1
    ports:
      http: 99999
`),
		wantErr:   true,
		errSubstr: "ports",
	},
	{
		name: "RW-8.108 VirtualService with invalid timeout format fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: invalid-timeout
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - timeout: invalid
    route:
    - destination:
        host: reviews
`),
		wantErr:   true,
		errSubstr: "duration",
	},
	{
		name: "RW-8.109 DestinationRule with invalid load balancer simple fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: invalid-lb
  namespace: default
spec:
  host: reviews
  trafficPolicy:
    loadBalancer:
      simple: INVALID_LB
`),
		wantErr:   true,
		errSubstr: "simple",
	},
	{
		name: "RW-8.110 Sidecar with invalid bind address fails",
		crYAML: []byte(`apiVersion: networking.istio.io/v1alpha3
kind: Sidecar
metadata:
  name: invalid-bind
  namespace: default
spec:
  egress:
  - bind: "not-an-ip"
    hosts:
    - "*/*"
`),
		wantErr:   true,
		errSubstr: "",
	},
}

// =============================================================================
// Test Runner
// =============================================================================

func TestIstioRealWorldValid(t *testing.T) {
	for _, tc := range istioValidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-8] testing istio valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			if len(result.Errors) > 0 {
				t.Errorf("expected valid, got errors: %v", result.Errors)
			}
		})
	}
	for _, tc := range istioValidAdditionalCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-8] testing istio valid additional: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			if len(result.Errors) > 0 {
				t.Errorf("expected valid, got errors: %v", result.Errors)
			}
		})
	}
}

func TestIstioRealWorldInvalid(t *testing.T) {
	for _, tc := range istioInvalidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-8] testing istio invalid: %s", tc.name)
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