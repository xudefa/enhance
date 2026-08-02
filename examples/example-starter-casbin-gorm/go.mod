module github.com/xudefa/enhance/examples/example-starter-casbin-gorm

go 1.25.12

require (
	github.com/casbin/casbin/v2 v2.123.0
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/casbin-gorm v0.0.5
	github.com/xudefa/enhance/starter/gorm v0.0.5
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/casbin/gorm-adapter/v3 v3.38.0 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/glebarez/go-sqlite v1.20.3 // indirect
	github.com/glebarez/sqlite v1.7.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/microsoft/go-mssqldb v1.8.2 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
	gorm.io/driver/sqlserver v1.5.3 // indirect
	gorm.io/gorm v1.31.2 // indirect
	gorm.io/plugin/dbresolver v1.6.0 // indirect
	modernc.org/libc v1.55.3 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/sqlite v1.20.4 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/casbin-gorm => ../../starter/casbin-gorm
	github.com/xudefa/enhance/starter/gorm => ../../starter/gorm
)
