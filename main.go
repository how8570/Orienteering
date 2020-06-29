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
	userUUID     string
	reachTime    string
	locationUUID string
	eventUUID    string
}

func (r *rawPunch) decode() (p *punch, err error) {
	p = new(punch)
	p.userUUID, err = uuid.FromString(r.userUUID)
	if err != nil {
		return nil, err
	}
	p.locationUUID, err = uuid.FromString(r.locationUUID)
	if err != nil {
		return nil, err
	}
	p.eventUUID, err = uuid.FromString(r.eventUUID)
	if err != nil {
		return nil, err
	}
	p.reachTime, err = time.Parse(time.RFC1123, r.reachTime)
	if err != nil {
		return nil, err
	}
	return
}

type punch struct {
	userUUID     uuid.UUID
	reachTime    time.Time
	locationUUID uuid.UUID
	eventUUID    uuid.UUID
}

func (p punch) String() string {
	return fmt.Sprintf("\n%v\n%v\n%v\n%v\n\n", p.userUUID, p.reachTime, p.locationUUID, p.eventUUID)
}

func handlePunch(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET": // reject get method
		http.Error(w, "404 page not found", 404)
	case "POST":

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "not correct content type", 400)
			return
		}

		if r.Body == nil {
			http.Error(w, "no json come", 400)
			return
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		json := jsoniter.ConfigCompatibleWithStandardLibrary
		if !json.Valid(bodyBytes) {
			http.Error(w, "bad request", 400)
			return
		}

		bodyString := string(bodyBytes)
		fmt.Println(bodyString)

		// decode json
		var r rawPunch
		err = json.NewDecoder(strings.NewReader(bodyString)).Decode(&r)
		if err != nil {
			log.Fatal(err)
		}

		// 如果 decode 出來的資訊是空的
		if r.eventUUID == "" || r.locationUUID == "" || r.reachTime == "" || r.userUUID == "" {
			http.Error(w, "not content enough arg or decode fail", 400)
			fmt.Println("Result", "ERROR_FAIL_DECODE")
			return
		}

		var p *punch
		p, err = r.decode()
		if err != nil {
			http.Error(w, "decode fail", 400)
			return
		}
		fmt.Println("json 解碼為 : ", p)

		db, err := sql.Open("sqlite3", "./data/punch.sqlite3")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		stmt, err := db.Prepare("SELECT * FROM Punch WHERE userUUID = ? AND locationUUID = ? AND eventUUID = ?;")
		if err != nil {
			log.Fatal(err)
		}

		q, err := stmt.Query(p.userUUID, p.locationUUID, p.eventUUID)
		if err != nil {
			log.Fatal(err)
		}
		defer q.Close()

		/** TODO **/
		// 已經踩過點了 不重重新插入
		if q.Next() {
			fmt.Println("Result", "ERROR_ALREADY_PUNCHED")
			return
		}

		stmt, err = db.Prepare("INSERT INTO Punch (userUUID, reachTime, locationUUID, eventUUID) VALUES   (?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}

		res, err := stmt.Exec(p.userUUID, p.reachTime, p.locationUUID, p.eventUUID)
		/** TODO **/
		// 寫入DB fail 可能是東西髒髒的
		if err != nil {
			fmt.Println("Result", "ERROR_DB_WRITE_FAIL")
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
		fmt.Println("Result", "OK")
		return

	default: // reject others http request method
		http.Error(w, "404 page not found", 404)
	}
}

func isValidUUID(u string) bool {
	_, err := uuid.FromString(u)
	return err == nil
}

func init() {
	extra.SupportPrivateFields()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
}

func main() {
	// new a uuid
	u1, _ := uuid.NewV4()
	fmt.Printf("UUIDv4: %s\n", u1)

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/punch", handlePunch)
	err := http.ListenAndServe(":80", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
