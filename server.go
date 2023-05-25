package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
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

		values, err := url.ParseQuery(string(body))
		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusInternalServerError)
			return
		}
		command := values.Get("command")
		fmt.Fprintf(w, "> <b>%s</b><br>", command)
		parts := strings.Split(command, " ")
		if len(parts) == 0 {
			fmt.Fprintf(w, "Error: no command")
			w.Write(shellForm)
			return
		}

		log.Printf("Running command: %s", parts)

		cmd := exec.Command(parts[0], parts[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error: %v, out: %s", err, out)
			outputString := formatError(out)
			fmt.Fprintf(w, "%s", outputString)
		} else {
			outputString := formatOutput(out)
			fmt.Fprintf(w, "%s", outputString)
		}

		w.Write(shellForm)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func formatOutput(out []byte) string {
	return fmt.Sprintf(
		"<div class=\"output\">%s</div>",
		strings.Replace(string(out), "\n", "<br>", -1),
	)
}

func formatError(out []byte) string {
	return fmt.Sprintf(
		"<div class=\"error\">%s</div>",
		strings.Replace(string(out), "\n", "<br>", -1),
	)
}
