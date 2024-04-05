package repository

import (
	"fmt"
	"sync"
	"testing"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	priceMapConn       PriceMapInterface
	priceMapForTesting map[string]map[string]chan model.Price

	priceMapInputValue1 = make(chan model.Price)
	pmKey1              = model.SymbOperDTO{
		Symbol: "symb1",
		UserID: "user1",
	}

	priceMapInputValue2 = make(chan model.Price)
	pmKey2              = model.SymbOperDTO{
		Symbol: "symb2",
		UserID: "user1",
	}

	priceMapInputValue3 = make(chan model.Price)
	pmKey3              = model.SymbOperDTO{
		Symbol: "symb3",
		UserID: "user1",
	}

	priceMapInputValue4 = make(chan model.Price)
	pmKey4              = model.SymbOperDTO{
		Symbol: "symb1",
		UserID: "user2",
	}

	priceMapInputValue5 = make(chan model.Price)
	pmKey5              = model.SymbOperDTO{
		Symbol: "symb1",
		UserID: "user3",
	}
)

func TestPriceMapAdd(t *testing.T) {

	type T struct {
		name  string
		key   model.SymbOperDTO
		value chan model.Price
	}
	testTable := []T{
		{
			name:  "standart input-1 with symb1",
			key:   pmKey1,
			value: priceMapInputValue1,
		},
		{
			name:  "standart input-2",
			key:   pmKey2,
			value: priceMapInputValue2,
		},
		{
			name:  "standart input-3",
			key:   pmKey3,
			value: priceMapInputValue3,
		},
		{
			name:  "standart input-4",
			key:   pmKey4,
			value: priceMapInputValue4,
		},
		{
			name:  "standart input-5",
			key:   pmKey5,
			value: priceMapInputValue5,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testTable {
		wg.Add(1)

		go func(test T) {
			defer wg.Done()
			priceMapConn.Add(test.key, test.value)
		}(testCase)
	}

	wg.Wait()
	log.Info("TestMapAdd finished!")
}

func TestGetAllChanForSymb(t *testing.T) {

	type T struct {
		name     string
		symb     string
		expected []chan model.Price
		hasError bool
	}

	testTable := []T{
		{
			name: "standart output-1-4-5",
			symb: "symb1",
			expected: []chan model.Price{
				priceMapInputValue1,
				priceMapInputValue4,
				priceMapInputValue5,
			},
			hasError: false,
		},
		{
			name: "standart input-2",
			symb: "symb2",
			expected: []chan model.Price{
				priceMapInputValue2,
			},
			hasError: false,
		},
		{
			name: "standart input-3",
			symb: "symb3",
			expected: []chan model.Price{
				priceMapInputValue3,
			},
			hasError: false,
		},
		{
			name:     "error input",
			symb:     "symb69",
			expected: nil,
			hasError: true,
		},
	}

	var wg sync.WaitGroup

	//In the code below extra loop were added to reapet test process 100 times to be shure that ouptut is stable
	for i := 0; i < 100; i++ {
		for _, testCase := range testTable {
			wg.Add(1)
			go func(test T, lp int) {
				defer wg.Done()
				actual, err := priceMapConn.GetAllChanForSymb(test.symb)
				testMsg := fmt.Sprint(test.name, " loop-", lp)
				if test.hasError {
					assert.Error(t, err, testMsg)
					assert.Nil(t, actual, testMsg)
					return
				}
				ok := assert.Nil(t, err, testMsg)
				if !ok {
					return
				}
				assert.ElementsMatch(t, test.expected, actual, testMsg)
			}(testCase, i)
		}
	}

	wg.Wait()
	log.Info("Test GetAllChanFromSymb finished!")
}

func TestPriceMapGet(t *testing.T) {
	type T struct {
		name     string
		key      model.SymbOperDTO
		expected chan model.Price
	}

	testTable := []T{
		{
			name:     "standart input-1",
			key:      pmKey1,
			expected: priceMapInputValue1,
		},
		{
			name:     "standart input-2",
			key:      pmKey2,
			expected: priceMapInputValue2,
		},
		{
			name:     "standart input-3",
			key:      pmKey3,
			expected: priceMapInputValue3,
		},
		{
			name:     "standart input-4",
			key:      pmKey4,
			expected: priceMapInputValue4,
		},
		{
			name:     "standart input-5",
			key:      pmKey5,
			expected: priceMapInputValue5,
		},
		{
			name: "non exist key",
			key: model.SymbOperDTO{
				Symbol: "non-exist",
				UserID: "user1",
			},
			expected: nil,
		},
		{
			name: "empty key",
			key: model.SymbOperDTO{
				Symbol: "",
				UserID: "",
			},
			expected: nil,
		},
	}

	var wg sync.WaitGroup
	for _, testCase := range testTable {
		wg.Add(1)
		go func(test T) {
			defer wg.Done()
			actual := priceMapConn.Get(test.key)
			assert.Equal(t, test.expected, actual, test.name)
		}(testCase)
	}
	wg.Wait()
	log.Info("TestPriceMapGet finished!")
}

func TestPriceMapDelete(t *testing.T) {
	testSlice := []model.SymbOperDTO{
		pmKey1,
		pmKey2,
		pmKey3,
		pmKey4,
		pmKey5,
	}

	var wg sync.WaitGroup
	for _, key := range testSlice {
		wg.Add(1)
		go func(key model.SymbOperDTO) {
			defer wg.Done()
			priceMapConn.Delete(key)
		}(key)
	}
	wg.Wait()

	for _, key := range testSlice {
		_, ok := priceMapForTesting[key.Symbol][key.UserID]
		if ok {
			t.Error("Folowing key is not deleted: ", key)
		}
	}

	log.Info("TestPriceMapDelete finished!")
}
