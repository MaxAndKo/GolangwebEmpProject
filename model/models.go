package model

type Employee struct {
	Id            int
	Name          string
	Surname       string
	Phone         string
	CompanyId     int
	EmpPassport   Passport
	EmpDepartment Department
}

type Passport struct {
	Type   string
	Number string
}

type Department struct {
	Name  string
	Phone string
}
