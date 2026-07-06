# Contributing to Argus

感谢你对 Argus 的关注！

## 如何贡献

### 报告 Bug

1. 在 [Issues](https://github.com/foxpup11/argus/issues) 中搜索是否已有相同问题
2. 如果没有，创建新 Issue，包含：
   - 问题描述
   - 复现步骤
   - 期望行为
   - 实际行为
   - 环境信息（OS、Go 版本、Wails 版本）

### 提交功能建议

1. 在 Issues 中创建新 Issue，标签选择 `enhancement`
2. 描述功能需求和使用场景

### 提交代码

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'feat: add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 创建 Pull Request

### 开发环境

```bash
# 克隆仓库
git clone https://github.com/your-name/argus.git
cd argus

# 安装依赖
go mod tidy

# 启动开发模式
wails dev

# 运行测试
go test ./...
```

### 代码规范

- 遵循 Go 官方代码规范
- 提交信息使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式
- 新功能需要添加测试
- PR 需要通过 CI 检查

## 行为准则

- 尊重每一位参与者
- 接受建设性批评
- 关注对社区最有利的事情
- 对其他成员表示同理心

## 问题反馈

- Issues: https://github.com/foxpup11/argus/issues
- Email: sizhen02621@gmail.com

感谢你的贡献！ 🎉
