## go-agw

一个用 Go 实现的轻量 API Gateway，支持插件、限流、上游调度、可观测性与 gRPC（h2c）。

### 功能
- 路由与反向代理：按 Path/Method 路由；简单反向代理至上游
- 插件体系：BeforeDispatch/AfterDispatch 生命周期钩子，可短路请求
- 调度：轮询（Round-Robin）在多个上游实例间分配请求
- 限流：按“客户端 IP + 路径”的令牌桶限流，支持 per-route 配置
- 可观测性：/healthz、/metrics（简单计数器）、/config 管理接口
- gRPC 支持：数据面开启 h2c；转发时处理 gRPC Header/Trailer

### 快速开始
```bash
go mod tidy
go run ./cmd/go-agw --config ./deploy/config.yaml
# 或通过环境变量
# export GO_AGW_CONFIG=./deploy/config.yaml && go run ./cmd/go-agw
```

管理接口：
- http://localhost:9000/healthz
- http://localhost:9000/metrics
- http://localhost:9000/config

数据面默认端口：:8080（可在配置中修改）

### 配置
配置文件为 YAML，主要字段：
- server: `http_addr`、`admin_addr`
- upstreams: 上游组与 target 列表
- routes: 路由规则（path、methods、upstream、rate_limit、plugins）
- plugins: 全局可用插件列表（按名称与 config 初始化）

示例（摘自 `deploy/config.yaml`）：
```yaml
server:
  http_addr: ":8080"
  admin_addr: ":9000"

upstreams:
  - name: echo
    targets: ["http://httpbin.org"]
    timeout_ms: 5000

routes:
  - path: "/"
    methods: ["GET"]
    upstream: "echo"
    rate_limit:
      rps: 100
      burst: 200
    plugins: []

plugins:
  available:
    - name: rewrite
      config:
        add_headers:
          X-AGW: go-agw
```

### 插件
- 生命周期：
  - BeforeDispatch(ctx) (handled bool, err error)：可改写请求、注入头、改写路径、选择上游；返回 handled=true 可直接向客户端返回并短路
  - AfterDispatch(ctx)：请求返回后做审计/指标/轻量改写
- 内置插件：`rewrite`
  - 支持 set_path、strip_prefix、add_prefix、add_headers、set_upstream
  - 示例：
    ```yaml
    plugins:
      available:
        - name: rewrite
          config:
            # 将 /api/foo -> /edge/foo
            strip_prefix: "/api"
            add_prefix: "/edge"
            # 将请求打上头字段
            add_headers:
              X-From: go-agw
            # 可覆盖路由里指定的上游
            set_upstream: echo
    ```

说明：当前插件链为全局链（按 `plugins.available` 顺序生效）。如需“每条路由单独的插件链”，可扩展 `RouteConfig.Plugins` 的装配逻辑。

### gRPC
- 数据面启用 h2c，可接收明文 HTTP/2（便于本地/内网场景）
- 识别 `application/grpc*` 的请求；转发时设置 `TE: trailers` 并转发 Header/Trailer
- 如需 TLS/HTTP2 强制策略，可扩展自定义 `http.Transport`

### 开发与测试
```bash
go build ./...
go test ./... -race
```

CI 位于 `.github/workflows/go.yml`：构建、vet、测试（race + 覆盖率）与覆盖率工件上传。
