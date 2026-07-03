package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/lib/pq"
)
type School struct {
	Name	string
	Addr	string
	Level	int
	Center	bool
	Tel		string
	Category	string
	Classes	int
	Students	int
	SmallClasses	int
	SmallStudents	int
	Id	int
	Lat	float64
	Lng	float64
}

func main(){
	// cfg := pq.Config{
	// 	Host: "localhost",
	// 	Port: 5432,
	// 	User: "youyou",
	// }
	fmt.Println("Server start...")
	r := mux.NewRouter()
	r.HandleFunc("/kindgarden",func(w http.ResponseWriter,r *http.Request){
		// vars := mux.Vars(r)
		// title := vars["title"]
		// page := vars["page"]
		db, err := sql.Open("postgres", "host=localhost dbname=youyou sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		row , err := db.Query(`SELECT * FROM kindgarden`)
		defer row.Close()
		var schools []School
		for row.Next() {
			var s School
			err := row.Scan(&s.Name, &s.Addr,&s.Level,&s.Center,&s.Tel,&s.Category,&s.Classes,&s.Students,&s.SmallClasses,&s.SmallStudents,&s.Id,&s.Lat,&s.Lng)
			if err != nil {
				log.Fatal(err)
			}
			schools = append(schools,s)
		}
		err = row.Err()
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Fprintf(w,"String only haha");
		json.NewEncoder(w).Encode(schools)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, you requested: %s\n", r.URL.Path)
    })
	// http.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
	// 	//w.Write([]byte("Hello World"))
	// 	fmt.Fprintf(w,"Hello,you requested :%s\n",r.URL.Path)
	// })
	fs := http.FileServer(http.Dir("/Users/youyou/Documents/practica/学前教育项目/cp/"))
	//http.Handle("/cp/",http.StripPrefix("/cp/",fs))
	r.PathPrefix("/cp/").Handler(http.StripPrefix("/cp/",fs))
	fmt.Println("Server start at :3333...")
	http.ListenAndServe(":3333",r)
}