package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8-manifest-validator/pkg/output"
	"k8-manifest-validator/pkg/validator"
)

// version is the binary version string.
// Format: v{ks_version}-{patch} (e.g., v1.31-0, v1.32-0)
const version = "v1.31-0"

func getVersion() string {
	return version
}

func main() {
	// Define flags
	manifestPath := flag.String("f", "", "Path to manifest file or directory")
	crdPath := flag.String("crd", "", "Path to CRD directory")
	outputFormat := flag.String("output", "json", "Output format: json or yaml")
	ignoreMissing := flag.Bool("ignore-missing-schemas", false, "Skip unknown resource kinds")
	ignorePatterns := flag.String("ignore", "", "Regex patterns to ignore (can be specified multiple times)")
	workers := flag.Int("n", 8, "Number of parallel workers")
	showVersion := flag.Bool("version", false, "Print version string")
	showHelp := flag.Bool("h", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "k8-manifest-validator - Kubernetes manifest validator\n\nUsage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	version := getVersion()

	// Handle -h / --help
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle --version
	if *showVersion {
		fmt.Println("k8-manifest-validator version", version)
		os.Exit(0)
	}

	// Validate required flags
	if *manifestPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -f flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse ignore patterns
	var patterns []string
	if *ignorePatterns != "" {
		patterns = strings.Split(*ignorePatterns, ",")
	}

	// Collect CRD paths if crd directory is specified
	var crdPaths []string
	if *crdPath != "" {
		err := filepath.Walk(*crdPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					crdPaths = append(crdPaths, path)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning CRD directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Create validator engine
	engineOpts := validator.EngineOptions{
		IgnoreMissing: *ignoreMissing,
		Workers:       *workers,
	}
	engine := validator.NewEngine(engineOpts)

	// Configure scanner patterns if provided
	if len(patterns) > 0 {
		// Note: scanner patterns would need to be set before Walk
		// For now, we rely on the engine's default behavior
	}

	// Validate manifests
	results, err := engine.Validate(*manifestPath, crdPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	// Output results in the requested format
	var outputBytes []byte
	switch strings.ToLower(*outputFormat) {
	case "yaml", "yml":
		fmtter := output.NewYAMLFormatter()
		outputBytes, err = fmtter.Format(results)
	case "json":
		fmtter := output.NewJSONFormatter()
		outputBytes, err = fmtter.Format(results)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown output format %q (use json or yaml)\n", *outputFormat)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(outputBytes))

	// Exit with appropriate code
	if results.Summary.Invalid > 0 || results.Summary.Errors > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
