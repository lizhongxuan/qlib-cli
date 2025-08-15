// 因子分析可视化组件
const { useState, useEffect } = React;

const FactorAnalysis = ({ 
    factorData = null,
    onNavigate = () => {} 
}) => {
    const [activeChart, setActiveChart] = useState('ic_series');
    const [selectedPeriod, setSelectedPeriod] = useState('monthly');
    const [showStatistics, setShowStatistics] = useState(true);

    // 模拟因子分析数据
    const mockFactorData = factorData || {
        name: '动量因子-V1.0',
        expression: 'Rank($close / Ref($close, 20) - 1)',
        statistics: {
            ic_mean: 0.0342,
            ic_std: 0.0876,
            ic_ir: 0.390,
            rank_ic_mean: 0.0456,
            rank_ic_std: 0.1123,
            rank_ic_ir: 0.406,
            turnover_mean: 0.234,
            coverage_mean: 0.892,
            valid_periods: 248
        },
        ic_series: Array.from({length: 60}, (_, i) => ({
            date: new Date(2022, 0, 1 + i * 7).toISOString().split('T')[0],
            ic: (Math.random() * 0.2 - 0.1).toFixed(4),
            rank_ic: (Math.random() * 0.25 - 0.125).toFixed(4),
            coverage: (Math.random() * 0.15 + 0.8).toFixed(3),
            turnover: (Math.random() * 0.1 + 0.2).toFixed(3)
        })),
        ic_distribution: Array.from({length: 20}, (_, i) => ({
            bin_start: (i * 0.02 - 0.2).toFixed(2),
            bin_end: ((i + 1) * 0.02 - 0.2).toFixed(2),
            frequency: Math.floor(Math.random() * 30 + 10),
            percentage: ((Math.floor(Math.random() * 30 + 10) / 248) * 100).toFixed(1)
        })),
        cumulative_ic: Array.from({length: 60}, (_, i) => {
            const baseValue = i * 0.005;
            return {
                date: new Date(2022, 0, 1 + i * 7).toISOString().split('T')[0],
                cum_ic: (baseValue + (Math.random() * 0.4 - 0.2)).toFixed(4),
                cum_rank_ic: (baseValue * 1.2 + (Math.random() * 0.5 - 0.25)).toFixed(4)
            };
        }),
        quantile_returns: Array.from({length: 10}, (_, i) => ({
            quantile: i + 1,
            mean_return: ((i + 1) * 0.002 + Math.random() * 0.008 - 0.004).toFixed(4),
            cumulative_return: ((i + 1) * 0.02 + Math.random() * 0.08 - 0.04).toFixed(3),
            sharpe: (Math.random() * 1.5 + 0.5).toFixed(2),
            max_drawdown: -(Math.random() * 0.15 + 0.05).toFixed(3)
        })),
        factor_exposure: {
            market: 0.12,
            size: -0.34,
            value: 0.23,
            growth: -0.18,
            momentum: 0.87,
            volatility: -0.45,
            liquidity: 0.15
        }
    };

    // 图表配置
    const chartTypes = [
        { id: 'ic_series', name: 'IC时序', icon: '📈' },
        { id: 'ic_distribution', name: 'IC分布', icon: '📊' },
        { id: 'cumulative_ic', name: '累积IC', icon: '📉' },
        { id: 'quantile_analysis', name: '分位数分析', icon: '🎯' },
        { id: 'factor_exposure', name: '因子暴露', icon: '🎪' }
    ];

    // 渲染IC时序图
    const renderICSeriesChart = () => {
        const data = mockFactorData.ic_series;
        const maxIC = Math.max(...data.map(d => Math.abs(parseFloat(d.ic))));
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>📈 IC时序走势</h3>
                    <div className="chart-controls">
                        <select 
                            value={selectedPeriod} 
                            onChange={(e) => setSelectedPeriod(e.target.value)}
                            className="period-selector"
                        >
                            <option value="daily">日度</option>
                            <option value="weekly">周度</option>
                            <option value="monthly">月度</option>
                        </select>
                    </div>
                </div>
                <div className="chart-content">
                    <svg viewBox="0 0 800 400" className="ic-chart">
                        {/* 网格 */}
                        <defs>
                            <pattern id="icGrid" width="50" height="40" patternUnits="userSpaceOnUse">
                                <path d="M 50 0 L 0 0 0 40" fill="none" stroke="#f5f5f5" strokeWidth="1"/>
                            </pattern>
                        </defs>
                        <rect width="800" height="400" fill="url(#icGrid)" />
                        
                        {/* 零线 */}
                        <line x1="50" y1="200" x2="750" y2="200" stroke="#ddd" strokeWidth="2" strokeDasharray="5,5"/>
                        
                        {/* IC曲线 */}
                        <polyline
                            points={data.map((d, i) => 
                                `${50 + i * 700/data.length},${200 - parseFloat(d.ic) * 1500}`
                            ).join(' ')}
                            fill="none"
                            stroke="#1890ff"
                            strokeWidth="2"
                        />
                        
                        {/* Rank IC曲线 */}
                        <polyline
                            points={data.map((d, i) => 
                                `${50 + i * 700/data.length},${200 - parseFloat(d.rank_ic) * 1200}`
                            ).join(' ')}
                            fill="none"
                            stroke="#52c41a"
                            strokeWidth="2"
                            strokeDasharray="3,3"
                        />
                        
                        {/* Y轴标签 */}
                        <text x="10" y="50" fill="#666" fontSize="12">0.1</text>
                        <text x="10" y="105" fill="#666" fontSize="12">0.05</text>
                        <text x="10" y="205" fill="#666" fontSize="12">0</text>
                        <text x="10" y="305" fill="#666" fontSize="12">-0.05</text>
                        <text x="10" y="360" fill="#666" fontSize="12">-0.1</text>
                        
                        {/* X轴标签 */}
                        {data.filter((_, i) => i % 10 === 0).map((d, i) => (
                            <text key={i} x={50 + i * 10 * 700/data.length - 20} y="390" fill="#666" fontSize="10">
                                {d.date.substring(5)}
                            </text>
                        ))}
                    </svg>
                    
                    <div className="chart-legend">
                        <div className="legend-item">
                            <div className="legend-color" style={{backgroundColor: '#1890ff'}}></div>
                            <span>IC值</span>
                        </div>
                        <div className="legend-item">
                            <div className="legend-color" style={{backgroundColor: '#52c41a', opacity: 0.7}}></div>
                            <span>Rank IC</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染IC分布图
    const renderICDistributionChart = () => {
        const data = mockFactorData.ic_distribution;
        const maxFreq = Math.max(...data.map(d => d.frequency));
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>📊 IC分布直方图</h3>
                </div>
                <div className="chart-content">
                    <svg viewBox="0 0 800 400" className="distribution-chart">
                        <rect width="800" height="400" fill="#fafafa" />
                        
                        {/* 分布柱状图 */}
                        {data.map((d, i) => (
                            <g key={i}>
                                <rect
                                    x={50 + i * 700/data.length}
                                    y={350 - (d.frequency / maxFreq) * 300}
                                    width={700/data.length - 2}
                                    height={(d.frequency / maxFreq) * 300}
                                    fill={parseFloat(d.bin_start) >= 0 ? "#52c41a" : "#ff4d4f"}
                                    opacity="0.8"
                                />
                                <text 
                                    x={50 + i * 700/data.length + (700/data.length)/2} 
                                    y={365} 
                                    fill="#666" 
                                    fontSize="8" 
                                    textAnchor="middle"
                                >
                                    {d.bin_start}
                                </text>
                            </g>
                        ))}
                        
                        {/* Y轴标签 */}
                        <text x="10" y="60" fill="#666" fontSize="12">{maxFreq}</text>
                        <text x="10" y="210" fill="#666" fontSize="12">{Math.round(maxFreq/2)}</text>
                        <text x="10" y="355" fill="#666" fontSize="12">0</text>
                        
                        {/* 轴线 */}
                        <line x1="50" y1="350" x2="750" y2="350" stroke="#333" strokeWidth="2"/>
                        <line x1="50" y1="350" x2="50" y2="50" stroke="#333" strokeWidth="2"/>
                    </svg>
                    
                    <div className="distribution-stats">
                        <div className="stat-item">
                            <span className="stat-label">均值:</span>
                            <span className="stat-value">{mockFactorData.statistics.ic_mean.toFixed(4)}</span>
                        </div>
                        <div className="stat-item">
                            <span className="stat-label">标准差:</span>
                            <span className="stat-value">{mockFactorData.statistics.ic_std.toFixed(4)}</span>
                        </div>
                        <div className="stat-item">
                            <span className="stat-label">信息比率:</span>
                            <span className="stat-value">{mockFactorData.statistics.ic_ir.toFixed(3)}</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染累积IC图
    const renderCumulativeICChart = () => {
        const data = mockFactorData.cumulative_ic;
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>📉 累积IC走势</h3>
                </div>
                <div className="chart-content">
                    <svg viewBox="0 0 800 300" className="cumulative-chart">
                        <rect width="800" height="300" fill="#fafafa" />
                        
                        {/* 累积IC曲线 */}
                        <polyline
                            points={data.map((d, i) => 
                                `${50 + i * 700/data.length},${150 - parseFloat(d.cum_ic) * 300}`
                            ).join(' ')}
                            fill="none"
                            stroke="#722ed1"
                            strokeWidth="3"
                        />
                        
                        {/* 累积Rank IC曲线 */}
                        <polyline
                            points={data.map((d, i) => 
                                `${50 + i * 700/data.length},${150 - parseFloat(d.cum_rank_ic) * 250}`
                            ).join(' ')}
                            fill="none"
                            stroke="#fa8c16"
                            strokeWidth="2"
                            strokeDasharray="4,4"
                        />
                        
                        {/* 零线 */}
                        <line x1="50" y1="150" x2="750" y2="150" stroke="#ddd" strokeWidth="1" strokeDasharray="5,5"/>
                    </svg>
                </div>
            </div>
        );
    };

    // 渲染分位数分析
    const renderQuantileAnalysis = () => {
        const data = mockFactorData.quantile_returns;
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>🎯 分位数收益分析</h3>
                </div>
                <div className="chart-content">
                    <div className="quantile-table">
                        <table>
                            <thead>
                                <tr>
                                    <th>分位数</th>
                                    <th>平均收益</th>
                                    <th>累积收益</th>
                                    <th>夏普比率</th>
                                    <th>最大回撤</th>
                                </tr>
                            </thead>
                            <tbody>
                                {data.map(d => (
                                    <tr key={d.quantile}>
                                        <td>Q{d.quantile}</td>
                                        <td className={parseFloat(d.mean_return) > 0 ? 'positive' : 'negative'}>
                                            {(parseFloat(d.mean_return) * 100).toFixed(2)}%
                                        </td>
                                        <td className={parseFloat(d.cumulative_return) > 0 ? 'positive' : 'negative'}>
                                            {(parseFloat(d.cumulative_return) * 100).toFixed(1)}%
                                        </td>
                                        <td>{d.sharpe}</td>
                                        <td className="negative">{(parseFloat(d.max_drawdown) * 100).toFixed(1)}%</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                    
                    {/* 分位数收益柱状图 */}
                    <svg viewBox="0 0 600 250" className="quantile-chart">
                        <rect width="600" height="250" fill="#fafafa" />
                        
                        {data.map((d, i) => (
                            <g key={i}>
                                <rect
                                    x={50 + i * 50}
                                    y={200 - Math.abs(parseFloat(d.mean_return)) * 50000}
                                    width="40"
                                    height={Math.abs(parseFloat(d.mean_return)) * 50000}
                                    fill={parseFloat(d.mean_return) > 0 ? "#52c41a" : "#ff4d4f"}
                                />
                                <text 
                                    x={70 + i * 50} 
                                    y={220} 
                                    fill="#666" 
                                    fontSize="10" 
                                    textAnchor="middle"
                                >
                                    Q{d.quantile}
                                </text>
                            </g>
                        ))}
                        
                        {/* 零线 */}
                        <line x1="50" y1="200" x2="550" y2="200" stroke="#333" strokeWidth="1"/>
                    </svg>
                </div>
            </div>
        );
    };

    // 渲染因子暴露分析
    const renderFactorExposure = () => {
        const exposure = mockFactorData.factor_exposure;
        const factors = Object.entries(exposure);
        
        return (
            <div className="chart-container">
                <div className="chart-header">
                    <h3>🎪 因子暴露分析</h3>
                </div>
                <div className="chart-content">
                    <div className="exposure-chart">
                        {factors.map(([factor, value]) => (
                            <div key={factor} className="exposure-item">
                                <div className="exposure-label">{factor}</div>
                                <div className="exposure-bar-container">
                                    <div 
                                        className={`exposure-bar ${value > 0 ? 'positive' : 'negative'}`}
                                        style={{
                                            width: `${Math.abs(value) * 100}%`,
                                            marginLeft: value < 0 ? `${(1 + value) * 100}%` : '50%'
                                        }}
                                    ></div>
                                </div>
                                <div className={`exposure-value ${value > 0 ? 'positive' : 'negative'}`}>
                                    {value.toFixed(2)}
                                </div>
                            </div>
                        ))}
                    </div>
                    
                    <div className="exposure-legend">
                        <div className="legend-item">
                            <div className="legend-color positive"></div>
                            <span>正暴露</span>
                        </div>
                        <div className="legend-item">
                            <div className="legend-color negative"></div>
                            <span>负暴露</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染当前选中的图表
    const renderActiveChart = () => {
        switch(activeChart) {
            case 'ic_series':
                return renderICSeriesChart();
            case 'ic_distribution':
                return renderICDistributionChart();
            case 'cumulative_ic':
                return renderCumulativeICChart();
            case 'quantile_analysis':
                return renderQuantileAnalysis();
            case 'factor_exposure':
                return renderFactorExposure();
            default:
                return renderICSeriesChart();
        }
    };

    return (
        <div className="factor-analysis">
            <div className="analysis-header">
                <h1>📊 因子分析报告</h1>
                <div className="factor-info-summary">
                    <h2>{mockFactorData.name}</h2>
                    <div className="expression-display">{mockFactorData.expression}</div>
                </div>
            </div>

            {/* 统计指标面板 */}
            {showStatistics && (
                <div className="statistics-panel">
                    <h3>📈 核心统计指标</h3>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <div className="stat-label">IC均值</div>
                            <div className={`stat-value ${mockFactorData.statistics.ic_mean > 0 ? 'positive' : 'negative'}`}>
                                {mockFactorData.statistics.ic_mean.toFixed(4)}
                            </div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">IC信息比率</div>
                            <div className="stat-value">{mockFactorData.statistics.ic_ir.toFixed(3)}</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">Rank IC均值</div>
                            <div className={`stat-value ${mockFactorData.statistics.rank_ic_mean > 0 ? 'positive' : 'negative'}`}>
                                {mockFactorData.statistics.rank_ic_mean.toFixed(4)}
                            </div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">Rank IC信息比率</div>
                            <div className="stat-value">{mockFactorData.statistics.rank_ic_ir.toFixed(3)}</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">平均换手率</div>
                            <div className="stat-value">{(mockFactorData.statistics.turnover_mean * 100).toFixed(1)}%</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">平均覆盖率</div>
                            <div className="stat-value">{(mockFactorData.statistics.coverage_mean * 100).toFixed(1)}%</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-label">有效期数</div>
                            <div className="stat-value">{mockFactorData.statistics.valid_periods}</div>
                        </div>
                    </div>
                </div>
            )}

            {/* 图表选择器 */}
            <div className="chart-selector">
                <h3>📊 分析图表</h3>
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

            {/* 操作按钮 */}
            <div className="analysis-actions">
                <button className="btn-primary" onClick={() => onNavigate('model')}>
                    🤖 使用此因子训练模型
                </button>
                <button className="btn-secondary" onClick={() => onNavigate('backtest')}>
                    📈 因子策略回测
                </button>
                <button className="btn-secondary">
                    📄 导出分析报告
                </button>
                <button className="btn-secondary">
                    💾 保存分析结果
                </button>
            </div>
        </div>
    );
};