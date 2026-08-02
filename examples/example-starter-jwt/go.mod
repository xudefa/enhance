module github.com/xudefa/enhance/examples/example-starter-jwt

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/jwt v0.0.5
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/jwt => ../../starter/jwt
)
