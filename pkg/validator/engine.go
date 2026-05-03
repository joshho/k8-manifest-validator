package validator

import (
	"fmt"
	"os"
	"sync"

	"k8-manifest-validator/pkg/types"
)

// Engine orchestrates all validators to validate Kubernetes manifests
type Engine struct {
	scanner        *Scanner
	decoder        *Decoder
	builtin       *BuiltinValidator
	crd           *CRDValidator
	cr            *CRValidator
	ignoreMissing bool
	workers       int
}

// EngineOptions configures the Engine
type EngineOptions struct {
	// IgnoreMissing determines how unknown kinds are handled:
	// true = skip unknown kinds (StatusSkipped)
	// false = return error for unknown kinds (StatusError)
	IgnoreMissing bool

	// Workers sets the number of concurrent validation workers (default: 1)
	Workers int
}

// NewEngine creates a new Engine with all required validators
func NewEngine(opts EngineOptions) *Engine {
	if opts.Workers <= 0 {
		opts.Workers = 1
	}

	scanner := NewScanner()
	decoder := NewDecoder()
	builtin := NewBuiltinValidator()
	crd := NewCRDValidator()
	cr := NewCRValidator(crd)

	return &Engine{
		scanner:        scanner,
		decoder:        decoder,
		builtin:        builtin,
		crd:            crd,
		cr:             cr,
		ignoreMissing:  opts.IgnoreMissing,
		workers:        opts.Workers,
	}
}

// Validate validates all Kubernetes manifests in the given path
// crdPaths can be provided to register CRDs before validation
func (e *Engine) Validate(path string, crdPaths []string) (*types.Results, error) {
	// Check single-file paths exist before Walk
	if info, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
	} else if !info.IsDir() {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
	}

	// Register CRDs if provided
	for _, crdPath := range crdPaths {
		crdData, err := os.ReadFile(crdPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CRD file %s: %w", crdPath, err)
		}
		if err := e.cr.RegisterCRD(crdData); err != nil {
			return nil, fmt.Errorf("failed to register CRD %s: %w", crdPath, err)
		}
	}

	// Scan for manifest files
	resources, err := e.scanner.Walk(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan path: %w", err)
	}

	// Validate all resources concurrently
	results := make([]types.Result, 0, len(resources))
	resultChan := make(chan types.Result, len(resources))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.workers)

	for _, res := range resources {
		wg.Add(1)
		go func(resource Resource) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := e.validateResource(resource)
			resultChan <- result
		}(res)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	// Build final results with summary
	return &types.Results{
		Summary:   types.CalculateSummary(results),
		Resources: results,
	}, nil
}

// validateResource validates a single resource and returns a types.Result
func (e *Engine) validateResource(resource Resource) types.Result {
	// Decode the manifest
	manifest, err := e.decoder.DecodeManifest(resource.Bytes)
	if err != nil {
		return types.NewResult(
			resource.Path,
			"",
			"",
			"",
			"",
			types.StatusError,
			[]types.ErrorItem{
				{Field: "manifest", Message: err.Error(), Code: types.ErrCodeInvalid},
			},
		)
	}

	// Determine the validator based on kind
	return e.routeToValidator(manifest, resource)
}

// routeToValidator routes the manifest to the appropriate validator
func (e *Engine) routeToValidator(manifest Manifest, resource Resource) types.Result {
	// Check if it's a CRD
	if manifest.Kind == "CustomResourceDefinition" {
		// Validate the CRD itself
		return e.validateCRD(resource.Bytes, manifest, resource.Path)
	}

	// Check if it's a known built-in kind
	if e.builtin.CanValidate(manifest.Kind) {
		return e.validateBuiltin(resource.Bytes, manifest, resource.Path)
	}

	// Check if it's a registered CR (Custom Resource)
	group, _ := parseAPIVersionGroup(manifest.APIVersion)
	if group != "" {
		crResult, err := e.cr.ValidateCR(resource.Bytes)
		if err == nil && crResult != nil {
			// Convert from validator.Result to types.Result
			return types.NewResult(
				resource.Path,
				crResult.Kind,
				crResult.Name,
				crResult.Namespace,
				crResult.APIVersion,
				crResult.Status,
				convertErrors(crResult.Errors),
			)
		}
	}

	// Unknown kind - handle based on ignoreMissing flag
	if e.ignoreMissing {
		return types.NewResult(
			resource.Path,
			manifest.Kind,
			manifest.Metadata.Name,
			manifest.Metadata.Namespace,
			manifest.APIVersion,
			types.StatusSkipped,
			[]types.ErrorItem{
				{Field: "kind", Message: fmt.Sprintf("unknown kind %q - no schema registered", manifest.Kind), Code: types.ErrCodeNotFound},
			},
		)
	}

	return types.NewResult(
		resource.Path,
		manifest.Kind,
		manifest.Metadata.Name,
		manifest.Metadata.Namespace,
		manifest.APIVersion,
		types.StatusError,
		[]types.ErrorItem{
			{Field: "kind", Message: fmt.Sprintf("unknown kind %q - use --ignore-missing-schemas to skip", manifest.Kind), Code: types.ErrCodeNotFound},
		},
	)
}

// validateBuiltin validates a built-in resource and returns a types.Result
func (e *Engine) validateBuiltin(data []byte, manifest Manifest, path string) types.Result {
	result := e.builtin.ValidateResource(data)
	// Convert from validator.Result to types.Result
	return types.NewResult(
		path,
		result.Kind,
		result.Name,
		result.Namespace,
		result.APIVersion,
		result.Status,
		convertErrors(result.Errors),
	)
}

// validateCRD validates a CustomResourceDefinition itself
func (e *Engine) validateCRD(data []byte, manifest Manifest, path string) types.Result {
	_, err := e.crd.ValidateCRD(data)
	if err != nil {
		return types.NewResult(
			path,
			manifest.Kind,
			manifest.Metadata.Name,
			manifest.Metadata.Namespace,
			manifest.APIVersion,
			types.StatusInvalid,
			[]types.ErrorItem{
				{Field: "crd", Message: fmt.Sprintf("CRD validation failed: %v", err), Code: types.ErrCodeInvalid},
			},
		)
	}

	// Register the CRD schema for use with custom resources
	if err := e.cr.RegisterCRD(data); err != nil {
		return types.NewResult(
			path,
			manifest.Kind,
			manifest.Metadata.Name,
			manifest.Metadata.Namespace,
			manifest.APIVersion,
			types.StatusError,
			[]types.ErrorItem{
				{Field: "crd", Message: fmt.Sprintf("failed to register CRD schema: %v", err), Code: types.ErrCodeInvalid},
			},
		)
	}

	return types.NewResult(
		path,
		manifest.Kind,
		manifest.Metadata.Name,
		manifest.Metadata.Namespace,
		manifest.APIVersion,
		types.StatusValid,
		nil,
	)
}

// convertErrors converts internal ErrorItem slice to types.ErrorItem slice
func convertErrors(errs []ErrorItem) []types.ErrorItem {
	if errs == nil {
		return nil
	}
	result := make([]types.ErrorItem, len(errs))
	for i, e := range errs {
		result[i] = types.ErrorItem{
			Field:   e.Field,
			Message: e.Message,
			Code:    e.Code,
		}
	}
	return result
}

// parseAPIVersionGroup splits "group/version" into group and version
func parseAPIVersionGroup(apiVersion string) (group, version string) {
	for i := 0; i < len(apiVersion); i++ {
		if apiVersion[i] == '/' {
			return apiVersion[:i], apiVersion[i+1:]
		}
	}
	return "", apiVersion
}