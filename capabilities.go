package main

import "strings"

// ModelCapabilities describes the tunable parameters a model accepts.
// This is served to the agent so it never has to guess valid values.
type ModelCapabilities struct {
	AspectRatios   []string `json:"aspect_ratios"`
	Resolutions    []string `json:"resolutions"`
	ThinkingLevels []string `json:"thinking_levels"` // empty => thinking not supported
}

// Common option sets for Gemini image ("nano banana") models on AI Studio.
var (
	commonAspectRatios = []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}
	flashResolutions   = []string{"1K", "2K"}
	proResolutions     = []string{"1K", "2K", "4K"}
	proThinkingLevels  = []string{"LOW", "HIGH"}
)

// CapabilitiesForModel returns the supported parameter values for a given model
// name. It matches on well-known model families and falls back to a safe flash
// profile for unknown image models.
func CapabilitiesForModel(model string) ModelCapabilities {
	m := strings.ToLower(model)

	// Pro image models support higher resolutions and a thinking budget.
	if strings.Contains(m, "pro") && strings.Contains(m, "image") {
		return ModelCapabilities{
			AspectRatios:   commonAspectRatios,
			Resolutions:    proResolutions,
			ThinkingLevels: proThinkingLevels,
		}
	}

	// Flash / flash-lite image models: no thinking, up to 2K.
	return ModelCapabilities{
		AspectRatios:   commonAspectRatios,
		Resolutions:    flashResolutions,
		ThinkingLevels: []string{},
	}
}

// validateOptions checks the requested options against a model's capabilities
// and returns a human-readable error if any value is unsupported.
func validateOptions(model string, opts ImageOptions) error {
	caps := CapabilitiesForModel(model)
	if opts.AspectRatio != "" && !contains(caps.AspectRatios, opts.AspectRatio) {
		return &optionError{"aspect_ratio", opts.AspectRatio, caps.AspectRatios}
	}
	if opts.Resolution != "" && !contains(caps.Resolutions, opts.Resolution) {
		return &optionError{"resolution", opts.Resolution, caps.Resolutions}
	}
	if opts.ThinkingLevel != "" {
		if len(caps.ThinkingLevels) == 0 {
			return &optionError{"thinking_level", opts.ThinkingLevel, caps.ThinkingLevels}
		}
		if !contains(caps.ThinkingLevels, strings.ToUpper(opts.ThinkingLevel)) {
			return &optionError{"thinking_level", opts.ThinkingLevel, caps.ThinkingLevels}
		}
	}
	return nil
}

type optionError struct {
	field   string
	value   string
	allowed []string
}

func (e *optionError) Error() string {
	if len(e.allowed) == 0 {
		return "'" + e.field + "' is not supported by this model (got '" + e.value + "')"
	}
	return "invalid '" + e.field + "' value '" + e.value + "'; allowed: " + strings.Join(e.allowed, ", ")
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v || strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}
