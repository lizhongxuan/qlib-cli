// 结果分析中心 - 统一的qlib分析和报告平台
const { useState, useEffect } = React;

const ResultsAnalysis = ({ 
    tasks = [],
    models = [],
    savedFactors = [],
    onNavigate = () => {} 
}) => {
    const [activeTab, setActiveTab] = useState('overview');
    const [selectedResults, setSelectedResults] = useState([]);
    const [analysisMode, setAnalysisMode] = useState('single');

    // 分析标签页配置
    const analysisTabs = [
        { 
            key: 'overview', 
            label: '结果概览', 
            icon: '📋', 
            desc: 'qlib工作流运行结果总览' 
        },
        { 
            key: 'model', 
            label: '模型分析', 
            icon: '🤖', 
            desc: '模型性能和因子重要性分析' 
        },
        { 
            key: 'strategy', 
            label: '策略分析', 
            icon: '📈', 
            desc: '策略收益和风险指标分析' 
        },
        { 
            key: 'comparison', 
            label: '对比分析', 
            icon: '⚖️', 
            desc: '多策略和多模型对比' 
        },
        { 
            key: 'report', 
            label: '研究报告', 
            icon: '📄', 
            desc: '生成和导出分析报告' 
        }
    ];

    // 过滤有效的任务结果
    const validTasks = tasks.filter(task => 
        task.type === 'qlib_workflow' && task.status === 'completed' && task.results
    );

    // 模拟更多测试数据
    const mockResults = [
        {
            id: 'result_001',
            name: 'LightGBM-Alpha158-CSI300',
            type: 'qlib_workflow',
            date: '2024-01-15',
            pipeline: {
                data: { market: 'csi300', benchmark: 'SH000300' },
                model: { class: 'LightGBM' },
                features: { handler: 'Alpha158' },
                strategy: { class: 'TopkDropoutStrategy', params: { topk: 50 } }
            },
            results: {
                model_performance: {
                    train_ic: 0.0456, valid_ic: 0.0398, test_ic: 0.0367,
                    train_rank_ic: 0.0612, valid_rank_ic: 0.0534, test_rank_ic: 0.0489
                },
                strategy_performance: {
                    annual_return: 0.1847, benchmark_return: 0.0956, excess_return: 0.0891,
                    volatility: 0.1623, sharpe_ratio: 1.138, max_drawdown: -0.0847, win_rate: 0.574
                },
                factor_analysis: {
                    top_factors: [
                        { name: 'RESI5', ic: 0.0423, importance: 0.125 },
                        { name: 'WVMA5', ic: 0.0389, importance: 0.098 },
                        { name: 'RSQR10', ic: 0.0356, importance: 0.087 }
                    ]
                }
            }
        },
        {
            id: 'result_002',
            name: 'XGBoost-Alpha360-CSI500',
            type: 'qlib_workflow',
            date: '2024-01-12',
            pipeline: {
                data: { market: 'csi500', benchmark: 'SH000905' },
                model: { class: 'XGBoost' },
                features: { handler: 'Alpha360' },
                strategy: { class: 'TopkDropoutStrategy', params: { topk: 30 } }
            },
            results: {
                model_performance: {
                    train_ic: 0.0398, valid_ic: 0.0345, test_ic: 0.0312,
                    train_rank_ic: 0.0567, valid_rank_ic: 0.0478, test_rank_ic: 0.0434
                },
                strategy_performance: {
                    annual_return: 0.2134, benchmark_return: 0.0823, excess_return: 0.1311,
                    volatility: 0.1891, sharpe_ratio: 1.129, max_drawdown: -0.1023, win_rate: 0.589
                },
                factor_analysis: {
                    top_factors: [
                        { name: 'QTURN20', ic: 0.0398, importance: 0.142 },
                        { name: 'BETA60', ic: 0.0367, importance: 0.108 },
                        { name: 'VSTD10', ic: 0.0334, importance: 0.095 }
                    ]
                }
            }
        }
    ];

    // 合并真实任务和模拟数据
    const allResults = [...validTasks, ...mockResults];

    // 处理结果选择
    const handleResultSelection = (result, isSelected) => {
        if (isSelected) {
            setSelectedResults(prev => [...prev, result]);
        } else {
            setSelectedResults(prev => prev.filter(r => r.id !== result.id));
        }
    };

    // 计算汇总统计
    const calculateSummaryStats = (results) => {
        if (results.length === 0) return null;

        const avgReturn = results.reduce((sum, r) => sum + (r.results?.strategy_performance?.annual_return || 0), 0) / results.length;
        const avgSharpe = results.reduce((sum, r) => sum + (r.results?.strategy_performance?.sharpe_ratio || 0), 0) / results.length;
        const avgIC = results.reduce((sum, r) => sum + (r.results?.model_performance?.test_ic || 0), 0) / results.length;
        const bestReturn = Math.max(...results.map(r => r.results?.strategy_performance?.annual_return || 0));
        const worstDrawdown = Math.min(...results.map(r => r.results?.strategy_performance?.max_drawdown || 0));

        return {
            avgReturn: avgReturn,
            avgSharpe: avgSharpe,
            avgIC: avgIC,
            bestReturn: bestReturn,
            worstDrawdown: worstDrawdown,
            totalResults: results.length
        };
    };

    const summaryStats = calculateSummaryStats(allResults);

    // 渲染结果概览
    const renderOverviewTab = () => (
        <div className="overview-content">
            <div className="overview-header">
                <h3>📋 qlib研究结果概览</h3>
                <div className="overview-stats">
                    {summaryStats && (
                        <>
                            <div className="stat-item">
                                <span className="stat-label">总运行次数</span>
                                <span className="stat-value">{summaryStats.totalResults}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">平均年化收益</span>
                                <span className="stat-value positive">
                                    {(summaryStats.avgReturn * 100).toFixed(2)}%
                                </span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">平均夏普比率</span>
                                <span className="stat-value">{summaryStats.avgSharpe.toFixed(3)}</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-label">平均测试IC</span>
                                <span className="stat-value">{summaryStats.avgIC.toFixed(4)}</span>
                            </div>
                        </>
                    )}
                </div>
            </div>

            {allResults.length === 0 ? (
                <div className="empty-results">
                    <div className="empty-icon">📊</div>
                    <div className="empty-title">暂无分析结果</div>
                    <div className="empty-desc">
                        请先运行qlib量化研究工作流以生成分析结果
                    </div>
                    <button 
                        className="btn-primary"
                        onClick={() => onNavigate('workflow')}
                    >
                        🚀 开始研究
                    </button>
                </div>
            ) : (
                <div className="results-grid">
                    {allResults.map(result => (
                        <div key={result.id} className="result-card">
                            <div className="result-header">
                                <div className="result-info">
                                    <h4>{result.name}</h4>
                                    <div className="result-meta">
                                        <span>{result.date}</span>
                                        <span>•</span>
                                        <span>{result.pipeline?.data?.market?.toUpperCase()}</span>
                                        <span>•</span>
                                        <span>{result.pipeline?.model?.class}</span>
                                    </div>
                                </div>
                                <div className="result-selector">
                                    <input
                                        type="checkbox"
                                        checked={selectedResults.some(r => r.id === result.id)}
                                        onChange={(e) => handleResultSelection(result, e.target.checked)}
                                    />
                                </div>
                            </div>

                            <div className="result-metrics">
                                <div className="metrics-row">
                                    <div className="metric">
                                        <span className="metric-label">年化收益</span>
                                        <span className="metric-value positive">
                                            {(result.results?.strategy_performance?.annual_return * 100).toFixed(2)}%
                                        </span>
                                    </div>
                                    <div className="metric">
                                        <span className="metric-label">夏普比率</span>
                                        <span className="metric-value">
                                            {result.results?.strategy_performance?.sharpe_ratio.toFixed(3)}
                                        </span>
                                    </div>
                                </div>
                                <div className="metrics-row">
                                    <div className="metric">
                                        <span className="metric-label">测试IC</span>
                                        <span className="metric-value">
                                            {result.results?.model_performance?.test_ic.toFixed(4)}
                                        </span>
                                    </div>
                                    <div className="metric">
                                        <span className="metric-label">最大回撤</span>
                                        <span className="metric-value negative">
                                            {(result.results?.strategy_performance?.max_drawdown * 100).toFixed(2)}%
                                        </span>
                                    </div>
                                </div>
                            </div>

                            <div className="result-actions">
                                <button 
                                    className="btn-sm btn-primary"
                                    onClick={() => {
                                        setSelectedResults([result]);
                                        setActiveTab('model');
                                    }}
                                >
                                    详细分析
                                </button>
                                <button className="btn-sm btn-secondary">
                                    导出结果
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );

    // 渲染模型分析标签页
    const renderModelTab = () => {
        const selectedResult = selectedResults[0];
        if (!selectedResult) {
            return (
                <div className="no-selection">
                    <div className="placeholder-icon">🤖</div>
                    <h4>请选择要分析的模型结果</h4>
                    <p>从结果概览中选择一个或多个qlib运行结果进行深入分析</p>
                    <button 
                        className="btn-primary"
                        onClick={() => setActiveTab('overview')}
                    >
                        返回概览
                    </button>
                </div>
            );
        }

        const modelPerf = selectedResult.results?.model_performance;
        const factorAnalysis = selectedResult.results?.factor_analysis;

        return (
            <div className="model-analysis">
                <div className="analysis-header">
                    <h3>🤖 模型性能分析</h3>
                    <div className="model-info">
                        <span>{selectedResult.name}</span>
                        <span>•</span>
                        <span>{selectedResult.pipeline?.model?.class}</span>
                        <span>•</span>
                        <span>{selectedResult.pipeline?.features?.handler}</span>
                    </div>
                </div>

                <div className="analysis-sections">
                    <div className="analysis-section">
                        <h4>📊 IC分析</h4>
                        <div className="ic-metrics">
                            <div className="ic-chart">
                                <div className="chart-header">
                                    <span>IC值对比</span>
                                </div>
                                <div className="ic-bars">
                                    <div className="ic-bar">
                                        <span className="bar-label">训练集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.train_ic) * 1000}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.train_ic}</span>
                                    </div>
                                    <div className="ic-bar">
                                        <span className="bar-label">验证集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.valid_ic) * 1000}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.valid_ic}</span>
                                    </div>
                                    <div className="ic-bar">
                                        <span className="bar-label">测试集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.test_ic) * 1000}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.test_ic}</span>
                                    </div>
                                </div>
                            </div>

                            <div className="rank-ic-chart">
                                <div className="chart-header">
                                    <span>Rank IC对比</span>
                                </div>
                                <div className="ic-bars">
                                    <div className="ic-bar">
                                        <span className="bar-label">训练集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.train_rank_ic) * 800}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.train_rank_ic}</span>
                                    </div>
                                    <div className="ic-bar">
                                        <span className="bar-label">验证集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.valid_rank_ic) * 800}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.valid_rank_ic}</span>
                                    </div>
                                    <div className="ic-bar">
                                        <span className="bar-label">测试集</span>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${Math.abs(modelPerf.test_rank_ic) * 800}px` }}
                                            ></div>
                                        </div>
                                        <span className="bar-value">{modelPerf.test_rank_ic}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="analysis-section">
                        <h4>🎯 因子重要性分析</h4>
                        <div className="factor-importance">
                            {factorAnalysis?.top_factors?.map((factor, index) => (
                                <div key={index} className="factor-item">
                                    <div className="factor-info">
                                        <span className="factor-name">{factor.name}</span>
                                        <span className="factor-ic">IC: {factor.ic}</span>
                                    </div>
                                    <div className="importance-bar">
                                        <div 
                                            className="importance-fill"
                                            style={{ width: `${factor.importance * 100}%` }}
                                        ></div>
                                    </div>
                                    <span className="importance-value">
                                        {(factor.importance * 100).toFixed(1)}%
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染策略分析标签页
    const renderStrategyTab = () => {
        const selectedResult = selectedResults[0];
        if (!selectedResult) {
            return (
                <div className="no-selection">
                    <div className="placeholder-icon">📈</div>
                    <h4>请选择要分析的策略结果</h4>
                    <p>从结果概览中选择一个qlib策略运行结果进行分析</p>
                </div>
            );
        }

        const strategyPerf = selectedResult.results?.strategy_performance;

        return (
            <div className="strategy-analysis">
                <div className="analysis-header">
                    <h3>📈 策略绩效分析</h3>
                    <div className="strategy-info">
                        <span>{selectedResult.name}</span>
                        <span>•</span>
                        <span>{selectedResult.pipeline?.strategy?.class}</span>
                        <span>•</span>
                        <span>TopK: {selectedResult.pipeline?.strategy?.params?.topk}</span>
                    </div>
                </div>

                <div className="performance-metrics">
                    <div className="metrics-grid">
                        <div className="perf-card">
                            <div className="perf-icon">💰</div>
                            <div className="perf-content">
                                <div className="perf-label">年化收益率</div>
                                <div className="perf-value positive">
                                    {(strategyPerf.annual_return * 100).toFixed(2)}%
                                </div>
                                <div className="perf-sub">
                                    vs 基准: +{(strategyPerf.excess_return * 100).toFixed(2)}%
                                </div>
                            </div>
                        </div>

                        <div className="perf-card">
                            <div className="perf-icon">📊</div>
                            <div className="perf-content">
                                <div className="perf-label">夏普比率</div>
                                <div className="perf-value">
                                    {strategyPerf.sharpe_ratio.toFixed(3)}
                                </div>
                                <div className="perf-sub">
                                    风险调整后收益
                                </div>
                            </div>
                        </div>

                        <div className="perf-card">
                            <div className="perf-icon">📉</div>
                            <div className="perf-content">
                                <div className="perf-label">最大回撤</div>
                                <div className="perf-value negative">
                                    {(strategyPerf.max_drawdown * 100).toFixed(2)}%
                                </div>
                                <div className="perf-sub">
                                    风险控制指标
                                </div>
                            </div>
                        </div>

                        <div className="perf-card">
                            <div className="perf-icon">🎯</div>
                            <div className="perf-content">
                                <div className="perf-label">胜率</div>
                                <div className="perf-value">
                                    {(strategyPerf.win_rate * 100).toFixed(1)}%
                                </div>
                                <div className="perf-sub">
                                    盈利交易占比
                                </div>
                            </div>
                        </div>

                        <div className="perf-card">
                            <div className="perf-icon">📈</div>
                            <div className="perf-content">
                                <div className="perf-label">波动率</div>
                                <div className="perf-value">
                                    {(strategyPerf.volatility * 100).toFixed(2)}%
                                </div>
                                <div className="perf-sub">
                                    收益率标准差
                                </div>
                            </div>
                        </div>

                        <div className="perf-card">
                            <div className="perf-icon">⚡</div>
                            <div className="perf-content">
                                <div className="perf-label">信息比率</div>
                                <div className="perf-value">
                                    {strategyPerf.information_ratio?.toFixed(3) || 'N/A'}
                                </div>
                                <div className="perf-sub">
                                    超额收益质量
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="risk-analysis">
                    <h4>🛡️ 风险分析</h4>
                    <div className="risk-metrics">
                        <div className="risk-item">
                            <span className="risk-label">收益风险比</span>
                            <div className="risk-bar">
                                <div 
                                    className="risk-fill positive"
                                    style={{ width: `${(strategyPerf.annual_return / strategyPerf.volatility) * 20}%` }}
                                ></div>
                            </div>
                            <span className="risk-value">
                                {(strategyPerf.annual_return / strategyPerf.volatility).toFixed(2)}
                            </span>
                        </div>
                        <div className="risk-item">
                            <span className="risk-label">回撤控制</span>
                            <div className="risk-bar">
                                <div 
                                    className="risk-fill negative"
                                    style={{ width: `${Math.abs(strategyPerf.max_drawdown) * 500}%` }}
                                ></div>
                            </div>
                            <span className="risk-value negative">
                                {(strategyPerf.max_drawdown * 100).toFixed(2)}%
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    // 渲染对比分析标签页
    const renderComparisonTab = () => (
        <div className="comparison-analysis">
            <div className="comparison-header">
                <h3>⚖️ 多策略对比分析</h3>
                <div className="comparison-controls">
                    <span>已选择 {selectedResults.length} 个结果</span>
                    {selectedResults.length < 2 && (
                        <div className="comparison-hint">
                            请从结果概览中选择至少2个结果进行对比
                        </div>
                    )}
                </div>
            </div>

            {selectedResults.length < 2 ? (
                <div className="insufficient-selection">
                    <div className="placeholder-icon">⚖️</div>
                    <h4>需要选择多个结果进行对比</h4>
                    <p>请返回结果概览，选择2个或更多qlib运行结果</p>
                    <button 
                        className="btn-primary"
                        onClick={() => setActiveTab('overview')}
                    >
                        返回选择
                    </button>
                </div>
            ) : (
                <div className="comparison-content">
                    <div className="comparison-table">
                        <table>
                            <thead>
                                <tr>
                                    <th>指标</th>
                                    {selectedResults.map(result => (
                                        <th key={result.id}>{result.name}</th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>年化收益率</td>
                                    {selectedResults.map(result => (
                                        <td key={result.id} className="positive">
                                            {(result.results?.strategy_performance?.annual_return * 100).toFixed(2)}%
                                        </td>
                                    ))}
                                </tr>
                                <tr>
                                    <td>夏普比率</td>
                                    {selectedResults.map(result => (
                                        <td key={result.id}>
                                            {result.results?.strategy_performance?.sharpe_ratio.toFixed(3)}
                                        </td>
                                    ))}
                                </tr>
                                <tr>
                                    <td>最大回撤</td>
                                    {selectedResults.map(result => (
                                        <td key={result.id} className="negative">
                                            {(result.results?.strategy_performance?.max_drawdown * 100).toFixed(2)}%
                                        </td>
                                    ))}
                                </tr>
                                <tr>
                                    <td>测试IC</td>
                                    {selectedResults.map(result => (
                                        <td key={result.id}>
                                            {result.results?.model_performance?.test_ic.toFixed(4)}
                                        </td>
                                    ))}
                                </tr>
                                <tr>
                                    <td>胜率</td>
                                    {selectedResults.map(result => (
                                        <td key={result.id}>
                                            {(result.results?.strategy_performance?.win_rate * 100).toFixed(1)}%
                                        </td>
                                    ))}
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <div className="comparison-charts">
                        <div className="comparison-chart">
                            <h4>收益率对比</h4>
                            <div className="chart-bars">
                                {selectedResults.map(result => (
                                    <div key={result.id} className="chart-bar">
                                        <div className="bar-label">{result.name}</div>
                                        <div className="bar-container">
                                            <div 
                                                className="bar-fill positive"
                                                style={{ width: `${result.results?.strategy_performance?.annual_return * 500}px` }}
                                            ></div>
                                        </div>
                                        <div className="bar-value">
                                            {(result.results?.strategy_performance?.annual_return * 100).toFixed(2)}%
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );

    // 渲染报告生成标签页
    const renderReportTab = () => (
        <div className="report-generation">
            <div className="report-header">
                <h3>📄 研究报告生成</h3>
                <div className="report-controls">
                    <select className="form-select">
                        <option>完整研究报告</option>
                        <option>模型分析报告</option>
                        <option>策略绩效报告</option>
                        <option>风险分析报告</option>
                    </select>
                    <button className="btn-primary">📊 生成报告</button>
                </div>
            </div>

            <div className="report-preview">
                <div className="report-section">
                    <h4>🎯 执行摘要</h4>
                    <p>
                        本次qlib量化研究共运行{allResults.length}个策略配置，
                        平均年化收益率{summaryStats ? (summaryStats.avgReturn * 100).toFixed(2) : 'N/A'}%，
                        平均夏普比率{summaryStats ? summaryStats.avgSharpe.toFixed(3) : 'N/A'}。
                        最佳策略实现了{summaryStats ? (summaryStats.bestReturn * 100).toFixed(2) : 'N/A'}%的年化收益率。
                    </p>
                </div>

                <div className="report-section">
                    <h4>📊 关键发现</h4>
                    <ul>
                        <li>LightGBM模型在Alpha158因子集上表现最佳</li>
                        <li>CSI300股票池的策略稳定性优于CSI500</li>
                        <li>TopK选股策略的风险收益比表现优异</li>
                        <li>价量类因子的重要性显著高于其他类型</li>
                    </ul>
                </div>

                <div className="report-section">
                    <h4>🔄 下一步建议</h4>
                    <ul>
                        <li>进一步优化Top5重要因子的组合权重</li>
                        <li>测试更多时间窗口的数据集分割方案</li>
                        <li>探索集成学习模型的应用可能</li>
                        <li>进行实盘模拟测试验证</li>
                    </ul>
                </div>
            </div>

            <div className="export-options">
                <h4>📤 导出选项</h4>
                <div className="export-buttons">
                    <button className="btn-secondary">📄 PDF报告</button>
                    <button className="btn-secondary">📊 Excel数据</button>
                    <button className="btn-secondary">📋 PowerPoint</button>
                    <button className="btn-secondary">🔗 分享链接</button>
                </div>
            </div>
        </div>
    );

    // 渲染当前标签页内容
    const renderActiveTab = () => {
        switch(activeTab) {
            case 'overview': return renderOverviewTab();
            case 'model': return renderModelTab();
            case 'strategy': return renderStrategyTab();
            case 'comparison': return renderComparisonTab();
            case 'report': return renderReportTab();
            default: return renderOverviewTab();
        }
    };

    return (
        <div className="results-analysis">
            <div className="analysis-header">
                <h1>📊 结果分析中心</h1>
                <div className="header-subtitle">
                    qlib量化研究结果的统一分析和报告平台
                </div>
            </div>

            {/* 分析标签页导航 */}
            <div className="analysis-tabs">
                {analysisTabs.map(tab => (
                    <button
                        key={tab.key}
                        className={`analysis-tab ${activeTab === tab.key ? 'active' : ''}`}
                        onClick={() => setActiveTab(tab.key)}
                    >
                        <span className="tab-icon">{tab.icon}</span>
                        <div className="tab-content">
                            <div className="tab-label">{tab.label}</div>
                            <div className="tab-desc">{tab.desc}</div>
                        </div>
                    </button>
                ))}
            </div>

            {/* 主要内容区域 */}
            <div className="analysis-content">
                {renderActiveTab()}
            </div>
        </div>
    );
};