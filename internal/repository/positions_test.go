package repository

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

var (
	conn DbInterface
)

func TestMain(m *testing.M) {
	ctx, _ := context.WithCancel(context.Background())

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Errorf("Could not construct pool: %s", err)
		return
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Hostname:   "postgres_test",
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=user",
			"POSTGRES_DB=chart",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	postgresHostAndPort := pgResource.GetHostPort("5432/tcp")
	postgresUrl := fmt.Sprintf("postgres://user:password@%s/chart?sslmode=disable", postgresHostAndPort)

	log.Info("Connecting to database on url: ", postgresUrl)

	var dbpool *pgxpool.Pool
	if err = pool.Retry(func() error {
		dbpool, err = pgxpool.New(ctx, postgresUrl)
		if err != nil {
			dbpool.Close()
			log.Error("can't connect to the pgxpool: %w", err)
		}
		return dbpool.Ping(ctx)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	///// ///// ///// ///// ///// ///// ///// ///// ///// ///// /////

	commandArr := []string{
		"-url=jdbc:postgresql://" + postgresHostAndPort + "/chart",
		"-user=user",
		"-password=password",
		"-locations=filesystem:../../migrations/",
		"-schemas=trading",
		"-connectRetries=60",
		"migrate",
	}
	cmd := exec.Command("flyway", commandArr[:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Printf("error: %s", err))
	}

	pool.MaxWait = 120 * time.Second
	conn = NewPostgresRepository(dbpool)

	code := m.Run()

	if err := pool.Purge(pgResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
