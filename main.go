package main

import (
	"database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	dbHost     = "localhost" 
	dbPort     = "5432"      
	dbUser     = "postgres"  
	dbPassword = "admin"  
	dbName     = "sre-test" 
)

type Kebab struct {
    ID    	int 	`json:"id,omitempty"`
    Flavor	string 	`json:"flavor,omitempty"`
    Price 	int    	`json:"price,omitempty"`
}

var db *sql.DB

func init() {
	connection := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connection)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}
}

func main() {
    r := mux.NewRouter()

    r.HandleFunc("/kebabs", getKebabs).Methods("GET")
    r.HandleFunc("/kebabs/{id}", getKebab).Methods("GET")
    r.HandleFunc("/kebabs", createKebab).Methods("POST")
    r.HandleFunc("/kebabs/{id}", updateKebab).Methods("PUT")
    r.HandleFunc("/kebabs/{id}", deleteKebab).Methods("DELETE")

    log.Fatal(http.ListenAndServe(":8080", r))
}

func getKebabs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, flavor, price FROM kebabs")
	if err != nil {
		http.Error(w, "Failed to fetch kebabs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	kebabs := []Kebab{}
	for rows.Next() {
		var kebab Kebab
		err := rows.Scan(&kebab.ID, &kebab.Flavor, &kebab.Price)
		if err != nil {
			http.Error(w, "Failed to fetch kebabs", http.StatusInternalServerError)
			return
		}
		kebabs = append(kebabs, kebab)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to fetch kebabs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(kebabs)
}

func getKebab(w http.ResponseWriter, r *http.Request) {
	kebabID := mux.Vars(r)["id"]

	row := db.QueryRow("SELECT id, flavor, price FROM kebabs WHERE id = $1", kebabID)

	var kebab Kebab
	err := row.Scan(&kebab.ID, &kebab.Flavor, &kebab.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Kebab not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get kebab", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(kebab)
}

func createKebab(w http.ResponseWriter, r *http.Request) {
	var kebab Kebab
	json.NewDecoder(r.Body).Decode(&kebab)

	fmt.Println(kebab.Flavor)
	fmt.Println(kebab.Price)

	insertSQL := "INSERT INTO kebabs (flavor, price) VALUES ($1, $2);"
	_, err := db.Exec(insertSQL, kebab.Flavor, kebab.Price)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create kebab", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(kebab)
}

func updateKebab(w http.ResponseWriter, r *http.Request) {
	kebabID := mux.Vars(r)["id"]

	var kebab Kebab
	json.NewDecoder(r.Body).Decode(&kebab)

	updateSQL := "UPDATE kebabs SET flavor = $1, price = $2 WHERE id = $3;"
	_, err := db.Exec(updateSQL, kebab.Flavor, kebab.Price, kebabID)
	if err != nil {
		http.Error(w, "Failed to update kebab", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(kebab)
}

func deleteKebab(w http.ResponseWriter, r *http.Request) {
	kebabID := mux.Vars(r)["id"]

	deleteSQL := "DELETE FROM kebabs WHERE id = $1;"
	_, err := db.Exec(deleteSQL, kebabID)
	if err != nil {
		http.Error(w, "Failed to delete kebab", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}