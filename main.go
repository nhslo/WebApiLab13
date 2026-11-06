package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Employee struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Position   string  `json:"position"`
	Salary     float64 `json:"salary"`
	IsActive   bool    `json:"isActive"`
}

var database *sql.DB

func send(responseWriter http.ResponseWriter, status int, value any) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(status)
	json.NewEncoder(responseWriter).Encode(value)
}

func idFrom(responseWriter http.ResponseWriter, request *http.Request) (int, bool) {
	id, operationError := strconv.Atoi(request.PathValue("id"))
	if operationError != nil || id <= 0 {
		http.Error(responseWriter, "Invalid ID", 400)
		return 0, false
	}
	return id, true
}

func serverError(responseWriter http.ResponseWriter, operationError error) {
	log.Println(operationError)
	http.Error(responseWriter, "Database error", 500)
}

func valid(employee Employee) bool {
	return strings.TrimSpace(employee.Name) != "" && strings.TrimSpace(employee.Department) != "" && strings.TrimSpace(employee.Position) != "" && employee.Salary >= 0
}

// Получаем сотрудников из PostgreSQL и при необходимости фильтруем по отделу.
func list(responseWriter http.ResponseWriter, request *http.Request) {
	department := request.URL.Query().Get("department")
	rows, operationError := database.Query("SELECT id,name,department,position,salary,is_active FROM employees WHERE ($1 = '' OR department = $1) ORDER BY id", department)
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	defer rows.Close()
	items := []Employee{}
	for rows.Next() {
		var employee Employee
		if operationError = rows.Scan(&employee.ID, &employee.Name, &employee.Department, &employee.Position, &employee.Salary, &employee.IsActive); operationError != nil {
			serverError(responseWriter, operationError)
			return
		}
		items = append(items, employee)
	}
	if operationError = rows.Err(); operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	send(responseWriter, 200, items)
}

func get(responseWriter http.ResponseWriter, request *http.Request) {
	id, success := idFrom(responseWriter, request)
	if !success {
		return
	}
	var employee Employee
	operationError := database.QueryRow("SELECT id,name,department,position,salary,is_active FROM employees WHERE id=$1", id).Scan(&employee.ID, &employee.Name, &employee.Department, &employee.Position, &employee.Salary, &employee.IsActive)
	if errors.Is(operationError, sql.ErrNoRows) {
		http.NotFound(responseWriter, request)
		return
	}
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	send(responseWriter, 200, employee)
}

func create(responseWriter http.ResponseWriter, request *http.Request) {
	var employee Employee
	if json.NewDecoder(request.Body).Decode(&employee) != nil || !valid(employee) {
		http.Error(responseWriter, "Invalid employee", 400)
		return
	}
	operationError := database.QueryRow("INSERT INTO employees(name,department,position,salary,is_active) VALUES($1,$2,$3,$4,$5) RETURNING id", employee.Name, employee.Department, employee.Position, employee.Salary, employee.IsActive).Scan(&employee.ID)
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	send(responseWriter, 201, employee)
}

func update(responseWriter http.ResponseWriter, request *http.Request) {
	id, success := idFrom(responseWriter, request)
	if !success {
		return
	}
	var employee Employee
	if json.NewDecoder(request.Body).Decode(&employee) != nil || !valid(employee) {
		http.Error(responseWriter, "Invalid employee", 400)
		return
	}
	result, operationError := database.Exec("UPDATE employees SET name=$1,department=$2,position=$3,salary=$4,is_active=$5 WHERE id=$6", employee.Name, employee.Department, employee.Position, employee.Salary, employee.IsActive, id)
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	numberValue, _ := result.RowsAffected()
	if numberValue == 0 {
		http.NotFound(responseWriter, request)
		return
	}
	get(responseWriter, request)
}

func remove(responseWriter http.ResponseWriter, request *http.Request) {
	id, success := idFrom(responseWriter, request)
	if !success {
		return
	}
	result, operationError := database.Exec("DELETE FROM employees WHERE id=$1", id)
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	numberValue, _ := result.RowsAffected()
	if numberValue == 0 {
		http.NotFound(responseWriter, request)
		return
	}
	responseWriter.WriteHeader(204)
}

// Количество сотрудников и среднюю зарплату считает PostgreSQL.
func statistics(responseWriter http.ResponseWriter, request *http.Request) {
	var all, active int
	var average float64
	operationError := database.QueryRow("SELECT COUNT(*),COUNT(*) FILTER (WHERE is_active),COALESCE(AVG(salary),0) FROM employees").Scan(&all, &active, &average)
	if operationError != nil {
		serverError(responseWriter, operationError)
		return
	}
	send(responseWriter, 200, map[string]any{"total": all, "active": active, "averageSalary": average})
}

func main() {
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("Set DATABASE_URL")
	}
	var operationError error
	database, operationError = sql.Open("pgx", connectionString)
	if operationError != nil {
		log.Fatal(operationError)
	}
	defer database.Close()
	if operationError = database.Ping(); operationError != nil {
		log.Fatal(operationError)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees", list)
	mux.HandleFunc("GET /employees/{id}", get)
	mux.HandleFunc("GET /employees/statistics", statistics)
	mux.HandleFunc("POST /employees", create)
	mux.HandleFunc("PUT /employees/{id}", update)
	mux.HandleFunc("DELETE /employees/{id}", remove)
	log.Println("http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
