# AgentScope Desktop - BUG 追踪文档

**审查日期:** 2026-06-30
**审查范围:** 全项目代码审查

---

## BUG 概览

| 优先级 | 数量 | 状态 |
|--------|------|------|
| P0 - 严重 | 3 | ✅ 已修复 |
| P1 - 高 | 4 | ✅ 已修复 |
| P2 - 中 | 4 | 待修复 |
| P3 - 低 | 4 | 待修复 |

---

## P0 - 严重 BUG (Critical)

### BUG #1: 结构体 JSON 标签错误
**文件:** `app.go:44`
**状态:** ✅ 已修复

**问题描述:**
`FileChangeInfo` 结构体的 `ActionCount` 字段标签缺少 `json:` 前缀。

**当前代码:**
```go
ActionCount int   `actionCount`
```

**修复方案:**
```go
ActionCount int   `json:"actionCount"`
```

**影响:** JSON 序列化时字段名不正确，前端无法正确接收数据。

---

### BUG #2: 风险评估结果未被使用
**文件:** `app.go:234`
**状态:** ✅ 已修复

**问题描述:**
`EvaluateAll` 返回评估后的 `[]session.FileChange`，但代码没有接收返回值，风险评估结果未生效。

**当前代码:**
```go
riskEngine.EvaluateAll(fileChanges)
```

**修复方案:**
```go
fileChanges = riskEngine.EvaluateAll(fileChanges)
```

**影响:** 所有文件都会显示为默认风险等级，无法正确识别高风险操作。

---

### BUG #3: HTML 模板引用不存在的字段
**文件:** `internal/export/templates.go:448`
**状态:** ✅ 已修复

**问题描述:**
模板引用 `.ActionCount` 字段，但 `session.FileChange` 结构体中没有定义此字段。

**修复方案:**
在 `session.FileChange` 结构体中添加 `ActionCount` 字段，或在导出时动态计算。

---

## P1 - 高优先级 BUG (High)

### BUG #4: Git Diff 函数传入错误路径类型
**文件:** `app.go:337`
**状态:** ✅ 已修复

**问题描述:**
`GetFilePatch` 期望相对路径，但传入的 `filePath` 可能是绝对路径。

**当前代码:**
```go
patch, err := diffEngine.GetFilePatch(filePath)
```

**修复方案:**
```go
relPath, _ := filepath.Rel(gitRoot, filePath)
patch, err := diffEngine.GetFilePatch(relPath)
```

---

### BUG #5: findRefBeforeTime 函数未完成实现
**文件:** `internal/diff/engine.go:203-213`
**状态:** ✅ 已修复

**问题描述:**
函数接收时间参数 `t` 但完全未使用，只是简单返回 reflog 的第一个 ref。

**修复方案:**
实现基于时间的 ref 查找逻辑。

---

### BUG #6: GetDiffWithActions 忽略 actions 参数
**文件:** `internal/diff/engine.go:149-172`
**状态:** ✅ 已修复

**问题描述:**
函数签名接收 `actions` 参数，但函数体中完全未使用。

**修复方案:**
移除未使用的参数，或实现参数预期的功能。

---

### BUG #7: GetDiffBetweenSession 获取 HEAD 顺序错误
**文件:** `internal/diff/engine.go:175-201`
**状态:** ✅ 已修复

**问题描述:**
逻辑顺序颠倒，应该先获取会话前的 ref，再获取当前 HEAD。

---

## P2 - 中优先级 BUG (Medium)

### BUG #8: Token 统计可能重复计算
**文件:** `internal/session/claude/reader.go:108-111`
**状态:** 待修复

**问题描述:**
如果同一条 assistant 消息的多行都包含 usage 数据，token 会被重复累加。

---

### BUG #9: parseNumstat 注释与实际输出不符
**文件:** `internal/diff/engine.go:226-228`
**状态:** 待修复

**问题描述:**
注释错误，git diff numstat 的实际输出格式与注释不符。

---

### BUG #10: 路径插入 HTML 属性存在 XSS 风险
**文件:** `frontend/app.js:271`
**状态:** 待修复

**问题描述:**
文件路径直接插入 `data-path` 属性，可能破坏 HTML 结构。

---

### BUG #11: updateLangToggle 缺少空值检查
**文件:** `frontend/app.js:141-147`
**状态:** 待修复

**问题描述:**
如果 `langToggle` 元素不存在，调用 `.querySelector` 会抛出异常。

---

## P3 - 低优先级问题 (Low)

### BUG #12: getSessionByID 性能问题
**文件:** `app.go:483-513`
**状态:** 待修复

**问题描述:**
每次调用都遍历所有项目目录和所有 JSONL 文件，性能问题。

---

### BUG #13: silentRefreshSessions 使用 JSON.stringify 比较
**文件:** `frontend/app.js:23`
**状态:** 待修复

**问题描述:**
JSON.stringify 比较效率低，应该使用更轻量的比较方式。

---

### BUG #14: joinStrings 函数可以使用 strings.Join
**文件:** `app.go:516-529`
**状态:** 待修复

**问题描述:**
标准库 `strings.Join` 已实现相同功能，无需重复实现。

---

### BUG #15: Scanner 缓冲区限制
**文件:** `internal/session/claude/reader.go:68`
**状态:** 待修复

**问题描述:**
如果单行 JSONL 超过 10MB，会被跳过。

---

## 修复进度

- [x] BUG #1 - 结构体 JSON 标签错误
- [x] BUG #2 - 风险评估结果未被使用
- [x] BUG #3 - HTML 模板引用不存在字段
- [x] BUG #4 - Git Diff 路径类型错误
- [x] BUG #5 - findRefBeforeTime 未完成实现
- [x] BUG #6 - GetDiffWithActions 忽略参数
- [x] BUG #7 - GetDiffBetweenSession 顺序错误
- [ ] BUG #8 - Token 统计重复计算
- [ ] BUG #9 - parseNumstat 注释错误
- [ ] BUG #10 - XSS 风险
- [ ] BUG #11 - 空值检查缺失
- [ ] BUG #12 - 性能问题
- [ ] BUG #13 - JSON 比较效率低
- [ ] BUG #14 - 重复实现
- [ ] BUG #15 - Scanner 缓冲区限制
