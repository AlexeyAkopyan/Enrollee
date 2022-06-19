package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Program struct {
	Name, Department string
	subjects         map[string]uint16
	plan             uint16
}

func (p Program) GetUrl() string {
	return "programs_" + strings.Replace(strings.ToLower(p.Name), " ", "_", -1)
}

func getPrograms(db *sql.DB) ([]Program, error) {
	rows, err := db.Query("SELECT name_program, subjects, plan, name_department, min_results FROM get_programs ()")
	var programs []Program
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var program Program
		var subjects, minResults string
		err := rows.Scan(&program.Name, &subjects, &program.plan, &program.Department, &minResults)
		if err != nil {
			return nil, err
		}
		program.subjects = make(map[string]uint16)
		for _, subject := range strings.Split(subjects, ";") {
			for _, s_minResult := range strings.Split(minResults, ";") {
				i_minResult, err := strconv.Atoi(s_minResult)
				if err != nil {
					return nil, err
				}
				program.subjects[subject] = uint16(i_minResult)
			}
		}
		log.Println(program)
		programs = append(programs, program)
	}

	return programs, nil

}

func checkUser(db *sql.DB, login string, password string) (bool, error, string) {
	row := db.QueryRow("SELECT last_name, first_name, middle_name FROM enrollee "+
		"WHERE login = $1 AND password_hash = $2",
		login, password)
	var last_name, first_name, middle_name string
	err := row.Scan(&last_name, &first_name, &middle_name)
	if errors.Is(err, sql.ErrNoRows) {
		log.Println(login, "has not registered")
		return false, nil, ""
	} else if err != nil {
		return false, err, ""
	}
	log.Println(last_name, first_name, middle_name, "has registered")
	return true, nil, first_name
}

func main() {
	connStr := "user=postgres password=" + getPassword() + " dbname=Enrollee sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var programs []Program
	programs, err = getPrograms(db)
	if err != nil {
		panic(err)
	}
	fmt.Println(programs)

	//http.HandleFunc("/singup", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "templates/sign_in.html")
	//})
	//
	http.HandleFunc("/programs", func(w http.ResponseWriter, r *http.Request) {

		tmp, err := template.ParseFiles("templates/programs.html")
		if err != nil {
			panic(err)
		}
		err = tmp.Execute(w, programs)
		if err != nil {
			return
		}
		//http.ServeFile(w, r, "templates/programs.html")
	})
	//
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "templates/sign_in.html")
	//})
	//
	//http.HandleFunc("/postform", func(w http.ResponseWriter, r *http.Request) {
	//
	//	login := r.FormValue("login")
	//	password := r.FormValue("password")
	//	isUser, err, name := checkUser(db, login, password)
	//	if err != nil {
	//		panic(err)
	//	} else if isUser != true {
	//		fmt.Fprintf(w, "no user")
	//
	//	} else {
	//		fmt.Fprintf(w, name+" is found")
	//	}
	//})
	//
	//log.Println("Server is listening")
	http.ListenAndServe("localhost:1213", nil)
}
