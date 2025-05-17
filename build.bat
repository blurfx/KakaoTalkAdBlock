@echo off

pushd winres

go-winres simply --icon icon.ico --arch amd64,386,arm64

popd

set GOARCH=amd64
go build -o KakaoTalkAdBlock_amd64.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go

set GOARCH=386
go build -o KakaoTalkAdBlock_i386.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go

set GOARCH=arm64
go build -o KakaoTalkAdBlock_arm64.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go
