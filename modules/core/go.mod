module github.com/sky-as-code/nikki-erp/modules/core

go 1.21

replace github.com/sky-as-code/nikki-erp/utility => ../../utility

require (
	github.com/golang/mock v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/sky-as-code/nikki-erp/utility v0.0.0-00010101000000-000000000000
	github.com/tidwall/gjson v1.18.0
	github.com/xgfone/go-cast v0.9.0
)

require (
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/xgfone/go-defaults v0.14.0 // indirect
)
