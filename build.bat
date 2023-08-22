@echo off

pushd winres

go-winres simply -icon=icon.ico

popd

go build -o KakaoTalkAdBlock.exe -ldflags "-H windowsgui -s -w" .\cmd\main.go
