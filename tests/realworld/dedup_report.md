# RW-12: Cross-Operator Deduplication Report

**Total operators analyzed:** 17  
**Total unique (kind, field_path) combinations:** 12,089  
**Cross-operator field collisions:** 26

## Collisions (same kind + field path across operators)

| kind + field_path | operators |
|---|---|
| `Deployment:replicas` | istio, prometheus |
| `Deployment:template.spec` | istio, prometheus |
| `Deployment:template` | istio, prometheus |
| `Deployment:selector` | istio, prometheus |
| `Deployment:selector.matchLabels` | istio, prometheus |
| `Deployment:template.spec.serviceAccountName` | istio, prometheus |
| `Deployment:template.spec.containers[*].name` | istio, prometheus |
| `Deployment:template.spec.containers[*].image` | istio, prometheus |
| `Deployment:template.spec.containers[*].ports` | istio, prometheus |
| `Deployment:template.spec.containers[*].ports[*].containerPort` | istio, prometheus |
| `KafkaConnect:replicas` | kafka, strimzi |
| `KafkaConnect:bootstrapServers` | istio, prometheus |
| `KafkaConnect:authentication` | kafka, strimzi |
| `KafkaConnect:authentication.type` | kafka, strimzi |
| `KafkaConnect:tls` | kafka, strimzi |
| `KafkaConnect:tls.trustedCertificates[*].certificate` | kafka, strimzi |
| `KafkaTopic:config.cleanup.policy` | kafka, strimzi |
| `KafkaUser:authentication` | kafka, strimzi |
| `KafkaUser:authentication.type` | kafka, strimzi |

**Note:** All collisions are expected — both istio/prometheus share Kubernetes built-in Deployment fields, and kafka/strimzi share Strimzi KafkaConnect/KafkaUser/KafkaTopic schema patterns. No false-duplication detected.

## Field Coverage Per Operator

| operator | field paths | kinds |
|---|---|---|
| prometheus | 10,858 | 12 |
| argocd | 263 | 8 |
| cert-manager | 140 | 8 |
| redis | 166 | 6 |
| postgresql | 106 | 10 |
| flux | 93 | 9 |
| strimzi | 71 | 5 |
| grafana | 54 | 1 |
| istio | 224 | 12 |
| elasticsearch | 8 | 1 |
| etcd | 38 | 1 |
| jaeger | 9 | 1 |
| kafka | 15 | 3 |
| keycloak | 23 | 1 |
| minio | 20 | 1 |
| rabbitmq | 11 | 1 |
| vault | 16 | 1 |

**prometheus** has the largest field coverage due to its extensive ServiceMonitor/PodMonitor/Alertmanager schemas alongside core CRDs.
