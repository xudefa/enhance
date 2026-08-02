module github.com/xudefa/enhance/examples/example-starter-validator

go 1.25.12

require (
	github.com/go-playground/validator/v10 v10.16.0
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/validator v0.0.5
)

require (
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/validator => ../../starter/validator
)
