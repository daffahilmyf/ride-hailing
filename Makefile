PROTO_DIR=proto

.PHONY: proto proto-clean tidy

proto:
	protoc -I $(PROTO_DIR) \
	  --go_out=paths=source_relative:$(PROTO_DIR) \
	  --go-grpc_out=paths=source_relative:$(PROTO_DIR) \
	  $(PROTO_DIR)/ride/v1/ride.proto \
	  $(PROTO_DIR)/matching/v1/matching.proto \
	  $(PROTO_DIR)/location/v1/location.proto

proto-clean:
	rm -f \
	  $(PROTO_DIR)/ride/v1/*.pb.go \
	  $(PROTO_DIR)/matching/v1/*.pb.go \
	  $(PROTO_DIR)/location/v1/*.pb.go

tidy:
	cd $(PROTO_DIR) && go mod tidy
	cd services/gateway && go mod tidy
