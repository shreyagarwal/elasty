.PHONY: build doc fmt lint run test buildall buildlinux

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.

default: build

build:
	goimports -v -w ./*.go
	go build -v -o _release/elasty

buildall:
	goimports -v -w ./*.go
	gox -verbose -output="_release/{{.Dir}}_{{.OS}}_{{.Arch}}"

buildlinux:
	goimports -v -w ./*.go
	gox -osarch="linux/amd64" -verbose -parallel=1 -output="_release/{{.Dir}}_{{.OS}}_{{.Arch}}"

githubrelease:
	goimports -v -w ./*.go

	# clean
	rm -rf ./_release/*

	# create for darwin/amd64
	mkdir -p ./_release/elasty_darwin_amd64
	go build -v -o _release/elasty_darwin_amd64/elasty_darwin_amd64
	# copy config files
	cp -r ./config  _release/elasty_darwin_amd64/
	# create Tar to release
	tar cvfz _release/elasty_darwin_amd64.tar.gz  -C _release/elasty_darwin_amd64 .
	# delete directory
	rm -rf _release/elasty_darwin_amd64

	# create for linux/amd64
	mkdir -p ./_release/elasty_linux_amd64
	gox -osarch="linux/amd64" -verbose -parallel=1 -output="_release/elasty_linux_amd64/{{.Dir}}_{{.OS}}_{{.Arch}}"
	# copy config files
	cp -r ./config  _release/elasty_linux_amd64/
	# create Tar to release
	tar cvfz _release/elasty_linux_amd64.tar.gz -C _release/elasty_linux_amd64 .
	# delete directory
	rm -rf _release/elasty_linux_amd64



clean:
	rm -r _release/*

#doc:
#    godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
#fmt:
#    go fmt ./src/...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
#lint:
#    golint ./src

#run: build
#    ./bin/main_app

#test:
#    go test ./src/...


