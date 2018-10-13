PACKAGE := ./cmd/... ./pkg/...

.DEFAULT_GOAL := test

.PHONY: test
test:
	go test -v -race $(PACKAGE)

.PHONY: format
format: test
	go fmt $(PACKAGE)
	go vet $(PACKAGE)

.PHONY: lint
lint: format
	golint $(PACKAGE)

.PHONY: apiserver
apiserver: test
	go build -o apiserver github.com/cfchou/icecream/cmd/apiserver

.PHONY: db
db:
	./scripts/recreate_db.sh

.PHONY: run
run: apiserver
	./apiserver -c icecream.yaml


