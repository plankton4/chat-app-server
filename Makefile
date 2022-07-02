PROTO_DIR = ../chat-app-proto
GO_OUT_DIR = .

GO_OUT_BUILD:
	@echo "Сборка go протоколов"
	protoc -I=$(PROTO_DIR) --go_out=$(GO_OUT_DIR) $(PROTO_DIR)/common.proto
	
