#!/usr/bin/env bats

load test_helpers/utilities

CONTAINER_NAME="app-sdk-grpc"
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

@test "Resource attributes can be set via environment variable" {
	env_result=$(spans_received | jq ".resource.attributes[] | select(.key == \"resource.example_set_in_env\") | .value.stringValue")
	assert_equal "$env_result" '"ENV"'
}

@test "Resource attributes can be set in code" {
	code_result=$(spans_received | jq ".resource.attributes[] | select(.key == \"resource.example_set_in_code\") | .value.stringValue")
	assert_equal "$code_result" '"CODE"'
}

@test "Resource attributes set in code win over matching key set in environment" {
	clobber_result=$(spans_received | jq ".resource.attributes[] | select(.key == \"resource.example_clobber\") | .value.stringValue")
	assert_equal "$clobber_result" '"CODE_WON"'
}
