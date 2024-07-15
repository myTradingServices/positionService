package consumer

import (
	"context"

	"github.com/mmfshirokan/positionService/internal/model"
	log "github.com/sirupsen/logrus"
)

type Close struct {
	mainCH     chan model.Position
	lisCloseCh chan model.Position
}

func NewCloser(closeChanel chan model.Position, positionBridgeChanel chan model.Position) *Close {
	return &Close{
		mainCH:     positionBridgeChanel,
		lisCloseCh: closeChanel,
	}
}

func (c *Close) Close(ctx context.Context) {
	log.Info("Close consumer started")
	defer log.Info("Close-consumer exited")

	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case pos := <-c.lisCloseCh:
			{
				log.Infof("Close-position received: %v, now sent to %v", c.lisCloseCh, c.mainCH)
				c.mainCH <- pos
				log.Infof("Close-position sent to %v compleet", c.mainCH)
			}
		}
	}
}
