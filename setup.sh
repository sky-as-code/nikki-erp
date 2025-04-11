#!/bin/bash

# Create module directories
mkdir -p modules/{core,api,db}

# Initialize core module
cd modules/core
go mod init github.com/sky-as-code/nikki-erp/modules/core
touch main.go

# Initialize api module
cd ../api
go mod init github.com/sky-as-code/nikki-erp/modules/api
touch main.go

# Initialize db module
cd ../db
go mod init github.com/sky-as-code/nikki-erp/modules/db
touch main.go

# Return to root and initialize workspace
cd ../..
go work init
go work use ./modules/*