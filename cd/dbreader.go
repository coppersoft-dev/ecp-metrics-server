package cd

import (
	"database/sql"
	"fmt"
)

func GetCDContent(username, password, dbHost, dbName string) ([]byte, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=require", username, password, dbHost, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed opening DB connection: %w", err)
	}

	row := db.QueryRow("select convert_from(lo_get(cd.directory_content), 'UTF-8') from component_directory cd")
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("failed querying for content: %w", err)
	}
	var res []byte
	if err := row.Scan(&res); err != nil {
		return nil, fmt.Errorf("failed scanning query result: %w", err)
	}

	return res, err
}
