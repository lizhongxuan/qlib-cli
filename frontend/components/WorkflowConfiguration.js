// 工作流配置组件
const { useState, useEffect } = React;

const WorkflowConfiguration = ({ 
    onNavigate = () => {},
    onAddTask = () => {}
}) => {
    const [activeStep, setActiveStep] = useState(0);
    const [config, setConfig] = useState({
        // 基础设置
        basic: {
            provider_uri: '~/.qlib/qlib_data/cn_data',
            region: 'cn',
            market: 'csi300',
            benchmark: 'SH000300'
        },
        // 数据处理
        data: {
            start_time: '2010-01-01',
            end_time: '2020-12-31',
            fit_start_time: '2010-01-01',
            fit_end_time: '2017-12-31',
            label: 'Ref($close, -1) / $close - 1'
        },
        // 模型配置
        model: {
            class: 'LightGBM',
            loss: 'mse',
            n_estimators: 200,
            seed: 2024,
            handler: 'Alpha158',
            train_start: '2010-01-01',
            train_end: '2017-12-31',
            valid_start: '2018-01-01',
            valid_end: '2018-12-31',
            test_start: '2019-01-01',
            test_end: '2020-12-31'
        },
        // 策略回测
        strategy: {
            class: 'TopkDropoutStrategy',
            topk: 50,
            n_drop: 5,
            backtest_start: '2019-01-01',
            backtest_end: '2020-12-31',
            account: 100000000,
            limit_threshold: 0.095,
            deal_price: 'close',
            open_cost: 0.0005,
            close_cost: 0.0015,
            min_cost: 5
        }
    });

    const [generatedYaml, setGeneratedYaml] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [runProgress, setRunProgress] = useState(0);
    const [runResults, setRunResults] = useState(null);

    // 配置步骤
    const steps = [
        {
            key: 'basic',
            title: '基础设置',
            icon: '⚙️',
            desc: '配置数据源、股票池和基准'
        },
        {
            key: 'data',
            title: '数据处理',
            icon: '📊',
            desc: '设置时间范围和标签定义'
        },
        {
            key: 'model',
            title: '模型训练',
            icon: '🤖',
            desc: '配置模型参数和数据集分割'
        },
        {
            key: 'strategy',
            title: '策略回测',
            icon: '📈',
            desc: '设置策略参数和回测配置'
        },
        {
            key: 'review',
            title: '配置预览',
            icon: '👀',
            desc: '检查配置并生成YAML文件'
        }
    ];

    // 股票池选项
    const marketOptions = [
        { value: 'csi300', label: 'CSI300 - 沪深300' },
        { value: 'csi500', label: 'CSI500 - 中证500' },
        { value: 'hs300', label: 'HS300 - 沪深300' },
        { value: 'all', label: 'ALL - 全市场' }
    ];

    // 模型选项
    const modelOptions = [
        { value: 'LightGBM', label: 'LightGBM - 轻量梯度提升' },
        { value: 'XGBoost', label: 'XGBoost - 极端梯度提升' },
        { value: 'CatBoost', label: 'CatBoost - 类别提升' },
        { value: 'Linear', label: 'Linear - 线性回归' },
        { value: 'MLP', label: 'MLP - 多层感知机' }
    ];

    // 因子库选项
    const handlerOptions = [
        { value: 'Alpha158', label: 'Alpha158 - 158个技术因子' },
        { value: 'Alpha360', label: 'Alpha360 - 360个综合因子' },
        { value: 'Alpha101', label: 'Alpha101 - 101个经典因子' }
    ];

    // 策略选项
    const strategyOptions = [
        { value: 'TopkDropoutStrategy', label: 'TopK选股策略' },
        { value: 'WeightStrategyBase', label: '权重策略基础版' },
        { value: 'EnhancedIndexingStrategy', label: '增强指数策略' }
    ];

    // 更新配置
    const updateConfig = (section, field, value) => {
        setConfig(prev => ({
            ...prev,
            [section]: {
                ...prev[section],
                [field]: value
            }
        }));
    };

    // 生成YAML配置
    const generateYaml = () => {
        const yamlContent = `# ===============================================
# Qlib 完整工作流配置文件
# ===============================================
# 该文件可直接用于 \`qrun\` 命令执行一个完整的量化研究流程

# -------------------------------------------
# 步骤 1: 基础设置
# -------------------------------------------
qlib_init:
    provider_uri: ${config.basic.provider_uri}
    region: ${config.basic.region}

market: ${config.basic.market}
benchmark: ${config.basic.benchmark}

# -------------------------------------------
# 步骤 2: 数据处理与因子工程
# -------------------------------------------
data_handler_config:
    start_time: ${config.data.start_time}
    end_time: ${config.data.end_time}
    fit_start_time: ${config.data.fit_start_time}
    fit_end_time: ${config.data.fit_end_time}
    instruments: market
    label: ["${config.data.label}"]

# -------------------------------------------
# 步骤 3: 模型训练任务
# -------------------------------------------
task:
    model:
        class: ${config.model.class}
        module_path: qlib.contrib.model.gbdt
        kwargs:
            loss: ${config.model.loss}
            n_estimators: ${config.model.n_estimators}
            seed: ${config.model.seed}

    dataset:
        class: DatasetH
        module_path: qlib.data.dataset
        kwargs:
            handler:
                class: ${config.model.handler}
                module_path: qlib.contrib.data.handler
                kwargs: {}
            
            segments:
                train: [${config.model.train_start}, ${config.model.train_end}]
                valid: [${config.model.valid_start}, ${config.model.valid_end}]
                test: [${config.model.test_start}, ${config.model.test_end}]

# -------------------------------------------
# 步骤 4: 策略回测
# -------------------------------------------
port_analysis_config:
    strategy:
        class: ${config.strategy.class}
        module_path: qlib.contrib.strategy
        kwargs:
            topk: ${config.strategy.topk}
            n_drop: ${config.strategy.n_drop}

    backtest:
        start_time: ${config.strategy.backtest_start}
        end_time: ${config.strategy.backtest_end}
        account: ${config.strategy.account}
        benchmark: ${config.basic.benchmark}
        exchange_kwargs:
            limit_threshold: ${config.strategy.limit_threshold}
            deal_price: ${config.strategy.deal_price}
            open_cost: ${config.strategy.open_cost}
            close_cost: ${config.strategy.close_cost}
            min_cost: ${config.strategy.min_cost}`;

        setGeneratedYaml(yamlContent);
        return yamlContent;
    };

    // 运行工作流
    const runWorkflow = async () => {
        setIsRunning(true);
        setRunProgress(0);
        
        // 模拟运行进度
        const progressSteps = [
            { progress: 10, message: '初始化Qlib环境...' },
            { progress: 25, message: '加载数据和因子...' },
            { progress: 45, message: '训练模型...' },
            { progress: 70, message: '模型验证...' },
            { progress: 85, message: '策略回测...' },
            { progress: 100, message: '生成分析报告...' }
        ];

        for (const step of progressSteps) {
            await new Promise(resolve => setTimeout(resolve, 2000));
            setRunProgress(step.progress);
        }

        // 模拟生成结果
        const mockResults = {
            model_performance: {
                train_ic: 0.0423,
                valid_ic: 0.0387,
                test_ic: 0.0356,
                train_rank_ic: 0.0521,
                valid_rank_ic: 0.0478,
                test_rank_ic: 0.0445
            },
            backtest_results: {
                total_return: 0.2847,
                benchmark_return: 0.1956,
                excess_return: 0.0891,
                sharpe_ratio: 1.654,
                max_drawdown: -0.0892,
                win_rate: 0.586,
                information_ratio: 1.423,
                volatility: 0.172
            },
            strategy_stats: {
                total_trades: 1247,
                avg_holding_period: 12.3,
                turnover_rate: 0.234,
                avg_position_size: 2.1
            }
        };

        setRunResults(mockResults);
        setIsRunning(false);

        // 添加到任务队列
        onAddTask({
            name: `工作流运行 - ${config.basic.market.toUpperCase()}`,
            type: 'workflow',
            config: config
        });
    };

    // 渲染基础设置步骤
    const renderBasicStep = () => (
        <div className="config-section">
            <h3>📊 基础设置</h3>
            
            <div className="form-grid">
                <div className="form-group">
                    <label>数据源路径</label>
                    <input
                        type="text"
                        value={config.basic.provider_uri}
                        onChange={(e) => updateConfig('basic', 'provider_uri', e.target.value)}
                        className="form-input"
                        placeholder="数据存放路径"
                    />
                    <div className="form-hint">请确保Qlib数据已正确初始化</div>
                </div>

                <div className="form-group">
                    <label>区域设置</label>
                    <select
                        value={config.basic.region}
                        onChange={(e) => updateConfig('basic', 'region', e.target.value)}
                        className="form-select"
                    >
                        <option value="cn">cn - 中国市场</option>
                        <option value="us">us - 美国市场</option>
                    </select>
                </div>

                <div className="form-group">
                    <label>股票池</label>
                    <select
                        value={config.basic.market}
                        onChange={(e) => updateConfig('basic', 'market', e.target.value)}
                        className="form-select"
                    >
                        {marketOptions.map(option => (
                            <option key={option.value} value={option.value}>
                                {option.label}
                            </option>
                        ))}
                    </select>
                </div>

                <div className="form-group">
                    <label>业绩基准</label>
                    <input
                        type="text"
                        value={config.basic.benchmark}
                        onChange={(e) => updateConfig('basic', 'benchmark', e.target.value)}
                        className="form-input"
                        placeholder="如: SH000300"
                    />
                    <div className="form-hint">基准指数代码</div>
                </div>
            </div>
        </div>
    );

    // 渲染数据处理步骤
    const renderDataStep = () => (
        <div className="config-section">
            <h3>📊 数据处理配置</h3>
            
            <div className="form-grid">
                <div className="form-group">
                    <label>数据开始时间</label>
                    <input
                        type="date"
                        value={config.data.start_time}
                        onChange={(e) => updateConfig('data', 'start_time', e.target.value)}
                        className="form-input"
                    />
                </div>

                <div className="form-group">
                    <label>数据结束时间</label>
                    <input
                        type="date"
                        value={config.data.end_time}
                        onChange={(e) => updateConfig('data', 'end_time', e.target.value)}
                        className="form-input"
                    />
                </div>

                <div className="form-group">
                    <label>预处理拟合开始时间</label>
                    <input
                        type="date"
                        value={config.data.fit_start_time}
                        onChange={(e) => updateConfig('data', 'fit_start_time', e.target.value)}
                        className="form-input"
                    />
                    <div className="form-hint">用于因子标准化等预处理的时间范围</div>
                </div>

                <div className="form-group">
                    <label>预处理拟合结束时间</label>
                    <input
                        type="date"
                        value={config.data.fit_end_time}
                        onChange={(e) => updateConfig('data', 'fit_end_time', e.target.value)}
                        className="form-input"
                    />
                </div>

                <div className="form-group full-width">
                    <label>标签定义（预测目标）</label>
                    <input
                        type="text"
                        value={config.data.label}
                        onChange={(e) => updateConfig('data', 'label', e.target.value)}
                        className="form-input"
                        placeholder="如: Ref($close, -1) / $close - 1"
                    />
                    <div className="form-hint">定义机器学习的预测目标，默认为次日收益率</div>
                </div>
            </div>
        </div>
    );

    // 渲染模型配置步骤
    const renderModelStep = () => (
        <div className="config-section">
            <h3>🤖 模型训练配置</h3>
            
            <div className="config-subsection">
                <h4>模型参数</h4>
                <div className="form-grid">
                    <div className="form-group">
                        <label>模型类型</label>
                        <select
                            value={config.model.class}
                            onChange={(e) => updateConfig('model', 'class', e.target.value)}
                            className="form-select"
                        >
                            {modelOptions.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="form-group">
                        <label>损失函数</label>
                        <select
                            value={config.model.loss}
                            onChange={(e) => updateConfig('model', 'loss', e.target.value)}
                            className="form-select"
                        >
                            <option value="mse">MSE - 均方误差</option>
                            <option value="mae">MAE - 平均绝对误差</option>
                            <option value="huber">Huber - 混合损失</option>
                        </select>
                    </div>

                    <div className="form-group">
                        <label>树的数量</label>
                        <input
                            type="number"
                            value={config.model.n_estimators}
                            onChange={(e) => updateConfig('model', 'n_estimators', parseInt(e.target.value))}
                            className="form-input"
                            min="50"
                            max="1000"
                        />
                    </div>

                    <div className="form-group">
                        <label>随机种子</label>
                        <input
                            type="number"
                            value={config.model.seed}
                            onChange={(e) => updateConfig('model', 'seed', parseInt(e.target.value))}
                            className="form-input"
                        />
                        <div className="form-hint">确保结果可复现</div>
                    </div>

                    <div className="form-group">
                        <label>因子库</label>
                        <select
                            value={config.model.handler}
                            onChange={(e) => updateConfig('model', 'handler', e.target.value)}
                            className="form-select"
                        >
                            {handlerOptions.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>
                </div>
            </div>

            <div className="config-subsection">
                <h4>数据集分割</h4>
                <div className="form-grid">
                    <div className="form-group">
                        <label>训练集开始</label>
                        <input
                            type="date"
                            value={config.model.train_start}
                            onChange={(e) => updateConfig('model', 'train_start', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>训练集结束</label>
                        <input
                            type="date"
                            value={config.model.train_end}
                            onChange={(e) => updateConfig('model', 'train_end', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>验证集开始</label>
                        <input
                            type="date"
                            value={config.model.valid_start}
                            onChange={(e) => updateConfig('model', 'valid_start', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>验证集结束</label>
                        <input
                            type="date"
                            value={config.model.valid_end}
                            onChange={(e) => updateConfig('model', 'valid_end', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>测试集开始</label>
                        <input
                            type="date"
                            value={config.model.test_start}
                            onChange={(e) => updateConfig('model', 'test_start', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>测试集结束</label>
                        <input
                            type="date"
                            value={config.model.test_end}
                            onChange={(e) => updateConfig('model', 'test_end', e.target.value)}
                            className="form-input"
                        />
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染策略配置步骤
    const renderStrategyStep = () => (
        <div className="config-section">
            <h3>📈 策略回测配置</h3>
            
            <div className="config-subsection">
                <h4>策略参数</h4>
                <div className="form-grid">
                    <div className="form-group">
                        <label>策略类型</label>
                        <select
                            value={config.strategy.class}
                            onChange={(e) => updateConfig('strategy', 'class', e.target.value)}
                            className="form-select"
                        >
                            {strategyOptions.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="form-group">
                        <label>选股数量 (TopK)</label>
                        <input
                            type="number"
                            value={config.strategy.topk}
                            onChange={(e) => updateConfig('strategy', 'topk', parseInt(e.target.value))}
                            className="form-input"
                            min="10"
                            max="200"
                        />
                        <div className="form-hint">每期选择得分最高的股票数</div>
                    </div>

                    <div className="form-group">
                        <label>淘汰数量 (Drop)</label>
                        <input
                            type="number"
                            value={config.strategy.n_drop}
                            onChange={(e) => updateConfig('strategy', 'n_drop', parseInt(e.target.value))}
                            className="form-input"
                            min="0"
                            max="50"
                        />
                        <div className="form-hint">每期淘汰排名最低的股票数</div>
                    </div>
                </div>
            </div>

            <div className="config-subsection">
                <h4>回测设置</h4>
                <div className="form-grid">
                    <div className="form-group">
                        <label>回测开始时间</label>
                        <input
                            type="date"
                            value={config.strategy.backtest_start}
                            onChange={(e) => updateConfig('strategy', 'backtest_start', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>回测结束时间</label>
                        <input
                            type="date"
                            value={config.strategy.backtest_end}
                            onChange={(e) => updateConfig('strategy', 'backtest_end', e.target.value)}
                            className="form-input"
                        />
                    </div>

                    <div className="form-group">
                        <label>初始资金</label>
                        <input
                            type="number"
                            value={config.strategy.account}
                            onChange={(e) => updateConfig('strategy', 'account', parseInt(e.target.value))}
                            className="form-input"
                            step="1000000"
                        />
                        <div className="form-hint">单位：元</div>
                    </div>

                    <div className="form-group">
                        <label>涨跌停限制</label>
                        <input
                            type="number"
                            value={config.strategy.limit_threshold}
                            onChange={(e) => updateConfig('strategy', 'limit_threshold', parseFloat(e.target.value))}
                            className="form-input"
                            step="0.001"
                            min="0"
                            max="0.2"
                        />
                        <div className="form-hint">0.095 = 9.5%</div>
                    </div>

                    <div className="form-group">
                        <label>开仓手续费率</label>
                        <input
                            type="number"
                            value={config.strategy.open_cost}
                            onChange={(e) => updateConfig('strategy', 'open_cost', parseFloat(e.target.value))}
                            className="form-input"
                            step="0.0001"
                            min="0"
                            max="0.01"
                        />
                        <div className="form-hint">0.0005 = 0.05%</div>
                    </div>

                    <div className="form-group">
                        <label>平仓手续费率</label>
                        <input
                            type="number"
                            value={config.strategy.close_cost}
                            onChange={(e) => updateConfig('strategy', 'close_cost', parseFloat(e.target.value))}
                            className="form-input"
                            step="0.0001"
                            min="0"
                            max="0.01"
                        />
                        <div className="form-hint">0.0015 = 0.15%</div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染配置预览步骤
    const renderReviewStep = () => {
        useEffect(() => {
            generateYaml();
        }, []);

        return (
            <div className="config-section">
                <h3>👀 配置预览</h3>
                
                <div className="config-summary">
                    <div className="summary-grid">
                        <div className="summary-item">
                            <div className="summary-label">股票池</div>
                            <div className="summary-value">{config.basic.market.toUpperCase()}</div>
                        </div>
                        <div className="summary-item">
                            <div className="summary-label">模型</div>
                            <div className="summary-value">{config.model.class}</div>
                        </div>
                        <div className="summary-item">
                            <div className="summary-label">因子库</div>
                            <div className="summary-value">{config.model.handler}</div>
                        </div>
                        <div className="summary-item">
                            <div className="summary-label">策略</div>
                            <div className="summary-value">Top{config.strategy.topk}</div>
                        </div>
                        <div className="summary-item">
                            <div className="summary-label">回测期间</div>
                            <div className="summary-value">
                                {config.strategy.backtest_start} 至 {config.strategy.backtest_end}
                            </div>
                        </div>
                        <div className="summary-item">
                            <div className="summary-label">初始资金</div>
                            <div className="summary-value">
                                {(config.strategy.account / 100000000).toFixed(1)}亿元
                            </div>
                        </div>
                    </div>
                </div>

                <div className="yaml-preview">
                    <div className="yaml-header">
                        <h4>📄 生成的YAML配置文件</h4>
                        <div className="yaml-actions">
                            <button 
                                className="btn-secondary"
                                onClick={() => navigator.clipboard.writeText(generatedYaml)}
                            >
                                📋 复制配置
                            </button>
                            <button 
                                className="btn-secondary"
                                onClick={() => {
                                    const blob = new Blob([generatedYaml], { type: 'text/yaml' });
                                    const url = URL.createObjectURL(blob);
                                    const a = document.createElement('a');
                                    a.href = url;
                                    a.download = 'qlib_workflow_config.yaml';
                                    a.click();
                                }}
                            >
                                💾 下载文件
                            </button>
                        </div>
                    </div>
                    <pre className="yaml-content">{generatedYaml}</pre>
                </div>
            </div>
        );
    };

    // 渲染当前步骤内容
    const renderStepContent = () => {
        switch (activeStep) {
            case 0: return renderBasicStep();
            case 1: return renderDataStep();
            case 2: return renderModelStep();
            case 3: return renderStrategyStep();
            case 4: return renderReviewStep();
            default: return null;
        }
    };

    return (
        <div className="workflow-configuration">
            <div className="workflow-header">
                <h1>⚙️ 工作流配置向导</h1>
                <div className="header-subtitle">
                    通过可视化界面配置完整的Qlib量化研究工作流
                </div>
            </div>

            {/* 步骤导航 */}
            <div className="steps-navigation">
                {steps.map((step, index) => (
                    <div
                        key={step.key}
                        className={`step-item ${index === activeStep ? 'active' : ''} ${index < activeStep ? 'completed' : ''}`}
                        onClick={() => setActiveStep(index)}
                    >
                        <div className="step-icon">{step.icon}</div>
                        <div className="step-content">
                            <div className="step-title">{step.title}</div>
                            <div className="step-desc">{step.desc}</div>
                        </div>
                        <div className="step-number">{index + 1}</div>
                    </div>
                ))}
            </div>

            {/* 配置内容 */}
            <div className="configuration-content">
                {renderStepContent()}
            </div>

            {/* 操作按钮 */}
            <div className="workflow-actions">
                <div className="actions-left">
                    {activeStep > 0 && (
                        <button 
                            className="btn-secondary"
                            onClick={() => setActiveStep(activeStep - 1)}
                        >
                            ← 上一步
                        </button>
                    )}
                </div>

                <div className="actions-right">
                    {activeStep < steps.length - 1 ? (
                        <button 
                            className="btn-primary"
                            onClick={() => setActiveStep(activeStep + 1)}
                        >
                            下一步 →
                        </button>
                    ) : (
                        <button 
                            className="btn-success"
                            onClick={runWorkflow}
                            disabled={isRunning}
                        >
                            {isRunning ? '🔄 运行中...' : '🚀 运行工作流'}
                        </button>
                    )}
                </div>
            </div>

            {/* 运行进度 */}
            {isRunning && (
                <div className="run-progress">
                    <div className="progress-header">
                        <h3>🔄 工作流运行中</h3>
                        <div className="progress-percentage">{runProgress}%</div>
                    </div>
                    <div className="progress-bar">
                        <div 
                            className="progress-fill" 
                            style={{ width: `${runProgress}%` }}
                        ></div>
                    </div>
                    <div className="progress-status">
                        正在处理: {
                            runProgress <= 10 ? '初始化Qlib环境...' :
                            runProgress <= 25 ? '加载数据和因子...' :
                            runProgress <= 45 ? '训练模型...' :
                            runProgress <= 70 ? '模型验证...' :
                            runProgress <= 85 ? '策略回测...' :
                            '生成分析报告...'
                        }
                    </div>
                </div>
            )}

            {/* 运行结果 */}
            {runResults && (
                <div className="run-results">
                    <div className="results-header">
                        <h3>✅ 工作流运行完成</h3>
                        <div className="results-actions">
                            <button 
                                className="btn-primary"
                                onClick={() => onNavigate('results')}
                            >
                                📊 查看详细报告
                            </button>
                        </div>
                    </div>

                    <div className="results-summary">
                        <div className="result-section">
                            <h4>📈 模型表现</h4>
                            <div className="metrics-grid">
                                <div className="metric-item">
                                    <span className="metric-label">训练IC</span>
                                    <span className="metric-value">{runResults.model_performance.train_ic.toFixed(4)}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">验证IC</span>
                                    <span className="metric-value">{runResults.model_performance.valid_ic.toFixed(4)}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">测试IC</span>
                                    <span className="metric-value">{runResults.model_performance.test_ic.toFixed(4)}</span>
                                </div>
                            </div>
                        </div>

                        <div className="result-section">
                            <h4>💰 回测结果</h4>
                            <div className="metrics-grid">
                                <div className="metric-item">
                                    <span className="metric-label">总收益率</span>
                                    <span className="metric-value positive">
                                        {(runResults.backtest_results.total_return * 100).toFixed(2)}%
                                    </span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">超额收益</span>
                                    <span className="metric-value positive">
                                        {(runResults.backtest_results.excess_return * 100).toFixed(2)}%
                                    </span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">夏普比率</span>
                                    <span className="metric-value">{runResults.backtest_results.sharpe_ratio.toFixed(3)}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">最大回撤</span>
                                    <span className="metric-value negative">
                                        {(runResults.backtest_results.max_drawdown * 100).toFixed(2)}%
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};