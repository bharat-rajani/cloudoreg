postgres:
	@echo "+$@"

	@docker rm -f postgres-sources || true

	@docker run \
		--detach \
		--net=host \
		--name postgres-sources \
		--env POSTGRES_PASSWORD=dev \
		--env PGDATA=/var/lib/postgresql/data/pgdata \
		--volume $(PWD)/_deployment/data/postgres:/var/lib/postgresql/data \
		postgres

	@sleep 2

	@docker exec -it postgres-sources psql \
		--host=localhost --user=postgres --password \
		-c 'create database sources_api_go_development;' || true

redis:
	@echo "+$@"

	@docker rm -f redis-sources || true
	@docker run --detach --name redis-sources --net=host redis

kafka:
	@echo "+$@"

	@docker rm -f kafka-sources || true
	@docker run --detach --rm -it --net=host \
		--name=kafka-sources \
		--env RUNTESTS=0 \
		--pull=always \
		lensesio/fast-data-dev

sources:
	@echo "+$@"

	# Source the deployment config.
	@. $(PWD)/_deployment/env.sh

	# Setup the clone destination so the operation is idempotent.
	@rm -rf $(PWD)/cloned/sources-api-go
	@mkdir -p $(PWD)/cloned/sources-api-go

	# Clone the repo.
	@git clone git@github.com:RedHatInsights/sources-api-go.git \
		$(PWD)/cloned/sources-api-go

	# Execute the required make commands to run sources.
	@make -C $(PWD)/cloned/sources-api-go setup
	@make -C $(PWD)/cloned/sources-api-go inlinerun

run: postgres redis kafka sources

bulk-create:
	@echo "+$@"

	# HTTP request for sources-api
	@curl -XPOST http://localhost:4000/api/sources/v3.1/bulk_create \
		-d _deployment/source_body.json -H "x-rh-sources-account-number: 1234466"
