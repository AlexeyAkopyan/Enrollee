package data

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
