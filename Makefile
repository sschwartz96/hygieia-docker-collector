build_docker:
	docker build -t hygieia-docker-collector .

run_docker:
	docker run --network hygieia \
		-v ${PWD}/config.docker.json:/hygieia-docker-collector/config.json \
		--name hygieia-docker-collector \
		hygieia-docker-collector
