package main

import (
	"fmt"
	// "io/ioutil"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func main() {
	// Check if the user provided a file name
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file name")
		os.Exit(1)
	}


	// Get the file name from the command line
	filename := os.Args[1]

	// Read the file
	// data, err := ioutil.ReadFile(filename)
	// if err != nil {
	// 	fmt.Println("Error reading file:", err)
	// 	os.Exit(1)
	// }

	mtype, err := mimetype.DetectFile(filename)

	if err != nil {
		fmt.Println("Error getting file type:", err)
		os.Exit(1)
	}

	fmt.Printf("mime: %s\t%s\n\n", mtype.String(), mtype.Extension())


	// Print the file's content to stdout
	// fmt.Println(string(data))
}

