package testcontainers

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultDbUser     = "database-service"
	defaultDbPassword = "password"
	defaultDbName     = "password"
)

type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

func (c PostgresContainer) Kill() {
	_ = c.Container.Terminate(context.Background())
}

var embedMigrations embed.FS

func SetupDb() PostgresContainer {
	env := map[string]string{
		"POSTGRES_HOST_AUTH_METHOD": "trust",
		"POSTGRES_PASSWORD":         defaultDbPassword,
		"POSTGRES_USER":             defaultDbUser,
		"POSTGRES_DB":               defaultDbName,
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16.2",
		ExposedPorts: []string{"5432/tcp"},
		Env:          env,
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
	}

	dsn := getDsn(host, mappedPort.Port())

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "../../migrations"); err != nil {
		panic(err)
	}

	return PostgresContainer{Container: container, DSN: dsn}
}

func getDsn(host, port string) string {

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC", host, port, defaultDbUser, defaultDbPassword, defaultDbName)

}

type RedisContainer struct {
	Container testcontainers.Container
	conn      string
}

func (c RedisContainer) Conn() string {
	return c.conn
}

func (c RedisContainer) Kill() {
	_ = c.Container.Terminate(context.Background())
}

func SetupRedis() RedisContainer {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7.2.4",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	var redisC RedisContainer

	redisC.Container = container

	endpoint, err := container.Endpoint(context.Background(), "")
	if err != nil {
		panic(err)
	}

	redisC.conn = endpoint

	if err != nil {
		panic(err)
	}

	return redisC
}
