Check the test coverage
```shell
go test  -covermode count -coverprofile coverage.out ./... && go tool cover -html=coverage.out
```

Check race conditions
```shell
go test -race ./...
```
