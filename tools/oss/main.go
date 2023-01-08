package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Module struct {
	Name      string
	Version   string
	Branch    string   `json:"Branch,omitempty"`
	CloneURLs []string `json:"CloneURLs,omitempty"`
}

type metadata struct {
	Info info
}

type info struct {
	Version      string
	ProjectURLs  urls     `json:"project_urls"`
	RequiresDist []string `json:"requires_dist"`
}

type urls struct {
	HomePage     string `json:"Homepage"`
	Repository   string `json:"Repository"`
	Source       string `json:"Source"`
	Code         string `json:"Code"`
	GitHub       string `json:"GitHub"`
	SourceCode   string `json:"Source Code"`
	IssueTracker string `json:"Issue Tracker"`
}

func main() {
	// 19 - S
	// 10 - J
	// 1 - A
	// readPackages(1, "aws-solutions-constructs-aws-dynamodb-stream-lambda-elasticsearch-kibana")

	// parse the data files and export dep information
	// parseAndPersistDeps([]string{"0.json"}, "0Deps.txt")

	parseAndPersistDeps([]string{"C3.json", "C4.json"}, "CDeps2.txt")
}

func readPackages(query int, startingName string) {
	resp, err := http.Get("https://pypi.org/simple/")
	if err != nil {
		log.Fatalln(err)
	}

	fileName := "0.json"
	if query > 0 {
		fileName = string(rune('A'-1+query)) + ".json"
	}
	fileName = "data/" + fileName
	fmt.Println("File requested: " + fileName)

	start := time.Now()
	total := 0
	count := make([]int, 27)
	scanner := bufio.NewScanner(resp.Body)
	modules := make([]Module, 0)
	started := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "href=\"/simple/") {
			continue
		}

		packageName := filepath.Base(strings.Split(line, "\"")[1])
		if packageName == startingName {
			started = true
		}
		first := []rune(packageName)[0]
		index := 0
		if unicode.IsDigit(first) {
			count[0] += 1
		} else {
			index = int(first - 'a' + 1)
			count[index] += 1
		}

		if index == query && started {
			module := getDetails(packageName)
			if len(module.Name) > 0 {
				modules = append(modules, module)
			} else {
				fmt.Println(" ..." + fmt.Sprint(len(modules)))
			}
		}

		if len(modules) == 500 {
			existing := getExisting(fileName)
			existing = append(existing, modules...)
			file, _ := json.MarshalIndent(existing, "", " ")
			_ = ioutil.WriteFile(fileName, file, 0644)
			fmt.Println("Completed "+fmt.Sprint(len(existing))+". Took ", time.Since(start))
			//start = time.Now()
			modules = make([]Module, 0)
		}
		total += 1
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error in scanner " + err.Error())
	}

	for idx, cnt := range count {
		if idx == 0 {
			fmt.Print("Count with digits: ")
			fmt.Println(cnt)
			continue
		}
		fmt.Print("Count with " + string(rune('A'-1+idx)) + ": ")
		fmt.Println(cnt)
	}
	fmt.Println(total)

	// persist
	if len(modules) > 0 {
		existing := getExisting(fileName)
		existing = append(existing, modules...)
		file, _ := json.MarshalIndent(existing, "", " ")
		_ = ioutil.WriteFile(fileName, file, 0644)
		fmt.Println("Successfully written " + fmt.Sprint(len(existing)) + " modules into " + fileName)
	}
}

func getExisting(fileName string) []Module {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("error opening the file: " + err.Error())
		return nil
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	// we initialize our Users array
	var existing []Module
	err = json.Unmarshal(byteValue, &existing)
	if err != nil {
		fmt.Println("error unmarshalling the file: " + err.Error())
		return nil
	}
	jsonFile.Close()

	return existing
}

func queryDetails(pkgName string, version string) metadata {
	time.Sleep(100 * time.Millisecond)

	var resp *http.Response
	var err error
	if version == "" {
		resp, err = http.Get("https://pypi.python.org/pypi/" + pkgName + "/json")
	} else {
		resp, err = http.Get("https://pypi.python.org/pypi/" + pkgName + "/" + version + "/json")
	}

	if err != nil {
		fmt.Println("ERRRRRRRRRRRRRR")
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Print("Package not found... " + pkgName)
		return metadata{}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Package read of response failed... " + pkgName)
		return metadata{}
	}

	var response metadata

	if err := json.Unmarshal(b, &response); err != nil {
		fmt.Println("Unmarshal of response failed... " + pkgName)
		return metadata{}
	}
	return response
}

func getDetails(pkgName string) Module {
	response := queryDetails(pkgName, "")
	result := Module{
		Name:    pkgName,
		Version: response.Info.Version,
	}

	urls := make([]string, 0)
	urls = addToList(response.Info.ProjectURLs.SourceCode, urls)
	urls = addToList(response.Info.ProjectURLs.Code, urls)
	urls = addToList(response.Info.ProjectURLs.Repository, urls)
	urls = addToList(response.Info.ProjectURLs.Source, urls)
	urls = addToList(response.Info.ProjectURLs.GitHub, urls)
	urls = addToList(response.Info.ProjectURLs.HomePage, urls)
	urls = addToList(response.Info.ProjectURLs.IssueTracker, urls)
	result.CloneURLs = urls

	return result
}

func addToList(url string, urls []string) []string {
	url = strings.TrimSpace(url)
	result := urls
	if len(url) > 0 && strings.Contains(url, "github.com") {
		result = append(urls, url)
	}

	return result
}

func isValidRepo(url string) bool {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{url},
	})

	if _, err := remote.List(&git.ListOptions{}); err != nil {
		return false
	}
	return true
}

func parseAndPersistDeps(files []string, opFile string) {
	lines := make([]string, 0)
	for _, file := range files {
		existing := getExisting("data/" + file)
		for _, module := range existing {
			metadata := queryDetails(module.Name, module.Version)
			if len(metadata.Info.RequiresDist) == 0 {
				continue
			}

			deps := make([]string, 0)
			for _, dist := range metadata.Info.RequiresDist {
				if strings.Contains(dist, "extra == ") {
					continue
				}
				deps = append(deps, dist)
			}
			if len(deps) == 0 {
				continue
			}
			newLine := fmt.Sprintf("%s,%s,%d,%s", module.Name, module.Version, len(deps), strings.Join(deps, "$"))
			// fmt.Println(newLine)
			lines = append(lines, newLine)
			if len(lines)%1000 == 0 {
				fmt.Print("Count is: ")
				fmt.Println(len(lines))
			}
		}
		fmt.Println(file + " completed......")
	}

	sort.Slice(lines, func(i, j int) bool {
		num1, _ := strconv.Atoi(strings.Split(lines[i], ",")[2])
		num2, _ := strconv.Atoi(strings.Split(lines[j], ",")[2])

		return num1 > num2
	})
	file, err := os.OpenFile(opFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range lines {
		fmt.Println(data)
		_, _ = datawriter.WriteString(data + "\n")
	}

	datawriter.Flush()
	file.Close()
}
