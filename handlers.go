package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Домашняя страница"))
}

func showEmp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	emp, err := readEmp(id)
	if err != nil {
		log.Println("Ошибка чтения работника")
		http.NotFound(w, r)
	} else {
		j, err := json.Marshal(emp)
		if err != nil {
			log.Println("Ошибка преобразования объекта работника")
		}

		w.Write([]byte(j))
	}

}

func showEmps(w http.ResponseWriter, r *http.Request) {

	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM employees")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var emp Employee
		err = rows.Scan(&emp.Id, &emp.Name, &emp.Surname, &emp.Phone, &emp.CompanyId, &emp.EmpDepartment.Name)
		if err != nil {
			log.Fatal(err)
		}

		query := fmt.Sprintf("SELECT * FROM departments WHERE name = '%s'", emp.EmpDepartment.Name)

		err = db.QueryRow(query).Scan(&emp.EmpDepartment.Name, &emp.EmpDepartment.Phone)
		if err != nil {
			log.Fatal(err)
		}

		query = fmt.Sprintf("SELECT * FROM passports WHERE employee_id = %d", emp.Id)

		err = db.QueryRow(query).Scan(&emp.EmpPassport.Number, &emp.EmpPassport.Type, &emp.Id)
		if err != nil {
			log.Fatal(err)
		}

		j, err := json.Marshal(emp)
		if err != nil {
			log.Fatal(err)
		}

		w.Write([]byte(j))
		w.Write([]byte("\n"))
	}
}

func createEmp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен", 405)
		return
	}

	emp, err := jsonToEmp(r.Body)
	if err != nil {
		log.Println("Ошибка преобразования тела запроса в объект работника")
		http.Error(w, "Неправильный формат тела запроса", 420)
		return
	}

	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	var id int
	query := fmt.Sprintf("INSERT INTO employees(name, surname, phone, company_id, department) VALUES ('%s', '%s', '%s', %d, '%s') RETURNING id",
		emp.Name, emp.Surname, emp.Phone, emp.CompanyId, emp.EmpDepartment.Name)
	err = db.QueryRow(query).Scan(&id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Неправильный формат тела запроса", 420)
		return
	}

	row := db.QueryRow(fmt.Sprintf("INSERT INTO passports VALUES ('%s', '%s', %d)", emp.EmpPassport.Number, emp.EmpPassport.Type, id))
	if row.Err() != nil {
		log.Println("Ошибка при добавлении паспорта")
		http.Error(w, "Неправильный формат тела запроса", 420)
		return
	}

	w.Write([]byte(fmt.Sprintf("Зарегестрирован рабочий c id %d", id)))
}

func removeEmp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	//fmt.Fprintf(w, "Удаление работника с ID %d\n", id)

	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("DELETE INTO employees WHERE id = %d", id))
	if row.Err() != nil {
		log.Println("Ошибка при удалении рабочего")
		http.NotFound(w, r)
	} else {
		fmt.Fprintf(w, "Рабочий был успешно удален")
	}

}

func updateEmp(w http.ResponseWriter, r *http.Request) {
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

	if newEmp.Name != "" && oldEmp.Name != newEmp.Name {
		oldEmp.Name = newEmp.Name
	}

	if newEmp.Surname != "" && oldEmp.Surname != newEmp.Surname {
		oldEmp.Surname = newEmp.Surname
	}

	if newEmp.Phone != "" && oldEmp.Phone != newEmp.Phone {
		oldEmp.Phone = newEmp.Phone
	}

	if newEmp.CompanyId != 0 && oldEmp.CompanyId != newEmp.CompanyId {
		oldEmp.CompanyId = newEmp.CompanyId
	}

	isNewPassport := false
	if newEmp.EmpPassport.Number != "" && oldEmp.EmpPassport.Number != newEmp.EmpPassport.Number {
		oldEmp.EmpPassport = newEmp.EmpPassport
		isNewPassport = true
	}

	if newEmp.EmpDepartment.Name != "" && oldEmp.EmpDepartment.Name != newEmp.EmpDepartment.Name {
		oldEmp.EmpDepartment.Name = newEmp.EmpDepartment.Name
	}

	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE employees SET name = '%s', surname = '%s', phone = '%s', company_id = %d, department = '%s' WHERE id = %d",
		oldEmp.Name, oldEmp.Surname, oldEmp.Phone, oldEmp.CompanyId, oldEmp.EmpDepartment.Name, oldEmp.Id)
	row := db.QueryRow(query)
	if row.Err() != nil {
		log.Println(err)
	}

	if isNewPassport {
		query := fmt.Sprintf("UPDATE passports SET passport_number = '%s', passport_type = '%s' WHERE employee_id = %d",
			oldEmp.EmpPassport.Number, oldEmp.EmpPassport.Type, oldEmp.Id)
		row := db.QueryRow(query)
		if row.Err() != nil {
			log.Println(err)
		}
	}

	w.Write([]byte("Рабочий успешно обновлен"))
}

func readEmp(id int) (emp Employee, err error) {
	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT * FROM employees WHERE id = %d", id))
	if row.Err() != nil {
		return
	}

	err = row.Scan(&emp.Id, &emp.Name, &emp.Surname, &emp.Phone, &emp.CompanyId, &emp.EmpDepartment.Name)
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

func jsonToEmp(r io.Reader) (emp Employee, err error) {
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
