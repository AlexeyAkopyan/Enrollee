package src

import (
	"Enrollee/src/data"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func BaseProgramPageHandler(program data.Program) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := program.GetEnrollees()
		log.Println("program enrollees", program.Enrollees)
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

func GetProgramsPageHandler(programs *data.Programs, currUser data.Applicant) http.Handler {
	type ProgramParams struct {
		Programs data.Programs
		CurrUser data.Applicant
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Println(programs.Programs)
		programs.SortBy = r.FormValue("sort_by")
		programs.KeyWords = strings.Split(r.FormValue("search_by"), " ")

		err := programs.GetPrograms()
		if err != nil {
			panic(err)
		}
		tmp, err := template.ParseFiles("templates/programs.html")
		if err != nil {
			panic(err)
		}

		programsParams := ProgramParams{*programs, currUser}
		err = tmp.Execute(w, programsParams)
		if err != nil {
			return
		}
		//http.ServeFile(w, r, "templates/programs.html")``
	})
}

func GetSignupPageHandler(db *sql.DB, currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.ServeFile(w, r, "templates/signup.html")
			return
		}

		login := r.FormValue("login")
		password := r.FormValue("password")
		log.Println("Get password", password)
		var user string

		row := db.QueryRow("SELECT last_name FROM enrollee WHERE login = $1", login)
		err := row.Scan(&user)

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
	})
}

func GetLoginPageHandeler(db *sql.DB, currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

func GetLogoutHangler(currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*currUser = data.Applicant{}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	})
}

func GetAccountPageHandler(currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(currUser.Login)
		if currUser.Login == "" {
			log.Println("Authorization is required")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "templates/account.html")
	})
}

func GetAccountPostFormHandler(db *sql.DB, currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if currUser.Login == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		currUser.FirstName = r.FormValue("firstName")
		currUser.LastName = r.FormValue("lastName")
		currUser.MiddleName = r.FormValue("middleName")
		currUser.Country = r.FormValue("country")

		log.Println(currUser)
		currUser.Subjects = make(map[string]uint16)

		var subjectNames []string
		var subjectName string
		rows, err := db.Query("SELECT name_subject FROM subject")
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			rows.Scan(&subjectName)
			subjectNames = append(subjectNames, subjectName)
		}

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
		err = currUser.AddToDB()
		if err != nil {
			log.Println(err)
			http.Error(w, "Server error, unable to create your account.", 500)
			return
		}

		log.Println("Redirect to /programs")
		http.Redirect(w, r, "/program_choice", http.StatusSeeOther)
		log.Println("Redirect to /programs")
		return
	})
}

func GetProgramChoicePageHandler(currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		log.Println(currUser.GetAvailablePrograms())
		err = tmp.Execute(w, currUser.GetAvailablePrograms())
	})
}

func GetProgramChoicePostFormHandler(currUser *data.Applicant) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//if r.Method != "POST" {
		//	http.ServeFile(w, r, "templates/account.html")
		//}

		if currUser.Login == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		//applicant := new(data.Applicant)
		currUser.Programs = make(map[string]int)
		currUser.Programs[r.FormValue("program_choice_1")] = 1
		currUser.Programs[r.FormValue("program_choice_2")] = 2
		currUser.Programs[r.FormValue("program_choice_3")] = 3

		currUser.AddSelectedProgramsToDB()

		log.Println("Redirect to /programs")
		http.Redirect(w, r, "/programs", http.StatusSeeOther)
		log.Println("Redirect to /programs")
		return
	})
}
