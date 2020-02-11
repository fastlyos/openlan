.PHONY: linux rpm win-zip test

linux:
	go build -mod=vendor -o resource/point.linux.x86_64 main/point_linux.go
	go build -mod=vendor -o resource/vswitch.linux.x86_64 main/vswitch.go

windows:
	go build -mod=vendor -o resource/point.windows.x86_64.exe main/point_windows.go
	go build -mod=vendor -o resource/vswitch.windows.x86_64.exe main/vswitch.go

osx: darwin

darwin:
	go build -mod=vendor -o resource/point.darwin.x86_64 main/point_darwin.go
	go build -mod=vendor -o resource/vswitch.darwin.x86_64 main/vswitch.go

rpm:
	./packaging/auto.sh
	rpmbuild -ba packaging/openlan-point.spec
	rpmbuild -ba packaging/openlan-vswitch.spec
	cp -rvf ~/rpmbuild/RPMS/x86_64/openlan-*.rpm resource

WIN_DIR = "openlan-windows-"$$(cat VERSION)

win-zip:
	rm -rf $(WIN_DIR) && mkdir -p $(WIN_DIR)
	cp -rvf resource/point.json $(WIN_DIR)
	cp -rvf resource/point.windows.x86_64.exe $(WIN_DIR)
	cp -rvf resource/vswitch.windows.x86_64.exe $(WIN_DIR)
	rm -rf resource/$(WIN_DIR).zip
	zip -r resource/$(WIN_DIR).zip $(WIN_DIR)

docker:
	docker build -t openlan-point -f packaging/point/Dockerfile .
	docker build -t openlan-vswitch -f packaging/vswitch/Dockerfile .
	# --env VS_ADDR=192.168.209.141 --env VS_AUTH=hi@admin:hi123$ --env VS_TLS=true
	# docker run -d --privileged openlan-point:latest
	# docker run -d  -p 10000:10000 -p 10002:10002 openlan-vswitch:latest

test:
	go test -mod=vendor -bench=. github.com/danieldin95/openlan-go/point
	go test -mod=vendor -bench=. github.com/danieldin95/openlan-go/libol
