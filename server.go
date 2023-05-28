package main

import (
	_ "embed"

	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

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
			fmt.Printf("Error: %v.\n", err)
			break
		}

		fmt.Printf("Received: %v\n", data)

		commandLine := data["command"].(string)

		parts := strings.Split(commandLine, " ")
		if len(parts) == 0 {
			fmt.Fprintf(ws, "Error: no command")
			return
		}

		if err != nil {
			fmt.Printf("Error: %v.\n", err)
			break
		}

		fmt.Printf("Received: %v\n", commandLine)

		execCommand(ws, commandLine)
	}
}

func execCommand(ws *websocket.Conn, command string) {
	if strings.HasPrefix(command, "clear") {
		clear(ws)
		return
	}

	cmd := exec.Command("/bin/sh", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errStr := fmt.Sprintf("%s", err.Error())
		streamHTMXOutput(ws, strings.NewReader(errStr))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		errStr := fmt.Sprintf("%s", err.Error())
		streamHTMXOutput(ws, strings.NewReader(errStr))
		return
	}

	header := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
		<hr>
		<b> > %s</b>
		<br></div>
		`,
		command,
	)
	_, _ = ws.Write([]byte(header))

	err = cmd.Start()
	if err != nil {
		errStr := fmt.Sprintf("%s", err.Error())
		streamHTMXOutput(ws, strings.NewReader(errStr))
		return
	}

	go streamHTMXOutput(ws, stdout)
	go streamHTMXOutput(ws, stderr)

	err = cmd.Wait()
	if err != nil {
		errStr := fmt.Sprintf("%s", err.Error())
		streamHTMXOutput(ws, strings.NewReader(errStr))
		return
	}

	footer := fmt.Sprintf(`
		<div id="notifications" hx-swap-oob="beforeend">
		<hr>
		</div>
		`,
	)
	_, _ = ws.Write([]byte(footer))
}

func clear(ws *websocket.Conn) {
	result := fmt.Sprintf(
		"<div id=\"notifications\" hx-swap-oob=\"outerHTML\"></div>",
	)
	_, _ = ws.Write([]byte(result))
}

func streamHTMXOutput(ws *websocket.Conn, r io.Reader) {
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
}
