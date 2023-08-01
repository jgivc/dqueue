.DEFAULT_GOAL := default

app_name = vapp

e2e:
	docker compose up --build --abort-on-container-exit --exit-code-from sipp_cl

default:
	go build -o $(app_name) cmd/*.go

