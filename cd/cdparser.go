package cd

import (
	"encoding/xml"
	"fmt"

	"github.com/coppersoft-dev/ecp-metrics-server/cd/types"
)

func ParseCDContent(content []byte) (types.Components, error) {
	// Unmarshal the XML data
	var components types.Components
	if err := xml.Unmarshal(content, &components); err != nil {
		return components, fmt.Errorf("failed unmarshaling XML: %w", err)
	}

	return components, nil
}
