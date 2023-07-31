.DEFAULT_GOAL := default

app_name = vapp

default:
	go build -o $(app_name) cmd/*.go

