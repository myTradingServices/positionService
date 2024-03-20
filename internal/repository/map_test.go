package repository

import (
	"sync"
	"testing"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	connMap    MapInterface[string]
	connMapMap MapInterface[model.SymbOperDTO]

	mapInputValue1 = make(chan model.Price)
	mapInputValue2 = make(chan model.Price)
	mapInputValue3 = make(chan model.Price)
)

func TestMapAdd(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      string
			value    chan model.Price
			hasError bool
		}
		testTable := []T{
			{
				name:     "map standart input-1",
				key:      "symb1",
				value:    mapInputValue1,
				hasError: false,
			},
			{
				name:     "map standart input-2",
				key:      "symb2",
				value:    mapInputValue2,
				hasError: false,
			},
			{
				name:     "map standart input-3",
				key:      "symb3",
				value:    mapInputValue3,
				hasError: false,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				err := connMap.Add(test.key, test.value)
				assert.Nil(t, err, test.name)
			}(testCase)
		}

		innerWG.Wait()

		err := connMap.Add("symb1", make(chan model.Price))
		assert.Error(t, err, "map error input")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      model.SymbOperDTO
			value    chan model.Price
			hasError bool
		}
		testTable := []T{
			{
				name: "map map standart input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "oper1",
				},
				value:    mapInputValue1,
				hasError: false,
			},
			{
				name: "map map standart input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb2",
					Operation: "oper2",
				},
				value:    mapInputValue2,
				hasError: false,
			},
			{
				name: "map map standart input-3",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper3",
				},
				value:    mapInputValue3,
				hasError: false,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				err := connMapMap.Add(test.key, test.value)
				assert.Nil(t, err, test.name)
			}(testCase)
		}

		innerWG.Wait()

		err := connMapMap.Add(model.SymbOperDTO{
			Symbol:    "symb1",
			Operation: "oper1",
		}, make(chan model.Price))
		assert.Error(t, err, "map map error input")
	}()

	wg.Wait()

	log.Info("TestMapAdd finished!")
}

func TestMapGet(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      string
			expected chan model.Price
			hasError bool
		}
		testTable := []T{
			{
				name:     "map standart input-1",
				key:      "symb1",
				expected: mapInputValue1,
				hasError: false,
			},
			{
				name:     "map standart input-2",
				key:      "symb2",
				expected: mapInputValue2,
				hasError: false,
			},
			{
				name:     "map standart input-3",
				key:      "symb3",
				expected: mapInputValue3,
				hasError: false,
			},
			{
				name:     "map error input-1",
				key:      "symb44",
				expected: nil,
				hasError: true,
			},
			{
				name:     "map error input-2",
				key:      "sym3",
				expected: nil,
				hasError: true,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				actual, err := connMap.Get(test.key)
				if test.hasError {
					assert.Error(t, err, test.name)
				} else {
					if ok := assert.Nil(t, err, test.name); !ok {
						return
					}
					assert.Equal(t, test.expected, actual, test.name)
				}
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      model.SymbOperDTO
			expected chan model.Price
			hasError bool
		}
		testTable := []T{
			{
				name: "map map standart input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "oper1",
				},
				expected: mapInputValue1,
				hasError: false,
			},
			{
				name: "map map standart input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb2",
					Operation: "oper2",
				},
				expected: mapInputValue2,
				hasError: false,
			},
			{
				name: "map map standart input-3",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper3",
				},
				expected: mapInputValue3,
				hasError: false,
			},
			{
				name: "map map error input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "non_exist_operation",
				},
				expected: nil,
				hasError: true,
			},
			{
				name: "map map error input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper2",
				},
				expected: nil,
				hasError: true,
			},
			{
				name: "map map error input-3",
				key: model.SymbOperDTO{
					Symbol:    "non_exist_symb",
					Operation: "oper3",
				},
				expected: nil,
				hasError: true,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				actual, err := connMapMap.Get(test.key)
				if test.hasError {
					assert.Error(t, err, test.name)
				} else {
					if ok := assert.Nil(t, err, test.name); !ok {
						return
					}
					assert.Equal(t, test.expected, actual, test.name)
				}
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Wait()

	log.Info("TestMapGet finished!")
}

func TestContains(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      string
			expected bool
		}
		testTable := []T{
			{
				name:     "map standart input-1",
				key:      "symb1",
				expected: true,
			},
			{
				name:     "map standart input-2",
				key:      "symb2",
				expected: true,
			},
			{
				name:     "map standart input-3",
				key:      "symb3",
				expected: true,
			},
			{
				name:     "map error input-1",
				key:      "symb44",
				expected: false,
			},
			{
				name:     "map error input-2",
				key:      "sym3",
				expected: false,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				actual := connMap.Contains(test.key)
				assert.Equal(t, test.expected, actual, test.name)
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      model.SymbOperDTO
			expected bool
		}
		testTable := []T{
			{
				name: "map map standart input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "oper1",
				},
				expected: true,
			},
			{
				name: "map map standart input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb2",
					Operation: "oper2",
				},
				expected: true,
			},
			{
				name: "map map standart input-3",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper3",
				},
				expected: true,
			},
			{
				name: "map map error input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "non_exist_operation",
				},
				expected: false,
			},
			{
				name: "map map error input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper2",
				},
				expected: false,
			},
			{
				name: "map map error input-3",
				key: model.SymbOperDTO{
					Symbol:    "non_exist_symb",
					Operation: "oper3",
				},
				expected: false,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				actual := connMapMap.Contains(test.key)
				assert.Equal(t, test.expected, actual, test.name)
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Wait()

	log.Info("TestMapContains finished!")
}

func TestGetKeys(t *testing.T) {
	actualMap := connMap.GetKeys()
	assert.ElementsMatch(t, []string{
		"symb1",
		"symb2",
		"symb3",
	}, actualMap)

	actualMapMap := connMapMap.GetKeys()
	assert.ElementsMatch(t, []model.SymbOperDTO{
		{
			Symbol:    "symb1",
			Operation: "oper1",
		},
		{
			Symbol:    "symb2",
			Operation: "oper2",
		},
		{
			Symbol:    "symb3",
			Operation: "oper3",
		},
	}, actualMapMap)

	log.Info("TestGetKeys finished!")
}

func TestDelete(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      string
			hasError bool
		}
		testTable := []T{
			{
				name:     "map standart input-1",
				key:      "symb1",
				hasError: false,
			},
			{
				name:     "map standart input-2",
				key:      "symb2",
				hasError: false,
			},
			{
				name:     "map standart input-3",
				key:      "symb3",
				hasError: false,
			},
			{
				name:     "map error input-1",
				key:      "symb44",
				hasError: true,
			},
			{
				name:     "map error input-2",
				key:      "sym3",
				hasError: true,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				err := connMap.Delete(test.key)
				if test.hasError {
					assert.Error(t, err, test.name)
				} else {
					assert.Nil(t, err, test.name)
				}
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		type T struct {
			name     string
			key      model.SymbOperDTO
			hasError bool
		}
		testTable := []T{
			{
				name: "map map standart input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "oper1",
				},
				hasError: false,
			},
			{
				name: "map map standart input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb2",
					Operation: "oper2",
				},
				hasError: false,
			},
			{
				name: "map map standart input-3",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper3",
				},
				hasError: false,
			},
			{
				name: "map map error input-1",
				key: model.SymbOperDTO{
					Symbol:    "symb1",
					Operation: "non_exist_operation",
				},
				hasError: true,
			},
			{
				name: "map map error input-2",
				key: model.SymbOperDTO{
					Symbol:    "symb3",
					Operation: "oper2",
				},
				hasError: true,
			},
			{
				name: "map map error input-3",
				key: model.SymbOperDTO{
					Symbol:    "non_exist_symb",
					Operation: "oper3",
				},
				hasError: true,
			},
		}

		var innerWG sync.WaitGroup

		for _, testCase := range testTable {
			innerWG.Add(1)

			go func(test T) {
				defer innerWG.Done()
				err := connMapMap.Delete(test.key)
				if test.hasError {
					assert.Error(t, err, test.name)
				} else {
					assert.Nil(t, err, test.name)
				}
			}(testCase)
		}

		innerWG.Wait()
	}()

	wg.Wait()

	log.Info("TestMapDelete finished!")
}
