// 因子工程工作坊组件
const { useState, useEffect, useRef } = React;

const FactorWorkshop = ({ 
    onSave = () => {}, 
    onNavigate = () => {},
    savedFactors = []
}) => {
    const [activeTab, setActiveTab] = useState('built-in');
    const [selectedCategory, setSelectedCategory] = useState('technical');
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFactor, setSelectedFactor] = useState(null);
    const [customExpression, setCustomExpression] = useState('');
    const [factorName, setFactorName] = useState('');
    const [factorDescription, setFactorDescription] = useState('');
    
    // AI对话相关状态
    const [aiChatMessages, setAiChatMessages] = useState([
        {
            id: 1,
            type: 'assistant',
            content: '你好！我是Qlib因子工程助手。我可以帮你生成因子表达式。请告诉我你想要什么类型的因子？',
            timestamp: new Date().toLocaleTimeString()
        }
    ]);
    const [aiInput, setAiInput] = useState('');
    const [isAiThinking, setIsAiThinking] = useState(false);
    const chatEndRef = useRef(null);

    // 内置因子库数据
    const builtInFactors = {
        technical: [
            {
                id: 'rsi',
                name: 'RSI相对强弱指数',
                expression: '(Sum(Max($close - Ref($close, 1), 0), 14) / Sum(Abs($close - Ref($close, 1)), 14)) * 100',
                description: '衡量价格变动速度和幅度的技术指标',
                category: 'technical',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['动量', '技术分析', '超买超卖']
            },
            {
                id: 'macd',
                name: 'MACD指标',
                expression: 'Mean($close, 12) - Mean($close, 26)',
                description: '移动平均收敛散度指标，用于判断趋势变化',
                category: 'technical',
                complexity: 'easy',
                returnPeriod: '中期',
                tags: ['趋势', '移动平均', '技术分析']
            },
            {
                id: 'bollinger_position',
                name: '布林带位置',
                expression: '($close - Mean($close, 20)) / (2 * Std($close, 20))',
                description: '股价在布林带中的相对位置',
                category: 'technical',
                complexity: 'medium',
                returnPeriod: '中期',
                tags: ['波动率', '均值回归', '技术分析']
            },
            {
                id: 'momentum_20',
                name: '20日动量',
                expression: '$close / Ref($close, 20) - 1',
                description: '20个交易日的价格动量',
                category: 'technical',
                complexity: 'easy',
                returnPeriod: '短期',
                tags: ['动量', '价格', '短期']
            }
        ],
        fundamental: [
            {
                id: 'pe_ratio',
                name: '市盈率因子',
                expression: '1 / $pe_ratio',
                description: '市盈率的倒数，用于价值投资',
                category: 'fundamental',
                complexity: 'easy',
                returnPeriod: '长期',
                tags: ['价值', '估值', '基本面']
            },
            {
                id: 'roe_growth',
                name: 'ROE增长率',
                expression: '($roe - Ref($roe, 252)) / Ref($roe, 252)',
                description: '净资产收益率的年度增长率',
                category: 'fundamental',
                complexity: 'medium',
                returnPeriod: '长期',
                tags: ['成长', '盈利能力', '基本面']
            },
            {
                id: 'debt_to_equity',
                name: '资产负债率因子',
                expression: '1 / (1 + $debt_to_equity)',
                description: '基于资产负债率的财务健康度指标',
                category: 'fundamental',
                complexity: 'easy',
                returnPeriod: '长期',
                tags: ['财务健康', '风险', '基本面']
            }
        ],
        volume: [
            {
                id: 'volume_momentum',
                name: '成交量动量',
                expression: 'Rank($volume / Mean($volume, 20))',
                description: '相对于历史平均的成交量排名',
                category: 'volume',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['成交量', '动量', '流动性']
            },
            {
                id: 'vwap_ratio',
                name: 'VWAP比率',
                expression: '$close / $vwap',
                description: '收盘价相对于成交量加权平均价格的比率',
                category: 'volume',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['成交量', '价格', 'VWAP']
            },
            {
                id: 'volume_price_trend',
                name: '量价趋势',
                expression: 'Corr($close / Ref($close, 1), $volume, 10)',
                description: '价格变化与成交量的相关性',
                category: 'volume',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['量价关系', '相关性', '趋势']
            }
        ],
        volatility: [
            {
                id: 'realized_volatility',
                name: '已实现波动率',
                expression: 'Std($close / Ref($close, 1), 20) * Sqrt(252)',
                description: '20日已实现波动率年化',
                category: 'volatility',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['波动率', '风险', '标准差']
            },
            {
                id: 'volatility_momentum',
                name: '波动率动量',
                expression: 'Std($close / Ref($close, 1), 5) / Std($close / Ref($close, 1), 20)',
                description: '短期波动率相对于长期波动率',
                category: 'volatility',
                complexity: 'medium',
                returnPeriod: '短期',
                tags: ['波动率', '动量', '比率']
            }
        ],
        cross_sectional: [
            {
                id: 'rank_return',
                name: '收益率排名',
                expression: 'Rank($close / Ref($close, 20) - 1)',
                description: '20日收益率的横截面排名',
                category: 'cross_sectional',
                complexity: 'easy',
                returnPeriod: '短期',
                tags: ['排名', '横截面', '收益率']
            },
            {
                id: 'zscore_volume',
                name: '成交量标准化',
                expression: 'Zscore($volume)',
                description: '成交量的横截面标准化',
                category: 'cross_sectional',
                complexity: 'easy',
                returnPeriod: '短期',
                tags: ['标准化', '成交量', '横截面']
            }
        ]
    };

    const factorCategories = {
        technical: { name: '技术指标', icon: '📈', desc: '基于价格和成交量的技术分析因子' },
        fundamental: { name: '基本面', icon: '📊', desc: '基于财务数据的基本面分析因子' },
        volume: { name: '成交量', icon: '📊', desc: '基于成交量特征的因子' },
        volatility: { name: '波动率', icon: '📉', desc: '基于价格波动特征的因子' },
        cross_sectional: { name: '横截面', icon: '🎯', desc: '横截面排名和标准化因子' }
    };

    // 过滤因子
    const filteredFactors = builtInFactors[selectedCategory]?.filter(factor =>
        factor.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        factor.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        factor.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
    ) || [];

    // AI对话功能
    const handleAiSubmit = async () => {
        if (!aiInput.trim()) return;

        const userMessage = {
            id: Date.now(),
            type: 'user',
            content: aiInput,
            timestamp: new Date().toLocaleTimeString()
        };

        setAiChatMessages(prev => [...prev, userMessage]);
        setAiInput('');
        setIsAiThinking(true);

        // 模拟AI响应
        setTimeout(() => {
            const aiResponse = generateAiResponse(aiInput);
            const assistantMessage = {
                id: Date.now() + 1,
                type: 'assistant',
                content: aiResponse.text,
                factorExpression: aiResponse.expression,
                timestamp: new Date().toLocaleTimeString()
            };

            setAiChatMessages(prev => [...prev, assistantMessage]);
            setIsAiThinking(false);
        }, 2000);
    };

    // AI响应生成（模拟）
    const generateAiResponse = (userInput) => {
        const input = userInput.toLowerCase();
        
        if (input.includes('动量') || input.includes('momentum')) {
            return {
                text: '基于你的需求，我为你生成了一个动量因子。这个因子衡量的是股票在过去N天的价格动量，适用于捕捉短期趋势：',
                expression: 'Rank($close / Ref($close, 20) - 1)'
            };
        } else if (input.includes('反转') || input.includes('mean reversion')) {
            return {
                text: '我为你生成了一个均值回归因子。这个因子通过对短期收益率取负号来捕捉反转效应：',
                expression: '-Rank(Sum($close / Ref($close, 1) - 1, 5))'
            };
        } else if (input.includes('波动率') || input.includes('volatility')) {
            return {
                text: '这里是一个波动率因子，用于衡量股票价格的波动程度。高波动率通常意味着更高的风险：',
                expression: 'Rank(Std($close / Ref($close, 1), 20))'
            };
        } else if (input.includes('成交量') || input.includes('volume')) {
            return {
                text: '我为你创建了一个成交量相关的因子。这个因子结合了价格变动和成交量信息：',
                expression: 'Rank(($close / Ref($close, 1) - 1) * $volume)'
            };
        } else if (input.includes('技术指标') || input.includes('rsi') || input.includes('macd')) {
            return {
                text: '基于技术分析，我推荐这个RSI改进版本。它结合了价格动量和成交量信息：',
                expression: '(Sum(Max($close - Ref($close, 1), 0) * $volume, 14) / Sum(Abs($close - Ref($close, 1)) * $volume, 14))'
            };
        } else {
            return {
                text: '基于你的描述，我为你生成了一个通用的价格动量因子。你可以根据需要调整参数：',
                expression: '($close - Mean($close, 20)) / Std($close, 20)'
            };
        }
    };

    // 使用AI生成的因子
    const useAiFactor = (expression, description) => {
        setCustomExpression(expression);
        setFactorDescription(description);
        setActiveTab('manual');
    };

    // 使用内置因子
    const useBuiltInFactor = (factor) => {
        setCustomExpression(factor.expression);
        setFactorName(factor.name);
        setFactorDescription(factor.description);
        setActiveTab('manual');
    };

    // 保存因子
    const handleSaveFactor = () => {
        if (!factorName.trim() || !customExpression.trim()) {
            alert('请填写因子名称和表达式');
            return;
        }

        const newFactor = {
            id: `custom_${Date.now()}`,
            name: factorName,
            expression: customExpression,
            description: factorDescription,
            category: 'custom',
            createTime: new Date().toLocaleString()
        };

        onSave(newFactor);
        alert('因子保存成功！');
        
        // 清空表单
        setFactorName('');
        setCustomExpression('');
        setFactorDescription('');
    };

    // 滚动到聊天底部
    useEffect(() => {
        chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [aiChatMessages]);

    return (
        <div className="factor-workshop">
            <div className="workshop-header">
                <h1>🧮 因子工程工作坊</h1>
                <div className="header-actions">
                    <button className="btn-secondary" onClick={() => onNavigate('factor')}>
                        📝 高级编辑器
                    </button>
                    <button className="btn-secondary" onClick={() => onNavigate('analysis')}>
                        📊 因子分析
                    </button>
                </div>
            </div>

            {/* 标签页导航 */}
            <div className="workshop-tabs">
                <button 
                    className={`workshop-tab ${activeTab === 'built-in' ? 'active' : ''}`}
                    onClick={() => setActiveTab('built-in')}
                >
                    📚 内置因子库
                </button>
                <button 
                    className={`workshop-tab ${activeTab === 'ai-chat' ? 'active' : ''}`}
                    onClick={() => setActiveTab('ai-chat')}
                >
                    🤖 AI助手
                </button>
                <button 
                    className={`workshop-tab ${activeTab === 'manual' ? 'active' : ''}`}
                    onClick={() => setActiveTab('manual')}
                >
                    ✏️ 手动编辑
                </button>
            </div>

            {/* 内置因子库 */}
            {activeTab === 'built-in' && (
                <div className="built-in-factors">
                    <div className="factors-sidebar">
                        <div className="search-box">
                            <input
                                type="text"
                                placeholder="搜索因子..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="search-input"
                            />
                        </div>
                        
                        <div className="category-list">
                            {Object.entries(factorCategories).map(([key, category]) => (
                                <div
                                    key={key}
                                    className={`category-item ${selectedCategory === key ? 'active' : ''}`}
                                    onClick={() => setSelectedCategory(key)}
                                >
                                    <span className="category-icon">{category.icon}</span>
                                    <div className="category-info">
                                        <div className="category-name">{category.name}</div>
                                        <div className="category-desc">{category.desc}</div>
                                        <div className="factor-count">
                                            {builtInFactors[key]?.length || 0} 个因子
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="factors-main">
                        <div className="factors-header">
                            <h3>{factorCategories[selectedCategory]?.name}因子</h3>
                            <div className="factors-count">共 {filteredFactors.length} 个因子</div>
                        </div>

                        <div className="factors-grid">
                            {filteredFactors.map(factor => (
                                <div key={factor.id} className="factor-card">
                                    <div className="factor-card-header">
                                        <h4>{factor.name}</h4>
                                        <div className="factor-badges">
                                            <span className={`complexity-badge ${factor.complexity}`}>
                                                {factor.complexity === 'easy' ? '简单' : 
                                                 factor.complexity === 'medium' ? '中等' : '复杂'}
                                            </span>
                                            <span className="period-badge">{factor.returnPeriod}</span>
                                        </div>
                                    </div>
                                    
                                    <div className="factor-expression-preview">
                                        {factor.expression}
                                    </div>
                                    
                                    <div className="factor-description">
                                        {factor.description}
                                    </div>
                                    
                                    <div className="factor-tags">
                                        {factor.tags.map(tag => (
                                            <span key={tag} className="factor-tag">{tag}</span>
                                        ))}
                                    </div>
                                    
                                    <div className="factor-actions">
                                        <button 
                                            className="btn-secondary"
                                            onClick={() => setSelectedFactor(factor)}
                                        >
                                            预览
                                        </button>
                                        <button 
                                            className="btn-primary"
                                            onClick={() => useBuiltInFactor(factor)}
                                        >
                                            使用
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {/* AI对话助手 */}
            {activeTab === 'ai-chat' && (
                <div className="ai-chat-container">
                    <div className="chat-messages">
                        {aiChatMessages.map(message => (
                            <div key={message.id} className={`message ${message.type}`}>
                                <div className="message-avatar">
                                    {message.type === 'user' ? '👤' : '🤖'}
                                </div>
                                <div className="message-content">
                                    <div className="message-text">{message.content}</div>
                                    {message.factorExpression && (
                                        <div className="generated-factor">
                                            <div className="factor-expression-code">
                                                {message.factorExpression}
                                            </div>
                                            <button 
                                                className="btn-primary btn-small"
                                                onClick={() => useAiFactor(message.factorExpression, message.content)}
                                            >
                                                使用这个因子
                                            </button>
                                        </div>
                                    )}
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
                        <div ref={chatEndRef} />
                    </div>

                    <div className="chat-input-container">
                        <div className="chat-suggestions">
                            <button 
                                className="suggestion-btn"
                                onClick={() => setAiInput('我想要一个动量因子')}
                            >
                                动量因子
                            </button>
                            <button 
                                className="suggestion-btn"
                                onClick={() => setAiInput('生成一个波动率因子')}
                            >
                                波动率因子
                            </button>
                            <button 
                                className="suggestion-btn"
                                onClick={() => setAiInput('创建一个均值回归因子')}
                            >
                                均值回归因子
                            </button>
                            <button 
                                className="suggestion-btn"
                                onClick={() => setAiInput('我需要一个成交量相关的因子')}
                            >
                                成交量因子
                            </button>
                        </div>
                        
                        <div className="chat-input-box">
                            <input
                                type="text"
                                placeholder="描述你想要的因子类型，比如：我想要一个捕捉短期动量的因子..."
                                value={aiInput}
                                onChange={(e) => setAiInput(e.target.value)}
                                onKeyPress={(e) => e.key === 'Enter' && handleAiSubmit()}
                                className="chat-input"
                                disabled={isAiThinking}
                            />
                            <button 
                                className="send-btn"
                                onClick={handleAiSubmit}
                                disabled={isAiThinking || !aiInput.trim()}
                            >
                                发送
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 手动编辑 */}
            {activeTab === 'manual' && (
                <div className="manual-editor">
                    <div className="editor-form">
                        <div className="form-section">
                            <h3>📋 因子信息</h3>
                            <div className="form-group">
                                <label>因子名称</label>
                                <input
                                    type="text"
                                    value={factorName}
                                    onChange={(e) => setFactorName(e.target.value)}
                                    placeholder="请输入因子名称"
                                    className="form-input"
                                />
                            </div>
                            <div className="form-group">
                                <label>因子描述</label>
                                <textarea
                                    value={factorDescription}
                                    onChange={(e) => setFactorDescription(e.target.value)}
                                    placeholder="请描述因子的作用和特点"
                                    className="form-textarea"
                                    rows="3"
                                />
                            </div>
                        </div>

                        <div className="form-section">
                            <h3>⚡ 因子表达式</h3>
                            <div className="form-group">
                                <textarea
                                    value={customExpression}
                                    onChange={(e) => setCustomExpression(e.target.value)}
                                    placeholder="请输入因子表达式，例如: ($close - Mean($close, 20)) / Std($close, 20)"
                                    className="expression-input"
                                    rows="6"
                                />
                            </div>
                            
                            <div className="editor-actions">
                                <button className="btn-secondary" onClick={() => onNavigate('factor')}>
                                    高级编辑器
                                </button>
                                <button className="btn-success" onClick={handleSaveFactor}>
                                    💾 保存因子
                                </button>
                            </div>
                        </div>
                    </div>

                    <div className="quick-examples">
                        <h3>💡 快速示例</h3>
                        <div className="examples-grid">
                            <div className="example-item" onClick={() => setCustomExpression('Rank($close / Ref($close, 20) - 1)')}>
                                <h4>价格动量</h4>
                                <code>Rank($close / Ref($close, 20) - 1)</code>
                            </div>
                            <div className="example-item" onClick={() => setCustomExpression('-Rank(Sum($close / Ref($close, 1) - 1, 5))')}>
                                <h4>短期反转</h4>
                                <code>-Rank(Sum($close / Ref($close, 1) - 1, 5))</code>
                            </div>
                            <div className="example-item" onClick={() => setCustomExpression('($close - Mean($close, 20)) / Std($close, 20)')}>
                                <h4>标准化偏离</h4>
                                <code>($close - Mean($close, 20)) / Std($close, 20)</code>
                            </div>
                            <div className="example-item" onClick={() => setCustomExpression('Rank($volume / Mean($volume, 20))')}>
                                <h4>成交量异常</h4>
                                <code>Rank($volume / Mean($volume, 20))</code>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* 因子详情模态框 */}
            {selectedFactor && (
                <div className="modal-overlay" onClick={() => setSelectedFactor(null)}>
                    <div className="modal factor-detail-modal" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>{selectedFactor.name}</h2>
                            <button className="close-btn" onClick={() => setSelectedFactor(null)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="factor-detail">
                                <div className="detail-section">
                                    <h3>表达式</h3>
                                    <div className="expression-display">{selectedFactor.expression}</div>
                                </div>
                                
                                <div className="detail-section">
                                    <h3>说明</h3>
                                    <p>{selectedFactor.description}</p>
                                </div>
                                
                                <div className="detail-section">
                                    <h3>特征</h3>
                                    <div className="factor-properties">
                                        <div className="property-item">
                                            <span className="property-label">复杂度:</span>
                                            <span className={`property-value ${selectedFactor.complexity}`}>
                                                {selectedFactor.complexity === 'easy' ? '简单' : 
                                                 selectedFactor.complexity === 'medium' ? '中等' : '复杂'}
                                            </span>
                                        </div>
                                        <div className="property-item">
                                            <span className="property-label">收益周期:</span>
                                            <span className="property-value">{selectedFactor.returnPeriod}</span>
                                        </div>
                                        <div className="property-item">
                                            <span className="property-label">类别:</span>
                                            <span className="property-value">{factorCategories[selectedFactor.category]?.name}</span>
                                        </div>
                                    </div>
                                </div>
                                
                                <div className="detail-section">
                                    <h3>标签</h3>
                                    <div className="factor-tags">
                                        {selectedFactor.tags.map(tag => (
                                            <span key={tag} className="factor-tag">{tag}</span>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setSelectedFactor(null)}>
                                关闭
                            </button>
                            <button className="btn-primary" onClick={() => {
                                useBuiltInFactor(selectedFactor);
                                setSelectedFactor(null);
                            }}>
                                使用此因子
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};