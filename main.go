/*
	License - generates a human-readable file about third-party licenses
	Copyright (C) 2024-2025  Martijn van der Kleijn

	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"text/template"

	"gopkg.in/yaml.v3"
)

// LicenseStatus holds the status (c/i/w) and optionally a reason.
type LicenseStatus struct {
	Status string `yaml:"status"`
	Reason string `yaml:"reason,omitempty"`
}

// Compatibility is a nested map for license compatibility lookup.
type Compatibility map[string]map[string]LicenseStatus

// Component represents the structure of a component in the input.
type Component struct {
	Name     string `json:"name" xml:"name"`
	Version  string `json:"version" xml:"version"`
	Licenses []struct {
		License struct {
			ID string `json:"id" xml:"id"`
		} `json:"license" xml:"license"`
	} `json:"licenses" xml:"licenses"`
}

// LoadCompatibility reads the YAML file at the given path and unmarshals it
// into the Compatibility structure.
func LoadCompatibility(filePath string) (Compatibility, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var comp Compatibility
	if err := yaml.Unmarshal(data, &comp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return comp, nil
}

// GetLicenseStatus looks up a main license and sub license in the Compatibility map.
// It returns the status (c/i/w), the optional reason, or an error if not found.
func GetLicenseStatus(comp Compatibility, mainLicense, subLicense string) (string, string, error) {
	// Check if the main license is in the map
	subMap, ok := comp[mainLicense]
	if !ok {
		return "", "", fmt.Errorf("no entry for main license: %s", mainLicense)
	}

	// Check if the sub license is in the sub-map
	ls, ok := subMap[subLicense]
	if !ok {
		return "", "", fmt.Errorf("no sub license: %s for main license: %s", subLicense, mainLicense)
	}

	return ls.Status, ls.Reason, nil
}

// BOM represents the overall structure of the input.
type BOM struct {
	XMLName    xml.Name    `xml:"bom"` // Matches the root element, e.g., <bom>
	Metadata   Component   `json:"metadata" xml:"metadata>component"`
	Components []Component `json:"components" xml:"components>component"`
}

// ComponentsByLicense groups components by license ID
type ComponentsByLicense map[string][]Component

func parse(filename string, format string) (BOM, error) {
	var bom BOM
	var err error

	if format == "json" {
		bom, err = parseJSON(filename)
	} else if format == "xml" {
		bom, err = parseXML(filename)
	} else {
		return BOM{}, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		if errors.Is(err, io.EOF) {
			return BOM{}, errors.New("unknown structure or empty file")
		}
		return BOM{}, err
	}

	return bom, nil
}

func parseJSON(filename string) (BOM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return BOM{}, err
	}
	defer file.Close()

	var bom BOM
	if err := json.NewDecoder(file).Decode(&bom); err != nil {
		return BOM{}, err
	}
	return bom, nil
}

func parseXML(filename string) (BOM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return BOM{}, err
	}
	defer file.Close()

	var bom BOM
	if err := xml.NewDecoder(file).Decode(&bom); err != nil {
		return BOM{}, err
	}
	return bom, nil
}

func main() {
	inputFileFlag := flag.String("i", "./sbom.json", "Input file path")
	outputFileFlag := flag.String("o", "./licenses.md", "Output file.")
	formatFlag := flag.String("f", "json", "Input format (json or xml)")
	templateFileFlag := flag.String("t", "./template.txt", "Golang template file to use for output.")
	validateLicenses := flag.Bool("validate", false, "Validate that dependency's licenses are compatible with the main license.")
	flag.Parse()

	var bom BOM
	var comp Compatibility
	var err error

	if *validateLicenses {
		comp, err = LoadCompatibility("compatibility.yaml")
		if err != nil {
			log.Fatal(err)
		}
	}

	bom, err = parse(*inputFileFlag, *formatFlag)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	// Retrieve main application license
	applicationLicense := bom.Metadata.Licenses[len(bom.Metadata.Licenses)-1]
	compissues := make(map[string]LicenseStatus)

	if len(bom.Components) == 0 {
		log.Fatalf("unknown structure or empty components in sbom")
	}

	// Group components by license
	componentsByLicense := make(ComponentsByLicense)
	for _, component := range bom.Components {
		licenseID := "No License"
		if len(component.Licenses) > 0 {
			licenseID = component.Licenses[0].License.ID
			if *validateLicenses {
				status, reason, err := GetLicenseStatus(comp, applicationLicense.License.ID, licenseID)
				if err != nil {
					compissues[component.Licenses[0].License.ID] = LicenseStatus{
						Status: "error",
						Reason: "ERROR - License not found in compatibility matrix.",
					}
				} else if status == "w" || status == "i" {
					compissues[component.Licenses[0].License.ID] = LicenseStatus{
						Status: status,
						Reason: reason,
					}
				}
			}
		}
		componentsByLicense[licenseID] = append(componentsByLicense[licenseID], component)
	}

	var licenseKeys []string
	for license := range componentsByLicense {
		licenseKeys = append(licenseKeys, license)
	}
	sort.Strings(licenseKeys)

	tmpl, err := template.ParseFiles(*templateFileFlag)
	if err != nil {
		log.Fatalf("failed to parse template file: %v", err)
	}

	outputFile, err := os.Create(*outputFileFlag)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Prepare data for the template
	data := struct {
		SortedKeys          []string
		ComponentsByLicense ComponentsByLicense
	}{
		SortedKeys:          licenseKeys,
		ComponentsByLicense: componentsByLicense,
	}

	if err := tmpl.Execute(outputFile, data); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	fmt.Printf("Components information has been written to %s\n", *outputFileFlag)

	// If there are any issues, exit with non-zero status
	if *validateLicenses && len(compissues) > 0 {
		fmt.Println("Found license compatibility issues:")
		for key, issue := range compissues {
			fmt.Printf("- %s: %s)\n", key, issue.Reason)
		}
		os.Exit(1) // Non-zero exit code for build pipeline
	}

	// Otherwise, exit normally
	fmt.Println("No issues found.")
	os.Exit(0)
}
