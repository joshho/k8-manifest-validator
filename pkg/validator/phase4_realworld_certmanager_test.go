package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

// =============================================================================
// RW-4 — cert-manager Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — cert-manager operator group
// Owner file: pkg/validator/phase4_realworld_certmanager_test.go
// CRDs covered: Certificate, Issuer, ClusterIssuer, CertificateRequest
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// cert-manager Operator CRD definitions (inline for test isolation)
// Based on cert-manager v1.14+ CRD schemas
// -----------------------------------------------------------------------

const certificateCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: certificates.cert-manager.io
spec:
  group: cert-manager.io
  names:
    kind: Certificate
    plural: certificates
    singular: certificate
    listKind: CertificateList
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
              secretName:
                type: string
              secretTemplate:
                type: object
                properties:
                  labels:
                    type: object
                    additionalProperties:
                      type: string
                  annotations:
                    type: object
                    additionalProperties:
                      type: string
              issuerRef:
                type: object
                properties:
                  name:
                    type: string
                  kind:
                    type: string
                  group:
                    type: string
                required:
                - name
              commonName:
                type: string
              duration:
                type: string
              renewBefore:
                type: string
              dnsNames:
                type: array
                items:
                  type: string
              ipAddresses:
                type: array
                items:
                  type: string
              emailSANs:
                type: array
                items:
                  type: string
              uriSANs:
                type: array
                items:
                  type: string
              usages:
                type: array
                items:
                  type: string
              privateKey:
                type: object
                properties:
                  rotationPolicy:
                    type: string
                  encoding:
                    type: string
                  algorithm:
                    type: string
                  size:
                    type: integer
              encodeUsagesInRequest:
                type: boolean
              nameConstraints:
                type: object
                properties:
                  permitted:
                    type: object
                    properties:
                      dns:
                        type: array
                        items:
                          type: string
                      email:
                        type: array
                        items:
                          type: string
                      uri:
                        type: array
                        items:
                          type: string
                  excluded:
                    type: object
              revisionHistoryLimit:
                type: integer
            required:
            - secretName
            - issuerRef
          status:
            type: object
            properties:
              conditions:
                type: array
              nextPrivateKeySecretName:
                type: string
              secretName:
                type: string
`

const issuerCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: issuers.cert-manager.io
spec:
  group: cert-manager.io
  names:
    kind: Issuer
    plural: issuers
    singular: issuer
    listKind: IssuerList
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
              acme:
                type: object
                properties:
                  server:
                    type: string
                  email:
                    type: string
                  privateKeySecretRef:
                    type: object
                    properties:
                      name:
                        type: string
                      key:
                        type: string
                    required:
                    - name
                  solvers:
                    type: array
                    items:
                      type: object
                      properties:
                        dns01:
                          type: object
                          properties:
                            route53:
                              type: object
                              properties:
                                region:
                                  type: string
                                hostedZoneID:
                                  type: string
                                clusterDNSName:
                                  type: string
                            cloudflare:
                              type: object
                              properties:
                                email:
                                  type: string
                                apiTokenSecretRef:
                                  type: object
                                  properties:
                                    name:
                                      type: string
                                    key:
                                      type: string
                            http01:
                              type: object
                              properties:
                                ingressClassName:
                                  type: string
                                serviceType:
                                  type: string
                        selector:
                          type: object
                          properties:
                            dnsZones:
                              type: array
                              items:
                                type: string
                            matchLabels:
                              type: object
                              additionalProperties:
                                type: string
                  externalAccountBinding:
                    type: object
                  preferredChain:
                    type: string
              ca:
                type: object
                properties:
                  secretName:
                    type: string
                  crlDistributionPoints:
                    type: array
                    items:
                      type: string
              vault:
                type: object
                properties:
                  path:
                    type: string
                  server:
                    type: string
                  caBundle:
                    type: string
                  auth:
                    type: object
                    properties:
                      tokenSecretRef:
                        type: object
                        properties:
                          name:
                            type: string
                          key:
                            type: string
                      appRole:
                        type: object
                        properties:
                          secretRef:
                            type: object
                            properties:
                              name:
                                type: string
                              key:
                                type: string
              selfSigned:
                type: object
                properties:
                  crlDistributionPoints:
                    type: array
                    items:
                      type: string
              venafi:
                type: object
                properties:
                  zone:
                    type: string
                  url:
                    type: string
                  apiTokenSecretRef:
                    type: object
                    properties:
                      name:
                        type: string
                      key:
                        type: string
          status:
            type: object
            properties:
              conditions:
                type: array
`

const clusterIssuerCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusterissuers.cert-manager.io
spec:
  group: cert-manager.io
  names:
    kind: ClusterIssuer
    plural: clusterissuers
    singular: clusterissuer
    listKind: ClusterIssuerList
  scope: Cluster
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
              acme:
                type: object
                properties:
                  server:
                    type: string
                  email:
                    type: string
                  privateKeySecretRef:
                    type: object
                    properties:
                      name:
                        type: string
                      key:
                        type: string
                    required:
                    - name
                  solvers:
                    type: array
                    items:
                      type: object
                      properties:
                        dns01:
                          type: object
                          properties:
                            route53:
                              type: object
                              properties:
                                region:
                                  type: string
                                hostedZoneID:
                                  type: string
                            cloudflare:
                              type: object
                              properties:
                                email:
                                  type: string
                                apiTokenSecretRef:
                                  type: object
                                  properties:
                                    name:
                                      type: string
                                    key:
                                      type: string
                            http01:
                              type: object
                              properties:
                                ingressClassName:
                                  type: string
                        selector:
                          type: object
                          properties:
                            dnsZones:
                              type: array
                              items:
                                type: string
                            matchLabels:
                              type: object
                              additionalProperties:
                                type: string
                  externalAccountBinding:
                    type: object
                  preferredChain:
                    type: string
              ca:
                type: object
                properties:
                  secretName:
                    type: string
              vault:
                type: object
                properties:
                  path:
                    type: string
                  server:
                    type: string
              selfSigned:
                type: object
              venafi:
                type: object
                properties:
                  zone:
                    type: string
                  url:
                    type: string
                  apiTokenSecretRef:
                    type: object
                    properties:
                      name:
                        type: string
                      key:
                        type: string
          status:
            type: object
            properties:
              conditions:
                type: array
`

const certificateRequestCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: certificaterequests.cert-manager.io
spec:
  group: cert-manager.io
  names:
    kind: CertificateRequest
    plural: certificaterequests
    singular: certificate-request
    listKind: CertificateRequestList
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
              issuerRef:
                type: object
                properties:
                  name:
                    type: string
                  kind:
                    type: string
                  group:
                    type: string
                required:
                - name
              commonName:
                type: string
              dnsNames:
                type: array
                items:
                  type: string
              ipAddresses:
                type: array
                items:
                  type: string
              emailSANs:
                type: array
                items:
                  type: string
              uriSANs:
                type: array
                items:
                  type: string
              usages:
                type: array
                items:
                  type: string
              request:
                type: string
                format: byte
              isCA:
                type: boolean
              encodeUsagesInRequest:
                type: boolean
              revision:
                type: integer
            required:
            - issuerRef
            - request
          status:
            type: object
            properties:
              conditions:
                type: array
              ca:
                type: string
                format: byte
              certificate:
                type: string
                format: byte
`

// -----------------------------------------------------------------------
// Table-driven test cases
// Pattern: name, crYAML (manifest under test), wantErr, errSubstr
// -----------------------------------------------------------------------

func TestCertManagerRealWorld(t *testing.T) {
	testCases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// =========================================================================
		// VALID CERTIFICATE TESTS (1-20)
		// =========================================================================
		{
			name: "RW-4.CertManager.Valid.1 Simple Certificate with secretName and issuerRef",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-cert
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  dnsNames:
    - example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.2 Certificate with commonName",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-cert-cn
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  commonName: example.com
  dnsNames:
    - example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.3 Certificate with multiple dnsNames",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: multi-dns-cert
  namespace: default
spec:
  secretName: multi-dns-tls
  issuerRef:
    name: my-issuer
  dnsNames:
    - example.com
    - www.example.com
    - api.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.4 Certificate with ipAddresses",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ip-cert
  namespace: default
spec:
  secretName: ip-cert-tls
  issuerRef:
    name: my-issuer
  ipAddresses:
    - "10.0.0.1"
    - "192.168.1.1"
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.5 Certificate with uriSANs",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: uri-cert
  namespace: default
spec:
  secretName: uri-cert-tls
  issuerRef:
    name: my-issuer
  uriSANs:
    - "spiffe://example.com/ns/default/sa/default"
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.6 Certificate with issuerRef kind ClusterIssuer",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: cluster-issuer-cert
  namespace: default
spec:
  secretName: cluster-issuer-tls
  issuerRef:
    name: my-cluster-issuer
    kind: ClusterIssuer
  dnsNames:
    - cluster.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.7 Certificate with issuerRef group",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: group-issuer-cert
  namespace: default
spec:
  secretName: group-issuer-tls
  issuerRef:
    name: my-issuer
    kind: Issuer
    group: cert-manager.io
  dnsNames:
    - group.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.8 Certificate with duration and renewBefore",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: duration-cert
  namespace: default
spec:
  secretName: duration-tls
  issuerRef:
    name: my-issuer
  duration: 2160h
  renewBefore: 360h
  dnsNames:
    - duration.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.9 Certificate with usages digital signature",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: usages-cert
  namespace: default
spec:
  secretName: usages-tls
  issuerRef:
    name: my-issuer
  usages:
    - digital signature
    - key encipherment
  dnsNames:
    - usages.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.10 Certificate with secretTemplate labels",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: template-cert
  namespace: default
spec:
  secretName: template-tls
  issuerRef:
    name: my-issuer
  secretTemplate:
    labels:
      environment: production
      team: platform
  dnsNames:
    - template.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.11 Certificate with secretTemplate annotations",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: template-anno-cert
  namespace: default
spec:
  secretName: template-anno-tls
  issuerRef:
    name: my-issuer
  secretTemplate:
    annotations:
      example.com/owner: platform-team
      example.com/version: "1.0"
  dnsNames:
    - template.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.12 Certificate with emailSANs",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: email-san-cert
  namespace: default
spec:
  secretName: email-san-tls
  issuerRef:
    name: my-issuer
  emailSANs:
    - admin@example.com
    - support@example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.13 Certificate with revisionHistoryLimit",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: revision-cert
  namespace: default
spec:
  secretName: revision-tls
  issuerRef:
    name: my-issuer
  revisionHistoryLimit: 5
  dnsNames:
    - revision.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.14 Certificate with privateKey encoding RSA",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: pk-cert
  namespace: default
spec:
  secretName: pk-tls
  issuerRef:
    name: my-issuer
  privateKey:
    encoding: PKCS1
    algorithm: RSA
    size: 2048
  dnsNames:
    - pk.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.15 Certificate with privateKey rotationPolicy",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: pk-rotation-cert
  namespace: default
spec:
  secretName: pk-rotation-tls
  issuerRef:
    name: my-issuer
  privateKey:
    rotationPolicy: Always
    encoding: PKCS8
    algorithm: ECDSAP256
  dnsNames:
    - pk-rotation.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.16 Certificate with encodeUsagesInRequest false",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: encode-false-cert
  namespace: default
spec:
  secretName: encode-false-tls
  issuerRef:
    name: my-issuer
  encodeUsagesInRequest: false
  dnsNames:
    - encode.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.17 Certificate with nameConstraints permitted dns",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: nc-cert
  namespace: default
spec:
  secretName: nc-tls
  issuerRef:
    name: my-issuer
  nameConstraints:
    permitted:
      dns:
        - example.com
        - "*.example.com"
  dnsNames:
    - nc.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.18 Issuer with ACME server and email",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: acme-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - dns01:
        route53:
          region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.19 Issuer with ACME HTTP01 solver",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: http01-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - selector:
        dnsZones:
          - example.com
      dns01:
        http01:
          ingressClassName: nginx
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.20 Issuer with CA secretName",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: ca-issuer
  namespace: default
spec:
  ca:
    secretName: ca-key-pair
`),
			wantErr: false,
		},
		// =========================================================================
		// VALID TESTS CONTINUED (21-40)
		// =========================================================================
		{
			name: "RW-4.CertManager.Valid.21 ClusterIssuer with ACME",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod-account-key
    solvers:
    - dns01:
        route53:
          region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.22 ClusterIssuer with Cloudflare DNS",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cloudflare-issuer
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: cloudflare-account-key
    solvers:
    - dns01:
        cloudflare:
          email: admin@example.com
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-key
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.23 CertificateRequest with basic spec",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  dnsNames:
    - csr.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.24 CertificateRequest with commonName",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-cn-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  commonName: csr.example.com
  request: TESTREQUESTBASE64==
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.25 CertificateRequest with usages",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-usages-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  usages:
    - digital signature
    - key encipherment
    - server auth
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.26 CertificateRequest with isCA true",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-ca-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  isCA: true
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.27 CertificateRequest with ipAddresses",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-ip-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  ipAddresses:
    - "10.0.0.1"
    - "192.168.1.100"
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.28 Issuer with selfSigned",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned:
    crlDistributionPoints:
      - http://crl.example.com/crl.pem
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.29 Issuer with Venafi",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: venafi-issuer
  namespace: default
spec:
  venafi:
    zone: my-zone
    url: https://example.com/vedsdk
    apiTokenSecretRef:
      name: venafi-api-token
      key: api-key
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.30 Certificate with all SAN types combined",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: all-san-cert
  namespace: default
spec:
  secretName: all-san-tls
  issuerRef:
    name: my-issuer
  commonName: allsan.example.com
  dnsNames:
    - allsan.example.com
    - www.allsan.example.com
  ipAddresses:
    - "10.0.0.1"
  emailSANs:
    - admin@allsan.example.com
  uriSANs:
    - spiffe://allsan.example.com/sa/default
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.31 Issuer ACME with externalAccountBinding",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: eab-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    externalAccountBinding:
      keyID: test-key-id
      keySecretRef:
        name: eab-secret
        key: key
    solvers:
    - dns01:
        route53:
          region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.32 Issuer ACME with preferredChain",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: preferred-chain-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    preferredChain: ISRG
    solvers:
    - dns01:
        route53:
          region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.33 Certificate with encodeUsagesInRequest true",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: encode-true-cert
  namespace: default
spec:
  secretName: encode-true-tls
  issuerRef:
    name: my-issuer
  encodeUsagesInRequest: true
  dnsNames:
    - encode.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.34 Issuer with Vault auth tokenSecretRef",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: vault-token-issuer
  namespace: default
spec:
  vault:
    server: https://vault.example.com:8200
    path: certs/example.com
    auth:
      tokenSecretRef:
        name: vault-token
        key: token
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.35 Issuer with Vault auth appRole",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: vault-approle-issuer
  namespace: default
spec:
  vault:
    server: https://vault.example.com:8200
    path: certs/example.com
    caBundle: TESTBUNDLEBASE64
    auth:
      appRole:
        secretRef:
          name: vault-approle-secret
          key: secret-id
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.36 ClusterIssuer with CA secretName",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: my-cluster-ca
spec:
  ca:
    secretName: cluster-ca-key-pair
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.37 Certificate with privateKey algorithm ECDSAP384",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ecdsa-384-cert
  namespace: default
spec:
  secretName: ecdsa-384-tls
  issuerRef:
    name: my-issuer
  privateKey:
    algorithm: ECDSAP384
  dnsNames:
    - ecdsa.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.38 Certificate with DNS01 selector matchLabels",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: matchlabels-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - selector:
        matchLabels:
          environment: production
          team: platform
      dns01:
        route53:
          region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.39 CertificateRequest with revision",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-rev-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  revision: 1
  dnsNames:
    - rev.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.40 Issuer ACME with multiple solvers",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: multi-solver-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - selector:
        dnsZones:
          - example.com
      dns01:
        route53:
          region: us-east-1
    - selector:
        dnsZones:
          - "*.example.com"
      dns01:
        cloudflare:
          email: admin@example.com
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-key
`),
			wantErr: false,
		},
		// =========================================================================
		// VALID TESTS CONTINUED (41-55)
		// =========================================================================
		{
			name: "RW-4.CertManager.Valid.41 Certificate with nameConstraints excluded dns",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: nc-excl-cert
  namespace: default
spec:
  secretName: nc-excl-tls
  issuerRef:
    name: my-issuer
  nameConstraints:
    excluded:
      dns:
        - excluded.example.com
  dnsNames:
    - nc.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.42 Certificate with nameConstraints permitted email",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: nc-email-cert
  namespace: default
spec:
  secretName: nc-email-tls
  issuerRef:
    name: my-issuer
  nameConstraints:
    permitted:
      email:
        - example.com
  dnsNames:
    - nc-email.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.43 Certificate with nameConstraints permitted uri",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: nc-uri-cert
  namespace: default
spec:
  secretName: nc-uri-tls
  issuerRef:
    name: my-issuer
  nameConstraints:
    permitted:
      uri:
        - spiffe://example.com
  dnsNames:
    - nc-uri.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.44 Issuer with route53 hostedZoneID",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: route53-hz-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - dns01:
        route53:
          region: us-east-1
          hostedZoneID: HOSTEDZONEID123
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.45 Issuer with route53 clusterDNSName",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cluster-dns-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - dns01:
        route53:
          region: us-east-1
          clusterDNSName: cluster.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.46 Issuer with HTTP01 serviceType ClusterIP",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: http01-svc-issuer
  namespace: default
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: acme-account-key
    solvers:
    - dns01:
        http01:
          ingressClassName: nginx
          serviceType: ClusterIP
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.47 Issuer with selfSigned empty",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-empty-issuer
  namespace: default
spec:
  selfSigned: {}
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.48 ClusterIssuer with selfSigned",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cluster-selfsigned
spec:
  selfSigned: {}
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.49 ClusterIssuer with venafi",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cluster-venafi
spec:
  venafi:
    zone: Production
    url: https://example.com/vedsdk
    apiTokenSecretRef:
      name: venafi-api-token
      key: api-key
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.50 ClusterIssuer with Vault",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cluster-vault
spec:
  vault:
    server: https://vault.example.com:8200
    path: pki/certs
    auth:
      tokenSecretRef:
        name: vault-token
        key: token
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.51 Certificate with privateKey size 4096",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: pk-4096-cert
  namespace: default
spec:
  secretName: pk-4096-tls
  issuerRef:
    name: my-issuer
  privateKey:
    encoding: PKCS1
    algorithm: RSA
    size: 4096
  dnsNames:
    - pk4096.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.52 Certificate with ECDSAP256",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ecdsa256-cert
  namespace: default
spec:
  secretName: ecdsa256-tls
  issuerRef:
    name: my-issuer
  privateKey:
    algorithm: ECDSAP256
  dnsNames:
    - ecdsa256.example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.53 CertificateRequest with emailSANs",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-email-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  emailSANs:
    - admin@example.com
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.54 CertificateRequest with uriSANs",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-uri-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  uriSANs:
    - spiffe://example.com/ns/default/sa/default
`),
			wantErr: false,
		},
		{
			name: "RW-4.CertManager.Valid.55 CertificateRequest with issuerRef group cert-manager.io",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: csr-group-cert
  namespace: default
spec:
  issuerRef:
    name: my-issuer
    kind: Issuer
    group: cert-manager.io
  request: TESTREQUESTBASE64==
`),
			wantErr: false,
		},
		// =========================================================================
		// INVALID CERTIFICATE TESTS (1-25)
		// =========================================================================
		{
			name: "RW-4.CertManager.Invalid.1 Certificate missing secretName fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: missing-secret
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "secretName",
		},
		{
			name: "RW-4.CertManager.Invalid.2 Certificate missing issuerRef fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: missing-issuerref
  namespace: default
spec:
  secretName: my-cert-tls
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "issuerRef",
		},
		{
			name: "RW-4.CertManager.Invalid.3 Certificate issuerRef missing name fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: issuerref-no-name
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    kind: Issuer
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-4.CertManager.Invalid.4 Certificate with empty string secretName fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: empty-secret
  namespace: default
spec:
  secretName: ""
  issuerRef:
    name: my-issuer
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "secretName",
		},
		{
			name: "RW-4.CertManager.Invalid.5 Certificate with empty string issuerRef name fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: empty-issuer-name
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: ""
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "issuerRef.name",
		},
		{
			name: "RW-4.CertManager.Invalid.6 Certificate with invalid duration format fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: bad-duration
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  duration: invalid-duration
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "duration",
		},
		{
			name: "RW-4.CertManager.Invalid.7 Certificate with invalid renewBefore format fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: bad-renewbefore
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  renewBefore: not-a-duration
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "renewBefore",
		},
		{
			name: "RW-4.CertManager.Invalid.8 Certificate with wrong type for dnsNames fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-dns-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  dnsNames: "example.com"
`),
			wantErr:   true,
			errSubstr: "dnsNames",
		},
		{
			name: "RW-4.CertManager.Invalid.9 Certificate with wrong type for ipAddresses fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-ip-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  ipAddresses: "10.0.0.1"
`),
			wantErr:   true,
			errSubstr: "ipAddresses",
		},
		{
			name: "RW-4.CertManager.Invalid.10 Certificate with wrong type for usages fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-usages-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  usages: digital signature
`),
			wantErr:   true,
			errSubstr: "usages",
		},
		{
			name: "RW-4.CertManager.Invalid.11 Certificate with invalid privateKey type fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-pk-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  privateKey: "RSA"
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "privateKey",
		},
		{
			name: "RW-4.CertManager.Invalid.12 Certificate with wrong type for secretTemplate fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-template-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  secretTemplate: labels
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "secretTemplate",
		},
		{
			name: "RW-4.CertManager.Invalid.13 Certificate with empty issuerRef object fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: empty-issuerref
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef: {}
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "issuerRef",
		},
		{
			name: "RW-4.CertManager.Invalid.14 Certificate with wrong type for nameConstraints fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-nc-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  nameConstraints: permitted
`),
			wantErr:   true,
			errSubstr: "nameConstraints",
		},
		{
			name: "RW-4.CertManager.Invalid.15 Certificate with non-integer revisionHistoryLimit fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-revlimit-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  revisionHistoryLimit: "5"
`),
			wantErr:   true,
			errSubstr: "revisionHistoryLimit",
		},
		{
			name: "RW-4.CertManager.Invalid.16 Certificate with non-boolean encodeUsagesInRequest fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-encode-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  encodeUsagesInRequest: "false"
`),
			wantErr:   true,
			errSubstr: "encodeUsagesInRequest",
		},
		{
			name: "RW-4.CertManager.Invalid.17 Certificate with non-integer privateKey size fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-pk-size-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  privateKey:
    size: "2048"
`),
			wantErr:   true,
			errSubstr: "size",
		},
		{
			name: "RW-4.CertManager.Invalid.18 Certificate with non-string commonName fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wrong-cn-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
  commonName: 12345
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "commonName",
		},
		{
			name: "RW-4.CertManager.Invalid.19 Certificate issuerRef with wrong type for kind fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: issuerref-wrong-kind-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
    kind: 123
`),
			wantErr:   true,
			errSubstr: "kind",
		},
		{
			name: "RW-4.CertManager.Invalid.20 Certificate issuerRef with wrong type for group fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: issuerref-wrong-group-type
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
    group: 123
`),
			wantErr:   true,
			errSubstr: "group",
		},
		// =========================================================================
		// INVALID ISSUER/CLUSTERISSUER TESTS (21-35)
		// =========================================================================
		{
			name: "RW-4.CertManager.Invalid.21 Issuer missing spec fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: missing-spec-issuer
  namespace: default
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-4.CertManager.Invalid.22 Issuer with empty spec fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: empty-spec-issuer
  namespace: default
spec: {}
`),
			wantErr:   true,
			errSubstr: "acme",
		},
		{
			name: "RW-4.CertManager.Invalid.23 Issuer ACME missing server fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: acme-missing-server
  namespace: default
spec:
  acme:
    email: admin@example.com
    privateKeySecretRef:
      name: acme-key
`),
			wantErr:   true,
			errSubstr: "server",
		},
		{
			name: "RW-4.CertManager.Invalid.24 Issuer ACME with wrong type for solvers fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: acme-wrong-solvers-type
  namespace: default
spec:
  acme:
    server: https://acme.example.com
    email: admin@example.com
    privateKeySecretRef:
      name: acme-key
    solvers: {}
`),
			wantErr:   true,
			errSubstr: "solvers",
		},
		{
			name: "RW-4.CertManager.Invalid.25 Issuer ACME with empty solvers array fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: acme-empty-solvers
  namespace: default
spec:
  acme:
    server: https://acme.example.com
    email: admin@example.com
    privateKeySecretRef:
      name: acme-key
    solvers: []
`),
			wantErr:   true,
			errSubstr: "solvers",
		},
		{
			name: "RW-4.CertManager.Invalid.26 Issuer ACME privateKeySecretRef missing name fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: acme-pk-no-name
  namespace: default
spec:
  acme:
    server: https://acme.example.com
    email: admin@example.com
    privateKeySecretRef:
      key: secret
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-4.CertManager.Invalid.27 Issuer CA missing secretName fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: ca-missing-secret
  namespace: default
spec:
  ca:
    secretName: ""
`),
			wantErr:   true,
			errSubstr: "secretName",
		},
		{
			name: "RW-4.CertManager.Invalid.28 Issuer with wrong type for acme fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wrong-acme-type
  namespace: default
spec:
  acme: server
`),
			wantErr:   true,
			errSubstr: "acme",
		},
		{
			name: "RW-4.CertManager.Invalid.29 Issuer with wrong type for ca fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wrong-ca-type
  namespace: default
spec:
  ca: secretName
`),
			wantErr:   true,
			errSubstr: "ca",
		},
		{
			name: "RW-4.CertManager.Invalid.30 Issuer with wrong type for selfSigned fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wrong-selfsigned-type
  namespace: default
spec:
  selfSigned: enabled
`),
			wantErr:   true,
			errSubstr: "selfSigned",
		},
		{
			name: "RW-4.CertManager.Invalid.31 Issuer with wrong type for venafi fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wrong-venafi-type
  namespace: default
spec:
  venafi: zone
`),
			wantErr:   true,
			errSubstr: "venafi",
		},
		{
			name: "RW-4.CertManager.Invalid.32 Issuer Venafi missing url fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: venafi-no-url
  namespace: default
spec:
  venafi:
    zone: my-zone
    apiTokenSecretRef:
      name: venafi-token
`),
			wantErr:   true,
			errSubstr: "url",
		},
		{
			name: "RW-4.CertManager.Invalid.33 Issuer Venafi missing zone fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: venafi-no-zone
  namespace: default
spec:
  venafi:
    url: https://example.com
    apiTokenSecretRef:
      name: venafi-token
`),
			wantErr:   true,
			errSubstr: "zone",
		},
		{
			name: "RW-4.CertManager.Invalid.34 ClusterIssuer missing spec fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: missing-spec-clusterissuer
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-4.CertManager.Invalid.35 ClusterIssuer with empty spec fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: empty-spec-clusterissuer
spec: {}
`),
			wantErr:   true,
			errSubstr: "acme",
		},
		// =========================================================================
		// INVALID CERTIFICATEREQUEST TESTS (36-50)
		// =========================================================================
		{
			name: "RW-4.CertManager.Invalid.36 CertificateRequest missing issuerRef fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: missing-issuerref-csr
  namespace: default
spec:
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "issuerRef",
		},
		{
			name: "RW-4.CertManager.Invalid.37 CertificateRequest missing request fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: missing-request-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
`),
			wantErr:   true,
			errSubstr: "request",
		},
		{
			name: "RW-4.CertManager.Invalid.38 CertificateRequest empty issuerRef fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: empty-issuerref-csr
  namespace: default
spec:
  issuerRef:
    name: ""
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "issuerRef",
		},
		{
			name: "RW-4.CertManager.Invalid.39 CertificateRequest empty request fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: empty-request-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: ""
`),
			wantErr:   true,
			errSubstr: "request",
		},
		{
			name: "RW-4.CertManager.Invalid.40 CertificateRequest with wrong type for issuerRef fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-issuerref-type-csr
  namespace: default
spec:
  issuerRef: my-issuer
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "issuerRef",
		},
		{
			name: "RW-4.CertManager.Invalid.41 CertificateRequest with wrong type for request fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-request-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: 12345
`),
			wantErr:   true,
			errSubstr: "request",
		},
		{
			name: "RW-4.CertManager.Invalid.42 CertificateRequest with wrong type for dnsNames fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-dnsnames-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  dnsNames: example.com
`),
			wantErr:   true,
			errSubstr: "dnsNames",
		},
		{
			name: "RW-4.CertManager.Invalid.43 CertificateRequest with wrong type for usages fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-usages-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  usages: digital signature
`),
			wantErr:   true,
			errSubstr: "usages",
		},
		{
			name: "RW-4.CertManager.Invalid.44 CertificateRequest with wrong type for isCA fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-isca-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  isCA: "true"
`),
			wantErr:   true,
			errSubstr: "isCA",
		},
		{
			name: "RW-4.CertManager.Invalid.45 CertificateRequest issuerRef missing name fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: issuerref-no-name-csr
  namespace: default
spec:
  issuerRef:
    kind: Issuer
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-4.CertManager.Invalid.46 CertificateRequest with wrong type for commonName fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-cn-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  commonName: 12345
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "commonName",
		},
		{
			name: "RW-4.CertManager.Invalid.47 CertificateRequest with wrong type for ipAddresses fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-ip-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  ipAddresses: "10.0.0.1"
`),
			wantErr:   true,
			errSubstr: "ipAddresses",
		},
		{
			name: "RW-4.CertManager.Invalid.48 CertificateRequest with wrong type for emailSANs fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-email-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  emailSANs: admin@example.com
`),
			wantErr:   true,
			errSubstr: "emailSANs",
		},
		{
			name: "RW-4.CertManager.Invalid.49 CertificateRequest with wrong type for uriSANs fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-uri-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  uriSANs: spiffe://example.com
`),
			wantErr:   true,
			errSubstr: "uriSANs",
		},
		{
			name: "RW-4.CertManager.Invalid.50 CertificateRequest with non-integer revision fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: CertificateRequest
metadata:
  name: wrong-rev-type-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
  revision: "1"
`),
			wantErr:   true,
			errSubstr: "revision",
		},
		// =========================================================================
		// INVALID ADDITIONAL TESTS (51-55)
		// =========================================================================
		{
			name: "RW-4.CertManager.Invalid.51 Certificate issuerRef with non-object name fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: issuerref-non-object-name
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name:
      first: value
  dnsNames:
    - example.com
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-4.CertManager.Invalid.52 Certificate with wrong kind fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: BadCertificate
metadata:
  name: wrong-kind-cert
  namespace: default
spec:
  secretName: my-cert-tls
  issuerRef:
    name: my-issuer
`),
			wantErr:   true,
			errSubstr: "kind",
		},
		{
			name: "RW-4.CertManager.Invalid.53 Issuer with wrong kind fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: BadIssuer
metadata:
  name: wrong-kind-issuer
  namespace: default
spec:
  acme:
    server: https://acme.example.com
    email: admin@example.com
    privateKeySecretRef:
      name: acme-key
`),
			wantErr:   true,
			errSubstr: "kind",
		},
		{
			name: "RW-4.CertManager.Invalid.54 ClusterIssuer with wrong kind fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: BadClusterIssuer
metadata:
  name: wrong-kind-clusterissuer
spec:
  acme:
    server: https://acme.example.com
    email: admin@example.com
    privateKeySecretRef:
      name: acme-key
`),
			wantErr:   true,
			errSubstr: "kind",
		},
		{
			name: "RW-4.CertManager.Invalid.55 CertificateRequest with wrong kind fails",
			crYAML: []byte(`apiVersion: cert-manager.io/v1
kind: BadCertificateRequest
metadata:
  name: wrong-kind-csr
  namespace: default
spec:
  issuerRef:
    name: my-issuer
  request: TESTREQUESTBASE64==
`),
			wantErr:   true,
			errSubstr: "kind",
		},
	}

	// Register CRDs
	crds := [][]byte{
		[]byte(certificateCRD),
		[]byte(issuerCRD),
		[]byte(clusterIssuerCRD),
		[]byte(certificateRequestCRD),
	}

	v := NewCRDValidator()
	for _, crd := range crds {
		if _, err := v.ValidateCRD(crd); err != nil {
			t.Fatalf("failed to register CRD: %v", err)
		}
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cr, err := v.ExtractSchemaFromYAML(tc.crYAML)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}

			gotErr := err != nil
			if gotErr && tc.wantErr {
				// Expected error and got one — check substr
				if tc.errSubstr != "" && !contains(errStr, tc.errSubstr) {
					t.Errorf("want errSubstr %q, got error %q", tc.errSubstr, errStr)
				}
			} else if gotErr && !tc.wantErr {
				t.Errorf("unexpected error: %v", err)
			} else if !gotErr && tc.wantErr {
				t.Errorf("expected error with substr %q, got nil", tc.errSubstr)
			}

			// Suppress unused variable warning
			_ = cr
		})
	}
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Make sure the types package is used to avoid unused import
var _ = types.Results{}