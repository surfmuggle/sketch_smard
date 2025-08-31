package smard_test

import (
	"context"
	"fmt"
	"log"
	"smard"
	"time"
)

func ExampleClient_GetTimestamps() {
	client := smard.NewClient()
	ctx := context.Background()

	// Get available timestamps for brown coal generation in Germany, hourly resolution
	timestamps, err := client.GetTimestamps(ctx, smard.FilterBrownCoal, smard.RegionGermany, smard.ResolutionHour)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d timestamps\n", len(timestamps))
	if len(timestamps) > 0 {
		// Convert first timestamp to human-readable time
		t := time.Unix(timestamps[0]/1000, (timestamps[0]%1000)*1000000)
		fmt.Printf("First timestamp: %s\n", t.Format(time.RFC3339))
	}
}

func ExampleClient_GetTimeseries() {
	client := smard.NewClient()
	ctx := context.Background()

	// First get the latest available timestamp
	latestTS, err := client.GetLatestTimestamp(ctx, smard.FilterPhotovoltaic, smard.RegionGermany, smard.ResolutionHour)
	if err != nil {
		log.Fatal(err)
	}

	// Get photovoltaic generation data for Germany
	data, err := client.GetTimeseries(ctx, smard.FilterPhotovoltaic, smard.RegionGermany, smard.ResolutionHour, latestTS)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved %d data points\n", len(data.Data))

	// Print first few data points
	for i, point := range data.Data[:min(5, len(data.Data))] {
		valueStr := "null"
		if point.Value != nil {
			valueStr = fmt.Sprintf("%.2f MW", *point.Value)
		}
		fmt.Printf("Point %d: %s = %s\n", i+1, point.Timestamp.Format("2006-01-02 15:04"), valueStr)
	}
}

func ExampleClient_GetTimeseriesRaw() {
	client := smard.NewClient()
	ctx := context.Background()

	// Use a specific timestamp (example from the API documentation)
	timestamp := int64(1627855200000)

	// Get raw data without parsing
	rawData, err := client.GetTimeseriesRaw(ctx, smard.FilterBrownCoal, smard.RegionGermany, smard.ResolutionHour, timestamp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Raw data contains %d points\n", len(rawData.Series))
	if len(rawData.Series) > 0 {
		fmt.Printf("First point: %v\n", rawData.Series[0])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
