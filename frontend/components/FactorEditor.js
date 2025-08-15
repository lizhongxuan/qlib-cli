// 因子表达式编辑器组件
const { useState, useEffect, useRef } = React;

const FactorEditor = ({ 
    onSave = () => {}, 
    onTestFactor = () => {}, 
    savedFactors = [],
    onNavigate = () => {} 
}) => {
    const [activeTab, setActiveTab] = useState('editor');
    const [expression, setExpression] = useState('');
    const [factorName, setFactorName] = useState('');
    const [description, setDescription] = useState('');
    const [testResult, setTestResult] = useState(null);
    const [isTestingFactor, setIsTestingFactor] = useState(false);
    const [syntaxErrors, setSyntaxErrors] = useState([]);
    const [suggestions, setSuggestions] = useState([]);
    const [cursorPosition, setCursorPosition] = useState(0);
    const editorRef = useRef(null);

    // 因子模板
    const factorTemplates = [
        {
            name: '价格趋势因子',
            expression: '($close - Mean($close, 20)) / Std($close, 20)',
            description: '20日价格偏离度'
        },
        {
            name: '成交量因子',
            expression: 'Rank(($close / Ref($close, 1) - 1) * $volume)',
            description: '成交量加权收益率排名'
        },
        {
            name: '反转因子',
            expression: '-Rank(Sum($close / Ref($close, 1) - 1, 5))',
            description: '5日收益率反转'
        },
        {
            name: '动量因子',
            expression: 'Rank($close / Ref($close, 20) - 1)',
            description: '20日价格动量'
        },
        {
            name: '波动率因子',
            expression: 'Rank(Std($close / Ref($close, 1), 20))',
            description: '20日收益率波动率'
        }
    ];

    // 操作符和函数库
    const operators = {
        '基础操作': ['+', '-', '*', '/', '(', ')'],
        '比较操作': ['>', '<', '>=', '<=', '==', '!='],
        '逻辑操作': ['&&', '||', '!'],
        '时序函数': ['Ref', 'Mean', 'Sum', 'Std', 'Max', 'Min', 'Delta', 'Corr'],
        '横截面函数': ['Rank', 'Zscore', 'Neutralize'],
        '条件函数': ['If', 'Sign', 'Abs', 'Log', 'Power'],
        '基础字段': ['$close', '$open', '$high', '$low', '$volume', '$amount']
    };

    // 语法检查
    const checkSyntax = (expr) => {
        const errors = [];
        
        // 检查括号匹配
        let openCount = 0;
        for (let char of expr) {
            if (char === '(') openCount++;
            if (char === ')') openCount--;
            if (openCount < 0) {
                errors.push('右括号多余');
                break;
            }
        }
        if (openCount > 0) {
            errors.push('缺少右括号');
        }

        // 检查基础字段格式
        const fieldPattern = /\$[a-zA-Z_][a-zA-Z0-9_]*/g;
        const validFields = ['$close', '$open', '$high', '$low', '$volume', '$amount'];
        const fieldMatches = expr.match(fieldPattern) || [];
        fieldMatches.forEach(field => {
            if (!validFields.includes(field)) {
                errors.push(`未知字段: ${field}`);
            }
        });

        // 检查函数格式
        const functionPattern = /([A-Z][a-zA-Z]*)\s*\(/g;
        const validFunctions = ['Ref', 'Mean', 'Sum', 'Std', 'Max', 'Min', 'Delta', 'Corr', 
                               'Rank', 'Zscore', 'Neutralize', 'If', 'Sign', 'Abs', 'Log', 'Power'];
        let match;
        while ((match = functionPattern.exec(expr)) !== null) {
            if (!validFunctions.includes(match[1])) {
                errors.push(`未知函数: ${match[1]}`);
            }
        }

        setSyntaxErrors(errors);
        return errors.length === 0;
    };

    // 模拟因子测试
    const handleTestFactor = async () => {
        if (!expression.trim()) {
            alert('请输入因子表达式');
            return;
        }

        if (!checkSyntax(expression)) {
            alert('表达式存在语法错误，请检查');
            return;
        }

        setIsTestingFactor(true);
        
        // 模拟API调用延迟
        setTimeout(() => {
            const mockResult = {
                ic: (Math.random() * 0.1 - 0.05).toFixed(4),
                icIR: (Math.random() * 2 + 0.5).toFixed(2),
                rank_ic: (Math.random() * 0.2 - 0.1).toFixed(4),
                rank_icIR: (Math.random() * 2 + 0.5).toFixed(2),
                turnover: (Math.random() * 0.3 + 0.1).toFixed(3),
                coverage: (Math.random() * 0.2 + 0.8).toFixed(3),
                validPeriods: Math.floor(Math.random() * 50 + 200),
                distribution: Array.from({length: 20}, (_, i) => ({
                    bin: (i * 0.1 - 1).toFixed(1),
                    count: Math.floor(Math.random() * 100 + 50)
                })),
                timeSeries: Array.from({length: 60}, (_, i) => ({
                    date: new Date(2023, 0, 1 + i * 5).toISOString().split('T')[0],
                    ic: (Math.random() * 0.2 - 0.1).toFixed(4),
                    coverage: (Math.random() * 0.1 + 0.85).toFixed(3)
                }))
            };
            setTestResult(mockResult);
            setIsTestingFactor(false);
        }, 2000);
    };

    // 保存因子
    const handleSaveFactor = () => {
        if (!factorName.trim()) {
            alert('请输入因子名称');
            return;
        }
        if (!expression.trim()) {
            alert('请输入因子表达式');
            return;
        }
        if (!checkSyntax(expression)) {
            alert('表达式存在语法错误，请检查');
            return;
        }

        const newFactor = {
            id: `factor_${Date.now()}`,
            name: factorName,
            expression: expression,
            description: description,
            createTime: new Date().toLocaleString(),
            status: 'active'
        };

        onSave(newFactor);
        alert('因子保存成功！');
        
        // 清空表单
        setFactorName('');
        setExpression('');
        setDescription('');
        setTestResult(null);
    };

    // 插入操作符或函数
    const insertOperator = (op) => {
        const textarea = editorRef.current;
        if (textarea) {
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const newExpr = expression.substring(0, start) + op + expression.substring(end);
            setExpression(newExpr);
            
            // 设置光标位置
            setTimeout(() => {
                textarea.focus();
                textarea.setSelectionRange(start + op.length, start + op.length);
            }, 0);
        }
    };

    // 加载模板
    const loadTemplate = (template) => {
        setExpression(template.expression);
        setFactorName(template.name);
        setDescription(template.description);
        setTestResult(null);
    };

    return (
        <div className="factor-editor">
            <div className="factor-editor-header">
                <h1>🧮 因子表达式编辑器</h1>
                <div className="editor-tabs">
                    <button 
                        className={`tab-btn ${activeTab === 'editor' ? 'active' : ''}`}
                        onClick={() => setActiveTab('editor')}
                    >
                        📝 编辑器
                    </button>
                    <button 
                        className={`tab-btn ${activeTab === 'library' ? 'active' : ''}`}
                        onClick={() => setActiveTab('library')}
                    >
                        📚 因子库
                    </button>
                    <button 
                        className={`tab-btn ${activeTab === 'templates' ? 'active' : ''}`}
                        onClick={() => setActiveTab('templates')}
                    >
                        📋 模板
                    </button>
                </div>
            </div>

            {activeTab === 'editor' && (
                <div className="editor-content">
                    <div className="editor-main">
                        <div className="editor-left">
                            {/* 因子信息 */}
                            <div className="factor-info">
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
                                        value={description}
                                        onChange={(e) => setDescription(e.target.value)}
                                        placeholder="请输入因子描述"
                                        className="form-textarea"
                                        rows="3"
                                    />
                                </div>
                            </div>

                            {/* 表达式编辑器 */}
                            <div className="expression-editor">
                                <h3>⚡ 表达式编辑</h3>
                                <div className="editor-wrapper">
                                    <textarea
                                        ref={editorRef}
                                        value={expression}
                                        onChange={(e) => {
                                            setExpression(e.target.value);
                                            checkSyntax(e.target.value);
                                        }}
                                        placeholder="请输入因子表达式，例如: ($close - Mean($close, 20)) / Std($close, 20)"
                                        className="expression-textarea"
                                        rows="8"
                                    />
                                    
                                    {/* 语法错误提示 */}
                                    {syntaxErrors.length > 0 && (
                                        <div className="syntax-errors">
                                            <h4>⚠️ 语法错误:</h4>
                                            {syntaxErrors.map((error, index) => (
                                                <div key={index} className="error-item">
                                                    {error}
                                                </div>
                                            ))}
                                        </div>
                                    )}

                                    {/* 操作按钮 */}
                                    <div className="editor-actions">
                                        <button 
                                            className="btn-primary"
                                            onClick={handleTestFactor}
                                            disabled={isTestingFactor || syntaxErrors.length > 0}
                                        >
                                            {isTestingFactor ? '🔄 测试中...' : '🔍 测试因子'}
                                        </button>
                                        <button 
                                            className="btn-success"
                                            onClick={handleSaveFactor}
                                            disabled={syntaxErrors.length > 0}
                                        >
                                            💾 保存因子
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="editor-right">
                            {/* 操作符面板 */}
                            <div className="operators-panel">
                                <h3>🛠️ 操作符库</h3>
                                {Object.entries(operators).map(([category, ops]) => (
                                    <div key={category} className="operator-category">
                                        <h4>{category}</h4>
                                        <div className="operator-grid">
                                            {ops.map(op => (
                                                <button
                                                    key={op}
                                                    className="operator-btn"
                                                    onClick={() => insertOperator(op)}
                                                    title={`插入 ${op}`}
                                                >
                                                    {op}
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>

                            {/* 语法说明 */}
                            <div className="syntax-help">
                                <h3>📖 语法说明</h3>
                                <div className="help-content">
                                    <div className="help-section">
                                        <h4>基础字段</h4>
                                        <p>$close, $open, $high, $low, $volume, $amount</p>
                                    </div>
                                    <div className="help-section">
                                        <h4>时序函数</h4>
                                        <p>Ref($close, 1) - 获取前一期数据</p>
                                        <p>Mean($close, 20) - 20期均值</p>
                                        <p>Std($close, 20) - 20期标准差</p>
                                    </div>
                                    <div className="help-section">
                                        <h4>横截面函数</h4>
                                        <p>Rank($close) - 横截面排名</p>
                                        <p>Zscore($close) - 横截面标准化</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* 测试结果 */}
                    {testResult && (
                        <div className="test-results">
                            <h3>📊 测试结果</h3>
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

                            {/* IC时序图 */}
                            <div className="ic-chart">
                                <h4>IC时序走势</h4>
                                <svg viewBox="0 0 800 200" className="ic-time-series">
                                    <defs>
                                        <pattern id="icGrid" width="40" height="20" patternUnits="userSpaceOnUse">
                                            <path d="M 40 0 L 0 0 0 20" fill="none" stroke="#f0f0f0" strokeWidth="1"/>
                                        </pattern>
                                    </defs>
                                    <rect width="800" height="200" fill="url(#icGrid)" />
                                    
                                    {/* 零线 */}
                                    <line x1="0" y1="100" x2="800" y2="100" stroke="#ddd" strokeWidth="1" strokeDasharray="5,5"/>
                                    
                                    {/* IC曲线 */}
                                    <polyline
                                        points={testResult.timeSeries.map((d, i) => 
                                            `${i * 800/testResult.timeSeries.length},${100 - parseFloat(d.ic) * 500}`
                                        ).join(' ')}
                                        fill="none"
                                        stroke="#1890ff"
                                        strokeWidth="2"
                                    />
                                    
                                    {/* Y轴标签 */}
                                    <text x="5" y="15" fill="#666" fontSize="10">0.1</text>
                                    <text x="5" y="105" fill="#666" fontSize="10">0.0</text>
                                    <text x="5" y="195" fill="#666" fontSize="10">-0.1</text>
                                </svg>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {activeTab === 'library' && (
                <div className="factor-library">
                    <h3>📚 已保存的因子</h3>
                    {savedFactors.length === 0 ? (
                        <div className="empty-state">
                            <div className="empty-icon">📭</div>
                            <div className="empty-text">暂无保存的因子</div>
                            <div className="empty-sub">开始创建你的第一个因子吧</div>
                        </div>
                    ) : (
                        <div className="factor-list">
                            {savedFactors.map(factor => (
                                <div key={factor.id} className="factor-item">
                                    <div className="factor-header">
                                        <h4>{factor.name}</h4>
                                        <div className="factor-status active">活跃</div>
                                    </div>
                                    <div className="factor-expression">{factor.expression}</div>
                                    <div className="factor-description">{factor.description}</div>
                                    <div className="factor-meta">
                                        <span>创建时间: {factor.createTime}</span>
                                    </div>
                                    <div className="factor-actions">
                                        <button 
                                            className="btn-secondary"
                                            onClick={() => {
                                                setExpression(factor.expression);
                                                setFactorName(factor.name);
                                                setDescription(factor.description);
                                                setActiveTab('editor');
                                            }}
                                        >
                                            编辑
                                        </button>
                                        <button className="btn-primary">测试</button>
                                        <button className="btn-danger">删除</button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}

            {activeTab === 'templates' && (
                <div className="factor-templates">
                    <h3>📋 因子模板</h3>
                    <div className="template-grid">
                        {factorTemplates.map((template, index) => (
                            <div key={index} className="template-card">
                                <div className="template-header">
                                    <h4>{template.name}</h4>
                                    <button 
                                        className="btn-primary"
                                        onClick={() => {
                                            loadTemplate(template);
                                            setActiveTab('editor');
                                        }}
                                    >
                                        使用模板
                                    </button>
                                </div>
                                <div className="template-expression">{template.expression}</div>
                                <div className="template-description">{template.description}</div>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};