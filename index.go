package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/spiretechnology/go-pool"
)

type Element struct {
	Directory  string `json:"Directory"`
	Name       string `json:"Name"`
	UpdateTime int    `json:"UpdateTime"`
}

type Patch struct {
	Url  string
	Name string
}

const NEW_SOURCE_URL = "##EXPORTED_DOMAIN_NAME##/"

func ReadGames() []Element {
	path := "/home/danger/.local/share/Tabletop Simulator/Mods/Workshop/"

	jsonFile := path + "WorkshopFileInfos.json"

	body, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
		return nil
	}

	var result []Element

	if err = json.Unmarshal(body, &result); err != nil {
		log.Fatalf("unable to parse the file: %v", err)
		return nil
	}

	return result
}

func ReadResource(name string, path string) {
	log.Default().Printf("reading file: %s", path)

	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "/", "", -1)
	name = strings.Replace(name, `"`, "", -1)
	name = strings.Replace(name, "\\", "", -1)

	target := "./output/" + name

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		log.Default().Printf("skip due to already synced: %s", target)
		return
	}

	f, err := os.OpenFile("logs/"+name+"_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	os.Mkdir(target, os.ModePerm)

	body, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read file [%s]: %v", path, err)
		return
	}

	rows := strings.Split(string(body), "\n")

	log.Default().Printf("file has %d rows", len(rows))
	positive := []string{}

	for _, row := range rows {
		if strings.Index(row, `: "http://`) != -1 || strings.Index(row, `: "https://`) != -1 {
			part := strings.Split(row, ":")

			url := strings.TrimSuffix(strings.Join(part[1:], ":"), ",")
			url = strings.Trim(url, ` `)
			url = strings.Trim(url, `"`)
			url = strings.Trim(url, ` `)

			if strings.HasPrefix(url, "https://melodice.org/") {
				continue
			}

			positive = append(positive, url)
		} else {
			if strings.Contains(row, "http") {
				log.Default().Printf("This row does not match %s", row)
			}
		}

	}

	log.Default().Printf("file has %d urls", len(positive))

	rows = removeDuplicates(positive)

	log.Default().Printf("file has %d unique urls", len(rows))

	patches := make([]Patch, len(rows))
	channels := make(chan Patch)

	p := pool.New(10)

	for _, url := range rows {
		urlCopy := url + ""
		p.Go(func() {
			name, err := StoreResourceTarget(urlCopy, target+"/")
			if err != nil {
				log.Default().Printf("[error] impossible to complete the download of %s: %v", urlCopy, err)
				channels <- Patch{}
				return
			}

			patch := Patch{
				Name: name,
				Url:  urlCopy,
			}

			channels <- patch
		})
	}

	log.Default().Printf("waiting...")
	for i := 0; i < len(rows); i++ {
		msg := <-channels
		log.Default().Printf("received message %v location %d", msg, i)
		patches[i] = msg
	}

	log.Default().Printf("complete...")
	p.Wait()

	split := strings.Split(path, "/")
	outputFile := target + "/" + split[len(split)-1]
	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("unable to create file [%s]: %v", outputFile, err)
		return
	}
	defer out.Close()

	outputPatched := string(body)

	newUrl := NEW_SOURCE_URL + name + "/"

	missingFiles := 0
	for _, patch := range patches {
		log.Default().Printf("patch %s -> %s", patch.Url, patch.Name)
		if patch != (Patch{}) {
			log.Default().Printf("apply patch %s -> %s", patch.Url, patch.Name)
			outputPatched = strings.Replace(outputPatched, patch.Url, newUrl+patch.Name, -1)
		} else {
			missingFiles = missingFiles + 1
		}
	}

	_, err = io.WriteString(out, outputPatched)
	if err != nil {
		log.Fatalf("unable to write resource on disk [%s]: %v", outputFile, err)
		return
	}

	log.Default().Printf("export completed with %d/%d errors", missingFiles, len(patches))
	log.Default().Printf("file: %s", outputFile)
	log.SetOutput(os.Stdout)
}

func StoreResourceTarget(url string, target string) (string, error) {
	log.Default().Printf("[download] start downloading the resource %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Default().Printf("[download][error] unable to get resource %s: %v", url, err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New(fmt.Sprintf("resource not found [%s]", url))
	}

	contentDisposition := resp.Header.Get("Content-Disposition")

	_, params, err := mime.ParseMediaType(contentDisposition)
	filename := params["filename"]

	if filename == "" {
		split := strings.Split(url, "/")
		filename = split[len(split)-1]
	}

	if len(filename) > 120 {
		log.Default().Printf("[download] reduce name length %s -> %s", filename, filename[len(filename)-120:])
		filename = filename[len(filename)-120:]
	}

	outputFile := target + filename
	log.Default().Printf("[download] resource %s has name %s", url, outputFile)

	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		log.Default().Printf("[download][error] skip due to already synced %s as %s", url, outputFile)
		return filename, nil
	}

	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("[download] unable to create resource %s to file %s: %v", url, outputFile, err)
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("[download] unable to write resource on disk %s as file %s: %v", url, outputFile, err)
		return "", err
	}

	log.Default().Printf("[download] url %s saved as %s", url, outputFile)

	return filename, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func removeDuplicates(strList []string) []string {
	list := []string{}
	for _, item := range strList {
		if contains(list, item) == false {
			list = append(list, item)
		}
	}
	return list
}

func main() {
	elements := ReadGames()

	log.Default().Printf("found %d games to export", len(elements))
	for _, element := range elements {
		log.Default().Printf("game: %s", element.Name)
		if strings.HasSuffix(element.Directory, ".json") {
			log.Default().Printf("exporting game: %s", element.Name)
			ReadResource(element.Name, element.Directory)
		} else {
			log.Default().Printf("skipped game: %s wrong format", element.Name)
		}
	}
}
