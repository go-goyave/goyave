module goyave.dev/goyave/v5

go 1.26.0

require (
	github.com/Code-Hex/uniseg v0.2.0
	github.com/andybalholm/brotli v1.2.3
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/klauspost/compress v1.19.2
	github.com/samber/lo v1.53.0
	github.com/stretchr/testify v1.12.1
	golang.org/x/crypto v0.55.0
	gorm.io/driver/bigquery v1.2.1
	gorm.io/driver/clickhouse v0.7.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.2
	gorm.io/driver/sqlite v1.6.0
	gorm.io/driver/sqlserver v1.6.4
	gorm.io/gorm v1.31.2
	goyave.dev/copier v0.4.4
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.23.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/bigquery v1.82.0 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.13.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/ClickHouse/ch-go v0.74.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.48.0 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.21 // indirect
	github.com/googleapis/gax-go/v2 v2.24.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.50 // indirect
	github.com/microsoft/go-mssqldb v1.11.0 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.29 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20260824195058-e88cd73687aa // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/telemetry v0.0.0-20260828145429-86cb5733f5b7 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.296.0 // indirect
	google.golang.org/genproto v0.0.0-20260831171406-18b4a7587f8a // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260831171406-18b4a7587f8a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260831171406-18b4a7587f8a // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

retract (
	v5.5.4 // Introduces another bug in the validation package, 5.5.5 fixes this bug
	v5.5.3 // Introduces bug in the validation package, 5.5.4 fixes this bug
)
