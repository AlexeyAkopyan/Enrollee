package main

//
//import (
//	"database/sql"
//	"errors"
//	"fmt"
//	_ "github.com/lib/pq"
//	"golang.org/x/crypto/bcrypt"
//	"log"
//	"net/http"
//)
//
//
//func checkUser(db *sql.DB, login string, password string) (bool, error, string) {
//	row := db.QueryRow("SELECT last_name, first_name, middle_name, password_hash FROM enrollee "+
//		"WHERE login = $1", login)
//
//	var last_name, first_name, middle_name string
//	var password_hash string
//	err := row.Scan(&last_name, &first_name, &middle_name, &password_hash)
//	if errors.Is(err, sql.ErrNoRows) {
//		log.Println(login, "has not registered")
//		return false, nil, ""
//	} else if err != nil {
//		return false, err, ""
//	}
//	//rows, err := db.Query("SELECT password_hash FROM enrollee")
//	//if err != nil {
//	//	return false, err, ""
//	//}
//	//var pass string
//	//for rows.Next() {
//	//	row.Scan(&pass)
//	pass_hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
//	log.Println(pass_hash, string(pass_hash))
//	log.Println([]byte(password_hash), password_hash)
//	//err = bcrypt.CompareHashAndPassword(pass_hash, []byte("cC7p6pknkWAP"))
//	//if err != nil {
//	//	log.Println("Here wrong password")
//	//	return false, nil, ""
//	//} else {
//	//	log.Println("Here correct password")
//	//}
//	//log.Println(string(pass_hash))
//	//}
//	//passed_password_hash, err := bcrypt.GenerateFromPassword(password, 14)
//	err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(password))
//	if err != nil {
//		log.Println("User is found, but wrong password is passed")
//		return false, nil, ""
//	} else {
//		log.Println("User is found, password is correct")
//	}
//	log.Println(last_name, first_name, middle_name, "has registered")
//	return true, nil, first_name
//
//}

//func main() {
//	connStr := "user=postgres password=" + getPassword() + " dbname=Enrollee sslmode=disable"
//	db, err := sql.Open("postgres", connStr)
//	if err != nil {
//		panic(err)
//	}
//	defer db.Close()
//
//	http.HandleFunc("/singup", func(w http.ResponseWriter, r *http.Request) {
//		http.ServeFile(w, r, "templates/sign_in.html")
//	})
//
//	http.HandleFunc("/programms", func(w http.ResponseWriter, r *http.Request) {
//		http.ServeFile(w, r, "templates/programs.html")
//	})
//
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		http.ServeFile(w, r, "templates/sign_in.html")
//	})
//
//	http.HandleFunc("/postform", func(w http.ResponseWriter, r *http.Request) {
//
//		login := r.FormValue("login")
//		password := r.FormValue("password")
//		isUser, err, name := checkUser(db, login, password)
//		if err != nil {
//			panic(err)
//		} else if isUser != true {
//			fmt.Fprintf(w, "no user")
//
//		} else {
//			fmt.Fprintf(w, name+" is found")
//		}
//	})
//
//	log.Println("Server is listening")
//	http.ListenAndServe("localhost:1212", nil)
//}
