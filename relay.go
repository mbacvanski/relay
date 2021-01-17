package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"relay/data"
)

var db data.RedisDB

func main() {
	_ = godotenv.Load()

	db.InitDB()

	http.HandleFunc("/", index)
	http.HandleFunc("/registerToken", registerToken)
	http.HandleFunc("/set", set)
	http.HandleFunc("/get", get)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"),
		handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func index(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintln(w, "Welcome to relay. Hit up /registerToken to register your first token.")
	if err != nil {
		fmt.Println("Oh no" + err.Error())
	}
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

	err, ret := db.Get(token, key)
	if !checkErr(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = fmt.Fprintln(w, ret)
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
