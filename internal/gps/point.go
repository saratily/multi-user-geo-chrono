// Package gps provides data structures and operations for GPS coordinate handling.
//
// @title GPS Data Structures Package
// @version 1.0
// @description Provides GPS coordinate handling with timestamps and metadata
// @description Supports geographical calculations and collection operations
// @description Designed for GPS tracking data from various sources
//
// Features:
// - GPS point structure with coordinates and metadata
// - Collection operations (sorting, filtering, bounds)
// - Geographical calculations
// - Time-based analysis
// - Movement pattern processing
package gps

import (
	"fmt"
	"sort"
	"time"
)

// Point represents a single GPS coordinate with associated metadata.
//
// @struct Point
// @description Single GPS coordinate with timestamp and metadata
// @description Contains essential information for mapping and display
// @property Timestamp time.Time When this GPS point was recorded
// @property Latitude float64 Latitude coordinate (-90.0 to 90.0 degrees)
// @property Longitude float64 Longitude coordinate (-180.0 to 180.0 degrees)
// @property Title string Display name for this location (optional)
// @property Description string Additional details about location (optional)
type Point struct {
	Timestamp   time.Time // @field Timestamp When this GPS point was recorded
	Latitude    float64   // @field Latitude Latitude coordinate (-90.0 to 90.0)
	Longitude   float64   // @field Longitude Longitude coordinate (-180.0 to 180.0)
	Title       string    // @field Title Display name for this location (optional)
	Description string    // @field Description Additional details about this location (optional)
}

// Points represents a collection of GPS points that can be manipulated as a group.
//
// @type Points []Point
// @description Collection of GPS points with group operations
// @description Provides methods for sorting, filtering, and analysis
// @methods SortByTimestamp, IsEmpty, First, Last, Bounds, etc.
type Points []Point

// SortByTimestamp sorts the GPS points in chronological order.
//
// @method SortByTimestamp
// @description Sorts GPS points by timestamp in chronological order
// @receiver p Points Collection of GPS points to sort
// @mutates Sorts the slice in place
// @essential Required for accurate trail visualization and path creation
func (p Points) SortByTimestamp() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Timestamp.Before(p[j].Timestamp)
	})
}

// IsEmpty returns true if the collection contains no GPS points.
// This is useful for validation before processing or displaying GPS data.
func (p Points) IsEmpty() bool {
	return len(p) == 0
}

// First returns a pointer to the first GPS point in the collection.
// Returns nil if the collection is empty. Useful for getting start points.
func (p Points) First() *Point {
	if len(p) == 0 {
		return nil
	}
	return &p[0]
}

// Last returns a pointer to the last GPS point in the collection.
// Returns nil if the collection is empty. Useful for getting end points.
func (p Points) Last() *Point {
	if len(p) == 0 {
		return nil
	}
	return &p[len(p)-1]
}

// RemoveDuplicates removes GPS points that have identical coordinates.
// This helps clean up GPS data by removing redundant points at the same location.
// The comparison is done with 6 decimal places precision (~0.1 meter accuracy).
func (p Points) RemoveDuplicates() Points {
	seen := make(map[string]bool)
	var result Points

	for _, point := range p {
		// Create unique key based on coordinates with reasonable precision
		key := fmt.Sprintf("%.6f,%.6f", point.Latitude, point.Longitude)
		if !seen[key] {
			seen[key] = true
			result = append(result, point)
		}
	}

	return result
}

// Bounds calculates the geographical bounding box that contains all GPS points.
// Returns the minimum and maximum latitude and longitude values.
// This is useful for setting appropriate map zoom levels and center points.
func (p Points) Bounds() (minLat, maxLat, minLng, maxLng float64) {
	if len(p) == 0 {
		return 0, 0, 0, 0
	}

	// Initialize bounds with first point
	minLat, maxLat = p[0].Latitude, p[0].Latitude
	minLng, maxLng = p[0].Longitude, p[0].Longitude

	// Find min/max values across all points
	for _, point := range p[1:] {
		if point.Latitude < minLat {
			minLat = point.Latitude
		}
		if point.Latitude > maxLat {
			maxLat = point.Latitude
		}
		if point.Longitude < minLng {
			minLng = point.Longitude
		}
		if point.Longitude > maxLng {
			maxLng = point.Longitude
		}
	}

	return minLat, maxLat, minLng, maxLng
}

// Center calculates the geographical center point of all GPS coordinates.
// This uses simple arithmetic mean, which works well for small areas.
// For larger areas spanning continents, more sophisticated calculations might be needed.
func (p Points) Center() (lat, lng float64) {
	if len(p) == 0 {
		return 0, 0
	}

	// Calculate average of all coordinates
	var totalLat, totalLng float64
	for _, point := range p {
		totalLat += point.Latitude
		totalLng += point.Longitude
	}

	return totalLat / float64(len(p)), totalLng / float64(len(p))
}

// TimeRange returns the earliest and latest timestamps in the GPS track.
// This is useful for understanding the duration and time span of the GPS data.
func (p Points) TimeRange() (start, end time.Time) {
	if len(p) == 0 {
		return time.Time{}, time.Time{}
	}

	// Initialize with first point's timestamp
	start, end = p[0].Timestamp, p[0].Timestamp

	// Find earliest and latest timestamps
	for _, point := range p[1:] {
		if point.Timestamp.Before(start) {
			start = point.Timestamp
		}
		if point.Timestamp.After(end) {
			end = point.Timestamp
		}
	}

	return start, end
}
