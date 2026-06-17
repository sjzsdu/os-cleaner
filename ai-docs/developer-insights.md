# 开发者视角洞察

> 最后更新于 2026-06-16

本文档分析 OS Cleaner 代码中体现的工程取舍、设计约束和演进线索，帮助后续开发者理解"为什么这样做"。

---

## 1. 零外部依赖（除了 Cobra）

项目在 `go.mod` 中只有一行直接依赖：`github.com/spf13/cobra`。

**为什么这样做？**

- 项目核心能力（文件系统操作、大小计算、命令执行）全部由 Go 标准库提供
- 减少版本管理负担和潜在的依赖冲突
- 构建产物体积小——含 Cobra 的二进制约 5-7MB
- 没有引入 `du` 的 Go 实现（如 `diskusage`），而是直接调用系统 `du` 命令

**代价**：
- JSON 格式化用标准库 `encoding/json`，功能够用但性能不如 `jsoniter`
- 颜色输出用 ANSI 转义码硬编码（`utils/size.go:30-52`），不支持 Windows 原生终端
- `active` 命令中的包描述文件解析器是手写的简易版（`cmd/active.go:166-353`），没有用真实 parser

## 2. 数据驱动而非代码驱动

**缓存类别是纯数据**（`internal/registry/categories.go`），而不是一个个子类或接口实现。

```go
// categories.go 的风格——声明式数据
var categories = []CacheCategory{
    {
        ID: "go-modules",
        Name: "Go Module Cache",
        Platforms: []string{"macos", "linux"},
        SafetyLevel: Safe,
        Paths: []PathRule{{Path: "~/go/pkg/mod"}},
        CleanCmd: "go clean -modcache",
    },
    // ... 50+ 条目
}
```

**好处**：
- 新增缓存类别只需追加数据，不涉及架构变化
- 适合社区贡献——任何人都能看懂并添加自己熟悉的缓存路径
- 运行时通过平台过滤自动适配，无需 `#ifdef` 或构建标签

**局限性**：
- 路径模式不支持通配符或正则匹配（`PathRule` 的 `Pattern` 字段定义了但未在扫描逻辑中使用）
- 缺少动态类别注册机制——想从外部配置文件添加类别需要额外开发

## 3. 系统命令调用 vs Go 原生实现

项目混合使用 Go 标准库和系统命令：

| 场景 | 实现方式 | 原因 | 代码位置 |
|------|----------|------|----------|
| 目录大小（扫描） | `du -sk` | 性能：C 实现遍历目录远快于 Go | `scanner/scanner.go:149-173` |
| 目录大小（清理前） | Go `filepath.Walk` | 需要精确值，且目录已被扫描过 | `utils/size.go:91-105` |
| 目录大小（top） | Go 递归 + `atomic.Int64` | 需要同时统计文件数，并发安全 | `topscan/topscan.go:130-157` |
| 可恢复删除 | `tar -czf` | 避免实现 tar 压缩逻辑 | `utils/size.go:108-111` |
| npm 缓存 | `npm cache ls` | 利用 npm CLI 的精确输出 | `cmd/inspect.go:215-249` |

**这个取舍要记住**：`du -sk` 极快但只返回 KB 大小和整数，所有 `scan` 命令中显示的文件数都是估算值（`size / 4096`，见 `scanner.go:166`）。当文件平均大小偏离 4KB 越多，估算偏差越大。`code-context` 的 `explain` 功能可以帮助定位这类隐含假设。

## 4. CLI 命令层的编排风格

每个子命令文件遵循统一模式：

1. **Flags 定义**（包级变量）
2. **Cobra Command 定义**（Use/Short/Long/RunE）
3. **init() 注册**（`rootCmd.AddCommand`）

其中 `RunE` 只做三件事：
1. 构造 Options 结构体
2. 调用内部包的业务函数
3. 返回 error

```go
// 典型模式——scan.go
RunE: func(cmd *cobra.Command, args []string) error {
    opts := scanner.ScanOptions{
        Parallel:  true,
        Verbose:   verbose,
        JSON:      jsonOutput,
        ShowStale: scanStale,
        OlderThan: parseDuration(scanOlderThan),
    }
    return scanner.Scan(opts)
}
```

**好处**：业务逻辑可测试（`scanner.Scan` 可以单独测试），CLI 层只做编排。

**注意**：当前版本没有 `*_test.go` 文件。这是未来可以改进的方向。

## 5. 对并发安全的处理方式演变

项目在不同地方用了不同的并发安全策略：

```go
// 方案1: channel 收集结果（scanner.go）
resultsChan := make(chan ScanResult, len(categories))
// goroutine 写入 chan → 主 goroutine 用 range 读取

// 方案2: atomic.Int64（topscan.go）
var size atomic.Int64
var count atomic.Int64
// 无需锁，适合简单计数器

// 方案3: sync.Mutex（progress.go）
type Progress struct {
    mu sync.Mutex
    completed int
    // ...
}
func (p *Progress) Increment(current string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.completed++
}
```

三种方案的选用原则：
- **channel**：需要收集一组结构体结果时用
- **atomic**：只需要累加一个整数时用，无锁开销最小
- **Mutex**：需要保护多个相关字段的一致性时用

## 6. 路径展开的重复实现

项目中存在两份路径展开逻辑：

- `registry/registry.go:121-144` — `ExpandPath()`
- `utils/size.go:61-76` — `ExpandPath()`（功能完全相同）

这是代码重复。两个函数都处理：
- `~` → `$HOME`
- `~/...` → `$HOME/...`
- `$ENV_VAR` → 展开

[推测] 这个重复可能是因为 `utils` 包最初是纯工具，后来 `registry` 引入了自己的展开逻辑以避免循环依赖。重构时可以考虑将展开逻辑统一移动到 `utils`，或让 `registry` 调用 `utils.ExpandPath`。

## 7. `inspect` 命令中的类型调度

`cmd/inspect.go:147-168` 用了一组 `is*()` 函数做类型分派：

```go
switch {
case isNPMCache(cat):
    inspectNPMCache(expandedPath, top)
case isGoModules(cat):
    inspectGoModules(expandedPath, top)
// ... 10+ 类型
default:
    inspectGenericCache(expandedPath, top)
}
```

这是典型的"基于标签的分派"模式。优点是可以针对不同缓存类型提供量身定制的查看体验（npm 看包名、Xcode 看项目维度、Homebrew 看 bottle 列表）。

**潜在改进**：如果未来类型数量增长，可以引入类型注册机制代替 `switch` 硬编码。

## 8. 包描述文件解析器的实现权衡

`cmd/active.go` 中的解析器**没有使用任何第三方解析库**，而是手写字符串处理：

```go
// 解析 package.json
func parsePackageJson(dir string) []string {
    content := string(data)
    // 不用 json.Unmarshal，而是逐行查找 key:value 模式
    for _, line := range lines {
        line = trimSpace(line)
        if contains(line, ":") && !contains(line, "//") {
            parts := splitOn(line, ":")
            // 提取包名...
        }
    }
}
```

**这有意为之**：
- 避免为每个语言引入解析器依赖（json、toml、xml、elixir...）
- 一次性解析多个格式时，这种简化方式"足够好"
- `go.mod` 直接用空格分割（`splitOn(line, " ")`），不处理复杂语法

**已知局限**：
- `package.json` 的解析可能被注释、特殊格式干扰
- `pom.xml` 直接返回 `["Maven project"]` 占位符，未实际解析 XML
- `Gemfile` 和 `mix.exs` 的解析非常粗糙

## 9. 终端输出的 ANSI 风格

项目使用硬编码的 ANSI 转义码实现颜色和样式（`internal/utils/size.go:30-52`）：

```go
func Bold(s string) string  { return "\033[1m" + s + "\033[0m" }
func Green(s string) string { return "\033[32m" + s + "\033[0m" }
func Dim(s string) string   { return "\033[2m" + s + "\033[0m" }
```

**为什么不引入 lipgloss 或 termenv ？**

- 保持零外部依赖
- 对于这个项目来说，ANSI 码级别够用了
- 不影响功能，纯视觉增强

**代价**：在非 ANSI 终端（如 Windows 旧版 cmd）或重定向到文件时，转义码会显示为乱码。但 `--json` 模式可以规避这个问题。

## 10. 演进线索与扩展点

### 来自 GitHub Actions 的线索

`release.yml` 构建 6 种平台的二进制（linux/darwin/windows × amd64/arm64），说明项目预期有跨平台用户。但 `categories.go` 中 Linux 的类别只有 apt/dnf/pacman，macOS 则覆盖了 30+ 类别。**Windows 类别当前为空**。

### 代码中预留但未使用的字段

- `PathRule.Pattern` — 定义了但扫描逻辑中未使用，可能是为 glob 过滤预留
- `CacheCategory.Dangerous` — SafetyLevel 定义了 `Dangerous`，但没有类别使用它
- `ScanOptions.Parallel` — 总是 true，未来可以支持顺序扫描（调试用）

### 适合贡献的方向

1. **添加 Windows 缓存类别**——这是最明显的缺口
2. **单元测试**——当前没有 `*_test.go`
3. **GitHub Issue 模板**——`CONTRIBUTING.md`
4. **动态类别配置**——支持从 YAML/JSON 文件加载自定义缓存路径
5. **清理结果缓存**——避免重复扫描（当前每次 scan 都重新 exec `du`）
