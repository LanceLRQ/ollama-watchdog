# Ollama Watchdog

## 项目概述

Ollama Watchdog 是一个基于 Vue 3 和 Go 语言开发的项目，主要用于监控和管理 GPU 资源。

项目前端使用 Vue 3 和 ECharts 进行数据可视化，后端使用 Go 语言进行数据处理和监控。

## 运行演示

![运行演示](./docs/demo.jpg)

## 安装

### curl
```bash
sudo curl -sSL https://raw.githubusercontent.com/LanceLRQ/ollama-watchdog/main/install.sh | sh
```

### wget
```bash
sudo wget -qO- https://raw.githubusercontent.com/LanceLRQ/ollama-watchdog/main/install.sh | sh
```

## 安装与运行(开发)

### 前端

1. 克隆项目到本地：
   ```bash
   git clone https://github.com/LanceLRQ/ollama-watchdog.git
   cd ollama-watchdog/web
   ```

2. 安装依赖：
   ```bash
   npm install
   ```

3. 启动开发服务器：
   ```bash
   npm run dev
   ```

4. 构建生产环境代码：
   ```bash
   npm run build
   ```

### 后端

1. 进入后端目录：
   ```bash
   cd src
   ```

2. 运行 Go 服务：
   ```bash
   go run -tags dev .
   # 或者
   sh ./dev.sh
   ```

3. 编译二进制：
   ```bash
   sh ./build.sh
   ```

## 编译环境

`go >= 1.23`


## 配置

项目默认配置文件位于：`~/.config/ollama-watchdog/server.yaml`

---

### 配置文件说明
配置文件采用 **YAML** 格式，支持动态修改（通过 `ollama-watchdog config set` 命令）。

---

#### `listen`
- **类型**: `string`
- **默认值**: `"0.0.0.0:23333"`
- **说明**: 看门狗主服务监听地址（IP + 端口）。
- **配置命令**:
  ```bash
  ollama-watchdog config set listen "0.0.0.0:23333"
  ```

---

#### `ollama_listens`
- **类型**: `string[]` (数组)
- **默认值**: `["http://127.0.0.1:11434"]`
- **说明**: 多 Ollama 服务地址（负载均衡或集群场景），用逗号分隔多个地址。
- **配置命令**:
  ```bash
  ollama-watchdog config set ollama_listens "http://127.0.0.1:11434,http://127.0.0.1:11435"
  ```

---

#### `ollama_services`
- **类型**: `string[]` (数组)
- **默认值**: `["ollama"]`
- **说明**: 需要监控的 Ollama 服务名称（用于进程管理），用逗号分隔。注意顺序需要和ollama_listens的一致，才能保证能够正确的在监控页面上重启ollama。
- **配置命令**:
  ```bash
  ollama-watchdog config set ollama_services "ollama-gpu0,ollama-gpu-1"
  ```

---

#### `nvidia_smi_path`
- **类型**: `string`
- **默认值**: `"/usr/bin/nvidia-smi"`
- **说明**: `nvidia-smi` 可执行文件路径（用于 GPU 监控）。
- **配置命令**:
  ```bash
  ollama-watchdog config set nvidia_smi_path "/usr/bin/nvidia-smi"
  ```

---

#### `gpu_sample_db`
- **类型**: `string`
- **默认值**: `"~/.config/ollama-watchdog/.gpu_samples"`
- **说明**: GPU 采样数据存储路径（SQLite 数据库）。
- **配置命令**:
  ```bash
  ollama-watchdog config set gpu_sample_db "~/.config/ollama-watchdog/.gpu_samples"
  ```

---

#### **注意事项**
1. **数组类型**：配置时用英文逗号分隔值（如 `"a,b,c"`）。
2. **动态生效**：配置完成后需要执行 `systemctl restart ollama-watchdog` 以使配置生效。
3. **默认值**：未显式配置时使用结构体中的默认值。


## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 许可证

本项目采用 MIT 许可证。
