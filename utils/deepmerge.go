package utils

// DeepMerge recursively merges src into dst; used for template data
func DeepMerge(dst *map[string]any, src map[string]any) {
	for k, srcVal := range src {
		dstMap := *dst
		dstVal, exists := dstMap[k]
		if !exists {
			dstMap[k] = srcVal
			continue
		}

		// If both values are maps, merge them
		srcMap, srcIsMap := srcVal.(map[string]any)
		dstMap2, dstIsMap := dstVal.(map[string]any)
		if srcIsMap && dstIsMap {
			// Create new map if destination is nil
			if dstMap2 == nil {
				dstMap2 = make(map[string]any)
				dstMap[k] = dstMap2
			}
			DeepMerge(&dstMap2, srcMap)
			continue
		}

		// Otherwise, overwrite with new value
		dstMap[k] = srcVal
	}
}
