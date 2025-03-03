package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//TODO: Надо бы накрутить транзакции
//При удалении файла нужен лок на время удаления

type IStorage interface {
	SaveImage(imageName string, size int, mimeType string, checksum string, createdAt time.Time) error
	ListFiles() ([]FileInfo, error)
	FindFileByName(fileName string) (string, error)
}

type Storage struct {
	db *sql.DB
}

type FileInfo struct {
	FileName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	fmt.Println(storagePath)
	fmt.Println(os.Getwd())

	return &Storage{db: db}, nil
}

func (s *Storage) SaveImage(imageName string, size int, mimeType string, checksum string, createdAt time.Time) error {
	const op = "storage.sqlite.SaveImage"

	insertStmt, err := s.db.Prepare(`
	INSERT INTO files (filename, path_to_file, size_kb, mime_type, checksum)
	VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := insertStmt.Exec(
		imageName,
		os.Getenv("PATH_TO_SAVED_IMAGES"),
		size,
		mimeType,
		checksum,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("res: ", result, op)
	return nil
}

func (s *Storage) ListFiles() ([]FileInfo, error) {
	const op = "storage.sqlite.ListFiles"

	rows, err := s.db.Query("SELECT filename, created_at, updated_at FROM files")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var files []FileInfo

	for rows.Next() {
		var file FileInfo
		err := rows.Scan(&file.FileName, &file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return files, nil
}

func (s *Storage) FindFileByName(fileName string) (string, error) {
	const op = "storage.sqlite.FindFileByName"

	var foundFileName string
	err := s.db.QueryRow("SELECT filename FROM files WHERE filename = ?", fileName).Scan(&foundFileName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Файл не найден, возвращаем пустую строку и nil ошибку
			return "", nil
		}
		// Другая ошибка базы данных
		return "", fmt.Errorf("%s: failed to find file: %w", op, err)
	}

	// Файл найден, возвращаем его имя
	return foundFileName, nil
}

func (s *Storage) CheckTable(ctx context.Context) error {
	const op = "storage.sqlite.CheckTable"

	// Insert test data
	insertStmt, err := s.db.Prepare(`
        INSERT INTO files (filename, path_to_file, size_kb, mime_type, checksum)
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer insertStmt.Close()

	result, err := insertStmt.ExecContext(ctx,
		"test_image.jpg",
		"/storage/images/test_image.jpg",
		1024,
		"image/jpeg",
		"abc123checksum",
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("Inserted row with ID: %d\n", id)

	// Select and print all records
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, filename, path_to_file, size_bytes, mime_type, created_at, checksum 
        FROM files 
        WHERE deleted_at IS NULL
    `)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id       int64
			filename string
			path     string
			size     int64
			mimeType string
			created  string
			checksum string
		)
		if err := rows.Scan(&id, &filename, &path, &size, &mimeType, &created, &checksum); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		fmt.Printf("File: %+v\n", map[string]interface{}{
			"id":       id,
			"filename": filename,
			"path":     path,
			"size":     size,
			"mimeType": mimeType,
			"created":  created,
			"checksum": checksum,
		})
	}

	return nil
}
