build: 
	@go install .

run: build 
	@./bin/gobank

test: 
	@go test -v ./...