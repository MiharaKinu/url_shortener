# URL Shortener

URL Shortener 是一个基于 Go 和 Gin 框架的 URL 缩短服务。该项目允许用户将长 URL 转换为短 URL，并支持短 URL 的解码和重定向功能。

## 功能

- **URL 缩短**：将长 URL 转换为短 URL。
- **URL 解码**：通过短 URL 获取原始长 URL。
- **URL 重定向**：访问短 URL 时重定向到原始长 URL。
- **过期处理**：定期清理过期的 URL 映射。

## 配置文件

项目的配置文件 `config.yaml` 示例：

```yaml
allowDomain:
  - example.com
expire: 3600
host: https://example.com
shortLength: 7
port: 8088
```

- `allowDomain`：允许缩短的域名列表。
- `expire`：URL 的过期时间（秒）。
- `host`：短 URL 的主机名。
- `shortLength`：生成的短 URL 的长度。
- `port`：服务运行的端口，默认8088
## 项目结构

- `main.go`：项目的主要代码文件。
- `config.yaml`：项目的配置文件。
- `Makefile`：项目的构建和运行脚本。

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/MiharaKinu/url_shortener.git
cd url_shortener
```

### 2. 安装依赖

```bash
make deps
```

### 3. 运行应用

```bash
make run
```

应用将运行在 `http://localhost:8088`。

### 4. 使用示例

- **缩短 URL**

  ```bash
  curl -X POST http://localhost:8088/short -H "Content-Type: application/json" -d '{"url": "https://example.com"}'
  ```

- **解码 URL**

  ```bash
  curl -X POST http://localhost:8088/decode -H "Content-Type: application/json" -d '{"url": "https://example.com/shortID"}'
  ```

- **重定向 URL**

  在浏览器中访问 `http://localhost:8088/shortID`。

## 构建项目

### 1. 编译应用

```bash
make build
```

### 2. 清理生成的文件

```bash
make clean
```

### 3. 运行测试

```bash
make test
```

## 反代说明

如果要反代程序到子目录，请按照下列伪静态规则：

```nginx
location /subdir/ {
        rewrite ^/subdir/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8088;
        proxy_redirect off;
}
```

如果Nginx运行于容器，请使用IP: `172.17.0.1:8088` 进行反代

## Ubuntu Server 发布说明

将`url_shortener`可执行文件、`config.yaml`配置文件上传到服务器

然后将以下脚本在服务器执行`sudo chmod +x ./install.sh && ./install.sh`

```bash
cat << EOF > /etc/systemd/system/url_shortener.service
[Unit]
Description=URL Shortener Service
After=network.target

[Service]
WorkingDirectory=$(pwd)
Environment=GIN_MODE=release
ExecStart=$(pwd)/url_shortener
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动 systemd 服务
systemctl enable url_shortener.service
systemctl start url_shortener.service
```

即可启动服务守护，可以通过以下命令检查运行情况

```bash
systemctl status url_shortener
```

测试可用性：

```bash
curl -X POST http://localhost:8088/short -H "Content-Type: application/json" -d '{"url": "https://example.com/"}' && echo
```

正常可以看到响应

```json
{"code":"200","data":{"url":"https://example.com/hBubVT1"}}
```

## License

项目基于 MIT License，详情请参阅 [LICENSE](./LICENSE) 文件。