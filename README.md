<p align="center">
  <img src="docs/images/Logo.png" alt="AgentScope Logo" width="280">
</p>

<h1 align="center">AgentScope</h1>

<p align="center">
  <strong>Claude Code 的运维控制台 -- 跨会话全局视角，对自身行为的元分析</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Wails-v2-5C2D91?style=for-the-badge" alt="Wails">
  <img src="https://img.shields.io/badge/Platform-Windows-lightgrey?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/badge/Version-v0.5.0-blue?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <a href="https://github.com/foxpup11/agentscope/releases/latest">下载</a> ·
  <a href="#-快速上手">快速上手</a> ·
  <a href="#-功能一览">功能一览</a> ·
  <a href="#-从源码构建">构建</a> ·
  <a href="#-roadmap">Roadmap</a>
</p>

---

## 这是什么？

**AgentScope** 运行在 Claude Code **外部**，拥有 Claude Code 自身看不到的 **跨会话全局视角**。

它直接读取 `~/.claude/projects/` 下的本地会话文件，帮你做到：

- **看清每次对话** -- 完整回放用户消息、AI 思考过程、工具调用
- **Token 费用透明** -- 实时统计今日/本月/累计消耗，趋势图一目了然
- **风险自动评估** -- 文件改动分级 (Safe / Review / Danger)，危险操作一眼识别
- **知识库管理** -- Plans、Memory、CLAUDE.md 集中管理，支持一键生成
- **[新] 会话连续性引擎** -- 跨会话分析，生成交接摘要，解决"AI 土拨鼠日"问题

> **零配置、零侵入** -- 不需要 API Key，不需要联网，它只是帮你更清楚地看到 Claude Code 已经做了什么。

---

## 快速上手

### 1. 下载

前往 [Releases](https://github.com/foxpup11/agentscope/releases/latest) 下载最新版本 (约 12MB，无需安装)。

### 2. 运行

双击 `agentscope-desktop.exe`，直接打开。

### 3. 开始使用

选择左侧任意会话，即可查看对话记录、文件改动、Token 消耗。

---

## 功能一览

### 会话连续性引擎 (v0.5.0 新增)

解决 Claude Code 最大的痛点 -- 每个新会话从零开始，重复执行已完成工作。

| 功能 | 说明 |
|------|------|
| 跨会话任务提取 | 从历史会话中自动识别已完成任务、待办事项、关键决策 |
| Git 交叉验证 | 对比 git 记录验证任务是否真正提交，防止"说过了但没做" |
| 文件概览 | 统计每个文件的操作次数和类型 (代码/测试/配置) |
| 问题发现 | 识别已知问题和陷阱 |
| 多种导出 | 导出到 memory 文件 / Markdown / 可粘贴 prompt 片段 |

### Token 仪表盘

| 功能 | 说明 |
|------|------|
| 5 张概览卡片 | 今日 / 本月 / 上月 / 累计 Token + 总会话数 |
| 30 天趋势图 | Input + Output 堆叠柱状图 |
| 项目维度分析 | 按项目分组统计 Token 消耗 |
| 模型维度分析 | 识别不同模型使用占比 |
| 月度对比 | 本月 vs 上月百分比变化 |

### 会话管理

| 功能 | 说明 |
|------|------|
| 全文搜索 | 搜索 Prompt、模型、分支、标签 |
| 标签系统 | 手动打标签 + 17 条自动识别规则 |
| 会话收藏 | 一键收藏重要会话 |
| 批量操作 | 批量收藏、导出、删除 |
| 项目分组 | 按项目自动分组，折叠/展开 |
| 实时监控 | 文件系统变化自动刷新 |

### 对话记录

| 功能 | 说明 |
|------|------|
| 完整回放 | 用户消息、AI 回复、思考过程 |
| 工具调用 | 展示 AI 调用了哪些工具 |
| 时间戳 | 每条消息精确到秒 |

### 文件改动 & Diff

| 功能 | 说明 |
|------|------|
| 风险评估 | Danger / Review / Safe 自动分级 |
| Diff 查看 | 语法高亮，支持未提交 / 会话对比两种模式 |
| 变更类型 | Created / Modified / Deleted 一目了然 |

### 知识库

| 功能 | 说明 |
|------|------|
| Plans | 管理 Claude Code 的计划文档 |
| Memory | 管理项目记忆文件 |
| CLAUDE.md 编辑器 | 分段编辑 + 实时预览 |
| CLAUDE.md 生成器 | 自动检测项目结构，一键生成 |

### 其他特性

| 功能 | 说明 |
|------|------|
| 主题 | 浅色 / 深色 / 跟随系统 |
| 自定义风险规则 | 按文件路径模式匹配自定义规则 |
| 导出 | 单条或批量导出为 Markdown |
| 国际化 | 中文 / English 一键切换 |
| 开箱即用 | 无需安装、无需 API Key、离线可用 |

---

## 预览

![Token 仪表盘](docs/images/preview.png)

![会话管理](docs/images/preview2.png)
![预览图](docs/images/preview3.png)

---

## 风险评估引擎

AgentScope 内置智能风险评估，自动识别危险操作：

| 风险等级 | 触发条件 | 你会看到 |
|----------|----------|----------|
| **Danger** | 删除文件、敏感文件、危险命令 | 红色警告 |
| **Review** | 修改依赖、多次编辑、删除代码 | 黄色提醒 |
| **Safe** | 新增文件、小改动、文档修改 | 绿色通过 |

**检测的危险操作：**
- 删除文件
- 敏感文件 (`.env`, `secret`, `password`, `token`)
- 危险命令 (`rm -rf`, `chmod 777`, `curl | bash`)
- 配置文件 (`.git/config`, `Dockerfile`, CI 配置)

---

## 下载安装

### 方式一：直接下载 (推荐)

前往 [Releases](https://github.com/foxpup11/agentscope/releases/latest) 下载最新版本。

| 文件 | 说明 |
|------|------|
| `agentscope-desktop.exe` | Windows 可执行文件，约 12MB |

### 方式二：从源码构建

```bash
# 前置条件
# - Go 1.21+
# - Wails v2: go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆仓库
git clone https://github.com/foxpup11/agentscope.git
cd agentscope

# 安装依赖
go mod tidy

# 开发模式 (热重载)
wails dev

# 构建生产版本
wails build -platform windows/amd64 -o agentscope-desktop.exe
```

---

## Roadmap

### 已完成

- [x] Claude Code 会话解析
- [x] 文件改动列表与风险评估
- [x] Diff 语法高亮
- [x] 中英文切换
- [x] 可拖拽布局
- [x] 实时监控
- [x] 会话导出
- [x] 深色主题
- [x] 自定义风险规则
- [x] 项目分组
- [x] Token 仪表盘 (5 卡片 + 趋势图 + 双表格)
- [x] 全文搜索 + 高级筛选
- [x] 标签系统 (手动 + 17 条自动识别)
- [x] 会话收藏 + 批量操作
- [x] 对话记录完整回放
- [x] CLAUDE.md 可视化编辑器
- [x] CLAUDE.md 模板生成器
- [x] 知识库管理 (Plans / Memory / CLAUDE.md)
- [x] **会话连续性引擎** -- 跨会话任务提取、Git 验证、交接摘要生成

### 计划中

- [ ] CLAUDE.md 规则遵守审计
- [ ] 成本异常预警系统
- [ ] 提示词效能分析器
- [ ] 上下文健康仪表盘
- [ ] 插件 & Hook 配置工作室
- [ ] macOS / Linux 支持

详细技术方案见 [docs/ROADMAP.md](docs/ROADMAP.md)。

---

## 贡献

欢迎贡献！无论是提交 Bug、建议新功能，还是直接贡献代码。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## License

MIT License -- 自由使用，自由分享。

---

<p align="center">
  <sub>Made by <a href="https://github.com/foxpup11">foxpup11</a></sub>
</p>
