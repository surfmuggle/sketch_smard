package smard

import "fmt"

// ValidateFilter checks if the provided filter is valid
func ValidateFilter(filter Filter) error {
	valid := map[Filter]bool{
		FilterBrownCoal:         true,
		FilterNuclear:           true,
		FilterWindOffshore:      true,
		FilterHydro:             true,
		FilterOtherConventional: true,
		FilterOtherRenewable:    true,
		FilterBiomass:           true,
		FilterWindOnshore:       true,
		FilterPhotovoltaic:      true,
		FilterHardCoal:          true,
		FilterPumpedStorage:     true,
		FilterNaturalGas:        true,
		FilterTotalConsumption:  true,
		FilterResidualLoad:      true,
		FilterPumpedStorageLoad: true,
	}

	if !valid[filter] {
		return fmt.Errorf("invalid filter: %s", filter)
	}
	return nil
}

// ValidateRegion checks if the provided region is valid
func ValidateRegion(region Region) error {
	valid := map[Region]bool{
		RegionGermany:    true,
		RegionAustria:    true,
		RegionLuxembourg: true,
		RegionDELU:       true,
		RegionDEATLU:     true,
		Region50Hertz:    true,
		RegionAmprion:    true,
		RegionTenneT:     true,
		RegionTransnetBW: true,
		RegionAPG:        true,
		RegionCreos:      true,
	}

	if !valid[region] {
		return fmt.Errorf("invalid region: %s", region)
	}
	return nil
}

// ValidateResolution checks if the provided resolution is valid
func ValidateResolution(resolution Resolution) error {
	valid := map[Resolution]bool{
		ResolutionHour:        true,
		ResolutionQuarterHour: true,
		ResolutionDay:         true,
		ResolutionWeek:        true,
		ResolutionMonth:       true,
		ResolutionYear:        true,
	}

	if !valid[resolution] {
		return fmt.Errorf("invalid resolution: %s", resolution)
	}
	return nil
}

// ValidateParameters validates all parameters at once
func ValidateParameters(filter Filter, region Region, resolution Resolution) error {
	if err := ValidateFilter(filter); err != nil {
		return err
	}
	if err := ValidateRegion(region); err != nil {
		return err
	}
	if err := ValidateResolution(resolution); err != nil {
		return err
	}
	return nil
}
