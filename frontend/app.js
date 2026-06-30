// @ts-check

// 全局状态
let currentSessionId = null;
let sessions = [];

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadSessions();
    setupEventListeners();
    initI18n();
    initMonitoring();

    // 每 2 分钟后台静默刷新会话列表
    setInterval(silentRefreshSessions, 120000);
});

// 静默刷新（不显示加载状态，不打扰用户）
async function silentRefreshSessions() {
    try {
        const newSessions = await window.go.main.App.GetSessions();
        // 只在数据有变化时更新
        if (JSON.stringify(newSessions) !== JSON.stringify(sessions)) {
            sessions = newSessions;
            renderSessionList(sessions);
        }
    } catch (error) {
        // 静默失败，不打扰用户
    }
}

// 初始化国际化
function initI18n() {
    updateLangToggle();
    updateUI();
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
    const knob = toggle.querySelector('.toggle-knob');
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
        <tr data-index="${index}" data-path="${file.path}">
            <td><span class="risk-badge risk-${(file.risk || 'review').toLowerCase()}">${t((file.risk || 'review').toLowerCase())}</span></td>
            <td title="${file.path}">${truncatePath(file.path)}</td>
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
        const diff = await window.go.main.App.GetDiff(currentSessionId, filePath);
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
}
