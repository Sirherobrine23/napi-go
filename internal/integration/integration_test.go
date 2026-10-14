package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNodeAPI(t *testing.T) {
	node := os.Getenv("NAPI_GO_NODE")
	if node == "" {
		node = "node"
	}
	if _, err := exec.LookPath(node); err != nil {
		t.Fatalf("Node.js is required: %v", err)
	}
	goCommand := os.Getenv("NAPI_GO_GO")
	if goCommand == "" {
		goCommand = filepath.Join(runtime.GOROOT(), "bin", "go")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	addon := filepath.Join(t.TempDir(), "integration.node")
	args := []string{"build", "-buildvcs=false", "-buildmode=c-shared", "-o", addon}
	if os.Getenv("NAPI_GO_TEST_RACE") == "1" {
		args = append(args, "-race")
	}
	args = append(args, "./addon")
	cmd := exec.CommandContext(ctx, goCommand, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build addon: %v\n%s", err, out)
	}
	nodeCtx, nodeCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer nodeCancel()
	cmd = exec.CommandContext(nodeCtx, node, "--expose-gc", "--test", "integration.cjs", "eventemitter.cjs", "eventemitter_thread.cjs", "objects.cjs", "builtins.cjs", "stream_io.cjs")
	cmd.Env = append(os.Environ(), "NAPI_GO_ADDON="+addon)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Node-API integration: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}
