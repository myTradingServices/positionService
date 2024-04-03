package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/service"
)

type closer struct {
	posMap  service.PositionMapInterface
	closeCh chan model.Position
}

type Closer interface {
	Close(ctx context.Context)
}

func NewCloser(posMap service.PositionMapInterface, closeCh chan model.Position, priceMap service.PriceMapInterface) Closer {
	return &closer{
		posMap:  posMap,
		closeCh: closeCh,
	}
}

func (c *closer) Close(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case pos := <-c.closeCh:
			{
				ch, ok := c.posMap.Get(pos.UserID.String())
				if !ok {
					panic("Fatal error closing chanel for position that does not exist")
				}

				ch <- pos
			}
		default:
		}
	}
}
