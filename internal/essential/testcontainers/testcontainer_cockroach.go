package testcontainers

import (
	"context"
	"fmt"
	"time"

	"github.com/guardlight/server/internal/essential/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type CockroachSQLContainer struct {
	testcontainers.Container
	MappedPort   string
	Host         string
	DatabaseName string
}

func (c CockroachSQLContainer) GetDSN() string {
	return fmt.Sprintf("postgresql://root@%s:%s/%s?sslmode=disable", c.Host, c.MappedPort, c.DatabaseName)
}

func NewCockroachSQLContainer(ctx context.Context) (*CockroachSQLContainer, error) {
	zap.S().Info("testcontainer: Starting cockroachDB container")

	req := testcontainers.ContainerRequest{
		ExposedPorts: []string{"26257/tcp", "8080/tcp"},
		Image:        "cockroachdb/cockroach:v22.2.2",
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
		Cmd:          []string{"start-single-node", "--insecure"},
		Env: map[string]string{
			"COCKROACH_DATABASE": config.Get().Database.Name,
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "26257")
	if err != nil {
		return nil, err
	}

	time.Sleep(10 * time.Second) // Wait for CockroachDB to create DB

	return &CockroachSQLContainer{
		Container:    container,
		MappedPort:   mappedPort.Port(),
		Host:         host,
		DatabaseName: config.Get().Database.Name,
	}, nil
}
