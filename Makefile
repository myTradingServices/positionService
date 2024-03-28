start: 
	go run main.go
gen rpc:
	cd ./proto &&  protoc --go_out=. --go-grpc_out=. *.proto && cd ../
gen mock:
	mockery --dir ./internal/service --name DBInterface --name MapInterface --output ./internal/rpc/mock --with-expecter;