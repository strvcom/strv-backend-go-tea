package util

import (
	"os"
	"testing"
)

const EnvVarTestCleanup = "GO_TEST_CLEANUP"

func CleanupAfterTest(t *testing.T) bool {
	t.Helper()
	return os.Getenv(EnvVarTestCleanup) != "false"
}
