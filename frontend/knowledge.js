// @ts-check
// Knowledge Tab — 知识库页面逻辑

// 知识库状态
let currentKnowledgeDocs = [];
let currentKnowledgeDoc = null;
let currentKnowledgeType = 'all';
let isKnowledgeEditing = false;

// ============================================
// Initialization
// ============================================

async function loadKnowledgeDocuments(type = 'all', project = '') {
    try {
        const docs = await window.go.main.App.GetKnowledgeDocuments(type, project);
        currentKnowledgeDocs = docs;
        renderKnowledgeDocList(docs);
    } catch (error) {
        console.error('Failed to load knowledge documents:', error);
        showToast(t('loadFailed') || '加载失败');
    }
}

// ============================================
// Document List
// ============================================

function renderKnowledgeDocList(docs) {
    const container = document.getElementById('knowledgeDocList');
    if (!container) return;

    if (!docs || docs.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <p>${t('noDocuments') || '暂无文档'}</p>
            </div>
        `;
        return;
    }

    // 按项目分组
    const groups = {};
    docs.forEach(doc => {
        const project = doc.project || 'plans';
        if (!groups[project]) {
            groups[project] = [];
        }
        groups[project].push(doc);
    });

    // 渲染分组列表
    let html = '';
    for (const [project, projectDocs] of Object.entries(groups)) {
        const displayName = formatProjectName(project);
        html += `
            <div class="knowledge-group">
                <div class="group-header" onclick="toggleKnowledgeGroup(this)">
                    <svg class="group-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M6 9l6 6 6-6"/>
                    </svg>
                    <span class="group-name">${escapeHtml(displayName)}</span>
                    <span class="group-count">${projectDocs.length}</span>
                </div>
                <div class="group-items">
                    ${projectDocs.map(doc => `
                        <div class="knowledge-doc-item ${currentKnowledgeDoc?.path === doc.path ? 'active' : ''}"
                             data-path="${escapeHtmlAttr(doc.path)}"
                             onclick="selectKnowledgeDoc(this.dataset.path)">
                            <div class="doc-icon">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    ${doc.type === 'plans'
                                        ? '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>'
                                        : '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>'}
                                </svg>
                            </div>
                            <div class="doc-info">
                                <div class="doc-title">${escapeHtml(doc.name)}</div>
                                <div class="doc-meta">
                                    <span class="doc-type">${doc.type === 'plans' ? 'Plan' : 'Memory'}</span>
                                    <span class="doc-time">${formatKnowledgeTime(doc.updatedAt)}</span>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    container.innerHTML = html;

    // 默认折叠所有分组
    container.querySelectorAll('.knowledge-group').forEach(group => {
        group.classList.add('collapsed');
    });
}

// 格式化项目名称
function formatProjectName(dirName) {
    if (dirName === 'plans') return 'Plans';
    // 将类似 "g--ltch-git-learn-agentscope-desktop" 转换为 "agentscope-desktop"
    const parts = dirName.split('-').filter(p => p.length > 0);
    if (parts.length >= 2) {
        return parts.slice(-2).join('-');
    }
    return dirName;
}

// 切换分组展开/折叠
function toggleKnowledgeGroup(header) {
    const group = header.closest('.knowledge-group');
    if (group) {
        group.classList.toggle('collapsed');
    }
}

// ============================================
// Document Selection
// ============================================

async function selectKnowledgeDoc(path) {
    try {
        const doc = await window.go.main.App.GetKnowledgeDocument(path);
        currentKnowledgeDoc = doc;

        // 更新列表高亮（使用 data-path 属性精确匹配）
        document.querySelectorAll('.knowledge-doc-item').forEach(item => {
            const itemPath = item.getAttribute('data-path');
            item.classList.toggle('active', itemPath === path);
        });

        // 更新工具栏
        const docName = document.getElementById('knowledgeDocName');
        const docType = document.getElementById('knowledgeDocType');
        if (docName) docName.textContent = doc.name;
        if (docType) {
            docType.textContent = doc.type === 'plans' ? 'Plan' : 'Memory';
            docType.className = `doc-type-badge ${doc.type}`;
        }

        // 显示预览
        renderKnowledgePreview(doc.content);

        // 重置编辑状态
        exitKnowledgeEdit();
    } catch (error) {
        console.error('Failed to load document:', error);
    }
}

// ============================================
// Preview / Editor
// ============================================

function renderKnowledgePreview(content) {
    const preview = document.getElementById('knowledgePreview');
    if (!preview) return;

    if (!content) {
        preview.innerHTML = `
            <div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                </svg>
                <p>${t('selectDocument') || '选择文档查看内容'}</p>
            </div>
        `;
        return;
    }

    // 解析 frontmatter 并提取 body
    const { frontmatter, body } = parseFrontmatter(content);

    // 构建预览 HTML
    let html = '';

    // 如果有 frontmatter，显示为元数据卡片
    if (frontmatter && Object.keys(frontmatter).length > 0) {
        html += '<div class="frontmatter-card">';
        for (const [key, value] of Object.entries(frontmatter)) {
            if (key !== 'metadata' && value) {
                html += `<div class="frontmatter-item"><span class="frontmatter-key">${escapeHtml(key)}:</span> <span class="frontmatter-value">${escapeHtml(value)}</span></div>`;
            }
        }
        html += '</div>';
    }

    // 渲染 Markdown body
    html += '<div class="markdown-content">';
    html += renderMarkdown(body);
    html += '</div>';

    preview.innerHTML = html;
}

// 解析 YAML frontmatter
function parseFrontmatter(content) {
    const frontmatter = {};
    let body = content;

    // 检查是否以 --- 开头
    if (!content.startsWith('---')) {
        return { frontmatter, body };
    }

    // 查找结束标记
    const endIndex = content.indexOf('---', 3);
    if (endIndex === -1) {
        return { frontmatter, body };
    }

    // 提取 frontmatter 部分
    const fmContent = content.substring(3, endIndex);
    body = content.substring(endIndex + 3).trim();

    // 简单解析 YAML（支持 key: value 格式）
    const lines = fmContent.split('\n');
    let currentKey = '';
    let currentValue = '';

    for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed === '' || trimmed.startsWith('#')) {
            continue;
        }

        // 检查是否是新的 key: value 对
        const colonIndex = trimmed.indexOf(':');
        if (colonIndex > 0) {
            // 保存上一个 key-value
            if (currentKey) {
                frontmatter[currentKey] = currentValue.trim();
            }
            currentKey = trimmed.substring(0, colonIndex).trim();
            currentValue = trimmed.substring(colonIndex + 1).trim();
        } else if (currentKey) {
            // 继续上一个 value（多行值）
            currentValue += ' ' + trimmed;
        }
    }

    // 保存最后一个 key-value
    if (currentKey) {
        frontmatter[currentKey] = currentValue.trim();
    }

    return { frontmatter, body };
}

function renderMarkdown(content) {
    if (!content) return '';

    // 先转义 HTML 特殊字符，防止 XSS
    let escaped = content
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');

    // 按行处理 Markdown
    const lines = escaped.split('\n');
    let html = '';
    let inList = false;
    let inCodeBlock = false;
    let codeBlockContent = '';

    for (let i = 0; i < lines.length; i++) {
        let line = lines[i];

        // 处理代码块
        if (line.trim().startsWith('```')) {
            if (inCodeBlock) {
                html += `<pre><code>${codeBlockContent}</code></pre>`;
                codeBlockContent = '';
                inCodeBlock = false;
            } else {
                inCodeBlock = true;
            }
            continue;
        }

        if (inCodeBlock) {
            codeBlockContent += (codeBlockContent ? '\n' : '') + line;
            continue;
        }

        // 关闭列表
        if (inList && !line.match(/^\s*[-*]\s/)) {
            html += '</ul>';
            inList = false;
        }

        // 处理标题
        if (line.match(/^#### /)) {
            html += `<h4>${line.substring(5)}</h4>`;
            continue;
        }
        if (line.match(/^### /)) {
            html += `<h3>${line.substring(4)}</h3>`;
            continue;
        }
        if (line.match(/^## /)) {
            html += `<h2>${line.substring(3)}</h2>`;
            continue;
        }
        if (line.match(/^# /)) {
            html += `<h1>${line.substring(2)}</h1>`;
            continue;
        }

        // 处理水平线
        if (line.match(/^---+$/)) {
            html += '<hr>';
            continue;
        }

        // 处理列表项
        if (line.match(/^\s*[-*]\s/)) {
            if (!inList) {
                html += '<ul>';
                inList = true;
            }
            const content = line.replace(/^\s*[-*]\s/, '');
            html += `<li>${processInlineMarkdown(content)}</li>`;
            continue;
        }

        // 处理空行
        if (line.trim() === '') {
            html += '<br>';
            continue;
        }

        // 处理普通段落
        html += `<p>${processInlineMarkdown(line)}</p>`;
    }

    // 关闭未关闭的列表
    if (inList) {
        html += '</ul>';
    }

    return html;
}

// 处理行内 Markdown（粗体、斜体、代码、链接）
function processInlineMarkdown(text) {
    return text
        // 行内代码（先处理，避免被其他规则影响）
        .replace(/`(.*?)`/g, '<code>$1</code>')
        // 粗体
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        // 斜体
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        // 链接
        .replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2" target="_blank">$1</a>');
}

function toggleKnowledgeEdit() {
    if (!currentKnowledgeDoc) return;

    isKnowledgeEditing = true;
    const preview = document.getElementById('knowledgePreview');
    const editor = document.getElementById('knowledgeEditor');
    const editBtn = document.getElementById('knowledgeEditBtn');
    const saveBtn = document.getElementById('knowledgeSaveBtn');
    const textarea = document.getElementById('knowledgeEditorContent');

    if (preview) preview.style.display = 'none';
    if (editor) editor.style.display = 'flex';
    if (editBtn) editBtn.style.display = 'none';
    if (saveBtn) saveBtn.style.display = 'inline-flex';
    if (textarea) textarea.value = currentKnowledgeDoc.content;
}

function exitKnowledgeEdit() {
    isKnowledgeEditing = false;
    const preview = document.getElementById('knowledgePreview');
    const editor = document.getElementById('knowledgeEditor');
    const editBtn = document.getElementById('knowledgeEditBtn');
    const saveBtn = document.getElementById('knowledgeSaveBtn');

    if (preview) preview.style.display = 'block';
    if (editor) editor.style.display = 'none';
    if (editBtn) editBtn.style.display = 'inline-flex';
    if (saveBtn) saveBtn.style.display = 'none';
}

// ============================================
// Document Operations
// ============================================

async function saveKnowledgeDoc() {
    if (!currentKnowledgeDoc) return;

    const textarea = document.getElementById('knowledgeEditorContent');
    if (!textarea) return;

    const content = textarea.value;
    try {
        await window.go.main.App.SaveKnowledgeDocument(currentKnowledgeDoc.path, content);
        currentKnowledgeDoc.content = content;
        exitKnowledgeEdit();
        renderKnowledgePreview(content);
        showToast(t('documentSaved') || '文档已保存');
    } catch (error) {
        console.error('Failed to save document:', error);
        showToast(t('saveFailed') || '保存失败');
    }
}

async function deleteKnowledgeDoc() {
    if (!currentKnowledgeDoc) return;

    const confirmMsg = t('confirmDeleteDocument') || '确定要删除这个文档吗？';
    if (!confirm(confirmMsg)) return;

    try {
        await window.go.main.App.DeleteKnowledgeDocument(currentKnowledgeDoc.path);
        currentKnowledgeDoc = null;
        loadKnowledgeDocuments(currentKnowledgeType);
        showToast(t('documentDeleted') || '文档已删除');

        // 清空预览区
        const preview = document.getElementById('knowledgePreview');
        if (preview) {
            preview.innerHTML = `
                <div class="empty-state">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                        <polyline points="14 2 14 8 20 8"/>
                        <line x1="16" y1="13" x2="8" y2="13"/>
                        <line x1="16" y1="17" x2="8" y2="17"/>
                    </svg>
                    <p>${t('selectDocument') || '选择文档查看内容'}</p>
                </div>
            `;
        }

        // 清空工具栏
        const docName = document.getElementById('knowledgeDocName');
        const docType = document.getElementById('knowledgeDocType');
        if (docName) docName.textContent = t('selectDocument') || '选择文档查看';
        if (docType) docType.textContent = '';
    } catch (error) {
        console.error('Failed to delete document:', error);
        showToast(t('deleteFailed') || '删除失败');
    }
}

async function createNewDocument() {
    const title = prompt(t('enterDocumentTitle') || '请输入文档标题:');
    if (!title) return;

    try {
        const path = await window.go.main.App.CreateKnowledgeDocument('plans', title, '', '');
        await loadKnowledgeDocuments(currentKnowledgeType);
        await selectKnowledgeDoc(path);
        toggleKnowledgeEdit();
    } catch (error) {
        console.error('Failed to create document:', error);
        showToast(t('createFailed') || '创建失败');
    }
}

// ============================================
// Filtering & Search
// ============================================

function filterByType(type) {
    currentKnowledgeType = type;
    document.querySelectorAll('.knowledge-filters .filter-btn').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-type') === type);
    });
    loadKnowledgeDocuments(type);
}

let knowledgeSearchTimeout;
function searchKnowledge() {
    clearTimeout(knowledgeSearchTimeout);
    knowledgeSearchTimeout = setTimeout(async () => {
        const input = document.getElementById('knowledgeSearchInput');
        if (!input) return;

        const query = input.value;
        try {
            const docs = await window.go.main.App.SearchKnowledgeDocuments(query, [], []);
            renderKnowledgeDocList(docs);
        } catch (error) {
            console.error('Failed to search:', error);
        }
    }, 300);
}

// ============================================
// Utility Functions
// ============================================

function formatKnowledgeTime(timeStr) {
    if (!timeStr) return '';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now - date;

    // 小于 1 小时
    if (diff < 3600000) {
        const minutes = Math.floor(diff / 60000);
        return `${minutes || 1} ${t('minutesAgo') || '分钟前'}`;
    }

    // 小于 24 小时
    if (diff < 86400000) {
        const hours = Math.floor(diff / 3600000);
        return `${hours} ${t('hoursAgo') || '小时前'}`;
    }

    // 小于 7 天
    if (diff < 604800000) {
        const days = Math.floor(diff / 86400000);
        return `${days} ${t('daysAgo') || '天前'}`;
    }

    // 超过 7 天，显示日期
    return date.toLocaleDateString();
}
