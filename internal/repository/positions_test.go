package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Note: using decimal.New(_, 1) results in assert error
// beter use asser for each field + decimal.Compare

var (
	conn DbInterface

	input1 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb1",
		OpenPrice:   decimal.New(19, 0),
		ClosePrice:  decimal.New(130, 0),
		Buy:         true,
	}

	input2 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb2",
		OpenPrice:   decimal.New(123, 0),
		ClosePrice:  decimal.New(11, 0),
		Buy:         false,
	}

	input3 = model.Position{
		OperationID: uuid.New(),
		UserID:      input1.UserID,
		Symbol:      "symb1",
		OpenPrice:   decimal.New(190, 0),
		ClosePrice:  decimal.New(183, 0),
		Buy:         true,
	}
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

	msg, err := cmd.Output()
	if err != nil {
		log.Errorf("Migtation error: %s", err) // Note: Error output is poor
	}

	str := string(msg)
	log.Info(
		strings.Replace(
			strings.Replace(str, "\n\n", " ", -1), "\n", " ", -1,
		),
	)

	pool.MaxWait = 120 * time.Second
	conn = NewPostgresRepository(dbpool)

	mapStrChan := make(map[string]chan model.Price)
	mapStrMapStrChan := make(map[string]map[string]chan model.Price)
	connMap = NewStringPrice(mapStrChan)
	connMapMap = NewSymbOperMap(mapStrMapStrChan)

	code := m.Run()

	if err := pool.Purge(pgResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestPostgresAdd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		input    model.Position
		hasError bool
	}{
		{
			name:     "standart input-1",
			input:    input1,
			hasError: false,
		},
		{
			name:     "standart input-2",
			input:    input2,
			hasError: false,
		},
		{
			name:     "standart input-3",
			input:    input3,
			hasError: false,
		},
	}

	for _, test := range testTable {
		err := conn.Add(ctx, test.input)

		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			assert.Nil(t, err, test.name)
		}
	}

	log.Info("TestPostgresAdd finished!")
}

func TestPostgresGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		input    uuid.UUID
		expected []model.Position
		hasError bool
	}{
		{
			name:  "standart input-1&3",
			input: input1.UserID,
			expected: []model.Position{
				input1,
				input3,
			},
			hasError: false,
		},
		{
			name:     "standart input-2",
			input:    input2.UserID,
			expected: []model.Position{input2},
			hasError: false,
		},
	}

	for _, test := range testTable {
		actual, err := conn.Get(ctx, test.input)

		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			if ok := assert.Nil(t, err, test.name); !ok {
				continue
			}

			assert.Equal(t, test.expected, actual, test.name)
		}
	}

	log.Info("TestPostgresGet finished!")
}

func TestPostgresDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		input    uuid.UUID
		hasError bool
	}{
		{
			name:     "standart input-1",
			input:    input1.OperationID,
			hasError: false,
		},
		{
			name:     "standart input-2",
			input:    input2.OperationID,
			hasError: false,
		},
		{
			name:     "standart input-3",
			input:    input2.OperationID,
			hasError: false,
		},
	}

	for _, test := range testTable {
		err := conn.Deleete(ctx, test.input)

		if test.hasError {
			assert.Error(t, err, test.name)
		} else {
			assert.Nil(t, err, test.name)
		}
	}

	log.Info("TestPostgresDelete finished!")
}
