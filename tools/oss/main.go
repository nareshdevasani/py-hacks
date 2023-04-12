package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	git "github.com/go-git/go-git/v5"
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
	Licence      string   `json:"license"`
	Classifiers  []string `json:"classifiers"`
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

	// start := time.Now()
	// parseAndPersistDeps([]string{"T3.json", "T4.json"}, "TDeps2.txt")
	// fmt.Println(time.Since(start))

	// pkgList := getListOfPHPPackages()

	// start := time.Now()
	// mstart := time.Now()
	// var prevCount int
	// var mcount int
	// var minElapsed int
	// for index, pkgName := range pkgList {
	// 	queryPHPPackage(pkgName)
	// 	if time.Since(start) >= time.Second {
	// 		fmt.Println("Total requests per second: " + fmt.Sprint(index-prevCount))
	// 		prevCount = index
	// 		start = time.Now()
	// 	}
	// 	if time.Since(mstart) >= time.Minute {
	// 		minElapsed++
	// 		fmt.Println("Total requests per minute: " + fmt.Sprint(minElapsed) + " : " + fmt.Sprint(index-mcount))
	// 		mcount = index
	// 		mstart = time.Now()
	// 	}
	// 	//time.Sleep(100 * time.Millisecond)
	// }

	// printPypiLicenses("")
	// utils.CreateRequirementsTxtFromUnresolvedDeps("/Users/nareshdevasani/endor/python/pv.json")

	// Read rubygems db dump
	readDbDump("/Users/nareshdevasani/Downloads/public_postgresql/databases/PostgreSQL.sql")
	// getDownloadURL()
	// err := isValidRemoteGitURL("https://github.com/googleapis/google-api-ruby-client")
	// if err != nil {
	// 	fmt.Println("Invalid GIT URL ****: " + err.Error())
	// } else {
	// 	fmt.Println("Valid GIT URL")
	// }
}

func printPypiLicenses(start string) {
	pkgs := getListOfPypiPackages()
	var started bool
	var count int
	for _, p := range pkgs {
		if start == "" || start == p {
			started = true
		}
		if started {
			queryPypiPackage(p)
			count++
		}
		if count > 0 && count%1000 == 0 {
			time.Sleep(10 * time.Second)
		}
	}
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
		fmt.Println("ERRRRRRRRRRRRRR --- Retrying")
		// log.Fatalln(err)
		return queryDetails(pkgName, version)
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
		totalModules := len(existing)
		for mno, module := range existing {
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
			if len(lines)%60 == 0 {
				fmt.Printf("%d / %d : ", mno, totalModules)
				fmt.Print(time.Now())
				fmt.Print(": Count is: ")
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

type PHPPackageNames struct {
	Names []string `json:"packageNames"`
}

type PHPPackage struct {
	Packages PackageVersions `json:"packages"`
}

type PackageVersions struct {
	PkgVersions map[string][]interface{}
}

func getListOfPHPPackages() []string {
	// var resp *http.Response
	// var err error
	resp, err := http.Get("https://packagist.org/packages/list.json")

	if err != nil {
		fmt.Println("ERRRRRRRRRRRRRR --- fetching list of php packages...")
		return []string{}
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Print("Error and code is invalid ")
		return []string{}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Package read of response failed... " + err.Error())
		return []string{}
	}

	var response PHPPackageNames

	if err := json.Unmarshal(b, &response); err != nil {
		fmt.Println("Unmarshal of response failed... " + err.Error())
		return response.Names
	}
	fmt.Println("******* count of packages: " + fmt.Sprint(len(response.Names)))
	return response.Names
}

func queryPHPPackage(pkgName string) {
	// var resp *http.Response
	// var err error
	resp, err := http.Get("https://repo.packagist.org/p2/" + pkgName + ".json")

	if err != nil {
		fmt.Println("ERRRRRRRRRRRRRR --- fetching each php package..." + pkgName + ", Err:" + err.Error())
		fmt.Println(resp)
		time.Sleep(5 * time.Second)
		queryPHPPackage(pkgName)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Print("Error and code is invalid " + pkgName)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Package read of response failed... " + pkgName + ", Err: " + err.Error())
		return
	}

	var response PHPPackage

	if err := json.Unmarshal(b, &response); err != nil {
		fmt.Println("Unmarshal of response failed... " + pkgName + ", Err: " + err.Error())
		return
	}
	// fmt.Println(response.Packages.PkgVersions)
}

func getListOfPypiPackages() []string {
	resp, err := http.Get("https://pypi.org/simple/")
	if err != nil {
		log.Fatalln(err)
	}

	scanner := bufio.NewScanner(resp.Body)
	packages := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "href=\"/simple/") {
			continue
		}

		packageName := filepath.Base(strings.Split(line, "\"")[1])
		packages = append(packages, packageName)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("error in scanner " + err.Error())
	}
	fmt.Println("******* count of packages: " + fmt.Sprint(len(packages)))
	return packages
}

func queryPypiPackage(pkgName string) {
	metadata := queryDetails(pkgName, "")

	fmt.Print(pkgName)
	fmt.Print("   ****   ")
	fmt.Print(metadata.Info.Licence)
	fmt.Print("   ****   ")
	for _, s := range metadata.Info.Classifiers {
		if strings.HasPrefix(s, "License") {
			parts := strings.Split(s, "::")
			fmt.Print(parts[len(parts)-1])
		}
	}
	fmt.Println()
}

type gemVersion struct {
	// sha
	checksum string
	// field3
	releaseDate time.Time
	// field7
	licenses []string
	// version number.
	name string
}

// gemMetadata holds the gems read from the DB dump.
type gemMetadata struct {
	// gem name.
	name string
	// source clone URL.
	repository string

	versions []*gemVersion
}

func readDbDump(path string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Println()
	inFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	buf := make([]byte, 256*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)

	// id to name map.
	lineCnt := 0
	nameLookup := make(map[string]string)
	versionLookup := make(map[string]*gemMetadata)
	linksetLookup := make(map[string]string)
	versionsCount := 0
	rubygemsCount := 0
	linksetCount := 0
	rubygemsStarted := false
	versionsStarted := false
	linksetStarted := false
	lastLine := ""
	maxLen := 0
	longLine := ""
	max2Len := 0
	long2Line := ""
	max3Len := 0
	long3Line := ""
	emptyURLCount := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lastLine = line
		if maxLen < len(line) {
			max3Len = max2Len
			long3Line = long2Line
			max2Len = maxLen
			long2Line = longLine
			maxLen = len(line)
			longLine = line
		} else if max2Len < len(line) {
			max3Len = max2Len
			long3Line = long2Line
			max2Len = len(line)
			long2Line = line
		} else if max3Len < len(line) {
			max3Len = len(line)
			long3Line = line
		}
		lineCnt++

		if lineCnt%500000 == 0 {
			fmt.Print("LINE# ")
			fmt.Println(lineCnt)
		}
		if line == "" || line == "\\." {
			rubygemsStarted = false
			versionsStarted = false
			linksetStarted = false
			continue
		}

		if strings.HasPrefix(line, "COPY public.linksets (id") {
			runtime.ReadMemStats(&m)
			fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
			fmt.Println()

			linksetStarted = true
			continue
		}

		if strings.HasPrefix(line, "COPY public.rubygems (id") {
			runtime.ReadMemStats(&m)
			fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
			fmt.Println()

			rubygemsStarted = true
			continue
		}

		if strings.HasPrefix(line, "COPY public.versions (id") {
			runtime.ReadMemStats(&m)
			fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
			fmt.Println()

			versionsStarted = true
			continue
		}

		if linksetStarted {
			linksetCount++
			row := strings.Split(line, "\t")
			if len(row) < 2 {
				fmt.Println("ERROR SPLIT ON TAB CHAR - rubygems table: " + line)
				continue
			}

			if row[6] != "\\N" {
				linksetLookup[row[1]] = row[6]
			}

			// when code is empty, use home.
			// TODO validate home url.
			// TODO remove /tree/ and extract source_ref.
			if linksetLookup[row[1]] == "" {
				linksetLookup[row[1]] = row[2]
				if strings.Contains(row[2], "gitlab") && strings.Contains(row[2], "/tree") {
					fmt.Println(row[2] + " -> " + row[1])
				}
			}
			continue
		}

		if rubygemsStarted {
			rubygemsCount++
			row := strings.Split(line, "\t")
			if len(row) < 2 {
				fmt.Println("ERROR SPLIT ON TAB CHAR - rubygems table: " + line)
				continue
			}

			nameLookup[row[0]] = row[1]
			continue
		}

		if versionsStarted {
			versionsCount++
			row := strings.Split(line, "\t")
			if len(row) < 25 {
				continue
			}
			gem, ok := versionLookup[nameLookup[row[4]]]
			if !ok {
				repository := getRepository(row[20], linksetLookup[row[4]])
				gem = &gemMetadata{
					name:       nameLookup[row[4]],
					repository: repository,
				}
				versionLookup[nameLookup[row[4]]] = gem
			} else {
				if gem.repository == "" {
					repository := getRepository(row[20], linksetLookup[row[4]])
					gem.repository = repository
				}
			}

			relTime, _ := time.Parse("2006-01-02 15:04:05.123456", row[9])
			sha := getSha(row[19])
			// if sha == "" {
			// 	fmt.Println("EMPTY SHA for: " + row[14])
			// }
			licenses := getLicenses(row[15])
			// if len(licenses) > 1 {
			// 	fmt.Println(fmt.Sprintf("******** More than one license : %s, %d: %v", row[14], len(licenses), licenses))
			// }
			gem.versions = append(gem.versions, &gemVersion{
				name:        row[3],
				checksum:    sha,
				releaseDate: relTime,
				licenses:    licenses,
			})
			continue
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Println("Last line ************: " + lastLine)
		panic(err)
	}

	// print first 10 records
	fmt.Println(maxLen)
	fmt.Println(longLine[:100])
	fmt.Println(max2Len)
	fmt.Println(long2Line[:100])
	fmt.Println(max3Len)
	fmt.Println(long3Line[:200])
	fmt.Println("Printing records *******")
	fmt.Print("Size of rubygems: ")
	fmt.Println(len(nameLookup))
	fmt.Print("Size of versions: ")
	fmt.Println(len(versionLookup))
	for _, v := range versionLookup {
		if v.repository == "" {
			emptyURLCount++
		}
	}

	fmt.Println(fmt.Sprintf("Count of gems with URLs: %d", (len(versionLookup) - emptyURLCount)))

	cnt := 0
	for k, v := range nameLookup {
		fmt.Print(k + " -> ")
		fmt.Println(v)
		cnt++
		if cnt == 20 {
			break
		}
	}

	cnt = 0
	for k, v := range versionLookup {
		fmt.Print(k + " -> " + v.name + " :URL: ")
		fmt.Println(v.repository)
		for _, ver := range v.versions {
			fmt.Println(ver.name + " -> " + ver.checksum + " -> " + fmt.Sprintf("%d: %v", len(ver.licenses), ver.licenses))
		}
		cnt++
		if cnt == 20 {
			break
		}
	}

	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
}

// COPY public.versions (
//	0		1							2							3		4			5					6							7						8			9							10			11			12			13		14				15				16		17				18						19												20			21							22			23									24						25			26					27
// id,		authors, 					description, 				number, rubygem_id, built_at, 			updated_at, 				summary, 				platform, 	created_at, 				indexed, 	prerelease, "position", latest, full_name, 		licenses, 		size, 	requirements, 	required_ruby_version, 	sha256, 										metadata, 	required_rubygems_version, 	yanked_at, 	info_checksum, 						yanked_info_checksum, 	pusher_id, 	canonical_number, 	cert_chain) FROM stdin;
// 634090	Austin G. Davis-Richardson	Like cowsay but with cats	0.2.1	95809		2014-11-17 00:00:00	2020-12-14 16:00:43.630801	Cats in your terminal	ruby		2014-11-17 01:10:06.711091	t			f			12			f		catsay-0.2.1	---\n- MIT\n	10240	--- []\n		>= 0					5M33+zp2ZnZNfDsGUxvziMdZfvWnzwTvVLgW8BIqG0A=				>= 0						\N			1848f2d9d098df0129209f628c0b0115	\N						\N			0.2.1				\N
// 1219495	Ryan Grove					Crass is a pure Ruby CSS.	1.0.6	85152		2020-01-12 00:00:00	2020-12-14 17:09:28.715023	CSS Level 3 spec.		ruby		2020-01-12 22:24:38.894954	t			f			0			t		crass-1.0.6		---\n- MIT\n	18432	--- []\n		>= 1.9.2				3FFgIqVuezsVYJmryBttKwjqHtEmdqx6VldhfwEr1F0=	"changelog_uri"=>"https://github.com/rgrove/crass/blob/v1.0.6/HISTORY.md", "bug_tracker_uri"=>"https://github.com/rgrove/crass/issues", "source_code_uri"=>"https://github.com/rgrove/crass/tree/v1.0.6", "documentation_uri"=>"https://www.rubydoc.info/gems/crass/1.0.6"	>= 0	\N	5758dfcc0400f10c45ba4ab4060f2d33	\N	763	1.0.6	\N

func getRepository(metadata string, urlFromLinkset string) string {
	parts := strings.Split(metadata, ",")
	for _, u := range parts {
		u = strings.TrimSpace(u)
		if strings.HasPrefix(u, "\"source_code_uri\"=>\"") {
			srcUrl := strings.TrimPrefix(u, "\"source_code_uri\"=>\"")
			return strings.TrimSuffix(srcUrl, "\"")
		}
	}

	return urlFromLinkset
}

func getSha(sha256 string) string {
	if sha256 == "\\N" {
		return ""
	}
	base64Decoded, err := base64.StdEncoding.DecodeString(sha256)
	if err != nil {
		fmt.Println(sha256)
		panic(err)
	}

	return hex.EncodeToString(base64Decoded)
}

func getLicenses(license string) []string {
	if license == "\\N" {
		return nil
	}
	parts := strings.Split(license, "\\n")
	licenses := make([]string, 0)
	for _, l := range parts {
		l = strings.TrimSpace(strings.TrimLeft(l, "-"))
		l = strings.TrimSpace(strings.TrimLeft(l, "."))
		if l == "" || l == "[]" {
			continue
		}
		licenses = append(licenses, l)
	}
	return licenses
}

const baseURL = "https://s3-us-west-2.amazonaws.com/rubygems-dumps/"
const prefix = "production/public_postgresql"

func getDownloadURL() string {
	resp, err := http.Get(fmt.Sprintf("%s?prefix=%s", baseURL, prefix))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("<Key>(.*?)</Key>")
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) > 0 {
		lastMatch := matches[len(matches)-1]
		return lastMatch[1]
	}
	return ""
}

func isValidRemoteGitURL(gitHTTPURL string) error {

	// stupid lib below will return true if the git URL is an empty string.
	if len(gitHTTPURL) == 0 {
		return fmt.Errorf("git URL is empty")
	}

	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{gitHTTPURL},
	})

	// gitOpt := &OptionConfig{}
	// for _, o := range gitOpts {
	// 	o.Apply(gitOpt)
	// }
	opts := &git.ListOptions{}
	if _, err := remote.List(opts); err != nil {
		return err
	}
	return nil
}
