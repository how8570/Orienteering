package apihandler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// HandleIndex is / handle function
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	pageByte, err := ioutil.ReadFile("./pages/index.html")
	if err != nil {
		log.Fatal("index get page: ", err)
	}

	page := string(pageByte[:])
	fmt.Fprintf(w, page)
}
