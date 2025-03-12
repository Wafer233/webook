# basic image
FROM ubuntu:20.04

#打包编译后的镜像，放到/app作为工作目录
COPY webook /app/webook
WORKDIR /app
#最佳，也可以用CMD，执行命令
ENTRYPOINT ["/app/webook"]

