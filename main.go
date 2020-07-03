package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	APIHandler "./apihandler"

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
	fmt.Printf("UUIDv4: %s\n", u1)
	fmt.Printf("%v", time.Now().Format(time.RFC1123))

	r := mux.NewRouter()
	r.HandleFunc("/", APIHandler.HandleIndex)
	r.HandleFunc("/punch", APIHandler.HandlePunch)
	r.HandleFunc("/event/names", APIHandler.HandleEventNames)
	r.HandleFunc("/event/{UUID:[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}}/points", APIHandler.HandleEventPoints)
	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
