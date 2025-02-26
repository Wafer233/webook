.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi  wafer233/webook:v0.0.1 . || true
	@docker build -t wafer233/webook:v0.0.1 .



