package apihandler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type event struct {
	eventUUID string
	title     string
	content   string
}

// HandleEventNames is /event/names handle function
func HandleEventNames(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", "./data/database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT UUID, title, content FROM Event ORDER BY id DESC;")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var e event
	var events []event

	for rows.Next() {
		err = rows.Scan(&e.eventUUID, &e.title, &e.content)
		events = append(events, e)
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	response, err := json.Marshal(events)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, string(response))
}
