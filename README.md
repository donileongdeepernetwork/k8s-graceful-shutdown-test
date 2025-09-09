# K8s Graceful Shutdown Test Program

## 概述

这是一个用于测试Kubernetes (K8s) graceful shutdown机制的测试程序。该程序包含一个服务端和一个客户端，通过WebSocket进行通信。程序模拟真实的业务场景，当服务端收到graceful shutdown信号时，能够优雅地通知客户端并完成剩余的业务数据传输后再关闭连接。

## 架构

- **服务端 (Server)**: 监听WebSocket连接，处理客户端请求，模拟业务逻辑，并在收到shutdown信号时执行graceful shutdown流程。部署为2个Pod实例，通过Kubernetes Service进行负载均衡。
- **客户端 (Client)**: 连接到服务端，模拟用户交互，发送和接收业务数据。部署为1个Pod，通过Service连接到服务端。当连接断开时，自动重连到另一个服务端实例。
- **Kubernetes Service**: 用于暴露服务端Pod，客户端通过Service名称连接，实现负载均衡和故障转移。

## 交互流程

1. **连接建立**:
   - 客户端发起WebSocket连接到服务端。
   - 服务端接受连接，建立WebSocket会话。

2. **业务模拟**:
   - 客户端和服务端开始模拟业务交互，例如发送和接收消息。
   - 业务数据可以是随机的字符串或JSON对象，模拟真实的API调用或数据传输。

3. **Graceful Shutdown触发**:
   - 当K8s发送SIGTERM信号到服务端Pod时（例如通过`kubectl delete pod`或滚动更新），服务端捕获该信号。
   - 服务端通过WebSocket向所有连接的客户端发送通知消息，告知客户端服务端即将关闭。

4. **剩余数据传输**:
   - 服务端生成随机数量的业务数据（例如1-10个），并通过WebSocket发送给客户端。
   - 客户端接收并处理这些数据。
   - 服务端等待所有数据传输完成。

5. **连接关闭**:
   - 服务端发送关闭消息给客户端。
   - 客户端和服务端优雅地关闭WebSocket连接。
   - 服务端进程退出。

6. **自动重连**:
   - 客户端检测到连接断开后，自动尝试重新连接到Service（可能连接到另一个服务端实例）。
   - 重新建立WebSocket会话，继续业务模拟。

## 技术栈

- **编程语言**: Go
- **WebSocket库**: gorilla/websocket
- **并发处理**: Go routines 和 channels
- **信号处理**: os/signal 包处理SIGTERM

## 部署和测试

1. **本地测试**:
   - 运行服务端: `go run cmd/server/main.go` (可设置环境变量 `PORT` 修改端口，默认8080)
   - 运行客户端: `go run cmd/client/main.go` (可设置环境变量 `SERVER_URL` 修改服务端地址，默认 `ws://localhost:8080/ws`)

2. **K8s部署**:
   - **服务端Deployment**: 创建一个Deployment，设置replicas为2，运行服务端容器。示例YAML:
     ```yaml
     apiVersion: apps/v1
     kind: Deployment
     metadata:
       name: ws-server
     spec:
       replicas: 2
       selector:
         matchLabels:
           app: ws-server
       template:
         metadata:
           labels:
             app: ws-server
         spec:
           containers:
           - name: server
             image: your-server-image
             ports:
             - containerPort: 8080
     ```
   - **Service**: 创建Service来暴露服务端Pod，实现负载均衡。示例YAML:
     ```yaml
     apiVersion: v1
     kind: Service
     metadata:
       name: ws-server-service
     spec:
       selector:
         app: ws-server
       ports:
       - protocol: TCP
         port: 8080
         targetPort: 8080
       type: ClusterIP
     ```
   - **客户端Deployment**: 创建一个Deployment，设置replicas为1，运行客户端容器。客户端通过Service名称（如ws-server-service:8080）连接。示例YAML:
     ```yaml
     apiVersion: apps/v1
     kind: Deployment
     metadata:
       name: ws-client
     spec:
       replicas: 1
       selector:
         matchLabels:
           app: ws-client
       template:
         metadata:
           labels:
             app: ws-client
         spec:
           containers:
           - name: client
             image: your-client-image
             env:
             - name: SERVER_URL
               value: "ws://ws-server-service:8080/ws"
     ```
   - 通过`kubectl delete pod`触发graceful shutdown，观察客户端是否自动重连到另一个实例。

3. **观察行为**:
   - 检查Pod日志，确认graceful shutdown流程和重连行为。
   - 验证客户端是否收到通知、处理剩余数据，并成功重连。
   - 使用`kubectl logs`监控服务端和客户端的日志。

## 预期效果

通过这个测试程序，可以验证：
- 服务端是否正确捕获SIGTERM信号。
- WebSocket连接是否在shutdown期间保持稳定。
- 客户端是否能收到shutdown通知并处理剩余数据。
- 服务端是否在完成所有任务后退出，避免数据丢失。
- 客户端是否能自动重连到另一个服务端实例，实现故障转移。
- Service是否正确进行负载均衡和故障转移。

## 扩展

- 添加更多业务逻辑模拟。
- 支持多个客户端并发连接。
- 集成健康检查端点。
- 添加metrics收集，监控shutdown时间等。
- 实现更复杂的负载均衡策略或会话亲和性（session affinity）。
