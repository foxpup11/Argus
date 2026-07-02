<p align="center">
  <img src="docs/images/Logo.png" alt="AgentScope Logo" width="280">
</p>

<h1 align="center">AgentScope</h1>

<p align="center">
  <strong>Claude Code 的上帝视角 -- 你再也回不去"盲人摸象"的时代了</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Wails-v2-5C2D91?style=for-the-badge" alt="Wails">
  <img src="https://img.shields.io/badge/Platform-Windows-lightgrey?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/badge/Version-v0.5.1-blue?style=for-the-badge" alt="Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <a href="https://github.com/foxpup11/agentscope/releases/latest">📥 下载 (12MB)</a> ·
  <a href="#-你是不是也遇到过这些问题">🎯 看痛点</a> ·
  <a href="#-它能帮你做什么">💡 看方案</a> ·
  <a href="#-为什么选择-agentscope">🔥 看优势</a>
</p>

---

## 你是不是也遇到过这些问题？

> **"每次开新会话，Claude 就像失忆了一样，把昨天做过的事又做一遍。"**

你花了 2 小时教它理解你的项目架构，第二天开新会话，它又从零开始问你"请问你的项目用的是什么技术栈？"

> **"月底账单出来，我才发现 token 消耗比我想的多了 10 倍。"**

你以为只是正常写代码，结果发现 Claude 陷入了循环读取同一个文件的陷阱，3 分钟烧掉了你 5 小时的配额。

> **"Claude 删了我关键文件，我完全不知道，直到 CI 挂了才发现。"**

它静悄悄地把你的配置文件改了、把测试删了，你从头到尾蒙在鼓里。

> **"CLAUDE.md 写了一堆规则，Claude 根本不遵守。"**

你说"先写测试再写实现"，它直接跳过了测试文件。你说"不要改 config"，它把你的 docker-compose.yml 改得面目全非。

**这些问题的根源是同一个：Claude Code 只看得到当前会话，它没有全局视野。**

而 AgentScope，就是给你的 Claude Code 装上一双"天眼"。

---

## 它能帮你做什么？

### 1. 跨会话连续性 -- 让 Claude "记住"一切

**不再从零开始。**

AgentScope 会自动从历史会话中提取已完成的任务、待办事项、关键决策，生成标准化的交接摘要。下一次开新会话？直接导入上下文，秒级进入状态。

| 没有 AgentScope | 有了 AgentScope |
|----------------|----------------|
| 每个新会话都要重新解释项目背景 | 一键导入上次会话的上下文 |
| Claude 重复执行已完成的工作 | 清楚知道什么做完了、什么还没做 |
| 你说"上次已经改了那个 bug"，它说"哪个 bug？" | 自动列出所有历史修改和决策 |
| 50-75% 时间浪费在"恢复记忆"上 | 零切换成本，直接继续 |

### 2. Token 费用透明 -- 每一分钱花在哪

**你值得知道自己的钱花在了什么上面。**

5 张仪表盘卡片实时展示今日/本月/累计消耗，30 天趋势图让你一眼看出哪些天烧得最多。按项目、按模型拆分分析，找到那个"偷吃"token 的元凶。

### 3. 风险自动评估 -- 你的 AI 操作守护者

**危险操作，一眼识别。**

AgentScope 内置智能风险评估引擎，自动给每次文件改动打分：

- **红色警告 (Danger)**：删除文件、修改 `.env`、执行 `rm -rf`
- **黄色提醒 (Review)**：修改依赖、删除代码、改动配置文件
- **绿色通过 (Safe)**：新增文件、小改动、文档修改

再也不用担心 Claude 静悄悄搞破坏。

### 4. 完整对话回放 -- AI 的"黑匣子"

**它到底想了什么、做了什么，全都在你眼前。**

- 完整回放用户消息、AI 思考过程、工具调用链
- 每条消息精确到秒的时间戳
- 一键搜索所有历史对话，再也不用翻聊天记录

### 5. 知识库管理 -- 项目记忆永不丢失

**CLAUDE.md、Plans、Memory，集中管理，一键生成。**

- CLAUDE.md 可视化编辑器：分段编辑 + 实时预览
- CLAUDE.md 智能生成器：自动检测项目结构，一键生成最佳实践配置
- 知识库集中管理：所有项目记忆、计划文档、配置文件在一个地方

---

## 为什么选择 AgentScope？

### 零配置，零侵入

- **不需要 API Key** -- 不调用任何外部服务
- **不需要联网** -- 完全离线运行，你的数据不会离开你的电脑
- **不需要安装** -- 一个 12MB 的 exe，双击即用
- **不需要配置** -- 自动扫描 `~/.claude/projects/`，开箱即用

### 真正的"外部观察者"

AgentScope 运行在 Claude Code **外部**，这意味着：

- 它能看到 Claude Code 自身看不到的全局数据
- 它不会干扰 Claude Code 的正常工作
- 它的安全性天然高于任何"插件"或"扩展"

### 你的数据，你做主

所有数据都存储在本地，AgentScope 只是读取 Claude Code 已经产生的会话文件。它不上传、不分析、不共享你的任何数据。

---

## 功能全景

<details>
<summary><strong>🔄 会话连续性引擎</strong></summary>

| 功能 | 说明 |
|------|------|
| 跨会话任务提取 | 从历史会话中自动识别已完成任务、待办事项、关键决策 |
| Git 交叉验证 | 对比 git 记录验证任务是否真正提交，防止"说过了但没做" |
| 文件概览 | 统计每个文件的操作次数和类型 (代码/测试/配置) |
| 问题发现 | 识别已知问题和陷阱 |
| 多种导出 | 导出到 memory 文件 / Markdown / 可粘贴 prompt 片段 |

</details>

<details>
<summary><strong>📊 Token 仪表盘</strong></summary>

| 功能 | 说明 |
|------|------|
| 5 张概览卡片 | 今日 / 本月 / 上月 / 累计 Token + 总会话数 |
| 30 天趋势图 | Input + Output 堆叠柱状图 |
| 项目维度分析 | 按项目分组统计 Token 消耗 |
| 模型维度分析 | 识别不同模型使用占比 |
| 月度对比 | 本月 vs 上月百分比变化 |

</details>

<details>
<summary><strong>💬 对话记录 & 工具调用</strong></summary>

| 功能 | 说明 |
|------|------|
| 完整回放 | 用户消息、AI 回复、思考过程 |
| 工具调用 | 展示 AI 调用了哪些工具 |
| 时间戳 | 每条消息精确到秒 |

</details>

<details>
<summary><strong>📝 文件改动 & Diff</strong></summary>

| 功能 | 说明 |
|------|------|
| 风险评估 | Danger / Review / Safe 自动分级 |
| Diff 查看 | 语法高亮，支持未提交 / 会话对比两种模式 |
| 变更类型 | Created / Modified / Deleted 一目了然 |

</details>

<details>
<summary><strong>📚 知识库管理</strong></summary>

| 功能 | 说明 |
|------|------|
| Plans | 管理 Claude Code 的计划文档 |
| Memory | 管理项目记忆文件 |
| CLAUDE.md 编辑器 | 分段编辑 + 实时预览 |
| CLAUDE.md 生成器 | 自动检测项目结构，一键生成 |

</details>

<details>
<summary><strong>🗂️ 会话管理</strong></summary>

| 功能 | 说明 |
|------|------|
| 全文搜索 | 搜索 Prompt、模型、分支、标签 |
| 标签系统 | 手动打标签 + 17 条自动识别规则 |
| 会话收藏 | 一键收藏重要会话 |
| 批量操作 | 批量收藏、导出、删除 |
| 项目分组 | 按项目自动分组，折叠/展开 |
| 实时监控 | 文件系统变化自动刷新 |

</details>

<details>
<summary><strong>🎨 其他特性</strong></summary>

| 功能 | 说明 |
|------|------|
| 主题 | 浅色 / 深色 / 跟随系统 |
| 自定义风险规则 | 按文件路径模式匹配自定义规则 |
| 导出 | 单条或批量导出为 Markdown |
| 国际化 | 中文 / English 一键切换 |

</details>

---

## 预览

(docs/images/preview0.png)

(docs/images/preview1.png)

(docs/images/preview2.png)

(docs/images/preview3.png)

(docs/images/preview4.png)

(docs/images/preview5.png)

---

## 快速上手

### 第一步：下载

前往 [Releases](https://github.com/foxpup11/agentscope/releases/latest) 下载最新版本。

| 文件 | 大小 | 说明 |
|------|------|------|
| `agentscope-desktop.exe` | ~12MB | Windows 可执行文件，免安装 |

### 第二步：双击运行

不需要安装，不需要配置，不需要 API Key。双击 `agentscope-desktop.exe`，直接打开。

### 第三步：开始使用

选择左侧任意会话，即可查看对话记录、文件改动、Token 消耗。

**就这么简单。**

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

- [ ] CLAUDE.md 规则遵守审计 -- 审计 Claude 是否遵守了你写的每一条规则
- [ ] 成本异常预警系统 -- 消耗速率异常时桌面通知告警
- [ ] 提示词效能分析器 -- 帮你优化 prompt，减少 token 浪费
- [ ] 上下文健康仪表盘 -- 实时监控上下文退化，提醒你何时该开新会话
- [ ] 插件 & Hook 配置工作室 -- 可视化管理 Claude Code 的插件生态
- [ ] macOS / Linux 支持

详细技术方案见 [docs/ROADMAP.md](docs/ROADMAP.md)。

---

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端 | Go 1.23 + Wails v2 | 高性能、跨平台桌面框架 |
| 前端 | HTML/CSS/JS | 原生前端，无框架依赖，轻量快速 |
| 存储 | SQLite + JSON | 本地存储，零网络依赖 |
| 文件监控 | fsnotify | 实时感知 Claude Code 会话变化 |

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
