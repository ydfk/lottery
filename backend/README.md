# 彩票智能推荐系统

这是一个基于Go语言开发的彩票智能推荐系统后端，使用AI技术来为用户提供彩票号码推荐。

## 项目特点

- 支持多种彩票类型（当前包括双色球和大乐透）
- 基于AI模型的号码推荐
- JWT认证和权限控制
- RESTful API接口
- 使用SQLite本地数据库
- 支持定时任务调度

## 技术栈

- Go 1.21+
- Fiber (Web框架)
- GORM (ORM库)
- SQLite (数据库)
- JWT (认证)
- OneAPI (AI接口)

## 项目结构

```
backend/
├── config/             # 配置文件目录
│   └── config.yaml     # 主配置文件
├── data/               # 数据文件目录
│   └── lottery.db      # SQLite数据库文件
├── internal/           # 内部包
│   ├── handlers/       # HTTP处理器
│   ├── models/         # 数据模型
│   └── pkg/            # 内部工具包
│       ├── ai/         # AI客户端
│       ├── config/     # 配置处理
│       ├── database/   # 数据库处理
│       └── scheduler/  # 任务调度器
├── main.go             # 程序入口
└── README.md           # 项目说明文件
```

## 配置说明

项目使用`config/config.yaml`文件进行配置：

```yaml
oneapi:
  base_url: "https://api.oneapi.com/v1"  # OneAPI服务地址
  api_key: ""                           # 你的API密钥
  allowed_models:                       # 允许使用的模型
    - "gpt-4"
    - "gpt-3.5-turbo"
    - "claude-2.1"
  timeout: 30s                          # 请求超时时间
  max_retries: 3                        # 最大重试次数

database:
  path: "./data/lottery.db"             # 数据库文件路径

server:
  port: 3000                            # 服务端口

jwt:
  secret: "your-jwt-secret-key"         # JWT密钥
  expiration: 86400                     # Token过期时间(秒)

# 初始用户配置
users:
  - username: "admin"                   # 管理员用户
    password: "admin123"
  - username: "user"                    # 普通用户
    password: "user123"

# 彩票类型配置
lottery_types:
  - name: "双色球"
    schedule_cron: "0 0 20 * * 2,4,7"   # 每周二、四、日20:00
    model_name: "gpt-4"
    is_active: true
  - name: "大乐透"
    schedule_cron: "0 0 20 * * 1,3,6"   # 每周一、三、六20:00
    model_name: "gpt-4"
    is_active: true
```

## 如何编译

### 前提条件

- 安装Go 1.21或更高版本
- 设置好GOPATH和GOROOT环境变量

### 编译步骤

1. 克隆仓库到本地

```bash
git clone https://your-repo/lottery.git
cd lottery/backend
```

2. 安装依赖

```bash
go mod tidy
```

3. 编译程序

```bash
go build -o lottery-backend main.go
```

## 如何运行

### 直接运行

```bash
go run main.go
```

### 运行编译后的程序

```bash
./lottery-backend    # Linux/Mac
lottery-backend.exe  # Windows
```

## 使用VS Code调试

要使用VS Code调试此项目，请按照以下步骤操作：

1. 确保已经安装了VS Code和Go扩展（ms-vscode.go）。

2. 在VS Code中打开项目文件夹（backend目录）。
   - **重要**：请确保在打开VS Code时，直接打开`lottery/backend`目录，而不是整个`lottery`仓库目录。这是因为Go模块系统需要在包含`go.mod`文件的目录中运行。

3. 创建`.vscode`目录和`launch.json`文件，内容如下：

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}",
            "env": {},
            "args": []
        }
    ]
}
```

4. 在代码中设置断点（点击行号左侧）。

5. 按F5键或点击调试视图中的绿色播放按钮开始调试。

### 调试故障排除

如果遇到以下错误：
```
Build Error: go build -o ... -gcflags all=-N -l .
go: cannot find main module, but found .git/config in ...
to create a module there, run:
go mod init (exit status 1)
```

这意味着调试器无法找到Go模块。请尝试以下解决方法：

1. 确保你是在`backend`目录下（包含`go.mod`文件的目录）进行调试。

2. 如果在根目录（lottery）下启动VS Code，请在VS Code设置中修改Go语言的工作区位置：
   - 打开VS Code设置 (Ctrl+,)
   - 搜索 "Go: Work Folder"
   - 将值设置为 "./backend"

3. 或者在`launch.json`中指定工作目录：
```json
{
    "name": "Launch Package",
    "type": "go",
    "request": "launch",
    "mode": "auto",
    "program": "${workspaceFolder}/backend",
    "cwd": "${workspaceFolder}/backend",
    "env": {},
    "args": []
}
```

4. 确保所有Go工具已正确安装。可以使用命令面板 (Ctrl+Shift+P) 运行 "Go: Install/Update Tools" 命令。

### 调试中的实用功能

6. 调试过程中可以：
   - 使用调试控制台查看变量值
   - 单步执行（F10）或步入函数（F11）
   - 在变量上悬停鼠标查看其值
   - 在调试控制台中执行表达式

7. 如果需要监视特定变量，可以使用调试面板中的"监视"部分添加变量。

8. 要在Delve控制台中运行命令，可以在调试控制台中使用`dlv`命令。

### 远程调试

如果需要远程调试，可以添加以下配置到`launch.json`：

```json
{
    "name": "Remote Debug",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "remotePath": "${workspaceFolder}",
    "port": 2345,
    "host": "127.0.0.1"
}
```

然后在远程机器上使用Delve启动应用：

```bash
dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient
```

## API接口

应用启动后，API服务会运行在配置文件中指定的端口（默认3000）。

### 主要接口

- `POST /api/auth/login` - 用户登录
- `GET /api/lottery-types` - 获取彩票类型列表
- `GET /api/recommendations` - 获取推荐列表
- `PUT /api/recommendations/:id/purchase` - 更新购买状态
- `GET /api/audit-logs` - 获取审计日志

## 开发说明

- 项目使用Go Modules管理依赖
- 数据库会在首次运行时自动初始化并创建必要的表
- 初始用户和彩票类型会在启动时自动创建

## 注意事项

- 默认配置中的密钥和密码仅用于开发环境，生产环境请务必更改
- OneAPI需要有效的API密钥才能使用AI功能
- 系统会在启动时自动创建配置文件中定义的用户和彩票类型