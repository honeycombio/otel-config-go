#!/usr/bin/env bats

load test_helpers/utilities

CONTAINER_NAME="app-sdk-http"
COLLECTOR_NAME="collector"
TRACER_NAME="my-app"

setup_file() {
	echo "# 🚧" >&3
	docker-compose up --build --detach collector
	wait_for_ready_collector ${COLLECTOR_NAME}
	docker-compose up --build --detach ${CONTAINER_NAME}
	wait_for_traces
}

teardown_file() {
	cp collector/data.json collector/data-results/data-${CONTAINER_NAME}.json
	docker-compose stop ${CONTAINER_NAME}
	docker-compose restart collector
	wait_for_flush
}

# TESTS

@test "Manual instrumentation produces span with name of span" {
	result=$(span_names_for ${TRACER_NAME})
	assert_equal "$result" '"doing-things"'
}
