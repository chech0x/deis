include ../includes.mk

# the filepath to this repository, relative to $GOPATH/src
repo_path = github.com/deis/deis/client

GO_FILES = $(wildcard *.go)
GO_PACKAGES = parser cmd controller/api controller/client $(wildcard controller/models/*) $(wildcard pkg/*)
GO_PACKAGES_REPO_PATH = $(addprefix $(repo_path)/,$(GO_PACKAGES))

COMPONENT = $(notdir $(repo_path))
IMAGE =  $(IMAGE_PREFIX)/$(COMPONENT):$(BUILD_TAG)

build:
	CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags '-s' -o talka .
	@$(call check-static-binary,talka)

install: build
	cp deis $$GOPATH/bin

installer: build
	@if [ ! -d makeself ]; then git clone -b single-binary https://github.com/deis/makeself.git; fi
	PATH=./makeself:$$PATH BINARY=talka makeself.sh --bzip2 --current --nox11 . \
		talka-cli-`cat talka-version`-`go env GOOS`-`go env GOARCH`.run \
		"Talka CLI" "echo \
		&& echo 'Talka se encuentra en el directorio actual. Por favor' \
		&& echo 'mueva talka a un directorio que se encuentre agregado' \
		&& echo 'en la variable de abiente PATH.' \
		&& echo \
		&& echo 'Vea http://talka.cl/ para más documentación.' \
		&& echo"

setup-root-gotools:
	sudo GOPATH=/tmp/tmpGOPATH go get -u -v golang.org/x/tools/cmd/cover
	sudo GOPATH=/tmp/tmpGOPATH go get -u -v golang.org/x/tools/cmd/vet
	sudo rm -rf /tmp/tmpGOPATH

setup-gotools:
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/cover
	go get -u golang.org/x/tools/cmd/vet

test: test-style test-unit

test-style:
# display output, then check
	$(GOFMT) $(GO_PACKAGES) $(GO_FILES)
	@$(GOFMT) $(GO_PACKAGES) $(GO_FILES) | read; if [ $$? == 0 ]; then echo "gofmt check failed."; exit 1; fi
	$(GOVET) $(repo_path) $(GO_PACKAGES_REPO_PATH)
	$(GOLINT) ./...

test-unit:
	$(GOTEST) ./...
