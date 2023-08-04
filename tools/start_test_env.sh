#!/bin/bash

# curl -v 10.29.2.2:12345/objects/test2 -XPUT -d"this is object test2"

# 设置环境变量
export RABBITMQ_SERVER=amqp://test:test@localhost:5672
export ES_SERVER=localhost:9200
export GODOSS="/home/tsingfa/GolandProject/go-distributed-oss"

start_rabbitmq_and_es() {
  # 检查RabbitMQ是否已经启动
  if ! pgrep -x "rabbitmq-server" > /dev/null; then
    echo "Starting RabbitMQ..."
    sudo systemctl start rabbitmq-server
  else
    echo "RabbitMQ is already running."
  fi

  # 检查Elasticsearch是否已经启动
  if ! pgrep -x "elasticsearch" > /dev/null; then
    echo "Starting Elasticsearch..."
    sudo systemctl start elasticsearch
  else
    echo "Elasticsearch is already running."
  fi
}

# 定义函数 - 启动 dataServer 进程
start_data_servers() {
    # 监听地址为init_test_env.sh中分配的虚拟地址
    LISTEN_ADDRESS="10.29.1.1:12345" STORAGE_ROOT="/tmp/1" go run "$GODOSS/dataServer/dataServer.go" &
    LISTEN_ADDRESS="10.29.1.2:12345" STORAGE_ROOT="/tmp/2" go run "$GODOSS/dataServer/dataServer.go" &
    LISTEN_ADDRESS="10.29.1.3:12345" STORAGE_ROOT="/tmp/3" go run "$GODOSS/dataServer/dataServer.go" &
    LISTEN_ADDRESS="10.29.1.4:12345" STORAGE_ROOT="/tmp/4" go run "$GODOSS/dataServer/dataServer.go" &
    LISTEN_ADDRESS="10.29.1.5:12345" STORAGE_ROOT="/tmp/5" go run "$GODOSS/dataServer/dataServer.go" &
    LISTEN_ADDRESS="10.29.1.6:12345" STORAGE_ROOT="/tmp/6" go run "$GODOSS/dataServer/dataServer.go" &
}

# 定义函数 - 启动 apiServer 进程
start_api_servers() {
    LISTEN_ADDRESS="10.29.2.1:12345" go run "$GODOSS/apiServer/apiServer.go" &
    LISTEN_ADDRESS="10.29.2.2:12345" go run "$GODOSS/apiServer/apiServer.go" &
}

# 定义函数 - 停止 dataServer 和 apiServer 进程
stop_servers() {
    pkill -x "dataServer"
    pkill -x "apiServer"
}

# 判断用户传递的参数并执行相应的操作
case "$1" in
    "start")
        start_rabbitmq_and_es
        start_data_servers
        start_api_servers
        ;;
    "stop")
        stop_servers
        ;;
    *)
        echo "Usage: $0 {start|stop}"
        exit 1
        ;;
esac
