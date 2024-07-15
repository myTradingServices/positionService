package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/config"
	"github.com/mmfshirokan/positionService/internal/consumer"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
	"github.com/mmfshirokan/positionService/internal/rpc"
	"github.com/mmfshirokan/positionService/internal/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Info("Starting position service")
	defer log.Info("Position service exited")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.New()
	if err != nil {
		log.Error("Config error:", err)
		return
	}

	conn, err := grpc.Dial(
		conf.PriceProviderURI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Errorf("grpc connection error on %v: %v", conf.PriceProviderURI, err)
		return
	}
	defer conn.Close()

	chPrice := make(chan model.Price)
	priceServRPC := rpc.NewPriceServer(conn, chPrice)
	priceData := repository.NewPrices(make(map[string]map[string]chan model.Price))
	priceDataServ := service.NewPrices(priceData)

	priceBrdg := consumer.NewPriceBridge(chPrice, priceDataServ)

	dbpool, err := pgxpool.New(ctx, conf.PostgresURI)
	if err != nil {
		log.Errorf("Error occurred while connecting yo postgresql pool: %v", err)
		return
	}
	defer dbpool.Close()

	openCh := make(chan model.Position)
	closeCh := make(chan model.Position)

	lis := repository.NewPgListen(openCh, closeCh, dbpool)

	posBridger := consumer.NewCloser(closeCh, openCh)

	localData := repository.NewLocalPosition(make(map[string]chan model.Position))
	localDataServ := service.NewLocalPositions(localData)

	posData := repository.NewPosition(dbpool)
	posDataServ := service.NewPosition(posData)

	opener := consumer.NewOpener(
		localDataServ,
		priceDataServ,
		posDataServ,
		openCh,
	)

	go opener.Open(ctx)
	go lis.Listen(ctx)
	go posBridger.Close(ctx)
	go priceServRPC.ReciveStream(ctx)
	go priceBrdg.PriceBridge(ctx)

	forever := make(chan struct{}) //destroy
	<-forever
}
