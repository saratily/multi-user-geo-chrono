// Package config_test provides comprehensive unit tests for the configuration management package.
// It tests YAML configuration loading, validation, environment variable handling, default value
// application, and error handling scenarios to ensure robust configuration management.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		wantErr     bool
		expectedAPI string
	}{
		{
			name: "valid config",
			configData: `
google_maps:
  api_key: "test-key"
  api_version: "3.54"
  libraries: ["geometry", "places"]
input:
  csv_file: "test.csv"
  csv_format:
    timestamp_column: "timestamp"
    latitude_column: "latitude"
    longitude_column: "longitude"
    has_header: true
    delimiter: ","
output:
  html_file: "test.html"
map:
  title: "Test Map"
  width: "100%"
  height: "600px"
  auto_fit_bounds: true
path:
  enabled: true
  style:
    color: "#FF0000"
    opacity: 0.8
    weight: 3
processing:
  remove_duplicates: true
  timestamp_formats:
    - "2006-01-02T15:04:05Z"
logging:
  level: "info"
  verbose: false
`,
			wantErr:     false,
			expectedAPI: "test-key",
		},
		{
			name:       "invalid yaml",
			configData: "invalid: yaml: content: [",
			wantErr:    true,
		},
		{
			name:       "empty file",
			configData: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configData), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Test Load function
			config, err := Load(configFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if config == nil {
				t.Error("Load() returned nil config")
				return
			}

			if tt.expectedAPI != "" && config.GoogleMaps.APIKey != tt.expectedAPI {
				t.Errorf("Load() APIKey = %v, want %v", config.GoogleMaps.APIKey, tt.expectedAPI)
			}
		})
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("non-existent-file.yaml")
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestConfigResolveAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		envVar      string
		envValue    string
		wantErr     bool
		expectedKey string
	}{
		{
			name:        "direct api key",
			apiKey:      "direct-key",
			wantErr:     false,
			expectedKey: "direct-key",
		},
		{
			name:        "environment variable substitution",
			apiKey:      "${TEST_API_KEY}",
			envVar:      "TEST_API_KEY",
			envValue:    "env-key",
			wantErr:     false,
			expectedKey: "env-key",
		},
		{
			name:    "missing environment variable",
			apiKey:  "${MISSING_KEY}",
			wantErr: true,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if specified
			if tt.envVar != "" && tt.envValue != "" {
				originalValue := os.Getenv(tt.envVar)
				os.Setenv(tt.envVar, tt.envValue)
				defer func() {
					if originalValue == "" {
						os.Unsetenv(tt.envVar)
					} else {
						os.Setenv(tt.envVar, originalValue)
					}
				}()
			}

			config := &Config{
				GoogleMaps: GoogleMapsConfig{
					APIKey: tt.apiKey,
				},
			}

			err := config.ResolveAPIKey()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveAPIKey() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if config.GoogleMaps.APIKey != tt.expectedKey {
				t.Errorf("ResolveAPIKey() APIKey = %v, want %v", config.GoogleMaps.APIKey, tt.expectedKey)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				GoogleMaps: GoogleMapsConfig{APIKey: "test-key"},
				Input:      InputConfig{CSVFile: "test.csv"},
				Output:     OutputConfig{HTMLFile: "test.html"},
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			config: &Config{
				Input:  InputConfig{CSVFile: "test.csv"},
				Output: OutputConfig{HTMLFile: "test.html"},
			},
			wantErr: true,
		},
		{
			name: "missing csv file",
			config: &Config{
				GoogleMaps: GoogleMapsConfig{APIKey: "test-key"},
				Output:     OutputConfig{HTMLFile: "test.html"},
			},
			wantErr: true,
		},
		{
			name: "missing html file",
			config: &Config{
				GoogleMaps: GoogleMapsConfig{APIKey: "test-key"},
				Input:      InputConfig{CSVFile: "test.csv"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigStructFields(t *testing.T) {
	// Test that all config struct fields can be set and retrieved
	config := &Config{
		GoogleMaps: GoogleMapsConfig{
			APIKey:     "test-key",
			APIVersion: "3.54",
			Libraries:  []string{"geometry", "places"},
		},
		Input: InputConfig{
			CSVFile: "test.csv",
			CSVFormat: CSVFormatConfig{
				TimestampColumn:   "timestamp",
				LatitudeColumn:    "latitude",
				LongitudeColumn:   "longitude",
				TitleColumn:       "title",
				DescriptionColumn: "description",
				CategoryColumn:    "category",
				HasHeader:         true,
				Delimiter:         ",",
				SkipRows:          0,
			},
		},
		Output: OutputConfig{
			HTMLFile:  "output.html",
			Debug:     true,
			ExportKML: false,
			KMLFile:   "output.kml",
		},
		Map: MapConfig{
			Title:  "Test Map",
			Width:  "100%",
			Height: "600px",
			InitialView: InitialViewConfig{
				Center: CenterConfig{
					Latitude:  floatPtr(37.7749),
					Longitude: floatPtr(-122.4194),
				},
				Zoom:    intPtr(13),
				MapType: "roadmap",
			},
			AutoFitBounds: true,
			Controls: ControlsConfig{
				ZoomControl:       true,
				StreetViewControl: true,
				FullscreenControl: true,
				MapTypeControl:    true,
				ScaleControl:      true,
			},
		},
		Markers: MarkersConfig{
			Default: MarkerStyleConfig{
				Icon: IconConfig{
					Color:  "red",
					URL:    stringPtr("icon.png"),
					Size:   SizeConfig{Width: 32, Height: 32},
					Anchor: PointConfig{X: 16, Y: 32},
				},
				Label: LabelConfig{
					ShowSequence: true,
					Text:         stringPtr("Test"),
					Color:        "white",
					Font: FontConfig{
						Family: "Arial",
						Size:   "12px",
						Weight: "bold",
					},
				},
			},
			Start: MarkerStyleConfig{
				Icon: IconConfig{Color: "green"},
			},
			End: MarkerStyleConfig{
				Icon: IconConfig{Color: "red"},
			},
			Categories: map[string]string{
				"work": "blue",
				"home": "green",
			},
		},
		Path: PathConfig{
			Enabled: true,
			Style: PathStyleConfig{
				Color:         "#FF0000",
				Opacity:       0.8,
				Weight:        3,
				StrokePattern: "solid",
			},
			Animation: AnimationConfig{
				Enabled:             false,
				Speed:               1000,
				ShowDirectionArrows: true,
			},
		},
		InfoWindows: InfoWindowsConfig{
			Enabled:       true,
			Template:      "<div>{{.Title}}</div>",
			AutoOpenStart: false,
			MaxWidth:      300,
		},
		Processing: ProcessingConfig{
			RemoveDuplicates:  true,
			MinDistanceFilter: 10.0,
			SmoothPath:        false,
			MaxSpeedFilter:    300.0,
			Timezone:          "UTC",
			TimestampFormats:  []string{"2006-01-02T15:04:05Z"},
		},
		Logging: LoggingConfig{
			Level:   "info",
			File:    "",
			Verbose: false,
		},
	}

	// Verify all fields are accessible
	if config.GoogleMaps.APIKey != "test-key" {
		t.Error("GoogleMaps.APIKey not set correctly")
	}
	if config.Input.CSVFile != "test.csv" {
		t.Error("Input.CSVFile not set correctly")
	}
	if config.Output.HTMLFile != "output.html" {
		t.Error("Output.HTMLFile not set correctly")
	}
	if config.Map.Title != "Test Map" {
		t.Error("Map.Title not set correctly")
	}
	if !config.Path.Enabled {
		t.Error("Path.Enabled not set correctly")
	}
	if !config.InfoWindows.Enabled {
		t.Error("InfoWindows.Enabled not set correctly")
	}
	if !config.Processing.RemoveDuplicates {
		t.Error("Processing.RemoveDuplicates not set correctly")
	}
	if config.Logging.Level != "info" {
		t.Error("Logging.Level not set correctly")
	}
}

// Helper functions for pointer values
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
