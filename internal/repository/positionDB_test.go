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

var (
	conn DBInterface

	input1 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb1",
		OpenPrice:   decimal.New(19, 0),
		Long:        true,
	}

	input2 = model.Position{
		OperationID: uuid.New(),
		UserID:      uuid.New(),
		Symbol:      "symb2",
		OpenPrice:   decimal.New(123, 0),
		Long:        false,
	}

	input3 = model.Position{
		OperationID: uuid.New(),
		UserID:      input1.UserID,
		Symbol:      "symb3",
		OpenPrice:   decimal.New(190, 0),
		Long:        true,
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

	openCh = make(chan model.Position)
	closeCh = make(chan model.Position)
	lis := NewPgListen(openCh, closeCh, dbpool)
	go lis.Listen(ctx)
	time.Sleep(time.Second)

	positionMapForTesting = make(map[string]chan model.Position)
	posMapConn = NewPositionMap(positionMapForTesting)

	priceMapForTesting = make(map[string]map[string]chan model.Price)
	priceMapConn = NewSymbOperMap(priceMapForTesting)

	code := m.Run()

	if err := pool.Purge(pgResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestAdd(t *testing.T) {
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

	log.Info("TestAdd finished!")
}

func TestGetAllOpened(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	opened, err := conn.GetAllOpened(ctx)
	assert.Nil(t, err, "GetAllOpened error is not nil")
	assert.ElementsMatch(t, opened, []model.Position{
		{
			UserID:    input1.UserID,
			Symbol:    input1.Symbol,
			Long:      input1.Long,
			OpenPrice: input1.OpenPrice,
		},
		{
			UserID:    input2.UserID,
			Symbol:    input2.Symbol,
			Long:      input2.Long,
			OpenPrice: input2.OpenPrice,
		},
		{
			UserID:    input3.UserID,
			Symbol:    input3.Symbol,
			Long:      input3.Long,
			OpenPrice: input3.OpenPrice,
		},
	})

	log.Info("TestGetAllOpened finished!")
}

// Note: test worling correctly even if input is incorrect
// Solutuion: change input data from (model.Position) to (userID, symbol, closePrice)
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
				UserID:     input1.UserID,
				Symbol:     input1.Symbol,
				ClosePrice: closePrice1,
			},
		},
		{
			name: "std input-2",
			input: model.Position{
				UserID:     input2.UserID,
				Symbol:     input2.Symbol,
				ClosePrice: closePrice2,
			},
		},
		{
			name: "std input-3",
			input: model.Position{
				UserID:     input3.UserID,
				Symbol:     input3.Symbol,
				ClosePrice: closePrice3,
			},
		},
	}

	for _, test := range testTable {
		err := conn.Update(ctx, test.input)

		assert.Nil(t, err, test.name)
	}

	log.Info("TestUpdate finished!")
}

func TestGet(t *testing.T) {
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
					Long:        input1.Long,
				},
				{
					OperationID: input3.OperationID,
					UserID:      input3.UserID,
					Symbol:      input3.Symbol,
					OpenPrice:   input3.OpenPrice,
					ClosePrice:  closePrice3,
					Long:        input3.Long,
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
				Long:        input2.Long,
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

			for i, actVal := range actual {
				assert.Equal(t, test.expected[i].Long, actVal.Long, test.name)
				assert.Equal(t, test.expected[i].ClosePrice, actVal.ClosePrice, test.name)
				assert.Equal(t, test.expected[i].OpenPrice, actVal.OpenPrice, test.name)
				assert.Equal(t, test.expected[i].OperationID, actVal.OperationID, test.name)
				assert.Equal(t, test.expected[i].Symbol, actVal.Symbol, test.name)
				assert.Equal(t, test.expected[i].UserID, actVal.UserID, test.name)
				if ok := actVal.CreatedAt.After(time.Now().Add(-time.Millisecond * 50)); !ok {
					t.Errorf("CreatedAt is not valid: %v.\n Created time: %v, Now: %v", test.name, actVal.CreatedAt.String(), time.Now())
				}
			}
		}
	}

	log.Info("TestGet finished!")
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

	log.Info("TestDelete finished!")
}
