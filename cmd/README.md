# cmd — 终端对话框（TUI）

`cmd` 包是 digtast 的终端对话框实现，基于 [bubbletea](https://github.com/charmbracelet/bubbletea)（Elm 架构：`Model` 通过 `Init/Update/View` 驱动界面）。本文说明**如何在程序运行后的终端里添加一个 `/命令`**，以 `/skills` 为完整示例。

## TUI 的基本结构

| 文件 | 职责 |
| --- | --- |
| `tui.go` | 屏幕（screen）枚举、`uiModel` 状态、Update/View 总路由、聊天界面 |
| `tui_models.go` | `/model`：厂商配置与模型启用流程 |
| `tui_skills.go` | `/skills`：skill 查看与启停 |
| `tui_channels.go` | `/channel`：消息渠道配置 |
| `tui_form.go` | 通用表单（textinput 字段、Tab 切换、提交校验） |
| `tui_picker.go` | 通用列表选择器（单选/多选、输入过滤、加载态） |
| `tui_style.go` | 样式（`styleTitle/styleBody/styleSelected/...`） |

核心概念：

- **screen**：一个全屏页面就是一个 `screenID` 枚举值（`tui.go` 顶部）。同一时刻只有一个 screen 处于活跃状态。
- **`uiModel`**：所有界面共享的唯一状态容器，字段按功能分组（`providers`、`skillsView`、`channelForm`…）。打开某屏时把该屏的 view 结构体挂到对应字段上，`View()` 里据此渲染。
- **三件套**：每屏由 `openXxx`（进入并装载状态）、`updateXxx`（按键处理）、`viewXxx`（渲染）三个方法构成。
- **命令面板**：输入 `/` 后自动弹出 `chatCommands` 里的命令提示，无需为命令做额外的注册。

## 添加一个命令的步骤

### 1. 在 `chatCommands` 里登记命令

`tui.go`：

```go
var chatCommands = []command{
    {name: "skills", desc: "查看扫描到的 skills，启用或停用"},
    // ...
}
```

登记后，用户输入 `/ski` 就能看到过滤提示，`Tab` 补全。

### 2. 在 `runCommand` 里分发

`tui.go` 的 `runCommand` switch 中加一个分支，进入对应 screen：

```go
case "skills", "skill":      // 复数为主名，单数做别名
    m.openSkills()
```

命令在流式回复期间默认被拒绝（`runCommand` 开头的判断），除非像 `help`/`exit` 那样加入白名单。

### 3. 声明 screen 并挂到路由

```go
const (
    screenChat screenID = iota
    // ...
    screenSkills    // 新增
    // ...
)
```

两处路由（都在 `tui.go`）：

```go
// Update：按键分发
case screenSkills:
    return m.updateSkills(msg)

// View：页面渲染（view 为空时回落到聊天页，防止空指针）
case screenSkills:
    if m.skillsView == nil {
        m.screen = screenChat
        return m.viewChat()
    }
    return m.viewSkills()
```

### 4. 实现三件套（tui_skills.go）

```go
// 视图状态：光标 + 条目
type skillView struct {
    skills []skill.SkillInfo
    cursor int
}

// 进入：装载状态并切换 screen
func (m *uiModel) openSkills() {
    skills, err := skill.NewManager().GetAllSkills()
    if err != nil {
        m.setStatus(true, "读取 skills 失败: %v", err)   // 不切换 screen
        return
    }
    m.skillsView = &skillView{skills: skills}
    m.screen = screenSkills
}

// 按键：up/down 移动、Space/Enter 切换、r 重扫、esc 返回
func (m *uiModel) updateSkills(msg tea.KeyMsg) (tea.Model, tea.Cmd) { ... }

// 渲染：标题 + 列表（当前行 styleSelected）+ 底部键位提示
func (m *uiModel) viewSkills() string { ... }
```

惯例：

- 返回聊天页时调用 `m.input.Focus()`，否则输入框不响应。
- 顶部 `styleTitle`，正文 `styleBody`，当前行 `styleSelected`，键位提示 `styleMuted`，路径/落点说明 `styleFooter`。
- 状态消息用 `m.setStatus(isErr, format, args...)`，渲染时 `m.statusLine()`。

### 5. 需要持久化时写 config

界面状态改动直接调用 `config` 包的保存函数（如 `config.SetSkillEnabled`、`config.SaveModelCredentials`），它们会写回 `~/.digtast/config.json`。新配置项要在 `config.Config` 与 `storedConfig` **两处**声明同名字段，否则重新加载时会丢失。

### 6. 需要异步工作时用 tea.Cmd + 自定义消息

耗时操作（网络请求、目录扫描）不能阻塞 Update。模式见 `tui_models.go` 的模型列表拉取：

```go
// 1. 定义消息类型
type modelsFetchedMsg struct {
    gen   int       // 代数，丢弃过期结果
    names []string
    err   error
}

// 2. 发起：返回 tea.Cmd，立即切到加载态 screen
func (m *uiModel) fetchProviderModels() tea.Cmd {
    m.fetchGen++
    gen := m.fetchGen
    // ...picker.loading = true
    return func() tea.Msg {                       // 在 goroutine 中执行
        names, err := provider.ListModels(...)
        return modelsFetchedMsg{gen: gen, names: names, err: err}
    }
}

// 3. Update 顶部处理消息，刷新视图
case modelsFetchedMsg:
    m.handleModelsFetched(msg)
    return m, nil
```

用 `gen`（自增代数）丢弃用户已离开该页面后返回的过期结果。

### 7. 更新帮助与测试

- `viewHelp` 里补一行命令说明。
- 测试写在 `tui_test.go`：直接构造 view 结构体断言行为（参考 `TestSkillViewToggle`），不需要启动 TUI。

## 键位约定

| 键 | 语义 |
| --- | --- |
| `↑/↓` 或 `k/j` | 列表移动光标 |
| `Enter` | 确认 / 使用 / 进入 |
| `Space` | 多选启用/取消（picker 与开关列表） |
| `Esc` | 返回上一级；聊天中先取消流式回复 |
| `Tab` / `Shift+Tab` | 表单字段切换 |
| `Ctrl+R` | 重新拉取（有网络请求的页面） |

注意：`pickerView.handleKey` 会消费方向键、退格、普通字符（过滤输入）和 `Esc`（清空过滤），先调用它再处理自己的键，未被消费的键才进入页面逻辑。
