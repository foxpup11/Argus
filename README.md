<p align="center">
  <h1 align="center">🔍 AgentScope</h1>
  <p align="center">
    <strong>AI Agent 改了你的代码？3 秒看清全貌。</strong>
  </p>
  <p align="center">
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go">
    <img src="https://img.shields.io/badge/Wails-v2-5C2D91?style=flat-square" alt="Wails">
    <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square" alt="Platform">
    <img src="https://img.shields.io/github/license/foxpup11/agentscope?style=flat-square" alt="License">
  </p>
</p>

---

## 🎯 一句话

**AgentScope** 让你一眼看清 AI Agent 对代码做了什么 —— 改了哪些文件、为什么改、风险多高。

## ✨ 为什么需要 AgentScope？

> *"Claude Code 改了 20 个文件，我根本不知道改了啥..."*

每次用 AI Agent 写代码后，你是不是也有这种感觉？

- 😵 改动太多，`git diff` 看不过来
- 🤔 不知道每个改动对应哪个操作
- ⚠️ 担心 Agent 误删文件或执行危险命令

**AgentScope 帮你解决这些问题。**

## 🚀 开箱即用

### 1. 下载

前往 [Releases](https://github.com/foxpup11/agentscope/releases) 下载对应平台的可执行文件：

| 平台 | 下载 |
|------|------|
| 🪟 Windows | `agentscope-windows-amd64.exe` |
| 🍎 macOS | `agentscope-darwin-arm64.zip` |
| 🐧 Linux | `agentscope-linux-amd64.tar.gz` |

### 2. 运行

双击运行，自动扫描 Claude Code 会话：

```bash
# Windows
agentscope.exe

# macOS / Linux
./agentscope
```

### 3. 查看

- 左侧选择会话
- 右侧查看文件改动
- 点击文件查看 Diff

**就这么简单。**

## 🛡️ 风险评估

AgentScope 自动评估每个改动的风险等级：

| 等级 | 触发条件 | 示例 |
|------|----------|------|
| 🟢 **Safe** | 新增文件、小改动、文档修改 | `README.md`, `docs/` |
| 🟡 **Review** | 删除代码、修改依赖、多次编辑 | `go.mod`, `package.json` |
| 🔴 **Danger** | 敏感文件、危险命令、大量改动 | `.env`, `rm -rf` |

## 🔧 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | [Wails](https://wails.io/) v2 |
| 后端 | Go |
| 前端 | HTML + CSS + JS |
| 设计风格 | Apple HIG / Material Design |

## 📦 从源码构建

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

## 🤝 Contributing

欢迎贡献！Fork → Branch → PR。

## 📝 License

MIT

---

<p align="center">
  <strong>如果 AgentScope 帮到了你，请给个 ⭐ Star 支持一下！</strong>
</p>
<p align="center">
  <sub>Your star motivates me to keep improving 🚀</sub>
</p>
