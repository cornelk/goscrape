
go.get:
	go get -u ./...

go.tidy:
	go mod tidy -compat=1.18

go.test:
	go test ./...

go.install:
	go install ./...

