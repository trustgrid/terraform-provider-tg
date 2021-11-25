
install:
	go build -o ~/.terraform.d/plugins/hashicorp.com/trustgrid/tg/0.1/linux_amd64/terraform-provider-tg main.go

run:
	cd examples && rm -rf .terraform* && terraform init && terraform apply -auto-approve

reset:
	cd examples && rm -rf .terraform* *.tfstate && terraform init && terraform apply -auto-approve