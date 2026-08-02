module github.com/xudefa/enhance/examples/example-starter-mongodb

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.4
	github.com/xudefa/enhance/starter/mongodb v0.0.4
	go.mongodb.org/mongo-driver v1.15.0
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/montanaflynn/stats v0.7.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/mongodb => ../../starter/mongodb
)
