@echo off
REM Windows 批处理脚本 - 生成 proto 文件

cd /d D:\work\temp\gocode\microg

echo 生成 user.proto ...
protoc --go_out=. --go-grpc_out=. ./examples/user/api/v1/user.proto

echo.
echo 生成 metadata.proto ...
protoc --go_out=. --go-grpc_out=. ./api/metadata/metadata.proto

echo.
echo 完成!
pause