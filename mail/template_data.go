package mail

import (
	"fmt"
	"time"
)

type TemplateData map[string]any

func NewTemplateData(cfg *Config) TemplateData {
	now := time.Now()

	data := TemplateData{
		"BaseURL":          cfg.BaseURL,
		"Copyright":        fmt.Sprintf("© %d %s. All rights reserved", now.Year(), cfg.CompanyName),
		"CompanyName":      cfg.CompanyName,
		"CompanyAddress1":  cfg.CompanyAddress1,
		"CompanyAddress2":  cfg.CompanyAddress2,
		"LogoURL":          cfg.LogoURL,
		"SupportEmail":     cfg.SupportEmail,
		"SupportPhone":     cfg.SupportPhone,
		"WebsiteName":      cfg.WebsiteName,
		"WebsiteURL":       cfg.WebsiteURL,
		"DateYear":         now.Year(),
		"DateTime":         now.Format("2006-01-02 15:04:05"),
		"DateDisplay":      now.Format("January 2, 2006"),
		"SiteLinks":        cfg.SiteLinks,
		"SocialMediaLinks": cfg.SocialMediaLinks,
	}

	return data
}

// Merge combines the current TemplateData with the provided data map.
func (td TemplateData) Merge(data map[string]any) TemplateData {
	merged := TemplateData{}

	// Copy the current TemplateData
	for key, value := range td {
		merged[key] = value
	}

	// Copy the provided data, allowing it to overwrite existing keys
	for key, value := range data {
		merged[key] = value
	}

	return merged
}

// MergeKeys merges data into the template data,
// combining maps for existing keys instead of overwriting the entire map. This allows for more granular updates.
// It can merge maps for existing keys as well as add new keys.
func (td TemplateData) MergeKeys(data map[string]any) TemplateData {
	merged := make(TemplateData)

	// Copy existing data
	for k, v := range td {
		merged[k] = v
	}

	// Merge new data, combining maps for existing keys
	for k, v := range data {
		if existingMap, ok := merged[k].(map[string]any); ok {
			if newMap, ok := v.(map[string]any); ok {
				// Combine maps for existing keys
				mergedMap := make(map[string]any)
				for ek, ev := range existingMap {
					mergedMap[ek] = ev
				}
				for nk, nv := range newMap {
					mergedMap[nk] = nv
				}
				merged[k] = mergedMap
				continue
			}
		}
		merged[k] = v
	}

	return merged
}
