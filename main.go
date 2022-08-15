package main

import (
	"Enrollee/src"
	"Enrollee/src/data"
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
)

var db *sql.DB
var err error
var currUser data.Applicant
var programs data.Programs

func main() {
	connStr := "user=postgres password=" + getPassword() + " dbname=Enrollee sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// inserDataToDB()

	//currUser.Login = "akimov_lev@gmail.com"
	currUser.SetDB(db)
	programs.SetDB(db)

	err := programs.GetPrograms()
	if err != nil {
		panic(err)
	}

	for _, program := range programs.Programs {
		fullPath := "/" + program.GetUrl()
		http.Handle(fullPath, src.BaseProgramPageHandler(program))
	}
	http.Handle("/programs", src.GetProgramsPageHandler(&programs, currUser))
	http.Handle("/signup", src.GetSignupPageHandler(db, &currUser))
	http.Handle("/login", src.GetLoginPageHandeler(db, &currUser))
	http.Handle("/logout", src.GetLogoutHangler(&currUser))
	http.Handle("/account", src.GetAccountPageHandler(&currUser))
	http.Handle("/account_postform", src.GetAccountPostFormHandler(db, &currUser))
	http.Handle("/program_choice", src.GetProgramChoicePageHandler(&currUser))
	http.Handle("/program_choice_postform", src.GetProgramChoicePostFormHandler(&currUser))

	err = http.ListenAndServe("localhost:1214", nil)
	if err != nil {
		panic(err)
	}
}
