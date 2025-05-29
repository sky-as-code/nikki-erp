package env

import (
	"os"
)

type EnvName string

const (
	// Development environment
	APP_ENV_LOCAL EnvName = "local"
	// Development environment running inside a container
	APP_ENV_LOCAL_CONTAINER EnvName = "localcontainer"
	// Development environment
	APP_ENV_DEV EnvName = "dev"
	// System integration testing environment
	APP_ENV_SIT EnvName = "sit"
	// Staging (production simulation) environment
	APP_ENV_STG EnvName = "stg"
	// Performance testing environment
	APP_ENV_PERF EnvName = "perf"
	// Production environment
	APP_ENV_PROD EnvName = "prod"
	// While running unit tests
	APP_ENV_TEST EnvName = "test"
)

func AppEnv() EnvName {
	return EnvName(os.Getenv("APP_ENV"))
}

func IsLocal() bool {
	curEnv := AppEnv()
	return curEnv == "" || curEnv == APP_ENV_LOCAL || curEnv == APP_ENV_LOCAL_CONTAINER || curEnv == APP_ENV_TEST
}

func IsNonProd() bool {
	curEnv := AppEnv()
	return IsLocal() || curEnv == APP_ENV_DEV || curEnv == APP_ENV_LOCAL_CONTAINER ||
		curEnv == APP_ENV_SIT || curEnv == APP_ENV_PERF
	// "Staging" should be considered equivalent to "production"
}

func IsProd() bool {
	curEnv := AppEnv()
	return curEnv == APP_ENV_STG || curEnv == APP_ENV_PROD
	// "Staging" should be considered equivalent to "production"
}

func IsRunningUnitTest() bool {
	curEnv := AppEnv()
	return curEnv == APP_ENV_TEST
}

func RunOnLocal(task func()) {
	if IsLocal() {
		task()
	}
}

func RunOnUnitTest(task func()) {
	if IsRunningUnitTest() {
		task()
	}
}

func RunOnNonProd(task func()) {
	if IsNonProd() {
		task()
	}
}

func RunOnProd(task func()) {
	if IsProd() {
		task()
	}
}

func Cwd() string {
	workDir := os.Getenv("WORKING_DIR")
	if len(workDir) == 0 {
		workDir, _ = os.Getwd()
	}
	return workDir
}
