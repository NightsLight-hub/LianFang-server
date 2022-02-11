# LianFang-server
联坊（取名自`风起洛阳`）的服务端  
联坊是用来监控机器上docker容器状态、操作容器的软件，目标是达到与`portainer`同样的能力  
项目刚刚启动，欢迎大家参与贡献代码  

# 启动方法
配置环境变量`DOCKER_HOST=tcp://172.24.108.219:2376`然后main函数启动即可


# 当前具备的功能
* 显示机器上的容器列表
* 实时显示容器的cpu、内存
* 启停容器
* 容器日志
* 容器web shell
* 容器内部文件系统浏览

# 主要依赖的三方库

* [gin](https://github.com/gin-gonic/gin)
* [websocket](http://github.com/gorilla/websocket)
* [docker](http://github.com/docker/docker)