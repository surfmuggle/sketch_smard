package main

import (
	"context"
	"fmt"
	"log"
	"smard"
	"time"
)

func main() {
	// Create a SMARD client
	client := smard.NewClient()
	ctx := context.Background()

	fmt.Println("SMARD API Client Demo")
	fmt.Println("====================\n")

	// Example 1: Get available timestamps for photovoltaic generation in Germany
	fmt.Println("1. Getting available timestamps for photovoltaic generation in Germany (hourly)...")
	timestamps, err := client.GetTimestamps(ctx, smard.FilterPhotovoltaic, smard.RegionGermany, smard.ResolutionHour)
	if err != nil {
		log.Printf("Error getting timestamps: %v", err)
	} else {
		fmt.Printf("   Found %d timestamps\n", len(timestamps))
		if len(timestamps) > 0 {
			// Show first and last timestamp
			first := time.Unix(timestamps[0]/1000, (timestamps[0]%1000)*1000000)
			last := time.Unix(timestamps[len(timestamps)-1]/1000, (timestamps[len(timestamps)-1]%1000)*1000000)
			fmt.Printf("   First: %s\n", first.Format("2006-01-02 15:04"))
			fmt.Printf("   Last:  %s\n\n", last.Format("2006-01-02 15:04"))
		}
	}

	// Example 2: Get latest timestamp and data
	fmt.Println("2. Getting latest photovoltaic data...")
	latestTS, err := client.GetLatestTimestamp(ctx, smard.FilterPhotovoltaic, smard.RegionGermany, smard.ResolutionHour)
	if err != nil {
		log.Printf("Error getting latest timestamp: %v", err)
	} else {
		fmt.Printf("   Latest timestamp: %s\n", time.Unix(latestTS/1000, (latestTS%1000)*1000000).Format("2006-01-02 15:04"))

		// Get the actual data
		data, err := client.GetTimeseries(ctx, smard.FilterPhotovoltaic, smard.RegionGermany, smard.ResolutionHour, latestTS)
		if err != nil {
			log.Printf("Error getting timeseries data: %v", err)
		} else {
			fmt.Printf("   Retrieved %d data points\n", len(data.Data))

			// Show first 5 data points
			fmt.Println("   First 5 data points:")
			for i, point := range data.Data[:min(5, len(data.Data))] {
				valueStr := "null"
				if point.Value != nil {
					valueStr = fmt.Sprintf("%.2f MW", *point.Value)
				}
				fmt.Printf("     %d. %s = %s\n", i+1, point.Timestamp.Format("2006-01-02 15:04"), valueStr)
			}
		}
	}

	fmt.Println("\nDemo completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
