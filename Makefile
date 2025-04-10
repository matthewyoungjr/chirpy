
run:
	go run main.go handlers.go models.go

build:
	go build main.go handlers.go models.go

air:
	air -c .air.toml