# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

#export CGO_ENABLED:=0

.PHONY: all
all: build

.PHONY: mod
mod:
	./hack/go-mod.sh

.PHONY: format
format:
	./hack/go-fmt.sh

.PHONY: go-generate
go-generate: mod
	$(Q)go generate ./...

.PHONY: vet
vet:
	./hack/go-vet.sh

.PHONY: test
test:
	./hack/go-test.sh

.PHONY: lint
lint:
	./hack/go-lint.sh
	#./hack/yaml-lint.sh

.PHONY: build
build:
	./hack/go-build.sh

.PHONY: clean
clean:
	rm -rf build/_output

.PHONY: addheaders
addheaders:
	./hack/addheaders.sh

.PHONY: install
install:
	./hack/install.sh

.PHONY: uninstall
uninstall:
	./hack/uninstall.sh

.PHONY: prepare-olm
version = ""
prepare-olm:
	./hack/pr-operatorhub.sh $(version)