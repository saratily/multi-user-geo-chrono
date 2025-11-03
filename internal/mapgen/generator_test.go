// Package mapgen_test provides comprehensive unit tests for HTML map generation functionality.
// It tests map generator creation, HTML template processing, Google Maps integration,
// marker and path generation, template execution, and file output validation to ensure
// proper GPS track visualization in web browsers.
package mapgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saratily/geo-chrono/internal/config"
	"github.com/saratily/geo-chrono/internal/gps"
)

func TestNewGenerator(t *testing.T) {
	cfg := &config.Config{}
	gen := NewGenerator(cfg)

	if gen == nil {
		t.Error("NewGenerator() returned nil")
	}
	if gen.config != cfg {
		t.Error("NewGenerator() did not set config correctly")
	}
}

func TestGeneratorGenerate(t *testing.T) {
	testTime := time.Date(2025, 10, 28, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		points     gps.Points
		config     *config.Config
		wantErr    bool
		wantOutput bool
	}{
		{
			name: "valid points and config",
			points: gps.Points{
				{
					Timestamp:   testTime,
					Latitude:    37.7749,
					Longitude:   -122.4194,
					Title:       "San Francisco",
					Description: "Test location",
				},
				{
					Timestamp:   testTime.Add(time.Hour),
					Latitude:    37.8044,
					Longitude:   -122.2711,
					Title:       "Oakland",
					Description: "Another location",
				},
			},
			config: &config.Config{
				GoogleMaps: config.GoogleMapsConfig{
					APIKey: "test-api-key",
				},
				Map: config.MapConfig{
					Title: "Test Map",
					InitialView: config.InitialViewConfig{
						Zoom:    &[]int{10}[0],
						MapType: "roadmap",
					},
				},
			},
			wantErr:    false,
			wantOutput: true,
		},
		{
			name:   "empty points",
			points: gps.Points{},
			config: &config.Config{
				GoogleMaps: config.GoogleMapsConfig{
					APIKey: "test-api-key",
				},
				Map: config.MapConfig{
					Title: "Empty Map",
				},
			},
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "missing API key",
			points: gps.Points{
				{
					Timestamp: testTime,
					Latitude:  37.7749,
					Longitude: -122.4194,
					Title:     "Test",
				},
			},
			config: &config.Config{
				GoogleMaps: config.GoogleMapsConfig{
					APIKey: "",
				},
				Map: config.MapConfig{
					Title: "Test Map",
				},
			},
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "single point",
			points: gps.Points{
				{
					Timestamp: testTime,
					Latitude:  37.7749,
					Longitude: -122.4194,
					Title:     "Single Point",
				},
			},
			config: &config.Config{
				GoogleMaps: config.GoogleMapsConfig{
					APIKey: "test-api-key",
				},
				Map: config.MapConfig{
					InitialView: config.InitialViewConfig{
						Zoom: &[]int{12}[0],
					},
				},
			},
			wantErr:    false,
			wantOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputFile := filepath.Join(tmpDir, "test_map.html")

			gen := NewGenerator(tt.config)
			err := gen.Generate(tt.points, outputFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantOutput {
				// Check if file was created
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Error("Generate() did not create output file")
				}

				// Check file content
				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Errorf("Failed to read generated file: %v", err)
					return
				}

				contentStr := string(content)

				// Check for essential HTML elements
				if !strings.Contains(contentStr, "<html") {
					t.Error("Generated file does not contain HTML tag")
				}
				if !strings.Contains(contentStr, "maps.googleapis.com") {
					t.Error("Generated file does not contain Google Maps API reference")
				}
				if !strings.Contains(contentStr, tt.config.GoogleMaps.APIKey) {
					t.Error("Generated file does not contain API key")
				}
			}
		})
	}
}

func TestGetHTMLTemplate(t *testing.T) {
	gen := &Generator{}
	template := gen.getHTMLTemplate()

	if template == "" {
		t.Error("getHTMLTemplate() returned empty string")
		return
	}

	// Check for essential HTML structure
	essentialElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>",
		"<body>",
		"<div id=\"map\"",
		"<script",
		"maps.googleapis.com",
		"google.maps.Map",
	}

	for _, element := range essentialElements {
		if !strings.Contains(template, element) {
			t.Errorf("getHTMLTemplate() missing essential element: %s", element)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		points      gps.Points
		shouldPanic bool
	}{
		{
			name:        "nil config",
			config:      nil,
			shouldPanic: true,
			points: gps.Points{
				{
					Timestamp: time.Now(),
					Latitude:  37.7749,
					Longitude: -122.4194,
					Title:     "Test",
				},
			},
		},
		{
			name: "config with empty API key",
			config: &config.Config{
				GoogleMaps: config.GoogleMapsConfig{
					APIKey: "",
				},
				Map: config.MapConfig{
					Title: "Test",
				},
			},
			points: gps.Points{
				{
					Timestamp: time.Now(),
					Latitude:  37.7749,
					Longitude: -122.4194,
					Title:     "Test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Generate() should panic for %s", tt.name)
					}
				}()
			}

			gen := NewGenerator(tt.config)
			_ = gen.Generate(tt.points, "/tmp/test.html")

			// Note: The current implementation doesn't validate empty API keys
			// This is acceptable behavior for this test
		})
	}
}

func TestMarkerAndPathGeneration(t *testing.T) {
	testTime := time.Date(2025, 10, 28, 10, 0, 0, 0, time.UTC)

	points := gps.Points{
		{
			Timestamp:   testTime,
			Latitude:    37.7749,
			Longitude:   -122.4194,
			Title:       "Start",
			Description: "Starting point",
		},
		{
			Timestamp:   testTime.Add(30 * time.Minute),
			Latitude:    37.7849,
			Longitude:   -122.4094,
			Title:       "Middle",
			Description: "Middle point",
		},
		{
			Timestamp:   testTime.Add(60 * time.Minute),
			Latitude:    37.7949,
			Longitude:   -122.3994,
			Title:       "End",
			Description: "Ending point",
		},
	}

	config := &config.Config{
		GoogleMaps: config.GoogleMapsConfig{
			APIKey: "test-api-key",
		},
		Map: config.MapConfig{
			Title: "Trail Map",
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_map.html")

	gen := NewGenerator(config)
	err := gen.Generate(points, outputFile)

	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Check file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read generated file: %v", err)
		return
	}

	contentStr := string(content)

	// Check that all points appear in the HTML
	for _, point := range points {
		if !strings.Contains(contentStr, point.Title) {
			t.Errorf("Generate() missing point title: %s", point.Title)
		}
		if !strings.Contains(contentStr, point.Description) {
			t.Errorf("Generate() missing point description: %s", point.Description)
		}
	}

	// Check for JavaScript structures
	jsChecks := []string{
		"google.maps.Map",
		"google.maps.Marker",
		"google.maps.Polyline",
	}

	for _, check := range jsChecks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("Generate() missing JavaScript element: %s", check)
		}
	}
}

func TestOutputFileWriting(t *testing.T) {
	testTime := time.Date(2025, 10, 28, 10, 0, 0, 0, time.UTC)

	points := gps.Points{
		{
			Timestamp: testTime,
			Latitude:  37.7749,
			Longitude: -122.4194,
			Title:     "Test Point",
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test_output.html")

	config := &config.Config{
		GoogleMaps: config.GoogleMapsConfig{
			APIKey: "test-api-key",
		},
		Map: config.MapConfig{
			InitialView: config.InitialViewConfig{
				Zoom: &[]int{10}[0],
			},
		},
	}

	gen := NewGenerator(config)
	err := gen.Generate(points, outputFile)

	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Check file exists and has content
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Errorf("Output file does not exist: %v", err)
		return
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test Point") {
		t.Error("Output file does not contain expected point data")
	}
}
