// Package config provides configuration management for the GeoChrono application.
//
// @title Configuration Management Package
// @version 1.0
// @description Handles YAML configuration loading, parsing, and validation
// @description Manages GPS processing settings and map generation options
// @description Supports environment variable substitution and validation
//
// Features:
// - YAML configuration file parsing
// - Environment variable substitution
// - Comprehensive validation
// - Default value application
// - Structured configuration hierarchy
package config

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v2"
)

// Config represents the complete configuration structure for GeoChrono.
//
// @struct Config
// @description Root configuration structure for entire application
// @description Encompases Google Maps, input/output, styling, and processing settings
// @property GoogleMaps GoogleMapsConfig Google Maps API integration settings
// @property Input InputConfig Input file and parsing configuration
// @property Output OutputConfig Output file generation settings
// @property Map MapConfig Map display and styling options
// @property Markers MarkersConfig GPS point marker customization
// @property Path PathConfig Path/trail visualization settings
// @property InfoWindows InfoWindowsConfig Popup window configuration
// @property Processing ProcessingConfig Data processing and filtering options
// @property Logging LoggingConfig Debug and logging settings
type Config struct {
	GoogleMaps  GoogleMapsConfig  `yaml:"google_maps"`  // @field GoogleMaps Google Maps API configuration
	Input       InputConfig       `yaml:"input"`        // @field Input Input file and format settings
	Output      OutputConfig      `yaml:"output"`       // @field Output Output file configuration
	Map         MapConfig         `yaml:"map"`          // @field Map Map display and styling options
	Markers     MarkersConfig     `yaml:"markers"`      // @field Markers GPS point marker configuration
	Path        PathConfig        `yaml:"path"`         // @field Path Path/trail visualization settings
	InfoWindows InfoWindowsConfig `yaml:"info_windows"` // @field InfoWindows Popup window configuration
	Processing  ProcessingConfig  `yaml:"processing"`   // @field Processing Data processing options
	Logging     LoggingConfig     `yaml:"logging"`      // @field Logging Logging and debug settings
}

// GoogleMapsConfig holds Google Maps API configuration settings.
// These settings control how the application integrates with Google Maps services.
type GoogleMapsConfig struct {
	APIKey     string   `yaml:"api_key"`     // Google Maps API key (supports env var substitution)
	APIVersion string   `yaml:"api_version"` // Google Maps API version to use
	Libraries  []string `yaml:"libraries"`   // Additional Google Maps libraries to load
}

// InputConfig holds input file configuration and parsing settings.
// This defines where to find GPS data and how to interpret it.
type InputConfig struct {
	CSVFile   string          `yaml:"csv_file"`   // Path to the input CSV file
	CSVFormat CSVFormatConfig `yaml:"csv_format"` // CSV parsing configuration
}

// CSVFormatConfig holds CSV file parsing configuration.
// This allows flexible parsing of various CSV formats and column layouts.
type CSVFormatConfig struct {
	TimestampColumn   string `yaml:"timestamp_column"`   // Name of timestamp column
	LatitudeColumn    string `yaml:"latitude_column"`    // Name of latitude column
	LongitudeColumn   string `yaml:"longitude_column"`   // Name of longitude column
	TitleColumn       string `yaml:"title_column"`       // Name of title/name column (optional)
	DescriptionColumn string `yaml:"description_column"` // Name of description column (optional)
	CategoryColumn    string `yaml:"category_column"`    // Name of category column (optional)
	HasHeader         bool   `yaml:"has_header"`         // Whether CSV file has a header row
	Delimiter         string `yaml:"delimiter"`          // Field delimiter (default: comma)
	SkipRows          int    `yaml:"skip_rows"`          // Number of rows to skip at beginning
}

// OutputConfig holds output file configuration and export options.
// This controls where and how the generated map and related files are saved.
type OutputConfig struct {
	HTMLFile  string `yaml:"html_file"`  // Path to output HTML file
	Debug     bool   `yaml:"debug"`      // Enable debug output in generated files
	ExportKML bool   `yaml:"export_kml"` // Whether to export KML file
	KMLFile   string `yaml:"kml_file"`   // Path to output KML file (if enabled)
}

// MapConfig holds map display and presentation configuration.
// This controls the overall appearance and behavior of the generated map.
type MapConfig struct {
	Title         string            `yaml:"title"`           // Map title displayed in browser
	Width         string            `yaml:"width"`           // Map width (CSS units)
	Height        string            `yaml:"height"`          // Map height (CSS units)
	InitialView   InitialViewConfig `yaml:"initial_view"`    // Initial map view settings
	AutoFitBounds bool              `yaml:"auto_fit_bounds"` // Auto-fit map to GPS points
	Controls      ControlsConfig    `yaml:"controls"`        // Map control visibility
}

// InitialViewConfig holds initial map view and positioning settings.
// This determines how the map appears when first loaded.
type InitialViewConfig struct {
	Center  CenterConfig `yaml:"center"`   // Initial center coordinates
	Zoom    *int         `yaml:"zoom"`     // Initial zoom level (1-20)
	MapType string       `yaml:"map_type"` // Map type (roadmap, satellite, hybrid, terrain)
}

// CenterConfig holds geographical center coordinates for map positioning.
// If not specified, the center will be calculated from GPS points.
type CenterConfig struct {
	Latitude  *float64 `yaml:"latitude"`  // Center latitude (-90 to 90)
	Longitude *float64 `yaml:"longitude"` // Center longitude (-180 to 180)
}

// ControlsConfig holds configuration for map control visibility.
// This allows customization of which map controls are displayed to users.
type ControlsConfig struct {
	ZoomControl       bool `yaml:"zoom_control"`        // Show zoom in/out buttons
	StreetViewControl bool `yaml:"street_view_control"` // Show street view control
	FullscreenControl bool `yaml:"fullscreen_control"`  // Show fullscreen button
	MapTypeControl    bool `yaml:"map_type_control"`    // Show map type selector
	ScaleControl      bool `yaml:"scale_control"`       // Show map scale indicator
}

// MarkersConfig holds configuration for GPS point markers on the map.
// This allows customization of how different types of points are displayed.
type MarkersConfig struct {
	Default    MarkerStyleConfig `yaml:"default"`    // Default marker style for regular points
	Start      MarkerStyleConfig `yaml:"start"`      // Special style for start point
	End        MarkerStyleConfig `yaml:"end"`        // Special style for end point
	Categories map[string]string `yaml:"categories"` // Category-specific marker colors/styles
}

// MarkerStyleConfig holds styling configuration for individual markers.
// This defines the visual appearance of GPS points on the map.
type MarkerStyleConfig struct {
	Icon  IconConfig  `yaml:"icon"`  // Icon appearance settings
	Label LabelConfig `yaml:"label"` // Label display settings
}

// IconConfig holds configuration for marker icons.
// This controls the visual representation of GPS points on the map.
type IconConfig struct {
	Color  string      `yaml:"color"`  // Icon color (hex color code)
	URL    *string     `yaml:"url"`    // Custom icon URL (optional)
	Size   SizeConfig  `yaml:"size"`   // Icon dimensions
	Anchor PointConfig `yaml:"anchor"` // Icon anchor point
}

// SizeConfig holds width and height dimensions.
// Used for defining the size of icons and other visual elements.
type SizeConfig struct {
	Width  int `yaml:"width"`  // Width in pixels
	Height int `yaml:"height"` // Height in pixels
}

// PointConfig holds X/Y coordinate configuration.
// Used for positioning elements like icon anchors.
type PointConfig struct {
	X int `yaml:"x"` // X coordinate offset
	Y int `yaml:"y"` // Y coordinate offset
}

// LabelConfig holds configuration for marker labels.
// This controls text displayed on or near GPS point markers.
type LabelConfig struct {
	ShowSequence bool       `yaml:"show_sequence"` // Show point sequence numbers
	Text         *string    `yaml:"text"`          // Custom label text template
	Color        string     `yaml:"color"`         // Label text color
	Font         FontConfig `yaml:"font"`          // Font styling options
}

// FontConfig holds font styling configuration.
// This defines the typography for labels and text elements.
type FontConfig struct {
	Family string `yaml:"family"` // Font family name
	Size   string `yaml:"size"`   // Font size (CSS units)
	Weight string `yaml:"weight"` // Font weight (normal, bold, etc.)
}

// PathConfig holds configuration for the GPS trail/path visualization.
// This controls how the chronological path between GPS points is displayed.
type PathConfig struct {
	Enabled   bool            `yaml:"enabled"`   // Whether to show connecting path
	Style     PathStyleConfig `yaml:"style"`     // Path visual styling
	Animation AnimationConfig `yaml:"animation"` // Path animation settings
}

// PathStyleConfig holds visual styling for the GPS path.
// This defines the appearance of the line connecting GPS points.
type PathStyleConfig struct {
	Color         string  `yaml:"color"`          // Path line color (hex code)
	Opacity       float64 `yaml:"opacity"`        // Path transparency (0.0-1.0)
	Weight        int     `yaml:"weight"`         // Path line thickness in pixels
	StrokePattern string  `yaml:"stroke_pattern"` // Line pattern (solid, dashed, etc.)
}

// AnimationConfig holds configuration for path animation effects.
// This controls how the GPS trail is animated to show movement over time.
type AnimationConfig struct {
	Enabled             bool `yaml:"enabled"`               // Enable path animation
	Speed               int  `yaml:"speed"`                 // Animation speed (1-10)
	ShowDirectionArrows bool `yaml:"show_direction_arrows"` // Show movement direction
}

// InfoWindowsConfig holds configuration for GPS point popup windows.
// This controls the information displayed when users click on GPS points.
type InfoWindowsConfig struct {
	Enabled       bool   `yaml:"enabled"`         // Enable popup windows
	Template      string `yaml:"template"`        // HTML template for popup content
	AutoOpenStart bool   `yaml:"auto_open_start"` // Auto-open popup for first point
	MaxWidth      int    `yaml:"max_width"`       // Maximum popup width in pixels
}

// ProcessingConfig holds configuration for GPS data processing and filtering.
// This controls how raw GPS data is cleaned and prepared for visualization.
type ProcessingConfig struct {
	RemoveDuplicates  bool     `yaml:"remove_duplicates"`   // Remove duplicate GPS points
	MinDistanceFilter float64  `yaml:"min_distance_filter"` // Minimum distance between points (meters)
	SmoothPath        bool     `yaml:"smooth_path"`         // Apply path smoothing algorithms
	MaxSpeedFilter    float64  `yaml:"max_speed_filter"`    // Maximum realistic speed (km/h)
	Timezone          string   `yaml:"timezone"`            // Timezone for timestamp processing
	TimestampFormats  []string `yaml:"timestamp_formats"`   // Supported timestamp formats
}

// LoggingConfig holds configuration for application logging and debugging.
// This controls how the application reports its operations and any issues.
type LoggingConfig struct {
	Level   string `yaml:"level"`   // Log level (debug, info, warn, error)
	File    string `yaml:"file"`    // Log file path (empty for stdout)
	Verbose bool   `yaml:"verbose"` // Enable verbose output
}

// Load reads and parses a YAML configuration file from the specified path.
//
// @function Load
// @description Loads and parses YAML configuration file into Config struct
// @param filename string Path to YAML configuration file
// @return *Config Parsed configuration structure with all settings
// @return error Error if file cannot be read or YAML is invalid
// @throws FileNotFoundError When config file doesn't exist
// @throws ParseError When YAML syntax is invalid
// @example config, err := Load("config.yaml")
func Load(filename string) (*Config, error) {
	// Open the configuration file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open config file %s: %w", filename, err)
	}
	defer file.Close()

	// Parse YAML content into Config struct
	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	return &config, nil
}

// ResolveAPIKey resolves the Google Maps API key from environment variables
// if the configuration uses environment variable placeholders.
//
// @method ResolveAPIKey
// @description Resolves Google Maps API key from environment variables
// @return error Error if API key is missing or environment variable not set
// @syntax Supports ${VARIABLE_NAME} environment variable substitution
// @example API key "${GOOGLE_MAPS_KEY}" resolves to env var value
func (c *Config) ResolveAPIKey() error {
	if c.GoogleMaps.APIKey == "" {
		return fmt.Errorf("google Maps API key is required (use 'DEMO' for demonstration)")
	}

	// Allow "DEMO" as a special demo API key
	if c.GoogleMaps.APIKey == "DEMO" {
		return nil
	}

	// Check for environment variable substitution syntax: ${VAR_NAME}
	if strings.HasPrefix(c.GoogleMaps.APIKey, "${") && strings.HasSuffix(c.GoogleMaps.APIKey, "}") {
		// Extract environment variable name
		envVar := strings.TrimSuffix(strings.TrimPrefix(c.GoogleMaps.APIKey, "${"), "}")

		// Replace with actual environment variable value
		if envValue := os.Getenv(envVar); envValue != "" {
			c.GoogleMaps.APIKey = envValue
		} else {
			return fmt.Errorf("environment variable %s is not set", envVar)
		}
	}

	return nil
}

// Validate performs comprehensive validation on the configuration to ensure
// all required fields are present and have valid values.
// It checks for missing API keys, file paths, and other critical settings.
func (c *Config) Validate() error {
	// Validate Google Maps API key (allow "DEMO" for demonstration purposes)
	if c.GoogleMaps.APIKey == "" {
		return fmt.Errorf("google Maps API key is required (use 'DEMO' for demonstration)")
	}

	// Validate input file path
	if c.Input.CSVFile == "" {
		return fmt.Errorf("input CSV file is required")
	}

	// Validate output file path
	if c.Output.HTMLFile == "" {
		return fmt.Errorf("output HTML file is required")
	}

	// All validation checks passed
	return nil
}
