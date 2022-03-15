package controller

import (
	"database/sql"
	"empProject/model"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func ConnectDB() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return
}

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	temp, err := template.ParseFiles("web/home_page.html")
	if err != nil {
		log.Println("Ошибка парсинга домашней страницы")
		http.Error(w, "Ошибка веб приложения", 500)
		return
	}
	temp.Execute(w, nil)
}

func ShowEmp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.Error(w, "Работника с таким id не существует", 404)
		return
	}

	emp, err := readEmp(id)
	if err != nil {
		log.Println("Ошибка чтения работника")
		http.Error(w, "Работника с таким id не существует", 404)
	} else {
		j, err := json.Marshal(emp)
		if err != nil {
			log.Println("Ошибка преобразования объекта работника")
			http.Error(w, "Ошибка преобразования в json", 400)
		}

		w.Write([]byte(j))
	}

}

func ShowEmps(w http.ResponseWriter, r *http.Request) {

	db, err := ConnectDB()
	if err != nil {
		log.Println("Ошибка открытия соединения с БД")
		http.Error(w, "Ошибка соединения приложения с БД", 500)
		db.Close()
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM employees")
	if err != nil {
		log.Println("Ошибка чтения работников")
		http.Error(w, "Внутрення ошибка сервера", 500)
		return
	}

	for rows.Next() {
		var emp model.Employee
		err = rows.Scan(&emp.Id, &emp.Name, &emp.Surname, &emp.Phone, &emp.CompanyId, &emp.EmpDepartment.Name)
		if err != nil {
			log.Println("Ошибка преобразования работника")
			http.Error(w, "Внутрення ошибка сервера", 500)
			return
		}

		query := fmt.Sprintf("SELECT * FROM departments WHERE name = '%s'", emp.EmpDepartment.Name)

		err = db.QueryRow(query).Scan(&emp.EmpDepartment.Name, &emp.EmpDepartment.Phone)
		if err != nil {
			log.Println("Ошибка чтения департамента")
			http.Error(w, "Внутрення ошибка сервера", 500)
			return
		}

		query = fmt.Sprintf("SELECT * FROM passports WHERE employee_id = %d", emp.Id)

		err = db.QueryRow(query).Scan(&emp.EmpPassport.Number, &emp.EmpPassport.Type, &emp.Id)
		if err != nil {
			log.Println("Ошибка чтения паспорта")
			http.Error(w, "Внутрення ошибка сервера", 500)
			return
		}

		j, err := json.Marshal(emp)
		if err != nil {
			log.Fatal(err)
		}

		w.Write([]byte(j))
		w.Write([]byte("\n"))
	}
}

func CreateEmp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен", 405)
		return
	}

	emp, err := jsonToEmp(r.Body)
	if err != nil {
		log.Println("Ошибка преобразования тела запроса в объект работника")
		http.Error(w, "Неправильный формат тела запроса", 400)
		return
	}
	if !checkEmp(emp) {
		log.Println("Ошибка заполнения тела запроса")
		http.Error(w, "Ошибка заполнения тела запроса", 400)
		return
	}

	db, err := ConnectDB()
	if err != nil {
		log.Println("Ошибка открытия соединения с БД")
		http.Error(w, "Ошибка соединения приложения с БД", 500)
		db.Close()
		return
	}
	defer db.Close()

	if !checkPassport(emp.EmpPassport.Number, db) {
		log.Println("Работник с таким паспортом уже существует")
		http.Error(w, "Работник с таким паспортом уже существует", 400)
		return
	}

	var id int
	query := fmt.Sprintf("INSERT INTO employees(name, surname, phone, company_id, department) VALUES ('%s', '%s', '%s', %d, '%s') RETURNING id",
		emp.Name, emp.Surname, emp.Phone, emp.CompanyId, emp.EmpDepartment.Name)
	err = db.QueryRow(query).Scan(&id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Неправильный формат тела запроса", 400)
		return
	}

	query = fmt.Sprintf("INSERT INTO passports VALUES ('%s', '%s', %d)", emp.EmpPassport.Number, emp.EmpPassport.Type, id)
	_, err = db.Exec(query)
	if err != nil {
		log.Print("Ошибка при добавлении паспорта:")
		log.Println(err)
		http.Error(w, "Неправильный формат тела запроса", 420)
		return
	}

	w.Write([]byte(fmt.Sprintf("Зарегистрирован рабочий c id %d", id)))

}

func RemoveEmp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	//fmt.Fprintf(w, "Удаление работника с ID %d\n", id)

	db, err := ConnectDB()
	if err != nil {
		log.Println("Ошибка открытия соединения с БД")
		http.Error(w, "Ошибка соединения приложения с БД", 500)
		db.Close()
		return
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM employees WHERE id = %d", id))
	if err != nil {
		log.Print("Ошибка при удалении рабочего: ")
		log.Println(err)
		http.Error(w, "Работника с таким id не существует", 404)
	} else {
		fmt.Fprintf(w, "Рабочий был успешно удален")
	}

}

func UpdateEmp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен", 405)
		return
	}

	newEmp, err := jsonToEmp(r.Body)
	if err != nil {
		log.Println("Ошибка преобразования тела запроса в объект работника")
		http.Error(w, "Неправильный формат тела запроса", 420)
		return
	}

	oldEmp, err := readEmp(newEmp.Id)
	if err != nil {
		log.Println("Ошибка чтения работника")
		http.NotFound(w, r)
		return
	}

	if newEmp.Name != "" {
		oldEmp.Name = newEmp.Name
	}

	if newEmp.Surname != "" {
		oldEmp.Surname = newEmp.Surname
	}

	if newEmp.Phone != "" {
		oldEmp.Phone = newEmp.Phone
	}

	if newEmp.CompanyId != 0 {
		oldEmp.CompanyId = newEmp.CompanyId
	}

	isNewPassport := false
	if newEmp.EmpPassport.Number != "" {
		oldEmp.EmpPassport = newEmp.EmpPassport
		isNewPassport = true
	}

	if newEmp.EmpDepartment.Name != "" {
		oldEmp.EmpDepartment.Name = newEmp.EmpDepartment.Name
	}

	db, err := ConnectDB()
	if err != nil {
		log.Println("Ошибка открытия соединения с БД")
		http.Error(w, "Ошибка соединения приложения с БД", 500)
		db.Close()
		return
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE employees SET name = '%s', surname = '%s', phone = '%s', company_id = %d, department = '%s' WHERE id = %d",
		oldEmp.Name, oldEmp.Surname, oldEmp.Phone, oldEmp.CompanyId, oldEmp.EmpDepartment.Name, oldEmp.Id)
	_, err = db.Exec(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка приложения", 500)
		return
	}

	if isNewPassport {
		query := fmt.Sprintf("UPDATE passports SET passport_number = '%s', passport_type = '%s' WHERE employee_id = %d",
			oldEmp.EmpPassport.Number, oldEmp.EmpPassport.Type, oldEmp.Id)
		_, err := db.Exec(query)
		if err != nil {
			log.Println(err)
			http.Error(w, "Ошибка приложения", 500)
			return
		}
	}

	w.Write([]byte("Рабочий успешно обновлен"))
}

func readEmp(id int) (emp model.Employee, err error) {
	db, err := ConnectDB()
	if err != nil {
		return
	}
	defer db.Close()

	err = db.QueryRow(fmt.Sprintf("SELECT * FROM employees WHERE id = %d", id)).Scan(&emp.Id, &emp.Name, &emp.Surname, &emp.Phone, &emp.CompanyId, &emp.EmpDepartment.Name)
	if err != nil {
		return
	}

	query := fmt.Sprintf("SELECT * FROM departments WHERE name = '%s'", emp.EmpDepartment.Name)

	err = db.QueryRow(query).Scan(&emp.EmpDepartment.Name, &emp.EmpDepartment.Phone)
	if err != nil {
		return
	}

	query = fmt.Sprintf("SELECT * FROM passports WHERE employee_id = %d", emp.Id)

	err = db.QueryRow(query).Scan(&emp.EmpPassport.Number, &emp.EmpPassport.Type, &emp.Id)
	if err != nil {
		return
	}

	return
}

func jsonToEmp(r io.Reader) (emp model.Employee, err error) {
	bodyBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	err = json.Unmarshal(bodyBytes, &emp)
	if err != nil {
		return
	}
	return
}

func checkPassport(number string, db *sql.DB) bool {
	query := fmt.Sprintf("SELECT * FROM passports WHERE passport_number = '%s'", number)

	var passportType string
	var id int
	err := db.QueryRow(query).Scan(&number, &passportType, &id)
	if err != nil {
		return true
	} else {
		return false
	}

}

func checkEmp(emp model.Employee) bool {
	if emp.Name == "" || emp.Surname == "" || emp.CompanyId == 0 || emp.Phone == "" || emp.EmpDepartment.Name == "" || emp.EmpPassport.Number == "" || emp.EmpPassport.Type == "" {
		return false
	}
	return true
}
