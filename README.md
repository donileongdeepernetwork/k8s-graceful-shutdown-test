# K8s Graceful Shutdown Test Program

## 概述

这是一个用于测试Kubernetes (K8s) graceful shutdown机制的测试程序。该程序包含一个服务端和一个客户端，通过WebSocket进行通信。程序模拟真实的业务场景，当服务端收到graceful shutdown信号时，能够优雅地通知客户端并完成剩余的业务数据传输后再关闭连接。

## Makefile 使用指南

本项目提供了完整的Makefile来简化构建、部署和测试流程。假设您已经有一个可用的Kubernetes集群，以下是主要的使用命令：

### 构建和部署

**构建流程说明**:
- Docker镜像构建由GitHub Actions自动处理
- 推送代码到master分支时自动触发构建
- 镜像推送到: `ghcr.io/donileongdeepernetwork/k8s-graceful-shutdown-test:latest`

```bash
# 部署到Kubernetes (命名空间、服务、Deployment)
make k8s-apply

# 重启部署以拉取最新镜像
make k8s-restart

# 删除所有Kubernetes资源
make k8s-delete
```

### 监控和调试

```bash
# 查看所有Pod状态
make k8s-list-pods

# 查看服务端日志 (包含Pod名称前缀)
make k8s-logs-server

# 查看客户端日志 (包含Pod名称前缀)
make k8s-logs-client

# 删除特定Pod (先用k8s-list-pods查看Pod名称)
make k8s-delete-pod POD_NAME=<pod-name>
```

### 测试Graceful Shutdown

1. **部署应用**:
   ```bash
   make k8s-apply
   ```

2. **观察正常运行**:
   ```bash
   make k8s-logs-server  # 观察服务端连接日志
   make k8s-logs-client  # 观察客户端接收数据
   ```

3. **触发Graceful Shutdown**:
   ```bash
   # 删除一个服务端Pod，观察客户端自动重连
   make k8s-list-pods
   make k8s-delete-pod POD_NAME=<server-pod-name>
   ```

4. **验证结果**:
   - 服务端会记录客户端连接和断开
   - 客户端会自动重连到另一个服务端实例
   - 日志会显示完整的graceful shutdown流程

### 注意事项

- 所有k8s命令都在 `k8s-graceful-shutdown-test` 命名空间中执行
- 镜像会自动从 `ghcr.io/donileongdeepernetwork/k8s-graceful-shutdown-test:latest` 拉取
- 日志命令会持续监控并显示Pod名称前缀
- 删除Pod时会触发graceful shutdown，客户端会自动重连

## 架构

- **服务端 (Server)**: 监听WebSocket连接，处理客户端请求，模拟业务逻辑，并在收到shutdown信号时执行graceful shutdown流程。部署为2个Pod实例，通过Kubernetes Service进行负载均衡。
- **客户端 (Client)**: 连接到服务端，模拟用户交互，发送和接收业务数据。部署为1个Pod，通过Service连接到服务端。当连接断开时，自动重连到另一个服务端实例。
- **Kubernetes Service**: 用于暴露服务端Pod，客户端通过Service名称连接，实现负载均衡和故障转移。

## 技术栈

- **编程语言**: Go
- **WebSocket库**: gorilla/websocket
- **容器化**: Docker
- **CI/CD**: GitHub Actions
- **配置管理**: Kustomize
- **并发处理**: Go routines 和 channels
- **信号处理**: os/signal 包处理SIGTERM

## 项目结构

```
.
├── Dockerfile                    # Docker镜像构建文件
├── Makefile                      # 构建和部署管理脚本
├── cmd/
│   ├── server/
│   │   └── main.go              # 服务端主程序
│   └── client/
│       └── main.go              # 客户端主程序
├── k8s/                         # Kubernetes配置
│   ├── namespace.yaml           # 命名空间定义
│   ├── service.yaml             # Service定义
│   ├── server-deployment.yaml   # 服务端Deployment
│   ├── client-deployment.yaml   # 客户端Deployment
│   └── kustomization.yaml       # Kustomize配置
├── .github/
│   └── workflows/
│       └── build-and-push.yml   # GitHub Actions CI/CD
├── go.mod
├── go.sum
└── README.md
```

## 详细使用说明

### Kubernetes 部署

1. **部署到K8s**:
   ```bash
   make k8s-apply
   ```

2. **查看Pod状态**:
   ```bash
   make k8s-list-pods
   ```

3. **查看日志**:
   ```bash
   # 服务端日志
   make k8s-logs-server

   # 客户端日志
   make k8s-logs-client
   ```

4. **重启部署（拉取最新镜像）**:
   ```bash
   make k8s-restart
   ```

5. **删除特定Pod**:
   ```bash
   # 先列出Pod
   make k8s-list-pods

   # 删除指定Pod
   make k8s-delete-pod POD_NAME=<pod-name>
   ```

6. **清理部署**:
   ```bash
   make k8s-delete
   ```

## 交互流程

1. **连接建立**:
   - 客户端发起WebSocket连接到服务端。
   - 服务端接受连接，建立WebSocket会话，并记录连接日志。

2. **业务模拟**:
   - 客户端和服务端开始模拟业务交互，每2秒发送业务数据。
   - 业务数据包含时间戳信息。

3. **Graceful Shutdown触发**:
   - 当K8s发送SIGTERM信号到服务端Pod时（例如通过`kubectl delete pod`或滚动更新），服务端捕获该信号。
   - 服务端通过WebSocket向所有连接的客户端发送"SHUTDOWN"通知消息。

4. **剩余数据传输**:
   - 服务端生成随机数量的业务数据（1-10个），并通过WebSocket发送给客户端。
   - 客户端接收并处理这些数据。
   - 服务端等待所有数据传输完成。

5. **连接关闭**:
   - 服务端发送"CLOSE"消息给客户端。
   - 客户端和服务端优雅地关闭WebSocket连接。
   - 服务端记录断开连接日志。

6. **自动重连**:
   - 客户端检测到连接断开后，自动尝试重新连接到Service（可能连接到另一个服务端实例）。
   - 重新建立WebSocket会话，继续业务模拟。

## 故障排除

### 常见问题

1. **Pod无法启动**:
   - 检查镜像是否正确推送: `docker pull ghcr.io/donileongdeepernetwork/k8s-graceful-shutdown-test:latest`
   - 验证健康检查配置

2. **客户端无法连接**:
   - 检查Service状态: `kubectl get svc -n k8s-graceful-shutdown-test`
   - 验证端点: `kubectl get endpoints -n k8s-graceful-shutdown-test`

3. **镜像推送失败**:
   - 确保已登录GHCR: `echo $GITHUB_TOKEN | docker login ghcr.io -u donileongdeepernetwork --password-stdin`
   - 检查GitHub Actions权限

## 扩展和改进

- **监控和可观测性**: 添加Prometheus metrics和Grafana仪表板
- **配置管理**: 使用ConfigMap管理环境变量
- **安全性**: 添加TLS加密和认证
- **性能测试**: 实现负载测试和性能基准测试
- **多环境支持**: 添加staging和production环境配置
- **日志聚合**: 集成ELK stack进行集中日志管理

## 贡献

1. Fork项目
2. 创建特性分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 创建Pull Request

## 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
