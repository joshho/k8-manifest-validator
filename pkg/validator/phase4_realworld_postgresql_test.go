package validator

import (
	"testing"
)

// =============================================================================
// RW-10 — PostgreSQL Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — postgresql-operator group
// Owner file: pkg/validator/phase4_realworld_postgresql_test.go
// CRDs covered: Pgcluster, Pgreplica, Pgbackup, Pgtask
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// PostgreSQL Operator CRD definitions (inline for test isolation)
// Based on Crunchy Data PostgreSQL Operator (postgres-operator.crunchydata.com)
// -----------------------------------------------------------------------

const pgclusterCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pgclusters.postgres-operator.crunchydata.com
spec:
  group: postgres-operator.crunchydata.com
  names:
    kind: Pgcluster
    plural: pgclusters
    singular: pgcluster
    listKind: PgclusterList
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
              name:
                type: string
              clusterName:
                type: string
              postgresVersion:
                type: integer
                minimum: 10
              instances:
                type: array
                items:
                  type: object
                  properties:
                    name:
                      type: string
                    replicas:
                      type: integer
                      minimum: 1
                    dataVolumeClaimSpec:
                      type: object
                      properties:
                        storageClassName:
                          type: string
                        accessModes:
                          type: array
                          items:
                            type: string
                        resources:
                          type: object
                      required:
                      - resources
              port:
                type: integer
                minimum: 1024
                maximum: 65535
              serviceType:
                type: string
                enum:
                - ClusterIP
                - NodePort
                - LoadBalancer
              image:
                type: string
              imagePullPolicy:
                type: string
                enum:
                - Always
                - IfNotPresent
                - Never
              backupURL:
                type: string
              backupSecret:
                type: string
              postgresConfig:
                type: string
              tls:
                type: object
                properties:
                  enabled:
                    type: boolean
                  secretName:
                    type: string
              proxy:
                type: object
                properties:
                  pgport:
                    type: integer
                  serviceType:
                    type: string
              userInterface:
                type: object
                properties:
                  serviceType:
                    type: string
              dataSource:
                type: object
                properties:
                  method:
                    type: string
                  url:
                    type: string
                  secretName:
                    type: string
            required:
            - name
            - postgresVersion
            - instances

          status:
            type: object
`

const pgreplicaCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pgreplicas.postgres-operator.crunchydata.com
spec:
  group: postgres-operator.crunchydata.com
  names:
    kind: Pgreplica
    plural: pgreplicas
    singular: pgreplica
    listKind: PgreplicaList
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
              clusterName:
                type: string
              replicaName:
                type: string
              source:
                type: object
                properties:
                  name:
                    type: string
                  namespace:
                    type: string
              image:
                type: string
              serviceType:
                type: string
                enum:
                - ClusterIP
                - NodePort
                - LoadBalancer
              resources:
                type: object
                properties:
                  requests:
                    type: object
                    properties:
                      cpu:
                        type: string
                      memory:
                        type: string
                  limits:
                    type: object
                    properties:
                      cpu:
                        type: string
                      memory:
                        type: string
            required:
            - clusterName
          status:
            type: object
`

const pgbackupCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pgbackups.postgres-operator.crunchydata.com
spec:
  group: postgres-operator.crunchydata.com
  names:
    kind: Pgbackup
    plural: pgbackups
    singular: pgbackup
    listKind: PgbackupList
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
              clusterName:
                type: string
              method:
                type: string
                enum:
                - full
                - incr
                - diff
              backupURL:
                type: string
              backupPath:
                type: string
              backupStatus:
                type: string
              storageType:
                type: string
              storageSource:
                type: object
                properties:
                  type:
                    type: string
                  endpoint:
                    type: string
                  bucket:
                    type: string
                  region:
                    type: string
              schedule:
                type: string
              retentionPeriod:
                type: string
            required:
            - clusterName
          status:
            type: object
`

const pgtaskCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pgtasks.postgres-operator.crunchydata.com
spec:
  group: postgres-operator.crunchydata.com
  names:
    kind: Pgtask
    plural: pgtasks
    singular: pgtask
    listKind: PgtaskList
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
              clusterName:
                type: string
              taskType:
                type: string
                enum:
                - backup
                - restore
                - clone
                - upgrade
                - migrate
              taskName:
                type: string
              parameters:
                type: object
                additionalProperties:
                  type: string
              status:
                type: string
              jobRef:
                type: object
                properties:
                  name:
                    type: string
                  namespace:
                    type: string
              owner:
                type: string
            required:
            - clusterName
            - taskType
          status:
            type: object
            properties:
              message:
                type: string
              state:
                type: string
              nextscheduletime:
                type: string
`

// -----------------------------------------------------------------------
// PostgreSQL Operator Valid Tests (RW-10.1 – RW-10.55)
// -----------------------------------------------------------------------

func TestPostgreSQL_Pgcluster_Valid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(pgclusterCRD))
	_ = engine.RegisterCRD([]byte(pgreplicaCRD))
	_ = engine.RegisterCRD([]byte(pgbackupCRD))
	_ = engine.RegisterCRD([]byte(pgtaskCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// Pgcluster valid tests (RW-10.1 – RW-10.15)
		{
			name: "RW-10.1 Pgcluster valid minimal",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.2 Pgcluster valid with port",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 14
  port: 5432
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.3 Pgcluster valid with ClusterIP service",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  serviceType: ClusterIP
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.4 Pgcluster valid with LoadBalancer service",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 14
  serviceType: LoadBalancer
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 10Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.5 Pgcluster valid with custom image",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  image: crunchydata/crunchy-postgres:centos7-13.4-2
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 5Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.6 Pgcluster valid with imagePullPolicy Always",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  image: crunchydata/crunchy-postgres:centos7-13.4-2
  imagePullPolicy: Always
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.7 Pgcluster valid with TLS enabled",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
  tls:
    enabled: true
    secretName: pg-tls-cert
`),
			wantErr: false,
		},
		{
			name: "RW-10.8 Pgcluster valid with multiple instances",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 14
  instances:
  - name: instance1
    replicas: 2
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 10Gi
  - name: instance2
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 10Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.9 Pgcluster valid with proxy config",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
  proxy:
    pgport: 5432
    serviceType: ClusterIP
`),
			wantErr: false,
		},
		{
			name: "RW-10.10 Pgcluster valid with backupURL",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  backupURL: s3://my-bucket/backups
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.11 Pgcluster valid with clusterName field",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  clusterName: my-postgres-cluster
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.12 Pgcluster valid with postgresConfig",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  postgresConfig: default
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.13 Pgcluster valid with storageClassName",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      storageClassName: standard
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.14 Pgcluster valid with accessModes",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.15 Pgcluster valid with empty proxy",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
  proxy: {}
`),
			wantErr: false,
		},

		// Pgreplica valid tests (RW-10.16 – RW-10.25)
		{
			name: "RW-10.16 Pgreplica valid minimal",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
`),
			wantErr: false,
		},
		{
			name: "RW-10.17 Pgreplica valid with source",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  source:
    name: my-postgres
    namespace: postgres-operator
`),
			wantErr: false,
		},
		{
			name: "RW-10.18 Pgreplica valid with custom image",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  image: crunchydata/crunchy-postgres:centos7-13.4-2
`),
			wantErr: false,
		},
		{
			name: "RW-10.19 Pgreplica valid with ClusterIP service",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  serviceType: ClusterIP
`),
			wantErr: false,
		},
		{
			name: "RW-10.20 Pgreplica valid with LoadBalancer service",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  serviceType: LoadBalancer
`),
			wantErr: false,
		},
		{
			name: "RW-10.21 Pgreplica valid with cpu memory requests",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  resources:
    requests:
      cpu: 500m
      memory: 256Mi
`),
			wantErr: false,
		},
		{
			name: "RW-10.22 Pgreplica valid with cpu memory limits",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  resources:
    limits:
      cpu: "1"
      memory: 1Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.23 Pgreplica valid with replicaName",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  replicaName: replica-west
`),
			wantErr: false,
		},
		{
			name: "RW-10.24 Pgreplica valid with full resource specs",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  replicaName: replica-east
  image: crunchydata/crunchy-postgres:centos7-14.2-0
  serviceType: NodePort
  resources:
    requests:
      cpu: 250m
      memory: 128Mi
    limits:
      cpu: "2"
      memory: 2Gi
`),
			wantErr: false,
		},
		{
			name: "RW-10.25 Pgreplica valid empty resources",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  resources: {}
`),
			wantErr: false,
		},

		// Pgbackup valid tests (RW-10.26 – RW-10.40)
		{
			name: "RW-10.26 Pgbackup valid minimal",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
`),
			wantErr: false,
		},
		{
			name: "RW-10.27 Pgbackup valid with full backup method",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: full
`),
			wantErr: false,
		},
		{
			name: "RW-10.28 Pgbackup valid with incr method",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: incr
`),
			wantErr: false,
		},
		{
			name: "RW-10.29 Pgbackup valid with diff method",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: diff
`),
			wantErr: false,
		},
		{
			name: "RW-10.30 Pgbackup valid with backupURL",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupURL: s3://my-bucket/postgres-backups/latest
`),
			wantErr: false,
		},
		{
			name: "RW-10.31 Pgbackup valid with backupPath",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupPath: /backups/my-postgres/20240514
`),
			wantErr: false,
		},
		{
			name: "RW-10.32 Pgbackup valid with storageSource",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: s3
    endpoint: https://s3.amazonaws.com
    bucket: my-backup-bucket
    region: us-east-1
`),
			wantErr: false,
		},
		{
			name: "RW-10.33 Pgbackup valid with schedule cron",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  schedule: "0 2 * * *"
`),
			wantErr: false,
		},
		{
			name: "RW-10.34 Pgbackup valid with retentionPeriod",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  retentionPeriod: "7d"
`),
			wantErr: false,
		},
		{
			name: "RW-10.35 Pgbackup valid with backupStatus",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupStatus: completed
`),
			wantErr: false,
		},
		{
			name: "RW-10.36 Pgbackup valid with storageType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageType: s3
`),
			wantErr: false,
		},
		{
			name: "RW-10.37 Pgbackup valid with empty storageSource",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource: {}
`),
			wantErr: false,
		},
		{
			name: "RW-10.38 Pgbackup valid full spec with all fields",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: full
  backupURL: s3://my-bucket/backups
  backupPath: /backups/my-postgres/latest
  backupStatus: in-progress
  storageType: s3
  storageSource:
    type: s3
    endpoint: https://s3.amazonaws.com
    bucket: my-bucket
    region: us-west-2
  schedule: "0 3 * * *"
  retentionPeriod: 30d
`),
			wantErr: false,
		},
		{
			name: "RW-10.39 Pgbackup valid with gcs storage source",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: gcs
    bucket: my-gcs-bucket
`),
			wantErr: false,
		},
		{
			name: "RW-10.40 Pgbackup valid with azure storage source",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: azure
    endpoint: https://myaccount.blob.core.windows.net
    bucket: my-container
`),
			wantErr: false,
		},

		// Pgtask valid tests (RW-10.41 – RW-10.55)
		{
			name: "RW-10.41 Pgtask valid minimal",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
`),
			wantErr: false,
		},
		{
			name: "RW-10.42 Pgtask valid with backup taskType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  taskName: nightly-backup
`),
			wantErr: false,
		},
		{
			name: "RW-10.43 Pgtask valid with restore taskType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: restore
`),
			wantErr: false,
		},
		{
			name: "RW-10.44 Pgtask valid with clone taskType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: clone
`),
			wantErr: false,
		},
		{
			name: "RW-10.45 Pgtask valid with upgrade taskType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: upgrade
`),
			wantErr: false,
		},
		{
			name: "RW-10.46 Pgtask valid with migrate taskType",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: migrate
`),
			wantErr: false,
		},
		{
			name: "RW-10.47 Pgtask valid with single parameter",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  parameters:
    backup-type: full
`),
			wantErr: false,
		},
		{
			name: "RW-10.48 Pgtask valid with multiple parameters",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  parameters:
    backup-type: full
    compression: gzip
    verify: "true"
`),
			wantErr: false,
		},
		{
			name: "RW-10.49 Pgtask valid with jobRef",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  jobRef:
    name: backup-job-12345
    namespace: postgres-operator
`),
			wantErr: false,
		},
		{
			name: "RW-10.50 Pgtask valid with owner reference",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  owner: my-postgres
`),
			wantErr: false,
		},
		{
			name: "RW-10.51 Pgtask valid with status message",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
status:
  message: backup completed successfully
  state: completed
`),
			wantErr: false,
		},
		{
			name: "RW-10.52 Pgtask valid with nextscheduletime",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
status:
  state: running
  nextscheduletime: "2024-05-15T02:00:00Z"
`),
			wantErr: false,
		},
		{
			name: "RW-10.53 Pgtask valid with full spec",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: restore
  taskName: restore-from-backup
  parameters:
    backup-url: s3://my-bucket/backups/latest
    target-timestamp: "2024-05-01T00:00:00Z"
  jobRef:
    name: restore-job-xyz
    namespace: postgres-operator
  owner: my-postgres
status:
  message: restore in progress
  state: running
`),
			wantErr: false,
		},
		{
			name: "RW-10.54 Pgtask valid with empty parameters",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  parameters: {}
`),
			wantErr: false,
		},
		{
			name: "RW-10.55 Pgtask valid with empty jobRef",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  jobRef: {}
`),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Validate(tc.crYAML)
			if tc.wantErr && len(result.Errors) == 0 {
				t.Errorf("expected error for %q, got none", tc.name)
			}
			if !tc.wantErr && len(result.Errors) > 0 {
				t.Errorf("unexpected error for %q: %v", tc.name, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// PostgreSQL Operator Invalid Tests (RW-10.56 – RW-10.110)
// -----------------------------------------------------------------------

func TestPostgreSQL_Pgcluster_Invalid(t *testing.T) {
	t.Parallel()
	engine := &Engine{}
	_ = engine.RegisterCRD([]byte(pgclusterCRD))
	_ = engine.RegisterCRD([]byte(pgreplicaCRD))
	_ = engine.RegisterCRD([]byte(pgbackupCRD))
	_ = engine.RegisterCRD([]byte(pgtaskCRD))

	cases := []struct {
		name      string
		crYAML    []byte
		wantErr   bool
		errSubstr string
	}{
		// Pgcluster invalid tests (RW-10.56 – RW-10.70)
		{
			name: "RW-10.56 Pgcluster missing name fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.57 Pgcluster missing postgresVersion fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "postgresVersion",
		},
		{
			name: "RW-10.58 Pgcluster missing instances fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
`),
			wantErr:   true,
			errSubstr: "instances",
		},
		{
			name: "RW-10.59 Pgcluster empty name fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: ""
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.60 Pgcluster invalid port low fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  port: 512
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "RW-10.61 Pgcluster invalid port high fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  port: 70000
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "port",
		},
		{
			name: "RW-10.62 Pgcluster wrong serviceType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  serviceType: InvalidType
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "serviceType",
		},
		{
			name: "RW-10.63 Pgcluster postgresVersion too low fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 9
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "postgresVersion",
		},
		{
			name: "RW-10.64 Pgcluster wrong imagePullPolicy fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  imagePullPolicy: InvalidPolicy
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "imagePullPolicy",
		},
		{
			name: "RW-10.65 Pgcluster empty instances fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances: []
`),
			wantErr:   true,
			errSubstr: "instances",
		},
		{
			name: "RW-10.66 Pgcluster instance missing name fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.67 Pgcluster instance missing dataVolumeClaimSpec fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
`),
			wantErr:   true,
			errSubstr: "dataVolumeClaimSpec",
		},
		{
			name: "RW-10.68 Pgcluster instance missing resources fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec: {}
`),
			wantErr:   true,
			errSubstr: "resources",
		},
		{
			name: "RW-10.69 Pgcluster instance replicas zero fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgcluster
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 0
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "replicas",
		},
		{
			name: "RW-10.70 Pgcluster wrong kind fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: WrongKind
metadata:
  name: my-postgres
  namespace: postgres-operator
spec:
  name: my-postgres
  postgresVersion: 13
  instances:
  - name: instance1
    replicas: 1
    dataVolumeClaimSpec:
      resources:
        requests:
          storage: 1Gi
`),
			wantErr:   true,
			errSubstr: "",
		},

		// Pgreplica invalid tests (RW-10.71 – RW-10.80)
		{
			name: "RW-10.71 Pgreplica missing clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec: {}
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.72 Pgreplica empty clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: ""
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.73 Pgreplica wrong serviceType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  serviceType: InvalidType
`),
			wantErr:   true,
			errSubstr: "serviceType",
		},
		{
			name: "RW-10.74 Pgreplica source missing name fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  source:
    namespace: postgres-operator
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.75 Pgreplica resources cpu non-string fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  resources:
    requests:
      cpu: 123
      memory: 256Mi
`),
			wantErr:   true,
			errSubstr: "cpu",
		},
		{
			name: "RW-10.76 Pgreplica resources memory non-string fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  resources:
    requests:
      cpu: 500m
      memory: 256
`),
			wantErr:   true,
			errSubstr: "memory",
		},
		{
			name: "RW-10.77 Pgreplica wrong kind fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: NotAPgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-10.78 Pgreplica empty spec fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec: {}
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.79 Pgreplica source empty name fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  source:
    name: ""
    namespace: postgres-operator
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.80 Pgreplica image as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgreplica
metadata:
  name: my-replica
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  image: 12345
`),
			wantErr:   true,
			errSubstr: "image",
		},

		// Pgbackup invalid tests (RW-10.81 – RW-10.95)
		{
			name: "RW-10.81 Pgbackup missing clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec: {}
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.82 Pgbackup empty clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: ""
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.83 Pgbackup invalid method fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: invalid-method
`),
			wantErr:   true,
			errSubstr: "method",
		},
		{
			name: "RW-10.84 Pgbackup wrong storageSource type fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: unknown
    bucket: my-bucket
`),
			wantErr:   true,
			errSubstr: "type",
		},
		{
			name: "RW-10.85 Pgbackup invalid storageType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageType: ftp
`),
			wantErr:   true,
			errSubstr: "storageType",
		},
		{
			name: "RW-10.86 Pgbackup empty spec fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec: {}
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.87 Pgbackup wrong kind fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: NotAPgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-10.88 Pgbackup backupURL as number fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupURL: 12345
`),
			wantErr:   true,
			errSubstr: "backupURL",
		},
		{
			name: "RW-10.89 Pgbackup schedule as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  schedule: 12345
`),
			wantErr:   true,
			errSubstr: "schedule",
		},
		{
			name: "RW-10.90 Pgbackup retentionPeriod as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  retentionPeriod: 30
`),
			wantErr:   true,
			errSubstr: "retentionPeriod",
		},
		{
			name: "RW-10.91 Pgbackup backupStatus as boolean fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupStatus: true
`),
			wantErr:   true,
			errSubstr: "backupStatus",
		},
		{
			name: "RW-10.92 Pgbackup backupPath as object fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  backupPath:
    path: /some/path
`),
			wantErr:   true,
			errSubstr: "backupPath",
		},
		{
			name: "RW-10.93 Pgbackup storageSource type as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: 1
`),
			wantErr:   true,
			errSubstr: "type",
		},
		{
			name: "RW-10.94 Pgbackup storageSource bucket as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  storageSource:
    type: s3
    bucket: 123
`),
			wantErr:   true,
			errSubstr: "bucket",
		},
		{
			name: "RW-10.95 Pgbackup method as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgbackup
metadata:
  name: my-backup
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  method: 1
`),
			wantErr:   true,
			errSubstr: "method",
		},

		// Pgtask invalid tests (RW-10.96 – RW-10.110)
		{
			name: "RW-10.96 Pgtask missing clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  taskType: backup
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.97 Pgtask missing taskType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
`),
			wantErr:   true,
			errSubstr: "taskType",
		},
		{
			name: "RW-10.98 Pgtask empty clusterName fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: ""
  taskType: backup
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.99 Pgtask empty taskType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: ""
`),
			wantErr:   true,
			errSubstr: "taskType",
		},
		{
			name: "RW-10.100 Pgtask invalid taskType fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: invalid-type
`),
			wantErr:   true,
			errSubstr: "taskType",
		},
		{
			name: "RW-10.101 Pgtask wrong kind fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: NotAPgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
`),
			wantErr:   true,
			errSubstr: "",
		},
		{
			name: "RW-10.102 Pgtask empty spec fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec: {}
`),
			wantErr:   true,
			errSubstr: "clusterName",
		},
		{
			name: "RW-10.103 Pgtask parameters non-object fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  parameters: "not-an-object"
`),
			wantErr:   true,
			errSubstr: "parameters",
		},
		{
			name: "RW-10.104 Pgtask jobRef name as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  jobRef:
    name: 12345
`),
			wantErr:   true,
			errSubstr: "name",
		},
		{
			name: "RW-10.105 Pgtask owner as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  owner: 42
`),
			wantErr:   true,
			errSubstr: "owner",
		},
		{
			name: "RW-10.106 Pgtask status message as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
status:
  message: 12345
`),
			wantErr:   true,
			errSubstr: "message",
		},
		{
			name: "RW-10.107 Pgtask status state as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
status:
  state: 123
`),
			wantErr:   true,
			errSubstr: "state",
		},
		{
			name: "RW-10.108 Pgtask taskName as boolean fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  taskName: true
`),
			wantErr:   true,
			errSubstr: "taskName",
		},
		{
			name: "RW-10.109 Pgtask nextscheduletime as integer fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
status:
  nextscheduletime: 12345
`),
			wantErr:   true,
			errSubstr: "nextscheduletime",
		},
		{
			name: "RW-10.110 Pgtask parameters value non-string fails",
			crYAML: []byte(`apiVersion: postgres-operator.crunchydata.com/v1
kind: Pgtask
metadata:
  name: my-task
  namespace: postgres-operator
spec:
  clusterName: my-postgres
  taskType: backup
  parameters:
    backup-type: 123
`),
			wantErr:   true,
			errSubstr: "backup-type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Validate(tc.crYAML)
			if tc.wantErr && len(result.Errors) == 0 {
				t.Errorf("expected error for %q, got none", tc.name)
			}
			if !tc.wantErr && len(result.Errors) > 0 {
				t.Errorf("unexpected error for %q: %v", tc.name, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasCRError(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}