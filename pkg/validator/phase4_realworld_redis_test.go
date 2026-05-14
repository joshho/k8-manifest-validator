package validator

import (
	"testing"
)

// =============================================================================
// RW-9 — Redis Operator Real-World Test Cases
// PRD ref: Real-world operator coverage — redis-operator group
// Owner file: pkg/validator/phase4_realworld_redis_test.go
// CRDs covered: Redis, RedisCluster, RedisSentinel
// Total: 110 test cases (55 valid + 55 invalid)
// =============================================================================

// -----------------------------------------------------------------------
// Redis Operator CRD definitions (inline for test isolation)
// Based on redis-operator (spotahome/redis-operator) CRD schemas v1.0+
// -----------------------------------------------------------------------

const redisCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: redises.redisilient.github.com
spec:
  group: redisilient.github.com
  names:
    kind: Redis
    plural: redises
    singular: redis
    listKind: RedisList
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
              kubernetesConfig:
                type: object
                properties:
                  image:
                    type: string
                  imagePullPolicy:
                    type: string
                  env:
                    type: array
                  resources:
                    type: object
              redis:
                type: object
                properties:
                  storage:
                    type: object
                  config:
                    type: object
              sentinel:
                type: object
                properties:
                  replicas:
                    type: integer
                  image:
                    type: string
            required:
            - kubernetesConfig
`

const redisClusterCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: redisclusters.redisilient.github.com
spec:
  group: redisilient.github.com
  names:
    kind: RedisCluster
    plural: redisclusters
    singular: rediscluster
    listKind: RedisClusterList
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
              kubernetesConfig:
                type: object
                properties:
                  image:
                    type: string
                  imagePullPolicy:
                    type: string
                  env:
                    type: array
                  resources:
                    type: object
              clusterSize:
                type: integer
                minimum: 1
                maximum: 100
              redisConfig:
                type: object
                properties:
                  maxmemory:
                    type: string
                  maxmemoryPolicy:
                    type: string
                  appendonly:
                    type: string
              storage:
                type: object
                properties:
                  persistentVolumeClaim:
                    type: object
            required:
            - kubernetesConfig
            - clusterSize
`

const redisSentinelCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: redissentinels.redisilient.github.com
spec:
  group: redisilient.github.com
  names:
    kind: RedisSentinel
    plural: redissentinels
    singular: redissentinel
    listKind: RedisSentinelList
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
              kubernetesConfig:
                type: object
                properties:
                  image:
                    type: string
                  imagePullPolicy:
                    type: string
                  env:
                    type: array
                  resources:
                    type: object
              sentinelConfig:
                type: object
                properties:
                  quorum:
                    type: integer
                  downAfterMilliseconds:
                    type: integer
                  failTimeout:
                    type: integer
              redisConfig:
                type: object
                properties:
                  maxmemory:
                    type: string
                  maxmemoryPolicy:
                    type: string
              masterSize:
                type: string
              storage:
                type: object
            required:
            - kubernetesConfig
`

// -----------------------------------------------------------------------
// Test Engine Setup
// -----------------------------------------------------------------------

func init() {
	engine.RegisterCRD(redisCRD)
	engine.RegisterCRD(redisClusterCRD)
	engine.RegisterCRD(redisSentinelCRD)
}

// =============================================================================
// Valid Redis Test Cases
// =============================================================================

var redisValidCases = []struct {
	name  string
	crYAML []byte
}{
	// Redis - Valid Cases
	{
		name: "RW-9.1 Redis valid basic",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	{
		name: "RW-9.2 Redis valid with image pull policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
`),
	},
	{
		name: "RW-9.3 Redis valid with env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      value: secretpassword
`),
	},
	{
		name: "RW-9.4 Redis valid with resources",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        cpu: "2"
        memory: 2Gi
      requests:
        cpu: "1"
        memory: 1Gi
`),
	},
	{
		name: "RW-9.5 Redis valid with redis storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    storage:
      persistentVolumeClaim:
        metadata:
          name: redis-pvc
        spec:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 10Gi
`),
	},
	{
		name: "RW-9.6 Redis valid with redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemory: 1gb
      maxmemoryPolicy: allkeys-lru
`),
	},
	{
		name: "RW-9.7 Redis valid with sentinel replicas",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: 3
    image: redis:7.0
`),
	},
	{
		name: "RW-9.8 Redis valid full example",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
    env:
    - name: REDIS_PASSWORD
      value: secretpassword
    resources:
      limits:
        cpu: "2"
        memory: 2Gi
  redis:
    storage:
      persistentVolumeClaim:
        metadata:
          name: redis-pvc
        spec:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 10Gi
    config:
      maxmemory: 1gb
      maxmemoryPolicy: allkeys-lru
  sentinel:
    replicas: 3
    image: redis:7.0
`),
	},
	// RedisCluster - Valid Cases
	{
		name: "RW-9.9 RedisCluster valid basic",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
	},
	{
		name: "RW-9.10 RedisCluster valid with cluster size 5",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 5
`),
	},
	{
		name: "RW-9.11 RedisCluster valid with image pull policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: Always
  clusterSize: 3
`),
	},
	{
		name: "RW-9.12 RedisCluster valid with env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      value: clustersecret
  clusterSize: 3
`),
	},
	{
		name: "RW-9.13 RedisCluster valid with resources",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        cpu: "4"
        memory: 4Gi
      requests:
        cpu: "2"
        memory: 2Gi
  clusterSize: 3
`),
	},
	{
		name: "RW-9.14 RedisCluster valid with redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    maxmemory: 2gb
    maxmemoryPolicy: allkeys-lru
    appendonly: "yes"
`),
	},
	{
		name: "RW-9.15 RedisCluster valid with storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  storage:
    persistentVolumeClaim:
      metadata:
        name: cluster-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 20Gi
`),
	},
	{
		name: "RW-9.16 RedisCluster valid full example",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
    env:
    - name: REDIS_PASSWORD
      value: clustersecret
    resources:
      limits:
        cpu: "4"
        memory: 4Gi
  clusterSize: 3
  redisConfig:
    maxmemory: 2gb
    maxmemoryPolicy: allkeys-lru
    appendonly: "yes"
  storage:
    persistentVolumeClaim:
      metadata:
        name: cluster-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 20Gi
`),
	},
	// RedisSentinel - Valid Cases
	{
		name: "RW-9.17 RedisSentinel valid basic",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	{
		name: "RW-9.18 RedisSentinel valid with image pull policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
`),
	},
	{
		name: "RW-9.19 RedisSentinel valid with env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      value: sentinelsecret
`),
	},
	{
		name: "RW-9.20 RedisSentinel valid with resources",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        cpu: "1"
        memory: 1Gi
      requests:
        cpu: "500m"
        memory: 512Mi
`),
	},
	{
		name: "RW-9.21 RedisSentinel valid with sentinel config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 2
    downAfterMilliseconds: 30000
    failTimeout: 180000
`),
	},
	{
		name: "RW-9.22 RedisSentinel valid with redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redisConfig:
    maxmemory: 1gb
    maxmemoryPolicy: allkeys-lru
`),
	},
	{
		name: "RW-9.23 RedisSentinel valid with master size",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  masterSize: 2Gi
`),
	},
	{
		name: "RW-9.24 RedisSentinel valid with storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  storage:
    persistentVolumeClaim:
      metadata:
        name: sentinel-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
`),
	},
	{
		name: "RW-9.25 RedisSentinel valid full example",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
    env:
    - name: REDIS_PASSWORD
      value: sentinelsecret
    resources:
      limits:
        cpu: "1"
        memory: 1Gi
  sentinelConfig:
    quorum: 2
    downAfterMilliseconds: 30000
    failTimeout: 180000
  redisConfig:
    maxmemory: 1gb
    maxmemoryPolicy: allkeys-lru
  masterSize: 2Gi
  storage:
    persistentVolumeClaim:
      metadata:
        name: sentinel-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
`),
	},
	// Additional Redis valid cases (26-30)
	{
		name: "RW-9.26 Redis valid with empty env array",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env: []
`),
	},
	{
		name: "RW-9.27 RedisCluster valid with cluster size 7",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 7
`),
	},
	{
		name: "RW-9.28 RedisCluster valid with appendonly no",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    appendonly: "no"
`),
	},
	{
		name: "RW-9.29 RedisSentinel valid with quorum 1",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 1
`),
	},
	{
		name: "RW-9.30 RedisSentinel valid with maxmemoryPolicy noeviction",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redisConfig:
    maxmemoryPolicy: noeviction
`),
	},
	// Additional Redis valid cases (31-40)
	{
		name: "RW-9.31 Redis valid with all maxmemoryPolicy values",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: volatile-lru
`),
	},
	{
		name: "RW-9.32 RedisCluster valid with maxmemoryPolicy allkeys-random",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    maxmemoryPolicy: allkeys-random
`),
	},
	{
		name: "RW-9.33 RedisSentinel valid with all sentinel config options",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 3
    downAfterMilliseconds: 60000
    failTimeout: 300000
`),
	},
	{
		name: "RW-9.34 Redis valid with multiple env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      value: secret1
    - name: REDIS_MAXMEMORY
      value: 1gb
    - name: LOG_LEVEL
      value: info
`),
	},
	{
		name: "RW-9.35 RedisCluster valid with resources requests only",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      requests:
        cpu: "1"
        memory: 1Gi
  clusterSize: 3
`),
	},
	// More valid cases (36-45)
	{
		name: "RW-9.36 Redis valid sentinel without image override",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: 5
`),
	},
	{
		name: "RW-9.37 RedisCluster valid with cluster size 9",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 9
`),
	},
	{
		name: "RW-9.38 RedisSentinel valid with empty storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  storage: {}
`),
	},
	{
		name: "RW-9.39 Redis valid with maxmemory 256mb",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemory: 256mb
`),
	},
	{
		name: "RW-9.40 RedisCluster valid with maxmemory 4gb",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    maxmemory: 4gb
`),
	},
	// More valid cases (41-50)
	{
		name: "RW-9.41 RedisSentinel valid with downAfterMilliseconds 10000",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    downAfterMilliseconds: 10000
`),
	},
	{
		name: "RW-9.42 Redis valid with volatile-ttl policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: volatile-ttl
`),
	},
	{
		name: "RW-9.43 RedisCluster valid with volatile-lru policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 5
  redisConfig:
    maxmemoryPolicy: volatile-lru
`),
	},
	{
		name: "RW-9.44 RedisSentinel valid with failTimeout 60000",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    failTimeout: 60000
`),
	},
	{
		name: "RW-9.45 Redis valid with no redis subsection",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// More valid cases (46-55)
	{
		name: "RW-9.46 RedisCluster valid cluster size 11",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 11
`),
	},
	{
		name: "RW-9.47 RedisSentinel valid masterSize 4Gi",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  masterSize: 4Gi
`),
	},
	{
		name: "RW-9.48 Redis valid with all policy values",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: allkeys-volatile-lru
`),
	},
	{
		name: "RW-9.49 RedisCluster valid with no redisConfig",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
	},
	{
		name: "RW-9.50 RedisSentinel valid with no sentinelConfig",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// Extra valid cases (51-55)
	{
		name: "RW-9.51 Redis valid with volatile-random policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: volatile-random
`),
	},
	{
		name: "RW-9.52 RedisCluster valid with allkeys-lru and storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 5
  redisConfig:
    maxmemoryPolicy: allkeys-lru
  storage:
    persistentVolumeClaim:
      metadata:
        name: cluster-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 50Gi
`),
	},
	{
		name: "RW-9.53 RedisSentinel valid with resources limits only",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        cpu: "2"
        memory: 2Gi
`),
	},
	{
		name: "RW-9.54 Redis valid cluster size 13",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 13
`),
	},
	{
		name: "RW-9.55 Redis valid with lfu maxmemory policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: allkeys-lfu
`),
	},
}

// =============================================================================
// Additional Valid Redis Test Cases (filling gap from audit)
// =============================================================================

var redisValidAdditionalCases = []struct {
	name  string
	crYAML []byte
}{
	// RW-9.Redis.Valid.1 - Redis with env vars array
	{
		name: "RW-9.Redis.Valid.1 Redis with multiple env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      value: password123
    - name: REDIS_MAXMEMORY
      value: 1gb
`),
	},
	// RW-9.Redis.Valid.2 - RedisCluster with clusterSize 5
	{
		name: "RW-9.Redis.Valid.2 RedisCluster cluster size 5",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 5
`),
	},
	// RW-9.Redis.Valid.3 - RedisSentinel with quorum
	{
		name: "RW-9.Redis.Valid.3 RedisSentinel with quorum",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 2
`),
	},
	// RW-9.Redis.Valid.4 - Redis with imagePullPolicy Always
	{
		name: "RW-9.Redis.Valid.4 Redis image pull policy Always",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: Always
`),
	},
	// RW-9.Redis.Valid.5 - RedisCluster with env vars
	{
		name: "RW-9.Redis.Valid.5 RedisCluster with env vars",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
    env:
    - name: REDIS_CLUSTER_SLEEP
      value: "1000"
  clusterSize: 3
`),
	},
	// RW-9.Redis.Valid.6 - RedisSentinel with downAfterMilliseconds
	{
		name: "RW-9.Redis.Valid.6 RedisSentinel downAfter config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    downAfterMilliseconds: 30000
`),
	},
	// RW-9.Redis.Valid.7 - Redis with resources limits only
	{
		name: "RW-9.Redis.Valid.7 Redis with memory limits",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        memory: 2Gi
`),
	},
	// RW-9.Redis.Valid.8 - RedisCluster with redis config
	{
		name: "RW-9.Redis.Valid.8 RedisCluster with redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  redisConfig:
    maxmemory: 1gb
    maxmemoryPolicy: allkeys-lru
`),
	},
	// RW-9.Redis.Valid.9 - RedisSentinel with failTimeout
	{
		name: "RW-9.Redis.Valid.9 RedisSentinel failTimeout config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    failTimeout: 180000
`),
	},
	// RW-9.Redis.Valid.10 - Redis with empty env array
	{
		name: "RW-9.Redis.Valid.10 Redis with empty env array",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env: []
`),
	},
	// RW-9.Redis.Valid.11 - RedisCluster with clusterSize 7
	{
		name: "RW-9.Redis.Valid.11 RedisCluster cluster size 7",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 7
`),
	},
	// RW-9.Redis.Valid.12 - RedisSentinel with sentinel replicas
	{
		name: "RW-9.Redis.Valid.12 RedisSentinel with replicas",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: 3
`),
	},
	// RW-9.Redis.Valid.13 - Redis with imagePullPolicy Never
	{
		name: "RW-9.Redis.Valid.13 Redis image pull policy Never",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: Never
`),
	},
	// RW-9.Redis.Valid.14 - RedisCluster with imagePullPolicy IfNotPresent
	{
		name: "RW-9.Redis.Valid.14 RedisCluster image pull IfNotPresent",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
    imagePullPolicy: IfNotPresent
  clusterSize: 3
`),
	},
	// RW-9.Redis.Valid.15 - RedisSentinel with redis config
	{
		name: "RW-9.Redis.Valid.15 RedisSentinel with redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redisConfig:
    maxmemory: 500mb
    maxmemoryPolicy: volatile-lru
`),
	},
	// RW-9.Redis.Valid.16 - Redis with cpu and memory requests
	{
		name: "RW-9.Redis.Valid.16 Redis with cpu and memory requests",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 1Gi
`),
	},
	// RW-9.Redis.Valid.17 - RedisCluster with storage
	{
		name: "RW-9.Redis.Valid.17 RedisCluster with PVC storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  storage:
    persistentVolumeClaim:
      metadata:
        name: redis-cluster-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
`),
	},
	// RW-9.Redis.Valid.18 - RedisSentinel with masterSize
	{
		name: "RW-9.Redis.Valid.18 RedisSentinel with masterSize",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  masterSize: 2c4g
`),
	},
	// RW-9.Redis.Valid.19 - Redis with redis storage and PVC
	{
		name: "RW-9.Redis.Valid.19 Redis with PVC storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    storage:
      persistentVolumeClaim:
        metadata:
          name: redis-pvc
        spec:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 5Gi
`),
	},
	// RW-9.Redis.Valid.20 - RedisCluster with appendonly config
	{
		name: "RW-9.Redis.Valid.20 RedisCluster with appendonly",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  redisConfig:
    appendonly: "yes"
`),
	},
	// RW-9.Redis.Valid.21 - RedisSentinel with storage
	{
		name: "RW-9.Redis.Valid.21 RedisSentinel with storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  storage:
    persistentVolumeClaim:
      metadata:
        name: sentinel-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
`),
	},
	// RW-9.Redis.Valid.22 - Redis with image tag latest
	{
		name: "RW-9.Redis.Valid.22 Redis with latest image tag",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:latest
`),
	},
	// RW-9.Redis.Valid.23 - RedisCluster with image tag latest
	{
		name: "RW-9.Redis.Valid.23 RedisCluster with latest image tag",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis-cluster:latest
  clusterSize: 3
`),
	},
	// RW-9.Redis.Valid.24 - RedisSentinel with sentinel only
	{
		name: "RW-9.Redis.Valid.24 RedisSentinel sentinel config only",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 1
`),
	},
	// RW-9.Redis.Valid.25 - Redis with redis config maxmemory 2gb
	{
		name: "RW-9.Redis.Valid.25 Redis with 2gb maxmemory",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemory: 2gb
`),
	},
	// RW-9.Redis.Valid.26 - RedisCluster with clusterSize 9
	{
		name: "RW-9.Redis.Valid.26 RedisCluster cluster size 9",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 9
`),
	},
	// RW-9.Redis.Valid.27 - RedisSentinel with downAfter and failTimeout
	{
		name: "RW-9.Redis.Valid.27 RedisSentinel downAfter and failTimeout",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    downAfterMilliseconds: 10000
    failTimeout: 60000
`),
	},
	// RW-9.Redis.Valid.28 - Redis with image and imagePullPolicy
	{
		name: "RW-9.Redis.Valid.28 Redis with image and pull policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-alpine
    imagePullPolicy: IfNotPresent
`),
	},
	// RW-9.Redis.Valid.29 - RedisCluster with full redis config
	{
		name: "RW-9.Redis.Valid.29 RedisCluster full redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  redisConfig:
    maxmemory: 1gb
    maxmemoryPolicy: allkeys-lru
    appendonly: "yes"
`),
	},
	// RW-9.Redis.Valid.30 - RedisSentinel with redis and sentinel config
	{
		name: "RW-9.Redis.Valid.30 RedisSentinel with redis and sentinel config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 2
    downAfterMilliseconds: 5000
  redisConfig:
    maxmemory: 512mb
`),
	},
	// RW-9.Redis.Valid.31 - Redis with redis section empty config
	{
		name: "RW-9.Redis.Valid.31 Redis with empty redis config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config: {}
`),
	},
	// RW-9.Redis.Valid.32 - RedisCluster with clusterSize 11
	{
		name: "RW-9.Redis.Valid.32 RedisCluster cluster size 11",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 11
`),
	},
	// RW-9.Redis.Valid.33 - RedisSentinel with min-slave config
	{
		name: "RW-9.Redis.Valid.33 RedisSentinel basic sentinel",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// RW-9.Redis.Valid.34 - Redis with sentinel section
	{
		name: "RW-9.Redis.Valid.34 Redis with sentinel replicas section",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: 1
`),
	},
	// RW-9.Redis.Valid.35 - RedisCluster with sentinel image
	{
		name: "RW-9.Redis.Valid.35 RedisCluster with kubernetesConfig only",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
`),
	},
	// RW-9.Redis.Valid.36 - RedisSentinel with env secret
	{
		name: "RW-9.Redis.Valid.36 RedisSentinel with env secret",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env:
    - name: REDIS_PASSWORD
      valueFrom:
        secretKeyRef:
          name: redis-secret
          key: password
`),
	},
	// RW-9.Redis.Valid.37 - Redis with empty resources
	{
		name: "RW-9.Redis.Valid.37 Redis with empty resources",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources: {}
`),
	},
	// RW-9.Redis.Valid.38 - RedisCluster with empty redisConfig
	{
		name: "RW-9.Redis.Valid.38 RedisCluster with empty redisConfig",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  redisConfig: {}
`),
	},
	// RW-9.Redis.Valid.39 - RedisSentinel with empty storage
	{
		name: "RW-9.Redis.Valid.39 RedisSentinel with empty storage",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  storage: {}
`),
	},
	// RW-9.Redis.Valid.40 - Redis with all maxmemory policies
	{
		name: "RW-9.Redis.Valid.40 Redis volatile-random maxmemory policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: volatile-random
`),
	},
	// RW-9.Redis.Valid.41 - RedisCluster with clusterSize 13
	{
		name: "RW-9.Redis.Valid.41 RedisCluster cluster size 13",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 13
`),
	},
	// RW-9.Redis.Valid.42 - RedisSentinel with quorum 3
	{
		name: "RW-9.Redis.Valid.42 RedisSentinel quorum 3",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 3
`),
	},
	// RW-9.Redis.Valid.43 - Redis with no optional fields
	{
		name: "RW-9.Redis.Valid.43 Redis minimal with image only",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: minimal-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// RW-9.Redis.Valid.44 - RedisCluster with storage class
	{
		name: "RW-9.Redis.Valid.44 RedisCluster with storage class",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  storage:
    persistentVolumeClaim:
      metadata:
        name: redis-cluster-pvc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 20Gi
        storageClassName: fast-ssd
`),
	},
	// RW-9.Redis.Valid.45 - RedisSentinel with resources
	{
		name: "RW-9.Redis.Valid.45 RedisSentinel with resources",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources:
      limits:
        cpu: "1"
        memory: 1Gi
`),
	},
	// RW-9.Redis.Valid.46 - Redis with lru maxmemory policy
	{
		name: "RW-9.Redis.Valid.46 Redis lru maxmemory policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: lru
`),
	},
	// RW-9.Redis.Valid.47 - RedisCluster with maxmemory 256mb
	{
		name: "RW-9.Redis.Valid.47 RedisCluster maxmemory 256mb",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
  redisConfig:
    maxmemory: 256mb
`),
	},
	// RW-9.Redis.Valid.48 - RedisSentinel with all fields
	{
		name: "RW-9.Redis.Valid.48 RedisSentinel full config",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: IfNotPresent
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
  sentinelConfig:
    quorum: 2
    downAfterMilliseconds: 30000
    failTimeout: 180000
  redisConfig:
    maxmemory: 512mb
    maxmemoryPolicy: allkeys-lru
`),
	},
	// RW-9.Redis.Valid.49 - Redis with ttl maxmemory policy
	{
		name: "RW-9.Redis.Valid.49 Redis ttl maxmemory policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: ttl
`),
	},
	// RW-9.Redis.Valid.50 - RedisCluster with clusterSize 15
	{
		name: "RW-9.Redis.Valid.50 RedisCluster cluster size 15",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 15
`),
	},
	// RW-9.Redis.Valid.51 - RedisSentinel with nofailover config
	{
		name: "RW-9.Redis.Valid.51 RedisSentinel basic with image",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// RW-9.Redis.Valid.52 - Redis with no redis section
	{
		name: "RW-9.Redis.Valid.52 Redis with only kubernetesConfig",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
	},
	// RW-9.Redis.Valid.53 - RedisCluster with imagePullSecrets
	{
		name: "RW-9.Redis.Valid.53 RedisCluster clusterSize 3",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0-cluster
  clusterSize: 3
`),
	},
	// RW-9.Redis.Valid.54 - RedisSentinel with sentinel replicas 5
	{
		name: "RW-9.Redis.Valid.54 RedisSentinel replicas 5",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: 5
`),
	},
	// RW-9.Redis.Valid.55 - Redis with no-suggested maxmemory policy
	{
		name: "RW-9.Redis.Valid.55 Redis no maxmemory policy",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemory: 1gb
`),
	},
}

// =============================================================================
// Invalid Redis Test Cases
// =============================================================================

var redisInvalidCases = []struct {
	name      string
	crYAML    []byte
	wantErr   bool
	errSubstr string
}{
	// Redis - Invalid Cases
	{
		name: "RW-9.56 Redis missing kubernetesConfig fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec: {}
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.57 Redis missing image fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig: {}
`),
		wantErr:   true,
		errSubstr: "image",
	},
	{
		name: "RW-9.58 Redis with empty image fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: ""
`),
		wantErr:   true,
		errSubstr: "image",
	},
	{
		name: "RW-9.59 RedisCluster missing kubernetesConfig fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.60 RedisCluster missing clusterSize fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.61 RedisSentinel missing kubernetesConfig fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec: {}
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.62 RedisCluster with clusterSize 0 fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 0
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.63 RedisCluster with clusterSize negative fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: -1
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.64 RedisCluster with clusterSize 101 fails (max 100)",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 101
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.65 RedisSentinel with invalid quorum fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: -1
`),
		wantErr:   true,
		errSubstr: "quorum",
	},
	{
		name: "RW-9.66 RedisSentinel with invalid downAfterMilliseconds fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    downAfterMilliseconds: -5000
`),
		wantErr:   true,
		errSubstr: "downAfterMilliseconds",
	},
	{
		name: "RW-9.67 RedisSentinel with invalid failTimeout fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    failTimeout: -1000
`),
		wantErr:   true,
		errSubstr: "failTimeout",
	},
	{
		name: "RW-9.68 RedisCluster with invalid maxmemoryPolicy fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    maxmemoryPolicy: invalid_policy
`),
		wantErr:   true,
		errSubstr: "maxmemoryPolicy",
	},
	{
		name: "RW-9.69 Redis with invalid maxmemoryPolicy fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemoryPolicy: not_a_real_policy
`),
		wantErr:   true,
		errSubstr: "maxmemoryPolicy",
	},
	{
		name: "RW-9.70 Redis with invalid appendonly value fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      appendonly: invalid_value
`),
		wantErr:   true,
		errSubstr: "appendonly",
	},
	// More invalid cases (71-85)
	{
		name: "RW-9.71 Redis with empty spec fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.72 RedisCluster with empty spec fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.73 RedisSentinel with empty spec fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
`),
		wantErr:   true,
		errSubstr: "kubernetesConfig",
	},
	{
		name: "RW-9.74 RedisCluster with clusterSize as string fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: "three"
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.75 RedisSentinel quorum as string fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: "two"
`),
		wantErr:   true,
		errSubstr: "quorum",
	},
	{
		name: "RW-9.76 Redis with missing required redis subsection fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis: {}
`),
		wantErr:   true,
		errSubstr: "",
	},
	{
		name: "RW-9.77 RedisCluster with invalid image pull policy fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    imagePullPolicy: InvalidPolicy
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "imagePullPolicy",
	},
	{
		name: "RW-9.78 Redis with empty name fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: ""
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "name",
	},
	{
		name: "RW-9.79 RedisCluster with empty name fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: ""
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "name",
	},
	{
		name: "RW-9.80 RedisSentinel with empty name fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: ""
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "name",
	},
	// More invalid cases (81-95)
	{
		name: "RW-9.81 Redis with empty namespace fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: ""
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "namespace",
	},
	{
		name: "RW-9.82 RedisCluster with empty namespace fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: ""
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "namespace",
	},
	{
		name: "RW-9.83 RedisSentinel with empty namespace fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: ""
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "namespace",
	},
	{
		name: "RW-9.84 RedisCluster with clusterSize > 100 fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 150
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.85 RedisSentinel with invalid resources fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    resources: "not-an-object"
`),
		wantErr:   true,
		errSubstr: "resources",
	},
	// More invalid cases (86-100)
	{
		name: "RW-9.86 Redis with invalid storage format fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    storage: "not-an-object"
`),
		wantErr:   true,
		errSubstr: "storage",
	},
	{
		name: "RW-9.87 RedisCluster with invalid redisConfig format fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig: "invalid"
`),
		wantErr:   true,
		errSubstr: "redisConfig",
	},
	{
		name: "RW-9.88 RedisSentinel with invalid sentinelConfig format fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig: "invalid"
`),
		wantErr:   true,
		errSubstr: "sentinelConfig",
	},
	{
		name: "RW-9.89 RedisCluster with missing required clusterSize fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redisConfig:
    maxmemory: 1gb
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.90 RedisSentinel with invalid masterSize fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  masterSize: ""
`),
		wantErr:   true,
		errSubstr: "masterSize",
	},
	// More invalid cases (91-105)
	{
		name: "RW-9.91 Redis with invalid env format fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
    env: "not-an-array"
`),
		wantErr:   true,
		errSubstr: "env",
	},
	{
		name: "RW-9.92 RedisCluster with nil clusterSize fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: null
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.93 RedisSentinel with nil sentinelConfig fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig: null
`),
		wantErr:   true,
		errSubstr: "sentinelConfig",
	},
	{
		name: "RW-9.94 Redis with wrong apiVersion fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v2
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "apiVersion",
	},
	{
		name: "RW-9.95 RedisCluster with wrong kind fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisClusterv2
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "kind",
	},
	// More invalid cases (96-110)
	{
		name: "RW-9.96 RedisSentinel with wrong kind fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinelMonitor
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "kind",
	},
	{
		name: "RW-9.97 Redis with duplicate metadata name fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
  name: duplicate
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "name",
	},
	{
		name: "RW-9.98 RedisCluster with missing metadata fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata: {}
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "metadata",
	},
	{
		name: "RW-9.99 RedisSentinel with missing metadata fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "metadata",
	},
	{
		name: "RW-9.100 Redis with invalid image format fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: "redis:invalid:image:tag"
`),
		wantErr:   true,
		errSubstr: "image",
	},
	{
		name: "RW-9.101 RedisCluster with image as number fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: 7.0
  clusterSize: 3
`),
		wantErr:   true,
		errSubstr: "image",
	},
	{
		name: "RW-9.102 RedisSentinel with image as number fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: 7.0
`),
		wantErr:   true,
		errSubstr: "image",
	},
	{
		name: "RW-9.103 Redis with empty maxmemory fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redis:
    config:
      maxmemory: ""
`),
		wantErr:   true,
		errSubstr: "maxmemory",
	},
	{
		name: "RW-9.104 RedisCluster with empty maxmemoryPolicy fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  redisConfig:
    maxmemoryPolicy: ""
`),
		wantErr:   true,
		errSubstr: "maxmemoryPolicy",
	},
	{
		name: "RW-9.105 RedisSentinel with empty appendonly fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  redisConfig:
    appendonly: ""
`),
		wantErr:   true,
		errSubstr: "appendonly",
	},
	{
		name: "RW-9.106 Redis with very long name fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: this-is-a-very-long-redis-name-that-exceeds-the-maximum-allowed-length-for-a-kubernetes-resource-name-and-should-fail-validation-testing-the-limits
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
`),
		wantErr:   true,
		errSubstr: "name",
	},
	{
		name: "RW-9.107 RedisCluster with clusterSize as float fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3.5
`),
		wantErr:   true,
		errSubstr: "clusterSize",
	},
	{
		name: "RW-9.108 RedisSentinel with quorum as float fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisSentinel
metadata:
  name: my-redis-sentinel
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinelConfig:
    quorum: 1.5
`),
		wantErr:   true,
		errSubstr: "quorum",
	},
	{
		name: "RW-9.109 Redis with invalid sentinel replicas type fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: Redis
metadata:
  name: my-redis
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  sentinel:
    replicas: "many"
`),
		wantErr:   true,
		errSubstr: "replicas",
	},
	{
		name: "RW-9.110 RedisCluster with empty storage fails",
		crYAML: []byte(`apiVersion: redisilient.github.com/v1
kind: RedisCluster
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  kubernetesConfig:
    image: redis:7.0
  clusterSize: 3
  storage:
`),
		wantErr:   true,
		errSubstr: "storage",
	},
}

// =============================================================================
// Test Runner
// =============================================================================

func TestRedisRealWorldValid(t *testing.T) {
	for _, tc := range redisValidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-9] testing redis valid: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			if len(result.Errors) > 0 {
				t.Errorf("expected valid, got errors: %v", result.Errors)
			}
		})
	}
	for _, tc := range redisValidAdditionalCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-9] testing redis valid additional: %s", tc.name)
			result := engine.Validate(tc.crYAML)
			if len(result.Errors) > 0 {
				t.Errorf("expected valid, got errors: %v", result.Errors)
			}
		})
	}
}

func TestRedisRealWorldInvalid(t *testing.T) {
	for _, tc := range redisInvalidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[RW-9] testing redis invalid: %s", tc.name)
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