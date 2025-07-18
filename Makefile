SWAGGER_OUT_DIR=./docs
SWAGGER_MAIN_FILE=./cmd/main.go

.PHONY: swagger tidy

swagger:
	swag init --generalInfo ./cmd/main.go --output ./docs --parseDependency --parseInternal

tidy:
	go mod tidy