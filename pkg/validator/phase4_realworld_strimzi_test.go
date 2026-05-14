package validator

import (
	"testing"

	"k8-manifest-validator/pkg/types"
)

// =============================================================================
// RW-5 — Strimzi Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — strimzi operator group
// Owner file: pkg/validator/phase4_realworld_strimzi_test.go
// CRDs covered: Kafka, KafkaTopic, KafkaUser, KafkaConnect, KafkaMirrorMaker2
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// Strimzi CRD definitions (inline for test isolation)
// These match the actual Strimzi CRD schemas (v1beta3 for Kafka, v1beta2 for others)
// -----------------------------------------------------------------------

// strimziKafkaCRD is the Kafka CRD covering Kafka, KafkaConnect, KafkaMirrorMaker2
const strimziKafkaCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kafka.strimzi.io
spec:
  group: strimzi.io
  names:
    kind: Kafka
    plural: kafkas
    singular: kafka
    listKind: KafkaList
  scope: Namespaced
  versions:
  - name: v1beta3
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              kafka:
                type: object
                properties:
                  replicas:
                    type: integer
                    minimum: 1
                  version:
                    type: string
                  replicas:
                    type: integer
                  resources:
                    type: object
                  listeners:
                    type: array
                  config:
                    type: object
                  storage:
                    type: object
                  listeners:
                    type: array
                    items:
                      type: object
                required:
                - replicas
              zookeeper:
                type: object
                properties:
                  replicas:
                    type: integer
                  resources:
                    type: object
              entityOperator:
                type: object
                properties:
                  topicOperator:
                    type: object
                  userOperator:
                    type: object
              cruiseControl:
                type: object
  storedVersions:
  - v1beta3
`

const strimziKafkaTopicCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kafkatopics.strimzi.io
spec:
  group: strimzi.io
  names:
    kind: KafkaTopic
    plural: kafkatopics
    singular: kafkatopic
    listKind: KafkaTopicList
  scope: Namespaced
  versions:
  - name: v1beta2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              partitions:
                type: integer
                minimum: 1
              replicas:
                type: integer
                minimum: 1
              config:
                type: object
              topicName:
                type: string
            required:
            - partitions
            - replicas
  storedVersions:
  - v1beta2
`

const strimziKafkaUserCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kafkausers.strimzi.io
spec:
  group: strimzi.io
  names:
    kind: KafkaUser
    plural: kafkausers
    singular: kafkauser
    listKind: KafkaUserList
  scope: Namespaced
  versions:
  - name: v1beta2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              authentication:
                type: object
              authorization:
                type: object
              quotas:
                type: object
            required:
            - authentication
  storedVersions:
  - v1beta2
`

const strimziKafkaConnectCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kafkaconnects.strimzi.io
spec:
  group: strimzi.io
  names:
    kind: KafkaConnect
    plural: kafkaconnects
    singular: kafkaconnect
    listKind: KafkaConnectList
  scope: Namespaced
  versions:
  - name: v1beta2
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
              bootstrapServers:
                type: string
              tls:
                type: object
              authentication:
                type: object
              logging:
                type: object
              resources:
                type: object
            required:
            - bootstrapServers
  storedVersions:
  - v1beta2
`

const strimziKafkaMirrorMaker2CRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: kafkamirrormaker2s.strimzi.io
spec:
  group: strimzi.io
  names:
    kind: KafkaMirrorMaker2
    plural: kafkamirrormaker2s
    singular: kafkamirrormaker2
    listKind: KafkaMirrorMaker2List
  scope: Namespaced
  versions:
  - name: v1beta2
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
              version:
                type: string
              connectCluster:
                type: string
              mirrors:
                type: array
              resources:
                type: object
            required:
            - replicas
            - connectCluster
            - mirrors
  storedVersions:
  - v1beta2
`

// -----------------------------------------------------------------------
// Test fixtures — 55 valid + 55 invalid strimzi resource cases
// -----------------------------------------------------------------------

func TestStrimzi_Kafka_Valid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaCRD)); err != nil {
		t.Fatalf("failed to register Kafka CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.01 – RW-5.11: Kafka valid cases
		{
			name: "RW-5.01 Kafka with all required fields passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    version: "3.5.0"
`),
			wantErr: false,
		},
		{
			name: "RW-5.02 Kafka with listeners array passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    listeners:
    - name: plain
      port: 9092
    - name: tls
      port: 9093
`),
			wantErr: false,
		},
		{
			name: "RW-5.03 Kafka with kafka.replicas and storage passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 1
    storage:
      type: jbod
`),
			wantErr: false,
		},
		{
			name: "RW-5.04 Kafka with entityOperator passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  entityOperator:
    topicOperator: {}
    userOperator: {}
`),
			wantErr: false,
		},
		{
			name: "RW-5.05 Kafka with zookeeper passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  zookeeper:
    replicas: 3
`),
			wantErr: false,
		},
		{
			name: "RW-5.06 Kafka with cruiseControl passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  cruiseControl:
    {}
`),
			wantErr: false,
		},
		{
			name: "RW-5.07 Kafka with kafka.config passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    config:
      num.partitions: 3
      num.io.threads: 8
`),
			wantErr: false,
		},
		{
			name: "RW-5.08 Kafka with zookeeper and resources passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  zookeeper:
    replicas: 3
    resources:
      limits:
        cpu: "1"
        memory: 2Gi
`),
			wantErr: false,
		},
		{
			name: "RW-5.09 Kafka with kafka.version string passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    version: "3.4.0"
`),
			wantErr: false,
		},
		{
			name: "RW-5.10 Kafka multiple components passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: prod-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 5
    version: "3.5.0"
    listeners:
    - name: plain
      port: 9092
    - name: tls
      port: 9093
  zookeeper:
    replicas: 3
  entityOperator:
    topicOperator: {}
    userOperator: {}
`),
			wantErr: false,
		},
		{
			name: "RW-5.11 Kafka with storage type ephemeral passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    storage:
      type: ephemeral
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaTopic_Valid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaTopicCRD)); err != nil {
		t.Fatalf("failed to register KafkaTopic CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.12 – RW-5.22: KafkaTopic valid cases
		{
			name: "RW-5.12 KafkaTopic with required partitions and replicas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
`),
			wantErr: false,
		},
		{
			name: "RW-5.13 KafkaTopic with topicName passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 6
  replicas: 3
  topicName: my-topic
`),
			wantErr: false,
		},
		{
			name: "RW-5.14 KafkaTopic with config passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config:
    retention.bytes: 1073741824
    cleanup.policy: delete
`),
			wantErr: false,
		},
		{
			name: "RW-5.15 KafkaTopic with partitions 1 passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: minimal-topic
  namespace: kafka
spec:
  partitions: 1
  replicas: 1
`),
			wantErr: false,
		},
		{
			name: "RW-5.16 KafkaTopic with high replica count passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: replicated-topic
  namespace: kafka
spec:
  partitions: 10
  replicas: 5
`),
			wantErr: false,
		},
		{
			name: "RW-5.17 KafkaTopic with retention config passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config:
    retention.ms: 86400000
    retention.bytes: -1
`),
			wantErr: false,
		},
		{
			name: "RW-5.18 KafkaTopic with segment bytes config passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config:
    segment.bytes: 1073741824
`),
			wantErr: false,
		},
		{
			name: "RW-5.19 KafkaTopic with cleanup policy compact passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: compact-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config:
    cleanup.policy: compact
`),
			wantErr: false,
		},
		{
			name: "RW-5.20 KafkaTopic with min.insync.replicas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config:
    min.insync.replicas: 2
`),
			wantErr: false,
		},
		{
			name: "RW-5.21 KafkaTopic with all config options passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: full-topic
  namespace: kafka
spec:
  partitions: 12
  replicas: 3
  config:
    retention.ms: 604800000
    retention.bytes: -1
    cleanup.policy: delete
    segment.bytes: 1073741824
    min.insync.replicas: 2
`),
			wantErr: false,
		},
		{
			name: "RW-5.22 KafkaTopic with unique name in different ns passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: cross-ns-topic
  namespace: kafka-ns
spec:
  partitions: 3
  replicas: 3
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaUser_Valid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaUserCRD)); err != nil {
		t.Fatalf("failed to register KafkaUser CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.23 – RW-5.33: KafkaUser valid cases
		{
			name: "RW-5.23 KafkaUser with required authentication passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
`),
			wantErr: false,
		},
		{
			name: "RW-5.24 KafkaUser with tls cert type passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: tls-user
  namespace: kafka
spec:
  authentication:
    type: tls
    certificates:
    - cert data
`),
			wantErr: false,
		},
		{
			name: "RW-5.25 KafkaUser with scram-sha-512 passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: scram-user
  namespace: kafka
spec:
  authentication:
    type: scram-sha-512
`),
			wantErr: false,
		},
		{
			name: "RW-5.26 KafkaUser with authorization passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  authorization:
    type: simple
`),
			wantErr: false,
		},
		{
			name: "RW-5.27 KafkaUser with quotas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  quotas:
    producer_byte_rate: 1048576
    consumer_byte_rate: 1048576
`),
			wantErr: false,
		},
		{
			name: "RW-5.28 KafkaUser with full spec passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: full-user
  namespace: kafka
spec:
  authentication:
    type: tls
  authorization:
    type: simple
    acls:
    - resource:
        type: topic
        name: my-topic
      operation: Read
  quotas:
    producer_byte_rate: 1048576
    consumer_byte_rate: 1048576
`),
			wantErr: false,
		},
		{
			name: "RW-5.29 KafkaUser with scram-sha-256 passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: scram-sha256-user
  namespace: kafka
spec:
  authentication:
    type: scram-sha-256
`),
			wantErr: false,
		},
		{
			name: "RW-5.30 KafkaUser with authorization simple type passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: auth-user
  namespace: kafka
spec:
  authentication:
    type: tls
  authorization:
    type: simple
`),
			wantErr: false,
		},
		{
			name: "RW-5.31 KafkaUser with quotas zero values passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  quotas:
    producer_byte_rate: 0
    consumer_byte_rate: 0
`),
			wantErr: false,
		},
		{
			name: "RW-5.32 KafkaUser in different namespace passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: ns-user
  namespace: kafka-prod
spec:
  authentication:
    type: tls
`),
			wantErr: false,
		},
		{
			name: "RW-5.33 KafkaUser with all authentication options passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: all-auth-user
  namespace: kafka
spec:
  authentication:
    type: scram-sha-512
  authorization:
    type: simple
  quotas:
    producer_byte_rate: 2097152
    consumer_byte_rate: 2097152
    request_percentage: 50
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaConnect_Valid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaConnectCRD)); err != nil {
		t.Fatalf("failed to register KafkaConnect CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.34 – RW-5.44: KafkaConnect valid cases
		{
			name: "RW-5.34 KafkaConnect with required bootstrapServers passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr: false,
		},
		{
			name: "RW-5.35 KafkaConnect with replicas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  replicas: 2
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr: false,
		},
		{
			name: "RW-5.36 KafkaConnect with tls passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9093
  tls: {}
`),
			wantErr: false,
		},
		{
			name: "RW-5.37 KafkaConnect with authentication passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  authentication:
    type: tls
`),
			wantErr: false,
		},
		{
			name: "RW-5.38 KafkaConnect with logging passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  logging:
    type: inline
`),
			wantErr: false,
		},
		{
			name: "RW-5.39 KafkaConnect with resources passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  resources:
    limits:
      cpu: "1"
      memory: 2Gi
`),
			wantErr: false,
		},
		{
			name: "RW-5.40 KafkaConnect with bootstrapServers containing port passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: cluster-1.example.com:9093,cluster-2.example.com:9093
`),
			wantErr: false,
		},
		{
			name: "RW-5.41 KafkaConnect with replicas 1 passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: single-connect
  namespace: kafka
spec:
  replicas: 1
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr: false,
		},
		{
			name: "RW-5.42 KafkaConnect with empty tls passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: tls-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9093
  tls:
    trustedCertificates:
    - secretName: my-cluster-cluster-ca-cert
      certificate: ca.crt
`),
			wantErr: false,
		},
		{
			name: "RW-5.43 KafkaConnect with multiple replicas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: multi-connect
  namespace: kafka
spec:
  replicas: 5
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr: false,
		},
		{
			name: "RW-5.44 KafkaConnect with all fields passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: full-connect
  namespace: kafka
spec:
  replicas: 3
  bootstrapServers: my-cluster-kafka:9092
  tls: {}
  authentication:
    type: tls
  logging:
    type: inline
  resources:
    limits:
      cpu: "2"
      memory: 4Gi
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaMirrorMaker2_Valid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaMirrorMaker2CRD)); err != nil {
		t.Fatalf("failed to register KafkaMirrorMaker2 CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.45 – RW-5.55: KafkaMirrorMaker2 valid cases
		{
			name: "RW-5.45 KafkaMirrorMaker2 with required fields passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: my-cluster-target
  mirrors:
  - sourceCluster: my-cluster-source
    targetCluster: my-cluster-target
    sourceTopic: topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.46 KafkaMirrorMaker2 with multiple mirrors passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-1
  - sourceCluster: source-b
    targetCluster: target-cluster
    sourceTopic: topic-2
`),
			wantErr: false,
		},
		{
			name: "RW-5.47 KafkaMirrorMaker2 with version passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  version: "3.5.0"
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.48 KafkaMirrorMaker2 with resources passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 2
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
  resources:
    limits:
      cpu: "1"
      memory: 2Gi
`),
			wantErr: false,
		},
		{
			name: "RW-5.49 KafkaMirrorMaker2 with high replica count passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 5
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.50 KafkaMirrorMaker2 with target endpoint override passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
    targetTopic: replicated-topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.51 KafkaMirrorMaker2 with multiple replica spec passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 3
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.52 KafkaMirrorMaker2 with version and replicas passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 2
  version: "3.4.0"
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr: false,
		},
		{
			name: "RW-5.53 KafkaMirrorMaker2 with resources and version passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  version: "3.5.0"
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
  resources:
    limits:
      cpu: "2"
      memory: 4Gi
`),
			wantErr: false,
		},
		{
			name: "RW-5.54 KafkaMirrorMaker2 with multiple mirrors and resources passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 2
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: cluster-us
    targetCluster: target-cluster
    sourceTopic: us-topic-1
  - sourceCluster: cluster-eu
    targetCluster: target-cluster
    sourceTopic: eu-topic-1
  resources:
    limits:
      cpu: "1"
      memory: 2Gi
`),
			wantErr: false,
		},
		{
			name: "RW-5.55 KafkaMirrorMaker2 with replication policy passes",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
    replicationPolicy: "prefix"
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// =============================================================================
// Invalid test cases — 55 invalid strimzi resource cases
// =============================================================================

func TestStrimzi_Kafka_Invalid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaCRD)); err != nil {
		t.Fatalf("failed to register Kafka CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.56 – RW-5.66: Kafka invalid cases
		{
			name: "RW-5.56 Kafka missing kafka.replicas fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    version: "3.5.0"
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.57 Kafka with kafka.replicas zero fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 0
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.58 Kafka with kafka.replicas string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: "three"
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.59 Kafka missing spec entirely fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-5.60 Kafka with kafka.replicas negative fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: -1
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.61 Kafka with zookeeper.replicas string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  zookeeper:
    replicas: "three"
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.62 Kafka with zookeeper.replicas negative fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
  zookeeper:
    replicas: -3
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.63 Kafka with empty string metadata.name fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: ""
  namespace: kafka
spec:
  kafka:
    replicas: 3
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-5.64 Kafka with spec containing invalid type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: three
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.65 Kafka with kafka.listeners wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    replicas: 3
    listeners: "not-an-array"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-5.66 Kafka with empty kafka.spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta3
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka: {}
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaTopic_Invalid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaTopicCRD)); err != nil {
		t.Fatalf("failed to register KafkaTopic CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.67 – RW-5.77: KafkaTopic invalid cases
		{
			name: "RW-5.67 KafkaTopic missing partitions fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  replicas: 3
`),
			wantErr:   true,
			errSubstr: "partitions",
		},
		{
			name: "RW-5.68 KafkaTopic missing replicas fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.69 KafkaTopic partitions zero fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 0
  replicas: 3
`),
			wantErr:   true,
			errSubstr: "partitions",
		},
		{
			name: "RW-5.70 KafkaTopic partitions string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: "three"
  replicas: 3
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.71 KafkaTopic replicas string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: "three"
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.72 KafkaTopic partitions negative fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: -5
  replicas: 3
`),
			wantErr:   true,
			errSubstr: "partitions",
		},
		{
			name: "RW-5.73 KafkaTopic replicas negative fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: -1
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.74 KafkaTopic empty spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec: {}
`),
			wantErr:   true,
			errSubstr: "partitions",
		},
		{
			name: "RW-5.75 KafkaTopic missing spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-5.76 KafkaTopic with config wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  config: "not-an-object"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.77 KafkaTopic topicName wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: kafka
spec:
  partitions: 3
  replicas: 3
  topicName: 12345
`),
			wantErr:   true,
			errSubstr: "string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaUser_Invalid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaUserCRD)); err != nil {
		t.Fatalf("failed to register KafkaUser CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.78 – RW-5.88: KafkaUser invalid cases
		{
			name: "RW-5.78 KafkaUser missing authentication fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authorization:
    type: simple
`),
			wantErr:   true,
			errSubstr: "authentication",
		},
		{
			name: "RW-5.79 KafkaUser with empty spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec: {}
`),
			wantErr:   true,
			errSubstr: "authentication",
		},
		{
			name: "RW-5.80 KafkaUser missing spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-5.81 KafkaUser with authentication wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication: "tls"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.82 KafkaUser with authorization wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  authorization: "simple"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.83 KafkaUser with quotas wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  quotas: []
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.84 KafkaUser with empty string name fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: ""
  namespace: kafka
spec:
  authentication:
    type: tls
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-5.85 KafkaUser with empty authentication type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: ""
`),
			wantErr:   true,
			errSubstr: "type",
		},
		{
			name: "RW-5.86 KafkaUser authentication type wrong value fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: unknown-auth-type
`),
			wantErr:   true,
			errSubstr: "unknown",
		},
		{
			name: "RW-5.87 KafkaUser with invalid quotas structure fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  quotas:
    producer_byte_rate: "unlimited"
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.88 KafkaUser authorization type unknown fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-user
  namespace: kafka
spec:
  authentication:
    type: tls
  authorization:
    type: unknown-type
`),
			wantErr:   true,
			errSubstr: "unknown",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaConnect_Invalid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaConnectCRD)); err != nil {
		t.Fatalf("failed to register KafkaConnect CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.89 – RW-5.99: KafkaConnect invalid cases
		{
			name: "RW-5.89 KafkaConnect missing bootstrapServers fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  replicas: 2
`),
			wantErr:   true,
			errSubstr: "bootstrapServers",
		},
		{
			name: "RW-5.90 KafkaConnect with empty spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec: {}
`),
			wantErr:   true,
			errSubstr: "bootstrapServers",
		},
		{
			name: "RW-5.91 KafkaConnect missing spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
`),
			wantErr:   true,
			errSubstr: "spec",
		},
		{
			name: "RW-5.92 KafkaConnect replicas string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  replicas: "two"
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.93 KafkaConnect bootstrapServers empty string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: ""
`),
			wantErr:   true,
			errSubstr: "bootstrapServers",
		},
		{
			name: "RW-5.94 KafkaConnect tls wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  tls: "enabled"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.95 KafkaConnect authentication wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  authentication: []
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.96 KafkaConnect logging wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  logging: "debug"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.97 KafkaConnect resources wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
  resources: "2Gi"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.98 KafkaConnect with empty metadata.name fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: ""
  namespace: kafka
spec:
  bootstrapServers: my-cluster-kafka:9092
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-5.99 KafkaConnect with bootstrapServers numeric fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: my-connect
  namespace: kafka
spec:
  bootstrapServers: 9092
`),
			wantErr:   true,
			errSubstr: "string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestStrimzi_KafkaMirrorMaker2_Invalid(t *testing.T) {
	t.Parallel()
	crdValidator := NewCRDValidator()
	crValidator := NewCRValidator(crdValidator)

	if err := crValidator.RegisterCRD([]byte(strimziKafkaMirrorMaker2CRD)); err != nil {
		t.Fatalf("failed to register KafkaMirrorMaker2 CRD: %v", err)
	}

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// RW-5.100 – RW-5.110: KafkaMirrorMaker2 invalid cases
		{
			name: "RW-5.100 KafkaMirrorMaker2 missing replicas fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.101 KafkaMirrorMaker2 missing connectCluster fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr:   true,
			errSubstr: "connectCluster",
		},
		{
			name: "RW-5.102 KafkaMirrorMaker2 missing mirrors fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
`),
			wantErr:   true,
			errSubstr: "mirrors",
		},
		{
			name: "RW-5.103 KafkaMirrorMaker2 replicas string fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: "one"
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr:   true,
			errSubstr: "integer",
		},
		{
			name: "RW-5.104 KafkaMirrorMaker2 empty spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec: {}
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-5.105 KafkaMirrorMaker2 mirrors wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors: "mirror-1"
`),
			wantErr:   true,
			errSubstr: "array",
		},
		{
			name: "RW-5.106 KafkaMirrorMaker2 connectCluster empty fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: ""
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr:   true,
			errSubstr: "connectCluster",
		},
		{
			name: "RW-5.107 KafkaMirrorMaker2 version wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  version: 3050
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
`),
			wantErr:   true,
			errSubstr: "string",
		},
		{
			name: "RW-5.108 KafkaMirrorMaker2 resources wrong type fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors:
  - sourceCluster: source-a
    targetCluster: target-cluster
    sourceTopic: topic-a
  resources: "2cpu"
`),
			wantErr:   true,
			errSubstr: "object",
		},
		{
			name: "RW-5.109 KafkaMirrorMaker2 mirrors empty array fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
spec:
  replicas: 1
  connectCluster: target-cluster
  mirrors: []
`),
			wantErr:   true,
			errSubstr: "mirrors",
		},
		{
			name: "RW-5.110 KafkaMirrorMaker2 missing spec fails",
			crYAML: []byte(`apiVersion: strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: my-mm2
  namespace: kafka
`),
			wantErr:   true,
			errSubstr: "spec",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := crValidator.ValidateCR(tc.crYAML)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotErr := result.Status != types.StatusValid
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v, status=%v, errors=%v", tc.wantErr, gotErr, result.Status, result.Errors)
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