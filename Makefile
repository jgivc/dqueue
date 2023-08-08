.DEFAULT_GOAL := default

app_name = dqueue

e2e:
	docker compose -f integration/docker-compose.yml up --build --abort-on-container-exit --exit-code-from sipp_cl

default:
	go build -o $(app_name) cmd/*.go

