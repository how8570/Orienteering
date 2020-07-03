package handler

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
)

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

// HandlePunch is /punch handle function
func HandlePunch(w http.ResponseWriter, r *http.Request) {

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
