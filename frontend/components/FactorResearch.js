// 因子研究工作台 - 整合因子工程、编辑器和分析功能
const { useState, useEffect } = React;

const FactorResearch = ({ 
    onSave = () => {}, 
    onTestFactor = () => {}, 
    savedFactors = [],
    onNavigate = () => {} 
}) => {
    const [activeModule, setActiveModule] = useState('discovery');
    const [selectedFactor, setSelectedFactor] = useState(null);
    const [workspaceFactors, setWorkspaceFactors] = useState([]);

    // 模块配置
    const modules = [
        { 
            key: 'discovery', 
            label: '因子发现', 
            icon: '🔍', 
            desc: '基于qlib内置库和AI助手发现因子' 
        },
        { 
            key: 'editor', 
            label: '表达式编辑', 
            icon: '✏️', 
            desc: '高级因子表达式编辑器' 
        },
        { 
            key: 'analysis', 
            label: '因子分析', 
            icon: '📊', 
            desc: '因子性能分析和可视化' 
        },
        { 
            key: 'library', 
            label: '因子库', 
            icon: '📚', 
            desc: '管理和组织您的因子' 
        }
    ];

    // 因子分类（基于qlib）
    const factorCategories = [
        {
            id: 'price',
            name: '价格类因子',
            icon: '💰',
            desc: '基于价格数据的技术指标',
            count: 45,
            factors: [
                { name: 'ROC', expression: '$close / Ref($close, 20) - 1', desc: '20日价格变化率' },
                { name: 'RSV', expression: '($close - Min($low, 9)) / (Max($high, 9) - Min($low, 9))', desc: 'RSV指标' },
                { name: 'BIAS', expression: '$close / Mean($close, 20) - 1', desc: '20日乖离率' },
                { name: 'CCI', expression: '($close - Mean($close, 14)) / (0.015 * Mean(Abs($close - Mean($close, 14)), 14))', desc: 'CCI指标' }
            ]
        },
        {
            id: 'volume',
            name: '成交量因子',
            icon: '📊',
            desc: '基于成交量的流动性指标',
            count: 28,
            factors: [
                { name: 'VSTD', expression: 'Std($volume, 20)', desc: '20日成交量标准差' },
                { name: 'VWAP', expression: 'Sum($volume * $close, 5) / Sum($volume, 5)', desc: '5日成交量加权平均价' },
                { name: 'VR', expression: 'Sum(If($close > Ref($close, 1), $volume, 0), 26) / Sum($volume, 26)', desc: '成交量比率' },
                { name: 'VROC', expression: '$volume / Ref($volume, 12) - 1', desc: '12日成交量变化率' }
            ]
        },
        {
            id: 'momentum',
            name: '动量因子',
            icon: '🚀',
            desc: '价格动量和趋势跟踪',
            count: 32,
            factors: [
                { name: 'MOM', expression: '$close / Ref($close, 10) - 1', desc: '10日动量' },
                { name: 'MACD', expression: 'EMA($close, 12) - EMA($close, 26)', desc: 'MACD指标' },
                { name: 'TRIX', expression: 'EMA(EMA(EMA(Log($close), 12), 12), 12)', desc: 'TRIX指标' },
                { name: 'UOS', expression: '(4*RSV(7) + 2*RSV(14) + RSV(28)) / 7', desc: '终极震荡指标' }
            ]
        },
        {
            id: 'mean_reversion',
            name: '均值回归',
            icon: '🔄',
            desc: '基于均值回归的反转因子',
            count: 19,
            factors: [
                { name: 'RSI', expression: '100 - 100 / (1 + Mean(Max($close - Ref($close, 1), 0), 14) / Mean(Abs($close - Ref($close, 1)), 14))', desc: 'RSI指标' },
                { name: 'WR', expression: '($high - $close) / ($high - $low)', desc: 'WR威廉指标' },
                { name: 'STOCH', expression: '($close - Min($low, 9)) / (Max($high, 9) - Min($low, 9))', desc: '随机指标KD' },
                { name: 'BBANDS', expression: '($close - Mean($close, 20)) / Std($close, 20)', desc: '布林带位置' }
            ]
        },
        {
            id: 'volatility',
            name: '波动率因子',
            icon: '📈',
            desc: '价格波动和风险度量',
            count: 15,
            factors: [
                { name: 'ATR', expression: 'Mean(Max($high - $low, Max(Abs($high - Ref($close, 1)), Abs($low - Ref($close, 1)))), 14)', desc: '真实波动范围' },
                { name: 'STDDEV', expression: 'Std($close / Ref($close, 1) - 1, 20)', desc: '20日收益率标准差' },
                { name: 'BETA', expression: 'Corr($close / Ref($close, 1), $benchmark_return, 60)', desc: '60日Beta系数' },
                { name: 'PVOL', expression: 'Mean(($high - $low) / $close, 10)', desc: '价格波动率' }
            ]
        }
    ];

    // AI 对话状态
    const [aiMessages, setAiMessages] = useState([
        {
            role: 'assistant',
            content: '您好！我是qlib因子研究助手。我可以帮您：\n• 发现和创建新因子\n• 解释因子含义和计算逻辑\n• 分析因子性能\n• 优化因子表达式\n\n请问您想要研究什么类型的因子？',
            timestamp: new Date().toLocaleTimeString()
        }
    ]);
    const [currentInput, setCurrentInput] = useState('');
    const [isAiThinking, setIsAiThinking] = useState(false);

    // 因子编辑器状态
    const [factorExpression, setFactorExpression] = useState('');
    const [factorName, setFactorName] = useState('');
    const [factorDescription, setFactorDescription] = useState('');
    const [syntaxErrors, setSyntaxErrors] = useState([]);
    const [testResult, setTestResult] = useState(null);

    // 处理因子选择
    const handleFactorSelect = (factor) => {
        setSelectedFactor(factor);
        setFactorExpression(factor.expression || '');
        setFactorName(factor.name || '');
        setFactorDescription(factor.desc || factor.description || '');
        
        // 如果选择了因子，切换到编辑器
        if (activeModule === 'discovery') {
            setActiveModule('editor');
        }
    };

    // 处理AI对话
    const handleAiMessage = async (message) => {
        if (!message.trim()) return;

        // 添加用户消息
        setAiMessages(prev => [...prev, {
            role: 'user',
            content: message,
            timestamp: new Date().toLocaleTimeString()
        }]);

        setCurrentInput('');
        setIsAiThinking(true);

        // 模拟AI响应
        setTimeout(() => {
            let response = '';
            if (message.includes('动量') || message.includes('momentum')) {
                response = `基于您的需求，我推荐几个动量因子：

**1. 多周期动量复合因子**
\`\`\`
(Rank($close / Ref($close, 5) - 1) + 
 Rank($close / Ref($close, 10) - 1) + 
 Rank($close / Ref($close, 20) - 1)) / 3
\`\`\`

**2. 成交量确认动量**
\`\`\`
($close / Ref($close, 10) - 1) * Rank($volume / Mean($volume, 20))
\`\`\`

这些因子结合了价格动量和成交量信息，通常在趋势市场中表现较好。您想了解哪个因子的详细分析？`;
            } else if (message.includes('反转') || message.includes('reversion')) {
                response = `反转因子通常捕捉短期过度反应，这里是几个有效的反转因子：

**1. 标准化反转因子**
\`\`\`
-Rank(Sum($close / Ref($close, 1) - 1, 5)) / Std($close / Ref($close, 1) - 1, 20)
\`\`\`

**2. RSI反转信号**
\`\`\`
If(RSI(14) > 70, -1, If(RSI(14) < 30, 1, 0))
\`\`\`

这些因子在震荡市场中通常有较好的预测能力。`;
            } else {
                response = `我理解您的问题。让我为您推荐一些qlib中常用的因子类型：

🔍 **技术指标类**：RSI, MACD, KDJ等经典指标
📊 **价量关系**：价格与成交量的相关性分析  
📈 **趋势因子**：移动平均、动量、波动率等
🔄 **均值回归**：基于统计特性的反转信号

您具体想研究哪个方向？我可以提供更详细的代码和解释。`;
            }

            setAiMessages(prev => [...prev, {
                role: 'assistant',
                content: response,
                timestamp: new Date().toLocaleTimeString()
            }]);
            setIsAiThinking(false);
        }, 2000);
    };

    // 语法检查
    const checkSyntax = (expression) => {
        const errors = [];
        
        // 检查括号匹配
        let openCount = 0;
        for (let char of expression) {
            if (char === '(') openCount++;
            if (char === ')') openCount--;
            if (openCount < 0) {
                errors.push('右括号多余');
                break;
            }
        }
        if (openCount > 0) errors.push('缺少右括号');

        // 检查qlib字段
        const fieldPattern = /\$[a-zA-Z_][a-zA-Z0-9_]*/g;
        const validFields = ['$close', '$open', '$high', '$low', '$volume', '$amount', '$factor', '$change'];
        const fieldMatches = expression.match(fieldPattern) || [];
        fieldMatches.forEach(field => {
            if (!validFields.includes(field)) {
                errors.push(`未知字段: ${field}`);
            }
        });

        // 检查qlib函数
        const functionPattern = /([A-Z][a-zA-Z]*)\s*\(/g;
        const validFunctions = [
            'Ref', 'Mean', 'Sum', 'Std', 'Max', 'Min', 'Delta', 'Corr',
            'Rank', 'Zscore', 'Neutralize', 'If', 'Sign', 'Abs', 'Log', 'Power',
            'EMA', 'RSI', 'MACD', 'ATR', 'BIAS', 'ROC'
        ];
        let match;
        while ((match = functionPattern.exec(expression)) !== null) {
            if (!validFunctions.includes(match[1])) {
                errors.push(`未知函数: ${match[1]}`);
            }
        }

        setSyntaxErrors(errors);
        return errors.length === 0;
    };

    // 测试因子
    const handleTestFactor = async () => {
        if (!factorExpression.trim()) {
            alert('请输入因子表达式');
            return;
        }

        if (!checkSyntax(factorExpression)) {
            alert('表达式存在语法错误，请检查');
            return;
        }

        // 模拟qlib因子测试
        setTimeout(() => {
            const mockResult = {
                ic: (Math.random() * 0.08 - 0.04).toFixed(4),
                icIR: (Math.random() * 1.5 + 0.8).toFixed(2),
                rank_ic: (Math.random() * 0.15 - 0.075).toFixed(4),
                rank_icIR: (Math.random() * 1.8 + 0.9).toFixed(2),
                turnover: (Math.random() * 0.4 + 0.2).toFixed(3),
                coverage: (Math.random() * 0.15 + 0.82).toFixed(3),
                validPeriods: Math.floor(Math.random() * 100 + 180),
                // 分年度表现
                yearlyPerformance: [
                    { year: 2020, ic: (Math.random() * 0.1 - 0.05).toFixed(4), rank_ic: (Math.random() * 0.12 - 0.06).toFixed(4) },
                    { year: 2021, ic: (Math.random() * 0.1 - 0.05).toFixed(4), rank_ic: (Math.random() * 0.12 - 0.06).toFixed(4) },
                    { year: 2022, ic: (Math.random() * 0.1 - 0.05).toFixed(4), rank_ic: (Math.random() * 0.12 - 0.06).toFixed(4) },
                    { year: 2023, ic: (Math.random() * 0.1 - 0.05).toFixed(4), rank_ic: (Math.random() * 0.12 - 0.06).toFixed(4) }
                ]
            };
            setTestResult(mockResult);
        }, 1500);
    };

    // 保存因子
    const handleSaveFactor = () => {
        if (!factorName.trim() || !factorExpression.trim()) {
            alert('请输入因子名称和表达式');
            return;
        }

        if (!checkSyntax(factorExpression)) {
            alert('表达式存在语法错误，请检查');
            return;
        }

        const newFactor = {
            id: `factor_${Date.now()}`,
            name: factorName,
            expression: factorExpression,
            description: factorDescription,
            createTime: new Date().toLocaleString(),
            performance: testResult,
            status: 'active'
        };

        onSave(newFactor);
        setWorkspaceFactors(prev => [...prev, newFactor]);
        alert('因子保存成功！');
    };

    // 渲染因子发现模块
    const renderDiscoveryModule = () => (
        <div className="discovery-module">
            <div className="module-layout">
                {/* 左侧：因子分类 */}
                <div className="factor-categories">
                    <div className="categories-header">
                        <h3>🔍 qlib因子库</h3>
                        <div className="search-box">
                            <input 
                                type="text" 
                                placeholder="搜索因子..." 
                                className="search-input"
                            />
                        </div>
                    </div>
                    
                    <div className="category-list">
                        {factorCategories.map(category => (
                            <div key={category.id} className="category-card">
                                <div className="category-header">
                                    <span className="category-icon">{category.icon}</span>
                                    <div className="category-info">
                                        <h4>{category.name}</h4>
                                        <p>{category.desc}</p>
                                        <span className="factor-count">{category.count}个因子</span>
                                    </div>
                                </div>
                                
                                <div className="factor-preview">
                                    {category.factors.slice(0, 3).map((factor, idx) => (
                                        <div 
                                            key={idx} 
                                            className="factor-item"
                                            onClick={() => handleFactorSelect(factor)}
                                        >
                                            <div className="factor-name">{factor.name}</div>
                                            <div className="factor-desc">{factor.desc}</div>
                                            <div className="factor-expression">{factor.expression}</div>
                                        </div>
                                    ))}
                                    {category.factors.length > 3 && (
                                        <div className="more-factors">
                                            +{category.factors.length - 3} 更多因子
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* 右侧：AI助手 */}
                <div className="ai-assistant">
                    <div className="ai-header">
                        <h3>🤖 Qlib因子研究助手</h3>
                        <div className="ai-status">在线</div>
                    </div>
                    
                    <div className="chat-messages">
                        {aiMessages.map((message, index) => (
                            <div key={index} className={`message ${message.role}`}>
                                <div className="message-avatar">
                                    {message.role === 'user' ? '👤' : '🤖'}
                                </div>
                                <div className="message-content">
                                    <div className="message-text">{message.content}</div>
                                    <div className="message-time">{message.timestamp}</div>
                                </div>
                            </div>
                        ))}
                        
                        {isAiThinking && (
                            <div className="message assistant">
                                <div className="message-avatar">🤖</div>
                                <div className="message-content">
                                    <div className="thinking-indicator">
                                        <span>正在思考</span>
                                        <div className="thinking-dots">
                                            <span>.</span><span>.</span><span>.</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                    
                    <div className="chat-input-container">
                        <div className="quick-questions">
                            <button onClick={() => handleAiMessage('推荐一些动量因子')}>动量因子</button>
                            <button onClick={() => handleAiMessage('反转因子有哪些？')}>反转因子</button>
                            <button onClick={() => handleAiMessage('如何创建成交量因子？')}>成交量因子</button>
                            <button onClick={() => handleAiMessage('解释RSI指标')}>技术指标</button>
                        </div>
                        
                        <div className="input-box">
                            <input
                                type="text"
                                value={currentInput}
                                onChange={(e) => setCurrentInput(e.target.value)}
                                onKeyPress={(e) => e.key === 'Enter' && handleAiMessage(currentInput)}
                                placeholder="询问因子相关问题..."
                                className="chat-input"
                            />
                            <button 
                                onClick={() => handleAiMessage(currentInput)}
                                disabled={!currentInput.trim() || isAiThinking}
                                className="send-button"
                            >
                                发送
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    // 渲染编辑器模块
    const renderEditorModule = () => (
        <div className="editor-module">
            <div className="editor-layout">
                <div className="editor-main">
                    <div className="factor-form">
                        <h3>📝 因子表达式编辑器</h3>
                        
                        <div className="form-section">
                            <div className="form-row">
                                <div className="form-group">
                                    <label>因子名称</label>
                                    <input
                                        type="text"
                                        value={factorName}
                                        onChange={(e) => setFactorName(e.target.value)}
                                        placeholder="输入因子名称"
                                        className="form-input"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>因子描述</label>
                                    <input
                                        type="text"
                                        value={factorDescription}
                                        onChange={(e) => setFactorDescription(e.target.value)}
                                        placeholder="描述因子的含义和用途"
                                        className="form-input"
                                    />
                                </div>
                            </div>
                        </div>

                        <div className="expression-editor">
                            <div className="editor-header">
                                <label>qlib因子表达式</label>
                                <div className="editor-actions">
                                    <button onClick={handleTestFactor} className="btn-primary btn-sm">
                                        🔍 测试因子
                                    </button>
                                    <button onClick={handleSaveFactor} className="btn-success btn-sm">
                                        💾 保存因子
                                    </button>
                                </div>
                            </div>
                            
                            <textarea
                                value={factorExpression}
                                onChange={(e) => {
                                    setFactorExpression(e.target.value);
                                    checkSyntax(e.target.value);
                                }}
                                placeholder="输入qlib因子表达式，如：($close - Mean($close, 20)) / Std($close, 20)"
                                className="expression-textarea"
                                rows="6"
                            />
                            
                            {syntaxErrors.length > 0 && (
                                <div className="syntax-errors">
                                    <h4>⚠️ 语法错误:</h4>
                                    {syntaxErrors.map((error, index) => (
                                        <div key={index} className="error-item">{error}</div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                <div className="editor-sidebar">
                    <div className="syntax-panel">
                        <h4>📖 qlib语法参考</h4>
                        
                        <div className="syntax-section">
                            <h5>基础字段</h5>
                            <div className="syntax-items">
                                <span className="syntax-item">$close</span>
                                <span className="syntax-item">$open</span>
                                <span className="syntax-item">$high</span>
                                <span className="syntax-item">$low</span>
                                <span className="syntax-item">$volume</span>
                                <span className="syntax-item">$amount</span>
                            </div>
                        </div>
                        
                        <div className="syntax-section">
                            <h5>时序函数</h5>
                            <div className="syntax-items">
                                <span className="syntax-item">Ref()</span>
                                <span className="syntax-item">Mean()</span>
                                <span className="syntax-item">Sum()</span>
                                <span className="syntax-item">Std()</span>
                                <span className="syntax-item">Max()</span>
                                <span className="syntax-item">Min()</span>
                            </div>
                        </div>
                        
                        <div className="syntax-section">
                            <h5>横截面函数</h5>
                            <div className="syntax-items">
                                <span className="syntax-item">Rank()</span>
                                <span className="syntax-item">Zscore()</span>
                                <span className="syntax-item">Neutralize()</span>
                            </div>
                        </div>
                        
                        <div className="syntax-section">
                            <h5>技术指标</h5>
                            <div className="syntax-items">
                                <span className="syntax-item">RSI()</span>
                                <span className="syntax-item">MACD()</span>
                                <span className="syntax-item">EMA()</span>
                                <span className="syntax-item">ATR()</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* 测试结果 */}
            {testResult && (
                <div className="test-results">
                    <h3>📊 因子测试结果</h3>
                    <div className="results-grid">
                        <div className="result-card">
                            <div className="result-label">IC值</div>
                            <div className={`result-value ${parseFloat(testResult.ic) > 0 ? 'positive' : 'negative'}`}>
                                {testResult.ic}
                            </div>
                        </div>
                        <div className="result-card">
                            <div className="result-label">IC信息比率</div>
                            <div className="result-value">{testResult.icIR}</div>
                        </div>
                        <div className="result-card">
                            <div className="result-label">Rank IC</div>
                            <div className={`result-value ${parseFloat(testResult.rank_ic) > 0 ? 'positive' : 'negative'}`}>
                                {testResult.rank_ic}
                            </div>
                        </div>
                        <div className="result-card">
                            <div className="result-label">换手率</div>
                            <div className="result-value">{testResult.turnover}</div>
                        </div>
                        <div className="result-card">
                            <div className="result-label">覆盖率</div>
                            <div className="result-value">{testResult.coverage}</div>
                        </div>
                        <div className="result-card">
                            <div className="result-label">有效期数</div>
                            <div className="result-value">{testResult.validPeriods}</div>
                        </div>
                    </div>
                    
                    <div className="yearly-performance">
                        <h4>分年度表现</h4>
                        <table className="performance-table">
                            <thead>
                                <tr>
                                    <th>年份</th>
                                    <th>IC值</th>
                                    <th>Rank IC</th>
                                </tr>
                            </thead>
                            <tbody>
                                {testResult.yearlyPerformance.map(year => (
                                    <tr key={year.year}>
                                        <td>{year.year}</td>
                                        <td className={parseFloat(year.ic) > 0 ? 'positive' : 'negative'}>
                                            {year.ic}
                                        </td>
                                        <td className={parseFloat(year.rank_ic) > 0 ? 'positive' : 'negative'}>
                                            {year.rank_ic}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}
        </div>
    );

    // 渲染分析模块
    const renderAnalysisModule = () => (
        <div className="analysis-module">
            <h3>📊 因子分析</h3>
            {selectedFactor ? (
                <FactorAnalysis 
                    factorData={selectedFactor} 
                    onNavigate={onNavigate}
                />
            ) : (
                <div className="no-factor-selected">
                    <div className="placeholder-icon">📊</div>
                    <h4>请选择要分析的因子</h4>
                    <p>从因子库中选择一个因子，或在编辑器中测试一个因子表达式</p>
                    <button 
                        className="btn-primary"
                        onClick={() => setActiveModule('library')}
                    >
                        浏览因子库
                    </button>
                </div>
            )}
        </div>
    );

    // 渲染因子库模块
    const renderLibraryModule = () => (
        <div className="library-module">
            <div className="library-header">
                <h3>📚 我的因子库</h3>
                <div className="library-stats">
                    <span>共 {savedFactors.length} 个因子</span>
                    <span>•</span>
                    <span>工作区 {workspaceFactors.length} 个</span>
                </div>
            </div>

            {savedFactors.length === 0 && workspaceFactors.length === 0 ? (
                <div className="empty-library">
                    <div className="empty-icon">📭</div>
                    <div className="empty-title">暂无保存的因子</div>
                    <div className="empty-desc">
                        从因子发现开始，或使用编辑器创建您的第一个因子
                    </div>
                    <div className="empty-actions">
                        <button 
                            className="btn-primary"
                            onClick={() => setActiveModule('discovery')}
                        >
                            🔍 因子发现
                        </button>
                        <button 
                            className="btn-secondary"
                            onClick={() => setActiveModule('editor')}
                        >
                            ✏️ 创建因子
                        </button>
                    </div>
                </div>
            ) : (
                <div className="factors-grid">
                    {[...savedFactors, ...workspaceFactors].map(factor => (
                        <div key={factor.id} className="factor-card">
                            <div className="factor-card-header">
                                <h4>{factor.name}</h4>
                                <div className={`factor-status ${factor.status}`}>
                                    {factor.status === 'active' ? '活跃' : '停用'}
                                </div>
                            </div>
                            
                            <div className="factor-expression">
                                {factor.expression}
                            </div>
                            
                            <div className="factor-description">
                                {factor.description || '暂无描述'}
                            </div>
                            
                            {factor.performance && (
                                <div className="factor-metrics">
                                    <div className="metric">
                                        <span>IC: </span>
                                        <span className={parseFloat(factor.performance.ic) > 0 ? 'positive' : 'negative'}>
                                            {factor.performance.ic}
                                        </span>
                                    </div>
                                    <div className="metric">
                                        <span>IR: </span>
                                        <span>{factor.performance.icIR}</span>
                                    </div>
                                </div>
                            )}
                            
                            <div className="factor-actions">
                                <button 
                                    className="btn-sm btn-secondary"
                                    onClick={() => handleFactorSelect(factor)}
                                >
                                    编辑
                                </button>
                                <button 
                                    className="btn-sm btn-primary"
                                    onClick={() => {
                                        setSelectedFactor(factor);
                                        setActiveModule('analysis');
                                    }}
                                >
                                    分析
                                </button>
                                <button 
                                    className="btn-sm btn-success"
                                    onClick={() => onNavigate('workflow')}
                                >
                                    使用
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );

    // 渲染当前模块
    const renderActiveModule = () => {
        switch(activeModule) {
            case 'discovery': return renderDiscoveryModule();
            case 'editor': return renderEditorModule();
            case 'analysis': return renderAnalysisModule();
            case 'library': return renderLibraryModule();
            default: return renderDiscoveryModule();
        }
    };

    return (
        <div className="factor-research">
            <div className="research-header">
                <h1>🧮 因子研究工作台</h1>
                <div className="header-subtitle">
                    基于qlib的因子发现、开发、测试和分析平台
                </div>
            </div>

            {/* 模块导航 */}
            <div className="module-navigator">
                {modules.map(module => (
                    <button
                        key={module.key}
                        className={`module-tab ${activeModule === module.key ? 'active' : ''}`}
                        onClick={() => setActiveModule(module.key)}
                    >
                        <span className="tab-icon">{module.icon}</span>
                        <div className="tab-content">
                            <div className="tab-label">{module.label}</div>
                            <div className="tab-desc">{module.desc}</div>
                        </div>
                    </button>
                ))}
            </div>

            {/* 主要内容 */}
            <div className="research-content">
                {renderActiveModule()}
            </div>
        </div>
    );
};