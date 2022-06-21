package src

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

type Applicant struct {
	prDB                                                *sql.DB
	Id, FirstName, LastName, MiddleName, Country, Login string
	Subjects                                            map[string]uint16 // subject name -> score
	Programs                                            map[string]uint16 // program name -> priority
	prAvailablePrograms                                 []string
	prNoAvailableProgram                                bool
}

func (ap *Applicant) SetDB(db *sql.DB) {
	ap.prDB = db
	return
}

func (ap Applicant) AddToDB(db *sql.DB) error {
	_, err := db.Exec("UPDATE enrollee SET first_name = $1, last_name = $2, middle_name = $3, country = $4"+
		" WHERE login = $5", ap.FirstName, ap.LastName, ap.MiddleName, ap.Country, ap.Login)
	if err != nil {
		return err
	}
	var enrollee_id int

	err = db.QueryRow("SELECT enrollee_id FROM enrollee WHERE login = $1", ap.Login).Scan(&enrollee_id)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(enrollee_id)
	log.Println("New applicant has successfully been added")

	for subject, score := range ap.Subjects {
		_, err = db.Exec("INSERT INTO enrollee_subject (enrollee_id, subject_id, result) "+
			"SELECT $1, subject_id, $2 FROM subject WHERE name_subject = $3", enrollee_id, score, subject)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	log.Println("New applicant's subject scores have successfully been added")

	return err
}

func (ap *Applicant) GetAvailablePrograms() []string {
	if (!ap.prNoAvailableProgram) && len(ap.prAvailablePrograms) == 0 {

		rows, err := ap.prDB.Query("SELECT name_program FROM get_available_to_enrollee_programs ($1)", ap.Login)
		if err != nil {
			panic(err)
		}
		var availProgram string
		for rows.Next() {
			err = rows.Scan(&availProgram)
			if err != nil {
				panic(err)
			}
			ap.prAvailablePrograms = append(ap.prAvailablePrograms, availProgram)
		}
		log.Println("Found ", len(ap.prAvailablePrograms), "programs")
	}
	return ap.prAvailablePrograms
}

type Enrollee struct {
	Id, FirstName, LastName, MiddleName, Verdict string
	TotalResult                                  uint16
}

func (en Enrollee) GetFullName() string {
	fullName := en.LastName + " " + en.FirstName
	if len(en.MiddleName) > 0 {
		fullName = fullName + " " + en.MiddleName
	}
	return fullName
}

type Program struct {
	Name, Department string
	Subjects         map[string]uint16
	Plan             uint16
	Enrollees        []Enrollee
}

func (p *Program) GetUrl() string {
	return "programs/" + strings.Replace(strings.ToLower(p.Name), " ", "_", -1)
}

type Programs struct {
	Programs []Program
}

func (ps *Programs) GetPrograms(db *sql.DB) error {
	rows, err := db.Query("SELECT name_program, subjects, plan, name_department, min_results " +
		"FROM get_programs ()")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var program Program
		var subjects, minResults string
		err := rows.Scan(&program.Name, &subjects, &program.Plan, &program.Department, &minResults)
		if err != nil {
			return err
		}
		program.Subjects = make(map[string]uint16)
		for _, subject := range strings.Split(subjects, ";") {
			for _, s_minResult := range strings.Split(minResults, ";") {
				i_minResult, err := strconv.Atoi(s_minResult)
				if err != nil {
					return err
				}
				program.Subjects[subject] = uint16(i_minResult)
			}
		}
		ps.Programs = append(ps.Programs, program)
		//log.Println(ps.Programs, "added to program list")
	}
	return nil
}

func (p *Program) GetEnrollees(db *sql.DB) error {
	p.Enrollees = make([]Enrollee, 0)
	rows, err := db.Query("SELECT first_name, last_name, middle_name, total_result "+
		"FROM get_enrollees_by_program_name ($1)", p.Name)
	if err != nil {
		log.Println("Error 1")
		return err
	}

	for rows.Next() {
		var enrollee Enrollee
		err = rows.Scan(&enrollee.FirstName, &enrollee.LastName, &enrollee.MiddleName, &enrollee.TotalResult)
		if err != nil {
			log.Println("Error appeared while scanning an enrollee")
			return err
		}
		enrollee.Verdict = "Recommended to enrolment"
		p.Enrollees = append(p.Enrollees, enrollee)
	}
	return nil
}
