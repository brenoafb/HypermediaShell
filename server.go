package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//go:embed index.html
var index []byte

//go:embed shell.html
var shellForm []byte

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})
	http.HandleFunc("/shell", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// The body is now in the `body` variable as a []byte.
		// If you're expecting a string, you can convert it like this:
		values, err := url.ParseQuery(string(body))
		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusInternalServerError)
			return
		}
        command := values.Get("command")
		fmt.Fprintf(w, "Hello, there: %s", command)
		w.Write(shellForm)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
