.PHONY: docs

install-osx:
	go build -o ~/.terraform.d/plugins/hashicorp.com/trustgrid/tg/0.1/darwin_amd64/terraform-provider-tg main.go

install-linux:
	go build -o ~/.terraform.d/plugins/hashicorp.com/trustgrid/tg/0.1/linux_amd64/terraform-provider-tg main.go

run:
	cd poc && rm -rf .terraform* && terraform init && terraform apply -auto-approve

reset:
	cd poc && rm -rf .terraform* *.tfstate && terraform init && terraform apply -auto-approve

docs:
	go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	go generate
	go mod tidy

build:
	go build main.go

test:
	go test -v ./...

sweep:
	go test -v ./acctests -v -sweep=all

testacc:
	TF_LOG=ERROR TF_ACC=1 go test -v ./...

lint:
	golangci-lint run --tests=false ./...