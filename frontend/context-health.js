// @ts-check
// Context Health Dashboard — 上下文健康仪表盘

let ctxGrowthChart = null;
let ctxScoreChart = null;

// ---- 加载上下文健康数据 ----

let ctxHealthLoading = false;

async function loadContextHealth() {
    if (ctxHealthLoading) return;
    ctxHealthLoading = true;

    // 显示加载状态
    const panel = document.getElementById('panel-context-health');
    if (panel) panel.classList.add('loading');

    try {
        const overview = await window.go.main.App.GetContextHealthOverview();
        if (!overview) return;

        console.log('[ContextHealth] overview:', JSON.stringify(overview, null, 2));

        renderContextOverviewCards(overview);
        renderContextGrowthChart(overview.topSessions);
        renderHealthScoreChart(overview);
        renderHealthSessionTable(overview.topSessions);
    } catch (error) {
        console.error('Failed to load context health:', error);
    } finally {
        ctxHealthLoading = false;
        if (panel) panel.classList.remove('loading');
    }
}

// ---- 概览卡片 ----

function renderContextOverviewCards(data) {
    document.getElementById('ctxAvgUsage').textContent = data.avgContextUsage.toFixed(1) + '%';
    document.getElementById('ctxMaxUsage').textContent = data.maxContextUsage.toFixed(1) + '%';
    document.getElementById('ctxAvgScore').textContent = Math.round(data.avgHealthScore) + '分';

    const alertCount = data.warningCount + data.criticalCount;
    const alertEl = document.getElementById('ctxAlertCount');
    alertEl.textContent = alertCount + '个';
    if (data.criticalCount > 0) {
        alertEl.style.color = 'var(--red)';
    } else if (data.warningCount > 0) {
        alertEl.style.color = 'var(--orange)';
    } else {
        alertEl.style.color = '';
    }
}

// ---- 采样数据点（避免图表过密）----

function sampleTurns(turns, maxPoints) {
    if (!turns || turns.length <= maxPoints) return turns;
    const step = Math.ceil(turns.length / maxPoints);
    const sampled = [];
    for (let i = 0; i < turns.length; i += step) {
        sampled.push(turns[i]);
    }
    // 确保最后一个点被包含
    const last = turns[turns.length - 1];
    if (sampled[sampled.length - 1] !== last) {
        sampled.push(last);
    }
    return sampled;
}

// ---- 上下文增长趋势图 (Chart.js) ----

function renderContextGrowthChart(sessions) {
    if (!sessions || sessions.length === 0) {
        console.warn('[ContextHealth] no sessions for growth chart');
        return;
    }

    const canvas = document.getElementById('ctxGrowthChart');
    if (!canvas) return;
    if (typeof Chart === 'undefined') {
        console.error('[ContextHealth] Chart.js is NOT loaded!');
        return;
    }

    if (ctxGrowthChart) {
        ctxGrowthChart.destroy();
        ctxGrowthChart = null;
    }

    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';

    const colors = isDark ? {
        text: '#8e8e93',
        grid: 'rgba(255, 255, 255, 0.06)',
        line1: 'rgba(99, 179, 237, 1)',
        fill1: 'rgba(99, 179, 237, 0.12)',
        line2: 'rgba(72, 187, 120, 1)',
        fill2: 'rgba(72, 187, 120, 0.12)',
        line3: 'rgba(255, 149, 0, 1)',
        fill3: 'rgba(255, 149, 0, 0.12)',
        line4: 'rgba(175, 82, 222, 1)',
        fill4: 'rgba(175, 82, 222, 0.12)',
        line5: 'rgba(255, 59, 48, 1)',
        fill5: 'rgba(255, 59, 48, 0.12)',
    } : {
        text: '#86868b',
        grid: 'rgba(0, 0, 0, 0.04)',
        line1: 'rgba(0, 122, 255, 1)',
        fill1: 'rgba(0, 122, 255, 0.08)',
        line2: 'rgba(52, 199, 89, 1)',
        fill2: 'rgba(52, 199, 89, 0.08)',
        line3: 'rgba(255, 149, 0, 1)',
        fill3: 'rgba(255, 149, 0, 0.08)',
        line4: 'rgba(175, 82, 222, 1)',
        fill4: 'rgba(175, 82, 222, 0.08)',
        line5: 'rgba(255, 59, 48, 1)',
        fill5: 'rgba(255, 59, 48, 0.08)',
    };

    const lineColors = [colors.line1, colors.line2, colors.line3, colors.line4, colors.line5];
    const fillColors = [colors.fill1, colors.fill2, colors.fill3, colors.fill4, colors.fill5];

    // 取前 5 个会话，每个最多 80 个数据点
    const topSessions = sessions.slice(0, 5);
    const MAX_POINTS = 80;

    let maxTurns = 0;
    const datasets = topSessions.map((s, i) => {
        const sampled = sampleTurns(s.turns, MAX_POINTS);
        const dataPoints = sampled.map(t => t.inputTokens);
        if (sampled.length > maxTurns) maxTurns = sampled.length;
        return {
            label: s.sessionId.substring(0, 8) + ' (' + s.model + ')',
            data: dataPoints,
            borderColor: lineColors[i % lineColors.length],
            backgroundColor: fillColors[i % fillColors.length],
            fill: false,
            tension: 0.2,
            pointRadius: 2,
            pointHoverRadius: 5,
            borderWidth: 2,
        };
    });

    const labels = Array.from({ length: maxTurns }, (_, i) => {
        if (maxTurns <= 20) return 'Turn ' + (i + 1);
        // 稀疏标签
        if (i % Math.ceil(maxTurns / 15) === 0 || i === maxTurns - 1) return 'T' + (i + 1);
        return '';
    });

    // 200K 上下文限制线
    const limitLine = {
        label: '200K 上下文限制',
        data: Array(maxTurns).fill(200000),
        borderColor: 'rgba(255, 59, 48, 0.4)',
        borderDash: [6, 4],
        borderWidth: 1.5,
        pointRadius: 0,
        fill: false,
    };

    ctxGrowthChart = new Chart(canvas, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [...datasets, limitLine],
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            animation: {
                duration: 600,
                easing: 'easeOutQuart',
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    align: 'end',
                    labels: {
                        color: colors.text,
                        font: { size: 10, family: '-apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", Roboto, sans-serif' },
                        usePointStyle: true,
                        pointStyleWidth: 8,
                        boxHeight: 6,
                        padding: 12,
                    }
                },
                tooltip: {
                    backgroundColor: isDark ? 'rgba(30,30,30,0.95)' : 'rgba(255,255,255,0.95)',
                    titleColor: isDark ? '#fff' : '#1d1d1f',
                    bodyColor: isDark ? '#fff' : '#1d1d1f',
                    padding: { top: 10, bottom: 10, left: 14, right: 14 },
                    cornerRadius: 10,
                    borderColor: isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.08)',
                    borderWidth: 1,
                    callbacks: {
                        label: function(ctx) {
                            return '  ' + ctx.dataset.label + ': ' + formatTokenCount(ctx.parsed.y);
                        }
                    }
                }
            },
            scales: {
                x: {
                    border: { display: false },
                    ticks: {
                        color: colors.text,
                        font: { size: 10 },
                        maxRotation: 0,
                        autoSkip: true,
                        maxTicksLimit: 15,
                        padding: 8,
                    },
                    grid: { display: false },
                },
                y: {
                    border: { display: false },
                    ticks: {
                        color: colors.text,
                        font: { size: 10 },
                        callback: v => formatTokenCount(v),
                        padding: 8,
                        maxTicksLimit: 6,
                    },
                    grid: { color: colors.grid, drawTicks: false },
                }
            }
        }
    });
}

// ---- 健康评分分布 (Chart.js 环形图) ----

function renderHealthScoreChart(data) {
    const canvas = document.getElementById('ctxScoreChart');
    if (!canvas) return;
    if (typeof Chart === 'undefined') return;

    if (ctxScoreChart) {
        ctxScoreChart.destroy();
        ctxScoreChart = null;
    }

    const isDark = document.documentElement.getAttribute('data-theme') === 'dark';

    // 统计各等级数量
    const counts = { excellent: 0, good: 0, warning: 0, critical: 0 };
    if (data.topSessions) {
        data.topSessions.forEach(s => {
            counts[s.healthLevel] = (counts[s.healthLevel] || 0) + 1;
        });
    }

    const total = counts.excellent + counts.good + counts.warning + counts.critical;
    if (total === 0) {
        const ctx = canvas.getContext('2d');
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = isDark ? '#8e8e93' : '#86868b';
        ctx.font = '13px -apple-system, sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText(t('noData'), canvas.width / 2, canvas.height / 2);
        return;
    }

    ctxScoreChart = new Chart(canvas, {
        type: 'doughnut',
        data: {
            labels: [t('excellent'), t('good'), t('warning'), t('critical')],
            datasets: [{
                data: [counts.excellent, counts.good, counts.warning, counts.critical],
                backgroundColor: [
                    'rgba(52, 199, 89, 0.8)',   // green
                    'rgba(0, 122, 255, 0.8)',    // blue
                    'rgba(255, 149, 0, 0.8)',    // orange
                    'rgba(255, 59, 48, 0.8)',    // red
                ],
                borderColor: isDark ? '#1c1c1e' : '#ffffff',
                borderWidth: 2,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            cutout: '60%',
            plugins: {
                legend: {
                    position: 'bottom',
                    labels: {
                        color: isDark ? '#8e8e93' : '#86868b',
                        font: { size: 11, family: '-apple-system, sans-serif' },
                        padding: 12,
                        usePointStyle: true,
                        pointStyleWidth: 8,
                    }
                },
                tooltip: {
                    backgroundColor: isDark ? 'rgba(30,30,30,0.95)' : 'rgba(255,255,255,0.95)',
                    titleColor: isDark ? '#fff' : '#1d1d1f',
                    bodyColor: isDark ? '#fff' : '#1d1d1f',
                    padding: 10,
                    cornerRadius: 8,
                    callbacks: {
                        label: function(ctx) {
                            const pct = total > 0 ? (ctx.raw / total * 100).toFixed(1) : 0;
                            return ' ' + ctx.label + ': ' + ctx.raw + ' (' + pct + '%)';
                        }
                    }
                }
            }
        }
    });
}

// ---- 会话健康列表 ----

function renderHealthSessionTable(sessions) {
    const tbody = document.getElementById('ctxHealthTableBody');
    if (!sessions || sessions.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state-cell">' + t('noData') + '</td></tr>';
        return;
    }

    tbody.innerHTML = sessions.map(s => {
        const levelClass = 'ctx-level-' + s.healthLevel;
        const levelText = t(s.healthLevel);
        const usageBar = `<div class="ctx-usage-bar"><div class="ctx-usage-fill ${levelClass}" style="width:${Math.min(s.contextUsagePct, 100)}%"></div></div>`;

        // 压缩事件标记
        let compressionTag = '';
        if (s.compressionEvents > 5) {
            compressionTag = '<span class="ctx-compress-tag high">' + s.compressionEvents + '</span>';
        } else if (s.compressionEvents > 0) {
            compressionTag = '<span class="ctx-compress-tag">' + s.compressionEvents + '</span>';
        }

        return `<tr>
            <td class="session-id-cell" title="${escapeHtmlAttr(s.sessionId)}">${escapeHtml(s.sessionId.substring(0, 12))}…</td>
            <td>${escapeHtml(s.model)}</td>
            <td>${usageBar}<span class="ctx-usage-text">${s.contextUsagePct.toFixed(1)}%</span></td>
            <td><span class="ctx-score-badge ${levelClass}">${s.healthScore}</span></td>
            <td><span class="ctx-level-badge ${levelClass}">${levelText}</span></td>
            <td>${compressionTag}</td>
        </tr>`;
    }).join('');
}
