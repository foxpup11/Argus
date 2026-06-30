<p align="center">
  <h1 align="center">🔍 AgentScope</h1>
  <p align="center">
    <strong>AI Agent 改了你的代码？3 秒看清全貌。</strong>
  </p>
  <p align="center">
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go">
    <img src="https://img.shields.io/badge/Wails-v2-5C2D91?style=flat-square" alt="Wails">
    <img src="https://img.shields.io/badge/Platform-Windows-lightgrey?style=flat-square" alt="Platform">
    <img src="https://img.shields.io/github/license/foxpup11/agentscope?style=flat-square" alt="License">
    <img src="https://img.shields.io/github/stars/foxpup11/agentscope?style=social" alt="Stars">
  </p>
</p>

---

## ✨ 这是什么？

**AgentScope** 是一个桌面应用，帮你一眼看清 **AI Agent（如 Claude Code）对代码做了什么**。

### 🎯 解决什么问题？

> *"Claude Code 改了 20 个文件，我根本不知道改了啥..."*

每次用 AI Agent 写代码后，你是不是也有这种感觉？

- 😵 改动太多，`git diff` 看不过来
- 🤔 不知道每个改动对应哪个操作
- ⚠️ 担心 Agent 误删文件或执行危险命令

**AgentScope 帮你解决这些问题。**

---

## 🚀 30 秒开始使用

### 第一步：下载

前往 [Releases](https://github.com/foxpup11/agentscope/releases/latest) 下载：

| 平台 | 文件 |
|------|------|
| 🪟 **Windows** | `agentscope-desktop.exe` (11MB) |

### 第二步：运行

**双击** `agentscope-desktop.exe`，无需安装。

### 第三步：查看

1. 左侧选择一个会话
2. 右侧查看文件改动
3. 点击文件查看 Diff

**就这么简单！**

---

## 📸 界面预览

<p align="center">
  <img src="docs/images/preview.png" alt="AgentScope Preview" width="800">
</p>

---

## 🛠️ 核心功能

| 功能 | 描述 |
|------|------|
| 📂 **会话列表** | 左侧面板展示所有 Claude Code 会话 |
| 📊 **文件改动** | 显示每个文件的风险等级、变更类型 |
| 🔍 **Diff 预览** | 语法高亮的代码差异查看 |
| 🛡️ **风险评估** | 自动标注 Safe / Review / Danger |
| 🔀 **中英文切换** | 支持中英文界面 |
| ↔️ **可拖拽布局** | 拖动调整侧边栏宽度 |

---

## 🛡️ 风险评估

AgentScope 自动评估每个改动的风险等级：

| 等级 | 触发条件 | 示例 |
|------|----------|------|
| 🟢 **Safe** | 新增文件、小改动、文档修改 | `README.md`, `docs/` |
| 🟡 **Review** | 删除代码、修改依赖、多次编辑 | `go.mod`, `package.json` |
| 🔴 **Danger** | 敏感文件、危险命令、大量改动 | `.env`, `rm -rf` |

---

## 🗺️ Roadmap

- [x] Claude Code 会话解析
- [x] 文件改动列表
- [x] Diff 语法高亮
- [x] 风险等级评估
- [x] 中英文切换
- [x] 可拖拽分隔栏
- [ ] 实时监控新会话
- [ ] 会话导出 (HTML/Markdown)
- [ ] 支持 Codex CLI
- [ ] 支持 OpenCode
- [ ] 支持 Aider

---

## 🔧 从源码构建

```bash
# 前置要求
# - Go 1.21+
# - Wails CLI: go install github.com/wailsapp/wails/v2/cmd/wails@latest

git clone https://github.com/foxpup11/agentscope.git
cd agentscope

# 安装依赖
go mod tidy

# 开发模式
wails dev

# 构建生产版本
wails build
```

---

## 🤝 Contributing

欢迎贡献！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 📝 License

MIT License

---

<p align="center">
  <strong>如果 AgentScope 帮到了你，请给个 ⭐ Star 支持一下！</strong>
</p>
<p align="center">
  <sub>Your star motivates me to keep improving 🚀</sub>
</p>
