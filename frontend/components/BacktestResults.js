// 增强的回测结果展示组件
const { useState, useEffect } = React;

const BacktestResults = ({ 
    backtestData = null, 
    onNavigate = () => {},
    onCompareStrategy = () => {} 
}) => {
    const [activeChart, setActiveChart] = useState('performance');
    const [selectedPeriod, setSelectedPeriod] = useState('daily');
    const [showDrawdown, setShowDrawdown] = useState(true);
    const [showBenchmark, setShowBenchmark] = useState(true);

    // 模拟详细回测数据
    const mockBacktestData = backtestData || {
        strategy_name: '动量TopK策略',
        model_name: 'XGBoost-v1.0',
        start_date: '2022-01-01',
        end_date: '2023-12-31',
        total_periods: 504,
        performance_metrics: {
            total_return: 0.235,
            annual_return: 0.182,
            sharpe_ratio: 1.85,
            calmar_ratio: 2.13,
            max_drawdown: -0.085,
            volatility: 0.124,
            skewness: -0.23,
            kurtosis: 3.45,
            var_95: -0.023,
            cvar_95: -0.031,
            win_rate: 0.623,
            profit_factor: 2.1,
            total_trades: 1250,
            avg_trade_return: 0.00018,
            best_trade: 0.078,
            worst_trade: -0.045
        },
        // 净值曲线数据
        nav_curve: Array.from({length: 504}, (_, i) => {
            const date = new Date(2022, 0, 1 + i);
            const baseReturn = i * 0.0004;
            const randomWalk = Math.sin(i * 0.05) * 0.02 + (Math.random() - 0.5) * 0.01;
            const strategy_nav = 1 + baseReturn + randomWalk;
            const benchmark_nav = 1 + i * 0.0002 + (Math.random() - 0.5) * 0.008;
            
            return {
                date: date.toISOString().split('T')[0],
                strategy_nav: Math.max(0.8, strategy_nav).toFixed(4),
                benchmark_nav: Math.max(0.9, benchmark_nav).toFixed(4),
                excess_return: (strategy_nav - benchmark_nav).toFixed(4),
                drawdown: Math.min(0, (strategy_nav - Math.max(...Array.from({length: i+1}, (_, j) => 1 + j * 0.0004 + Math.sin(j * 0.05) * 0.02))) / Math.max(...Array.from({length: i+1}, (_, j) => 1 + j * 0.0004 + Math.sin(j * 0.05) * 0.02))).toFixed(4)
            };
        }),
        // 月度收益
        monthly_returns: Array.from({length: 24}, (_, i) => {
            const year = 2022 + Math.floor(i / 12);
            const month = (i % 12) + 1;
            return {
                period: `${year}-${month.toString().padStart(2, '0')}`,
                strategy_return: ((Math.random() - 0.4) * 0.2).toFixed(4),
                benchmark_return: ((Math.random() - 0.45) * 0.15).toFixed(4),
                excess_return: ((Math.random() - 0.3) * 0.1).toFixed(4)
            };
        }),
        // 持仓分布
        position_distribution: [
            { sector: '金融', weight: 0.28, count: 8, avg_return: 0.023 },
            { sector: '科技', weight: 0.25, count: 7, avg_return: 0.034 },
            { sector: '消费', weight: 0.18, count: 5, avg_return: 0.019 },
            { sector: '医药', weight: 0.15, count: 4, avg_return: 0.028 },
            { sector: '工业', weight: 0.14, count: 6, avg_return: 0.021 }
        ],
        // 交易分析
        trade_analysis: {
            daily_turnover: Array.from({length: 60}, (_, i) => ({
                date: new Date(2023, 6, 1 + i).toISOString().split('T')[0],
                turnover: (Math.random() * 0.3 + 0.1).toFixed(3),
                transaction_cost: (Math.random() * 0.005 + 0.001).toFixed(4)
            })),
            holding_period_dist: [
                { period: '1天', frequency: 123, percentage: 15.2 },
                { period: '2-5天', frequency: 298, percentage: 36.8 },
                { period: '6-10天', frequency: 201, percentage: 24.8 },
                { period: '11-20天', frequency: 134, percentage: 16.5 },
                { period: '>20天', frequency: 54, percentage: 6.7 }
            ],
            top_positions: [
                { symbol: '000001.SZ', name: '平安银行', weight: 0.045, return: 0.078, days_held: 12 },
                { symbol: '000002.SZ', name: '万科A', weight: 0.042, return: 0.034, days_held: 8 },
                { symbol: '600036.SH', name: '招商银行', weight: 0.041, return: 0.056, days_held: 15 },
                { symbol: '000858.SZ', name: '五粮液', weight: 0.038, return: 0.023, days_held: 7 },
                { symbol: '600519.SH', name: '贵州茅台', weight: 0.037, return: 0.045, days_held: 18 }
            ]
        },
        // 风险指标
        risk_metrics: {
            beta: 0.95,
            alpha: 0.045,
            information_ratio: 1.23,
            treynor_ratio: 0.156,
            tracking_error: 0.089,
            correlation_with_benchmark: 0.87,
            downside_deviation: 0.078,
            upside_capture: 1.15,
            downside_capture: 0.82
        }
    };

    // 图表类型配置
    const chartTypes = [
        { id: 'performance', name: '净值走势', icon: '📈' },
        { id: 'returns', name: '收益分布', icon: '📊' },
        { id: 'drawdown', name: '回撤分析', icon: '📉' },
        { id: 'positions', name: '持仓分析', icon: '🎯' },
        { id: 'trades', name: '交易分析', icon: '💹' },
        { id: 'risk', name: '风险指标', icon: '⚠️' }
    ];

    // 渲染净值走势图
    const renderPerformanceChart = () => {
        const data = mockBacktestData.nav_curve.filter((_, i) => i % 5 === 0); // 抽样显示
        const maxNav = Math.max(...data.map(d => Math.max(parseFloat(d.strategy_nav), parseFloat(d.benchmark_nav))));
        const minNav = Math.min(...data.map(d => Math.min(parseFloat(d.strategy_nav), parseFloat(d.benchmark_nav))));
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>📈 净值走势对比</h3>
                    <div className="chart-controls">
                        <label className="checkbox-label">
                            <input 
                                type="checkbox" 
                                checked={showBenchmark} 
                                onChange={(e) => setShowBenchmark(e.target.checked)}
                            />
                            显示基准
                        </label>
                        <label className="checkbox-label">
                            <input 
                                type="checkbox" 
                                checked={showDrawdown} 
                                onChange={(e) => setShowDrawdown(e.target.checked)}
                            />
                            显示回撤
                        </label>
                    </div>
                </div>
                
                <div className="chart-content">
                    <svg viewBox="0 0 900 500" className="performance-chart">
                        {/* 网格 */}
                        <defs>
                            <pattern id="performanceGrid" width="45" height="50" patternUnits="userSpaceOnUse">
                                <path d="M 45 0 L 0 0 0 50" fill="none" stroke="#f5f5f5" strokeWidth="1"/>
                            </pattern>
                            <linearGradient id="drawdownGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                                <stop offset="0%" stopColor="#ff4d4f" stopOpacity="0.3"/>
                                <stop offset="100%" stopColor="#ff4d4f" stopOpacity="0.1"/>
                            </linearGradient>
                        </defs>
                        <rect width="900" height="500" fill="url(#performanceGrid)" />
                        
                        {/* 回撤填充区域 */}
                        {showDrawdown && (
                            <path
                                d={`M 50 ${250 - (1 - minNav) * 400} 
                                    ${data.map((d, i) => 
                                        `L ${50 + i * 800/data.length} ${250 - (parseFloat(d.strategy_nav) - minNav) * 400/(maxNav - minNav) + Math.abs(parseFloat(d.drawdown)) * 200}`
                                    ).join(' ')} 
                                    ${data.map((d, i) => 
                                        `L ${50 + (data.length - 1 - i) * 800/data.length} ${250 - (parseFloat(d.strategy_nav) - minNav) * 400/(maxNav - minNav)}`
                                    ).reverse().join(' ')} Z`}
                                fill="url(#drawdownGradient)"
                            />
                        )}
                        
                        {/* 策略净值线 */}
                        <polyline
                            points={data.map((d, i) => 
                                `${50 + i * 800/data.length},${250 - (parseFloat(d.strategy_nav) - minNav) * 400/(maxNav - minNav)}`
                            ).join(' ')}
                            fill="none"
                            stroke="#1890ff"
                            strokeWidth="3"
                        />
                        
                        {/* 基准净值线 */}
                        {showBenchmark && (
                            <polyline
                                points={data.map((d, i) => 
                                    `${50 + i * 800/data.length},${250 - (parseFloat(d.benchmark_nav) - minNav) * 400/(maxNav - minNav)}`
                                ).join(' ')}
                                fill="none"
                                stroke="#52c41a"
                                strokeWidth="2"
                                strokeDasharray="5,5"
                            />
                        )}
                        
                        {/* Y轴标签 */}
                        <text x="10" y="60" fill="#666" fontSize="12">{maxNav.toFixed(2)}</text>
                        <text x="10" y="260" fill="#666" fontSize="12">{((maxNav + minNav)/2).toFixed(2)}</text>
                        <text x="10" y="460" fill="#666" fontSize="12">{minNav.toFixed(2)}</text>
                        
                        {/* X轴标签 */}
                        {data.filter((_, i) => i % 20 === 0).map((d, i) => (
                            <text key={i} x={50 + i * 20 * 800/data.length - 25} y="485" fill="#666" fontSize="10">
                                {d.date.substring(5)}
                            </text>
                        ))}
                    </svg>
                    
                    <div className="chart-legend">
                        <div className="legend-item">
                            <div className="legend-color" style={{backgroundColor: '#1890ff'}}></div>
                            <span>策略净值</span>
                            <span className="legend-value">+{(mockBacktestData.performance_metrics.total_return * 100).toFixed(1)}%</span>
                        </div>
                        {showBenchmark && (
                            <div className="legend-item">
                                <div className="legend-color" style={{backgroundColor: '#52c41a'}}></div>
                                <span>基准指数</span>
                                <span className="legend-value">+7.2%</span>
                            </div>
                        )}
                        <div className="legend-item">
                            <div className="legend-color" style={{backgroundColor: '#722ed1'}}></div>
                            <span>超额收益</span>
                            <span className="legend-value">+{((mockBacktestData.performance_metrics.total_return - 0.072) * 100).toFixed(1)}%</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染月度收益热力图
    const renderReturnsChart = () => {
        const monthlyData = mockBacktestData.monthly_returns;
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>📊 月度收益热力图</h3>
                </div>
                
                <div className="chart-content">
                    <div className="heatmap-container">
                        <div className="heatmap-grid">
                            {monthlyData.map((month, i) => {
                                const returnValue = parseFloat(month.strategy_return);
                                const intensity = Math.min(Math.abs(returnValue) * 5, 1);
                                const color = returnValue > 0 ? `rgba(82, 196, 26, ${intensity})` : `rgba(255, 77, 79, ${intensity})`;
                                
                                return (
                                    <div 
                                        key={i} 
                                        className="heatmap-cell"
                                        style={{backgroundColor: color}}
                                        title={`${month.period}: ${(returnValue * 100).toFixed(2)}%`}
                                    >
                                        <div className="cell-period">{month.period.substring(5)}</div>
                                        <div className="cell-value">{(returnValue * 100).toFixed(1)}%</div>
                                    </div>
                                );
                            })}
                        </div>
                        
                        <div className="heatmap-legend">
                            <span>负收益</span>
                            <div className="color-scale">
                                <div style={{backgroundColor: '#ff4d4f'}}></div>
                                <div style={{backgroundColor: '#ffa39e'}}></div>
                                <div style={{backgroundColor: '#f5f5f5'}}></div>
                                <div style={{backgroundColor: '#b7eb8f'}}></div>
                                <div style={{backgroundColor: '#52c41a'}}></div>
                            </div>
                            <span>正收益</span>
                        </div>
                    </div>
                    
                    {/* 收益统计 */}
                    <div className="returns-stats">
                        <div className="stat-item">
                            <span className="stat-label">胜率:</span>
                            <span className="stat-value">{(mockBacktestData.performance_metrics.win_rate * 100).toFixed(1)}%</span>
                        </div>
                        <div className="stat-item">
                            <span className="stat-label">最佳月份:</span>
                            <span className="stat-value positive">+{(Math.max(...monthlyData.map(m => parseFloat(m.strategy_return))) * 100).toFixed(1)}%</span>
                        </div>
                        <div className="stat-item">
                            <span className="stat-label">最差月份:</span>
                            <span className="stat-value negative">{(Math.min(...monthlyData.map(m => parseFloat(m.strategy_return))) * 100).toFixed(1)}%</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染持仓分析
    const renderPositionsChart = () => {
        const sectorData = mockBacktestData.position_distribution;
        const topPositions = mockBacktestData.trade_analysis.top_positions;
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>🎯 持仓结构分析</h3>
                </div>
                
                <div className="chart-content">
                    <div className="positions-layout">
                        {/* 行业分布饼图 */}
                        <div className="sector-distribution">
                            <h4>行业分布</h4>
                            <svg viewBox="0 0 300 300" className="pie-chart">
                                <g transform="translate(150,150)">
                                    {sectorData.map((sector, i) => {
                                        const startAngle = sectorData.slice(0, i).reduce((sum, s) => sum + s.weight, 0) * 2 * Math.PI;
                                        const endAngle = startAngle + sector.weight * 2 * Math.PI;
                                        const largeArcFlag = sector.weight > 0.5 ? 1 : 0;
                                        
                                        const x1 = Math.cos(startAngle) * 80;
                                        const y1 = Math.sin(startAngle) * 80;
                                        const x2 = Math.cos(endAngle) * 80;
                                        const y2 = Math.sin(endAngle) * 80;
                                        
                                        const pathData = [
                                            `M 0 0`,
                                            `L ${x1} ${y1}`,
                                            `A 80 80 0 ${largeArcFlag} 1 ${x2} ${y2}`,
                                            'Z'
                                        ].join(' ');
                                        
                                        const colors = ['#1890ff', '#52c41a', '#faad14', '#f759ab', '#13c2c2'];
                                        
                                        return (
                                            <path
                                                key={i}
                                                d={pathData}
                                                fill={colors[i % colors.length]}
                                                opacity="0.8"
                                                stroke="#fff"
                                                strokeWidth="2"
                                            />
                                        );
                                    })}
                                </g>
                            </svg>
                            
                            <div className="sector-legend">
                                {sectorData.map((sector, i) => {
                                    const colors = ['#1890ff', '#52c41a', '#faad14', '#f759ab', '#13c2c2'];
                                    return (
                                        <div key={i} className="legend-item">
                                            <div className="legend-color" style={{backgroundColor: colors[i % colors.length]}}></div>
                                            <span>{sector.sector}</span>
                                            <span className="legend-value">{(sector.weight * 100).toFixed(1)}%</span>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                        
                        {/* 重仓股表格 */}
                        <div className="top-positions">
                            <h4>重仓股票</h4>
                            <table className="positions-table">
                                <thead>
                                    <tr>
                                        <th>股票代码</th>
                                        <th>股票名称</th>
                                        <th>权重</th>
                                        <th>收益率</th>
                                        <th>持有天数</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {topPositions.map((pos, i) => (
                                        <tr key={i}>
                                            <td>{pos.symbol}</td>
                                            <td>{pos.name}</td>
                                            <td>{(pos.weight * 100).toFixed(1)}%</td>
                                            <td className={pos.return > 0 ? 'positive' : 'negative'}>
                                                {(pos.return * 100).toFixed(1)}%
                                            </td>
                                            <td>{pos.days_held}天</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染风险指标
    const renderRiskChart = () => {
        const riskData = mockBacktestData.risk_metrics;
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>⚠️ 风险指标分析</h3>
                </div>
                
                <div className="chart-content">
                    <div className="risk-metrics-grid">
                        <div className="risk-category">
                            <h4>收益风险指标</h4>
                            <div className="metrics-list">
                                <div className="metric-row">
                                    <span className="metric-name">夏普比率</span>
                                    <span className="metric-value">{mockBacktestData.performance_metrics.sharpe_ratio.toFixed(2)}</span>
                                    <div className="metric-bar">
                                        <div className="bar-fill" style={{width: `${Math.min(mockBacktestData.performance_metrics.sharpe_ratio * 20, 100)}%`}}></div>
                                    </div>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">卡玛比率</span>
                                    <span className="metric-value">{mockBacktestData.performance_metrics.calmar_ratio.toFixed(2)}</span>
                                    <div className="metric-bar">
                                        <div className="bar-fill" style={{width: `${Math.min(mockBacktestData.performance_metrics.calmar_ratio * 15, 100)}%`}}></div>
                                    </div>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">信息比率</span>
                                    <span className="metric-value">{riskData.information_ratio.toFixed(2)}</span>
                                    <div className="metric-bar">
                                        <div className="bar-fill" style={{width: `${Math.min(riskData.information_ratio * 25, 100)}%`}}></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div className="risk-category">
                            <h4>波动率指标</h4>
                            <div className="metrics-list">
                                <div className="metric-row">
                                    <span className="metric-name">年化波动率</span>
                                    <span className="metric-value">{(mockBacktestData.performance_metrics.volatility * 100).toFixed(1)}%</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">跟踪误差</span>
                                    <span className="metric-value">{(riskData.tracking_error * 100).toFixed(1)}%</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">下行偏差</span>
                                    <span className="metric-value">{(riskData.downside_deviation * 100).toFixed(1)}%</span>
                                </div>
                            </div>
                        </div>
                        
                        <div className="risk-category">
                            <h4>市场相关性</h4>
                            <div className="metrics-list">
                                <div className="metric-row">
                                    <span className="metric-name">Beta系数</span>
                                    <span className="metric-value">{riskData.beta.toFixed(2)}</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">Alpha系数</span>
                                    <span className="metric-value">{(riskData.alpha * 100).toFixed(1)}%</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">相关系数</span>
                                    <span className="metric-value">{riskData.correlation_with_benchmark.toFixed(2)}</span>
                                </div>
                            </div>
                        </div>
                        
                        <div className="risk-category">
                            <h4>极值风险</h4>
                            <div className="metrics-list">
                                <div className="metric-row">
                                    <span className="metric-name">VaR(95%)</span>
                                    <span className="metric-value negative">{(mockBacktestData.performance_metrics.var_95 * 100).toFixed(1)}%</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">CVaR(95%)</span>
                                    <span className="metric-value negative">{(mockBacktestData.performance_metrics.cvar_95 * 100).toFixed(1)}%</span>
                                </div>
                                <div className="metric-row">
                                    <span className="metric-name">最大回撤</span>
                                    <span className="metric-value negative">{(mockBacktestData.performance_metrics.max_drawdown * 100).toFixed(1)}%</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染当前选中的图表
    const renderActiveChart = () => {
        switch(activeChart) {
            case 'performance':
                return renderPerformanceChart();
            case 'returns':
                return renderReturnsChart();
            case 'drawdown':
                return renderPerformanceChart(); // 可以复用，通过showDrawdown控制
            case 'positions':
                return renderPositionsChart();
            case 'trades':
                return renderPositionsChart(); // 可以扩展为交易分析
            case 'risk':
                return renderRiskChart();
            default:
                return renderPerformanceChart();
        }
    };

    return (
        <div className="backtest-results">
            <div className="results-header">
                <div className="strategy-info">
                    <h1>📈 {mockBacktestData.strategy_name}</h1>
                    <div className="strategy-meta">
                        <span>模型: {mockBacktestData.model_name}</span>
                        <span>•</span>
                        <span>期间: {mockBacktestData.start_date} 至 {mockBacktestData.end_date}</span>
                        <span>•</span>
                        <span>交易日: {mockBacktestData.total_periods}天</span>
                    </div>
                </div>
                
                <div className="results-actions">
                    <button className="btn-secondary" onClick={() => onCompareStrategy()}>
                        📊 策略对比
                    </button>
                    <button className="btn-secondary">
                        📤 导出报告
                    </button>
                    <button className="btn-primary" onClick={() => onNavigate('portfolio')}>
                        💼 实盘部署
                    </button>
                </div>
            </div>

            {/* 核心指标概览 */}
            <div className="metrics-overview">
                <div className="metrics-grid">
                    <div className="metric-card highlight">
                        <div className="metric-icon">💰</div>
                        <div className="metric-content">
                            <div className="metric-label">总收益率</div>
                            <div className="metric-value positive">
                                +{(mockBacktestData.performance_metrics.total_return * 100).toFixed(1)}%
                            </div>
                            <div className="metric-sub">年化 +{(mockBacktestData.performance_metrics.annual_return * 100).toFixed(1)}%</div>
                        </div>
                    </div>
                    
                    <div className="metric-card">
                        <div className="metric-icon">⚡</div>
                        <div className="metric-content">
                            <div className="metric-label">夏普比率</div>
                            <div className="metric-value">{mockBacktestData.performance_metrics.sharpe_ratio.toFixed(2)}</div>
                            <div className="metric-sub">优秀水平</div>
                        </div>
                    </div>
                    
                    <div className="metric-card">
                        <div className="metric-icon">📉</div>
                        <div className="metric-content">
                            <div className="metric-label">最大回撤</div>
                            <div className="metric-value negative">{(mockBacktestData.performance_metrics.max_drawdown * 100).toFixed(1)}%</div>
                            <div className="metric-sub">控制良好</div>
                        </div>
                    </div>
                    
                    <div className="metric-card">
                        <div className="metric-icon">🎯</div>
                        <div className="metric-content">
                            <div className="metric-label">胜率</div>
                            <div className="metric-value">{(mockBacktestData.performance_metrics.win_rate * 100).toFixed(1)}%</div>
                            <div className="metric-sub">{mockBacktestData.performance_metrics.total_trades}笔交易</div>
                        </div>
                    </div>
                    
                    <div className="metric-card">
                        <div className="metric-icon">📊</div>
                        <div className="metric-content">
                            <div className="metric-label">波动率</div>
                            <div className="metric-value">{(mockBacktestData.performance_metrics.volatility * 100).toFixed(1)}%</div>
                            <div className="metric-sub">年化波动</div>
                        </div>
                    </div>
                    
                    <div className="metric-card">
                        <div className="metric-icon">💹</div>
                        <div className="metric-content">
                            <div className="metric-label">盈亏比</div>
                            <div className="metric-value">{mockBacktestData.performance_metrics.profit_factor.toFixed(1)}</div>
                            <div className="metric-sub">风险控制</div>
                        </div>
                    </div>
                </div>
            </div>

            {/* 图表选择器 */}
            <div className="chart-selector">
                <h3>📊 详细分析</h3>
                <div className="chart-tabs">
                    {chartTypes.map(chart => (
                        <button
                            key={chart.id}
                            className={`chart-tab ${activeChart === chart.id ? 'active' : ''}`}
                            onClick={() => setActiveChart(chart.id)}
                        >
                            <span className="tab-icon">{chart.icon}</span>
                            <span className="tab-name">{chart.name}</span>
                        </button>
                    ))}
                </div>
            </div>

            {/* 图表展示区域 */}
            <div className="chart-display">
                {renderActiveChart()}
            </div>

            {/* 策略评级 */}
            <div className="strategy-rating">
                <h3>⭐ 策略评级</h3>
                <div className="rating-grid">
                    <div className="rating-item">
                        <div className="rating-label">收益能力</div>
                        <div className="rating-stars">
                            {'★★★★☆'}
                        </div>
                        <div className="rating-score">4.2/5.0</div>
                    </div>
                    <div className="rating-item">
                        <div className="rating-label">风险控制</div>
                        <div className="rating-stars">
                            {'★★★★★'}
                        </div>
                        <div className="rating-score">4.8/5.0</div>
                    </div>
                    <div className="rating-item">
                        <div className="rating-label">稳定性</div>
                        <div className="rating-stars">
                            {'★★★★☆'}
                        </div>
                        <div className="rating-score">4.1/5.0</div>
                    </div>
                    <div className="rating-item">
                        <div className="rating-label">综合评级</div>
                        <div className="rating-stars">
                            {'★★★★☆'}
                        </div>
                        <div className="rating-score">4.4/5.0</div>
                    </div>
                </div>
                
                <div className="rating-summary">
                    <h4>📝 评级总结</h4>
                    <p>该策略在风险控制方面表现优秀，回撤控制良好，收益稳定性较高。建议在实盘部署时适当调整仓位管理和止损策略，以进一步提升收益风险比。</p>
                </div>
            </div>
        </div>
    );
};