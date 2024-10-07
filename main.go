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
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"text/template"
)

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

// BOM represents the overall structure of the input.
type BOM struct {
	XMLName    xml.Name    `xml:"bom"` // Matches the root element, e.g., <bom>
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
	flag.Parse()

	var bom BOM

	var err error

	bom, err = parse(*inputFileFlag, *formatFlag)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	if len(bom.Components) == 0 {
		log.Fatalf("unknown structure or empty components in sbom")
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
