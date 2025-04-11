module github.com/sky-as-code/nikki-erp/main

go 1.21

replace github.com/sky-as-code/nikki-erp/modules/core => ../modules/core

replace github.com/sky-as-code/nikki-erp/modules/shared => ../modules/shared

replace github.com/sky-as-code/nikki-erp/utility => ../utility

require (
	github.com/sky-as-code/nikki-erp/modules/core v0.0.0-00010101000000-000000000000
	github.com/sky-as-code/nikki-erp/modules/shared v0.0.0-00010101000000-000000000000
	github.com/sky-as-code/nikki-erp/utility v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/labstack/echo/v4 v4.13.3 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/thoas/go-funk v0.9.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xgfone/go-cast v0.9.0 // indirect
	github.com/xgfone/go-defaults v0.14.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	gorm.io/gorm v1.25.12 // indirect
)
