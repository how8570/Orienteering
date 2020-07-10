package handler

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// HandleEventImage is /event/<event-UUID>/image handle function
func HandleEventImage(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	vars := mux.Vars(r)
	eventUUID := vars["UUID"]

	w.Header().Set("content-type", "image/png")

	file, err := os.Open("./asserts/event_img/" + eventUUID + ".png")
	if err != nil {
		// fmt.Println(err)
		file, _ = os.Open("./asserts/event_img/emt2.jpg")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println(err)
	}

	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, img)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(buffer.Bytes())
}
