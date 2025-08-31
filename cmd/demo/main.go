package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"smard"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type PowerData struct {
	Timestamp         time.Time
	BrownCoal         *float64 // Braunkohle
	HardCoal          *float64 // Steinkohle
	WindOnshore       *float64 // Wind Onshore
	WindOffshore      *float64 // Wind Offshore
	Photovoltaic      *float64 // Photovoltaik
	Nuclear           *float64 // Kernenergie
	Hydro             *float64 // Wasserkraft
	NaturalGas        *float64 // Erdgas
	PumpedStorage     *float64 // Pumpspeicher
	Biomass           *float64 // Biomasse
	OtherConventional *float64 // Sonstige Konventionelle
	OtherRenewable    *float64 // Sonstige Erneuerbare
	TotalConsumption  *float64 // Gesamt Verbrauch
	ResidualLoad      *float64 // Residuallast
	PumpedStorageLoad *float64 // Pumpspeicher Verbrauch
}

func main() {
	// Remove existing database for fresh start
	if err := os.Remove("smard_data.db"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Could not remove existing database: %v", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", "smard_data.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create tables
	if err := createTables(db); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Create SMARD client
	client := smard.NewClient()
	ctx := context.Background()

	fmt.Println("SMARD SQLite Demo")
	fmt.Println("=================")
	fmt.Println()

	// Get and store timestamps for quarter-hour resolution
	fmt.Println("Fetching available timestamps for quarter-hour resolution...")
	timestamps, err := getAndStoreTimestamps(ctx, client, db)
	if err != nil {
		log.Fatal("Failed to get timestamps:", err)
	}

	fmt.Printf("Found and stored %d timestamps\n", len(timestamps))
	if len(timestamps) > 0 {
		first := time.Unix(timestamps[0]/1000, (timestamps[0]%1000)*1000000)
		last := time.Unix(timestamps[len(timestamps)-1]/1000, (timestamps[len(timestamps)-1]%1000)*1000000)
		fmt.Printf("Range: %s to %s\n\n", first.Format("2006-01-02 15:04"), last.Format("2006-01-02 15:04"))
	}

	// Fetch and store power generation data for the latest timestamp
	if len(timestamps) > 0 {
		latestTimestamp := timestamps[len(timestamps)-1]
		fmt.Printf("Fetching power generation data for latest timestamp: %s\n",
			time.Unix(latestTimestamp/1000, (latestTimestamp%1000)*1000000).Format("2006-01-02 15:04"))

		powerData, err := fetchPowerData(ctx, client, latestTimestamp)
		if err != nil {
			log.Fatal("Failed to fetch power data:", err)
		}

		if err := storePowerData(db, powerData); err != nil {
			log.Fatal("Failed to store power data:", err)
		}

		fmt.Printf("Successfully stored %d power generation records\n\n", len(powerData))
	}

	// Display some statistics
	displayStatistics(db)

	fmt.Println("Demo completed! Data stored in smard_data.db")
}

func createTables(db *sql.DB) error {
	// Create timestamps table
	timestampsSQL := `
	CREATE TABLE IF NOT EXISTS timestamps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp_ms INTEGER UNIQUE NOT NULL,
		timestamp_dt DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Create power generation data table
	powerDataSQL := `
	CREATE TABLE IF NOT EXISTS power_generation (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp_dt DATETIME NOT NULL,
		brown_coal REAL,          -- Braunkohle
		hard_coal REAL,           -- Steinkohle  
		wind_onshore REAL,        -- Wind Onshore
		wind_offshore REAL,       -- Wind Offshore
		photovoltaic REAL,        -- Photovoltaik
		nuclear REAL,             -- Kernenergie
		hydro REAL,               -- Wasserkraft
		natural_gas REAL,         -- Erdgas
		pumped_storage REAL,      -- Pumpspeicher
		biomass REAL,             -- Biomasse
		other_conventional REAL,  -- Sonstige Konventionelle
		other_renewable REAL,     -- Sonstige Erneuerbare
		total_consumption REAL,   -- Gesamt Verbrauch
		residual_load REAL,       -- Residuallast
		pumped_storage_load REAL, -- Pumpspeicher Verbrauch
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(timestamp_dt)
	);
	`

	// Create indexes for better performance
	indexSQL := `
	CREATE INDEX IF NOT EXISTS idx_timestamps_dt ON timestamps(timestamp_dt);
	CREATE INDEX IF NOT EXISTS idx_power_timestamp ON power_generation(timestamp_dt);
	`

	for _, sqlStmt := range []string{timestampsSQL, powerDataSQL, indexSQL} {
		if _, err := db.Exec(sqlStmt); err != nil {
			return fmt.Errorf("executing SQL: %w", err)
		}
	}

	return nil
}

func getAndStoreTimestamps(ctx context.Context, client *smard.Client, db *sql.DB) ([]int64, error) {
	// Get timestamps from SMARD API (using brown coal as representative)
	timestamps, err := client.GetTimestamps(ctx, smard.FilterBrownCoal, smard.RegionGermany, smard.ResolutionQuarterHour)
	if err != nil {
		return nil, fmt.Errorf("getting timestamps from API: %w", err)
	}

	// Store timestamps in database
	stmt, err := db.Prepare(`INSERT OR IGNORE INTO timestamps (timestamp_ms, timestamp_dt) VALUES (?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("preparing timestamp insert statement: %w", err)
	}
	defer stmt.Close()

	for _, ts := range timestamps {
		timestamp := time.Unix(ts/1000, (ts%1000)*1000000)
		if _, err := stmt.Exec(ts, timestamp); err != nil {
			return nil, fmt.Errorf("inserting timestamp %d: %w", ts, err)
		}
	}

	return timestamps, nil
}

func fetchPowerData(ctx context.Context, client *smard.Client, timestamp int64) ([]PowerData, error) {
	// Define all power sources we want to fetch
	powerSources := map[smard.Filter]string{
		smard.FilterBrownCoal:         "Brown Coal",
		smard.FilterHardCoal:          "Hard Coal",
		smard.FilterWindOnshore:       "Wind Onshore",
		smard.FilterWindOffshore:      "Wind Offshore",
		smard.FilterPhotovoltaic:      "Photovoltaic",
		smard.FilterNuclear:           "Nuclear",
		smard.FilterHydro:             "Hydro",
		smard.FilterNaturalGas:        "Natural Gas",
		smard.FilterPumpedStorage:     "Pumped Storage",
		smard.FilterBiomass:           "Biomass",
		smard.FilterOtherConventional: "Other Conventional",
		smard.FilterOtherRenewable:    "Other Renewable",
		smard.FilterTotalConsumption:  "Total Consumption",
		smard.FilterResidualLoad:      "Residual Load",
		smard.FilterPumpedStorageLoad: "Pumped Storage Load",
	}

	// Create a map to collect all data points by timestamp
	dataMap := make(map[time.Time]*PowerData)

	// Fetch data for each power source
	for filter, name := range powerSources {
		fmt.Printf("  Fetching %s data...\n", name)

		data, err := client.GetTimeseries(ctx, filter, smard.RegionGermany, smard.ResolutionQuarterHour, timestamp)
		if err != nil {
			log.Printf("  Warning: Failed to fetch %s data: %v", name, err)
			continue
		}

		// Process each data point
		for _, point := range data.Data {
			if _, exists := dataMap[point.Timestamp]; !exists {
				dataMap[point.Timestamp] = &PowerData{Timestamp: point.Timestamp}
			}

			// Set the appropriate field based on the filter
			setPowerValue(dataMap[point.Timestamp], filter, point.Value)
		}
	}

	// Convert map to slice
	var result []PowerData
	for _, data := range dataMap {
		result = append(result, *data)
	}

	return result, nil
}

func setPowerValue(data *PowerData, filter smard.Filter, value *float64) {
	switch filter {
	case smard.FilterBrownCoal:
		data.BrownCoal = value
	case smard.FilterHardCoal:
		data.HardCoal = value
	case smard.FilterWindOnshore:
		data.WindOnshore = value
	case smard.FilterWindOffshore:
		data.WindOffshore = value
	case smard.FilterPhotovoltaic:
		data.Photovoltaic = value
	case smard.FilterNuclear:
		data.Nuclear = value
	case smard.FilterHydro:
		data.Hydro = value
	case smard.FilterNaturalGas:
		data.NaturalGas = value
	case smard.FilterPumpedStorage:
		data.PumpedStorage = value
	case smard.FilterBiomass:
		data.Biomass = value
	case smard.FilterOtherConventional:
		data.OtherConventional = value
	case smard.FilterOtherRenewable:
		data.OtherRenewable = value
	case smard.FilterTotalConsumption:
		data.TotalConsumption = value
	case smard.FilterResidualLoad:
		data.ResidualLoad = value
	case smard.FilterPumpedStorageLoad:
		data.PumpedStorageLoad = value
	}
}

func storePowerData(db *sql.DB, powerData []PowerData) error {
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO power_generation (
			timestamp_dt, brown_coal, hard_coal, wind_onshore, wind_offshore, 
			photovoltaic, nuclear, hydro, natural_gas, pumped_storage, 
			biomass, other_conventional, other_renewable, total_consumption, 
			residual_load, pumped_storage_load
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing insert statement: %w", err)
	}
	defer stmt.Close()

	for _, data := range powerData {
		_, err := stmt.Exec(
			data.Timestamp,
			data.BrownCoal,
			data.HardCoal,
			data.WindOnshore,
			data.WindOffshore,
			data.Photovoltaic,
			data.Nuclear,
			data.Hydro,
			data.NaturalGas,
			data.PumpedStorage,
			data.Biomass,
			data.OtherConventional,
			data.OtherRenewable,
			data.TotalConsumption,
			data.ResidualLoad,
			data.PumpedStorageLoad,
		)
		if err != nil {
			return fmt.Errorf("inserting power data for %v: %w", data.Timestamp, err)
		}
	}

	return nil
}

func displayStatistics(db *sql.DB) {
	fmt.Println("Database Statistics:")
	fmt.Println("====================")

	// Count timestamps
	var timestampCount int
	db.QueryRow("SELECT COUNT(*) FROM timestamps").Scan(&timestampCount)
	fmt.Printf("Stored timestamps: %d\n", timestampCount)

	// Count power generation records
	var powerRecordCount int
	db.QueryRow("SELECT COUNT(*) FROM power_generation").Scan(&powerRecordCount)
	fmt.Printf("Power generation records: %d\n", powerRecordCount)

	// Show sample data
	rows, err := db.Query(`
		SELECT timestamp_dt, brown_coal, wind_onshore, photovoltaic, total_consumption
		FROM power_generation 
		ORDER BY timestamp_dt DESC 
		LIMIT 5
	`)
	if err != nil {
		log.Printf("Error querying sample data: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nSample Power Generation Data (Latest 5 records):")
	fmt.Println("Timestamp\t\t\tBrown Coal\tWind Onshore\tPhotovoltaic\tTotal Consumption")
	fmt.Println("================================================================================")

	for rows.Next() {
		var timestamp time.Time
		var brownCoal, windOnshore, photovoltaic, totalConsumption sql.NullFloat64

		if err := rows.Scan(&timestamp, &brownCoal, &windOnshore, &photovoltaic, &totalConsumption); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("%s\t", timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("%.1f MW\t\t", nullFloatToString(brownCoal))
		fmt.Printf("%.1f MW\t\t", nullFloatToString(windOnshore))
		fmt.Printf("%.1f MW\t\t", nullFloatToString(photovoltaic))
		fmt.Printf("%.1f MW\n", nullFloatToString(totalConsumption))
	}

	fmt.Println()
}

func nullFloatToString(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
}
