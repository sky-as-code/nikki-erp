module github.com/sky-as-code/nikki-erp

go 1.25.0

require (
	entgo.io/ent v0.14.4
	github.com/ThreeDotsLabs/watermill v1.4.7
	github.com/ThreeDotsLabs/watermill-redisstream v1.4.3
	github.com/XSAM/otelsql v0.44.0
	github.com/aws/aws-sdk-go-v2 v1.43.8
	github.com/aws/aws-sdk-go-v2/config v1.32.35
	github.com/aws/aws-sdk-go-v2/credentials v1.19.34
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.22.41
	github.com/aws/aws-sdk-go-v2/service/s3 v1.107.4
	github.com/aws/smithy-go v1.28.0
	github.com/aws/smithy-go/tracing/smithyoteltracing v1.0.28
	github.com/go-playground/validator/v10 v10.27.0
	github.com/go-sanitize/sanitize v1.1.0
	github.com/goccy/go-yaml v1.18.0
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang/mock v1.7.0-rc.1
	github.com/google/uuid v1.6.0
	github.com/huandu/go-sqlbuilder v1.39.1
	github.com/invopop/validation v0.8.0
	github.com/joho/godotenv v1.5.1
	github.com/json-iterator/go v1.1.12
	github.com/labstack/echo/v5 v5.1.0
	github.com/lib/pq v1.10.9
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.5.0
	github.com/redis/go-redis/extra/redisotel/v9 v9.22.0
	github.com/redis/go-redis/v9 v9.22.0
	github.com/samber/lo v1.53.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/shopspring/decimal v1.4.0
	github.com/stretchr/testify v1.12.1
	github.com/thoas/go-funk v0.9.3
	github.com/tidwall/gjson v1.18.0
	github.com/xgfone/go-cast v0.9.0
	github.com/xuri/excelize/v2 v2.10.1
	go.bryk.io/pkg v0.0.0-20250623171018-8f3f4709eea0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.46.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.46.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.46.0
	go.opentelemetry.io/otel/sdk v1.46.0
	go.opentelemetry.io/otel/trace v1.46.0
	go.uber.org/dig v1.19.0
	golang.org/x/crypto v0.55.0
	golang.org/x/mod v0.38.0
	golang.org/x/text v0.41.0
	gopkg.in/jeevatkm/go-model.v1 v1.1.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.19 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.40 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.32 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.39 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.40 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.33.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.38.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.45.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.22.0 // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.46.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/grpc v1.83.1 // indirect
)

require (
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/gabriel-vasile/mimetype v1.4.8
	github.com/go-co-op/gocron/v2 v2.21.2
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/huandu/go-clone v1.7.3 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xgfone/go-defaults v0.14.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
