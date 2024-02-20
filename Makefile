start_docker_without_api:
	docker-compose -f deployment/docker-compose.yml up -d grafana prometheus otel-collector zipkin-all-in-one

start_docker:
	docker-compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker-compose -f deployment/docker-compose.yml down
