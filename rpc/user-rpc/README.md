rpc 服务代码生成命令
在文件目录
/Users/wangyingwen/work/golang/micro-go-zero/rpc/user-rpc
下执行一下命令
goctl rpc protoc  ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=.

后面统一写在目录
/Users/wangyingwen/work/golang/micro-go-zero 
下的 makefile 中使用make 工具统一生成