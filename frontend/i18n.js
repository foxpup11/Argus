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
