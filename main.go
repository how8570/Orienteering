package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	// handler functions
	handler "./api/handler"

	"github.com/gorilla/mux"
	"github.com/json-iterator/go/extra"
	uuid "github.com/satori/go.uuid"

	// sqlite3 connect lib
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	extra.SupportPrivateFields()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
}

func main() {
	// new a uuid
	u1, _ := uuid.NewV4()
	fmt.Println("UUIDv4: %s", u1)
	fmt.Println("%v", time.Now().Format(time.RFC1123))

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HandleIndex)
	r.HandleFunc("/punch", handler.HandlePunch)
	r.HandleFunc("/event/names", handler.HandleEventNames)
	// "/event/<UUID>/imgae"
	r.HandleFunc("/event/{UUID:[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}}/image", handler.HandleEventImage)
	// "/event/<UUID>/points"
	r.HandleFunc("/event/{UUID:[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}}/points", handler.HandleEventPoints)
	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
