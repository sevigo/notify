BASEDIR=${CURDIR}
TMP=${BASEDIR}/tmp
LOCAL_BIN:=${TMP}/bin

install-mockgen:
	GOPATH=${TMP} go get github.com/golang/mock/gomock
	GOPATH=${TMP} go install github.com/golang/mock/mockgen

mockgen: install-mockgen
	${LOCAL_BIN}/mockgen -destination=mocks/mock_gen.go -package=mocks github.com/sevigo/notify/core DirectoryWatcher

test:
	go test -timeout 10s -v -cover ./...
	CGO_ENABLED=0 go test -timeout 10s -tags fake -v