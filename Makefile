clean:
	rm -rf ./smoke-tests/collector/data.json
	rm -rf ./smoke-tests/collector/data-results/*.json
	rm -rf ./smoke-tests/report.*

test:
	go test -v ./...

smoke-tests/collector/data.json:
	@echo ""
	@echo "+++ Zhuzhing smoke test's Collector data.json"
	@touch $@ && chmod o+w $@

smoke-sdk-grpc: smoke-tests/collector/data.json
	@echo ""
	@echo "+++ Running gRPC smoke tests."
	@echo ""
	cd smoke-tests && bats ./smoke-sdk-grpc.bats --report-formatter junit --output ./

smoke-sdk-http: smoke-tests/collector/data.json
	@echo ""
	@echo "+++ Running HTTP smoke tests."
	@echo ""
	cd smoke-tests && bats ./smoke-sdk-http.bats --report-formatter junit --output ./

smoke-sdk: smoke-sdk-grpc smoke-sdk-http

smoke: docker_compose_present
	@echo ""
	@echo "+++ Smoking all the tests."
	@echo ""
	cd smoke-tests && bats . --report-formatter junit --output ./

unsmoke: docker_compose_present
	@echo ""
	@echo "+++ Spinning down the smokers."
	@echo ""
	cd smoke-tests && docker-compose down --volumes

#: use this for local smoke testing
resmoke: unsmoke smoke

.PHONY: clean-smoke-tests example smoke unsmoke resmoke smoke-sdk-grpc smoke-sdk-http smoke-sdk

.PHONY: docker_compose_present
docker_compose_present:
	@which docker-compose || (echo "Required docker-compose command is missing"; exit 1)