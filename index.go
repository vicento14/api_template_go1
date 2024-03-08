package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
)

// var db *sql.DB

type UserAccounts struct {
	Id       int    `json:"Id"`
	IdNumber string `json:"IdNumber"`
	FullName string `json:"FullName"`
	Username string `json:"Username"`
	Password string `json:"Password"`
	Section  string `json:"Section"`
	Role     string `json:"Role"`
}

// Database Connection
func connectToDatabase() (*sql.DB, error) {
	// Capture connection properties.
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "web_template",
		AllowNativePasswords: true,
	}

	// Open the database connection.
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure the connection is established.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// UserAccounts Functions

func getUserAccounts(db *sql.DB) ([]UserAccounts, error) {
	// A user_accounts slice to hold data from returned rows.
	var user_accounts []UserAccounts

	rows, err := db.Query("SELECT * FROM user_accounts")
	if err != nil {
		return nil, fmt.Errorf("getUserAccounts : %v", err)
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var ua UserAccounts
		if err := rows.Scan(&ua.Id, &ua.IdNumber, &ua.FullName, &ua.Username, &ua.Password, &ua.Section, &ua.Role); err != nil {
			return nil, fmt.Errorf("getUserAccounts : %v", err)
		}
		user_accounts = append(user_accounts, ua)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getUserAccounts : %v", err)
	}
	return user_accounts, nil
}

func getUserAccountsSearch(id_number string, full_name string, role string, db *sql.DB) ([]UserAccounts, error) {
	// A user_accounts slice to hold data from returned rows.
	var user_accounts []UserAccounts

	// Start building the query.
	query := "SELECT * FROM user_accounts WHERE 1=1"

	// Slice to hold the arguments for the query.
	var args []interface{}

	// Add conditions to the query based on the parameters.
	if id_number != "" {
		query += " AND id_number LIKE ?"
		id_number = id_number + "%"
		args = append(args, id_number)
	}
	if full_name != "" {
		query += " AND full_name LIKE ?"
		full_name = full_name + "%"
		args = append(args, full_name)
	}
	if role != "" {
		query += " AND role = ?"
		args = append(args, role)
	}

	// Prepare the statement.
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getUserAccountsSearch: %v", err)
	}
	defer stmt.Close()

	// Execute the query.
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("getUserAccountsSearch : %v", err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var ua UserAccounts
		if err := rows.Scan(&ua.Id, &ua.IdNumber, &ua.FullName, &ua.Username, &ua.Password, &ua.Section, &ua.Role); err != nil {
			return nil, fmt.Errorf("getUserAccountsSearch : %v", err)
		}
		user_accounts = append(user_accounts, ua)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getUserAccountsSearch : %v", err)
	}
	return user_accounts, nil
}

func getUserAccountsById(id int64, db *sql.DB) (UserAccounts, error) {
	// A ua to hold data from the returned row.
	var ua UserAccounts

	row := db.QueryRow("SELECT * FROM user_accounts WHERE id = ?", id)
	if err := row.Scan(&ua.Id, &ua.IdNumber, &ua.FullName, &ua.Username, &ua.Password, &ua.Section, &ua.Role); err != nil {
		if err == sql.ErrNoRows {
			return ua, fmt.Errorf("getUserAccountsById %d: no such UserAccount", id)
		}
		return ua, fmt.Errorf("getUserAccountsById %d: %v", id, err)
	}
	return ua, nil
}

func countUserAccounts(id_number string, full_name string, role string, db *sql.DB) (int, error) {
	// A variable to hold the count.
	var count int

	// Start building the query.
	query := "SELECT COUNT(*) FROM user_accounts WHERE 1=1"

	// Slice to hold the arguments for the query.
	var args []interface{}

	// Add conditions to the query based on the parameters.
	if id_number != "" {
		query += " AND id_number LIKE ?"
		id_number = id_number + "%"
		args = append(args, id_number)
	}
	if full_name != "" {
		query += " AND full_name LIKE ?"
		full_name = full_name + "%"
		args = append(args, full_name)
	}
	if role != "" {
		query += " AND role = ?"
		args = append(args, role)
	}

	// Prepare the statement.
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("countUserAccounts: %v", err)
	}
	defer stmt.Close()

	// Execute the query.
	row := stmt.QueryRow(args...)
	err = row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("countUserAccounts: %v", err)
	}

	return count, nil
}

func insertUserAccount(ua UserAccounts, db *sql.DB) (int64, error) {
	result, err := db.Exec("INSERT INTO user_accounts (id_number, full_name, username, password, section, role) VALUES (?, ?, ?, ?, ?, ?)", ua.IdNumber, ua.FullName, ua.Username, ua.Password, ua.Section, ua.Role)
	if err != nil {
		return 0, fmt.Errorf("insertUserAccount: %v", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("insertUserAccount: %v", err)
	}
	return inserted, nil
}

func updateUserAccount(ua UserAccounts, db *sql.DB) (int64, error) {
	result, err := db.Exec("UPDATE user_accounts SET id_number = ?, full_name = ?, username = ?, password = ?, section = ?, role = ? WHERE id = ?", ua.IdNumber, ua.FullName, ua.Username, ua.Password, ua.Section, ua.Role, ua.Id)
	if err != nil {
		return 0, fmt.Errorf("updateUserAccount: %v", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("updateUserAccount: %v", err)
	}
	return updated, nil
}

func deleteUserAccount(id int, db *sql.DB) (int64, error) {
	result, err := db.Exec("DELETE FROM user_accounts WHERE id = ?", id)
	if err != nil {
		return 0, fmt.Errorf("deleteUserAccount: %v", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("deleteUserAccount: %v", err)
	}
	return deleted, nil
}

// Database Middleware
func dbMiddleware(f func(w http.ResponseWriter, r *http.Request, db *sql.DB)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := connectToDatabase()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error connecting to the database")
			return
		}
		f(w, r, db)
	}
}

// Route Functions

func userAccountsRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	user_accounts, err := getUserAccounts(db)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// fmt.Fprint(w, user_accounts)
	json.NewEncoder(w).Encode(user_accounts)
}

func userAccountsSearchRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse the query parameters from the URL
	query := r.URL.Query()

	id_number := query.Get("id_number")
	full_name := query.Get("full_name")
	role := query.Get("role")

	user_accounts, err := getUserAccountsSearch(id_number, full_name, role, db)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user_accounts)
}

func userAccountsCountRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse the query parameters from the URL
	query := r.URL.Query()

	id_number := query.Get("id_number")
	full_name := query.Get("full_name")
	role := query.Get("role")

	count, err := countUserAccounts(id_number, full_name, role, db)
	if err != nil {
		log.Print(err)
	}

	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, strconv.Itoa(count))
}

func userAccountsIdRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse the query parameters from the URL
	query := r.URL.Query()

	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect ID")
		return
	}

	user_account, err := getUserAccountsById(id, db)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user_account)
}

func userAccountsInsertRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		// Handle POST request data
		user_account := new(UserAccounts)

		if err := json.NewDecoder(r.Body).Decode(&user_account); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect UserAccounts")
			return
		}

		message := ""

		inserted, err := insertUserAccount(UserAccounts{
			Id:       0,
			IdNumber: user_account.IdNumber,
			FullName: user_account.FullName,
			Username: user_account.Username,
			Password: user_account.Password,
			Section:  user_account.Section,
			Role:     user_account.Role,
		}, db)
		if err != nil {
			log.Print(err)
		}

		if inserted > 0 {
			message = "success"
		}

		fmt.Fprint(w, message)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST requests are allowed on this route"))
		fmt.Fprint(w)
	}
}

func userAccountsUpdateRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		// Handle POST request data
		user_account := new(UserAccounts)

		if err := json.NewDecoder(r.Body).Decode(&user_account); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect UserAccounts")
			return
		}

		message := ""

		updated, err := updateUserAccount(UserAccounts{
			Id:       user_account.Id,
			IdNumber: user_account.IdNumber,
			FullName: user_account.FullName,
			Username: user_account.Username,
			Password: user_account.Password,
			Section:  user_account.Section,
			Role:     user_account.Role,
		}, db)
		if err != nil {
			log.Print(err)
		}

		if updated > 0 {
			message = "success"
		}

		fmt.Fprint(w, message)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST requests are allowed on this route"))
		fmt.Fprint(w)
	}
}

func userAccountsDeleteRoute(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		// Handle POST request data
		user_account := new(UserAccounts)

		if err := json.NewDecoder(r.Body).Decode(&user_account); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect UserAccounts")
			return
		}

		message := ""

		deleted, err := deleteUserAccount(user_account.Id, db)
		if err != nil {
			log.Print(err)
		}

		if deleted > 0 {
			message = "success"
		}

		fmt.Fprint(w, message)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST requests are allowed on this route"))
		fmt.Fprint(w)
	}
}

// Main Function

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlElement := "<p>"
		htmlElement += "API TEMPLATE GO 1 (localhost:914)"
		htmlElement += "</p>"
		htmlElement += "<p>"
		htmlElement += "Developed By : Vince Dale D. Alcantara"
		htmlElement += "</p>"
		htmlElement += "<p>"
		htmlElement += "Version 1.0.0"
		htmlElement += "</p>"

		fmt.Fprint(w, htmlElement)
	})

	mux.HandleFunc("/UserAccounts", dbMiddleware(userAccountsRoute))
	mux.HandleFunc("/UserAccounts/Search", dbMiddleware(userAccountsSearchRoute))
	mux.HandleFunc("/UserAccounts/Count", dbMiddleware(userAccountsCountRoute))
	mux.HandleFunc("/UserAccounts/Id", dbMiddleware(userAccountsIdRoute))
	mux.HandleFunc("/UserAccounts/Insert", dbMiddleware(userAccountsInsertRoute))
	mux.HandleFunc("/UserAccounts/Update", dbMiddleware(userAccountsUpdateRoute))
	mux.HandleFunc("/UserAccounts/Delete", dbMiddleware(userAccountsDeleteRoute))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	})

	// Set up CORS middleware with default options (all origins accepted with simple methods)
	handler := c.Handler(mux)

	http.ListenAndServe(":914", handler)
}
