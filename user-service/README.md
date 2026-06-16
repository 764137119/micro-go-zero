api 服务代码生成命令
需要在文件*.api 目录执行命令
一键清理 rm -rf ./userservice.go ./etc ./internal
cd user-service && goctl api go --api ./user-service.api -dir ./