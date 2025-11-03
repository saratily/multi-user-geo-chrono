// Package csv_test provides comprehensive unit tests for CSV parsing functionality.
// It tests CSV file reading, column detection, data validation, flexible format handling,
// timestamp parsing, error recovery, and various CSV format scenarios commonly encountered
// in GPS tracking data.
package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saratily/geo-chrono/internal/config"
)

func TestNewReader(t *testing.T) {
	csvConfig := &config.CSVFormatConfig{
		HasHeader: true,
		Delimiter: ",",
	}
	procConfig := &config.ProcessingConfig{
		RemoveDuplicates: true,
	}

	reader := NewReader(csvConfig, procConfig)

	if reader == nil {
		t.Error("NewReader() returned nil")
	}
	if reader.config != csvConfig {
		t.Error("NewReader() did not set config correctly")
	}
	if reader.processing != procConfig {
		t.Error("NewReader() did not set processing config correctly")
	}
}

func TestReaderReadFile(t *testing.T) {
	tests := []struct {
		name       string
		csvContent string
		csvConfig  *config.CSVFormatConfig
		procConfig *config.ProcessingConfig
		wantPoints int
		wantErr    bool
		wantFirst  string
	}{
		{
			name: "valid csv with header",
			csvContent: `timestamp,latitude,longitude,title,description
2025-10-28T10:00:00Z,37.7749,-122.4194,Golden Gate Park,Starting point
2025-10-28T11:00:00Z,37.8044,-122.2711,Bay Bridge,Ending point`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader:         true,
				Delimiter:         ",",
				TimestampColumn:   "timestamp",
				LatitudeColumn:    "latitude",
				LongitudeColumn:   "longitude",
				TitleColumn:       "title",
				DescriptionColumn: "description",
			},
			procConfig: &config.ProcessingConfig{
				RemoveDuplicates: false,
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 2,
			wantErr:    false,
			wantFirst:  "Golden Gate Park",
		},
		{
			name: "csv without header",
			csvContent: `2025-10-28T10:00:00Z,37.7749,-122.4194,Point1,Desc1
2025-10-28T11:00:00Z,37.8044,-122.2711,Point2,Desc2`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader: false,
				Delimiter: ",",
			},
			procConfig: &config.ProcessingConfig{
				RemoveDuplicates: false,
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 2,
			wantErr:    false,
			wantFirst:  "Point1",
		},
		{
			name: "csv with duplicates",
			csvContent: `timestamp,latitude,longitude
2025-10-28T10:00:00Z,37.7749,-122.4194
2025-10-28T11:00:00Z,37.7749,-122.4194
2025-10-28T12:00:00Z,37.8044,-122.2711`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader: true,
				Delimiter: ",",
			},
			procConfig: &config.ProcessingConfig{
				RemoveDuplicates: true,
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 2,
			wantErr:    false,
		},
		{
			name: "custom delimiter",
			csvContent: `timestamp;latitude;longitude
2025-10-28T10:00:00Z;37.7749;-122.4194`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader: true,
				Delimiter: ";",
			},
			procConfig: &config.ProcessingConfig{
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 1,
			wantErr:    false,
		},
		{
			name: "skip rows issue",
			csvContent: `# This is a comment
# Another comment
timestamp,latitude,longitude
2025-10-28T10:00:00Z,37.7749,-122.4194`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader: true,
				Delimiter: ",",
				SkipRows:  2,
			},
			procConfig: &config.ProcessingConfig{
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 0,
			wantErr:    true,
		},
		{
			name: "missing required columns",
			csvContent: `timestamp,title
2025-10-28T10:00:00Z,Test Point`,
			csvConfig: &config.CSVFormatConfig{
				HasHeader: true,
				Delimiter: ",",
			},
			procConfig: &config.ProcessingConfig{
				TimestampFormats: []string{"2006-01-02T15:04:05Z"},
			},
			wantPoints: 0,
			wantErr:    true,
		},
		{
			name:       "empty file",
			csvContent: "",
			csvConfig: &config.CSVFormatConfig{
				HasHeader: true,
			},
			procConfig: &config.ProcessingConfig{},
			wantPoints: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary CSV file
			tmpDir := t.TempDir()
			csvFile := filepath.Join(tmpDir, "test.csv")

			err := os.WriteFile(csvFile, []byte(tt.csvContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test CSV file: %v", err)
			}

			// Test ReadFile
			reader := NewReader(tt.csvConfig, tt.procConfig)
			points, err := reader.ReadFile(csvFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadFile() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(points) != tt.wantPoints {
				t.Errorf("ReadFile() returned %d points, want %d", len(points), tt.wantPoints)
			}

			if tt.wantPoints > 0 && tt.wantFirst != "" {
				if points[0].Title != tt.wantFirst {
					t.Errorf("ReadFile() first point title = %v, want %v", points[0].Title, tt.wantFirst)
				}
			}
		})
	}
}

func TestReaderReadFileNonExistent(t *testing.T) {
	reader := NewReader(&config.CSVFormatConfig{}, &config.ProcessingConfig{})
	_, err := reader.ReadFile("non-existent-file.csv")

	if err == nil {
		t.Error("ReadFile() should return error for non-existent file")
	}
}

func TestMatchesColumn(t *testing.T) {
	reader := &Reader{}

	tests := []struct {
		name       string
		colName    string
		configName string
		defaults   []string
		want       bool
	}{
		{
			name:       "exact config match",
			colName:    "timestamp",
			configName: "timestamp",
			defaults:   []string{"time", "datetime"},
			want:       true,
		},
		{
			name:       "case insensitive config match",
			colName:    "timestamp",
			configName: "TIMESTAMP",
			defaults:   []string{"time"},
			want:       true,
		},
		{
			name:       "default match",
			colName:    "time",
			configName: "",
			defaults:   []string{"time", "datetime"},
			want:       true,
		},
		{
			name:       "no match",
			colName:    "other",
			configName: "",
			defaults:   []string{"time", "datetime"},
			want:       false,
		},
		{
			name:       "config overrides default",
			colName:    "time",
			configName: "timestamp",
			defaults:   []string{"time"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reader.matchesColumn(tt.colName, tt.configName, tt.defaults)
			if got != tt.want {
				t.Errorf("matchesColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	reader := &Reader{
		processing: &config.ProcessingConfig{
			TimestampFormats: []string{
				"2006-01-02T15:04:05Z",
				"2006-01-02 15:04:05",
			},
		},
	}

	tests := []struct {
		name         string
		timestamp    string
		wantErr      bool
		expectedYear int
	}{
		{
			name:         "ISO 8601 UTC",
			timestamp:    "2025-10-28T10:00:00Z",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "simple datetime",
			timestamp:    "2025-10-28 10:00:00",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:      "invalid timestamp",
			timestamp: "invalid-timestamp",
			wantErr:   true,
		},
		{
			name:         "whitespace",
			timestamp:    "  2025-10-28T10:00:00Z  ",
			wantErr:      false,
			expectedYear: 2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reader.parseTimestamp(tt.timestamp)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimestamp() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Year() != tt.expectedYear {
				t.Errorf("parseTimestamp() year = %v, want %v", result.Year(), tt.expectedYear)
			}
		})
	}
}

func TestParseDefaultTimestamp(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    string
		wantErr      bool
		expectedYear int
	}{
		{
			name:         "ISO 8601 UTC",
			timestamp:    "2025-10-28T15:04:05Z",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "ISO 8601 with timezone",
			timestamp:    "2025-10-28T15:04:05-07:00",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "simple datetime",
			timestamp:    "2025-10-28 15:04:05",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "US format",
			timestamp:    "10/28/2025 15:04:05",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "European format",
			timestamp:    "28/10/2025 15:04:05",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "ISO 8601 with milliseconds",
			timestamp:    "2025-10-28T15:04:05.000Z",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "date only",
			timestamp:    "2025-10-28",
			wantErr:      false,
			expectedYear: 2025,
		},
		{
			name:         "unix timestamp",
			timestamp:    "1730131445", // Roughly 2024-10-28
			wantErr:      false,
			expectedYear: 2024,
		},
		{
			name:      "invalid format",
			timestamp: "not-a-timestamp",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDefaultTimestamp(tt.timestamp)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDefaultTimestamp() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseDefaultTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Year() != tt.expectedYear {
				t.Errorf("parseDefaultTimestamp() year = %v, want %v", result.Year(), tt.expectedYear)
			}
		})
	}
}

func TestFindColumnIndices(t *testing.T) {
	reader := &Reader{
		config: &config.CSVFormatConfig{
			TimestampColumn:   "time",
			LatitudeColumn:    "lat",
			LongitudeColumn:   "lng",
			TitleColumn:       "name",
			DescriptionColumn: "desc",
			HasHeader:         true,
		},
	}

	tests := []struct {
		name          string
		records       [][]string
		wantTimestamp int
		wantLatitude  int
		wantLongitude int
		wantTitle     int
		wantDesc      int
		wantErr       bool
	}{
		{
			name: "configured columns",
			records: [][]string{
				{"time", "lat", "lng", "name", "desc"},
				{"2025-10-28T10:00:00Z", "37.7749", "-122.4194", "SF", "San Francisco"},
			},
			wantTimestamp: 0,
			wantLatitude:  1,
			wantLongitude: 2,
			wantTitle:     3,
			wantDesc:      4,
			wantErr:       false,
		},
		{
			name: "default column names - should fail",
			records: [][]string{
				{"timestamp", "latitude", "longitude"},
				{"2025-10-28T10:00:00Z", "37.7749", "-122.4194"},
			},
			wantErr: true,
		},
		{
			name: "missing required column",
			records: [][]string{
				{"timestamp", "latitude"},
				{"2025-10-28T10:00:00Z", "37.7749"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indices, err := reader.findColumnIndices(tt.records)

			if tt.wantErr {
				if err == nil {
					t.Errorf("findColumnIndices() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("findColumnIndices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if indices.timestamp != tt.wantTimestamp {
				t.Errorf("findColumnIndices() timestamp = %v, want %v", indices.timestamp, tt.wantTimestamp)
			}
			if indices.latitude != tt.wantLatitude {
				t.Errorf("findColumnIndices() latitude = %v, want %v", indices.latitude, tt.wantLatitude)
			}
			if indices.longitude != tt.wantLongitude {
				t.Errorf("findColumnIndices() longitude = %v, want %v", indices.longitude, tt.wantLongitude)
			}
			if indices.title != tt.wantTitle {
				t.Errorf("findColumnIndices() title = %v, want %v", indices.title, tt.wantTitle)
			}
			if indices.description != tt.wantDesc {
				t.Errorf("findColumnIndices() description = %v, want %v", indices.description, tt.wantDesc)
			}
		})
	}
}

func TestParseRecord(t *testing.T) {
	reader := &Reader{
		processing: &config.ProcessingConfig{
			TimestampFormats: []string{"2006-01-02T15:04:05Z"},
		},
	}

	indices := &columnIndices{
		timestamp:   0,
		latitude:    1,
		longitude:   2,
		title:       3,
		description: 4,
	}

	tests := []struct {
		name      string
		record    []string
		wantErr   bool
		wantTitle string
		wantLat   float64
		wantLng   float64
	}{
		{
			name:      "valid record",
			record:    []string{"2025-10-28T10:00:00Z", "37.7749", "-122.4194", "San Francisco", "Test location"},
			wantErr:   false,
			wantTitle: "San Francisco",
			wantLat:   37.7749,
			wantLng:   -122.4194,
		},
		{
			name:    "insufficient columns",
			record:  []string{"2025-10-28T10:00:00Z", "37.7749"},
			wantErr: true,
		},
		{
			name:    "invalid timestamp",
			record:  []string{"invalid-time", "37.7749", "-122.4194"},
			wantErr: true,
		},
		{
			name:    "invalid latitude",
			record:  []string{"2025-10-28T10:00:00Z", "invalid", "-122.4194"},
			wantErr: true,
		},
		{
			name:    "invalid longitude",
			record:  []string{"2025-10-28T10:00:00Z", "37.7749", "invalid"},
			wantErr: true,
		},
		{
			name:      "whitespace handling",
			record:    []string{"2025-10-28T10:00:00Z", " 37.7749 ", " -122.4194 ", " San Francisco ", " Test location "},
			wantErr:   false,
			wantTitle: "San Francisco",
			wantLat:   37.7749,
			wantLng:   -122.4194,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point, err := reader.parseRecord(tt.record, indices, 1)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseRecord() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("parseRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if point == nil {
				t.Error("parseRecord() returned nil point")
				return
			}

			if point.Title != tt.wantTitle {
				t.Errorf("parseRecord() title = %v, want %v", point.Title, tt.wantTitle)
			}
			if point.Latitude != tt.wantLat {
				t.Errorf("parseRecord() latitude = %v, want %v", point.Latitude, tt.wantLat)
			}
			if point.Longitude != tt.wantLng {
				t.Errorf("parseRecord() longitude = %v, want %v", point.Longitude, tt.wantLng)
			}
		})
	}
}
