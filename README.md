# ulms-go
Template for services written on Go for the ULMS project

## REST API spec
There is [specification](api/v1.yml) described using [swagger format version 2.0](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md). One also can try it using [online Swagger editor](http://editor.swagger.io), or by opening `/swagger-ui/` URL of the deployed service.

## Installation
[Install Go](https://golang.org/doc/install) and run following commands:

    go get -u -v github.com/golang/dep/cmd/dep github.com/rubenv/sql-migrate/...
    go get -v github.com/netology-group/ulms-go/...

### Dependencies
From the project root folder (`$GOPATH/src/github.com/netology-group/ulms-go`) execute following command:

    dep ensure -v

### PostgreSQL
    docker run --detach --name postgres-11 --publish 5432:5432 postgres:11-alpine
    
### DB schema
    docker exec --tty --interactive --user postgres postgres-11 createdb ulms
    sql-migrate up
    
## Development
### Run
    go run cmd/api/api.go

### Making new migration
    sql-migrate new my-new-migration

### Tests
Make sure you have created DB for tests:

    docker exec --tty --interactive --user postgres postgres-11 createdb ulms-test
    
Run following command to execute tests:

    go test -v ./...
