package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	URL           string
	TargetPath    string
	TotalSections int
}

func main() {
	startTime := time.Now()

	download := Download{
		// Provide the URL to download, example: https://www.dropbox.com/s/lgvhj/sample.mp4?dl=1
		URL: "",
		// Provide the target file path with extension, example: sample.mp4
		TargetPath: "",
		// Number of sections/connections to make to the server
		TotalSections: 10,
	}
	err := download.Do()
	if err != nil {
		log.Printf("An error occurred while downloading the file: %s\n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

// Start the download
func (download Download) Do() error {
	fmt.Println("Checking URL")

	res, err := download.getNewRequest("HEAD")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(res)
	if err != nil {
		return err
	}
	fmt.Printf("Got %v\n", resp.StatusCode)

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	fmt.Printf("Size is %v bytes\n", size)

	var sections = make([][2]int, download.TotalSections)
	eachSize := size / download.TotalSections
	fmt.Printf("Each size is %v bytes\n", eachSize)

	// example: if file size is 100 bytes, our sections : [[0 10] [11 21] [22 32] [33 43] [44 54] [55 65] [66 76] [77 87] [88 98] [99 99]]
	for i := range sections {
		if i == 0 {
			// starting byte of first section
			sections[i][0] = 0
		} else {
			// starting byte of other sections
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < download.TotalSections-1 {
			// ending byte of other sections
			sections[i][1] = sections[i][0] + eachSize
		} else {
			// ending byte of other sections
			sections[i][1] = size - 1
		}
	}

	log.Println(sections)
	var wg sync.WaitGroup

	// download each section concurrently
	for i, s := range sections {
		wg.Add(1)
		go func(i int, s [2]int) {
			defer wg.Done()
			err = download.downloadSection(i, s)
			if err != nil {
				panic(err)
			}
		}(i, s)
	}
	wg.Wait()

	return download.mergeFiles(sections)
}

// Download a single section and save content to a tmp file
func (download Download) downloadSection(i int, c [2]int) error {
	r, err := download.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	fmt.Printf("Downloaded %v bytes for section %v\n", resp.Header.Get("Content-Length"), i)
	
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Get new http request
func (download Download) getNewRequest(method string) (*http.Request, error) {
	resp, err := http.NewRequest(
		method,
		download.URL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp.Header.Set("User-Agent", "Silly Download Manager v001")
	return resp, nil
}

// Merge tmp files to single file and delete tmp files
func (download Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(download.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := range sections {
		tmpFileName := fmt.Sprintf("section-%v.tmp", i)
		b, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%v bytes merged\n", n)
	}
	return nil
}