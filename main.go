package main

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	dbName   = "gophone"
	user     = "postgres"
	password = "evolvus888#"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	db, err := sql.Open("postgres", psqlInfo)
	must(err)
	err = resetDb(db, dbName)
	must(err)
	db.Close()
	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbName)
	db, err = sql.Open("postgres", psqlInfo)
	must(err)

	defer db.Close()

	must(createPhoneNumberTable(db))
	_, err = insertPhoneNumber(db, "1234567890")
	must(err)
	_, err = insertPhoneNumber(db, "123 456 7891")
	must(err)
	_, err = insertPhoneNumber(db, "(123) 456 7892")
	must(err)
	_, err = insertPhoneNumber(db, "(123) 456-7893")
	must(err)
	_, err = insertPhoneNumber(db, "123-456-7894")
	must(err)
	id, err := insertPhoneNumber(db, "(123)456-7892")
	must(err)

	p, err := getPhone(db, id)
	must(err)
	fmt.Println(p)

	numbers, err := getAllNumbers(db)
	must(err)
	for _, p := range numbers {
		number := normalize(p.value)
		if number != p.value {
			fmt.Println("Updating or removing...", number)
			existing, err := findPhone(db, number)
			must(err)
			if existing != nil {
				must(deletePhone(db, p.id))
			} else {
				p.value = number
				must(updatePhone(db, p))
			}
		} else {
			fmt.Println("No changes required")
		}
	}

}
func insertPhoneNumber(db *sql.DB, phone string) (int, error) {
	sql := "insert into phone_numbers (value) values ($1) returning id"
	var id int
	err := db.QueryRow(sql, phone).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

type phone struct {
	id    int
	value string
}

func findPhone(db *sql.DB, number string) (*phone, error) {
	var p phone
	row := db.QueryRow("SELECT * FROM phone_numbers WHERE value=$1", number)
	err := row.Scan(&p.id, &p.value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func updatePhone(db *sql.DB, p phone) error {
	statement := `UPDATE phone_numbers SET value=$2 WHERE id=$1`
	_, err := db.Exec(statement, p.id, p.value)
	return err
}

func deletePhone(db *sql.DB, id int) error {
	statement := `DELETE FROM phone_numbers WHERE id=$1`
	_, err := db.Exec(statement, id)
	return err
}

func getAllNumbers(db *sql.DB) ([]phone, error) {
	var ret []phone
	rows, err := db.Query("select * from phone_numbers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p phone
		if err := rows.Scan(&p.id, &p.value); err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
func getPhone(db *sql.DB, id int) (string, error) {
	var number string
	err := db.QueryRow("select value from phone_numbers where id = $1", id).Scan(&number)
	if err != nil {
		return "", err
	}
	return number, nil
}

func createPhoneNumberTable(db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS phone_numbers(
			id SERIAL,
			value varchar(255)
		)
	`
	_, err := db.Exec(sql)
	return err
}

func createDb(db *sql.DB, name string) error {
	_, err := db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return err
	}
	return nil
}
func resetDb(db *sql.DB, name string) error {
	_, err := db.Exec("DROP DATABASE IF EXISTS " + name)
	if err != nil {
		return err
	}
	return createDb(db, name)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func normalize(phone string) string {

	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(phone, "")
}

// func normalize(phone string) string {

// 	re := regexp.MustCompile("\\D")
// 	return re.ReplaceAllString(phone, "")
// }

// func normalize(phone string) string {

// 	var buf bytes.Buffer
// 	for _, ch := range phone {
// 		if ch >= '0' && ch <= '9' {
// 			buf.WriteRune(ch)
// 		}
// 	}

// 	return buf.String()
// }
