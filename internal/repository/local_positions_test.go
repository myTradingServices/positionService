package repository

import (
	"sync"
	"testing"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	positionMapForTesting map[string]chan model.Position
	posMapConn            *Positions

	posMapInputValue1 = make(chan model.Position)
	posMapInputValue2 = make(chan model.Position)
	posMapInputValue3 = make(chan model.Position)
	posMapInputValue4 = make(chan model.Position)
	posMapInputValue5 = make(chan model.Position)
)

func TestPosMapAdd(t *testing.T) {

	type T struct {
		name   string
		userID string
		value  chan model.Position
	}
	testTable := []T{
		{
			name:   "standart input-1",
			userID: "user1",
			value:  posMapInputValue1,
		},
		{
			name:   "standart input-2",
			userID: "user2",
			value:  posMapInputValue2,
		},
		{
			name:   "standart input-3",
			userID: "user3",
			value:  posMapInputValue3,
		},
		{
			name:   "standart input-4",
			userID: "user4",
			value:  posMapInputValue4,
		},
		{
			name:   "standart input-5",
			userID: "user5",
			value:  posMapInputValue5,
		},
	}

	var wg sync.WaitGroup
	for _, testCase := range testTable {
		wg.Add(1)
		go func(test T) {
			defer wg.Done()
			posMapConn.Add(test.userID, test.value)
		}(testCase)
	}
	wg.Wait()
	log.Info("TestPosMapAdd finished!")
}

func TestPosMapGet(t *testing.T) {

	type T struct {
		name   string
		userID string
		value  chan model.Position
		ok     bool
	}

	testTable := []T{
		{
			name:   "standart input-1",
			userID: "user1",
			value:  posMapInputValue1,
			ok:     true,
		},
		{
			name:   "standart input-2",
			userID: "user2",
			value:  posMapInputValue2,
			ok:     true,
		},
		{
			name:   "standart input-3",
			userID: "user3",
			value:  posMapInputValue3,
			ok:     true,
		},
		{
			name:   "standart input-4",
			userID: "user4",
			value:  posMapInputValue4,
			ok:     true,
		},
		{
			name:   "standart input-5",
			userID: "user5",
			value:  posMapInputValue5,
			ok:     true,
		},
		{
			name:   "not ok input-1",
			userID: "user124",
			value:  nil,
			ok:     false,
		},
		{
			name:   "not ok input-2",
			userID: "blablabla",
			value:  nil,
			ok:     false,
		},
		{
			name:   "not ok input-3",
			userID: "blablASDdsadASDabla",
			value:  nil,
			ok:     false,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testTable {
		wg.Add(1)
		go func(test T) {
			defer wg.Done()
			ch, ok := posMapConn.Get(test.userID)
			assert.Equal(t, test.ok, ok, test.name)
			assert.Equal(t, test.value, ch, test.name)

		}(testCase)
	}

	wg.Wait()
	log.Info("TestPosMapGet finished!")
}

func TestPosMapDelete(t *testing.T) {
	testTable := []string{
		"user1",
		"user2",
		"user3",
		"user4",
		"user5",
	}

	var wg sync.WaitGroup

	for _, key := range testTable {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			posMapConn.Delete(userID)
		}(key)
	}
	wg.Wait()
	for _, userID := range testTable {
		if _, ok := positionMapForTesting[userID]; ok {
			t.Error("Folowing userID is not deleted: ", userID)
		}
	}

	log.Info("TestPosMapDelete finished!")
}
