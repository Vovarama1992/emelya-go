include .env
export

SWAGGER_OUT_DIR=./docs
SWAGGER_MAIN_FILE=./cmd/main.go

.PHONY: swagger tidy refresh

swagger:
	swag init --generalInfo ./cmd/main.go --output ./docs --parseDependency --parseInternal

tidy:
	go mod tidy

refresh:
	git pull origin master
	docker-compose build app
	docker-compose up -d app db redis pgadmin
	until docker-compose exec db pg_isready -U emelya; do sleep 1; done
	sleep 3
	until docker-compose exec app migrate -path ./migrations -database $$DATABASE_URL up; do sleep 1; done
print:
	echo $$DATABASE_URL