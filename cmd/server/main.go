package main

import (
	_ "embed"

	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

//go:embed index.html
var index []byte

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})
	http.Handle("/ws", websocket.Handler(WSServer))

	port := 8081
	fmt.Printf("Starting server at port %d\n", port)
	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func WSServer(ws *websocket.Conn) {
	fmt.Printf("New connection\n")

	for {
		data := make(map[string]interface{})
		err := websocket.JSON.Receive(ws, &data)
		if err != nil {
			streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
			continue
		}

		fmt.Printf("Received: %v\n", data)

		commandLine := data["command"].(string)

		if err != nil {
			streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
			continue
		}

		fmt.Printf("Received: %v\n", commandLine)

		execCommand(ws, commandLine)
	}
}

func cd(dir string) error {
	if dir == "" {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		dir = usr.HomeDir
	}
	return os.Chdir(dir)
}

func execCommand(ws *websocket.Conn, command string) {
	if strings.HasPrefix(command, "clear") {
		clear(ws)
		return
	}

	header := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
		<hr>
		<b>%s</b>
		<br></div>
		`,
		command,
	)
	_, _ = ws.Write([]byte(header))

	defer func() {
		footer := `
		<div id="notifications" hx-swap-oob="beforeend">
		<hr>
		</div>
		`

		_, _ = ws.Write([]byte(footer))
	}()

	if strings.HasPrefix(command, "cd") {
		target := ""
		parts := strings.Split(command, " ")
		if len(parts) == 2 {
			target = parts[1]
		}
		err := cd(target)
		if err != nil {
			streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
		}
		return
	}

	cmd := exec.Command("/bin/sh", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
		return
	}

	err = cmd.Start()
	if err != nil {
		streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
		return
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go streamHTMXOutput(ws, stdout, &wg)
	go streamHTMXOutput(ws, stderr, &wg)

	err = cmd.Wait()
	if err != nil {
		streamHTMXOutput(ws, strings.NewReader(err.Error()), nil)
		return
	}
	wg.Wait()
}

func clear(ws *websocket.Conn) {
	result :=
		"<div id=\"notifications\" hx-swap-oob=\"outerHTML\"></div>"
	_, _ = ws.Write([]byte(result))
}

func streamHTMXOutput(ws *websocket.Conn, r io.Reader, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		result := fmt.Sprintf(`
			<div id="notifications" hx-swap-oob="beforeend">
			%s
			<br></div>
			`,
			text,
		)
		_, _ = ws.Write([]byte(result))
	}

	if wg != nil {
		wg.Done()
	}
}
