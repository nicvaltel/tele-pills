package postgresql

import (
	"Pills/utls"
	"bufio"
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func getPostgresqlConnStr() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

func getPostgresqlConnStrForMigration() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

func openDB() *sql.DB { //TODO return error

	psqlconn := getPostgresqlConnStr()

	// open database
	db, err := sql.Open("postgres", psqlconn)
	utls.PanicError(err)

	// check db
	err = db.Ping()
	utls.PanicError(err)

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

func RunMigration(migrationFile string, migrationHashFile string) error {
	hash, err := computeMD5(migrationFile)
	if err != nil {
		return err
	}

	hashBytes := []byte(hash)

	hashFile, err := openOrCreateFile(migrationHashFile)
	if err != nil {
		return fmt.Errorf("Error opening hash file: %v", err)
	}
	defer hashFile.Close()

	hashFileBytes := make([]byte, len(hashBytes))
	_, err = hashFile.Read(hashFileBytes)
	if err != nil || !bytes.Equal(hashBytes, hashFileBytes) {
		// rewrite migrationHashFile with new hash
		_, err = hashFile.Write(hashBytes)
		if err != nil {
			return fmt.Errorf("Error writing hash to file: %v", err)
		}
		// run migration
		db := openDB()
		defer db.Close()

		err := execSqlFile(db, migrationFile)
		if err != nil {
			return err
		}
		log.Printf("Migration done.")
	} else {
		log.Printf("Migration ok - No need to run migration.")
	}

	return nil
}

func computeMD5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func openOrCreateFile(filename string) (*os.File, error) {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		// File does not exist, create a new one
		file, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		return file, nil
	} else if err == nil {
		// File exists, open it
		file, err := os.OpenFile(filename, os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return file, nil
	} else {
		// Some other error occurred
		return nil, err
	}
}
