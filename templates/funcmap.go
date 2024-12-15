package templates

import (
	"html/template"

	"github.com/patrickward/hop/templates/funcmap/collections"
	"github.com/patrickward/hop/templates/funcmap/conversions"
	"github.com/patrickward/hop/templates/funcmap/debug"
	"github.com/patrickward/hop/templates/funcmap/html"
	"github.com/patrickward/hop/templates/funcmap/maps"
	"github.com/patrickward/hop/templates/funcmap/numbers"
	"github.com/patrickward/hop/templates/funcmap/slices"
	"github.com/patrickward/hop/templates/funcmap/strings"
	"github.com/patrickward/hop/templates/funcmap/time"
	"github.com/patrickward/hop/templates/funcmap/values"
)

// MergeFuncMaps merges the provided function maps into a single function map.
func MergeFuncMaps(maps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}
	return result
}

// cachedFuncMap holds the cached function map for the templates package.
var cachedFuncMap template.FuncMap

// FuncMap returns the complete function map for the templates package.
func FuncMap() template.FuncMap {
	if cachedFuncMap != nil {
		return cachedFuncMap
	}

	cachedFuncMap = MergeFuncMaps(
		collections.FuncMap(),
		conversions.FuncMap(),
		debug.FuncMap(),
		html.FuncMap(),
		maps.FuncMap(),
		numbers.FuncMap(),
		slices.FuncMap(),
		strings.FuncMap(),
		time.FuncMap(),
		values.FuncMap(),
	)

	return cachedFuncMap
}
