package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Package struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	var packages []string

	// Query the Packagist API multiple times to get a total of 1000 packages
	for i := 1; i <= 10; i++ {
		// Send HTTP GET request to the Packagist API
		url := fmt.Sprintf("https://packagist.org/explore/popular.json?page=%d&per_page=100", i)
		fmt.Println(url)
		response, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		// Unmarshal the JSON response into a struct
		var data struct {
			Packages []Package `json:"packages"`
			Next     string    `json:"next"`
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}

		// Append the packages from the current response to the overall packages list
		for _, p := range data.Packages {
			packages = append(packages, p.Name)
		}

		// If there is no "next" URL, break out of the loop
		if data.Next == "" {
			break
		}
	}

	// Create a slice of package names
	var packageNames []string
	packageNames = append(packageNames, packages...)

	// Create the composer.json file
	content := `{
	"require": {
		` + strings.Join(getRequireStatements(packageNames), ",\n\t\t") + `
	}
}`

	// Write the content to the composer.json file
	err := ioutil.WriteFile("composer.json", []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	// Success message
	println("composer.json file created successfully.")
}

// Helper function to format package names as require statements
func getRequireStatements(packages []string) []string {
	var requireStatements []string
	for _, pkg := range packages {
		requireStatements = append(requireStatements, `"`+pkg+`": "*"`)
	}
	return requireStatements
}
