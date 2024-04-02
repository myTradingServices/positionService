package consumer

// import (
// 	"context"

// 	"github.com/mmfshirokan/positionService/internal/model"
// 	"github.com/mmfshirokan/positionService/internal/service"
// )

// type closer struct {
// 	posMap   service.UserMapInterface
// 	closeCh  chan model.Position
// 	priceMap service.PriceMapInterface
// }

// type Closer interface {
// 	Close(ctx context.Context)
// }

// func NewCloser(posMap service.UserMapInterface, closeCh chan model.Position, priceMap service.PriceMapInterface) Closer {
// 	return &closer{
// 		posMap:   posMap,
// 		closeCh:  closeCh,
// 		priceMap: priceMap,
// 	}
// }

// // Закрывает канал если его нету поднимает панику
// // Т.к. закрытие для канала пришло но канала до этого не было
// func (c *closer) Close(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			{
// 				return
// 			}
// 		case pos := <-c.closeCh:
// 			{
// 				ch, ok := c.posMap.GetOrCreate(pos.UserID.String())
// 				if !ok {
// 					panic("Fatal error closing chanel for position that does not exist")
// 				}

// 				c.priceMap.Delete(model.SymbOperDTO{
// 					Symbol: pos.Symbol,
// 					UserID: pos.UserID.String(),
// 				})

// 				close(ch)
// 			}
// 		default:
// 		}
// 	}
// }
