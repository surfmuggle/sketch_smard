package smard

import "time"

// Resolution represents the temporal resolution of data
type Resolution string

const (
	ResolutionHour        Resolution = "hour"
	ResolutionQuarterHour Resolution = "quarterhour"
	ResolutionDay         Resolution = "day"
	ResolutionWeek        Resolution = "week"
	ResolutionMonth       Resolution = "month"
	ResolutionYear        Resolution = "year"
)

// Filter represents data type filters for power generation and consumption
type Filter string

const (
	// Power Generation - Conventional
	FilterBrownCoal         Filter = "1223" // Stromerzeugung: Braunkohle
	FilterNuclear           Filter = "1224" // Stromerzeugung: Kernenergie
	FilterWindOffshore      Filter = "1225" // Stromerzeugung: Wind Offshore
	FilterHydro             Filter = "1226" // Stromerzeugung: Wasserkraft
	FilterOtherConventional Filter = "1227" // Stromerzeugung: Sonstige Konventionelle
	FilterOtherRenewable    Filter = "1228" // Stromerzeugung: Sonstige Erneuerbare
	FilterBiomass           Filter = "4066" // Stromerzeugung: Biomasse
	FilterWindOnshore       Filter = "4067" // Stromerzeugung: Wind Onshore
	FilterPhotovoltaic      Filter = "4068" // Stromerzeugung: Photovoltaik
	FilterHardCoal          Filter = "4069" // Stromerzeugung: Steinkohle
	FilterPumpedStorage     Filter = "4070" // Stromerzeugung: Pumpspeicher
	FilterNaturalGas        Filter = "4071" // Stromerzeugung: Erdgas

	// Power Consumption
	FilterTotalConsumption  Filter = "410"  // Stromverbrauch: Gesamt (Netzlast)
	FilterResidualLoad      Filter = "4359" // Stromverbrauch: Residuallast
	FilterPumpedStorageLoad Filter = "4387" // Stromverbrauch: Pumpspeicher
)

// Region represents geographic areas, control zones, and market areas
type Region string

const (
	// Countries
	RegionGermany    Region = "DE"
	RegionAustria    Region = "AT"
	RegionLuxembourg Region = "LU"

	// Market Areas
	RegionDELU   Region = "DE-LU"    // Marktgebiet: DE/LU (ab 01.10.2018)
	RegionDEATLU Region = "DE-AT-LU" // Marktgebiet: DE/AT/LU (bis 30.09.2018)

	// Control Zones Germany
	Region50Hertz    Region = "50Hertz"    // Regelzone (DE): 50Hertz
	RegionAmprion    Region = "Amprion"    // Regelzone (DE): Amprion
	RegionTenneT     Region = "TenneT"     // Regelzone (DE): TenneT
	RegionTransnetBW Region = "TransnetBW" // Regelzone (DE): TransnetBW

	// Control Zones Other
	RegionAPG   Region = "APG"   // Regelzone (AT): APG
	RegionCreos Region = "Creos" // Regelzone (LU): Creos
)

// TimestampResponse represents the response from the timestamps API
type TimestampResponse struct {
	Timestamps []int64 `json:"timestamps"`
}

// DataPoint represents a single data point in a time series
type DataPoint [2]interface{} // [timestamp, value] where value can be float64 or null

// TimeseriesResponse represents the response from the timeseries API
type TimeseriesResponse struct {
	Series []DataPoint `json:"series"`
}

// TimeseriesData represents parsed time series data
type TimeseriesData struct {
	Timestamp time.Time
	Value     *float64 // nil if value is null
}

// ParsedTimeseries represents the parsed response with proper Go types
type ParsedTimeseries struct {
	Data []TimeseriesData
}
