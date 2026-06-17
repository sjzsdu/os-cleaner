# OS Cleaner 项目概览

> 最后更新于 2026-06-16

## 一句话概括

**OS Cleaner** 是一个跨平台（macOS / Linux）的系统缓存清理 CLI 工具，用 Go 编写。它可以扫描、展示、深入检查和安全清理各类系统缓存与开发工具缓存。

## 解决的问题

开发者的机器上积累了大量的系统缓存、语言包管理器缓存、IDE 缓存和浏览器缓存，这些文件通常占用几十 GB 的磁盘空间。OS Cleaner 提供了一条命令即可看清哪些可以清理、清理多少空间，并安全执行清理的能力。

## 核心功能

| 命令 | 功能 | 入口文件 |
|------|------|----------|
| `scan` | 扫描所有缓存类别，按安全等级汇总空间占用，支持`--stale`与`--older-than`按时间过滤 | `cmd/scan.go` → `internal/scanner/scanner.go` |
| `list` | 列出所有可清理的缓存类别及其描述 | `cmd/list.go` → `internal/formatter/formatter.go` |
| `clean` | 清理指定缓存，支持`--dry-run`预览、`--safe`仅清理安全类别、`--recoverable`删除前压缩备份 | `cmd/clean.go` → `internal/cleaner/cleaner.go` |
| `inspect` | 深入查看某个缓存类别的内容（支持 npm/go/pip/cargo 等类型专属分析） | `cmd/inspect.go` |
| `top` | 扫描目录下最大文件和文件夹，100MB 黄色预警、1GB 红色预警 | `cmd/top.go` → `internal/topscan/topscan.go` |
| `active` | 检测项目中使用的包（通过 `package.json`/`go.mod`/`Cargo.toml` 等识别） | `cmd/active.go` |

## 技术栈

| 技术 | 用途 | 选型理由 |
|------|------|----------|
| **Go 1.24** | 编译语言 | 单二进制分发、跨平台编译、零运行时依赖 |
| **spf13/cobra** | CLI 框架 | Go CLI 生态事实标准，提供子命令、flags、help 自动生成 |
| **Go 标准库** | 文件系统操作、并发、JSON、OS 命令调用 | 零外部依赖，`os/exec` 调用系统工具（`du`、`tar`） |
| **GitHub Actions** | CI/CD + 发布 | 多平台交叉编译（linux/darwin/windows × amd64/arm64），自动创建 Release |

## 缓存类别覆盖

代码内置了 **50+ 缓存类别**（`internal/registry/categories.go`），按领域分组：

- **macOS 系统缓存**：`~/Library/Caches`、`/Library/Caches`、日志等
- **Xcode**：DerivedData、Simulator、Archives、DeviceSupport
- **Docker**：镜像、容器、构建缓存
- **语言生态**：npm/yarn/pnpm、Go Modules、Cargo、pip/pyenv、Maven、Gradle、gem、Hex
- **包管理**：Homebrew（macOS）、apt/dnf/pacman（Linux）
- **浏览器**：Chrome、Firefox、Safari
- **IDE**：VSCode、JetBrains、Trae、Cursor、Claude 等
- **macOS 应用**：Lark/Feishu、Discord、Doubao、Google 等

每个类别都标注了安全等级（Safe / Caution / Dangerous），并映射到具体路径。

## 安全设计

- **SafetyLevel 枚举**：代码层面区分清理风险，`clean --safe` 只处理 Safe 等级
- **Dry-Run 预览**：`--dry-run` 查看会删什么但不实际执行
- **Recoverable 模式**：`--recoverable` 在删除前压缩为 `.tar.gz` 到 `~/.os-cleaner/recover/`
- **权限降级处理**：对权限不足的文件给出明确提示，建议 `sudo` 重试

## 快速上手指南

```bash
# 安装
go build -o os-cleaner . && mv os-cleaner ~/.local/bin/

# 扫描——看看能释放多少空间
os-cleaner scan

# 预览——清理 npm 缓存但不真删
os-cleaner clean npm-cache --dry-run

# 真删——清理所有安全类别的缓存
os-cleaner clean --safe

# 搜索大文件
os-cleaner top ~/Downloads --top 10

# 看看哪些包被项目真正使用
os-cleaner active ~/projects
```

## 项目结构

```
os-cleaner/
├── main.go                    # 入口：调用 cmd.Execute()
├── cmd/                       # CLI 命令层（6 个子命令）
│   ├── root.go                # 根命令、全局 flags、格式化/颜色辅助函数
│   ├── scan.go                # os-cleaner scan
│   ├── list.go                # os-cleaner list
│   ├── clean.go               # os-cleaner clean
│   ├── inspect.go             # os-cleaner inspect（含多种缓存专属分析逻辑）
│   ├── active.go              # os-cleaner active（含多语言包描述文件解析器）
│   └── top.go                 # os-cleaner top
├── internal/
│   ├── registry/              # 缓存类别注册表
│   │   ├── categories.go      # 50+ 内置缓存类别的数据定义
│   │   └── registry.go        # 类别查找、平台过滤、路径展开
│   ├── scanner/               # 扫描引擎
│   │   └── scanner.go         # 并发扫描、大小计算、时间过滤、表格/JSON 输出
│   ├── cleaner/               # 清理引擎
│   │   └── cleaner.go         # 安全删除、dry-run、recoverable 压缩、权限错误处理
│   ├── topscan/               # 大文件扫描引擎
│   │   └── topscan.go         # 并发递归统计、排序、颜色阈值输出
│   ├── formatter/             # 列表格式化
│   │   └── formatter.go       # 缓存类别列表的表格/JSON 输出
│   └── utils/                 # 通用工具
│       ├── size.go            # 大小格式化、路径工具、终端颜色、目录遍历
│       └── progress.go        # 终端旋转进度条
├── setup.sh                   # 安装脚本
└── .github/workflows/         # CI/CD：多平台交叉编译 + GitHub Release
```
