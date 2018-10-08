PACKAGE := ./cmd/... ./pkg/...

.DEFAULT_GOAL := format-and-test

.PHONY: test
test: format
	go test -v -race $(PACKAGE)

.PHONY: format
format:
	go fmt $(PACKAGE)
	go vet $(PACKAGE)

.PHONY: lint
lint: format
	golint $(PACKAGE)

.PHONY: apiserver
apiserver: format
	go build -o apiserver github.com/cfchou/icecream/cmd/apiserver

.PHONY: db
db:
	./scripts/recreate_db.sh

.PHONY: run
run: apiserver
	./apiserver -c icecream.yaml


