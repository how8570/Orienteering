package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	uuid "github.com/satori/go.uuid"

	// sqlite3 connect lib
	_ "github.com/mattn/go-sqlite3"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	pageByte, err := ioutil.ReadFile("./pages/index.html")
	if err != nil {
		log.Fatal("index get page: ", err)
	}

	page := string(pageByte[:])
	fmt.Fprintf(w, page)
}

type rawPunch struct {
	userID       string
	locationUUID string
	eventUUID    string
}

func (r *rawPunch) decode() (p *punch, err error) {
	p = new(punch)
	p.userID = r.userID
	p.locationUUID, err = uuid.FromString(r.locationUUID)
	if err != nil {
		return nil, err
	}
	p.eventUUID, err = uuid.FromString(r.eventUUID)
	if err != nil {
		return nil, err
	}

	p.reachTime = time.Now()
	return
}

type punch struct {
	userID       string
	reachTime    time.Time
	locationUUID uuid.UUID
	eventUUID    uuid.UUID
}

func (p punch) String() string {
	return fmt.Sprintf("\n%v\n%v\n%v\n%v\n\n", p.userID, p.reachTime, p.locationUUID, p.eventUUID)
}

func handlePunch(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "not correct content type", http.StatusBadRequest)
			return
		}

		if r.Body == nil {
			http.Error(w, "no json come", http.StatusBadRequest)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		json := jsoniter.ConfigCompatibleWithStandardLibrary
		if !json.Valid(bodyBytes) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		bodyString := string(bodyBytes)
		// fmt.Println(bodyString)

		// decode json
		var r rawPunch
		err = json.NewDecoder(strings.NewReader(bodyString)).Decode(&r)
		if err != nil {
			log.Fatal(err)
		}

		// 如果 decode 出來的資訊是空的
		if r.eventUUID == "" || r.locationUUID == "" || r.userID == "" {
			http.Error(w, "not content enough arg or decode fail", http.StatusBadRequest)
			fmt.Println("Result", "ERROR_FAIL_DECODE")
			return
		}

		var p *punch
		p, err = r.decode()
		if err != nil {
			http.Error(w, "decode fail", http.StatusBadRequest)
			return
		}
		// fmt.Println("json 解碼為 : ", p)

		db, err := sql.Open("sqlite3", "./data/database.sqlite3")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		stmt, err := db.Prepare("SELECT * FROM Punch WHERE userID = ? AND locationUUID = ? AND eventUUID = ?;")
		if err != nil {
			log.Fatal(err)
		}

		q, err := stmt.Query(p.userID, p.locationUUID, p.eventUUID)
		if err != nil {
			log.Fatal(err)
		}
		defer q.Close()

		// 已經踩過點了 不重重新插入
		if q.Next() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"Result": "ERROR_ALREADY_PUNCHED"}`)
			return
		}

		stmt, err = db.Prepare("INSERT INTO Punch (userID, reachTime, locationUUID, eventUUID) VALUES   (?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}

		res, err := stmt.Exec(p.userID, p.reachTime.Format(time.RFC1123), p.locationUUID, p.eventUUID)
		// 寫入DB fail 可能是東西髒髒的
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"Result": "ERROR_DB_WRITE_FAIL"}`)
			return
		}

		rows, err := res.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		if rows != 1 {
			log.Fatalf("expected to affect 1 row, affected %d", rows)
		}

		/** TODO **/
		// 打卡成功
		fmt.Println("User ", p.userID, " reach ", p.locationUUID, " !!")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"Result": "OK"}`)
		return
	}
	// reject get/Others method
	http.Error(w, "404 page not found", http.StatusNotFound)

}

// func isValidUUID(u string) bool {
// 	_, err := uuid.FromString(u)
// 	return err == nil
// }

type event struct {
	eventUUID string
	title     string
	content   string
}

func handleEventNames(w http.ResponseWriter, r *http.Request) {

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

func handleEventPoints(w http.ResponseWriter, r *http.Request) {

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
	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/punch", handlePunch)
	r.HandleFunc("/event/names", handleEventNames)
	r.HandleFunc("/event/{UUID:[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}}/points", handleEventPoints)
	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
