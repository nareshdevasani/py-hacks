package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type GetPackageVersions struct {
	List ObjectList `json:"list"`
}

type ObjectList struct {
	Objects []PackageVersion `json:"objects"`
}

type PackageVersion struct {
	Spec SpecObject `json:"spec"`
}

type SpecObject struct {
	UnresolvedDependencies []PYPIObject `json:"unresolved_dependencies"`
}

type PYPIObject struct {
	Pypi Pypi `json:"pypi"`
}

type Pypi struct {
	Name               string `json:"name"`
	VersionConstraints string `json:"version_constraints"`
}

func CreateRequirementsTxtFromUnresolvedDeps(jsonFile string) {
	packageVersions := getReadPackageVersionsJSON(jsonFile)

	for _, req := range packageVersions.List.Objects[0].Spec.UnresolvedDependencies {
		if req.Pypi.VersionConstraints == "*" {
			fmt.Println(req.Pypi.Name)
		} else {
			fmt.Println(req.Pypi.Name + req.Pypi.VersionConstraints)
		}
	}
}

func getReadPackageVersionsJSON(fileName string) GetPackageVersions {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("error opening the file: " + err.Error())
		return GetPackageVersions{}
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	// we initialize our Users array
	var packageVersions GetPackageVersions
	err = json.Unmarshal(byteValue, &packageVersions)
	if err != nil {
		fmt.Println("error unmarshalling the file: " + err.Error())
		return GetPackageVersions{}
	}
	jsonFile.Close()

	return packageVersions
}
