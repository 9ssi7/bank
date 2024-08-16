
proto:
	protoc --go_out=. --go-grpc_out=.  api/rpc/protos/*.proto

jwt-key:
	ssh-keygen -t rsa -b 4096 -m PEM -f ./tmp/bank_jwtRS256.key

jwt-pub:
	openssl rsa -in ./tmp/bank_jwtRS256.key -pubout -outform PEM -out ./tmp/bank_jwtRS256.key.pub

jwt: jwt-key jwt-pub

jwt-register:
	docker secret create bank_private_key ./tmp/bank_jwtRS256.key
	docker secret create bank_public_key ./tmp/bank_jwtRS256.key.pub

compose:
	docker-compose -f ./deployments/docker-compose.yml up -d

compose-build:
	docker-compose -f ./deployments/docker-compose.yml up -d --build --remove-orphans

compose-down:
	docker-compose -f ./deployments/docker-compose.yml down

network:
	docker network create --driver overlay --attachable bank

build-srv:
	docker build -t github.com/9ssi7/bank:latest .

start-srv:
	docker service create --name 9ssi7bank --publish 4000:4000 --publish 5000:5000 --secret bank_private_key --secret bank_public_key --replicas 3 --mount type=bind,source=./deployments/config.yaml,target=/config.yaml --network bank github.com/9ssi7/bank:latest

stop-srv:
	docker service rm 9ssi7bank

once: jwt jwt-register network

born: once compose

dev: build-srv-dev run-srv-dev

clean:
	rm -rf deployments/bank_jwtRS256.key
	rm -rf deployments/bank_jwtRS256.key.pub

clean-docker:
	docker service rm 9ssi7bank
	docker secret rm bank_private_key
	docker secret rm bank_public_key
	docker network rm bank
	docker rmi github.com/9ssi7/bank:latest
	docker rmi github.com/9ssi7/bank:dev

.PHONY: proto jwt-key jwt-pub jwt jwt-register compose compose-build compose-down network build-srv start-srv stop-srv build-srv-dev run-srv-dev once born dev clean clean-docker