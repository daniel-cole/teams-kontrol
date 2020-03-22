test:
	go test ./...

dev:
	make test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/teams-kontrol
	docker build . -t teams-kontrol:dev
	kind load docker-image teams-kontrol:dev

