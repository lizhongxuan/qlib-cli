// 策略回测组件
const { useState, useEffect, useCallback } = React;

const StrategyBacktest = ({ models, datasets, onAddTask }) => {
    const [activeTab, setActiveTab] = useState('config');
    const [backtestResults, setBacktestResults] = useState([]);
    const [showForm, setShowForm] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        model: '',
        dataset: '',
        strategy: 'topk',
        startDate: '2022-01-01',
        endDate: '2023-12-31',
        initialCash: 1000000,
        topK: 30,
        rebalanceFreq: 'daily',
        commission: 0.0003
    });

    const strategyTypes = [
        { value: 'topk', label: 'TopK策略', desc: '选择评分最高的K只股票' },
        { value: 'long_short', label: '多空策略', desc: '同时做多和做空' },
        { value: 'market_neutral', label: '市场中性', desc: '保持市场敞口为零' }
    ];

    const handleBacktest = () => {
        const backtestId = `backtest_${Date.now()}`;
        const newBacktest = {
            id: backtestId,
            name: formData.name || `${formData.strategy}-回测`,
            strategy: formData.strategy,
            model: formData.model,
            dataset: formData.dataset,
            status: 'running',
            progress: 0,
            startTime: new Date().toLocaleString(),
            results: null
        };

        setBacktestResults(prev => [...prev, newBacktest]);
        setActiveTab('results');

        onAddTask({
            name: `策略回测: ${newBacktest.name}`,
            type: 'backtest'
        });

        // 模拟回测过程
        simulateBacktest(backtestId);
        setShowForm(false);
    };

    const simulateBacktest = (backtestId) => {
        let progress = 0;
        const interval = setInterval(() => {
            progress += Math.random() * 10;
            
            setBacktestResults(prev => prev.map(bt => 
                bt.id === backtestId ? {
                    ...bt,
                    progress: Math.min(progress, 100)
                } : bt
            ));

            if (progress >= 100) {
                const mockResults = {
                    totalReturn: '23.5%',
                    annualReturn: '18.2%',
                    sharpeRatio: 1.85,
                    maxDrawdown: '8.5%',
                    winRate: '62.3%',
                    volatility: '12.4%',
                    trades: 1250,
                    profitFactor: 2.1
                };

                setBacktestResults(prev => prev.map(bt => 
                    bt.id === backtestId ? {
                        ...bt,
                        status: 'completed',
                        progress: 100,
                        results: mockResults
                    } : bt
                ));
                clearInterval(interval);
            }
        }, 800);
    };

    return (
        <div className="strategy-backtest">
            <div className="page-header">
                <h1>📈 策略回测</h1>
                <div className="header-actions">
                    <button className="btn-secondary">📊 策略对比</button>
                    <button className="btn-primary" onClick={() => setShowForm(true)}>
                        + 新建回测
                    </button>
                </div>
            </div>

            {/* 标签页导航 */}
            <div className="tab-navigation">
                <button 
                    className={`tab-btn ${activeTab === 'config' ? 'active' : ''}`}
                    onClick={() => setActiveTab('config')}
                >
                    ⚙️ 策略配置
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'results' ? 'active' : ''}`}
                    onClick={() => setActiveTab('results')}
                >
                    📈 回测结果
                    <span className="tab-badge">{backtestResults.length}</span>
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'analysis' ? 'active' : ''}`}
                    onClick={() => setActiveTab('analysis')}
                >
                    📊 分析报告
                </button>
            </div>

            {/* 标签页内容 */}
            <div className="tab-content">
                {activeTab === 'config' && (
                    <div className="config-content">
                        <div className="config-wizard">
                            <div className="wizard-steps">
                                <div className="step active">1. 基础配置</div>
                                <div className="step">2. 策略参数</div>
                                <div className="step">3. 风控设置</div>
                                <div className="step">4. 执行回测</div>
                            </div>
                            
                            <div className="config-cards">
                                <div className="config-card">
                                    <div className="card-header">
                                        <h3>📋 基础配置</h3>
                                        <p>选择模型、数据集和回测时间范围</p>
                                    </div>
                                    <div className="card-content">
                                        <div className="form-group">
                                            <label>选择模型</label>
                                            <select defaultValue="">
                                                <option value="">请选择预测模型</option>
                                                {models.filter(m => m.status === 'trained').map(model => (
                                                    <option key={model.id} value={model.id}>
                                                        {model.name} (IC: {model.ic})
                                                    </option>
                                                ))}
                                            </select>
                                        </div>
                                        <div className="form-group">
                                            <label>选择数据集</label>
                                            <select defaultValue="">
                                                <option value="">请选择数据集</option>
                                                {datasets.filter(d => d.status === 'ready').map(dataset => (
                                                    <option key={dataset.id} value={dataset.id}>
                                                        {dataset.name}
                                                    </option>
                                                ))}
                                            </select>
                                        </div>
                                        <div className="form-row">
                                            <div className="form-group">
                                                <label>开始日期</label>
                                                <input type="date" defaultValue="2022-01-01" />
                                            </div>
                                            <div className="form-group">
                                                <label>结束日期</label>
                                                <input type="date" defaultValue="2023-12-31" />
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div className="config-card">
                                    <div className="card-header">
                                        <h3>📈 策略配置</h3>
                                        <p>选择策略类型和关键参数</p>
                                    </div>
                                    <div className="card-content">
                                        <div className="strategy-types">
                                            {strategyTypes.map(strategy => (
                                                <div key={strategy.value} className="strategy-option">
                                                    <input 
                                                        type="radio" 
                                                        name="strategy" 
                                                        value={strategy.value}
                                                        defaultChecked={strategy.value === 'topk'}
                                                    />
                                                    <div className="strategy-info">
                                                        <div className="strategy-name">{strategy.label}</div>
                                                        <div className="strategy-desc">{strategy.desc}</div>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                        <div className="form-row">
                                            <div className="form-group">
                                                <label>持仓数量</label>
                                                <input type="number" defaultValue="30" min="1" max="100" />
                                            </div>
                                            <div className="form-group">
                                                <label>调仓频率</label>
                                                <select defaultValue="daily">
                                                    <option value="daily">每日</option>
                                                    <option value="weekly">每周</option>
                                                    <option value="monthly">每月</option>
                                                </select>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <div className="config-card">
                                    <div className="card-header">
                                        <h3>💰 资金设置</h3>
                                        <p>配置初始资金和交易成本</p>
                                    </div>
                                    <div className="card-content">
                                        <div className="form-group">
                                            <label>初始资金</label>
                                            <input type="number" defaultValue="1000000" step="10000" />
                                            <span className="input-suffix">元</span>
                                        </div>
                                        <div className="form-row">
                                            <div className="form-group">
                                                <label>手续费率</label>
                                                <input type="number" defaultValue="0.0003" step="0.0001" />
                                                <span className="input-suffix">%</span>
                                            </div>
                                            <div className="form-group">
                                                <label>冲击成本</label>
                                                <input type="number" defaultValue="0.0005" step="0.0001" />
                                                <span className="input-suffix">%</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'results' && (
                    <div className="results-content">
                        {backtestResults.length === 0 ? (
                            <div className="empty-state">
                                <div className="empty-icon">📈</div>
                                <div className="empty-text">暂无回测结果</div>
                                <div className="empty-sub">创建新的回测任务以查看结果</div>
                            </div>
                        ) : (
                            backtestResults.map(backtest => (
                                <div key={backtest.id} className="backtest-result">
                                    <div className="result-header">
                                        <div className="result-info">
                                            <h3>{backtest.name}</h3>
                                            <div className="result-meta">
                                                {backtest.strategy} • {backtest.startTime}
                                            </div>
                                        </div>
                                        <div className="result-status">
                                            <span className={`status-badge ${backtest.status}`}>
                                                {backtest.status === 'running' ? '🔄 运行中' : '✅ 已完成'}
                                            </span>
                                        </div>
                                    </div>

                                    {backtest.status === 'running' && (
                                        <div className="progress-section">
                                            <div className="progress-bar-container">
                                                <div className="progress-bar" style={{width: `${backtest.progress}%`}}></div>
                                            </div>
                                            <div className="progress-text">{Math.floor(backtest.progress)}%</div>
                                        </div>
                                    )}

                                    {backtest.status === 'completed' && backtest.results && (
                                        <div className="result-metrics">
                                            <div className="metrics-grid">
                                                <div className="metric-item">
                                                    <div className="metric-label">总收益率</div>
                                                    <div className="metric-value positive">{backtest.results.totalReturn}</div>
                                                </div>
                                                <div className="metric-item">
                                                    <div className="metric-label">年化收益</div>
                                                    <div className="metric-value">{backtest.results.annualReturn}</div>
                                                </div>
                                                <div className="metric-item">
                                                    <div className="metric-label">夏普比率</div>
                                                    <div className="metric-value">{backtest.results.sharpeRatio}</div>
                                                </div>
                                                <div className="metric-item">
                                                    <div className="metric-label">最大回撤</div>
                                                    <div className="metric-value negative">{backtest.results.maxDrawdown}</div>
                                                </div>
                                                <div className="metric-item">
                                                    <div className="metric-label">胜率</div>
                                                    <div className="metric-value">{backtest.results.winRate}</div>
                                                </div>
                                                <div className="metric-item">
                                                    <div className="metric-label">交易次数</div>
                                                    <div className="metric-value">{backtest.results.trades}</div>
                                                </div>
                                            </div>
                                            <div className="result-actions">
                                                <button className="btn-text">📊 详细报告</button>
                                                <button className="btn-text">📈 净值图表</button>
                                                <button className="btn-text">📋 交易记录</button>
                                                <button className="btn-text">📤 导出结果</button>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))
                        )}
                    </div>
                )}

                {activeTab === 'analysis' && (
                    <div className="analysis-content">
                        <div className="analysis-placeholder">
                            <div className="placeholder-icon">📊</div>
                            <h3>分析报告</h3>
                            <p>深入分析策略表现和风险特征</p>
                            <div className="analysis-features">
                                <div className="feature-item">📈 收益归因分析</div>
                                <div className="feature-item">📉 风险指标详解</div>
                                <div className="feature-item">📋 交易行为分析</div>
                                <div className="feature-item">🎯 策略优化建议</div>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* 回测配置表单 */}
            {showForm && (
                <div className="modal-overlay" onClick={() => setShowForm(false)}>
                    <div className="modal large" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>🚀 创建回测任务</h2>
                            <button className="close-btn" onClick={() => setShowForm(false)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="form-group">
                                <label>回测名称</label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={e => setFormData({...formData, name: e.target.value})}
                                    placeholder="输入回测任务名称"
                                />
                            </div>

                            <div className="form-row">
                                <div className="form-group">
                                    <label>选择模型</label>
                                    <select
                                        value={formData.model}
                                        onChange={e => setFormData({...formData, model: e.target.value})}
                                    >
                                        <option value="">请选择模型</option>
                                        {models.filter(m => m.status === 'trained').map(model => (
                                            <option key={model.id} value={model.id}>
                                                {model.name} (IC: {model.ic})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>选择数据集</label>
                                    <select
                                        value={formData.dataset}
                                        onChange={e => setFormData({...formData, dataset: e.target.value})}
                                    >
                                        <option value="">请选择数据集</option>
                                        {datasets.filter(d => d.status === 'ready').map(dataset => (
                                            <option key={dataset.id} value={dataset.id}>
                                                {dataset.name}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            </div>

                            <div className="form-group">
                                <label>策略类型</label>
                                <select
                                    value={formData.strategy}
                                    onChange={e => setFormData({...formData, strategy: e.target.value})}
                                >
                                    {strategyTypes.map(strategy => (
                                        <option key={strategy.value} value={strategy.value}>
                                            {strategy.label}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div className="form-row">
                                <div className="form-group">
                                    <label>开始日期</label>
                                    <input
                                        type="date"
                                        value={formData.startDate}
                                        onChange={e => setFormData({...formData, startDate: e.target.value})}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>结束日期</label>
                                    <input
                                        type="date"
                                        value={formData.endDate}
                                        onChange={e => setFormData({...formData, endDate: e.target.value})}
                                    />
                                </div>
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setShowForm(false)}>
                                取消
                            </button>
                            <button className="btn-primary" onClick={handleBacktest}>
                                开始回测
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};