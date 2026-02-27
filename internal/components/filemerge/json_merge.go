package filemerge

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func MergeJSONObjects(baseJSON []byte, overlayJSON []byte) ([]byte, error) {
	base := map[string]any{}
	if len(bytes.TrimSpace(baseJSON)) > 0 {
		if err := json.Unmarshal(baseJSON, &base); err != nil {
			return nil, fmt.Errorf("unmarshal base json: %w", err)
		}
	}

	overlay := map[string]any{}
	if len(bytes.TrimSpace(overlayJSON)) > 0 {
		if err := json.Unmarshal(overlayJSON, &overlay); err != nil {
			return nil, fmt.Errorf("unmarshal overlay json: %w", err)
		}
	}

	merged := mergeObjects(base, overlay)
	encoded, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal merged json: %w", err)
	}

	return append(encoded, '\n'), nil
}

func mergeObjects(base map[string]any, overlay map[string]any) map[string]any {
	result := make(map[string]any, len(base)+len(overlay))
	for key, value := range base {
		result[key] = value
	}

	for key, overlayValue := range overlay {
		baseValue, ok := result[key]
		if !ok {
			result[key] = overlayValue
			continue
		}

		baseMap, baseIsMap := baseValue.(map[string]any)
		overlayMap, overlayIsMap := overlayValue.(map[string]any)
		if baseIsMap && overlayIsMap {
			result[key] = mergeObjects(baseMap, overlayMap)
			continue
		}

		result[key] = overlayValue
	}

	return result
}
