package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestNodeAPI(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node.js is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	addon := filepath.Join(t.TempDir(), "integration.node")
	cmd := exec.CommandContext(ctx, "go", "build", "-buildmode=c-shared", "-o", addon, "./addon")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build addon: %v\n%s", err, out)
	}
	nodeCtx, nodeCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer nodeCancel()
	cmd = exec.CommandContext(nodeCtx, "node", "--expose-gc", "--test", "integration.cjs")
	cmd.Env = append(os.Environ(), "NAPI_GO_ADDON="+addon)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Node-API integration: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}
