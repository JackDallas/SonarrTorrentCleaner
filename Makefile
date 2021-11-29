.POSIX:
.SUFFIXES:

SERVICE = SonarrTorrentCleaner
GO = go
RM = rm
GOFLAGS =
PREFIX = /usr/local
BUILDDIR = build

all: clean build

deps:
	cd web && npm install
	go mod download

build: deps	build/web build/app
	
build/app:
	go build -o $(BUILDDIR)/$(SERVICE)

build/web:
	mkdir build
	cd web && npm run build
	mkdir -p build/static/ && cp -r web/public/* build/static/
	cp init/SonarrTorrentCleaner.service build/

clean:
	$(RM) -rf build

