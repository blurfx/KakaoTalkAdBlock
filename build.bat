@echo off

pushd winres

go-winres simply -icon=icon.ico

popd

set GOARCH=amd64
go build -o KakaoTalkAdBlock_amd64.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go

set GOARCH=386
go build -o KakaoTalkAdBlock_i386.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go
