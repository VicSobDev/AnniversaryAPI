package db

import "time"

type Image struct {
	ID         int       `json:"id,omitempty"`
	UploadedBy int       `json:"uploaded_by,omitempty"`
	Name       string    `json:"name,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

func (s *SQLiteDB) GetImages() ([]Image, error) {
	rows, err := s.db.Query("SELECT id, uploaded_by, name, created_at FROM images")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		err := rows.Scan(&image.ID, &image.UploadedBy, &image.Name, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func (s *SQLiteDB) GetImagesPaginated(limit, offset int) ([]Image, error) {
	rows, err := s.db.Query("SELECT id, uploaded_by, name, created_at FROM images LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		err := rows.Scan(&image.ID, &image.UploadedBy, &image.Name, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func (s *SQLiteDB) CreateImage(uploadedBy int, name string, createdAt int64) error {
	_, err := s.db.Exec("INSERT INTO images (uploaded_by, name, created_at) VALUES (?, ?, ?)", uploadedBy, name, createdAt)
	return err
}

func (s *SQLiteDB) DeleteImage(id int) error {
	_, err := s.db.Exec("DELETE FROM images WHERE id = ?", id)
	return err
}

func (s *SQLiteDB) GetImage(id int) (Image, error) {
	var image Image
	err := s.db.QueryRow("SELECT id, uploaded_by, name, created_at FROM images WHERE id = ?", id).Scan(&image.ID, &image.UploadedBy, &image.Name, &image.CreatedAt)
	return image, err
}

func (s *SQLiteDB) GetImagesByUser(userID int) ([]Image, error) {
	rows, err := s.db.Query("SELECT id, uploaded_by, name, created_at FROM images WHERE uploaded_by = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		err := rows.Scan(&image.ID, &image.UploadedBy, &image.Name, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}
