package rpc

import (
	"sync"
	"time"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
)

type pricePosition struct {
	symbPrice     service.MapInterface[string]
	symbOperPrice service.MapInterface[model.SymbOperDTO]
	operPrice     service.MapInterface[string]
	wg            sync.WaitGroup
}

func NewPricePositionServer(sp service.MapInterface[string], sop service.MapInterface[model.SymbOperDTO], tmpOperPrice service.MapInterface[string]) *pricePosition {
	return &pricePosition{
		symbPrice:     sp,
		symbOperPrice: sop,
		operPrice:     tmpOperPrice,
	}
}

func (p *pricePosition) Mapper() {
	tmpDataChan := make(chan model.SymbOperDTO)
	go func(ch chan<- model.SymbOperDTO) {
		p.wg.Add(1)
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic was resived, now exiting Gorutine: ", err)
			}
			p.wg.Done()
		}()

		for {
			arr := p.symbOperPrice.GetKeys()
			for _, val := range arr {
				ok := p.operPrice.Contains(val.Operation)
				if !ok {
					priceCahn, err := p.symbPrice.Get(val.Symbol)
					if err != nil {
						log.Error(err)
					}

					p.operPrice.Add(
						val.Operation,
						priceCahn,
					)
					ch <- val
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}(tmpDataChan)

	go func(ch <-chan model.SymbOperDTO) {
		p.wg.Add(1)
		for {
			key, opend := <-ch
			if !opend {
				break
			}

			go func(k model.SymbOperDTO) {
				for {
					priceRecvChan, err := p.symbOperPrice.Get(k)
					if err != nil {
						log.Error("Error geting chan for symbOperPrice:", err)
						break
					}

					priceChan, err := p.operPrice.Get(key.Symbol)
					if err != nil {
						log.Error("Error geting chan for symbPrice:", err)
						break
					}

					priceRecvChan <- <-priceChan

					time.Sleep(time.Millisecond * 100)
				}
			}(key)
			time.Sleep(time.Millisecond * 100)
		}
		p.wg.Done()
	}(tmpDataChan)

	p.wg.Wait()
}

// Inside for loop:
// p.mut.RLock()
// for symb, operMap := range p.symbOperPrice {

// 	for _, priceChan := range operMap {
// 		p.mut.RUnlock()
// 		priceChan <- <-p.symbPrice[symb]
// 	}
// }
