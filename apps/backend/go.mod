module github.com/borgehl/bruttice/apps/backend

go 1.25.3

replace github.com/borgehl/bruttice/packages/go-pkg => ../../packages/go-pkg

require (
	github.com/borgehl/bruttice/packages/go-pkg v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/gorilla/feeds v1.2.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/IPGeolocation/ip-geolocation-go-sdk v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
)
