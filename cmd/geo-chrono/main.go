// Package main provides the command-line interface for GeoChrono.
//
// @title GeoChrono CLI Application
// @version 1.0
// @description Command-line tool for generating interactive Google Maps from GPS CSV data
// @description Creates HTML maps with walking trails and chronological GPS visualization
//
// @usage geo-chrono [flags]
// @flags
//
//	-config string    Path to configuration file (default "config.yaml")
//	-csv string       Path to CSV file (overrides config)
//	-apikey string    Google Maps API key (overrides config)
//	-out string       Output HTML file (overrides config)
//	-title string     Map title (overrides config)
//
// @example geo-chrono -csv data.csv -out map.html -title "My Walking Trail"
//
// Features:
// - CSV GPS data processing
// - Interactive Google Maps generation
// - Configurable styling and markers
// - Command-line flag override support
// - YAML configuration file support
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/saratily/geo-chrono/internal/config"
	"github.com/saratily/geo-chrono/internal/csv"
	"github.com/saratily/geo-chrono/internal/gps"
	"github.com/saratily/geo-chrono/internal/mapgen"
)

// main is the entry point for the GeoChrono application.
//
// @function main
// @description Application entry point and orchestration function
// @steps Parse flags, Load config, Override settings, Process CSV, Generate map
// @workflow Configuration → CSV Reading → GPS Processing → Map Generation
// @exit Exits with status code 1 on any error, 0 on success
func main() {
	// Parse command line flags to get user input
	flags := parseFlags()

	// Load configuration from YAML file
	cfg, err := config.Load(flags.ConfigFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Override configuration values with command line flags if provided
	overrideConfigWithFlags(cfg, flags)

	// Resolve Google Maps API key from environment variables if needed
	if err := cfg.ResolveAPIKey(); err != nil {
		log.Fatalf("Error resolving API key: %v", err)
	}

	// Validate that all required configuration values are present
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Create CSV reader with appropriate format configuration
	reader := csv.NewReader(&cfg.Input.CSVFormat, &cfg.Processing)

	// Read and parse GPS points from the CSV file
	points, err := reader.ReadFile(cfg.Input.CSVFile)
	if err != nil {
		log.Fatalf("Error reading CSV file: %v", err)
	}

	// Ensure we have valid GPS data to work with
	if points.IsEmpty() {
		log.Fatal("No valid GPS points found in CSV file")
	}

	// Sort GPS points by timestamp to create chronological path
	points.SortByTimestamp()

	// Log detailed information about loaded GPS points if verbose mode is enabled
	if cfg.Logging.Verbose {
		logPointsInfo(points, cfg.Input.CSVFile)
	}

	// Create map generator and generate interactive HTML map
	generator := mapgen.NewGenerator(cfg)
	if err := generator.Generate(points, cfg.Output.HTMLFile); err != nil {
		log.Fatalf("Error generating map: %v", err)
	}

	// Inform user of successful completion
	fmt.Printf("Map generated successfully: %s\n", cfg.Output.HTMLFile)
	fmt.Printf("Open the file in your browser to view the interactive map\n")
}

// Flags holds command line flag values that can override configuration file settings.
// This allows users to customize behavior without modifying the config file.
type Flags struct {
	ConfigFile string // Path to YAML configuration file
	CSVFile    string // Path to input CSV file with GPS data
	APIKey     string // Google Maps API key for map generation
	Output     string // Path to output HTML file
	Title      string // Title to display on the generated map
}

// parseFlags parses and validates command line arguments.
// Returns a Flags struct containing all parsed values with appropriate defaults.
func parseFlags() *Flags {
	flags := &Flags{}

	// Define command line flags with descriptions and defaults
	flag.StringVar(&flags.ConfigFile, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&flags.CSVFile, "csv", "", "Path to CSV file (overrides config)")
	flag.StringVar(&flags.APIKey, "apikey", "", "Google Maps API key (overrides config)")
	flag.StringVar(&flags.Output, "out", "", "Output HTML file (overrides config)")
	flag.StringVar(&flags.Title, "title", "", "Map title (overrides config)")

	// Parse all provided command line arguments
	flag.Parse()

	return flags
}

// overrideConfigWithFlags applies command line flag values to the configuration,
// allowing users to override specific settings without modifying the config file.
// Command line flags take precedence over configuration file values.
func overrideConfigWithFlags(cfg *config.Config, flags *Flags) {
	// Override input CSV file path if provided
	if flags.CSVFile != "" {
		cfg.Input.CSVFile = flags.CSVFile
	}

	// Override Google Maps API key if provided via flag
	if flags.APIKey != "" {
		cfg.GoogleMaps.APIKey = flags.APIKey
	} else if cfg.GoogleMaps.APIKey == "" || cfg.GoogleMaps.APIKey == "${GOOGLE_MAPS_API_KEY}" {
		// Fallback: try to get API key from environment variable
		if envKey := os.Getenv("GOOGLE_MAPS_API_KEY"); envKey != "" {
			cfg.GoogleMaps.APIKey = envKey
		}
	}

	// Override output HTML file path if provided
	if flags.Output != "" {
		cfg.Output.HTMLFile = flags.Output
	}

	// Override map title if provided
	if flags.Title != "" {
		cfg.Map.Title = flags.Title
	}
}

// logPointsInfo displays detailed information about the loaded GPS points,
// including the total count and time range of the data.
// This helps users understand the scope and coverage of their GPS data.
func logPointsInfo(points gps.Points, filename string) {
	// Get the time span of the GPS data
	start, end := points.TimeRange()

	// Display summary information
	fmt.Printf("Loaded %d GPS points from %s\n", len(points), filename)
	fmt.Printf("Time range: %s to %s\n",
		start.Format("2006-01-02 15:04:05"),
		end.Format("2006-01-02 15:04:05"))
}
