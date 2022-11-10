package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
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
	Branch    string
	CloneURLs []string
}

type metadata struct {
	Info info
}

type info struct {
	Version     string
	ProjectURLs urls `json:"project_urls"`
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
	readPackages(24)
}

func readPackages(query int) {
	resp, err := http.Get("https://pypi.org/simple/")
	if err != nil {
		log.Fatalln(err)
	}

	fileName := "0.json"
	if query > 0 {
		fileName = string(rune('A'-1+query)) + ".json"
	}
	fmt.Println("File requested: " + fileName)

	total := 0
	count := make([]int, 27)
	scanner := bufio.NewScanner(resp.Body)
	modules := make([]Module, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "href=\"/simple/") {
			continue
		}

		packageName := filepath.Base(strings.Split(line, "\"")[1])
		first := []rune(packageName)[0]
		index := 0
		if unicode.IsDigit(first) {
			count[0] += 1
		} else {
			index = int(first - 'a' + 1)
			count[index] += 1
		}

		if index == query {
			module := getDetails(packageName)
			if len(module.Name) > 0 {
				modules = append(modules, module)
			}
		}

		if len(modules) > 0 && len(modules)%500 == 0 {
			fmt.Println("Completed " + fmt.Sprint(len(modules)))
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
		fileName = "data/" + fileName
		file, _ := json.MarshalIndent(modules, "", " ")
		_ = ioutil.WriteFile(fileName, file, 0644)
		fmt.Println("Successfully written " + fmt.Sprint(len(modules)) + " modules into " + fileName)
	}
}

func getDetails(pkgName string) Module {
	time.Sleep(250 * time.Millisecond)

	resp, err := http.Get("https://pypi.python.org/pypi/" + pkgName + "/json")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Println("Package not found... " + pkgName)
		return Module{}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Package read of response failed... " + pkgName)
		return Module{}
	}

	var response metadata

	if err := json.Unmarshal(b, &response); err != nil {
		fmt.Println("Unmarshal of response failed... " + pkgName)
		return Module{}
	}

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
