package cd

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Components is the top-level structure
type Components struct {
	XMLName       xml.Name      `xml:"http://mades.entsoe.eu/componentDirectory components"`
	ComponentList ComponentList `xml:"componentList"`
	Metadata      Metadata      `xml:"metadata"`
}

// ComponentList contains the list of components
type ComponentList struct {
	Brokers              []Broker             `xml:"broker"`
	Endpoints            []Endpoint           `xml:"endpoint"`
	ComponentDirectories []ComponentDirectory `xml:"componentDirectory"`
}

type Component struct {
	MADESImplementation MADESImplementation `xml:"madesImplementation"`
}

type ComponentDirectory struct {
	Component
	Organization          string       `xml:"organization"`
	Person                string       `xml:"person"`
	Email                 string       `xml:"email"`
	Code                  string       `xml:"code"`
	Type                  string       `xml:"type"`
	Networks              Networks     `xml:"networks"`
	URLs                  URLs         `xml:"urls"`
	Certificates          Certificates `xml:"certificates"`
	CreationTimestamp     time.Time    `xml:"creationTimestamp"`
	ModificationTimestamp time.Time    `xml:"modificationTimestamp"`
	ComponentDirectory    string       `xml:"componentDirectory"`
}

// Broker represents a broker component
type Broker struct {
	Component
	Organization          string       `xml:"organization"`
	Person                string       `xml:"person"`
	Email                 string       `xml:"email"`
	Code                  string       `xml:"code"`
	Type                  string       `xml:"type"`
	Networks              Networks     `xml:"networks"`
	URLs                  URLs         `xml:"urls"`
	Certificates          Certificates `xml:"certificates"`
	CreationTimestamp     time.Time    `xml:"creationTimestamp"`
	ModificationTimestamp time.Time    `xml:"modificationTimestamp"`
	ComponentDirectory    string       `xml:"componentDirectory"`
	Restriction           Restriction  `xml:"restriction"`
}

// Endpoint represents an endpoint component
type Endpoint struct {
	Component
	Organization          string       `xml:"organization"`
	Person                string       `xml:"person"`
	Email                 string       `xml:"email"`
	Code                  string       `xml:"code"`
	Type                  string       `xml:"type"`
	Networks              Networks     `xml:"networks"`
	URLs                  URLs         `xml:"urls"`
	Certificates          Certificates `xml:"certificates"`
	CreationTimestamp     time.Time    `xml:"creationTimestamp"`
	ModificationTimestamp time.Time    `xml:"modificationTimestamp"`
	ComponentDirectory    string       `xml:"componentDirectory"`
	Paths                 Paths        `xml:"paths"`
}

// Networks contains a list of network names
type Networks struct {
	Network []string `xml:"network"`
}

// URLs contains a list of URL entries
type URLs struct {
	URL []URL `xml:"url"`
}

// URL represents a network URL
type URL struct {
	Network string `xml:"network,attr"`
	Value   string `xml:",chardata"`
}

// Certificates contains a list of certificates
type Certificates struct {
	Certificate []Certificate `xml:"certificate"`
}

// Certificate represents a certificate entry
type Certificate struct {
	CertificateID string    `xml:"certificateID"`
	Type          string    `xml:"type"`
	Certificate   string    `xml:"certificate"`
	ValidFrom     time.Time `xml:"validFrom"`
	ValidTo       time.Time `xml:"validTo"`
}

// MADESImplementation contains MADES implementation details
type MADESImplementation struct {
	Name               string `xml:"name,attr"`
	Version            string `xml:"version,attr"`
	CompatibleVersions string `xml:"compatibleVersions,attr"`
	MADESVersion       string `xml:"madesVersion,attr"`
}

// Restriction defines restrictions for a component
type Restriction struct {
	Components   RestrictionComponents `xml:"components"`
	MessageTypes MessageTypes          `xml:"messageTypes"`
}

// RestrictionComponents contains component restrictions
type RestrictionComponents struct {
	Component []string `xml:"component"`
}

// MessageTypes contains message type restrictions
type MessageTypes struct {
	MessageType []string `xml:"messageType"`
}

// Paths contains routing path information
type Paths struct {
	Path []Path `xml:"path"`
}

// Path represents a routing path entry
type Path struct {
	SenderComponent SenderComponent `xml:"senderComponent"`
	MessageType     string          `xml:"messageType"`
	Path            string          `xml:"path"`
	ValidFrom       time.Time       `xml:"validFrom"`
}

// SenderComponent contains sender component details
type SenderComponent struct {
	Component []string `xml:"component"`
}

// Metadata contains metadata about the component directory
type Metadata struct {
	ComponentDirectoryMetadata ComponentDirectoryMetadata `xml:"componentDirectoryMetadata"`
}

// ComponentDirectoryMetadata contains directory metadata details
type ComponentDirectoryMetadata struct {
	ComponentDirectory string `xml:"componentDirectory"`
	TTL                int    `xml:"ttl"`
	ContentID          int    `xml:"contentID"`
}

func ParseCDContent(content []byte) (Components, error) {
	// Unmarshal the XML data
	var components Components
	if err := xml.Unmarshal(content, &components); err != nil {
		return components, fmt.Errorf("failed unmarshaling XML: %w", err)
	}

	return components, nil
}
