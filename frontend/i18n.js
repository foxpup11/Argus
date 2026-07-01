// 国际化支持
const i18n = {
    zh: {
        // 标题栏
        appTitle: 'AgentScope',
        refresh: '刷新',
        monitor: '监控',

        // Tab 导航
        dashboard: 'Token 概览',
        sessionsTab: '会话管理',

        // 监控状态
        monitoringActive: '实时监控已开启',
        monitoringInactive: '实时监控已关闭',
        monitoringStarted: '已开启实时监控',
        monitoringStopped: '已关闭实时监控',
        monitoringFailed: '监控启动失败',
        sessionUpdated: '检测到新会话',

        // 导出功能
        export: '导出',
        exporting: '正在导出...',
        exportSuccess: '导出成功',
        exportFailed: '导出失败',
        selectSessionFirst: '请先选择一个会话',

        // 侧边栏
        sessions: '会话',
        searchPlaceholder: '搜索...',
        files: '文件',
        actions: '操作',

        // 文件表格
        fileChanges: '文件改动',
        risk: 'Risk',
        file: 'File',
        change: 'Change',
        ops: 'Ops',

        // 风险等级
        safe: 'Safe',
        review: 'Review',
        danger: 'Danger',

        // 变更类型
        created: '新增',
        modified: '修改',
        deleted: '删除',

        // Diff 视图
        diff: 'Diff',
        selectFile: '选择文件查看差异',
        noDiff: '暂无 diff 数据',
        diffModeUncommitted: '未提交改动',
        diffModeSession: '会话对比',

        // 状态栏
        session: '会话',
        branch: '分支',
        tokenIn: 'in',
        tokenOut: 'out',

        // 空状态
        selectSession: '选择会话查看改动',
        noChanges: '暂无文件改动',
        loading: '加载中...',
        loadFailed: '加载失败',
        noSessions: '暂无会话数据',

        // 对话记录
        conversation: '对话记录',
        userMessage: '用户',
        aiMessage: 'AI',
        toolMessage: '工具',
        thinking: '思考过程',
        toolCall: '工具调用',
        toolResult: '工具结果',
        copyCommand: '复制命令',
        commandCopied: '已复制',
        loadingConversation: '加载对话记录...',
        noConversation: '暂无对话记录',

        // 语言切换
        langSwitch: 'EN',

        // 设置面板
        settings: '设置',
        theme: '主题',
        themeLight: '浅色',
        themeDark: '深色',
        themeAuto: '跟随系统',
        customRules: '自定义风险规则',
        addRule: '添加规则',
        deleteRule: '删除',
        noRules: '暂无自定义规则',
        ruleNamePlaceholder: '规则名称:',
        ruleDescPlaceholder: '规则描述:',
        ruleLevelPlaceholder: '风险等级 (safe/review/danger):',
        rulePatternPlaceholder: '文件路径匹配模式:',
        invalidLevel: '无效的风险等级',
        ruleAdded: '规则已添加',
        confirmDelete: '确定要删除这个规则吗？',
        ruleDeleted: '规则已删除',

        // Dashboard — 仪表盘
        todayTokens: '今日 Token',
        thisMonthTokens: '本月 Token',
        lastMonthTokens: '上月 Token',
        totalTokensAll: '累计 Token',
        totalSessions: '总会话数',
        vsLastMonth: '较上月',
        tokens: 'Token',
        tokenTrend: 'Token 趋势（近30天）',
        projectBreakdown: '项目 Token 分布',
        modelDistribution: '模型分布',
        project: '项目',
        model: '模型',
        projectSessions: '会话数',
        noData: '暂无数据',

        // 会话管理增强
        // 高级搜索
        advancedSearch: '高级搜索',
        searchFields: '搜索范围',
        prompt: '提示词',
        tags: '标签',
        filterByTag: '按标签筛选',
        filterByStatus: '按状态筛选',
        favorited: '已收藏',
        applyFilter: '应用筛选',
        searchFailed: '搜索失败',
        noTags: '暂无标签',

        // 快捷筛选
        allSessions: '全部',
        favorites: '收藏',
        today: '今天',

        // 批量操作
        batchMode: '批量',
        selected: '已选择',
        cancel: '取消',
        selectSessionsFirst: '请先选择会话',
        batchFavoriteSuccess: '已收藏 {count} 个会话',
        batchExportSuccess: '已导出 {count} 个会话',
        batchDeleteSuccess: '已删除 {count} 个会话',
        batchOperationFailed: '批量操作失败',
        confirmBatchDelete: '确定要删除 {count} 个会话吗？此操作不可撤销。',

        // 会话元数据
        favorite: '收藏',
        unfavorite: '取消收藏',
        addedToFavorite: '已添加到收藏',
        removedFromFavorite: '已从收藏中移除',
        addTag: '添加标签',
        removeTag: '移除标签',
        tagAdded: '标签已添加',
        tagRemoved: '标签已移除',
        enterTagName: '请输入标签名称:',
        addNote: '添加备注',
        noteSaved: '备注已保存',
        enterNote: '请输入备注:',
        autoTagsApplied: '已应用自动标签: {tags}',
        noNewAutoTags: '没有新的自动标签',

        // 知识库
        knowledge: '知识库',
        knowledgeTab: '知识库',
        plans: '计划',
        memory: '记忆',
        newDocument: '新建',
        searchDocuments: '搜索文档...',
        selectDocument: '选择文档查看',
        selectDocumentToView: '选择文档查看内容',
        editDocument: '编辑',
        saveDocument: '保存',
        deleteDocument: '删除',
        documentSaved: '文档已保存',
        documentDeleted: '文档已删除',
        confirmDeleteDocument: '确定要删除这个文档吗？',
        enterDocumentTitle: '请输入文档标题:',
        noDocuments: '暂无文档',
        all: '全部',
        minutesAgo: '分钟前',
        hoursAgo: '小时前',
        daysAgo: '天前',

        // CLAUDE.md 编辑器
        claudeMD: 'CLAUDE.md',
        sectionEditor: '分节编辑器',
        overview: '概述',
        techStack: '技术栈',
        conventions: '代码规范',
        architecture: '架构',
        commands: '常用命令',
        noSections: '无分节',
        inputMarkdown: '输入 Markdown 内容...',
        characters: '字符',
        preview: '预览',
        generateFromProject: '从项目生成',
        detecting: '检测中...',
        detectionFailed: '检测失败',
        projectDetected: '检测结果',
        saveAllSections: '保存全部',
        copyToClipboard: '复制到剪贴板',
        savedToProject: '已保存到项目',
        copyFailed: '复制失败',
        selectProject: '选择项目',
        selectProjectToGenerate: '选择项目以生成 CLAUDE.md',
        hasCLAUDE: '已有 CLAUDE.md',
        noCLAUDE: '无 CLAUDE.md',
        language: '语言',
        framework: '框架',
        buildTool: '构建工具',
        features: '特性',
        noProjects: '未找到项目',
        cancel: '取消',
    },
    en: {
        // Header
        appTitle: 'AgentScope',
        refresh: 'Refresh',
        monitor: 'Monitor',

        // Tab Navigation
        dashboard: 'Token Overview',
        sessionsTab: 'Sessions',

        // Monitoring Status
        monitoringActive: 'Real-time monitoring active',
        monitoringInactive: 'Real-time monitoring stopped',
        monitoringStarted: 'Monitoring started',
        monitoringStopped: 'Monitoring stopped',
        monitoringFailed: 'Failed to start monitoring',
        sessionUpdated: 'New session detected',

        // Export
        export: 'Export',
        exporting: 'Exporting...',
        exportSuccess: 'Export successful',
        exportFailed: 'Export failed',
        selectSessionFirst: 'Please select a session first',

        // Sidebar
        sessions: 'Sessions',
        searchPlaceholder: 'Search...',
        files: 'Files',
        actions: 'Actions',

        // File Table
        fileChanges: 'File Changes',
        risk: 'RISK',
        file: 'FILE',
        change: 'CHANGE',
        ops: 'OPS',

        // Risk Levels
        safe: 'Safe',
        review: 'Review',
        danger: 'Danger',

        // Change Types
        created: 'Created',
        modified: 'Modified',
        deleted: 'Deleted',

        // Diff View
        diff: 'Diff',
        selectFile: 'Select a file to view diff',
        noDiff: 'No diff data',
        diffModeUncommitted: 'Uncommitted',
        diffModeSession: 'Session Diff',

        // Status Bar
        session: 'Session',
        branch: 'Branch',
        tokenIn: 'in',
        tokenOut: 'out',

        // Empty States
        selectSession: 'Select a session to view changes',
        noChanges: 'No file changes',
        loading: 'Loading...',
        loadFailed: 'Failed to load',
        noSessions: 'No sessions found',

        // Conversation
        conversation: 'Conversation',
        userMessage: 'User',
        aiMessage: 'AI',
        toolMessage: 'Tool',
        thinking: 'Thinking',
        toolCall: 'Tool Call',
        toolResult: 'Tool Result',
        copyCommand: 'Copy Command',
        commandCopied: 'Copied',
        loadingConversation: 'Loading conversation...',
        noConversation: 'No conversation',

        // Language Switch
        langSwitch: '中文',

        // Settings Panel
        settings: 'Settings',
        theme: 'Theme',
        themeLight: 'Light',
        themeDark: 'Dark',
        themeAuto: 'System',
        customRules: 'Custom Risk Rules',
        addRule: 'Add Rule',
        deleteRule: 'Delete',
        noRules: 'No custom rules',
        ruleNamePlaceholder: 'Rule name:',
        ruleDescPlaceholder: 'Rule description:',
        ruleLevelPlaceholder: 'Risk level (safe/review/danger):',
        rulePatternPlaceholder: 'File path pattern:',
        invalidLevel: 'Invalid risk level',
        ruleAdded: 'Rule added',
        confirmDelete: 'Are you sure you want to delete this rule?',
        ruleDeleted: 'Rule deleted',

        // Dashboard
        todayTokens: 'Today\'s Tokens',
        thisMonthTokens: 'This Month',
        lastMonthTokens: 'Last Month',
        totalTokensAll: 'Total Tokens',
        totalSessions: 'Total Sessions',
        vsLastMonth: 'vs last month',
        tokens: 'Tokens',
        tokenTrend: 'Token Trend (Last 30 Days)',
        projectBreakdown: 'Project Token Breakdown',
        modelDistribution: 'Model Distribution',
        project: 'Project',
        model: 'Model',
        projectSessions: 'Sessions',
        noData: 'No data available',

        // Session Management Enhancement
        // Advanced Search
        advancedSearch: 'Advanced Search',
        searchFields: 'Search Fields',
        prompt: 'Prompt',
        tags: 'Tags',
        filterByTag: 'Filter by Tag',
        filterByStatus: 'Filter by Status',
        favorited: 'Favorited',
        applyFilter: 'Apply Filter',
        searchFailed: 'Search failed',
        noTags: 'No tags',

        // Quick Filters
        allSessions: 'All',
        favorites: 'Favorites',
        today: 'Today',

        // Batch Operations
        batchMode: 'Batch',
        selected: 'selected',
        cancel: 'Cancel',
        selectSessionsFirst: 'Please select sessions first',
        batchFavoriteSuccess: 'Favorited {count} sessions',
        batchExportSuccess: 'Exported {count} sessions',
        batchDeleteSuccess: 'Deleted {count} sessions',
        batchOperationFailed: 'Batch operation failed',
        confirmBatchDelete: 'Are you sure you want to delete {count} sessions? This cannot be undone.',

        // Session Metadata
        favorite: 'Favorite',
        unfavorite: 'Unfavorite',
        addedToFavorite: 'Added to favorites',
        removedFromFavorite: 'Removed from favorites',
        addTag: 'Add Tag',
        removeTag: 'Remove Tag',
        tagAdded: 'Tag added',
        tagRemoved: 'Tag removed',
        enterTagName: 'Enter tag name:',
        addNote: 'Add Note',
        noteSaved: 'Note saved',
        enterNote: 'Enter note:',
        autoTagsApplied: 'Auto tags applied: {tags}',
        noNewAutoTags: 'No new auto tags',

        // Knowledge
        knowledge: 'Knowledge',
        knowledgeTab: 'Knowledge',
        plans: 'Plans',
        memory: 'Memory',
        newDocument: 'New',
        searchDocuments: 'Search documents...',
        selectDocument: 'Select document',
        selectDocumentToView: 'Select a document to view',
        editDocument: 'Edit',
        saveDocument: 'Save',
        deleteDocument: 'Delete',
        documentSaved: 'Document saved',
        documentDeleted: 'Document deleted',
        confirmDeleteDocument: 'Are you sure you want to delete this document?',
        enterDocumentTitle: 'Enter document title:',
        noDocuments: 'No documents',
        all: 'All',
        minutesAgo: 'minutes ago',
        hoursAgo: 'hours ago',
        daysAgo: 'days ago',

        // CLAUDE.md Editor
        claudeMD: 'CLAUDE.md',
        sectionEditor: 'Section Editor',
        overview: 'Overview',
        techStack: 'Tech Stack',
        conventions: 'Conventions',
        architecture: 'Architecture',
        commands: 'Commands',
        noSections: 'No sections',
        inputMarkdown: 'Enter Markdown content...',
        characters: 'characters',
        preview: 'Preview',
        generateFromProject: 'Generate from Project',
        detecting: 'Detecting...',
        detectionFailed: 'Detection failed',
        projectDetected: 'Detection Result',
        saveAllSections: 'Save All',
        copyToClipboard: 'Copy to Clipboard',
        savedToProject: 'Saved to project',
        copyFailed: 'Copy failed',
        selectProject: 'Select Project',
        selectProjectToGenerate: 'Select a project to generate CLAUDE.md',
        hasCLAUDE: 'Has CLAUDE.md',
        noCLAUDE: 'No CLAUDE.md',
        language: 'Language',
        framework: 'Framework',
        buildTool: 'Build Tool',
        features: 'Features',
        noProjects: 'No projects found',
    }
};

// 当前语言
let currentLang = localStorage.getItem('lang') || 'zh';

// 获取翻译
function t(key) {
    return i18n[currentLang][key] || key;
}

// 切换语言
function switchLang() {
    currentLang = currentLang === 'zh' ? 'en' : 'zh';
    localStorage.setItem('lang', currentLang);
    updateUI();
    return currentLang;
}

// 获取当前语言
function getCurrentLang() {
    return currentLang;
}
