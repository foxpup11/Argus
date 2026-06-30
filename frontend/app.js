// @ts-check

// 全局状态
let currentSessionId = null;
let currentFileIndex = 0;

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadSessions();
    setupEventListeners();
});

// 加载会话列表
async function loadSessions() {
    const sessionList = document.getElementById('sessionList');
    sessionList.innerHTML = '<div class="loading">加载中...</div>';

    try {
        const sessions = await window.go.main.App.GetSessions();
        renderSessionList(sessions);
    } catch (error) {
        sessionList.innerHTML = `<div class="loading">加载失败: ${error}</div>`;
    }
}

// 渲染会话列表
function renderSessionList(sessions) {
    const sessionList = document.getElementById('sessionList');

    if (!sessions || sessions.length === 0) {
        sessionList.innerHTML = `
            <div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                <p>暂无会话数据</p>
            </div>
        `;
        return;
    }

    sessionList.innerHTML = sessions.map(session => `
        <div class="session-item" data-id="${session.id}">
            <div class="session-id">${session.id.substring(0, 8)}</div>
            <div class="session-model">${session.model || '-'}</div>
            <div class="session-prompt">${session.prompt || '-'}</div>
            <div class="session-meta">
                <span>文件 ${session.fileCount}</span>
                <span>操作 ${session.actionCount}</span>
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
    document.getElementById('statusSession').textContent = `会话: ${detail.id.substring(0, 8)}`;
    document.getElementById('statusBranch').textContent = `分支: ${detail.branch || '-'}`;
    document.getElementById('statusTokens').textContent = `Token: ${formatNumber(detail.tokenUsage.inputTokens)} in / ${formatNumber(detail.tokenUsage.outputTokens)} out`;

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
                    <p>暂无文件改动</p>
                </td>
            </tr>
        `;
        fileCount.textContent = '0 个文件';
        return;
    }

    fileCount.textContent = `${detail.fileChanges.length} 个文件`;

    fileTableBody.innerHTML = detail.fileChanges.map((file, index) => `
        <tr data-index="${index}" data-path="${file.path}">
            <td><span class="risk-badge risk-${(file.risk || 'review').toLowerCase()}">${getRiskLabel(file.risk)}</span></td>
            <td title="${file.path}">${truncatePath(file.path)}</td>
            <td><span class="change-badge change-${(file.changeType || 'modified').toLowerCase()}">${getChangeTypeLabel(file.changeType)}</span></td>
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
        document.getElementById('diffView').innerHTML = `<code>加载 diff 失败: ${error}</code>`;
    }
}

// 渲染 diff
function renderDiff(diff, filePath) {
    const diffView = document.getElementById('diffView');

    if (!diff) {
        diffView.innerHTML = '<code>暂无 diff 数据</code>';
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

// 获取风险标签
function getRiskLabel(risk) {
    switch (risk) {
        case 'Safe': return 'Safe';
        case 'Review': return 'Review';
        case 'Danger': return 'Danger';
        default: return risk;
    }
}

// 获取变更类型标签
function getChangeTypeLabel(changeType) {
    switch (changeType) {
        case 'Created': return '新增';
        case 'Modified': return '修改';
        case 'Deleted': return '删除';
        default: return changeType;
    }
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

    // 搜索框
    document.getElementById('searchInput').addEventListener('input', (e) => {
        const keyword = e.target.value.toLowerCase();
        document.querySelectorAll('.session-item').forEach(item => {
            const text = item.textContent.toLowerCase();
            item.style.display = text.includes(keyword) ? '' : 'none';
        });
    });
}
