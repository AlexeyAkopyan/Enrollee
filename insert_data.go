package main

// TODO: correct the insert data process

import (
	"bufio"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"strconv"
	"strings"
)

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

var departmentNames = []string{
	"Faculty of Mathematics and Mechanics",
	"Faculty of Arts",
	"Faculty of Biology",
	"Institute for Geosciences",
	"Institute of Chemistry",
	"Faculty of Mathematics and Computer Science",
	"Faculty of Applied Mathematics and Control Processes",
}

func parsePrograms() {
	var subject_ids map[string]int
	subject_ids = make(map[string]int)
	var subject_id int
	for _, subject := range subjectNames {
		_, err := db.Exec("INSERT INTO subject (name_subject) VALUES ($1)", subject)
		if err != nil {
			panic(err)
		}
		row := db.QueryRow("SELECT subject_id FROM subject WHERE name_subject = $1", subject)
		err = row.Scan(&subject_id)
		if err != nil {
			panic(err)
		}
		subject_ids[subject] = subject_id
	}

	log.Println("Subjects have successfully been added to data base")

	var department_ids map[string]int
	department_ids = make(map[string]int)
	var department_id int
	for _, department := range departmentNames {
		_, err := db.Exec("INSERT INTO department (name_department) VALUES ($1)", department)
		if err != nil {
			panic(err)
		}
		row := db.QueryRow("SELECT department_id FROM department WHERE name_department = $1", department)
		err = row.Scan(&department_id)
		if err != nil {
			panic(err)
		}
		department_ids[department] = department_id
	}
	log.Println("Departments have successfully been added to data base")

	f, _ := os.Open("programs")
	scanner := bufio.NewScanner(f)
	i := 0
	var program_id int
	for scanner.Scan() {
		if i%4 == 0 {
			splitted := strings.Split(scanner.Text(), "; ")
			plan, err := strconv.Atoi(splitted[2])
			if err != nil {
				panic(err)
			}
			_, err = db.Exec("INSERT INTO program (name_program, plan, department_id) VALUES($1, $2, $3)",
				splitted[0], plan, department_ids[splitted[1]])
			if err != nil {
				panic(err)
			}
			row := db.QueryRow("SELECT program_id FROM program WHERE name_program = $1", splitted[0])
			err = row.Scan(&program_id)
			if err != nil {
				panic(err)
			}

		} else {
			splitted := strings.Split(scanner.Text(), "; ")
			min_result, err := strconv.Atoi(splitted[1])
			if err != nil {
				panic(err)
			}
			subject_id = subject_ids[splitted[0]]
			_, err = db.Exec("INSERT INTO program_subject (program_id, subject_id, min_result, priority)"+
				" VALUES($1, $2, $3, $4)", program_id, subject_id, min_result, i%4)
			if err != nil {
				panic(err)
			}
		}
		i += 1
	}

	log.Println("Programs have successfully been added to data base")

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
