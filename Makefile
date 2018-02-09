default:
	@export GOPATH=$$GOPATH:$$(pwd)  && go install main
run: default
	@bin/main
	@echo ""
clean:
	@rm -rf pkg bin
edit:
	@export GOPATH=$$GOPATH:$$(pwd) && atom .
test:
	@export GOPATH=$$GOPATH:$$(pwd) && go test ./...
export: clean default
	@mv bin/main bin/reporter
	@tar -czf bin/reporter_$$(date +"%Y%m%d_%H%M").tgz bin/reporter
setup:
	go get -u github.com/aws/aws-sdk-go/...
