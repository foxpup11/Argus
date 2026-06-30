# Branch Protection Configuration

## 推荐设置（个人开发者友好）

### Main Branch Protection

**Settings → Branches → Add rule**

```
Branch name pattern: main
```

#### 基础设置（推荐）

| 选项 | 设置 | 理由 |
|------|------|------|
| ✅ Require pull request before merging | 0 approvals | 允许自己批准 |
| ❌ Require approvals from code owners | 不勾选 | 个人项目不需要 |
| ✅ Require status checks to pass | 空列表 | 允许直接推送 |
| ❌ Require conversation resolution | 不勾选 | 简化流程 |
| ❌ Require linear history | 不勾选 | 允许 merge commit |
| ❌ Require signed commits | 不勾选 | 个人项目不需要 |
| ❌ Require branches to be up to date | 不勾选 | 简化流程 |

#### Bypass 限制

```
Allow specified actors to bypass:
  ✅ sizhen (你的用户名)
```

这样你可以直接推送 main 分支，其他人需要 PR。

---

## GitHub Ruleset 设置（新功能）

**Settings → Rules → Rulesets → New ruleset → New branch ruleset**

### 配置

```
Name: main-protection
Enforcement: Active
Target: Branch
Branches: main
```

### Rules

```yaml
- type: pull_request
  parameters:
    required_approving_review_count: 0
    dismiss_stale_reviews_on_push: false
    require_code_owner_review: false

- type: required_status_checks
  parameters:
    required_status_checks: []
    strict_required_status_checks_policy: false

- type: non_fast_forward
  parameters: {}

- type: block_force_push
  parameters: {}
```

### Bypass Actors

```
Actor: sizhen
Type: User
Bypass mode: Always
```

---

## 快速检查清单

- [ ] 创建 Branch Protection Rule
- [ ] 设置 `Required approvals: 0`
- [ ] 添加自己到 Bypass 列表
- [ ] 测试：直接 push main 分支
- [ ] 测试：创建 PR 不需要审批

---

## 注意事项

1. **0 approvals** = 自己可以批准自己的 PR
2. **Bypass** = 完全绕过所有规则
3. **个人项目**建议保持简单
4. 如果以后团队协作，再增加限制
