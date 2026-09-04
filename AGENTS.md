# OpenSearch integration tests

Use the persistent two-step workflow when testing against OpenSearch:

```sh
make start-opensearch-test-service
make run-opensearch-tests
```

`start-opensearch-test-service` waits for the local OpenSearch container to be
healthy. `run-opensearch-tests` uses that running service and does not stop it.
Leave the service running after the tests; the human operator will stop it when
appropriate.

For manual cleanup only, use:

```sh
make stop-opensearch-test-service
```

Do not use `make test-opensearch` for this workflow: it stops and removes the
test service, including its data volume, after the test run.
