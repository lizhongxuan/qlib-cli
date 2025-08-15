// 总览页面组件
const { useState, useEffect, useCallback } = React;

const Dashboard = ({ tasks = [], datasets = [], models = [], onNavigate = () => {} }) => {
    const [timeRange, setTimeRange] = useState('1M');
    
    // 模拟性能数据
    const performanceData = [
        { date: '01-01', value: 100, benchmark: 100, volume: 85000000 },
        { date: '01-08', value: 105, benchmark: 102, volume: 92000000 },
        { date: '01-15', value: 108, benchmark: 101, volume: 88000000 },
        { date: '01-22', value: 115, benchmark: 105, volume: 95000000 },
        { date: '02-01', value: 112, benchmark: 103, volume: 91000000 },
        { date: '02-08', value: 118, benchmark: 106, volume: 98000000 },
        { date: '02-15', value: 125, benchmark: 108, volume: 103000000 },
        { date: '02-22', value: 122, benchmark: 107, volume: 87000000 },
    ];

    // 市场数据
    const marketData = {
        sh000300: { value: 3456.78, change: '+1.23%', trend: 'up' },
        sz399905: { value: 6789.12, change: '-0.45%', trend: 'down' },
        sh000905: { value: 5432.10, change: '+0.67%', trend: 'up' }
    };

    const stats = {
        totalReturn: '+25.0%',
        sharpeRatio: '1.85',
        maxDrawdown: '-8.5%',
        winRate: '62.3%'
    };

    return (
        <div className="dashboard">
            <div className="dashboard-header">
                <h1>🏠 系统总览</h1>
                <div className="time-range-selector">
                    {['1D', '1W', '1M', '3M', '1Y'].map(range => (
                        <button 
                            key={range}
                            className={`time-btn ${timeRange === range ? 'active' : ''}`}
                            onClick={() => setTimeRange(range)}
                        >
                            {range}
                        </button>
                    ))}
                </div>
            </div>
            
            {/* 快速操作区域 */}
            <div className="quick-actions">
                <h2>🚀 快速开始</h2>
                <div className="action-cards">
                    <div className="action-card" onClick={() => onNavigate('data')}>
                        <div className="action-icon">💾</div>
                        <div className="action-content">
                            <h3>准备数据</h3>
                            <p>创建数据集，选择股票池和特征</p>
                        </div>
                        <div className="action-arrow">→</div>
                    </div>
                    <div className="action-card" onClick={() => onNavigate('factor-workshop')}>
                        <div className="action-icon">🧮</div>
                        <div className="action-content">
                            <h3>因子工程</h3>
                            <p>查看内置因子，手动编辑或AI生成</p>
                        </div>
                        <div className="action-arrow">→</div>
                    </div>
                    <div className="action-card" onClick={() => onNavigate('model')}>
                        <div className="action-icon">🤖</div>
                        <div className="action-content">
                            <h3>训练模型</h3>
                            <p>配置和训练量化预测模型</p>
                        </div>
                        <div className="action-arrow">→</div>
                    </div>
                    <div className="action-card" onClick={() => onNavigate('backtest')}>
                        <div className="action-icon">📈</div>
                        <div className="action-content">
                            <h3>策略回测</h3>
                            <p>测试策略历史表现</p>
                        </div>
                        <div className="action-arrow">→</div>
                    </div>
                </div>
            </div>

            {/* 市场概览 */}
            <div className="market-overview">
                <h2>📊 市场概览</h2>
                <div className="market-cards">
                    <div className="market-card">
                        <div className="market-name">沪深300</div>
                        <div className="market-value">{marketData.sh000300.value}</div>
                        <div className={`market-change ${marketData.sh000300.trend}`}>
                            {marketData.sh000300.change}
                        </div>
                    </div>
                    <div className="market-card">
                        <div className="market-name">中证500</div>
                        <div className="market-value">{marketData.sz399905.value}</div>
                        <div className={`market-change ${marketData.sz399905.trend}`}>
                            {marketData.sz399905.change}
                        </div>
                    </div>
                    <div className="market-card">
                        <div className="market-name">中证1000</div>
                        <div className="market-value">{marketData.sh000905.value}</div>
                        <div className={`market-change ${marketData.sh000905.trend}`}>
                            {marketData.sh000905.change}
                        </div>
                    </div>
                </div>
            </div>
            
            {/* 统计卡片 */}
            <div className="stats-section">
                <h2>📈 策略表现</h2>
                <div className="stats-grid">
                    <div className="stat-card">
                        <div className="stat-icon">💰</div>
                        <div className="stat-content">
                            <div className="stat-label">累计收益</div>
                            <div className="stat-value positive">{stats.totalReturn}</div>
                        </div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-icon">⚡</div>
                        <div className="stat-content">
                            <div className="stat-label">夏普比率</div>
                            <div className="stat-value">{stats.sharpeRatio}</div>
                        </div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-icon">📉</div>
                        <div className="stat-content">
                            <div className="stat-label">最大回撤</div>
                            <div className="stat-value negative">{stats.maxDrawdown}</div>
                        </div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-icon">🎯</div>
                        <div className="stat-content">
                            <div className="stat-label">胜率</div>
                            <div className="stat-value">{stats.winRate}</div>
                        </div>
                    </div>
                </div>
            </div>

            {/* 图表区域 */}
            <div className="chart-container">
                <div className="chart-header">
                    <h2>📊 净值走势</h2>
                    <div className="chart-controls">
                        <button className="chart-btn active">净值对比</button>
                        <button className="chart-btn">收益分布</button>
                    </div>
                </div>
                <div className="chart-content">
                    <div className="chart-wrapper">
                        <svg viewBox="0 0 800 300" className="performance-chart">
                            {/* 网格线 */}
                            <defs>
                                <pattern id="grid" width="50" height="30" patternUnits="userSpaceOnUse">
                                    <path d="M 50 0 L 0 0 0 30" fill="none" stroke="#f0f0f0" strokeWidth="1"/>
                                </pattern>
                            </defs>
                            <rect width="800" height="300" fill="url(#grid)" />
                            
                            {/* 策略收益线 */}
                            <polyline
                                points={performanceData.map((d, i) => 
                                    `${i * 100},${300 - (d.value - 90) * 10}`
                                ).join(' ')}
                                fill="none"
                                stroke="#1890ff"
                                strokeWidth="3"
                                style={{filter: 'drop-shadow(0 2px 4px rgba(24,144,255,0.3))'}}
                            />
                            
                            {/* 基准收益线 */}
                            <polyline
                                points={performanceData.map((d, i) => 
                                    `${i * 100},${300 - (d.benchmark - 90) * 10}`
                                ).join(' ')}
                                fill="none"
                                stroke="#52c41a"
                                strokeWidth="2"
                                strokeDasharray="5,5"
                            />
                            
                            {/* 数据点 */}
                            {performanceData.map((d, i) => (
                                <g key={i}>
                                    <circle
                                        cx={i * 100}
                                        cy={300 - (d.value - 90) * 10}
                                        r="4"
                                        fill="#1890ff"
                                        style={{cursor: 'pointer'}}
                                    />
                                </g>
                            ))}
                            
                            {/* Y轴标签 */}
                            <text x="10" y="20" fill="#666" fontSize="12">125</text>
                            <text x="10" y="120" fill="#666" fontSize="12">110</text>
                            <text x="10" y="220" fill="#666" fontSize="12">100</text>
                            <text x="10" y="290" fill="#666" fontSize="12">90</text>
                            
                            {/* X轴标签 */}
                            {performanceData.map((d, i) => (
                                <text key={i} x={i * 100 - 15} y="295" fill="#666" fontSize="10">
                                    {d.date}
                                </text>
                            ))}
                        </svg>
                    </div>
                    <div className="chart-legend">
                        <div className="legend-item">
                            <div className="legend-color strategy"></div>
                            <span>量化策略</span>
                            <span className="legend-value">+22.0%</span>
                        </div>
                        <div className="legend-item">
                            <div className="legend-color benchmark"></div>
                            <span>基准指数</span>
                            <span className="legend-value">+7.0%</span>
                        </div>
                        <div className="legend-item">
                            <div className="legend-color excess"></div>
                            <span>超额收益</span>
                            <span className="legend-value">+15.0%</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* 系统状态 */}
            <div className="bottom-section">
                <div className="system-status">
                    <h2>🖥️ 系统状态</h2>
                    <div className="status-grid">
                        <div className="status-card">
                            <div className="status-header">
                                <span className="status-title">数据集</span>
                                <div className="status-indicator online"></div>
                            </div>
                            <div className="status-count">{datasets.length}</div>
                            <div className="status-detail">
                                {datasets.filter(d => d.status === 'ready').length} 个可用
                            </div>
                            <div className="status-progress">
                                <div className="progress-bar" style={{
                                    width: `${(datasets.filter(d => d.status === 'ready').length / datasets.length) * 100}%`
                                }}></div>
                            </div>
                        </div>
                        
                        <div className="status-card">
                            <div className="status-header">
                                <span className="status-title">模型</span>
                                <div className="status-indicator online"></div>
                            </div>
                            <div className="status-count">{models.length}</div>
                            <div className="status-detail">
                                {models.filter(m => m.status === 'trained').length} 个已训练
                            </div>
                            <div className="status-progress">
                                <div className="progress-bar" style={{
                                    width: `${(models.filter(m => m.status === 'trained').length / models.length) * 100}%`
                                }}></div>
                            </div>
                        </div>
                        
                        <div className="status-card">
                            <div className="status-header">
                                <span className="status-title">任务队列</span>
                                <div className="status-indicator processing"></div>
                            </div>
                            <div className="status-count">{tasks.filter(t => t.status === 'running').length}</div>
                            <div className="status-detail">
                                总计 {tasks.length} 个任务
                            </div>
                            <div className="status-progress">
                                <div className="progress-bar running" style={{width: '35%'}}></div>
                            </div>
                        </div>
                        
                        <div className="status-card">
                            <div className="status-header">
                                <span className="status-title">系统资源</span>
                                <div className="status-indicator online"></div>
                            </div>
                            <div className="status-count">65%</div>
                            <div className="status-detail">
                                CPU使用率
                            </div>
                            <div className="status-progress">
                                <div className="progress-bar" style={{width: '65%'}}></div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* 最近活动 */}
                <div className="recent-activities">
                    <h2>📋 最近活动</h2>
                    <div className="activity-list">
                        {tasks.length === 0 ? (
                            <div className="empty-state">
                                <div className="empty-icon">📭</div>
                                <div className="empty-text">暂无任务记录</div>
                                <div className="empty-sub">开始创建数据集或训练模型</div>
                            </div>
                        ) : (
                            tasks.slice(-5).reverse().map(task => (
                                <div key={task.id} className="activity-item">
                                    <div className="activity-icon">
                                        {task.status === 'completed' ? '✅' : 
                                         task.status === 'running' ? '⏳' : '❌'}
                                    </div>
                                    <div className="activity-content">
                                        <div className="activity-title">{task.name}</div>
                                        <div className="activity-time">
                                            {task.startTime}
                                            {task.status === 'running' && (
                                                <span className="activity-duration"> • 运行中</span>
                                            )}
                                        </div>
                                    </div>
                                    <div className="activity-status">
                                        {task.status === 'running' ? 
                                            <div className="progress-mini">
                                                <div className="progress-mini-bar" 
                                                     style={{width: `${Math.floor(task.progress)}%`}}>
                                                </div>
                                                <span>{Math.floor(task.progress)}%</span>
                                            </div> : 
                                            <span className={`status-text ${task.status}`}>
                                                {task.status === 'completed' ? '已完成' : '失败'}
                                            </span>
                                        }
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};