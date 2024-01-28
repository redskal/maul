/*
 * Maul
 *
 * Use it to build wordlists from Burp targets. Export URLs and run
 * through Maul to get subdomain, path, file and parameter fuzzing
 * lists which can then be merged into master wordlists with something
 * like TomNomNom's anew tool.
 *
 * The goal is to create more realistic fuzzing lists than are publicly
 * available from what we're seeing in the wild.
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/redskal/maul/internal/helpers"
)

var (
	// output files
	outputFiles = map[string]string{
		"files":  "files.txt",
		"paths":  "paths.txt",
		"subs":   "subdomains.txt",
		"params": "parameters.txt",
	}

	// for de-duplication
	mapFiles      sync.Map
	mapPaths      sync.Map
	mapSubdomains sync.Map
	mapParameters sync.Map

	usage = `
Usage of Maul:
-f string
		File to process.
-ef string
		Exclude files with given extensions. Comma-separated list. (default ".png,.jpg,.svg,.woff,.ttf,.eot")
-o string
		Directory to output files to. (default "./")
-t int
		Amount of threads to run. (default 50)

Input can also be supplied by piping it in.
Eg.
	$ cat urls.txt | maul
	$ maul < urls.txt

Output files are:
	files.txt      - any filenames found
	paths.txt      - any paths up to a depth of 2 (/path/here)
	subdomains.txt - any subdomains it can identify
	parameters.txt - names of any parameters it finds
`
)

type empty struct{}

type values struct {
	subdomain  string
	path       string
	file       string
	parameters []string
}

func main() {
	inputFile := flag.String("f", "", "File to process.")
	outputDir := flag.String("o", "./", "Directory to output files to.")
	excludeFiles := flag.String("ef", ".png,.jpg,.svg,.woff,.ttf,.eot", "Exclude files with given extensions. Comma-separated list.")
	threadCount := flag.Int("t", 50, "Amount of threads to run.")
	flag.Usage = func() {
		fmt.Println(usage)
	}
	flag.Parse()

	if !hasStdin() && len(*inputFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	gather := make(chan values)
	urls := make(chan string)
	tracker := make(chan empty)

	// start workers
	for i := 0; i < *threadCount; i++ {
		go worker(tracker, gather, urls)
	}

	// magic the gathering
	go func() {
		for v := range gather {
			// file
			if len(v.file) > 0 && !isExcludedFile(v.file, *excludeFiles) && unique(v.file, &mapFiles) {
				err := appendFile(filepath.Join(*outputDir, outputFiles["files"]), v.file)
				if err != nil {
					log.Println(err)
				}
			}

			// path
			if len(v.path) > 0 && unique(v.path, &mapPaths) {
				err := appendFile(filepath.Join(*outputDir, outputFiles["paths"]), v.path)
				if err != nil {
					log.Println(err)
				}
			}

			// subdomain
			if len(v.subdomain) > 0 && unique(v.subdomain, &mapSubdomains) {
				err := appendFile(filepath.Join(*outputDir, outputFiles["subs"]), v.subdomain)
				if err != nil {
					log.Println(err)
				}
			}

			// parameters
			for _, param := range v.parameters {
				if len(param) > 0 && unique(param, &mapParameters) {
					err := appendFile(filepath.Join(*outputDir, outputFiles["params"]), param)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
		var e empty
		tracker <- e
	}()

	// process input into urls channel
	if hasStdin() {
		// add lines to urls channel
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			urls <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println(scanner.Err().Error())
		}
	}
	if *inputFile != "" {
		// try to open file and add lines to urls channel
		f, err := os.Open(*inputFile)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			urls <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}

	// wait for workers
	close(urls)
	for i := 0; i < *threadCount; i++ {
		<-tracker
	}

	// close gathering channel and wait for thread
	close(gather)
	<-tracker
}

// worker processes URLs into a values struct and adds to the gathering
// channel
func worker(tracker chan empty, gather chan values, urls chan string) {
	for url := range urls {
		v := processUrl(url)
		gather <- v
	}

	var e empty
	tracker <- e
}

// processUrl grabs subdomain, path, filename and parameter
// names from the given URL. Outputs a values struct.
func processUrl(url string) values {
	var v values

	subdomain, err := helpers.GetSubdomain(url)
	if err == nil {
		v.subdomain = subdomain
	}

	path, err := helpers.GetPath(url)
	if err == nil {
		v.path = path
	}

	f, err := helpers.GetFile(url)
	if err == nil {
		v.file = f
	}

	params, err := helpers.GetParameterNames(url)
	if err == nil {
		v.parameters = append(v.parameters, params...)
	}

	return v
}

// appendFile appends a string to the end of a file.
func appendFile(fileName, s string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// we need to add newlines because WriteString() won't
	outString := fmt.Sprintf("%s\n", s)
	if _, err = f.WriteString(outString); err != nil {
		return err
	}

	return nil
}

// hasStdin checks for piped input from character devices
// or FIFO. Original:
// https://github.com/projectdiscovery/fileutil/blob/380e33ef95825c6b781f289d8cd9c0d48d6c67f5/file.go#L141
func hasStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	mode := stat.Mode()
	return (mode&os.ModeCharDevice) == 0 || (mode&os.ModeNamedPipe) != 0
}

// unique checks if the given string was unique for the given map.
// Modified version of https://github.com/hakluke/hakrawler/blob/master/hakrawler.go#L285
func unique(s string, syncMap *sync.Map) bool {
	_, present := syncMap.Load(s)
	if present {
		return false
	}
	syncMap.Store(s, true)
	return true
}

// isExcludedFile checks file suffix is not in the given list
// of extensions.
func isExcludedFile(file, extensions string) bool {
	extensionList := strings.Split(extensions, ",")
	for _, v := range extensionList {
		if strings.HasSuffix(file, v) {
			return true
		}
	}
	return false
}
