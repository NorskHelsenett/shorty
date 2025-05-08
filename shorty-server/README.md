# BUILD

1. Set version ```export SHORTY_VERSION=[version]```

2. Build executable
```bash
go mod tidy
swag init
CGO_ENABLED=0 go build -ldflags "-w -extldflags '-static' -X shorty/main.Version=$SHORTY_VERSION" -o "dist/kort" main.go
```
3. Build and upload docker image
```bash
docker build . -t ncr.sky.nhn.no/nhn/kort:$SHORTY_VERSION
docker push ncr.sky.nhn.no/nhn/kort:$SHORTY_VERSION
```
4. Cleanup
```bash
rm dist/kort
```