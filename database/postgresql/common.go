package postgresql

import (
	"Pills/utls"
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func openDB() *sql.DB {

	host := os.Getenv("POSTGRES_HOST")

	port := os.Getenv("POSTGRES_PORT")

	user := os.Getenv("POSTGRES_USER")

	password := os.Getenv("POSTGRES_PASSWORD")

	dbname := os.Getenv("POSTGRES_DB")

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	utls.CheckError(err)

	// check db
	err = db.Ping()
	utls.CheckError(err)

	log.Printf("Connected to database!")

	return db
}

func parseSQLFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var queries []string
	scanner := bufio.NewScanner(file)
	query := ""
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "--") || strings.TrimSpace(line) == "" {
			continue
		}

		query += line + " "

		// Check if the line ends with a semicolon
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			queries = append(queries, query)
			query = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

func execSqlFile(db *sql.DB, filePath string) error {
	queryes, err := parseSQLFile(filePath)
	if err != nil {
		return err
	}

	for _, q := range queryes {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunMigration() { // TODO rewrite to ordinary migration
	db := openDB()
	defer db.Close()
	err := execSqlFile(db, "database/migration/PostgreSQL/Migrations_00_Create_Tables/0001_create.sql")
	if err != nil {
		panic(err)
	}
}
