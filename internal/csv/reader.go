// Package csv provides functionality for reading and parsing GPS data from CSV files.
//
// @title CSV Reader Package
// @version 1.0
// @description Provides flexible CSV parsing for GPS tracking data with configurable format support
// @description Supports column mapping, timestamp formats, data validation, and error handling
// @description Can process files with/without headers, custom delimiters, and various layouts
//
// Features:
// - Flexible column detection and mapping
// - Multiple timestamp format support
// - Data validation and cleaning
// - Duplicate removal capabilities
// - Comprehensive error handling
package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/saratily/geo-chrono/internal/config"
	"github.com/saratily/geo-chrono/internal/gps"
)

// Reader handles CSV file reading and parsing with configurable format support.
//
// @struct Reader
// @description CSV reader with flexible format configuration
// @description Adapts to different layouts and provides robust error handling
// @property config CSVFormatConfig CSV format settings (columns, delimiters, headers)
// @property processing ProcessingConfig Data processing options (formats, validation, filters)
type Reader struct {
	config     *config.CSVFormatConfig  // @field config CSV format configuration (columns, delimiters, etc.)
	processing *config.ProcessingConfig // @field processing Data processing options (formats, filters, etc.)
}

// NewReader creates a new CSV reader with the specified configuration.
//
// @function NewReader
// @description Creates configured CSV reader instance
// @param csvConfig CSVFormatConfig CSV format settings and column mapping
// @param procConfig ProcessingConfig Data processing and validation options
// @return Reader Configured CSV reader instance
// @example reader := NewReader(csvConfig, procConfig)
func NewReader(csvConfig *config.CSVFormatConfig, procConfig *config.ProcessingConfig) *Reader {
	return &Reader{
		config:     csvConfig,
		processing: procConfig,
	}
}

// ReadFile reads and parses GPS points from a CSV file.
//
// @method ReadFile
// @description Processes CSV file and extracts GPS tracking data
// @param filename string Path to the CSV file to process
// @return gps.Points Collection of parsed GPS points
// @return error Error if file cannot be read or parsed
// @throws FileNotFoundError When CSV file cannot be opened
// @throws ParseError When CSV structure is invalid
// @throws ValidationError When required columns are missing
// @example points, err := reader.ReadFile("tracking.csv")
func (r *Reader) ReadFile(filename string) (gps.Points, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", filename, err)
	}
	defer file.Close()

	// Configure CSV reader with appropriate delimiter
	reader := csv.NewReader(file)
	if r.config.Delimiter != "" {
		reader.Comma = rune(r.config.Delimiter[0])
	}

	// Read all CSV records into memory
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("cannot read CSV: %w", err)
	}

	// Parse records into GPS points
	return r.parseRecords(records)
}

// parseRecords processes CSV records and converts them into GPS points.
//
// @method parseRecords
// @description Converts raw CSV data into structured GPS points
// @param records [][]string Raw CSV records from file
// @return gps.Points Collection of validated GPS points
// @return error Error if parsing or validation fails
// @internal true
// @steps Skip configured header rows, Detect column indices, Parse each record, Validate coordinates
func (r *Reader) parseRecords(records [][]string) (gps.Points, error) {
	// Skip initial rows if configured (e.g., for metadata or comments)
	if r.config.SkipRows > 0 && len(records) > r.config.SkipRows {
		records = records[r.config.SkipRows:]
	}

	// Validate that we have data to process
	if len(records) < 1 {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	// Determine starting row based on header configuration
	startRow := 0
	if r.config.HasHeader {
		if len(records) < 2 {
			return nil, fmt.Errorf("CSV file must have at least a header and one data row")
		}
		startRow = 1 // Skip header row
	}

	// Determine column positions for required fields
	colIndices, err := r.findColumnIndices(records)
	if err != nil {
		return nil, err
	}

	// Process each data row and convert to GPS points
	var points gps.Points
	for i, record := range records[startRow:] {
		point, err := r.parseRecord(record, colIndices, i+startRow+1)
		if err != nil {
			// Log warning but continue processing other rows
			fmt.Printf("Warning: Skipping row %d - %v\n", i+startRow+1, err)
			continue
		}
		points = append(points, *point)
	}

	// Apply data processing filters as configured
	if r.processing.RemoveDuplicates {
		points = points.RemoveDuplicates()
	}

	return points, nil
}

// columnIndices holds the column positions for different data fields.
//
// @struct columnIndices
// @description Maps logical field names to physical CSV column positions
// @property timestamp int Column index for timestamp data (-1 if not found)
// @property latitude int Column index for latitude coordinates (-1 if not found)
type columnIndices struct {
	timestamp   int // @field timestamp Column index for timestamp data
	latitude    int // @field latitude Column index for latitude coordinates
	longitude   int // @field longitude Column index for longitude coordinates
	title       int // @field title Column index for location title/name (optional, -1 if not used)
	description int // @field description Column index for location description (optional, -1 if not used)
}

// findColumnIndices determines the column positions for required and optional fields.
//
// @method findColumnIndices
// @description Maps CSV columns to data fields using headers or default positions
// @param records [][]string All CSV records including potential header row
// @return *columnIndices Column position mapping structure
// @return error Error if required columns cannot be located
// @internal true
// @logic Uses header names when available, falls back to positional defaults
func (r *Reader) findColumnIndices(records [][]string) (*columnIndices, error) {
	indices := &columnIndices{
		timestamp:   -1,
		latitude:    -1,
		longitude:   -1,
		title:       -1,
		description: -1,
	}

	if r.config.HasHeader && len(records) > 0 {
		// Parse header row to find column positions
		header := records[0]
		for i, col := range header {
			colLower := strings.ToLower(col)

			// Match timestamp column using configured name or common defaults
			if r.matchesColumn(colLower, r.config.TimestampColumn, []string{"timestamp", "time", "datetime"}) {
				indices.timestamp = i
			}

			// Match latitude column using configured name or common defaults
			if r.matchesColumn(colLower, r.config.LatitudeColumn, []string{"latitude", "lat"}) {
				indices.latitude = i
			}

			// Match longitude column using configured name or common defaults
			if r.matchesColumn(colLower, r.config.LongitudeColumn, []string{"longitude", "lon", "lng"}) {
				indices.longitude = i
			}

			// Match optional title column (exact match required if configured)
			if r.config.TitleColumn != "" && colLower == strings.ToLower(r.config.TitleColumn) {
				indices.title = i
			}

			// Match optional description column (exact match required if configured)
			if r.config.DescriptionColumn != "" && colLower == strings.ToLower(r.config.DescriptionColumn) {
				indices.description = i
			}
		}
	} else {
		// Use default column positions when no header is present
		// Assumed order: timestamp, latitude, longitude, [title], [description]
		indices.timestamp = 0
		indices.latitude = 1
		indices.longitude = 2

		// Optional columns if enough columns are present
		if len(records) > 0 && len(records[0]) > 3 {
			indices.title = 3
		}
		if len(records) > 0 && len(records[0]) > 4 {
			indices.description = 4
		}
	}

	// Validate that all required columns were found
	if indices.timestamp == -1 || indices.latitude == -1 || indices.longitude == -1 {
		return nil, fmt.Errorf("CSV must contain timestamp, latitude, and longitude columns")
	}

	return indices, nil
}

// matchesColumn checks if a column name matches either the configured name or default alternatives.
//
// @method matchesColumn
// @description Flexible column name matching for CSV format variations
// @param colName string Column name from CSV header (normalized to lowercase)
// @param configName string User-configured column name (takes precedence)
// @param defaults []string List of common default column names to try
// @return bool True if column name matches configured or default names
// @internal true
// @logic Prioritizes configured names, then tries common variations
func (r *Reader) matchesColumn(colName, configName string, defaults []string) bool {
	// If a specific column name is configured, use exact match
	if configName != "" {
		return colName == strings.ToLower(configName)
	}

	// Otherwise, check against common default column names
	for _, defaultName := range defaults {
		if colName == defaultName {
			return true
		}
	}
	return false
}

// parseRecord processes a single CSV record and converts it to a GPS point.
//
// @method parseRecord
// @description Converts single CSV row to validated GPS point
// @param record []string Individual CSV record fields
// @param indices *columnIndices Column position mapping
// @param rowNum int Row number for error context
// @return *gps.Point Parsed GPS point with coordinates and metadata
// @return error Error if parsing or validation fails
// @internal true
// @validation Checks column count, coordinate ranges, timestamp formats
func (r *Reader) parseRecord(record []string, indices *columnIndices, rowNum int) (*gps.Point, error) {
	// Validate that the record has enough columns for all required fields
	if len(record) <= indices.timestamp || len(record) <= indices.latitude || len(record) <= indices.longitude {
		return nil, fmt.Errorf("insufficient columns")
	}

	// Parse the timestamp field using configured or default formats
	timestamp, err := r.parseTimestamp(record[indices.timestamp])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp '%s': %w", record[indices.timestamp], err)
	}

	// Parse latitude coordinate, trimming whitespace for robustness
	lat, err := strconv.ParseFloat(strings.TrimSpace(record[indices.latitude]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude '%s': %w", record[indices.latitude], err)
	}

	// Parse longitude coordinate, trimming whitespace for robustness
	lng, err := strconv.ParseFloat(strings.TrimSpace(record[indices.longitude]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude '%s': %w", record[indices.longitude], err)
	}

	// Create the GPS point with required fields
	point := &gps.Point{
		Timestamp: timestamp,
		Latitude:  lat,
		Longitude: lng,
	}

	// Add optional title field if configured and present in the record
	if indices.title != -1 && indices.title < len(record) {
		point.Title = strings.TrimSpace(record[indices.title])
	}

	// Add optional description field if configured and present in the record
	if indices.description != -1 && indices.description < len(record) {
		point.Description = strings.TrimSpace(record[indices.description])
	}

	return point, nil
}

// parseTimestamp attempts to parse a timestamp string using configured formats first,
// then falls back to common default formats.
//
// @method parseTimestamp
// @description Parses timestamp strings using multiple format attempts
// @param s string Input timestamp string to parse
// @return time.Time Parsed timestamp value
// @return error Error if no format successfully parses the input
// @internal true
// @formats Tries configured formats first, then common defaults
func (r *Reader) parseTimestamp(s string) (time.Time, error) {
	// Clean the input string by trimming whitespace
	s = strings.TrimSpace(s)

	// Try configured timestamp formats first (user-specified formats take precedence)
	for _, format := range r.processing.TimestampFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// If no configured formats work, fallback to common default formats
	return parseDefaultTimestamp(s)
}

// parseDefaultTimestamp attempts to parse timestamp strings using a comprehensive set
// of common timestamp formats.
//
// @function parseDefaultTimestamp
// @description Parses timestamps using common format patterns
// @param s string Input timestamp string
// @return time.Time Parsed timestamp value
// @return error Error if all format attempts fail
// @internal true
// @formats ISO8601, database, regional, Unix timestamp
// @fallback Tries Unix timestamp as last resort
func parseDefaultTimestamp(s string) (time.Time, error) {
	// Define common timestamp formats ordered by likelihood and specificity
	formats := []string{
		"2006-01-02T15:04:05Z",      // ISO 8601 UTC (most common in APIs)
		"2006-01-02T15:04:05-07:00", // ISO 8601 with timezone offset
		"2006-01-02 15:04:05",       // Simple datetime (database format)
		"01/02/2006 15:04:05",       // US format (MM/DD/YYYY HH:MM:SS)
		"02/01/2006 15:04:05",       // European format (DD/MM/YYYY HH:MM:SS)
		"2006-01-02T15:04:05.000Z",  // ISO 8601 with milliseconds
		"2006-01-02",                // Date only (time assumed as midnight)
	}

	// Try each format until one succeeds
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// As a last resort, try parsing as Unix timestamp (seconds since epoch)
	if unix, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(unix, 0), nil
	}

	// If all parsing attempts fail, return an error with the problematic input
	return time.Time{}, fmt.Errorf("cannot parse timestamp format: %s", s)
}
