build:
	go build -o server.exe ./cmd/server/main.go
run:
	go run ./cmd/server/main.go

migrateUp:
	go run ./cmd/migrator/ --storage-path=./internal/storage/sqlite/image.db --migrations-path=./migrations

