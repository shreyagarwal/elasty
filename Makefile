.PHONY: build doc fmt lint run test buildall buildlinux githubrelease clean deb

# Prepend our _vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.

default: build

build:
	goimports -v -w ./*.go
	go build -v -ldflags "-s -w" -o _release/elasty

buildall:
	goimports -v -w ./*.go
	gox -verbose -ldflags "-s -w" -output="_release/{{.Dir}}_{{.OS}}_{{.Arch}}"

buildlinux:
	goimports -v -w ./*.go
	gox -osarch="linux/amd64" -ldflags "-s -w" -verbose -parallel=1 -output="_release/{{.Dir}}_{{.OS}}_{{.Arch}}"

githubrelease:
	goimports -v -w ./*.go

	# clean
	rm -rf ./_release/*

	# create for darwin/amd64
	mkdir -p ./_release/elasty_darwin_amd64
	go build -v -ldflags "-s -w" -o _release/elasty_darwin_amd64/elasty_darwin_amd64
	# copy config files
	cp -r ./config  _release/elasty_darwin_amd64/
	# create Tar to release
	tar cvfz _release/elasty_darwin_amd64.tar.gz  -C _release/elasty_darwin_amd64 .
	# delete directory
	rm -rf _release/elasty_darwin_amd64

	# create for linux/amd64
	mkdir -p ./_release/elasty_linux_amd64
	gox -osarch="linux/amd64" -ldflags "-s -w" -verbose -parallel=1 -output="_release/elasty_linux_amd64/{{.Dir}}_{{.OS}}_{{.Arch}}"
	# copy config files
	cp -r ./config  _release/elasty_linux_amd64/
	# create Tar to release
	tar cvfz _release/elasty_linux_amd64.tar.gz -C _release/elasty_linux_amd64 .
	# delete directory
	rm -rf _release/elasty_linux_amd64


# make deb file
deb:
	# create folder structure
	# |-- DEBIAN
	# |   `-- control
	# |-- etc
	# |   |-- elasty
	# |   |-- init.d
	# |   `-- logrotate.d
	# |-- usr
	# |   |-- sbin
	# |   |   `-- elasty
	# |   `-- share
	# |       `-- doc
	# |           `-- elasty
	# |               |-- changelog.Debian.gz
	# |               |-- changelog.gz
	# |               |-- copyright
	# |               `-- README
	# `-- var
	#     `-- log
	#         `-- elasty

	# first build for linux amd64

	# copy from skeleton folder structure

	# Edit control file to reflect version

	# copy compiled binary

	# copy ReadME

	# create Deb

	# remove copied folder structure

clean:
	rm -r _release/*

#doc:
#    godoc -http=:6060 -index

fmt:
	go fmt -v -w ./*.go

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
#lint:
#    golint ./src


#test:
#    go test ./src/...


