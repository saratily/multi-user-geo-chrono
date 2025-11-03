// Package gps_test provides comprehensive unit tests for GPS data structures and operations.
// It tests GPS point creation, geographical calculations, bounds computation, time range analysis,
// sorting functionality, and mathematical accuracy for coordinate-based operations.
package gps

import (
	"math"
	"testing"
	"time"
)

func TestPoint(t *testing.T) {
	// Test Point creation and field access
	timestamp := time.Date(2025, 10, 28, 12, 0, 0, 0, time.UTC)
	point := Point{
		Timestamp:   timestamp,
		Latitude:    37.7749,
		Longitude:   -122.4194,
		Title:       "San Francisco",
		Description: "Test location",
	}

	if point.Timestamp != timestamp {
		t.Errorf("Point.Timestamp = %v, want %v", point.Timestamp, timestamp)
	}
	if point.Latitude != 37.7749 {
		t.Errorf("Point.Latitude = %v, want %v", point.Latitude, 37.7749)
	}
	if point.Longitude != -122.4194 {
		t.Errorf("Point.Longitude = %v, want %v", point.Longitude, -122.4194)
	}
	if point.Title != "San Francisco" {
		t.Errorf("Point.Title = %v, want %v", point.Title, "San Francisco")
	}
	if point.Description != "Test location" {
		t.Errorf("Point.Description = %v, want %v", point.Description, "Test location")
	}
}

func TestPointsSortByTimestamp(t *testing.T) {
	now := time.Now()
	points := Points{
		{Timestamp: now.Add(2 * time.Hour), Latitude: 3, Longitude: 3},
		{Timestamp: now, Latitude: 1, Longitude: 1},
		{Timestamp: now.Add(1 * time.Hour), Latitude: 2, Longitude: 2},
	}

	points.SortByTimestamp()

	if points[0].Latitude != 1 || points[1].Latitude != 2 || points[2].Latitude != 3 {
		t.Error("SortByTimestamp() did not sort points correctly")
	}
}

func TestPointsIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		points Points
		want   bool
	}{
		{
			name:   "empty points",
			points: Points{},
			want:   true,
		},
		{
			name:   "nil points",
			points: nil,
			want:   true,
		},
		{
			name: "non-empty points",
			points: Points{
				{Latitude: 1, Longitude: 1},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.points.IsEmpty(); got != tt.want {
				t.Errorf("Points.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointsFirst(t *testing.T) {
	tests := []struct {
		name   string
		points Points
		want   *Point
	}{
		{
			name:   "empty points",
			points: Points{},
			want:   nil,
		},
		{
			name:   "nil points",
			points: nil,
			want:   nil,
		},
		{
			name: "single point",
			points: Points{
				{Latitude: 1, Longitude: 1, Title: "First"},
			},
			want: &Point{Latitude: 1, Longitude: 1, Title: "First"},
		},
		{
			name: "multiple points",
			points: Points{
				{Latitude: 1, Longitude: 1, Title: "First"},
				{Latitude: 2, Longitude: 2, Title: "Second"},
			},
			want: &Point{Latitude: 1, Longitude: 1, Title: "First"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.points.First()
			if tt.want == nil {
				if got != nil {
					t.Errorf("Points.First() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("Points.First() = nil, want %v", tt.want)
				return
			}
			if got.Latitude != tt.want.Latitude || got.Longitude != tt.want.Longitude || got.Title != tt.want.Title {
				t.Errorf("Points.First() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointsLast(t *testing.T) {
	tests := []struct {
		name   string
		points Points
		want   *Point
	}{
		{
			name:   "empty points",
			points: Points{},
			want:   nil,
		},
		{
			name:   "nil points",
			points: nil,
			want:   nil,
		},
		{
			name: "single point",
			points: Points{
				{Latitude: 1, Longitude: 1, Title: "Only"},
			},
			want: &Point{Latitude: 1, Longitude: 1, Title: "Only"},
		},
		{
			name: "multiple points",
			points: Points{
				{Latitude: 1, Longitude: 1, Title: "First"},
				{Latitude: 2, Longitude: 2, Title: "Last"},
			},
			want: &Point{Latitude: 2, Longitude: 2, Title: "Last"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.points.Last()
			if tt.want == nil {
				if got != nil {
					t.Errorf("Points.Last() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("Points.Last() = nil, want %v", tt.want)
				return
			}
			if got.Latitude != tt.want.Latitude || got.Longitude != tt.want.Longitude || got.Title != tt.want.Title {
				t.Errorf("Points.Last() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPointsRemoveDuplicates(t *testing.T) {
	points := Points{
		{Latitude: 37.7749, Longitude: -122.4194, Title: "SF1"},
		{Latitude: 37.7749, Longitude: -122.4194, Title: "SF2"}, // Duplicate
		{Latitude: 40.7128, Longitude: -74.0060, Title: "NYC"},
		{Latitude: 37.7749, Longitude: -122.4194, Title: "SF3"}, // Another duplicate
	}

	result := points.RemoveDuplicates()

	if len(result) != 2 {
		t.Errorf("RemoveDuplicates() returned %d points, want 2", len(result))
	}

	// Should keep the first occurrence of each unique location
	if result[0].Title != "SF1" {
		t.Errorf("RemoveDuplicates() first point title = %v, want SF1", result[0].Title)
	}
	if result[1].Title != "NYC" {
		t.Errorf("RemoveDuplicates() second point title = %v, want NYC", result[1].Title)
	}
}

func TestPointsRemoveDuplicatesEmpty(t *testing.T) {
	var points Points
	result := points.RemoveDuplicates()

	if len(result) != 0 {
		t.Errorf("RemoveDuplicates() on empty points returned %d points, want 0", len(result))
	}
}

func TestPointsBounds(t *testing.T) {
	tests := []struct {
		name                                           string
		points                                         Points
		wantMinLat, wantMaxLat, wantMinLng, wantMaxLng float64
	}{
		{
			name:       "empty points",
			points:     Points{},
			wantMinLat: 0, wantMaxLat: 0, wantMinLng: 0, wantMaxLng: 0,
		},
		{
			name: "single point",
			points: Points{
				{Latitude: 37.7749, Longitude: -122.4194},
			},
			wantMinLat: 37.7749, wantMaxLat: 37.7749, wantMinLng: -122.4194, wantMaxLng: -122.4194,
		},
		{
			name: "multiple points",
			points: Points{
				{Latitude: 37.7749, Longitude: -122.4194}, // San Francisco
				{Latitude: 40.7128, Longitude: -74.0060},  // New York
				{Latitude: 34.0522, Longitude: -118.2437}, // Los Angeles
			},
			wantMinLat: 34.0522, wantMaxLat: 40.7128, wantMinLng: -122.4194, wantMaxLng: -74.0060,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minLat, maxLat, minLng, maxLng := tt.points.Bounds()

			if minLat != tt.wantMinLat {
				t.Errorf("Points.Bounds() minLat = %v, want %v", minLat, tt.wantMinLat)
			}
			if maxLat != tt.wantMaxLat {
				t.Errorf("Points.Bounds() maxLat = %v, want %v", maxLat, tt.wantMaxLat)
			}
			if minLng != tt.wantMinLng {
				t.Errorf("Points.Bounds() minLng = %v, want %v", minLng, tt.wantMinLng)
			}
			if maxLng != tt.wantMaxLng {
				t.Errorf("Points.Bounds() maxLng = %v, want %v", maxLng, tt.wantMaxLng)
			}
		})
	}
}

func TestPointsCenter(t *testing.T) {
	tests := []struct {
		name    string
		points  Points
		wantLat float64
		wantLng float64
	}{
		{
			name:    "empty points",
			points:  Points{},
			wantLat: 0,
			wantLng: 0,
		},
		{
			name: "single point",
			points: Points{
				{Latitude: 37.7749, Longitude: -122.4194},
			},
			wantLat: 37.7749,
			wantLng: -122.4194,
		},
		{
			name: "two points",
			points: Points{
				{Latitude: 0, Longitude: 0},
				{Latitude: 10, Longitude: 20},
			},
			wantLat: 5,
			wantLng: 10,
		},
		{
			name: "multiple points",
			points: Points{
				{Latitude: 37.7749, Longitude: -122.4194}, // San Francisco
				{Latitude: 40.7128, Longitude: -74.0060},  // New York
				{Latitude: 34.0522, Longitude: -118.2437}, // Los Angeles
			},
			wantLat: (37.7749 + 40.7128 + 34.0522) / 3,
			wantLng: (-122.4194 + -74.0060 + -118.2437) / 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng := tt.points.Center()

			if math.Abs(lat-tt.wantLat) > 0.0001 {
				t.Errorf("Points.Center() lat = %v, want %v", lat, tt.wantLat)
			}
			if math.Abs(lng-tt.wantLng) > 0.0001 {
				t.Errorf("Points.Center() lng = %v, want %v", lng, tt.wantLng)
			}
		})
	}
}

func TestPointsTimeRange(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	later := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		points    Points
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "empty points",
			points:    Points{},
			wantStart: time.Time{},
			wantEnd:   time.Time{},
		},
		{
			name: "single point",
			points: Points{
				{Timestamp: now, Latitude: 1, Longitude: 1},
			},
			wantStart: now,
			wantEnd:   now,
		},
		{
			name: "multiple points ordered",
			points: Points{
				{Timestamp: earlier, Latitude: 1, Longitude: 1},
				{Timestamp: now, Latitude: 2, Longitude: 2},
				{Timestamp: later, Latitude: 3, Longitude: 3},
			},
			wantStart: earlier,
			wantEnd:   later,
		},
		{
			name: "multiple points unordered",
			points: Points{
				{Timestamp: now, Latitude: 2, Longitude: 2},
				{Timestamp: later, Latitude: 3, Longitude: 3},
				{Timestamp: earlier, Latitude: 1, Longitude: 1},
			},
			wantStart: earlier,
			wantEnd:   later,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := tt.points.TimeRange()

			if !start.Equal(tt.wantStart) {
				t.Errorf("Points.TimeRange() start = %v, want %v", start, tt.wantStart)
			}
			if !end.Equal(tt.wantEnd) {
				t.Errorf("Points.TimeRange() end = %v, want %v", end, tt.wantEnd)
			}
		})
	}
}

func TestPointsLenMethod(t *testing.T) {
	tests := []struct {
		name   string
		points Points
		want   int
	}{
		{
			name:   "empty points",
			points: Points{},
			want:   0,
		},
		{
			name:   "nil points",
			points: nil,
			want:   0,
		},
		{
			name: "single point",
			points: Points{
				{Latitude: 1, Longitude: 1},
			},
			want: 1,
		},
		{
			name: "multiple points",
			points: Points{
				{Latitude: 1, Longitude: 1},
				{Latitude: 2, Longitude: 2},
				{Latitude: 3, Longitude: 3},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(tt.points); got != tt.want {
				t.Errorf("len(Points) = %v, want %v", got, tt.want)
			}
		})
	}
}
