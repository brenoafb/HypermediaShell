package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gookit/goutil/dump"
)

//go:embed index.html
var index []byte

// TODO actually use sessions
type Session struct {
	ID      uint32
	History []string
}

func main() {

	openFile := ""
	openFileMutex := sync.Mutex{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})
	http.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		openFileMutex.Lock()
		defer openFileMutex.Unlock()

		if openFile == "" {
			fmt.Fprintf(w, "No file open")
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// body will have the format
		// editor=<contents>
		// where contents is the url encoding of the file contents
		values, err := url.ParseQuery(string(body))
		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusInternalServerError)
			return
		}

		dump.P(values)

		contents, ok := values["editor"]
		if !ok {
			http.Error(w, "Error: no editor field", http.StatusInternalServerError)
			return
		}

		if len(contents) != 1 {
			http.Error(w, "Error: too many editor fields", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Contents: %s\n", contents[0])

		// write the contents to the file
		err = ioutil.WriteFile(openFile, []byte(contents[0]), 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error writing file: %v", err), http.StatusInternalServerError)
			return
		}

		// return ok
		fmt.Fprintf(w, "Done editing file %s", openFile)
	})

	http.HandleFunc("/shell", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// pretty print the body
		dump.P("body", string(body))

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
			return
		}

		if parts[0] == "edit" {
			if len(parts) < 2 {
				fmt.Fprintf(w, "Error: no filename")
				return
			}
			filename := parts[1]
			fmt.Fprintf(w, "Editing %s<br>", filename)
			openFileMutex.Lock()
			defer openFileMutex.Unlock()

			openFile = filename

			// check if the file exists.
			// if so, get the contents.
			// otherwise, initialize the contents to ""

			contents := ""

			fileinfo, err := os.Stat(filename)
			if err == nil {
				if fileinfo.IsDir() {
					fmt.Fprintf(w, "Error: %s is a directory", filename)
					return
				}

				if fileinfo.Size() > 0 {
					// read the file
					var err error
					bytes, err := ioutil.ReadFile(filename)
					if err != nil {
						fmt.Fprintf(w, "Error: %v", err)
						return
					}
					contents = string(bytes)
				}
			}

			// send the response with the contents
			response := fmt.Sprintf(`
			<form>
		  <div hx-target="this" hx-swap="outerHTML">
				<textarea name="editor" rows="5" cols="50">%s</textarea>
				<button type="submit" hx-post="/edit">Save</button>
				<button type="submit" hx-post="/edit">Close</button>
  		</div>	
			</form>
			`, contents)
			fmt.Fprintf(w, "%s", response)

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
	})
	port := 8080
	fmt.Printf("Starting server at port %d\n", port)
	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
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

// func disp(filename string) string {
// }
