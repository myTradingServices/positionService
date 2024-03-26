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
// beter use assert for each field + decimal.Compare

// TODO: remove hasError filed from testTables where it's unnesasary
var (
	conn DBInterface

	input1 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb1",
		OpenPrice:   decimal.New(19, 0),
		Buy:         true,
		Open:        true,
	}

	input2 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb2",
		OpenPrice:   decimal.New(123, 0),
		Buy:         false,
		Open:        true,
	}

	input3 = model.Position{
		OperationID: uuid.New(),
		UserID:      input1.UserID,
		Symbol:      "symb1",
		OpenPrice:   decimal.New(190, 0),
		Buy:         true,
		Open:        true,
	}

	closePrice1 = decimal.New(130, 0)
	closePrice2 = decimal.New(11, 0)
	closePrice3 = decimal.New(183, 0)
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

//TODO change to GetLaterThen

// func TestGetAllOpened(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
// 	defer cancel()

// 	opened, err := conn.GetAllOpend(ctx)
// 	assert.Nil(t, err, "GetAllOpened error is not nil")
// 	assert.ElementsMatch(t, opened, []model.Position{
// 		{
// 			OperationID: input1.OperationID,
// 			Symbol:      input1.Symbol,
// 			Buy:         input1.Buy,
// 			OpenPrice:   input1.OpenPrice,
// 		},
// 		{
// 			OperationID: input2.OperationID,
// 			Symbol:      input2.Symbol,
// 			Buy:         input2.Buy,
// 			OpenPrice:   input2.OpenPrice,
// 		},
// 		{
// 			OperationID: input3.OperationID,
// 			Symbol:      input3.Symbol,
// 			Buy:         input3.Buy,
// 			OpenPrice:   input3.OpenPrice,
// 		},
// 	})

// 	log.Info("TestGetAllOpened finished!")
// }

func TestUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name  string
		input model.Position
	}{
		{
			name: "std input-1",
			input: model.Position{
				OperationID: input1.OperationID,
				ClosePrice:  closePrice1,
				Open:        false,
			},
		},
		{
			name: "std input-2",
			input: model.Position{
				OperationID: input2.OperationID,
				ClosePrice:  closePrice2,
				Open:        false,
			},
		},
		{
			name: "std input-3",
			input: model.Position{
				OperationID: input3.OperationID,
				ClosePrice:  closePrice3,
				Open:        true,
			},
		},
	}

	for _, test := range testTable {
		err := conn.Update(ctx, test.input)

		assert.Nil(t, err, test.name)
	}

	log.Info("TestUpdate finished!")
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
				{
					OperationID: input1.OperationID,
					UserID:      input1.UserID,
					Symbol:      input1.Symbol,
					OpenPrice:   input1.OpenPrice,
					ClosePrice:  closePrice1,
					Buy:         input1.Buy,
					Open:        false,
				},
				{
					OperationID: input3.OperationID,
					UserID:      input3.UserID,
					Symbol:      input3.Symbol,
					OpenPrice:   input3.OpenPrice,
					ClosePrice:  closePrice3,
					Buy:         input3.Buy,
					Open:        true,
				},
			},
			hasError: false,
		},
		{
			name:  "standart input-2",
			input: input2.UserID,
			expected: []model.Position{{
				OperationID: input2.OperationID,
				UserID:      input2.UserID,
				Symbol:      input2.Symbol,
				OpenPrice:   input2.OpenPrice,
				ClosePrice:  closePrice2,
				Buy:         input2.Buy,
				Open:        false,
			}},
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

func TestGetOneState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	testTable := []struct {
		name     string
		input    uuid.UUID
		expected bool
	}{
		{
			name:     "std input-1",
			input:    input1.OperationID,
			expected: false,
		},
		{
			name:     "std input-2",
			input:    input2.OperationID,
			expected: false,
		},
		{
			name:     "std input-1",
			input:    input3.OperationID,
			expected: true,
		},
	}

	for _, test := range testTable {
		actual, err := conn.GetOneState(ctx, test.input)
		assert.Nil(t, err, test.name)
		assert.Equal(t, test.expected, actual, test.name)
	}

	log.Info("TestGetOneState finished!")
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

// func TestUpdate(t *testing.T) {

// }

// func TestGetAllOpened(t *testing.T) {

// }

// func TestGetOneState(t *testing.T) {

// }
