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
	go generate

build:
	go build main.go

test:
	go test -v ./...