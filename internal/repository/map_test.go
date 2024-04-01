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
	connMap MapInterface

	mapInputValue1 = make(chan model.Price)
	mapInputValue2 = make(chan model.Price)
	mapInputValue3 = make(chan model.Price)
	mapInputValue4 = make(chan model.Price)
	mapInputValue5 = make(chan model.Price)
)

func TestMapAdd(t *testing.T) {

	type T struct {
		name  string
		key   model.SymbOperDTO
		value chan model.Price
	}
	testTable := []T{
		{
			name: "standart input-1",
			key: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: "oper1",
			},
			value: mapInputValue1,
		},
		{
			name: "standart input-2",
			key: model.SymbOperDTO{
				Symbol: "symb2",
				UserID: "oper2",
			},
			value: mapInputValue2,
		},
		{
			name: "standart input-3",
			key: model.SymbOperDTO{
				Symbol: "symb3",
				UserID: "oper3",
			},
			value: mapInputValue3,
		},
		{
			name: "standart input-4",
			key: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: "oper4",
			},
			value: mapInputValue4,
		},
		{
			name: "standart input-5",
			key: model.SymbOperDTO{
				Symbol: "symb1",
				UserID: "oper5",
			},
			value: mapInputValue5,
		},
	}

	var wg sync.WaitGroup

	for _, testCase := range testTable {
		wg.Add(1)

		go func(test T) {
			defer wg.Done()
			err := connMap.Add(test.key, test.value)
			assert.Nil(t, err, test.name)
		}(testCase)
	}

	wg.Wait()

	err := connMap.Add(model.SymbOperDTO{
		Symbol: "symb1",
		UserID: "oper1",
	}, make(chan model.Price))
	assert.Error(t, err, "error input")

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
				mapInputValue1,
				mapInputValue4,
				mapInputValue5,
			},
			hasError: false,
		},
		{
			name: "standart input-2",
			symb: "symb2",
			expected: []chan model.Price{
				mapInputValue2,
			},
			hasError: false,
		},
		{
			name: "standart input-3",
			symb: "symb3",
			expected: []chan model.Price{
				mapInputValue3,
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
				actual, err := connMap.GetAllChanForSymb(test.symb)
				testMsg := fmt.Sprint(test.name, "loop", lp)
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
	log.Info("TestGetAllChanFromSymb finished!")
}
