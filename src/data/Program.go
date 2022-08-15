package data

import (
	"database/sql"
	"log"
	"strings"
)

type Program struct {
	prDB             *sql.DB
	Name, Department string
	Subjects         map[string]uint16
	Plan             uint16
	Enrollees        []Enrollee
}

func (p *Program) GetUrl() string {
	return "programs/" + strings.Replace(strings.ToLower(p.Name), " ", "_", -1)
}

func (p *Program) SetDB(db *sql.DB) {
	p.prDB = db
}

func (p *Program) GetEnrollees() error {
	p.Enrollees = make([]Enrollee, 0)
	rows, err := p.prDB.Query("SELECT first_name, last_name, middle_name, total_result "+
		"FROM get_enrollees_by_program_name ($1)", p.Name)
	if err != nil {
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
