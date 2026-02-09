package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const e2eGateEnv = "ISSUE2MD_E2E"
const e2eAddrEnv = "ISSUE2MD_E2E_ADDR"

func TestIssue2mdWebE2EJourney(t *testing.T) {
	if strings.TrimSpace(os.Getenv(e2eGateEnv)) != "1" {
		t.Skip("set ISSUE2MD_E2E=1 to run E2E tests")
	}
	if !canBindLocalhost() {
		t.Skip("local tcp listen is not permitted in current environment")
	}

	repoRoot := repoRootDir(t)
	if _, err := os.Stat(filepath.Join(repoRoot, "docs", "swagger.json")); err != nil {
		t.Skip("docs/swagger.json not found, run make swagger first")
	}

	addr := strings.TrimSpace(os.Getenv(e2eAddrEnv))
	if addr == "" {
		addr = "127.0.0.1:18080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/issue2mdweb")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "ISSUE2MD_WEB_ADDR="+addr)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start issue2mdweb: %v", err)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
		close(waitCh)
	}()

	defer stopProcess(t, cmd, waitCh)

	baseURL := "http://" + addr
	client := &http.Client{Timeout: 3 * time.Second}

	if err := waitForWebReady(ctx, client, baseURL, waitCh); err != nil {
		t.Fatalf("wait for web ready: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}

	tcs := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantInBody string
	}{
		{
			name:       "index page",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
			wantInBody: "issue2md Web",
		},
		{
			name:       "swagger page",
			method:     http.MethodGet,
			path:       "/swagger",
			wantStatus: http.StatusOK,
			wantInBody: "/openapi.json",
		},
		{
			name:       "openapi spec",
			method:     http.MethodGet,
			path:       "/openapi.json",
			wantStatus: http.StatusOK,
			wantInBody: "\"swagger\": \"2.0\"",
		},
		{
			name:       "convert invalid url",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"invalid-url"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusBadRequest,
			wantInBody: "invalid github url",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			reqCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, tc.method, baseURL+tc.path, strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("do request: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status=%d want=%d body=%q", resp.StatusCode, tc.wantStatus, string(body))
			}
			if tc.wantInBody != "" && !strings.Contains(string(body), tc.wantInBody) {
				t.Fatalf("body=%q should contain %q", string(body), tc.wantInBody)
			}
		})
	}
}

func waitForWebReady(ctx context.Context, client *http.Client, baseURL string, waitCh <-chan error) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		reqCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, baseURL+"/", nil)
		if err != nil {
			cancel()
			return fmt.Errorf("create readiness request: %w", err)
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			cancel()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		} else {
			cancel()
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for ready: %w", ctx.Err())
		case err, ok := <-waitCh:
			if !ok {
				return fmt.Errorf("web process exited before ready")
			}
			return fmt.Errorf("web process exited before ready: %w", err)
		case <-ticker.C:
		}
	}
}

func stopProcess(t *testing.T, cmd *exec.Cmd, waitCh <-chan error) {
	t.Helper()

	if cmd.Process == nil {
		return
	}

	_ = cmd.Process.Signal(os.Interrupt)
	select {
	case _, ok := <-waitCh:
		if !ok {
			return
		}
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		select {
		case <-waitCh:
		case <-time.After(1 * time.Second):
		}
	}
}

func repoRootDir(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func canBindLocalhost() bool {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
