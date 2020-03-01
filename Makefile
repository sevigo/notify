BASEDIR=${CURDIR}

test:
	go test -timeout 10s -v -cover ./...