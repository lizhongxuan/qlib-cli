// Qlib量化研究工作流 - 核心引擎
const { useState, useEffect } = React;

const QlibWorkflow = ({ 
    onNavigate = () => {},
    onAddTask = () => {},
    savedFactors = [],
    models = [],
    onAddModel = () => {},
    datasets = []
}) => {
    const [activeStep, setActiveStep] = useState(0);
    const [pipeline, setPipeline] = useState({
        data: {
            provider_uri: '~/.qlib/qlib_data/cn_data',
            region: 'cn',
            market: 'csi300',
            benchmark: 'SH000300',
            start_time: '2018-01-01',
            end_time: '2023-12-31'
        },
        features: {
            handler: 'Alpha158',
            factors: [],
            label: 'Ref($close, -1) / $close - 1'
        },
        model: {
            class: 'LightGBM',
            params: {
                n_estimators: 200,
                learning_rate: 0.1,
                max_depth: 6,
                seed: 2024
            },
            segments: {
                train: ['2018-01-01', '2021-12-31'],
                valid: ['2022-01-01', '2022-12-31'], 
                test: ['2023-01-01', '2023-12-31']
            }
        },
        strategy: {
            class: 'TopkDropoutStrategy',
            params: {
                topk: 50,
                n_drop: 5
            }
        },
        backtest: {
            start_time: '2023-01-01',
            end_time: '2023-12-31',
            account: 100000000,
            benchmark: 'SH000300',
            exchange: {
                limit_threshold: 0.095,
                deal_price: 'close',
                open_cost: 0.0005,
                close_cost: 0.0015,
                min_cost: 5
            }
        }
    });

    const [isRunning, setIsRunning] = useState(false);
    const [runProgress, setRunProgress] = useState(0);
    const [runResults, setRunResults] = useState(null);
    const [generatedConfig, setGeneratedConfig] = useState('');

    // 工作流步骤定义
    const workflowSteps = [
        {
            key: 'data',
            title: 'qlib数据配置',
            icon: '💾',
            desc: '配置qlib数据源和股票池',
            status: 'completed'
        },
        {
            key: 'features',
            title: '特征工程',
            icon: '🧮', 
            desc: '选择因子和特征处理',
            status: 'pending'
        },
        {
            key: 'model',
            title: 'qlib模型训练',
            icon: '🤖',
            desc: '配置和训练预测模型',
            status: 'pending'
        },
        {
            key: 'strategy',
            title: '投资策略',
            icon: '📈',
            desc: '配置交易策略和参数',
            status: 'pending'
        },
        {
            key: 'backtest',
            title: '策略回测',
            icon: '🎯',
            desc: '执行回测和性能分析',
            status: 'pending'
        }
    ];

    // qlib预设配置
    const qlibPresets = {
        datasets: [
            { value: 'csi300', label: 'CSI300 - 沪深300成分股', desc: 'qlib内置中国A股主要指数' },
            { value: 'csi500', label: 'CSI500 - 中证500', desc: 'qlib内置中盘股指数' },
            { value: 'all', label: 'All - 全A股市场', desc: 'qlib全市场股票数据' }
        ],
        handlers: [
            { value: 'Alpha158', label: 'Alpha158 - qlib经典158因子', desc: '包含价格、成交量、技术指标等158个因子' },
            { value: 'Alpha360', label: 'Alpha360 - qlib增强360因子', desc: '扩展的360个综合因子库' },
            { value: 'Alpha101', label: 'Alpha101 - WorldQuant101因子', desc: 'WorldQuant开源的101个经典因子' }
        ],
        models: [
            { value: 'LightGBM', label: 'LightGBM - 梯度提升树', desc: 'qlib优化的LightGBM实现，适合表格数据' },
            { value: 'CatBoost', label: 'CatBoost - 类别提升', desc: 'Yandex开发的梯度提升算法' },
            { value: 'XGBoost', label: 'XGBoost - 极端梯度提升', desc: '经典的梯度提升框架' },
            { value: 'Linear', label: 'Linear - 线性回归', desc: '简单快速的线性模型' },
            { value: 'MLP', label: 'MLP - 多层感知机', desc: 'qlib内置的深度学习模型' },
            { value: 'LSTM', label: 'LSTM - 长短期记忆网络', desc: '时序数据的深度学习模型' }
        ],
        strategies: [
            { value: 'TopkDropoutStrategy', label: 'TopK选股策略', desc: 'qlib经典的TopK选股+Dropout策略' },
            { value: 'WeightStrategyBase', label: '权重分配策略', desc: '基于模型预测的权重分配' },
            { value: 'EnhancedIndexingStrategy', label: '增强指数策略', desc: '指数增强型投资策略' }
        ]
    };

    // 更新流水线配置
    const updatePipeline = (step, config) => {
        setPipeline(prev => ({
            ...prev,
            [step]: { ...prev[step], ...config }
        }));
    };

    // 生成qlib配置
    const generateQlibConfig = () => {
        const config = `# qlib量化研究工作流配置
# 此配置文件可直接用于qrun命令执行

# qlib初始化
qlib_init:
    provider_uri: ${pipeline.data.provider_uri}
    region: ${pipeline.data.region}

market: ${pipeline.data.market}
benchmark: ${pipeline.data.benchmark}

# 数据处理配置
data_handler_config:
    start_time: ${pipeline.data.start_time}
    end_time: ${pipeline.data.end_time}
    fit_start_time: ${pipeline.data.start_time}
    fit_end_time: ${pipeline.model.segments.train[1]}
    instruments: market
    label: ["${pipeline.features.label}"]

# 模型训练任务
task:
    model:
        class: ${pipeline.model.class}
        module_path: qlib.contrib.model.gbdt
        kwargs:
            n_estimators: ${pipeline.model.params.n_estimators}
            learning_rate: ${pipeline.model.params.learning_rate}
            max_depth: ${pipeline.model.params.max_depth}
            seed: ${pipeline.model.params.seed}

    dataset:
        class: DatasetH
        module_path: qlib.data.dataset
        kwargs:
            handler:
                class: ${pipeline.features.handler}
                module_path: qlib.contrib.data.handler
                kwargs: {}
            
            segments:
                train: [${pipeline.model.segments.train[0]}, ${pipeline.model.segments.train[1]}]
                valid: [${pipeline.model.segments.valid[0]}, ${pipeline.model.segments.valid[1]}]
                test: [${pipeline.model.segments.test[0]}, ${pipeline.model.segments.test[1]}]

# 策略回测配置
port_analysis_config:
    strategy:
        class: ${pipeline.strategy.class}
        module_path: qlib.contrib.strategy
        kwargs:
            topk: ${pipeline.strategy.params.topk}
            n_drop: ${pipeline.strategy.params.n_drop}

    backtest:
        start_time: ${pipeline.backtest.start_time}
        end_time: ${pipeline.backtest.end_time}
        account: ${pipeline.backtest.account}
        benchmark: ${pipeline.backtest.benchmark}
        exchange_kwargs:
            limit_threshold: ${pipeline.backtest.exchange.limit_threshold}
            deal_price: ${pipeline.backtest.exchange.deal_price}
            open_cost: ${pipeline.backtest.exchange.open_cost}
            close_cost: ${pipeline.backtest.exchange.close_cost}
            min_cost: ${pipeline.backtest.exchange.min_cost}`;

        setGeneratedConfig(config);
        return config;
    };

    // 运行qlib工作流
    const runQlibWorkflow = async () => {
        setIsRunning(true);
        setRunProgress(0);

        const progressSteps = [
            { progress: 15, message: '初始化qlib环境...', detail: '加载数据源配置' },
            { progress: 30, message: '加载股票数据...', detail: `处理${pipeline.data.market}数据集` },
            { progress: 45, message: '特征工程...', detail: `计算${pipeline.features.handler}因子` },
            { progress: 60, message: '训练模型...', detail: `使用${pipeline.model.class}算法` },
            { progress: 75, message: '策略配置...', detail: `配置${pipeline.strategy.class}策略` },
            { progress: 90, message: '执行回测...', detail: '计算策略收益和风险指标' },
            { progress: 100, message: '生成分析报告...', detail: '完成量化研究流程' }
        ];

        for (const step of progressSteps) {
            await new Promise(resolve => setTimeout(resolve, 2000));
            setRunProgress(step.progress);
        }

        // 模拟qlib运行结果
        const mockResults = {
            model_performance: {
                train_ic: 0.0456,
                valid_ic: 0.0398,
                test_ic: 0.0367,
                train_rank_ic: 0.0612,
                valid_rank_ic: 0.0534,
                test_rank_ic: 0.0489,
                model_path: `/qlib/models/${pipeline.model.class}_${Date.now()}.pkl`
            },
            strategy_performance: {
                annual_return: 0.1847,
                benchmark_return: 0.0956,
                excess_return: 0.0891,
                volatility: 0.1623,
                sharpe_ratio: 1.138,
                information_ratio: 0.549,
                max_drawdown: -0.0847,
                win_rate: 0.574,
                calmar_ratio: 2.18
            },
            backtest_details: {
                total_trades: 2341,
                avg_holding_days: 8.7,
                turnover_rate: 0.234,
                trading_cost: 0.0156,
                net_return: 0.1691
            },
            factor_analysis: {
                top_factors: [
                    { name: 'RESI5', ic: 0.0423, weight: 0.125 },
                    { name: 'WVMA5', ic: 0.0389, weight: 0.098 },
                    { name: 'RSQR10', ic: 0.0356, weight: 0.087 },
                    { name: 'CORR20', ic: 0.0341, weight: 0.076 },
                    { name: 'STD20', ic: 0.0298, weight: 0.065 }
                ]
            }
        };

        setRunResults(mockResults);
        setIsRunning(false);

        // 添加到任务队列
        onAddTask({
            name: `qlib量化研究 - ${pipeline.data.market.toUpperCase()}`,
            type: 'qlib_workflow',
            pipeline: pipeline,
            results: mockResults,
            config: generateQlibConfig(),
            status: 'completed'
        });
    };

    // 渲染数据配置步骤
    const renderDataStep = () => (
        <div className="step-content">
            <h3>💾 qlib数据配置</h3>
            
            <div className="config-sections">
                <div className="config-section">
                    <h4>数据源配置</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>qlib数据路径</label>
                            <input
                                type="text"
                                value={pipeline.data.provider_uri}
                                onChange={(e) => updatePipeline('data', { provider_uri: e.target.value })}
                                className="form-input"
                            />
                            <div className="form-hint">请确保qlib数据已正确初始化</div>
                        </div>
                        
                        <div className="form-group">
                            <label>市场区域</label>
                            <select
                                value={pipeline.data.region}
                                onChange={(e) => updatePipeline('data', { region: e.target.value })}
                                className="form-select"
                            >
                                <option value="cn">cn - 中国A股市场</option>
                                <option value="us">us - 美国股票市场</option>
                            </select>
                        </div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>股票池和基准</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>股票池</label>
                            <select
                                value={pipeline.data.market}
                                onChange={(e) => updatePipeline('data', { market: e.target.value })}
                                className="form-select"
                            >
                                {qlibPresets.datasets.map(dataset => (
                                    <option key={dataset.value} value={dataset.value}>
                                        {dataset.label}
                                    </option>
                                ))}
                            </select>
                            <div className="form-hint">选择qlib内置的股票数据集</div>
                        </div>

                        <div className="form-group">
                            <label>基准指数</label>
                            <input
                                type="text"
                                value={pipeline.data.benchmark}
                                onChange={(e) => updatePipeline('data', { benchmark: e.target.value })}
                                className="form-input"
                                placeholder="如: SH000300"
                            />
                        </div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>时间范围</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>开始时间</label>
                            <input
                                type="date"
                                value={pipeline.data.start_time}
                                onChange={(e) => updatePipeline('data', { start_time: e.target.value })}
                                className="form-input"
                            />
                        </div>

                        <div className="form-group">
                            <label>结束时间</label>
                            <input
                                type="date"
                                value={pipeline.data.end_time}
                                onChange={(e) => updatePipeline('data', { end_time: e.target.value })}
                                className="form-input"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染特征工程步骤
    const renderFeaturesStep = () => (
        <div className="step-content">
            <h3>🧮 特征工程</h3>
            
            <div className="config-sections">
                <div className="config-section">
                    <h4>qlib因子库</h4>
                    <div className="form-group">
                        <label>选择因子处理器</label>
                        <select
                            value={pipeline.features.handler}
                            onChange={(e) => updatePipeline('features', { handler: e.target.value })}
                            className="form-select"
                        >
                            {qlibPresets.handlers.map(handler => (
                                <option key={handler.value} value={handler.value}>
                                    {handler.label}
                                </option>
                            ))}
                        </select>
                        <div className="form-hint">qlib内置的因子处理器，包含预定义的技术指标</div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>自定义因子</h4>
                    <div className="custom-factors">
                        {savedFactors.length > 0 ? (
                            <div className="factors-list">
                                <label>选择已保存的因子:</label>
                                {savedFactors.map(factor => (
                                    <div key={factor.id} className="factor-item">
                                        <input
                                            type="checkbox"
                                            id={factor.id}
                                            checked={pipeline.features.factors?.some(f => f.id === factor.id) || false}
                                            onChange={(e) => {
                                                const factors = pipeline.features.factors || [];
                                                if (e.target.checked) {
                                                    updatePipeline('features', {
                                                        factors: [...factors, factor]
                                                    });
                                                } else {
                                                    updatePipeline('features', {
                                                        factors: factors.filter(f => f.id !== factor.id)
                                                    });
                                                }
                                            }}
                                        />
                                        <label htmlFor={factor.id}>
                                            <strong>{factor.name}</strong>
                                            <div className="factor-expression">{factor.expression}</div>
                                        </label>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="no-factors">
                                <p>暂无自定义因子</p>
                                <button 
                                    className="btn-secondary btn-sm"
                                    onClick={() => onNavigate('factor')}
                                >
                                    前往因子研究
                                </button>
                            </div>
                        )}
                    </div>
                </div>

                <div className="config-section">
                    <h4>标签定义</h4>
                    <div className="form-group">
                        <label>预测目标 (Y)</label>
                        <input
                            type="text"
                            value={pipeline.features.label}
                            onChange={(e) => updatePipeline('features', { label: e.target.value })}
                            className="form-input"
                            placeholder="如: Ref($close, -1) / $close - 1"
                        />
                        <div className="form-hint">定义机器学习的预测目标，通常为未来收益率</div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染模型训练步骤
    const renderModelStep = () => (
        <div className="step-content">
            <h3>🤖 qlib模型训练</h3>
            
            <div className="config-sections">
                <div className="config-section">
                    <h4>模型选择</h4>
                    <div className="model-options">
                        {qlibPresets.models.map(model => (
                            <div 
                                key={model.value}
                                className={`model-card ${pipeline.model.class === model.value ? 'selected' : ''}`}
                                onClick={() => updatePipeline('model', { class: model.value })}
                            >
                                <div className="model-name">{model.label}</div>
                                <div className="model-desc">{model.desc}</div>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="config-section">
                    <h4>模型参数</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>树的数量/迭代次数</label>
                            <input
                                type="number"
                                value={pipeline.model.params.n_estimators}
                                onChange={(e) => updatePipeline('model', {
                                    params: { ...pipeline.model.params, n_estimators: parseInt(e.target.value) }
                                })}
                                className="form-input"
                                min="10" max="1000"
                            />
                        </div>

                        <div className="form-group">
                            <label>学习率</label>
                            <input
                                type="number"
                                value={pipeline.model.params.learning_rate}
                                onChange={(e) => updatePipeline('model', {
                                    params: { ...pipeline.model.params, learning_rate: parseFloat(e.target.value) }
                                })}
                                className="form-input"
                                min="0.001" max="1" step="0.001"
                            />
                        </div>

                        <div className="form-group">
                            <label>最大深度</label>
                            <input
                                type="number"
                                value={pipeline.model.params.max_depth}
                                onChange={(e) => updatePipeline('model', {
                                    params: { ...pipeline.model.params, max_depth: parseInt(e.target.value) }
                                })}
                                className="form-input"
                                min="1" max="20"
                            />
                        </div>

                        <div className="form-group">
                            <label>随机种子</label>
                            <input
                                type="number"
                                value={pipeline.model.params.seed}
                                onChange={(e) => updatePipeline('model', {
                                    params: { ...pipeline.model.params, seed: parseInt(e.target.value) }
                                })}
                                className="form-input"
                            />
                        </div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>数据分割</h4>
                    <div className="segments-config">
                        <div className="segment-group">
                            <label>训练集</label>
                            <div className="date-range">
                                <input
                                    type="date"
                                    value={pipeline.model.segments.train[0]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            train: [e.target.value, pipeline.model.segments.train[1]]
                                        }
                                    })}
                                    className="form-input"
                                />
                                <span>至</span>
                                <input
                                    type="date"
                                    value={pipeline.model.segments.train[1]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            train: [pipeline.model.segments.train[0], e.target.value]
                                        }
                                    })}
                                    className="form-input"
                                />
                            </div>
                        </div>

                        <div className="segment-group">
                            <label>验证集</label>
                            <div className="date-range">
                                <input
                                    type="date"
                                    value={pipeline.model.segments.valid[0]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            valid: [e.target.value, pipeline.model.segments.valid[1]]
                                        }
                                    })}
                                    className="form-input"
                                />
                                <span>至</span>
                                <input
                                    type="date"
                                    value={pipeline.model.segments.valid[1]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            valid: [pipeline.model.segments.valid[0], e.target.value]
                                        }
                                    })}
                                    className="form-input"
                                />
                            </div>
                        </div>

                        <div className="segment-group">
                            <label>测试集</label>
                            <div className="date-range">
                                <input
                                    type="date"
                                    value={pipeline.model.segments.test[0]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            test: [e.target.value, pipeline.model.segments.test[1]]
                                        }
                                    })}
                                    className="form-input"
                                />
                                <span>至</span>
                                <input
                                    type="date"
                                    value={pipeline.model.segments.test[1]}
                                    onChange={(e) => updatePipeline('model', {
                                        segments: {
                                            ...pipeline.model.segments,
                                            test: [pipeline.model.segments.test[0], e.target.value]
                                        }
                                    })}
                                    className="form-input"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染策略配置步骤
    const renderStrategyStep = () => (
        <div className="step-content">
            <h3>📈 投资策略配置</h3>
            
            <div className="config-sections">
                <div className="config-section">
                    <h4>策略类型</h4>
                    <div className="strategy-options">
                        {qlibPresets.strategies.map(strategy => (
                            <div
                                key={strategy.value}
                                className={`strategy-card ${pipeline.strategy.class === strategy.value ? 'selected' : ''}`}
                                onClick={() => updatePipeline('strategy', { class: strategy.value })}
                            >
                                <div className="strategy-name">{strategy.label}</div>
                                <div className="strategy-desc">{strategy.desc}</div>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="config-section">
                    <h4>策略参数</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>选股数量 (TopK)</label>
                            <input
                                type="number"
                                value={pipeline.strategy.params.topk}
                                onChange={(e) => updatePipeline('strategy', {
                                    params: { ...pipeline.strategy.params, topk: parseInt(e.target.value) }
                                })}
                                className="form-input"
                                min="10" max="200"
                            />
                            <div className="form-hint">每期选择模型预测得分最高的股票数量</div>
                        </div>

                        <div className="form-group">
                            <label>淘汰数量 (Dropout)</label>
                            <input
                                type="number"
                                value={pipeline.strategy.params.n_drop}
                                onChange={(e) => updatePipeline('strategy', {
                                    params: { ...pipeline.strategy.params, n_drop: parseInt(e.target.value) }
                                })}
                                className="form-input"
                                min="0" max="50"
                            />
                            <div className="form-hint">每期淘汰持仓中排名最低的股票数量</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染回测配置步骤
    const renderBacktestStep = () => (
        <div className="step-content">
            <h3>🎯 策略回测配置</h3>
            
            <div className="config-sections">
                <div className="config-section">
                    <h4>回测时间</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>回测开始时间</label>
                            <input
                                type="date"
                                value={pipeline.backtest.start_time}
                                onChange={(e) => updatePipeline('backtest', { start_time: e.target.value })}
                                className="form-input"
                            />
                        </div>

                        <div className="form-group">
                            <label>回测结束时间</label>
                            <input
                                type="date"
                                value={pipeline.backtest.end_time}
                                onChange={(e) => updatePipeline('backtest', { end_time: e.target.value })}
                                className="form-input"
                            />
                        </div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>资金和基准</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>初始资金 (元)</label>
                            <input
                                type="number"
                                value={pipeline.backtest.account}
                                onChange={(e) => updatePipeline('backtest', { account: parseInt(e.target.value) })}
                                className="form-input"
                                step="1000000"
                            />
                        </div>

                        <div className="form-group">
                            <label>业绩基准</label>
                            <input
                                type="text"
                                value={pipeline.backtest.benchmark}
                                onChange={(e) => updatePipeline('backtest', { benchmark: e.target.value })}
                                className="form-input"
                            />
                        </div>
                    </div>
                </div>

                <div className="config-section">
                    <h4>交易成本设置</h4>
                    <div className="form-grid">
                        <div className="form-group">
                            <label>涨跌停限制</label>
                            <input
                                type="number"
                                value={pipeline.backtest.exchange.limit_threshold}
                                onChange={(e) => updatePipeline('backtest', {
                                    exchange: { ...pipeline.backtest.exchange, limit_threshold: parseFloat(e.target.value) }
                                })}
                                className="form-input"
                                step="0.001" min="0" max="0.2"
                            />
                            <div className="form-hint">0.095 = 9.5%</div>
                        </div>

                        <div className="form-group">
                            <label>开仓手续费率</label>
                            <input
                                type="number"
                                value={pipeline.backtest.exchange.open_cost}
                                onChange={(e) => updatePipeline('backtest', {
                                    exchange: { ...pipeline.backtest.exchange, open_cost: parseFloat(e.target.value) }
                                })}
                                className="form-input"
                                step="0.0001" min="0" max="0.01"
                            />
                            <div className="form-hint">0.0005 = 0.05%</div>
                        </div>

                        <div className="form-group">
                            <label>平仓手续费率</label>
                            <input
                                type="number"
                                value={pipeline.backtest.exchange.close_cost}
                                onChange={(e) => updatePipeline('backtest', {
                                    exchange: { ...pipeline.backtest.exchange, close_cost: parseFloat(e.target.value) }
                                })}
                                className="form-input"
                                step="0.0001" min="0" max="0.01"
                            />
                            <div className="form-hint">0.0015 = 0.15%</div>
                        </div>

                        <div className="form-group">
                            <label>最低手续费 (元)</label>
                            <input
                                type="number"
                                value={pipeline.backtest.exchange.min_cost}
                                onChange={(e) => updatePipeline('backtest', {
                                    exchange: { ...pipeline.backtest.exchange, min_cost: parseInt(e.target.value) }
                                })}
                                className="form-input"
                                min="0"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染当前步骤内容
    const renderStepContent = () => {
        switch(activeStep) {
            case 0: return renderDataStep();
            case 1: return renderFeaturesStep();
            case 2: return renderModelStep();
            case 3: return renderStrategyStep();
            case 4: return renderBacktestStep();
            default: return renderDataStep();
        }
    };

    useEffect(() => {
        generateQlibConfig();
    }, [pipeline]);

    return (
        <div className="qlib-workflow">
            <div className="workflow-header">
                <h1>⚙️ qlib量化研究工作流</h1>
                <div className="header-subtitle">
                    基于qlib框架的端到端量化投资研究平台
                </div>
            </div>

            {/* 工作流步骤导航 */}
            <div className="workflow-steps">
                {workflowSteps.map((step, index) => (
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
                        {index < workflowSteps.length - 1 && <div className="step-connector"></div>}
                    </div>
                ))}
            </div>

            {/* 主配置区域 */}
            <div className="workflow-main">
                <div className="configuration-panel">
                    {renderStepContent()}
                </div>

                {/* 配置预览 */}
                <div className="preview-panel">
                    <div className="preview-header">
                        <h4>📄 qlib配置预览</h4>
                        <div className="preview-actions">
                            <button 
                                className="btn-secondary btn-sm"
                                onClick={() => navigator.clipboard.writeText(generatedConfig)}
                            >
                                📋 复制
                            </button>
                            <button 
                                className="btn-secondary btn-sm"
                                onClick={() => {
                                    const blob = new Blob([generatedConfig], { type: 'text/yaml' });
                                    const url = URL.createObjectURL(blob);
                                    const a = document.createElement('a');
                                    a.href = url;
                                    a.download = `qlib_config_${Date.now()}.yaml`;
                                    a.click();
                                }}
                            >
                                💾 下载
                            </button>
                        </div>
                    </div>
                    <pre className="config-preview">{generatedConfig}</pre>
                </div>
            </div>

            {/* 操作控制 */}
            <div className="workflow-controls">
                <div className="controls-left">
                    {activeStep > 0 && (
                        <button 
                            className="btn-secondary"
                            onClick={() => setActiveStep(activeStep - 1)}
                        >
                            ← 上一步
                        </button>
                    )}
                </div>

                <div className="controls-right">
                    <button 
                        className="btn-success btn-large"
                        onClick={runQlibWorkflow}
                        disabled={isRunning}
                    >
                        {isRunning ? '🔄 运行中...' : '🚀 运行qlib工作流'}
                    </button>
                </div>
            </div>

            {/* 运行进度 */}
            {isRunning && (
                <div className="run-progress">
                    <div className="progress-header">
                        <h3>🔄 qlib工作流运行中</h3>
                        <div className="progress-percentage">{runProgress}%</div>
                    </div>
                    <div className="progress-bar">
                        <div 
                            className="progress-fill" 
                            style={{ width: `${runProgress}%` }}
                        ></div>
                    </div>
                    <div className="progress-details">
                        {runProgress <= 15 && '初始化qlib环境...'}
                        {runProgress > 15 && runProgress <= 30 && `加载股票数据... (${pipeline.data.market})`}
                        {runProgress > 30 && runProgress <= 45 && `特征工程... (${pipeline.features.handler})`}
                        {runProgress > 45 && runProgress <= 60 && `训练模型... (${pipeline.model.class})`}
                        {runProgress > 60 && runProgress <= 75 && `策略配置... (${pipeline.strategy.class})`}
                        {runProgress > 75 && runProgress <= 90 && '执行回测...'}
                        {runProgress > 90 && '生成分析报告...'}
                    </div>
                </div>
            )}

            {/* 运行结果 */}
            {runResults && (
                <div className="run-results">
                    <div className="results-header">
                        <h3>✅ qlib工作流完成</h3>
                        <div className="results-actions">
                            <button 
                                className="btn-primary"
                                onClick={() => onNavigate('results')}
                            >
                                📊 查看详细分析
                            </button>
                        </div>
                    </div>

                    <div className="results-summary">
                        <div className="result-section">
                            <h4>🤖 模型性能</h4>
                            <div className="metrics-row">
                                <div className="metric-item">
                                    <span className="metric-label">训练IC</span>
                                    <span className="metric-value">{runResults.model_performance.train_ic}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">验证IC</span>
                                    <span className="metric-value">{runResults.model_performance.valid_ic}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">测试IC</span>
                                    <span className="metric-value">{runResults.model_performance.test_ic}</span>
                                </div>
                            </div>
                        </div>

                        <div className="result-section">
                            <h4>📈 策略表现</h4>
                            <div className="metrics-row">
                                <div className="metric-item">
                                    <span className="metric-label">年化收益</span>
                                    <span className="metric-value positive">
                                        {(runResults.strategy_performance.annual_return * 100).toFixed(2)}%
                                    </span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">夏普比率</span>
                                    <span className="metric-value">{runResults.strategy_performance.sharpe_ratio}</span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">最大回撤</span>
                                    <span className="metric-value negative">
                                        {(runResults.strategy_performance.max_drawdown * 100).toFixed(2)}%
                                    </span>
                                </div>
                                <div className="metric-item">
                                    <span className="metric-label">信息比率</span>
                                    <span className="metric-value">{runResults.strategy_performance.information_ratio}</span>
                                </div>
                            </div>
                        </div>

                        <div className="result-section">
                            <h4>🔍 因子分析</h4>
                            <div className="top-factors">
                                {runResults.factor_analysis.top_factors.slice(0, 3).map((factor, idx) => (
                                    <div key={idx} className="factor-metric">
                                        <span className="factor-name">{factor.name}</span>
                                        <span className="factor-ic">IC: {factor.ic}</span>
                                        <span className="factor-weight">权重: {(factor.weight * 100).toFixed(1)}%</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};