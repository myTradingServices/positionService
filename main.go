package main

import (
	"context"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmfshirokan/positionService/internal/config"
	"github.com/mmfshirokan/positionService/internal/model"
	"github.com/mmfshirokan/positionService/internal/repository"
	"github.com/mmfshirokan/positionService/internal/rpc"
	"github.com/mmfshirokan/positionService/internal/service"
	"github.com/mmfshirokan/positionService/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	conf, err := config.New()
	if err != nil {
		log.Error("Config error:", err)
	}

	pgxPool, err := pgxpool.New(ctx, conf.PpstgresURI)
	if err != nil {
		log.Panic("pgxPool fatal error:", err)
	}

	pgxConn, err := pgxPool.Acquire(ctx)
	if err != nil {
		log.Panic("pgxPool fatal conection error:", err)
	}

	rpcOptions := grpc.WithTransportCredentials(insecure.NewCredentials())
	rpcConn, err := grpc.Dial(conf.PpstgresURI, rpcOptions)
	if err != nil {
		log.Panic("fail to launch rpc:", err)
	}
	defer rpcConn.Close()

	symbOperPriceMap := make(map[string]map[string]chan model.Price)
	symbPriceMap := make(map[string]chan model.Price)
	operPriceMap := make(map[string]chan model.Price)

	symbOperPriceServ := repository.NewSymbOperMap(symbOperPriceMap)
	symblPriceServ := repository.NewStringPrice(symbPriceMap)
	operPriceServ := repository.NewStringPrice(operPriceMap)

	dbRepo := repository.NewPostgresRepository(pgxConn)
	serv := service.NewPositionService(dbRepo)

	balanceChan := make(chan model.Position)

	mapper := rpc.NewPricePositionServer(symblPriceServ, symbOperPriceServ, operPriceServ)
	priceRPC := rpc.NewPriceServer(rpcConn, symblPriceServ)

	positionRPC := rpc.NewPositionServer(symbOperPriceServ, balanceChan, serv)
	balancerRPC := rpc.NewBalancerServer(balanceChan)

	go positionServerStart(conf.PositionServerURI, positionRPC)
	go balanceServerStart(conf.BalanceServerURI, balancerRPC)
	go priceRPC.Recive(ctx)
	go mapper.Mapper()
}

func positionServerStart(posServPort string, positionRPC pb.PositionServer) {
	lis, err := net.Listen("tcp", posServPort)
	if err != nil {
		log.Errorf("Critical error on listen port %v: %v", posServPort, err)
		return
	}

	rpcServer := grpc.NewServer()

	pb.RegisterPositionServer(rpcServer, positionRPC)

	err = rpcServer.Serve(lis)
	if err != nil {
		log.Errorf("Critical error on rpcServer start with port %v: %v", posServPort, err)
	}
}

func balanceServerStart(balanceServPort string, balanceRPC pb.BalanceServer) {
	lis, err := net.Listen("tcp", balanceServPort)
	if err != nil {
		log.Errorf("Critical error on listen port %v: %v", balanceServPort, err)
		return
	}

	rpcServer := grpc.NewServer()

	pb.RegisterBalanceServer(rpcServer, balanceRPC)

	err = rpcServer.Serve(lis)
	if err != nil {
		log.Errorf("Critical error on rpcServer start with port %v: %v", balanceServPort, err)
	}
}
