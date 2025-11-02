

build:
	@mkdir bin || true
	-rm bin/sentiment
	CGO_ENABLED=0 GOOS=linux go build -o bin/sentiment .

up:
	-docker-compose -f ./test/docker-compose.yml down
	docker-compose -f ./test/docker-compose.yml up --build
