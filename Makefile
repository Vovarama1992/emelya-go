SWAGGER_OUT_DIR=./docs
SWAGGER_MAIN_FILE=./cmd/main.go

.PHONY: swagger tidy refresh

swagger:
	swag init --generalInfo ./cmd/main.go --output ./docs --parseDependency --parseInternal

tidy:
	go mod tidy

refresh:
	git pull origin master && \
	docker-compose build --no-cache && \
	docker-compose up -d && \
	docker-compose exec app migrate -path ./migrations -database $$DATABASE_URL up