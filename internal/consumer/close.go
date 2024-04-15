package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
)

type Close struct {
	posBridgeCh chan model.Position
	lisCloseCh  chan model.Position
}

func NewCloser(closeChanel chan model.Position, positionBridgeChanel chan model.Position) *Close {
	return &Close{
		posBridgeCh: positionBridgeChanel,
		lisCloseCh:  closeChanel,
	}
}

func (c *Close) Close(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case pos := <-c.lisCloseCh:
			{
				c.posBridgeCh <- pos
			}
		}
	}
}
