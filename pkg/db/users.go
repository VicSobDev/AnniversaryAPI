package db

type User struct {
	ID       int
	Username string
	Password string
}

func (s *SQLiteDB) CreateUser(username, password string) (*User, error) {
	_, err := s.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		return nil, err
	}
	return s.GetUser(username)
}

func (s *SQLiteDB) GetUser(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
	return user, err
}

func (s *SQLiteDB) CreateKey(key string) error {
	_, err := s.db.Exec("INSERT INTO keys (key) VALUES (?)", key)
	return err
}

func (s *SQLiteDB) ContainsKey(key string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM keys WHERE key = ?", key).Scan(&count)
	return count > 0, err
}

func (s *SQLiteDB) DeleteKey(key string) error {
	_, err := s.db.Exec("DELETE FROM keys WHERE key = ?", key)
	return err
}

func (s *SQLiteDB) GetKeys() ([]string, error) {
	rows, err := s.db.Query("SELECT key FROM keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *SQLiteDB) AddKey(key string) error {
	_, err := s.db.Exec("INSERT INTO keys (key) VALUES (?)", key)
	return err
}
