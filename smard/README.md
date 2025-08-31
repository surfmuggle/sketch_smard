# SMARD Go Client

A Go client library for accessing the SMARD API (Bundesnetzagentur German electricity market data).

## Installation

```bash
go get ./smard
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "./smard"
)

func main() {
    client := smard.NewClient()
    ctx := context.Background()
    
    // Get available timestamps
    timestamps, err := client.GetTimestamps(ctx, 
        smard.FilterPhotovoltaic, 
        smard.RegionGermany, 
        smard.ResolutionHour)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d timestamps\n", len(timestamps))
    
    // Get time series data using the latest timestamp
    if len(timestamps) > 0 {
        latest := timestamps[len(timestamps)-1]
        data, err := client.GetTimeseries(ctx, 
            smard.FilterPhotovoltaic, 
            smard.RegionGermany, 
            smard.ResolutionHour, 
            latest)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("Retrieved %d data points\n", len(data.Data))
        
        // Print first few data points
        for i, point := range data.Data[:5] {
            valueStr := "null"
            if point.Value != nil {
                valueStr = fmt.Sprintf("%.2f MW", *point.Value)
            }
            fmt.Printf("Point %d: %s = %s\n", 
                i+1, 
                point.Timestamp.Format("2006-01-02 15:04"), 
                valueStr)
        }
    }
}
```

### Available Filters (Data Types)

#### Power Generation - Conventional
- `FilterBrownCoal` - Brown coal generation
- `FilterNuclear` - Nuclear power generation
- `FilterHardCoal` - Hard coal generation
- `FilterNaturalGas` - Natural gas generation
- `FilterPumpedStorage` - Pumped storage generation
- `FilterOtherConventional` - Other conventional sources

#### Power Generation - Renewable
- `FilterPhotovoltaic` - Solar/Photovoltaic generation
- `FilterWindOnshore` - Onshore wind generation
- `FilterWindOffshore` - Offshore wind generation
- `FilterHydro` - Hydroelectric generation
- `FilterBiomass` - Biomass generation
- `FilterOtherRenewable` - Other renewable sources

#### Power Consumption
- `FilterTotalConsumption` - Total electricity consumption (grid load)
- `FilterResidualLoad` - Residual load
- `FilterPumpedStorageLoad` - Pumped storage consumption

### Available Regions

#### Countries
- `RegionGermany` - Germany (DE)
- `RegionAustria` - Austria (AT)
- `RegionLuxembourg` - Luxembourg (LU)

#### Market Areas
- `RegionDELU` - DE/LU market area (from 01.10.2018)
- `RegionDEATLU` - DE/AT/LU market area (until 30.09.2018)

#### German Control Zones
- `Region50Hertz` - 50Hertz control zone
- `RegionAmprion` - Amprion control zone
- `RegionTenneT` - TenneT control zone
- `RegionTransnetBW` - TransnetBW control zone

#### Other Control Zones
- `RegionAPG` - APG (Austria)
- `RegionCreos` - Creos (Luxembourg)

### Available Resolutions

- `ResolutionHour` - Hourly data
- `ResolutionQuarterHour` - Quarter-hourly data
- `ResolutionDay` - Daily data
- `ResolutionWeek` - Weekly data
- `ResolutionMonth` - Monthly data
- `ResolutionYear` - Yearly data

### Client Methods

#### `GetTimestamps(ctx, filter, region, resolution) ([]int64, error)`

Retrieves available timestamps for the specified parameters. Returns Unix timestamps in milliseconds.

#### `GetTimeseries(ctx, filter, region, resolution, timestamp) (*ParsedTimeseries, error)`

Retrieves time series data starting from the specified timestamp. Returns parsed data with proper Go types:
- Timestamps are converted to `time.Time`
- Values are `*float64` (nil for missing data)

#### `GetTimeseriesRaw(ctx, filter, region, resolution, timestamp) (*TimeseriesResponse, error)`

Retrieves raw time series data without parsing. Useful if you need to handle the original JSON format.

#### `GetLatestTimestamp(ctx, filter, region, resolution) (int64, error)`

Convenience method to get the most recent available timestamp for the given parameters.

### Error Handling

All methods validate input parameters and return descriptive errors for:
- Invalid filter, region, or resolution values
- Network/HTTP errors
- JSON parsing errors
- API error responses

### Custom HTTP Client

```go
import "net/http"

// Use custom HTTP client with different timeout
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}
client := smard.NewClientWithHTTP(httpClient)
```

## API Reference

This client implements the SMARD API as documented by Bundesnetzagentur:
- Base URL: `https://www.smard.de/app/chart_data`
- Timestamp endpoint: `/{filter}/{region}/index_{resolution}.json`
- Timeseries endpoint: `/{filter}/{region}/{filter}_{region}_{resolution}_{timestamp}.json`

Note: The API has a quirky design where filter and region parameters must be duplicated in the timeseries URL. This client handles that automatically.
