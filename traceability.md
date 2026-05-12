# Traceability Map — k8-manifest-validator (swe-k8-manifest-validator branch)

**Maintainer:** Architect (OX 🦾)
**Last updated:** 2026-05-11 (AWU-18.1–18.6 wave1-qa verified ✅)
**Branch:** swe-k8-manifest-validator

---

## 1. Phase 3.1 Deep-Nested AWU Status

| AWU ID | Title | Test File | Status | Notes |
|---|---|---|---|---|
| AWU-17.1 | Deep Nested — EnvVar.valueFrom chain | `phase3_deepenv_test.go` | ✅ Implemented | secretKeyRef + configMapKeyRef, ≥4 cases |
| AWU-17.2 | Deep Nested — EnvFrom.chain | `phase3_deepenvfrom_test.go` | ✅ Implemented | configMapRef + secretRef, Optional passthrough, ≥4 cases |
| AWU-17.3 | Deep Nested — ConfigMap/Secret Volume items | `phase3_deepitems_test.go` | ✅ Implemented | items array passthrough + azureFile.secretName, ≥5 cases |
| AWU-17.4 | Deep Nested — Lifecycle Handler sub-fields | `phase3_deeplifecycle_test.go` | ✅ Implemented | postStart/preStop exec/httpGet/tcpSocket, ≥4 cases |
| **AWU-17.5** | **Deep Nested — Probe sub-fields** | **`phase3_deepprobe_test.go`** | **✅ Complete** | **Final AWU; ≥5 cases covering liveness/readiness/startup probe sub-fields** |

---

## 2. AWU-17.5 Acceptance Criteria

Per PRD Review R2 (2026-05-10):

- [ ] ≥5 test cases covering probe sub-field traversal:
  1. **Valid:** `livenessProbe.tcpSocket.port` set (Required:true)
  2. **Invalid:** `livenessProbe.tcpSocket.port` empty/zero (Required:true → error)
  3. **Valid:** `readinessProbe.httpGet` with host + path + port + scheme
  4. **Valid:** `startupProbe.exec.command` with array of strings
  5. **Invalid:** `livenessProbe.httpGet.port` empty/zero (Required:true → error)
  6. **Valid:** probe with `failureThreshold`, `periodSeconds`, `timeoutSeconds` (all Optional)
- [ ] All tests pass: `go test ./pkg/validator/... -run Phase3Deep`
- [ ] Correct modifier helpers (`withLivenessProbe`, `withReadinessProbe`, `withStartupProbe`)
- [ ] No new dependencies

---

## 3. Codegen Reference for Probes

Probe sub-field entries confirmed in `pkg/validator/generated_structural.go` (lines ~505–765):

```
livenessProbe.tcpSocket.port        Required:true
livenessProbe.httpGet.port         Required:true
livenessProbe.exec.command         Type:array, Required:false
readinessProbe.tcpSocket.port      Required:true
readinessProbe.httpGet.port        Required:true
readinessProbe.exec.command        Type:array, Required:false
startupProbe.tcpSocket.port        Required:true
startupProbe.httpGet.port          Required:true
startupProbe.exec.command          Type:array, Required:false
```

All three probe types (liveness, readiness, startup) share identical sub-field structure.

---

## 4. Dependency Chain

```
AWU-16.3 (Phase 3 integration tests — complete)
  └─ AWU-17.1 (deep env valueFrom) — complete
        └─ AWU-17.2 (deep envFrom) — complete
              └─ AWU-17.3 (configmap/secret items) — complete
                    └─ AWU-17.4 (lifecycle handler) — complete
                          └─ AWU-17.5 (probe sub-fields) — IN PROGRESS
```

---

---

## 6. AWU-18 Series — Volume Source Type Gap Coverage

| AWU ID | Title | Test File | Status | Wave | Codegen Lines |
|---|---|---|---|---|---|
| AWU-18.1 | Azure File Volume — shareName + secretName Required | `phase3_volume_azurefile_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.2 | AWS EBS Volume — volumeID Required | `phase3_volume_awsebs_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.3 | GCE PD Volume — pdName Required | `phase3_volume_gcepd_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.4 | Cinder Volume — volumeID Required | `phase3_volume_cinder_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.5 | PhotonPD Volume — pdID Required | `phase3_volume_photonpd_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.6 | vSphere Volume — volumePath Required | `phase3_volume_vsphere_test.go` | ✅ QA Verified (Wave 1) | wave1-qa 2026-05-11 |
| AWU-18.8 | NFS Volume — server + path Required | `phase3_volume_nfs_test.go` | 📋 Planned | Wave 2 | 542–545 |
| AWU-18.9 | GlusterFS Volume — endpoints + path Required | `phase3_volume_glusterfs_test.go` | 📋 Planned | Wave 2 | 425–428 |
| AWU-18.10 | CephFS Volume — monitors array Required | `phase3_volume_cephfs_test.go` | 📋 Planned | Wave 2 | 327–333 |
| AWU-18.11 | iSCSI Volume — targetPortal + iqn + lun Required | `phase3_volume_iscsi_test.go` | 📋 Planned | Wave 2 | 455–466 |
| AWU-18.12 | RBD Volume — image + monitors Required | `phase3_volume_rbd_test.go` | 📋 Planned | Wave 3 | 631–639 |
| AWU-18.13 | ScaleIO Volume — gateway + system + secretRef Required | `phase3_volume_scaleio_test.go` | 📋 Planned | Wave 3 | 680–690 |
| AWU-18.14 | CSI Volume — driver Required | `phase3_volume_csi_test.go` | 📋 Planned | Wave 4 | 355–360 |
| AWU-18.15 | FlexVolume — driver Required | `phase3_volume_flexvol_test.go` | 📋 Planned | Wave 4 | 403–408 |
| AWU-18.16 | HostPath Volume — path Required | `phase3_volume_hostpath_test.go` | 📋 Planned | Wave 5 | 435–437 |
| AWU-18.17 | GitRepo Volume — repository Required | `phase3_volume_gitrepo_test.go` | 📋 Planned | Wave 5 | 421–424 |
| AWU-18.18 | Quobyte Volume — registry + volume Required | `phase3_volume_quobyte_test.go` | 📋 Planned | Wave 5 | 624–630 |
| AWU-18.19 | Portworx Volume — volumeID Required | `phase3_volume_portworx_test.go` | 📋 Planned | Wave 5 | 584–587 |
| AWU-18.20 | FC Volume — all Optional (passthrough) | `phase3_volume_fc_test.go` | 📋 Planned | Wave 6 | 394–399 |

**Deferred:** AWU-18.7 (azureDisk) — all Optional per codegen; AWU-18.7 (flocker) — all Optional per codegen.

**Total:** 88 test cases across 20 AWU files. All independent — no inter-AWU dependencies.

---

## 7. AWU-18 Acceptance Criteria (Volume Source Types)

- [ ] AWU-18.1: ≥5 cases — azureFile.shareName + secretName Required enforcement
- [ ] AWU-18.2: ≥5 cases — awsElasticBlockStore.volumeID Required enforcement
- [x] AWU-18.3: ≥5 cases — gcePersistentDisk.pdName Required enforcement
- [ ] AWU-18.4: ≥5 cases — cinder.volumeID Required enforcement
- [ ] AWU-18.5: ≥4 cases — photonPersistentDisk.pdID Required enforcement
- [ ] AWU-18.6: ≥5 cases — vsphereVolume.volumePath Required enforcement
- [ ] AWU-18.8: ≥6 cases — nfs.server + nfs.path Required enforcement
- [ ] AWU-18.9: ≥6 cases — glusterfs.endpoints + path Required enforcement
- [ ] AWU-18.10: ≥5 cases — cephfs.monitors array Required enforcement
- [ ] AWU-18.11: ≥7 cases — iscsi.targetPortal + iqn + lun Required enforcement
- [ ] AWU-18.12: ≥6 cases — rbd.image + monitors Required enforcement
- [ ] AWU-18.13: ≥6 cases — scaleIO.gateway + system + secretRef Required enforcement
- [ ] AWU-18.14: ≥5 cases — csi.driver Required enforcement
- [ ] AWU-18.15: ≥5 cases — flexVolume.driver Required enforcement
- [ ] AWU-18.16: ≥4 cases — hostPath.path Required enforcement
- [ ] AWU-18.17: ≥4 cases — gitRepo.repository Required enforcement
- [ ] AWU-18.18: ≥5 cases — quobyte.registry + volume Required enforcement
- [ ] AWU-18.19: ≥4 cases — portworxVolume.volumeID Required enforcement
- [ ] AWU-18.20: ≥4 cases — fc all-Optional passthrough
- [ ] All pass: `go test ./pkg/validator/... -run "AWU-18"`
- [ ] No new dependencies

---

## 8. Reference Files

- PRD Review R2: `/home/node/.openclaw/workspace/github/SWE/swe-k8-manifest-validator/prd_review_deepnested_r2.md`
- Plan fragment: `/home/node/.openclaw/workspace/github/SWE/swe-k8-manifest-validator/plan_fragment.md`
- Lifecycle test (pattern reference): `pkg/validator/phase3_deeplifecycle_test.go`
- Codegen metadata: `pkg/validator/generated_structural.go` (lines ~505–520, ~642–665, ~748–765)
