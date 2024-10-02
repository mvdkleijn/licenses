/*
	License - generates a human-readable file about third-party licenses
	Copyright (C) 2024  Martijn van der Kleijn

	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"text/template"
)

// Component represents the structure of a component in the input JSON.
type Component struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Licenses []struct {
		License struct {
			ID string `json:"id"`
		} `json:"license"`
	} `json:"licenses"`
}

// BOM represents the overall structure of the input JSON.
type BOM struct {
	Components []Component `json:"components"`
}

// ComponentsByLicense groups components by license ID
type ComponentsByLicense map[string][]Component

func main() {
	// Define command-line flags
	inputFileFlag := flag.String("i", "./sbom.json", "SBOM input file in CycloneDX json format.")
	outputFileFlag := flag.String("o", "./licenses.md", "Output file.")
	templateFileFlag := flag.String("t", "./template.txt", "Golang template file to use for output.")

	flag.Parse()

	file, err := os.Open(*inputFileFlag)
	if err != nil {
		log.Fatalf("failed to open input JSON file: %v", err)
	}
	defer file.Close()

	var bom BOM
	if err := json.NewDecoder(file).Decode(&bom); err != nil {
		log.Fatalf("failed to decode JSON: %v", err)
	}

	// Group components by license
	componentsByLicense := make(ComponentsByLicense)
	for _, component := range bom.Components {
		licenseID := "No License"
		if len(component.Licenses) > 0 {
			licenseID = component.Licenses[0].License.ID
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
}
