package data

import (
	"database/sql"
	"log"
	"sort"
	"strconv"
	"strings"
)

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

type Programs struct {
	prDB     *sql.DB
	Programs []Program
	SortBy   string
	KeyWords []string
}

func (ps *Programs) SetDB(db *sql.DB) {
	ps.prDB = db
}

func (ps *Programs) GetPrograms() error {
	var orderBy string

	switch ps.SortBy {
	case "", "department":
		orderBy = "department"
	case "name":
		orderBy = "program"
	case "plan":
		orderBy = "plan"
	case "recommendation":
		orderBy = "program"
	}

	log.Println("Programs will be sorted by ", orderBy)

	for i, word := range ps.KeyWords {
		ps.KeyWords[i] = strings.ToLower(word)
	}

	ps.Programs = make([]Program, 0)

	rows, err := ps.prDB.Query("SELECT name_program, subjects, plan, name_department, min_results " +
		"FROM get_programs_order_by_" + orderBy + " ()")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var program Program
		var subjects, minResults string

		err := rows.Scan(&program.Name, &subjects, &program.Plan, &program.Department, &minResults)
		if err != nil {
			panic(err)
		}
		program.SetDB(ps.prDB)
		program.Subjects = make(map[string]uint16)
		for _, subject := range strings.Split(subjects, ";") {
			for _, s_minResult := range strings.Split(minResults, ";") {
				i_minResult, err := strconv.Atoi(s_minResult)
				if err != nil {
					panic(err)
				}
				program.Subjects[subject] = uint16(i_minResult)
			}
		}
		ps.Programs = append(ps.Programs, program)
	}
	if len(ps.Programs) == 0 {
		log.Println("WARNING: no program has been found")
		return nil
	}

	if len(ps.KeyWords) > 0 {
		var programWeights []int
		var returnProgram []Program
		var value int
		for _, program := range ps.Programs {
			value = 0
			for _, word := range ps.KeyWords {
				value += 3*Btoi(strings.Contains(strings.ToLower(program.Name), word)) +
					2*Btoi(strings.Contains(strings.ToLower(program.Department), word))
				for subject, _ := range program.Subjects {
					value += Btoi(strings.Contains(strings.ToLower(subject), word))
				}
			}
			log.Println(program.Name, value)
			if value > 0 {
				programWeights = append(programWeights, value)
				returnProgram = append(returnProgram, program)
			}
		}

		sort.SliceStable(returnProgram, func(i, j int) bool {
			if programWeights[i] > programWeights[j] {
				return true
			} else {
				return false
			}
		})
		ps.Programs = returnProgram
	}
	return nil
}


