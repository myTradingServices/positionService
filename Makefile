start: 
	go run main.go
gen rpc:
	cd ./proto &&  protoc --go_out=. --go-grpc_out=. *.proto && cd ../
gen mock:
	mockery --dir ./internal/rpc --name PositionManipulator --name Reciver --output ./internal/rpc/mock --with-expecter;
	mockery --dir ./internal/consumer --all --output ./internal/consumer/mock --with-expecter;