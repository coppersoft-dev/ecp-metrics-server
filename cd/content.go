package cd

import (
	"fmt"
)

func (s *Service) GetCDContent() ([]byte, error) {
	row := s.db.QueryRow("select convert_from(lo_get(component_directory.directory_content), 'UTF-8') from component_directory")
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("failed querying for content: %w", err)
	}
	var res []byte
	if err := row.Scan(&res); err != nil {
		return nil, fmt.Errorf("failed scanning query result: %w", err)
	}

	return res, nil
}
