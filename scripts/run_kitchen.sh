set -e
go build dronutz/cmd/kitchen/kitchen.go
./kitchen "$@"