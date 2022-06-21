package main

import (
	"Enrollee/src"
	"bufio"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var db *sql.DB
var err error
var currUser src.Applicant

var subjectNames = []string{
	"Russian language",
	"Mathematics",
	"Physics",
	"Chemistry",
	"History",
	"Social studies",
	"Computer science",
	"Biology",
	"Geography",
	"Foreign languages",
	"Literature",
}

func updatePasswords() {
	f, _ := os.Open("passwords")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		splitted := strings.Split(scanner.Text(), "; ")
		log.Println(splitted)
		login := splitted[0]
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(splitted[1]), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		_, err = db.Exec("UPDATE enrollee SET password_hash = $1 WHERE login = $2", passwordHash, login)
		if err != nil {
			panic(err)
		}
	}
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

func baseProgramPageHandlerFunc(program src.Program) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := program.GetEnrollees(db)
		if err != nil {
			panic(err)
		}
		tmp, err := template.New("program.html").Funcs(template.FuncMap{"inc": func(a int) int { return a + 1 }}).ParseFiles("templates/program.html")

		if err != nil {
			panic(err)
		}
		err = tmp.Execute(w, program)
		if err != nil {
			panic(err)
		}
	})
}

func signupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "templates/signup.html")
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")
	log.Println("Get password", password)
	var user string

	row := db.QueryRow("SELECT last_name FROM enrollee WHERE login = $1", login)
	err = row.Scan(&user)

	switch {
	case err == sql.ErrNoRows:
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			//http.Error(w, "Server error, unable to create your account.", 500)
			return
		}

		log.Println(passwordHash)
		_, err = db.Exec("INSERT INTO enrollee(login, password_hash) VALUES($1, $2)", login, passwordHash)
		if err != nil {
			log.Println(err)
			http.Error(w, "Server error, unable to create your account.", 500)
			return
		}

		//w.Write([]byte("User created!"))
		currUser.Login = login
		http.Redirect(w, r, "/account", 301)
		return
	case err != nil:
		http.Error(w, "Server error, unable to create your account.", 500)
		return
		//default:
		//http.Redirect(w, r, "/signup", 301)
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	var truePasswordHash string

	err := db.QueryRow("SELECT password_hash FROM enrollee WHERE login=$1", login).Scan(&truePasswordHash)

	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", 301)
		return
	}
	log.Println(truePasswordHash, password)
	log.Println([]byte(truePasswordHash), []byte(password))
	err = bcrypt.CompareHashAndPassword([]byte(truePasswordHash), []byte(password))

	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", 301)
		return
	}
	currUser.Login = login
	http.Redirect(w, r, "/programs", 301)
	return
}

func accountPostformPage(w http.ResponseWriter, r *http.Request) {
	//if r.Method != "POST" {
	//	http.ServeFile(w, r, "templates/account.html")
	//}

	if currUser.Login == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	//applicant := new(src.Applicant)
	currUser.FirstName = r.FormValue("firstName")
	currUser.LastName = r.FormValue("lastName")
	currUser.MiddleName = r.FormValue("middleName")
	currUser.Country = r.FormValue("country")

	log.Println(currUser)
	currUser.Subjects = make(map[string]uint16)
	var s_score string
	for _, subject := range subjectNames {
		s_score = r.FormValue(strings.Replace(strings.ToLower(subject), " ", "_", 1))
		log.Println(subject, s_score)
		if len(s_score) > 0 {
			i_score, err := strconv.Atoi(s_score)
			if err == nil && i_score > 0 && i_score <= 100 {
				currUser.Subjects[subject] = uint16(i_score)
			} else if err != nil {
				log.Println(err)
			}
		}
	}

	currUser.Login = currUser.Login
	log.Println(currUser)
	err = currUser.AddToDB(db)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, unable to create your account.", 500)
		return
	}

	log.Println("Redirect to /programs")
	http.Redirect(w, r, "/programs", http.StatusSeeOther)
	log.Println("Redirect to /programs")
	return
}

func main() {
	connStr := "user=postgres password=" + getPassword() + " dbname=Enrollee sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var programs src.Programs
	err = programs.GetPrograms(db)
	currUser.Login = "akimov_lev@gmail.com"
	currUser.SetDB(db)
	//log.Println(programs.Programs)
	if err != nil {
		panic(err)
	}
	//http.HandleFunc("/singup", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "templates/sign_in.html")
	//})
	//

	http.HandleFunc("/programs", func(w http.ResponseWriter, r *http.Request) {
		//log.Println(programs.Programs)
		tmp, err := template.ParseFiles("templates/programs.html")
		if err != nil {
			panic(err)
		}
		err = tmp.Execute(w, programs.Programs)
		if err != nil {
			return
		}
		//http.ServeFile(w, r, "templates/programs.html")
	})

	//http.HandleFunc("/sign_up", func(w http.ResponseWriter, r *http.Request) {
	//	login := r.FormValue("login")
	//	password := r.FormValue("password")
	//	password_repeat := r.FormValue("password repeat")
	//
	//})

	for _, program := range programs.Programs {
		fullPath := "/" + program.GetUrl()
		http.Handle(fullPath, baseProgramPageHandlerFunc(program))
	}

	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {

		log.Println(currUser.Login)
		if currUser.Login == "" {
			log.Println("Authorization is required")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "templates/account.html")
	})
	http.HandleFunc("/program_choose", func(w http.ResponseWriter, r *http.Request) {
		log.Println(currUser.Login)
		if currUser.Login == "" {
			log.Println("Authorization is required")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		tmp, err := template.ParseFiles("templates/program_choose.html")
		if err != nil {
			panic(err)
		}
		err = tmp.Execute(w, currUser.GetAvailablePrograms())
	})
	http.HandleFunc("/postform_account", accountPostformPage)

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
	http.ListenAndServe("localhost:1214", nil)
}
