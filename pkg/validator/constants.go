package validator

// ValidCapabilities is the set of valid Linux capability constants as defined
// by Kubernetes and the Linux kernel. These are the only values allowed in
// securityContext.capabilities.add and securityContext.capabilities.drop.
// Ref: k8s.io/api/core/v1/types.go — Capability type
var ValidCapabilities = map[string]bool{
	"AUDIT_CONTROL":     true,
	"AUDIT_READ":        true,
	"AUDIT_WRITE":       true,
	"BLOCK_SUSPEND":     true,
	"CHOWN":             true,
	"DAC_OVERRIDE":      true,
	"DAC_READ_SEARCH":   true,
	"FOWNER":            true,
	"FSETID":            true,
	"IPC_LOCK":          true,
	"IPC_SETOWN":        true,
	"KILL":              true,
	"LEASE":             true,
	"LINUX_IMMUTABLE":   true,
	"NET_BIND_SERVICE":  true,
	"NET_BROADCAST":     true,
	"NET_ADMIN":         true,
	"NET_RAW":           true,
	"OFFSET":            true,
	"OVERLAY_FS":        true,
	"OWNER":             true,
	"READ_IMMUTABLE":    true,
	"REBOOT":            true,
	"RESTART":           true,
	"ROOT":              true,
	"SETGID":            true,
	"SETFCAP":           true,
	"SETPCAP":           true,
	"SETUID":            true,
	"SYS_ADMIN":         true,
	"SYS_BOOT":          true,
	"SYS_CHROOT":        true,
	"SYS_MODULE":        true,
	"SYS_NICE":          true,
	"SYS_PACCT":         true,
	"SYS_PTRACE":        true,
	"SYS_RAWIO":         true,
	"SYS_REBOOT":        true,
	"SYS_RESOURCE":      true,
	"SYS_TIME":          true,
	"SYS_TTY_CONFIG":    true,
	"SYSLOG":            true,
	"WAKE_ALARM":        true,
}

// ValidSeccompProfileTypes is the set of valid seccompProfile.type values.
// Ref: k8s.io/api/core/v1/types.go — SeccompProfile type
var ValidSeccompProfileTypes = map[string]bool{
	"RuntimeDefault": true,
	"Unconfined":     true,
	"Localhost":      true,
}
