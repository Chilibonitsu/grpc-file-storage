module imagestorage

go 1.24.0

require (
	github.com/caarlos0/env/v11 v11.3.1
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/grpc v1.70.0
	imagestorage/contracts v0.0.0-00010101000000-000000000000
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace imagestorage/contracts => ./contracts
