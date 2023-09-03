package main

import (
	"fmt"
)

func main() {
	response := fmt.Sprintf(`<form>
	<div hx-target="this" hx-swap="outerHTML">
		<textarea name="editor" rows="5" cols="50">%s</textarea>
		<button type="submit" hx-post="/edit">Save</button>
		<button type="submit" hx-post="/edit">Close</button>
	</div>	
	</form>
	`, contents)

	fmt.Println(response)
}
