package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"k8-manifest-validator/pkg/types"
)

// Options configures the Validator library
type Options struct {
	CRDs                 []byte // CRD YAML content
	IgnoreMissingSchemas bool   // Skip unknown kinds
	Workers              int    // Parallel workers (default 8)
}

// Validator is the high-level validation API
type Validator struct {
	engine *Engine
}

// New creates a new Validator with the given options
func New(opts Options) *Validator {
	workers := opts.Workers
	if workers <= 0 {
		workers = 8
	}

	engine := NewEngine(EngineOptions{
		IgnoreMissing: opts.IgnoreMissingSchemas,
		Workers:       workers,
	})

	v := &Validator{engine: engine}

	// Register CRDs if provided
	if len(opts.CRDs) > 0 {
		if err := v.engine.cr.RegisterCRD(opts.CRDs); err != nil {
			// Log but continue - may be valid CRD that will be validated later
			fmt.Fprintf(os.Stderr, "Warning: failed to pre-register CRDs: %v\n", err)
		}
	}

	return v
}

// Validate validates Kubernetes manifests from raw YAML bytes and returns results.
// It writes manifests to a temp directory for the engine to process.
func (v *Validator) Validate(manifests []byte) ([]types.Result, error) {
	// If manifests is empty, return empty results
	if len(manifests) == 0 {
		return []types.Result{}, nil
	}

	// Create a temporary directory for the manifests
	tmpDir, err := os.MkdirTemp("", "k8s-validator-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write manifests to temp file
	manifestPath := filepath.Join(tmpDir, "manifests.yaml")
	if err := os.WriteFile(manifestPath, manifests, 0644); err != nil {
		return nil, fmt.Errorf("failed to write manifests: %w", err)
	}

	// Use engine to validate (no CRD paths since CRDs were registered at init)
	results, err := v.engine.Validate(manifestPath, nil)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return results.Resources, nil
}

// ValidateWithCRDs validates Kubernetes manifests with additional CRDs provided.
// This allows passing CRDs at validation time rather than at construction time.
func (v *Validator) ValidateWithCRDs(manifests []byte, crds []byte) ([]types.Result, error) {
	// If manifests is empty, return empty results
	if len(manifests) == 0 {
		return []types.Result{}, nil
	}

	// Create a temporary directory for the manifests
	tmpDir, err := os.MkdirTemp("", "k8s-validator-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write manifests to temp file
	manifestPath := filepath.Join(tmpDir, "manifests.yaml")
	if err := os.WriteFile(manifestPath, manifests, 0644); err != nil {
		return nil, fmt.Errorf("failed to write manifests: %w", err)
	}

	// Write CRDs to temp files if provided
	var crdPaths []string
	if len(crds) > 0 {
		crdPath := filepath.Join(tmpDir, "crds.yaml")
		if err := os.WriteFile(crdPath, crds, 0644); err != nil {
			return nil, fmt.Errorf("failed to write CRDs: %w", err)
		}
		crdPaths = append(crdPaths, crdPath)
	}

	// Use engine to validate with CRD paths
	results, err := v.engine.Validate(manifestPath, crdPaths)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return results.Resources, nil
}
