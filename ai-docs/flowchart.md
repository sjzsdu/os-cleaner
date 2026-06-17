# 核心流程说明

> 最后更新于 2026-06-16

本文档用流程图说明 OS Cleaner 的 3 个核心工作流，以及类别匹配的解析流程。

---

## 1. 扫描流程（scan）

这是最常用的命令，展示所有缓存类别占用的磁盘空间。

```mermaid
flowchart TD
    START(["os-cleaner scan [--stale] [--older-than 90d] [--json]"])
    START --> ParseOpts["cmd/scan.go\n构建 ScanOptions"]
    ParseOpts --> GetCats["registry.GetCategoriesByPlatform()\n获取当前平台的所有类别"]

    GetCats --> InitProgress["初始化 Progress 进度条"]
    InitProgress --> LaunchGoroutines["为每个类别启动 goroutine\n（并发扫描）"]

    subgraph PerCategory["每个 goroutine: scanCategory()"]
        direction TB
        LoopPaths["遍历 cat.Paths"] --> ExpandPath["registry.ExpandPath()\n展开 ~ 和 $HOME"]
        ExpandPath --> StatPath["os.Stat(expandedPath)"]
        StatPath --> PathExists{"路径存在?"}
        PathExists -->|否| Skip["跳过"]
        PathExists -->|是| IsDir{"是目录?"}
        IsDir -->|是| FastSize["fastGetSize()\n调用 du -sk\n获取大小和文件数"]
        IsDir -->|否| SingleFile["info.Size()\n单个文件直接取大小"]
        FastSize --> CheckStale{"--stale 或\n--older-than?"}
        SingleFile --> CheckStale
        CheckStale -->|是| WalkTime["filepath.Walk\n遍历计算最早 ModTime"]
        CheckStale -->|否| Done["返回 ScanResult"]
        WalkTime --> Done
    end

    LaunchGoroutines --> PerCategory
    PerCategory --> Collect["wg.Wait()\n从 chan 收集所有结果"]
    Collect --> Sort["按大小降序排序"]
    Sort --> FilterStale{"过滤条件:"}

    FilterStale -->|--stale| StaleFilter["只保留 LastAccess\n> 30 天前的"]
    FilterStale -->|--older-than| AgeFilter["只保留超过\n指定时长的"]
    FilterStale -->|无过滤| NoFilter["保留全部"]

    StaleFilter --> Output
    AgeFilter --> Output
    NoFilter --> Output

    Output{"--json?"}
    Output -->|是| JSON["json.MarshalIndent\n输出到 stdout"]
    Output -->|否| Table["终端表格输出"]
    Table --> Summary["汇总: Total / Safe / Caution"]
    Summary --> Tip["提示下一步\nclean 或 clean --safe"]

    JSON --> END(["结束"])
    Tip --> END
```

### 扫描的关键细节

**大小计算策略**（`internal/scanner/scanner.go:149-173`）：

```
fastGetSize():
  ┌─ macOS/Linux: du -sk <path>
  │  输出: "12345    /path/to/dir"
  │  解析: 12345 KB → 乘以 1024 → bytes
  │  文件数: 估算值 size / 4096（平均文件大小 4KB）
  │
  └─ 为什么用 du 而非 filepath.Walk？
     - du 是 C 实现的，遍历目录远快于 Go 的 filepath.Walk
     - 对于包含数十万文件的大缓存目录，差距可达 10x-100x
```

**时间过滤机制**（`internal/scanner/scanner.go:175-201`）：

```
calculateDirSizeAndTime():
  ┌─ 遍历目录所有文件，记录：
  │   - 最新 ModTime（latestAccess）
  │   - 最早 ModTime（oldestAccess）
  │
  └─ 这只在 --stale 或 --older-than 时执行
     普通 scan 不会触发遍历，以保持速度
```

---

## 2. 清理流程（clean）

```mermaid
flowchart TD
    START(["os-cleaner clean [category]\n[--safe] [--dry-run] [--recoverable]"])
    START --> ParseMode{"选择模式:"}

    ParseMode -->|--safe| Safe["GetSafeCategories()"]
    ParseMode -->|--caution| Caution["GetCautionCategories()"]
    ParseMode -->|指定类别| ByID["resolveCategory(args[0])\n匹配 ID/名称"]
    ByID --> Found{"找到?"}
    Found -->|否| ShowHelp["列出可用类别"]
    Found -->|是| Single[["单个类别"]]

    Safe --> CleanList["待清理列表"]
    Caution --> CleanList
    Single --> CleanList

    CleanList --> Loop["遍历每个类别"]

    subgraph PerCatClean["每个类别的清理流程"]
        direction TB
        A["清空 totalSize"] --> B["遍历 cat.Paths"]
        B --> C["路径存在?"]
        C -->|否| Skip["跳过"]
        C -->|是| D["计算目录大小"]
        D --> DryRun{"--dry-run?"}
        DryRun -->|是| E["status=dry-run\n不执行删除"]
        DryRun -->|否| Recover{"--recoverable?"}
        Recover -->|是| F["tar -czf 压缩\n→ ~/.os-cleaner/recover/"]
        Recover -->|否| G["直接删除"]
        F --> G
        G --> PermCheck{"检查权限错误"}
        PermCheck -->|无错误| H["status=cleaned"]
        PermCheck -->|部分错误| I["status=partial\n计算已删大小"]
        PermCheck -->|全部错误| J["status=error\n建议 sudo"]
    end

    Loop --> PerCatClean
    PerCatClean --> Results["收集 CleanResult"]

    Results --> Output{"--json?"}
    Output -->|是| JOut["JSON 输出"]
    Output -->|否| TOut["表格输出\n显示每个类别的\n状态/大小/路径"]
    TOut --> Total["汇总释放空间"]
```

### 类别名称解析过程（`cmd/clean.go:69-94`）

```mermaid
flowchart LR
    A["用户输入\nxcode-sim"] --> B["registry.GetCategoryByID()\n精确匹配 ID"]
    B --> C{"找到?"}
    C -->|是| D["返回该类别"]
    C -->|否| E["遍历 GetCategoriesByPlatform()\nName 不区分大小写匹配"]
    E --> F{"找到?"}
    F -->|是| D
    F -->|否| G["遍历\nstrings.Contains\nID 或 Name 包含输入"]
    G --> H{"找到?"}
    H -->|是| D
    H -->|否| I["返回 nil→显示可用类别"]
```

这意味着用户输入 `xcode` 即可匹配 `xcode-deriveddata`、`xcode-simulator` 等，无需完整 ID。

---

## 3. 大文件扫描流程（top）

```mermaid
flowchart TD
    START(["os-cleaner top [path] [-n 20] [--json]"])
    START --> SetPath{"指定了路径?"}
    SetPath -->|否| CWD["使用当前工作目录"]
    SetPath -->|是| UserPath["使用用户指定路径"]
    CWD --> CheckDir{"是有效目录?"}
    UserPath --> CheckDir
    CheckDir -->|否| Error["返回错误"]
    CheckDir -->|是| ReadDir["os.ReadDir()\n读取一级条目"]

    ReadDir --> InitProgress["初始化进度条"]
    InitProgress --> Parallel["为每个条目启动 goroutine"]

    subgraph PerEntry["每个 goroutine"]
        direction TB
        E1["获取文件信息"] --> E2{"是目录?"}
        E2 -->|是| E3["递归遍历子目录\natomic.Int64 累加大小和文件数"]
        E2 -->|否| E4["取 info.Size()"]
        E3 --> E5["返回 TopItem"]
        E4 --> E5
    end

    Parallel --> PerEntry
    PerEntry --> Wait["wg.Wait()"]
    Wait --> Sort["按大小降序排序"]
    Sort --> Limit["截取前 N 项\n（默认 20）"]
    Limit --> Output{"--json?"}

    Output -->|是| JOut["JSON 输出\n含路径和完整列表"]
    Output -->|否| Table["终端表格输出"]
    Table --> Header["打印目录和总大小"]
    Header --> Items["逐行输出"]
    Items --> Colorize{"大小阈值"}
    Colorize -->|≥ 1GB| Red["红色 ⚠️"]
    Colorize -->|≥ 100MB| Yellow["黄色 ⚡"]
    Colorize -->|< 100MB| Normal["默认颜色"]
    Items --> Summary["汇总: ≥1GB 项数、≥100MB 项数"]
```

### 大小阈值颜色编码

| 大小 | 颜色 | 标记 | 代码位置 |
|------|------|------|----------|
| ≥ 1GB (1073741824 bytes) | 红色 | `⚠️` | `topscan.go:32` `DangerThreshold` |
| ≥ 100MB (104857600 bytes) | 黄色 | `⚡` | `topscan.go:31` `WarnThreshold` |
| < 100MB | 默认 | 无 | — |

---

## 4. 完整命令关系总图

```mermaid
flowchart TD
    CLI["os-cleaner"] --> Scan["scan"]
    CLI --> List["list"]
    CLI --> Clean["clean"]
    CLI --> Inspect["inspect"]
    CLI --> Top["top"]
    CLI --> Active["active"]

    Scan --> S1["并发扫描所有缓存路径"]
    S1 --> S2["按等级汇总空间"]
    S2 --> S3["按时间过滤（可选）"]

    List --> L1["列出所有类别定义"]
    L1 --> L2["含 ID/安全等级/平台"]

    Clean --> C1["按类别/安全模式筛选"]
    C1 --> C2["dry-run 预览"]
    C1 --> C3["recoverable 压缩"]
    C2 --> C4["安全删除+权限处理"]

    Inspect --> I1["查找缓存路径"]
    I1 --> I2{"类型识别"}
    I2 --> I2a["npm: 调用 npm cache ls + 注册表分析"]
    I2 --> I2b["Go: du 扫描 mod 缓存"]
    I2 --> I2c["Python: pip wheel / pyenv 版本"]
    I2 --> I2d["Cargo: registry cache 分析"]
    I2 --> I2e["Homebrew: 缓存 bottle 列表"]
    I2 --> I2f["Xcode: DerivedData 项目维度"]
    I2 --> I2g["通用: du 显示子目录"]

    Top --> T1["并发统计目录条目"]
    T1 --> T2["排序+阈值颜色"]

    Active --> A1["递归查找项目文件"]
    A1 --> A2{"文件类型"}
    A2 --> A2a["package.json → Node.js"]
    A2 --> A2b["go.mod → Go"]
    A2 --> A2c["Cargo.toml → Rust"]
    A2 --> A2d["requirements.txt → Python"]
    A2 --> A2e["Gemfile → Ruby"]
    A2 --> A2f["mix.exs → Elixir"]
    A2 --> A2g["pom.xml → Maven"]
    A2 --> A2h["其他 → 跳过"]
    A2a --> A3["汇总所有项目使用的包"]
    A2b --> A3
    A2c --> A3
    A2d --> A3
    A2e --> A3
    A2f --> A3
    A2g --> A3
    A3 --> A4["对比 inspect 结果 →\n发现未使用的包"]
```
