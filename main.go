package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("Welcome to my Download Manager !")

	fileUrl := "https://file-examples-com.github.io/uploads/2017/11/file_example_MP3_5MG.mp3"
	fileTargetPath := "abc.mp3"

	startTime := time.Now()
	// Initialize a download manager
	dm := Download{
		Url:           fileUrl,
		Targetpath:    fileTargetPath,
		TotalSections: 10,
	}

	err := dm.Do()
	if err != nil {
		log.Printf("Error occurred: %s", err)
	}
	endTime := time.Now()

	fmt.Printf("Download completed successfully in %v seconds. \n", endTime.Sub(startTime).Seconds())
}
