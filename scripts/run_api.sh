set -e
go build dronutz/cmd/api/api.go
./api "$@"