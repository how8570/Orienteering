package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

type eventPoints struct {
	eventUUID string
	points    []point
}

type point struct {
	pointUUID  string
	pointOrder int
	longitude  float64
	latitude   float64
	title      string
	content    string
}

// HandleEventPoints is /event/<event-UUID>/points handle function
func HandleEventPoints(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	eventUUID := vars["UUID"]

	db, err := sql.Open("sqlite3", "./data/database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT * FROM Event WHERE UUID = ?;")
	if err != nil {
		log.Fatal(err)
	}

	q, err := stmt.Query(eventUUID)
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()

	if !q.Next() {
		http.Error(w, "404 page not found!", http.StatusNotFound)
		return
	}

	stmt, err = db.Prepare("SELECT pointUUID, pointOrder FROM Event_Point WHERE eventUUID = ?;")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := stmt.Query(eventUUID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var ep eventPoints
	var p point

	ep.eventUUID = eventUUID

	for rows.Next() {
		err = rows.Scan(&p.pointUUID, &p.pointOrder)

		stmt, err := db.Prepare("SELECT longitude, latitude, title, content FROM Point WHERE UUID = ?;")
		if err != nil {
			log.Fatal(err)
		}

		q := stmt.QueryRow(p.pointUUID)
		err = q.Scan(&p.longitude, &p.latitude, &p.title, &p.content)
		if err != nil {
			log.Fatal(err)
		}

		ep.points = append(ep.points, p)

	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	response, err := json.Marshal(ep)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, string(response))
}
