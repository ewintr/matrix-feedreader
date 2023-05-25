docker-push:
	docker build . -t matrix-bots
	docker tag matrix-bots registry.ewintr.nl/matrix-feedreader
	docker push registry.ewintr.nl/matrix-feedreader
