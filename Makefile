
GO=go
export GOBIN ?= $(PWD)/.bin

ARGS_TOOLS_MODFILE=-modfile go.tools.mod

run: generate-swagger
	@echo "🚀 launch..."
	@$(GO) run main.go

$(GOBIN)/swagger: 
	@echo "ℹ️  download go-swagger..."
	@$(GO) get $(ARGS_TOOLS_MODFILE) github.com/go-swagger/go-swagger/cmd/swagger

$(GOBIN)/air:
	@echo "ℹ️  download air..."
	@$(GO) get $(ARGS_TOOLS_MODFILE) github.com/cosmtrek/air

run: $(GOBIN)/air generate-swagger

generate-swagger: $(GOBIN)/swagger
	@echo "ℹ️  generate api..."
	@rm -rf internal/api2/gen && mkdir internal/api2/gen
	@$(GOBIN)/swagger generate server --quiet --spec internal/api2/swagger.yml --exclude-main --keep-spec-order --target=internal/api2/gen
	
