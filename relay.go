package main

import (
	"fmt"
	"log"
	"net/http"
	db2 "relay/db"
)

var db db2.RedisDB

func main() {
	db.InitDB()

	http.HandleFunc("/", index)
	http.HandleFunc("/registerToken", registerToken)
	http.HandleFunc("/set", set)
	http.HandleFunc("/get", get)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintln(w, "Welcome to relay. Hit up /registerToken to register your first token.")
	if err != nil {
		fmt.Println("Oh no" + err.Error())
	}
	fmt.Println("Sent response")
}

func set(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	checkErr(err)

	token := r.FormValue("token")
	key := r.FormValue("key")
	val := r.FormValue("value")

	if !db.CheckIfTokenExists(token) {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintln(w, "Token does not exist")
		checkErr(err)
		return
	}

	fmt.Printf("Setting token [%s], key [%s], val [%s]\n", token, key, val)

	if !checkErr(db.Set(token, key, val)) {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	checkErr(err)

	token := r.FormValue("token")
	key := r.FormValue("key")

	if !db.CheckIfTokenExists(token) {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintln(w, "Token does not exist")
		checkErr(err)
		return
	}

	err, data := db.Get(token, key)
	if !checkErr(err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprintln(w, data)
	checkErr(err)
}

func registerToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if db.CheckIfTokenExists(token) {
		w.WriteHeader(http.StatusConflict) // 409
		_, err := fmt.Fprintf(w, "Token already in use")
		checkErr(err)
	} else {
		err := db.RegisterToken(token)
		checkErr(err)
		w.WriteHeader(http.StatusCreated)
		_, err = fmt.Fprintf(w, token)
		checkErr(err)
	}
}

/*
Returns false if there was an error, otherwise true.
 */
func checkErr(err error) bool {
	if err != nil {
		// Do more logging here
		fmt.Println(fmt.Errorf("Error: %s\n", err.Error()))
	}
	return err == nil
}
