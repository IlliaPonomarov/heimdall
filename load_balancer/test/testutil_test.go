package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func nginxConf(body string) string {
	return fmt.Sprintf(`events {}
http {
    server {
        listen 80;
        location / {
            return 200 '%s';
            add_header Content-Type text/plain;
        }
        location /health {
            return 200 'ok';
            add_header Content-Type text/plain;
        }
    }
}`, body)
}

type testBackend struct {
	container testcontainers.Container
	url       string
}

func startBackends(t testing.TB, ctx context.Context, n int) ([]testBackend, func()) {
	t.Helper()
	backends := make([]testBackend, 0, n)

	cleanup := func() {
		for _, b := range backends {
			if err := testcontainers.TerminateContainer(b.container); err != nil {
				t.Logf("failed to terminate container: %v", err)
			}
		}
	}

	for i := 0; i < n; i++ {
		body := fmt.Sprintf("backend-%d", i)
		conf := nginxConf(body)

		container, err := testcontainers.Run(ctx, "nginx:alpine",
			testcontainers.WithFiles(testcontainers.ContainerFile{
				Reader:            strings.NewReader(conf),
				ContainerFilePath: "/etc/nginx/nginx.conf",
			}),
			testcontainers.WithExposedPorts("80/tcp"),
			testcontainers.WithWaitStrategy(wait.ForHTTP("/").WithPort("80/tcp")),
		)
		if err != nil {
			cleanup()
			t.Fatalf("failed to start backend-%d: %v", i, err)
		}

		endpoint, err := container.Endpoint(ctx, "http")
		if err != nil {
			cleanup()
			t.Fatalf("failed to get endpoint for backend-%d: %v", i, err)
		}

		backends = append(backends, testBackend{container: container, url: endpoint})
	}

	return backends, cleanup
}
