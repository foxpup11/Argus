// @ts-check

// 全局状态
let currentSessionId = null;
let currentDiffMode = 'uncommitted'; // 默认：未提交改动
let sessions = [];

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadSessions();
    setupEventListeners();
    initI18n();
    initMonitoring();
    initTheme(); // 初始化主题

    // 每 2 分钟后台静默刷新会话列表
    setInterval(silentRefreshSessions, 120000);
});

// 静默刷新（不显示加载状态，不打扰用户）
async function silentRefreshSessions() {
    try {
        const newSessions = await window.go.main.App.GetSessions();
        // 使用轻量级比较：检查数量和最新会话 ID
        if (sessionsChanged(sessions, newSessions)) {
            sessions = newSessions;
            renderSessionList(sessions);
        }
    } catch (error) {
        // 静默失败，不打扰用户
    }
}

// 检查会话列表是否发生变化
function sessionsChanged(oldSessions, newSessions) {
    // 数量不同，肯定变化了
    if (oldSessions.length !== newSessions.length) {
        return true;
    }
    // 空列表
    if (oldSessions.length === 0) {
        return false;
    }
    // 比较第一个和最后一个会话的 ID（假设按时间排序）
    return oldSessions[0].id !== newSessions[0].id ||
           oldSessions[oldSessions.length - 1].id !== newSessions[newSessions.length - 1].id;
}

// 初始化国际化
function initI18n() {
    updateLangToggle();
    updateUI();
}

// 初始化主题
async function initTheme() {
    try {
        const settings = await window.go.main.App.GetSettings();
        if (settings && settings.theme) {
            applyTheme(settings.theme);
        }
    } catch (error) {
        // 静默失败，使用默认主题
    }
}

// 监控状态
let isMonitoring = false;

// 初始化监控
async function initMonitoring() {
    // 检查监控状态
    try {
        isMonitoring = await window.go.main.App.IsMonitoring();
        updateMonitoringUI();
    } catch (error) {
        console.error('检查监控状态失败:', error);
    }

    // 监听会话更新事件
    window.runtime.EventsOn('session-updated', () => {
        silentRefreshSessions();
        showToast(t('sessionUpdated'));
    });
}

// 切换监控状态
async function toggleMonitoring() {
    try {
        if (isMonitoring) {
            await window.go.main.App.StopMonitoring();
            isMonitoring = false;
            showToast(t('monitoringStopped'));
        } else {
            const started = await window.go.main.App.StartMonitoring();
            if (started) {
                isMonitoring = true;
                showToast(t('monitoringStarted'));
            }
        }
        updateMonitoringUI();
    } catch (error) {
        console.error('切换监控状态失败:', error);
        showToast(t('monitoringFailed'));
    }
}

// 更新监控 UI
function updateMonitoringUI() {
    const indicator = document.getElementById('monitorIndicator');
    if (indicator) {
        indicator.classList.toggle('active', isMonitoring);
        indicator.title = isMonitoring ? t('monitoringActive') : t('monitoringInactive');
    }
}

// 显示提示消息
function showToast(message) {
    // 创建 toast 元素
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    document.body.appendChild(toast);

    // 显示动画
    setTimeout(() => toast.classList.add('show'), 10);

    // 3秒后移除
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// 导出功能
async function exportSession() {
    if (!currentSessionId) {
        showToast(t('selectSessionFirst'));
        return;
    }

    try {
        // 直接弹出目录选择对话框
        const path = await window.go.main.App.SelectDirectory();
        if (!path) {
            // 用户取消了选择
            return;
        }

        showToast(t('exporting'));

        // 调用导出 API，传递选择的路径
        const result = await window.go.main.App.ExportSession(
            currentSessionId,
            'markdown',
            path
        );

        if (result && result.filePath) {
            showToast(t('exportSuccess'));
        }
    } catch (error) {
        console.error('导出会话失败:', error);
        showToast(t('exportFailed') + ': ' + error);
    }
}

// 更新语言切换按钮
function updateLangToggle() {
    const toggle = document.getElementById('langToggle');
    if (!toggle) return;
    const knob = toggle.querySelector('.toggle-knob');
    if (!knob) return;
    const isEn = getCurrentLang() === 'en';

    toggle.classList.toggle('active', isEn);
    knob.textContent = isEn ? '中' : 'EN';
}

// 更新所有 UI 文本
function updateUI() {
    updateLangToggle();

    // 更新 data-i18n 元素
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (key) {
            el.textContent = t(key);
        }
    });

    // 更新 placeholder
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        if (key) {
            el.placeholder = t(key);
        }
    });

    // 重新渲染会话列表
    if (sessions.length > 0) {
        renderSessionList(sessions);
    }
}

// 加载会话列表
async function loadSessions() {
    const sessionList = document.getElementById('sessionList');
    sessionList.innerHTML = `<div class="loading">${t('loading')}</div>`;

    try {
        sessions = await window.go.main.App.GetSessions();
        renderSessionList(sessions);
    } catch (error) {
        sessionList.innerHTML = `<div class="loading">${t('loadFailed')}: ${error}</div>`;
    }
}

// 渲染会话列表
function renderSessionList(sessionData) {
    const sessionList = document.getElementById('sessionList');

    if (!sessionData || sessionData.length === 0) {
        sessionList.innerHTML = `
            <div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                <p>${t('noSessions')}</p>
            </div>
        `;
        return;
    }

    sessionList.innerHTML = sessionData.map(session => `
        <div class="session-item" data-id="${session.id}">
            <div class="session-id">${session.id.substring(0, 8)}</div>
            <div class="session-model">${session.model || '-'}</div>
            <div class="session-prompt">${session.prompt || '-'}</div>
            <div class="session-meta">
                <span>${t('files')} ${session.fileCount}</span>
                <span>${t('actions')} ${session.actionCount}</span>
            </div>
        </div>
    `).join('');

    // 绑定点击事件
    document.querySelectorAll('.session-item').forEach(item => {
        item.addEventListener('click', () => {
            const sessionId = item.getAttribute('data-id');
            selectSession(sessionId);
        });
    });
}

// 选择会话
async function selectSession(sessionId) {
    currentSessionId = sessionId;

    // 更新选中状态
    document.querySelectorAll('.session-item').forEach(item => {
        item.classList.toggle('active', item.getAttribute('data-id') === sessionId);
    });

    try {
        const detail = await window.go.main.App.GetSession(sessionId);
        renderSessionDetail(detail);
    } catch (error) {
        console.error('加载会话详情失败:', error);
    }
}

// 渲染会话详情
function renderSessionDetail(detail) {
    // 更新状态栏
    document.getElementById('statusSession').innerHTML = `${t('session')}: ${detail.id.substring(0, 8)}`;
    document.getElementById('statusBranch').innerHTML = `${t('branch')}: ${detail.branch || '-'}`;
    document.getElementById('statusTokens').innerHTML = `Token: ${formatNumber(detail.tokenUsage.inputTokens)} ${t('tokenIn')} / ${formatNumber(detail.tokenUsage.outputTokens)} ${t('tokenOut')}`;

    // 更新文件表格
    const fileTableBody = document.getElementById('fileTableBody');
    const fileCount = document.getElementById('fileCount');

    if (!detail.fileChanges || detail.fileChanges.length === 0) {
        fileTableBody.innerHTML = `
            <tr>
                <td colspan="4" class="empty-state">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                        <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                    </svg>
                    <p>${t('noChanges')}</p>
                </td>
            </tr>
        `;
        fileCount.textContent = '0';
        return;
    }

    fileCount.textContent = detail.fileChanges.length;

    fileTableBody.innerHTML = detail.fileChanges.map((file, index) => `
        <tr data-index="${index}" data-path="${escapeHtmlAttr(file.path)}">
            <td><span class="risk-badge risk-${(file.risk || 'review').toLowerCase()}">${t((file.risk || 'review').toLowerCase())}</span></td>
            <td title="${escapeHtmlAttr(file.path)}">${truncatePath(file.path)}</td>
            <td><span class="change-badge change-${(file.changeType || 'modified').toLowerCase()}">${t((file.changeType || 'modified').toLowerCase())}</span></td>
            <td>${file.actionCount || 0}</td>
        </tr>
    `).join('');

    // 绑定点击事件
    document.querySelectorAll('#fileTableBody tr').forEach(row => {
        row.addEventListener('click', () => {
            const path = row.getAttribute('data-path');
            selectFile(path);
        });
    });

    // 默认选中第一个文件
    if (detail.fileChanges.length > 0) {
        selectFile(detail.fileChanges[0].path);
    }
}

// 选择文件
async function selectFile(filePath) {
    // 更新选中状态
    document.querySelectorAll('#fileTableBody tr').forEach(row => {
        row.classList.toggle('active', row.getAttribute('data-path') === filePath);
    });

    document.getElementById('diffFileName').textContent = filePath;

    try {
        let diff = '';

        if (currentDiffMode === 'session') {
            // 会话前后对比模式：获取整个会话的 diff，然后找到指定文件
            const diffInfo = await window.go.main.App.GetSessionDiff(currentSessionId, 'session');
            if (diffInfo && diffInfo.diffs) {
                const fileDiff = diffInfo.diffs.find(d => d.filePath === filePath);
                if (fileDiff) {
                    diff = fileDiff.patch;
                }
            }
        } else {
            // 默认模式：获取单个文件的未提交改动
            diff = await window.go.main.App.GetDiff(currentSessionId, filePath);
        }

        renderDiff(diff, filePath);
    } catch (error) {
        document.getElementById('diffView').innerHTML = `<code>${t('loadFailed')}: ${error}</code>`;
    }
}

// 渲染 diff
function renderDiff(diff, filePath) {
    const diffView = document.getElementById('diffView');

    if (!diff) {
        diffView.innerHTML = `<code>${t('noDiff')}</code>`;
        return;
    }

    // 高亮 diff
    const highlighted = highlightDiff(diff);
    diffView.innerHTML = `<code>${highlighted}</code>`;
}

// Diff 高亮
function highlightDiff(diff) {
    return diff.split('\n').map(line => {
        if (line.startsWith('+')) {
            return `<span class="diff-add">${escapeHtml(line)}</span>`;
        } else if (line.startsWith('-')) {
            return `<span class="diff-remove">${escapeHtml(line)}</span>`;
        } else if (line.startsWith('@@') || line.startsWith('diff') || line.startsWith('index')) {
            return `<span class="diff-header-line">${escapeHtml(line)}</span>`;
        }
        return escapeHtml(line);
    }).join('\n');
}

// HTML 转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// HTML 属性转义（防止 XSS）
function escapeHtmlAttr(text) {
    return text
        .replace(/&/g, '&amp;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}

// 截断路径
function truncatePath(path) {
    if (path.length > 40) {
        return '...' + path.slice(-37);
    }
    return path;
}

// 格式化数字
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
}

// 设置事件监听
function setupEventListeners() {
    // 刷新按钮
    document.getElementById('refreshBtn').addEventListener('click', loadSessions);

    // 语言切换按钮
    document.getElementById('langToggle').addEventListener('click', () => {
        switchLang();
    });

    // 搜索框
    document.getElementById('searchInput').addEventListener('input', (e) => {
        const keyword = e.target.value.toLowerCase();
        document.querySelectorAll('.session-item').forEach(item => {
            const text = item.textContent.toLowerCase();
            item.style.display = text.includes(keyword) ? '' : 'none';
        });
    });

    // Diff 模式切换按钮
    document.querySelectorAll('.diff-mode-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const mode = btn.getAttribute('data-mode');
            switchDiffMode(mode);
        });
    });
}

// 切换 diff 对比模式
function switchDiffMode(mode) {
    currentDiffMode = mode;

    // 更新按钮状态
    document.querySelectorAll('.diff-mode-btn').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-mode') === mode);
    });

    // 如果有选中的文件，重新加载 diff
    const diffFileName = document.getElementById('diffFileName');
    if (diffFileName && diffFileName.textContent && currentSessionId) {
        selectFile(diffFileName.textContent);
    }
}

// ============================================
// Settings Panel
// ============================================

let currentSettings = null;

// 打开设置面板
async function openSettings() {
    const modal = document.getElementById('settingsModal');
    modal.style.display = 'flex';
    setTimeout(() => modal.classList.add('show'), 10);

    // 加载设置
    try {
        currentSettings = await window.go.main.App.GetSettings();
        renderSettings();
    } catch (error) {
        console.error('加载设置失败:', error);
    }
}

// 关闭设置面板
function closeSettings() {
    const modal = document.getElementById('settingsModal');
    modal.classList.remove('show');
    setTimeout(() => modal.style.display = 'none', 200);
}

// 渲染设置
function renderSettings() {
    if (!currentSettings) return;

    // 更新主题按钮状态
    document.querySelectorAll('.theme-btn').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-theme') === currentSettings.theme);
    });

    // 渲染自定义规则
    renderCustomRules();
}

// 设置主题
async function setTheme(theme) {
    try {
        await window.go.main.App.UpdateTheme(theme);
        currentSettings.theme = theme;

        // 更新按钮状态
        document.querySelectorAll('.theme-btn').forEach(btn => {
            btn.classList.toggle('active', btn.getAttribute('data-theme') === theme);
        });

        // 应用主题
        applyTheme(theme);
        showToast(t('theme') + ': ' + t('theme' + theme.charAt(0).toUpperCase() + theme.slice(1)));
    } catch (error) {
        console.error('设置主题失败:', error);
    }
}

// 应用主题
function applyTheme(theme) {
    if (theme === 'auto') {
        // 跟随系统
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        document.documentElement.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
    } else {
        document.documentElement.setAttribute('data-theme', theme);
    }
}

// 渲染自定义规则
function renderCustomRules() {
    const container = document.getElementById('customRulesList');
    if (!container || !currentSettings) return;

    if (currentSettings.customRules.length === 0) {
        container.innerHTML = `<p style="color: var(--text-tertiary); font-size: 13px;">${t('noRules') || '暂无自定义规则'}</p>`;
        return;
    }

    container.innerHTML = currentSettings.customRules.map((rule, index) => `
        <div class="rule-item">
            <div class="rule-info">
                <div class="rule-name">${escapeHtml(rule.name)}</div>
                <div class="rule-desc">${escapeHtml(rule.description)}</div>
            </div>
            <span class="rule-level ${(rule.level || 'review').toLowerCase()}">${rule.level || 'Review'}</span>
            <button class="rule-delete" data-rule-index="${index}" title="${t('deleteRule')}">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
            </button>
        </div>
    `).join('');

    // 使用事件委托绑定删除按钮事件
    container.querySelectorAll('.rule-delete').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const index = parseInt(btn.getAttribute('data-rule-index'));
            if (!isNaN(index) && currentSettings.customRules[index]) {
                deleteRule(currentSettings.customRules[index].name);
            }
        });
    });
}

// 显示添加规则表单（简化版：使用 prompt）
function showAddRuleForm() {
    const name = prompt(t('ruleNamePlaceholder') || '规则名称:');
    if (!name) return;

    const description = prompt(t('ruleDescPlaceholder') || '规则描述:');
    if (!description) return;

    const level = prompt(t('ruleLevelPlaceholder') || '风险等级 (safe/review/danger):', 'review');
    if (!level || !['safe', 'review', 'danger'].includes(level.toLowerCase())) {
        showToast(t('invalidLevel') || '无效的风险等级');
        return;
    }

    const pattern = prompt(t('rulePatternPlaceholder') || '文件路径匹配模式:');
    if (!pattern) return;

    addRule(name, description, level.toLowerCase(), pattern);
}

// 添加规则
async function addRule(name, description, level, pattern) {
    try {
        await window.go.main.App.AddCustomRule(name, description, level, pattern);
        currentSettings.customRules.push({ name, description, level, pattern, enabled: true });
        renderCustomRules();
        showToast(t('ruleAdded') || '规则已添加');
    } catch (error) {
        console.error('添加规则失败:', error);
    }
}

// 删除规则
async function deleteRule(name) {
    if (!confirm(t('confirmDelete') || '确定要删除这个规则吗？')) {
        return;
    }

    try {
        await window.go.main.App.RemoveCustomRule(name);
        currentSettings.customRules = currentSettings.customRules.filter(r => r.name !== name);
        renderCustomRules();
        showToast(t('ruleDeleted') || '规则已删除');
    } catch (error) {
        console.error('删除规则失败:', error);
    }
}

// 点击模态框背景关闭
document.addEventListener('click', (e) => {
    if (e.target.id === 'settingsModal') {
        closeSettings();
    }
});
