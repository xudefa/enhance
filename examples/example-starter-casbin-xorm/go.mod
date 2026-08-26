module github.com/xudefa/enhance/examples/example-starter-casbin-xorm

go 1.25.12

require (
	github.com/casbin/casbin/v2 v2.123.0
	github.com/xudefa/enhance v0.0.6
	github.com/xudefa/enhance/starter/casbin-xorm v0.0.6
	github.com/xudefa/enhance/starter/xorm v0.0.6
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
	github.com/casbin/xorm-adapter/v2 v2.4.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	modernc.org/cc/v3 v3.41.0 // indirect
	xorm.io/builder v0.3.13 // indirect
	xorm.io/xorm v1.3.11 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/casbin-xorm => ../../starter/casbin-xorm
	github.com/xudefa/enhance/starter/xorm => ../../starter/xorm
)
