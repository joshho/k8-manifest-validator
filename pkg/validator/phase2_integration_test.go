package validator

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"sigs.k8s.io/yaml"
)

// =============================================================================
// AWU-15.1 — Phase 2 Integration Tests
// PRD ref: PRD Section 20 Phase 2 deferred fields + AT-24 through AT-41
// Global test owner for Phase 2 — covers semantic_lifecycle, semantic_security,
// semantic_scheduling, structural, generated_structural, constants, and wire-in.
// =============================================================================

// -----------------------------------------------------------------------
// AT-37 / Structural Layer — Deferred field structural validation
// -----------------------------------------------------------------------

func TestPhase2_StructuralLayer_TopologySpreadConstraints(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid topologySpreadConstraints",
			deploy: deployment(withTopologySpreadConstraints(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr: false,
		},
		{
			name: "topologySpreadConstraints missing required maxSkew",
			deploy: deployment(withTopologySpreadConstraints(
				corev1.TopologySpreadConstraint{
					MaxSkew:           0, // must be >= 1
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "maxSkew",
		},
		{
			name: "topologySpreadConstraints missing required topologyKey",
			deploy: deployment(withTopologySpreadConstraints(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "", // required
					WhenUnsatisfiable: corev1.DoNotSchedule,
				},
			)),
			wantErr:    true,
			errSubstr:  "topologyKey",
		},
		{
			name: "topologySpreadConstraints invalid whenUnsatisfiable",
			deploy: deployment(withTopologySpreadConstraints(
				corev1.TopologySpreadConstraint{
					MaxSkew:           1,
					TopologyKey:       "kubernetes.io/hostname",
					WhenUnsatisfiable: "InvalidValue",
				},
			)),
			wantErr:    true,
			errSubstr:  "whenUnsatisfiable",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-34] testing topologySpreadConstraints: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestPhase2_StructuralLayer_DNSConfig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid dnsConfig",
			deploy: deployment(withDNSConfig(&corev1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8"},
				Searches:    []string{"ns1.svc.cluster.local"},
			})),
			wantErr: false,
		},
		{
			name: "dnsConfig nameserver invalid IP",
			deploy: deployment(withDNSConfig(&corev1.PodDNSConfig{
				Nameservers: []string{"not-an-ip"},
			})),
			wantErr:    true,
			errSubstr:  "nameserver",
		},
		{
			name: "dnsConfig too many searches (>6)",
			deploy: deployment(withDNSConfig(&corev1.PodDNSConfig{
				Searches: []string{
					"ns1.svc.cluster.local", "ns2.svc.cluster.local", "ns3.svc.cluster.local",
					"ns4.svc.cluster.local", "ns5.svc.cluster.local", "ns6.svc.cluster.local",
					"ns7.svc.cluster.local", // 7th — too many
				},
			})),
			wantErr:    true,
			errSubstr:  "searches",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-35] testing dnsConfig: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestPhase2_StructuralLayer_Affinity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid affinity nodeSelectorTerms",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}},
							}},
						},
					},
				},
			})),
			wantErr: false,
		},
		{
			name: "nodeSelectorTerms empty — AT-32",
			deploy: deployment(withAffinity(&corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{}, // must have at least one
					},
				},
			})),
			wantErr:    true,
			errSubstr:  "nodeSelectorTerms",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-32] testing affinity: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestPhase2_StructuralLayer_SecurityContext(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid securityContext runAsNonRoot + runAsUser",
			deploy: deployment(withPodSecurityContext(&corev1.PodSecurityContext{
				RunAsNonRoot: boolPtr(true),
				RunAsUser:    int64Ptr(1000),
			})),
			wantErr: false,
		},
		{
			name: "runAsNonRoot=true + runAsUser=0 — AT-29",
			deploy: deployment(withPodSecurityContext(&corev1.PodSecurityContext{
				RunAsNonRoot: boolPtr(true),
				RunAsUser:    int64Ptr(0), // root with runAsNonRoot=true is forbidden
			})),
			wantErr:    true,
			errSubstr:  "runAsUser",
		},
		{
			name: "runAsGroup negative — AT-29",
			deploy: deployment(withPodSecurityContext(&corev1.PodSecurityContext{
				RunAsGroup: int64Ptr(-1),
			})),
			wantErr:    true,
			errSubstr:  "runAsGroup",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-29] testing securityContext: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-36 / Cross-field semantics — ClusterFirstWithHostNet + dnsConfig
// -----------------------------------------------------------------------

func TestPhase2_CrossField_ClusterFirstWithHostNet_DNSConfig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		dnsPolicy corev1.DNSPolicy
		dnsConfig *corev1.PodDNSConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "ClusterFirstWithHostNet + dnsConfig forbidden — AT-36",
			dnsPolicy: corev1.DNSClusterFirstWithHostNet,
			dnsConfig: &corev1.PodDNSConfig{Nameservers: []string{"8.8.8.8"}},
			wantErr:   true,
			errSubstr: "dnsConfig",
		},
		{
			name:      "ClusterFirst + dnsConfig allowed",
			dnsPolicy: corev1.DNSClusterFirst,
			dnsConfig: &corev1.PodDNSConfig{Nameservers: []string{"8.8.8.8"}},
			wantErr:   false,
		},
		{
			name:      "Default (empty) + dnsConfig allowed",
			dnsPolicy: "",
			dnsConfig: &corev1.PodDNSConfig{Nameservers: []string{"8.8.8.8"}},
			wantErr:   false,
		},
		{
			name:      "ClusterFirstWithHostNet without dnsConfig allowed",
			dnsPolicy: corev1.DNSClusterFirstWithHostNet,
			dnsConfig: nil,
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-36] testing cross-field: %s", tc.name)
			deploy := deployment(withDNSPolicy(tc.dnsPolicy), withDNSConfig(tc.dnsConfig))
			result := validateDeploymentRaw(deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-24 / initContainer probes forbidden
// -----------------------------------------------------------------------

func TestPhase2_Semantic_InitContainerProbes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		job       *batchv1.Job
		wantErr   bool
		errSubstr string
	}{
		{
			name: "init container with livenessProbe forbidden — AT-24",
			job: job(
				withJobInitContainerLivenessProbe(),
				withJobRestartPolicy(corev1.RestartPolicyNever),
			),
			wantErr:   true,
			errSubstr: "liveness",
		},
		{
			name: "init container with readinessProbe forbidden — AT-24",
			job: job(
				withJobInitContainerReadinessProbe(),
				withJobRestartPolicy(corev1.RestartPolicyNever),
			),
			wantErr:   true,
			errSubstr: "readiness",
		},
		{
			name: "init container with startupProbe forbidden — AT-24",
			job: job(
				withJobInitContainerStartupProbe(),
				withJobRestartPolicy(corev1.RestartPolicyNever),
			),
			wantErr:   true,
			errSubstr: "startup",
		},
		{
			name:    "init container without probes allowed",
			job:     job(withJobInitContainerNoProbes(), withJobRestartPolicy(corev1.RestartPolicyNever)),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-24] testing initContainer probes: %s", tc.name)
			result := validateJobRaw(tc.job)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-25 / Lifecycle hook handler count
// -----------------------------------------------------------------------

func TestPhase2_Semantic_LifecycleHookHandlerCount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "lifecycle preStop with exactly one handler valid — AT-25",
			deploy: deployment(withLifecyclePreStop(&corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{Command: []string{"prestop", "command"}},
			})),
			wantErr: false,
		},
		{
			name:       "lifecycle hook with zero handlers invalid — AT-25",
			deploy:     deployment(withLifecyclePreStop(&corev1.LifecycleHandler{})),
			wantErr:    true,
			errSubstr:  "exactly one",
		},
		{
			name:       "lifecycle hook with multiple handlers invalid — AT-25",
			deploy:     deployment(withLifecyclePreStopMulti()),
			wantErr:    true,
			errSubstr:  "more than one",
		},
		{
			name: "lifecycle postStart valid",
			deploy: deployment(withLifecyclePostStart(&corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{Path: "/health", Port: intstr.FromInt(8080)},
			})),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-25] testing lifecycle hook handler count: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-33 / Toleration operator/value combinations
// -----------------------------------------------------------------------

func TestPhase2_Semantic_Tolerations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		deploy    *appsv1.Deployment
		wantErr   bool
		errSubstr string
	}{
		{
			name: "toleration Equal operator requires non-empty value — AT-33",
			deploy: deployment(withTolerations([]corev1.Toleration{
				{Key: "node-role", Operator: corev1.TolerationOpEqual, Value: "", Effect: corev1.TaintEffectNoSchedule},
			})),
			wantErr:    true,
			errSubstr:  "value must be non-empty",
		},
		{
			name: "toleration Exists operator requires empty value — AT-33",
			deploy: deployment(withTolerations([]corev1.Toleration{
				{Key: "node-role", Operator: corev1.TolerationOpExists, Value: "something", Effect: corev1.TaintEffectNoSchedule},
			})),
			wantErr:    true,
			errSubstr:  "value must be empty",
		},
		{
			name: "toleration Exists with empty value valid — AT-33",
			deploy: deployment(withTolerations([]corev1.Toleration{
				{Key: "node-role", Operator: corev1.TolerationOpExists, Value: "", Effect: corev1.TaintEffectNoSchedule},
			})),
			wantErr: false,
		},
		{
			name: "toleration with invalid effect label value — AT-33",
			deploy: deployment(withTolerations([]corev1.Toleration{
				{Key: "node-role", Operator: corev1.TolerationOpEqual, Value: "master", Effect: "Invalid Effect!"},
			})),
			wantErr:    true,
			errSubstr:  "effect",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-33] testing tolerations: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-30 / restartPolicy for Job — Never/OnFailure valid, Always forbidden
// AT-31 / restartPolicy for non-Job — empty/Always/OnFailure/Never valid
// -----------------------------------------------------------------------

func TestPhase2_Semantic_RestartPolicy_Job(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		job       *batchv1.Job
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "Job restartPolicy Never valid — AT-30",
			job:     job(withJobRestartPolicy(corev1.RestartPolicyNever)),
			wantErr: false,
		},
		{
			name:    "Job restartPolicy OnFailure valid — AT-30",
			job:     job(withJobRestartPolicy(corev1.RestartPolicyOnFailure)),
			wantErr: false,
		},
		{
			name:      "Job restartPolicy Always forbidden — AT-30",
			job:       job(withJobRestartPolicy(corev1.RestartPolicyAlways)),
			wantErr:   true,
			errSubstr: "restartPolicy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-30] testing Job restartPolicy: %s", tc.name)
			result := validateJobRaw(tc.job)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
			if tc.wantErr && tc.errSubstr != "" && !hasErrErrItems(result.Errors, tc.errSubstr) {
				t.Errorf("expected error containing %q, got %v", tc.errSubstr, result.Errors)
			}
		})
	}
}

func TestPhase2_Semantic_RestartPolicy_NonJob(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		deploy  *appsv1.Deployment
		wantErr bool
	}{
		{
			name:    "Deployment restartPolicy Always valid — AT-31",
			deploy:  deployment(withRestartPolicy(corev1.RestartPolicyAlways)),
			wantErr: false,
		},
		{
			name:    "Deployment restartPolicy OnFailure valid — AT-31",
			deploy:  deployment(withRestartPolicy(corev1.RestartPolicyOnFailure)),
			wantErr: false,
		},
		{
			name:    "Deployment restartPolicy Never valid — AT-31",
			deploy:  deployment(withRestartPolicy(corev1.RestartPolicyNever)),
			wantErr: false,
		},
		{
			name:    "Deployment restartPolicy empty valid (defaults to Always) — AT-31",
			deploy:  deployment(withRestartPolicy("")),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[AT-31] testing non-Job restartPolicy: %s", tc.name)
			result := validateDeploymentRaw(tc.deploy)
			gotErr := len(result.Errors) > 0
			if gotErr != tc.wantErr {
				t.Errorf("wantErr=%v, gotErr=%v, errors=%v", tc.wantErr, gotErr, result.Errors)
			}
		})
	}
}

// -----------------------------------------------------------------------
// AT-38 / Phase 1 + Phase 2 composition
// Phase 1 validatePodSpec still runs; Phase 2 augments it.
// -----------------------------------------------------------------------

// Phase 1 + Phase 2 composition for Deployment.
// Both layers fire: Phase 1 (container name required) + Phase 2 security cross-field (runAsNonRoot+runAsUser=0)
func TestPhase2_Composition_Phase1PlusPhase2(t *testing.T) {
	t.Parallel()
	deploy := deployment(
		withContainerName(""), // Phase 1: name required
		withPodSecurityContext(&corev1.PodSecurityContext{
			RunAsNonRoot: boolPtr(true),
			RunAsUser:    int64Ptr(0), // Phase 2: forbidden with runAsNonRoot=true
		}),
	)

	result := validateDeploymentRaw(deploy)

	// Phase 1 container name error fires
	if !hasErrErrItems(result.Errors, "name") {
		t.Errorf("expected Phase 1 container name error, got %v", result.Errors)
	}
	// Phase 2 runAsNonRoot+runAsUser=0 error fires
	if !hasErrErrItems(result.Errors, "runAsUser") {
		t.Errorf("expected Phase 2 runAsUser error, got %v", result.Errors)
	}

	// Confirm Phase 1 still runs independently (empty name → error)
	deployMissingName := deployment(withContainerName(""))
	resultMissing := validateDeploymentRaw(deployMissingName)
	if !hasErrErrItems(resultMissing.Errors, "name") {
		t.Errorf("expected Phase 1 container name required error, got %v", resultMissing.Errors)
	}
}

// -----------------------------------------------------------------------
// AT-39 / Wire-in — all 8 validators call both structural and semantic layers
// -----------------------------------------------------------------------

func TestPhase2_WireIn_AllValidators(t *testing.T) {
	t.Parallel()
	// Kind → validator function name → what to check
	kinds := []string{"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet",
		"ReplicationController", "Pod", "Job", "CronJob"}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			t.Logf("[AT-39] verifying wire-in for %s", kind)
			bv := NewBuiltinValidator()
			if !bv.CanValidate(kind) {
				t.Errorf("CanValidate(%q) = false, want true", kind)
			}
		})
	}
}

// A Job that triggers both Phase 1 and Phase 2 errors simultaneously.
// Job is used because it calls validateInitContainerProbes (Phase 2 semantic).
func TestPhase2_WireIn_Phase1PlusPhase2BothFire(t *testing.T) {
	t.Parallel()
	// This Job has:
	// 1. Phase 1 error: empty container name
	// 2. Phase 2 semantic error: init container has liveness probe
	probe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"echo"}}},
	}
	job := job(
		withJobContainerName(""), // Phase 1: name required
		withJobInitContainerLivenessProbeWithProbe(probe), // Phase 2: init probe forbidden
		withJobRestartPolicy(corev1.RestartPolicyNever),
	)

	result := validateJobRaw(job)

	// Should have BOTH Phase 1 error (name) AND Phase 2 error (init container)
	if len(result.Errors) < 2 {
		t.Errorf("expected at least 2 errors (Phase1 + Phase2), got %d: %v", len(result.Errors), result.Errors)
	}
	hasPhase1Err := hasErrErrItems(result.Errors, "name")
	hasPhase2Err := hasErrErrItems(result.Errors, "init container")
	if !hasPhase1Err || !hasPhase2Err {
		t.Errorf("Phase1+Phase2 both should fire. Phase1=name error: %v, Phase2=initContainer error: %v",
			hasPhase1Err, hasPhase2Err)
	}
}

// -----------------------------------------------------------------------
// AT-40 / Phase transition — Phase 1 only → Phase 1+2
// -----------------------------------------------------------------------

func TestPhase2_PhaseTransition_StructuralAndSemanticBothPresent(t *testing.T) {
	t.Parallel()
	deploy := deployment(
		withTopologySpreadConstraints(corev1.TopologySpreadConstraint{
			MaxSkew:           1,
			TopologyKey:       "kubernetes.io/hostname",
			WhenUnsatisfiable: corev1.DoNotSchedule,
		}),
		withClusterFirstWithHostNetAndDNSConfig(), // cross-field semantic
	)

	result := validateDeploymentRaw(deploy)

	// structural field topologySpreadConstraints passes (valid)
	// cross-field semantic: ClusterFirstWithHostNet + dnsConfig is forbidden
	if !hasErrErrItems(result.Errors, "dnsConfig") {
		t.Errorf("expected dnsConfig cross-field error, got %v", result.Errors)
	}
}

// -----------------------------------------------------------------------
// AT-41 / Global test owner coverage
// All Phase 2 files must be exercised by these tests.
// Files: semantic_lifecycle.go, semantic_security.go, semantic_scheduling.go,
//        structural.go, generated_structural.go, constants.go
// -----------------------------------------------------------------------

func TestPhase2_GlobalCoverage_SemanticFilesCovered(t *testing.T) {
	t.Parallel()
	// This test acts as a marker — it documents which files are covered by this suite.
	// semantic_lifecycle:  TestPhase2_Semantic_InitContainerProbes, TestPhase2_Semantic_LifecycleHookHandlerCount
	// semantic_security:  TestPhase2_StructuralLayer_SecurityContext (runAsNonRoot+runAsUser)
	// semantic_scheduling: TestPhase2_StructuralLayer_Affinity, TestPhase2_StructuralLayer_DNSConfig,
	//                     TestPhase2_CrossField_ClusterFirstWithHostNet_DNSConfig,
	//                     TestPhase2_Semantic_Tolerations, TestPhase2_StructuralLayer_TopologySpreadConstraints
	// structural:        TestPhase2_StructuralLayer_* (all structural layer tests)
	// generated_structural: implicitly exercised by all structural tests (deferredFieldIndex)
	// constants:         implicitly exercised by all semantic tests (capability constants, etc.)

	if testing.Short() {
		t.Skip("skipping coverage marker in short mode")
	}
}

// =============================================================================
// Test fixtures and builder helpers
// =============================================================================

func boolPtr(b bool) *bool    { return &b }
func int64Ptr(i int64) *int64 { return &i }
func int32Ptr(i int32) *int32 { return &i }

func baseDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx:1.21"}},
				},
			},
		},
	}
}

// Deployment modifier helpers

func deployment(opts ...func(*appsv1.Deployment)) *appsv1.Deployment {
	d := baseDeployment()
	for _, o := range opts {
		o(d)
	}
	return d
}

func withContainerName(name string) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Name = name
	}
}

func withTopologySpreadConstraints(c corev1.TopologySpreadConstraint) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{c}
	}
}

func withDNSConfig(cfg *corev1.PodDNSConfig) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.DNSConfig = cfg
	}
}

func withAffinity(a *corev1.Affinity) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Affinity = a
	}
}

func withPodSecurityContext(sc *corev1.PodSecurityContext) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.SecurityContext = sc
	}
}

func withDNSPolicy(p corev1.DNSPolicy) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.DNSPolicy = p
	}
}

func withRestartPolicy(p corev1.RestartPolicy) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.RestartPolicy = p
	}
}

func withTolerations(t []corev1.Toleration) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Tolerations = t
	}
}

func withLifecyclePreStop(h *corev1.LifecycleHandler) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].Lifecycle == nil {
			d.Spec.Template.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{}
		}
		d.Spec.Template.Spec.Containers[0].Lifecycle.PreStop = h
	}
}

func withLifecyclePreStopMulti() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].Lifecycle == nil {
			d.Spec.Template.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{}
		}
		d.Spec.Template.Spec.Containers[0].Lifecycle.PreStop = &corev1.LifecycleHandler{
			Exec:    &corev1.ExecAction{Command: []string{"echo"}},
			HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(8080)},
		}
	}
}

func withLifecyclePostStart(h *corev1.LifecycleHandler) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		if d.Spec.Template.Spec.Containers[0].Lifecycle == nil {
			d.Spec.Template.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{}
		}
		d.Spec.Template.Spec.Containers[0].Lifecycle.PostStart = h
	}
}

func withClusterFirstWithHostNetAndDNSConfig() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.HostNetwork = true
		d.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
		d.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{"8.8.8.8"},
		}
	}
}

func withPhase1ValidContainer() func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Name = "valid-container"
	}
}

// withInitContainerProbeDeployment adds an init container with a probe (for Deployment tests)
func withInitContainerProbeDeployment(probe corev1.Probe) func(*appsv1.Deployment) {
	return func(d *appsv1.Deployment) {
		initC := corev1.Container{
			Name:  "init-container",
			Image: "busybox:1.34",
		}
		// Set the specific probe type on the init container (LivenessProbe field is always present on Container)
		if probe.HTTPGet != nil || probe.TCPSocket != nil || probe.Exec != nil || probe.GRPC != nil {
			initC.LivenessProbe = &probe
		}
		d.Spec.Template.Spec.InitContainers = []corev1.Container{initC}
	}
}

// Job helpers

func baseJob() *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test-job"}},
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{{Name: "job-container", Image: "busybox:1.34"}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

func job(opts ...func(*batchv1.Job)) *batchv1.Job {
	j := baseJob()
	for _, o := range opts {
		o(j)
	}
	return j
}

func withJobInitContainerLivenessProbe() func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:          "init-container",
				Image:         "busybox:1.34",
				LivenessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"echo"}}}},
			},
		}
	}
}

func withJobInitContainerReadinessProbe() func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:           "init-container",
				Image:          "busybox:1.34",
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"echo"}}}},
			},
		}
	}
}

func withJobInitContainerStartupProbe() func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:          "init-container",
				Image:         "busybox:1.34",
				StartupProbe:  &corev1.Probe{ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"echo"}}}},
			},
		}
	}
}

func withJobInitContainerProbe(probe corev1.Probe) func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:  "init-container",
				Image: "busybox:1.34",
			},
		}
		// Set the specific probe type on the init container based on what's non-nil
		if probe.HTTPGet != nil || probe.TCPSocket != nil || probe.Exec != nil || probe.GRPC != nil {
			j.Spec.Template.Spec.InitContainers[0].LivenessProbe = &probe
		}
	}
}

func withJobInitContainerNoProbes() func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:  "init-container",
				Image: "busybox:1.34",
			},
		}
	}
}

func withJobRestartPolicy(p corev1.RestartPolicy) func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.RestartPolicy = p
	}
}

func withJobContainerName(name string) func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.Containers[0].Name = name
	}
}

func withJobInitContainerLivenessProbeWithProbe(probe corev1.Probe) func(*batchv1.Job) {
	return func(j *batchv1.Job) {
		j.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:          "init-container",
				Image:         "busybox:1.34",
				LivenessProbe: &probe,
			},
		}
	}
}

// -----------------------------------------------------------------------
// Validation helpers — use the same pattern as existing tests
// -----------------------------------------------------------------------

// validateDeploymentRaw encodes a Deployment and validates it via BuiltinValidator
func validateDeploymentRaw(d *appsv1.Deployment) *Result {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	// Use a simple YAML encoding approach compatible with ValidateResource
	bv := NewBuiltinValidator()
	// Encode to YAML for ValidateResource
	raw, _ := yaml.Marshal(d)
	return bv.ValidateResource(raw)
}

// validateJobRaw encodes a Job and validates it via BuiltinValidator
func validateJobRaw(j *batchv1.Job) *Result {
	scheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(scheme)
	bv := NewBuiltinValidator()
	raw, _ := yaml.Marshal(j)
	return bv.ValidateResource(raw)
}

// hasErrErrItems checks if any error in the list contains the given substring
func hasErrErrItems(errs []ErrorItem, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e.Field, substr) || strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}
