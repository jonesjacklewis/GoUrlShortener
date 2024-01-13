package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl = template.Must(template.ParseFiles("./forms/shorten.gohtml"))
var databaseName = "./db/shorten.db"

func main() {
	createSqliteDBFile(databaseName)
	createDatabaseTables(databaseName)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./forms/index.html")
	})

	r.HandleFunc("/submit", handleSubmit)

	r.HandleFunc("/shorten", handleShorten)

	r.HandleFunc("/r/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		url := getLongURL(id)

		if url == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, url, 301)

	})

	http.ListenAndServe(":8080", r)

	fmt.Println("Server is running at port 8080")

}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		// throw 401 error
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// get form values
	url := r.FormValue("url")

	// insert url into database
	insertURL(url)

	// redirect to /shortern, with url as encoded query param
	http.Redirect(w, r, "/shorten?url="+url, 301)

	//http.Redirect(w, r, "/", 301)
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	// get query param
	url := r.URL.Query().Get("url")

	// generate short url
	shortURL := getShortURL(url)

	// render shorten.gohtml template
	tmpl.Execute(w, struct {
		ShortURL string
		LongURL  string
	}{
		ShortURL: shortURL,
		LongURL:  url,
	})
}

func getShortURL(url string) string {
	// get from database

	query := "SELECT id FROM shortenUrls WHERE url = ?"

	db, err := sql.Open("sqlite3", databaseName)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	rows, err := db.Query(query, url)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	var id int

	for rows.Next() {
		err = rows.Scan(&id)

		if err != nil {
			fmt.Println(err)
			return ""
		}
	}

	defer db.Close()

	return fmt.Sprintf("http://localhost:8080/r/%d", id)
}

func getLongURL(id string) string {
	// get from database

	query := "SELECT url FROM shortenUrls WHERE id = ?"

	db, err := sql.Open("sqlite3", databaseName)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	rows, err := db.Query(query, id)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	var url string

	for rows.Next() {
		err = rows.Scan(&url)

		if err != nil {
			fmt.Println(err)
			return ""
		}
	}

	defer db.Close()

	return url
}

func createSqliteDBFile(databaseName string) {

	// check if db folder exists
	if _, err := os.Stat("./db"); os.IsNotExist(err) {
		os.Mkdir("./db", 0755)
	}

	// check if database exists

	if _, err := os.Stat(databaseName); os.IsNotExist(err) {
		file, err := os.Create(databaseName)

		if err != nil {
			fmt.Println(err)
		}

		defer file.Close()
	}
}

func createDatabaseTables(databaseName string) {
	// read db/createTables.sql

	sqlFile, err := os.Open("./db/createTables.sql")

	if err != nil {
		fmt.Println(err)
		return
	}

	sqlFileContent, err := io.ReadAll(sqlFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	sqlString := string(sqlFileContent)

	// execute sql statements
	db, err := sql.Open("sqlite3", databaseName)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec(sqlString)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()
	defer sqlFile.Close()
}

func insertURL(url string) {
	db, err := sql.Open("sqlite3", databaseName)
	defer db.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	stmt, err := db.Prepare("INSERT OR IGNORE INTO shortenUrls(url) VALUES(?)")

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = stmt.Exec(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Inserted url into database")

	return

}
