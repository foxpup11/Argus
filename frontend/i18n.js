// 国际化支持
const i18n = {
    zh: {
        // 标题栏
        appTitle: 'AgentScope',
        refresh: '刷新',
        monitor: '监控',

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
    },
    en: {
        // Header
        appTitle: 'AgentScope',
        refresh: 'Refresh',
        monitor: 'Monitor',

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
