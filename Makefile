VERSION = 0.2.0

build-dev:
	@echo "Building gocho"
	go install -i github.com/temorfeouz/gocho/cmd/gocho

clean:
	rm -rf dist/*

# dist: clean ui generate
# 	@echo "Building gocho for Linux x86_64..."
# 	mkdir -p dist/linux64
# 	go build -o dist/linux64/gocho cmd/gocho/gocho.go
# 	@zip -j dist/gocho_${VERSION}_linux64.zip dist/linux64/gocho

dist-all:  dist-linux32 dist-linux64 dist-mips dist-mipsle dist-win32 dist-win64 dist-darwin dist-mips64 dist-mips64le dist-arm dist-arm64
generate:
	go generate cmd/gocho/gocho.go

dist-linux32:
	@echo "Building gocho for Linux 32bits..."
	mkdir -p dist/linux386
	GOOS=linux GOARCH=386 go build -o dist/linux386/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_linux386.zip dist/linux386/gocho


dist-linux64:
	@echo "Building gocho for Linux 64bits..."
	mkdir -p dist/linux64
	GOOS=linux GOARCH=amd64 go build -o dist/linux64/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_linux64.zip dist/linux64/gocho

dist-win32:
	@echo "Building gocho for Windows 32bits..."
	mkdir -p dist/win32
	GOOS=windows GOARCH=386 go build -o dist/win32/gocho.exe cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_win32.zip dist/win32/gocho.exe

dist-win64:
	@echo "Building gocho for Windows 64bits..."
	mkdir -p dist/win64
	GOOS=windows GOARCH=amd64 go build -o dist/win64/gocho.exe cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_win64.zip dist/win64/gocho.exe

dist-darwin:
	@echo "Building gocho for Darwin 64bits..."
	mkdir -p dist/darwin
	GOOS=darwin GOARCH=amd64 go build -o dist/darwin/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_darwin.zip dist/darwin/gocho

dist-mips:
	@echo "Building gocho for mips 32bits..."
	mkdir -p dist/mips
	GOOS=linux GOARCH=mips go build -o dist/mips/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_mips.zip dist/mips/gocho

dist-mipsle:
	@echo "Building gocho for mipsle 32bits..."
	mkdir -p dist/mipsle
	GOOS=linux GOARCH=mipsle go build -o dist/mipsle/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_mipsle.zip dist/mipsle/gocho

dist-mips64:
	@echo "Building gocho for mips64 64bits..."
	mkdir -p dist/mips64
	GOOS=linux GOARCH=mips64 go build -o dist/mips64/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_mips64.zip dist/mips64/gocho

dist-mips64le:
	@echo "Building gocho for mips64le 64bits..."
	mkdir -p dist/mips64le
	GOOS=linux GOARCH=mips64le go build -o dist/mips64le/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_mips64le.zip dist/mips64le/gocho

dist-arm:
	@echo "Building gocho for arm 32bits..."
	mkdir -p dist/arm
	GOOS=linux GOARCH=arm go build -o dist/arm/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_arm.zip dist/arm/gocho

dist-arm64:
	@echo "Building gocho for arm 32bits..."
	mkdir -p dist/arm64
	GOOS=linux GOARCH=arm64 go build -o dist/arm64/gocho cmd/gocho/gocho.go
	@zip -j dist/gocho_${VERSION}_arm64.zip dist/arm64/gocho


docker: dist
	docker build . -t temorfeouz/gocho

start:
	docker run -it -p "1337:1337" --rm temorfeouz/gocho gocho start --debug || true

test:
	docker run -it --rm temorfeouz/gocho || true

clean-dashboard:
	rm -rf assets/assets_gen.go

ui: clean-dashboard
	cd ui \
	&& yarn build
