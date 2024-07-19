run:
	go run main.go ../..

build:
	go build

install:	build
	cp summarizefiles ~/bin/sf
	ls -l ~/bin/sf

static:
	# This static build is not successful yet. If I remove the libmagic dep, static compile might not be necessary.
	# E.g. go should statically build without needing the C dll deps
	echo sudo dnf install file-static glibc-static -y
	go build -ldflags "-linkmode 'external' -extldflags '-static'"
