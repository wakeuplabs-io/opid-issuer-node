//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/joho/godotenv/cmd/godotenv"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
