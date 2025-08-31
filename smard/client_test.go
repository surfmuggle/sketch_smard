package smard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateFilter(t *testing.T) {
	tests := []struct {
		filter  Filter
		wantErr bool
	}{
		{FilterBrownCoal, false},
		{FilterPhotovoltaic, false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateFilter(tt.filter)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateFilter(%q) error = %v, wantErr %v", tt.filter, err, tt.wantErr)
		}
	}
}

func TestValidateRegion(t *testing.T) {
	tests := []struct {
		region  Region
		wantErr bool
	}{
		{RegionGermany, false},
		{RegionAustria, false},
		{Region50Hertz, false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateRegion(tt.region)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateRegion(%q) error = %v, wantErr %v", tt.region, err, tt.wantErr)
		}
	}
}

func TestValidateResolution(t *testing.T) {
	tests := []struct {
		resolution Resolution
		wantErr    bool
	}{
		{ResolutionHour, false},
		{ResolutionDay, false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateResolution(tt.resolution)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateResolution(%q) error = %v, wantErr %v", tt.resolution, err, tt.wantErr)
		}
	}
}

func TestGetTimestamps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/1223/DE/index_hour.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"timestamps": [1627855200000, 1627858800000]}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{Timeout: 30 * time.Second}, server.URL)

	ctx := context.Background()
	timestamps, err := client.GetTimestamps(ctx, FilterBrownCoal, RegionGermany, ResolutionHour)
	if err != nil {
		t.Fatalf("GetTimestamps failed: %v", err)
	}

	expected := []int64{1627855200000, 1627858800000}
	if len(timestamps) != len(expected) {
		t.Errorf("Expected %d timestamps, got %d", len(expected), len(timestamps))
	}

	for i, ts := range timestamps {
		if ts != expected[i] {
			t.Errorf("Expected timestamp %d, got %d at index %d", expected[i], ts, i)
		}
	}
}

func TestGetTimestampsValidation(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test invalid filter
	_, err := client.GetTimestamps(ctx, "invalid", RegionGermany, ResolutionHour)
	if err == nil {
		t.Error("Expected error for invalid filter")
	}

	// Test invalid region
	_, err = client.GetTimestamps(ctx, FilterBrownCoal, "invalid", ResolutionHour)
	if err == nil {
		t.Error("Expected error for invalid region")
	}

	// Test invalid resolution
	_, err = client.GetTimestamps(ctx, FilterBrownCoal, RegionGermany, "invalid")
	if err == nil {
		t.Error("Expected error for invalid resolution")
	}
}

func TestGetTimeseries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/1223/DE/1223_DE_hour_1627855200000.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"series": [[1627855200000, 1234.5], [1627858800000, null]]}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{Timeout: 30 * time.Second}, server.URL)

	ctx := context.Background()
	data, err := client.GetTimeseries(ctx, FilterBrownCoal, RegionGermany, ResolutionHour, 1627855200000)
	if err != nil {
		t.Fatalf("GetTimeseries failed: %v", err)
	}

	if len(data.Data) != 2 {
		t.Errorf("Expected 2 data points, got %d", len(data.Data))
	}

	// Check first data point
	firstPoint := data.Data[0]
	if firstPoint.Value == nil {
		t.Error("Expected first point to have a value")
	} else if *firstPoint.Value != 1234.5 {
		t.Errorf("Expected first point value 1234.5, got %f", *firstPoint.Value)
	}

	// Check second data point (should be null)
	secondPoint := data.Data[1]
	if secondPoint.Value != nil {
		t.Error("Expected second point value to be nil")
	}
}

func TestGetTimeseriesValidation(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test invalid timestamp
	_, err := client.GetTimeseries(ctx, FilterBrownCoal, RegionGermany, ResolutionHour, 0)
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}

	_, err = client.GetTimeseries(ctx, FilterBrownCoal, RegionGermany, ResolutionHour, -1)
	if err == nil {
		t.Error("Expected error for negative timestamp")
	}
}

func TestGetLatestTimestamp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"timestamps": [1627855200000, 1627858800000, 1627851600000]}`))
	}))
	defer server.Close()

	client := NewClientWithOptions(&http.Client{Timeout: 30 * time.Second}, server.URL)

	ctx := context.Background()
	latest, err := client.GetLatestTimestamp(ctx, FilterBrownCoal, RegionGermany, ResolutionHour)
	if err != nil {
		t.Fatalf("GetLatestTimestamp failed: %v", err)
	}

	// Should return the maximum timestamp
	expected := int64(1627858800000)
	if latest != expected {
		t.Errorf("Expected latest timestamp %d, got %d", expected, latest)
	}
}

func TestParseTimeseries(t *testing.T) {
	response := TimeseriesResponse{
		Series: []DataPoint{
			{1627855200000.0, 1234.5},
			{1627858800000.0, nil},
			{1627862400000.0, 5678.9},
		},
	}

	parsed := parseTimeseries(response)

	if len(parsed.Data) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(parsed.Data))
	}

	// Check first point
	first := parsed.Data[0]
	expectedTime := time.Unix(1627855200, 0)
	if !first.Timestamp.Equal(expectedTime) {
		t.Errorf("Expected timestamp %v, got %v", expectedTime, first.Timestamp)
	}
	if first.Value == nil || *first.Value != 1234.5 {
		t.Errorf("Expected value 1234.5, got %v", first.Value)
	}

	// Check second point (null value)
	second := parsed.Data[1]
	if second.Value != nil {
		t.Errorf("Expected nil value, got %v", second.Value)
	}

	// Check third point
	third := parsed.Data[2]
	if third.Value == nil || *third.Value != 5678.9 {
		t.Errorf("Expected value 5678.9, got %v", third.Value)
	}
}
