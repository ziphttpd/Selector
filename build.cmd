rem go generate -v
go run github.com/rakyll/statik -f -src=static
go build -o selector.exe main.go
