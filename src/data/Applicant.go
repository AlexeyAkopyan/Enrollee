package data

import (
	"database/sql"
	"log"
)

type Applicant struct {
	Id                                              int
	prDB                                            *sql.DB
	FirstName, LastName, MiddleName, Country, Login string
	Subjects                                        map[string]uint16 // subject name -> score
	Programs                                        map[string]int    // program name -> priority
	prAvailablePrograms                             []string
	prNoAvailableProgram                            bool
}

func (ap *Applicant) SetDB(db *sql.DB) {
	ap.prDB = db
}

func (ap Applicant) AddToDB() error {
	_, err := ap.prDB.Exec("UPDATE enrollee SET first_name = $1, last_name = $2, middle_name = $3, country = $4"+
		" WHERE login = $5", ap.FirstName, ap.LastName, ap.MiddleName, ap.Country, ap.Login)
	if err != nil {
		return err
	}

	err = ap.prDB.QueryRow("SELECT enrollee_id FROM enrollee WHERE login = $1", ap.Login).Scan(&ap.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(ap.Id)
	log.Println("New applicant has successfully been added")

	ap.prDB.Exec("DELETE FROM enrollee_subject WHERE enrollee_id = $1", ap.Id)
	if err != nil {
		panic(err)
	}

	for subject, score := range ap.Subjects {
		_, err = ap.prDB.Exec("INSERT INTO enrollee_subject (enrollee_id, subject_id, result) "+
			"SELECT $1, subject_id, $2 FROM subject WHERE name_subject = $3", ap.Id, score, subject)
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

func (ap *Applicant) AddSelectedProgramsToDB() {

	err := ap.prDB.QueryRow("SELECT enrollee_id FROM enrollee WHERE login = $1", ap.Login).Scan(&ap.Id)
	if err != nil {
		panic(err)
	}

	ap.prDB.Exec("DELETE FROM program_enrollee WHERE enrollee_id = $1", ap.Id)
	if err != nil {
		panic(err)
	}
	for program, priority := range ap.Programs {
		log.Println(program, priority)
		_, err := ap.prDB.Exec("INSERT INTO program_enrollee (program_id, enrollee_id, priority) "+
			"SELECT program_id, $1, $2 FROM program WHERE name_program = $3", ap.Id, priority, program)
		if err != nil {
			panic(err)
		}
	}
}
