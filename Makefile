APPNAME?="urlshortener"
REPONAME?="url-shortener"
TEST?=$$(go list ./... | grep -v '/vendor/')
VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
REV?=$$(git rev-parse --short HEAD)
BRANCH?=$$(git rev-parse --abbrev-ref HEAD)
BUILDFILES?=$$(find bin -mindepth 1 -maxdepth 1 -type f)
VERSION?="0.0.1"
MAIN_PATH="./cmd/url-shortener/"
SCRIPT_PATH="./script"
DOCKER_REPO?="${REPONAME}/${APPNAME}"
XC_OS=$$(go env GOOS)
XC_ARCH=$$(go env GOARCH)
SHELL := /bin/bash


default: quick

lazy: version fmt lint vet test

quick: version fmt lint vet docker-build

vendor:
	@dep ensure

version:
	@echo "SOFTWARE VERSION"
	@echo "\tbranch:\t\t" ${BRANCH}
	@echo "\trevision:\t" ${REV}
	@echo "\tversion:\t" ${VERSION}

ci: buildonly generate-package
	@echo "CI BUILD..."

generate-package:
	@echo "GENERATE PACKAGE..."
	bash script/build-package.sh full


tools:
	@echo "GO TOOLS installation..."
	@go get -u golang.org/x/tools/cmd/cover
	@go get -u github.com/golang/lint/golint
	@curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

build: version vendor test
	@echo "GO BUILD..."
	@CGO_ENABLED=0 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=${XC_OS}/${XC_ARCH}" -v -o ./bin/${APPNAME} ${MAIN_PATH}

buildonly: vendor
	@CGO_ENABLED=0 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=${XC_OS}/${XC_ARCH}" -v -o ./bin/${APPNAME} ${MAIN_PATH}

crosscompile: vendor
	@./script/docker-build.sh linux-build
	@./script/docker-build.sh darwin-build
	@./script/docker-build.sh freebsd-build
	@./script/docker-build.sh windows-build
	@./script/docker-build.sh tar-everything
	@./script/docker-build.sh shasums
	@echo "crosscompile done..."

docker:
	SHELL=/bin/bash
	@eval $$(${SCRIPT_PATH}/bin/minikube docker-env -u) && ./script/docker-build.sh docker-build
	@if [ ! -e "bin/linux-amd64/${APPNAME}" ]; then \
		echo "Please run crosscompile before running docker command." ; \
		exit 1 ; \
	fi
	eval $$(${SCRIPT_PATH}/bin/minikube docker-env ); \
		docker build -t ${DOCKER_REPO}:${VERSION} -q --build-arg CONT_IMG_VER=${VERSION} . ; \
		docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:latest; \
		eval $$(${SCRIPT_PATH}/bin/minikube docker-env -u )


docker-run: docker
	@./script/docker-build.sh buildonly
	docker run -it --rm -v /etc/ssl/certs/:/etc/ssl/certs/ -p 8080:8080 \
	--name ${APPNAME} ${DOCKER_REPO}:latest -fakeLoad -http.addr :8080

build-in-k8s:
	@echo "building docker image for k8s cluster"
	@./script/build-local-k8s.sh
	@$(MAKE) docker

deploy-in-k8s:
	@echo "deploying in k8s"
	@./script/deploy-local-k8s.sh

run-k8s: docker-clean build-in-k8s deploy-in-k8s

tar-everything:
	@echo "tar-everything..."
	@tar -zcvf bin/${APPNAME}-linux-386-${VERSION}.tgz bin/linux-386
	@tar -zcvf bin/${APPNAME}-linux-amd64-${VERSION}.tgz bin/linux-amd64
	@tar -zcvf bin/${APPNAME}-linux-arm-${VERSION}.tgz bin/linux-arm
	@tar -zcvf bin/${APPNAME}-darwin-386-${VERSION}.tgz bin/darwin-386
	@tar -zcvf bin/${APPNAME}-darwin-amd64-${VERSION}.tgz bin/darwin-amd64
	@tar -zcvf bin/${APPNAME}-freebsd-386-${VERSION}.tgz bin/freebsd-386
	@tar -zcvf bin/${APPNAME}-freebsd-amd64-${VERSION}.tgz bin/freebsd-amd64
	@zip -9 -y -r bin/${APPNAME}-windows-386-${VERSION}.zip bin/windows-386
	@zip -9 -y -r bin/${APPNAME}-windows-amd64-${VERSION}.zip bin/windows-amd64

shasums:
	@shasum -a 256 $(BUILDFILES) > bin/${APPNAME}-${VERSION}.shasums

gpg:
	@gpg --output bin/${APPNAME}-${VERSION}.sig --detach-sig bin/${APPNAME}-${VERSION}.shasums

gpg-verify:
	@gpg --verify bin/${APPNAME}-${VERSION}.sig bin/${APPNAME}-${VERSION}.shasums


docker-build: vendor
	@echo "linux build... amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=linux/amd64" -v -o ./bin/linux-amd64/${APPNAME} ${MAIN_PATH}

linux-build: vendor
	@echo "linux build... 386"
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=linux/386" -v -o ./bin/linux-386/${APPNAME} ${MAIN_PATH} 2>/dev/null
	@echo "linux build... amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=linux/amd64" -v -o ./bin/linux-amd64/${APPNAME} ${MAIN_PATH} 2>/dev/null
	@echo "linux build... arm"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=linux/arm" -v -o ./bin/linux-arm/${APPNAME} ${MAIN_PATH} 2>/dev/null

darwin-build: vendor
	@echo "darwin build... 386"
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=darwin/386" -v -o ./bin/darwin-386/${APPNAME} ${MAIN_PATH} 2>/dev/null
	@echo "darwin build... amd64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=darwin/amd64" -v -o ./bin/darwin-amd64/${APPNAME} ${MAIN_PATH} 2>/dev/null

freebsd-build: vendor
	@echo "freebsd build... 386"
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=freebsd/386" -v -o ./bin/freebsd-386/${APPNAME} ${MAIN_PATH} 2>/dev/null
	@echo "freebsd build... amd64"
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=freebsd/amd64" -v -o ./bin/freebsd-amd64/${APPNAME} ${MAIN_PATH} 2>/dev/null

windows-build: vendor
	@echo "windows build... 386"
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=windows/386" -v -o ./bin/windows-386/${APPNAME}.exe ${MAIN_PATH} 2>/dev/null
	@echo "windows build... amd64"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -X main.Build=${VERSION} -X main.Revision=${REV} -X main.Branch=${BRANCH} -X main.OSArch=windows/amd64" -v -o ./bin/windows-amd64/${APPNAME}.exe ${MAIN_PATH} 2>/dev/null

lint:
	@echo "GO LINT..."
	@for pkg in $$(go list ./... |grep -v /vendor/) ; do \
        golint -set_exit_status $$pkg ; \
    done

test: fmt generate lint vet
	@echo "GO TEST..."
	@go test $(TEST) $(TESTARGS) -v -timeout=30s -parallel=4 -bench=. -benchmem -cover

cover:
	@echo "GO TOOL COVER..."
	@go cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	@go test $(TEST) -coverprofile=coverage.out
	@go cover -html=coverage.out
	@rm coverage.out

generate:
	@echo "GO GENERATE..."
	@go generate $$(go list ./... | grep -v /vendor/)



# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo "GO VET..."
	@go vet $(VETARGS) $$(go list ./... | grep -v /vendor/); if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	@echo "GO FMT..."
	@gofmt -w -s $(GOFMT_FILES)


docker-clean:
	@echo "cleaning docker files"
	@eval $$(${SCRIPT_PATH}/bin/minikube docker-env ); \
		docker rmi -f ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:latest; \
		eval $$(${SCRIPT_PATH}/bin/minikube docker-env -u )

destroy:
	@$(MAKE) clean
	${SCRIPT_PATH}/bin/minikube stop
	${SCRIPT_PATH}/bin/minikube delete

clean:
	@echo "cleaning files"
	rm -rf bin/
	@$(MAKE) docker-clean

.PHONY: tools default docker buildonly clean docker-clean destroy
