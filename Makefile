
export GO111MODULE=on

.PHONY: test
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/kubectl-storageos-bundle github.com/croomes/kubectl-plugin/cmd/bundle
	go build -o bin/kubectl-storageos-preflight github.com/croomes/kubectl-plugin/cmd/preflight

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...
