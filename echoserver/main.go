package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func callback(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("Body reading error: %v", err)
		return
	}

	fmt.Printf("Receiving cluster info: %s\n", string(bodyBytes))
	fmt.Fprintf(w, string(bodyBytes))
}

func main() {
	port := "8090"
	fmt.Printf("Starting request echo server on port %v\n", port)
	http.HandleFunc("/callback", callback)
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
