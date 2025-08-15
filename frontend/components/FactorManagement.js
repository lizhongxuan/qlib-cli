// 因子管理组件 - 整合因子工程、编辑器和分析功能
const { useState, useEffect } = React;

const FactorManagement = ({ 
    onSave = () => {}, 
    onTestFactor = () => {}, 
    savedFactors = [],
    onNavigate = () => {} 
}) => {
    const [activeTab, setActiveTab] = useState('workshop');
    const [selectedFactor, setSelectedFactor] = useState(null);

    // 标签页配置
    const tabs = [
        { 
            key: 'workshop', 
            label: '因子工程', 
            icon: '🏭', 
            desc: '内置因子库和AI助手' 
        },
        { 
            key: 'editor', 
            label: '因子编辑器', 
            icon: '✏️', 
            desc: '高级表达式编辑和测试' 
        },
        { 
            key: 'analysis', 
            label: '因子分析', 
            icon: '📊', 
            desc: '因子效果分析和可视化' 
        },
        { 
            key: 'library', 
            label: '因子库', 
            icon: '📚', 
            desc: '已保存的因子管理' 
        }
    ];

    // 处理标签页切换
    const handleTabChange = (tabKey) => {
        setActiveTab(tabKey);
    };

    // 处理因子选择（用于分析）
    const handleFactorSelect = (factor) => {
        setSelectedFactor(factor);
        setActiveTab('analysis');
    };

    // 从工程或编辑器切换到分析
    const navigateToAnalysis = (factor) => {
        setSelectedFactor(factor);
        setActiveTab('analysis');
    };

    return (
        <div className="factor-management">
            <div className="factor-management-header">
                <h1>🧮 因子管理中心</h1>
                <div className="header-subtitle">
                    集成因子工程、编辑器和分析功能的统一工作平台
                </div>
            </div>

            {/* 主导航标签页 */}
            <div className="management-tabs">
                {tabs.map(tab => (
                    <button
                        key={tab.key}
                        className={`management-tab ${activeTab === tab.key ? 'active' : ''}`}
                        onClick={() => handleTabChange(tab.key)}
                    >
                        <span className="tab-icon">{tab.icon}</span>
                        <div className="tab-content">
                            <div className="tab-label">{tab.label}</div>
                            <div className="tab-desc">{tab.desc}</div>
                        </div>
                    </button>
                ))}
            </div>

            {/* 内容区域 */}
            <div className="management-content">
                {activeTab === 'workshop' && (
                    <FactorWorkshop 
                        onSave={onSave}
                        onNavigate={(page) => {
                            if (page === 'factor') {
                                setActiveTab('editor');
                            } else if (page === 'analysis') {
                                setActiveTab('analysis');
                            } else {
                                onNavigate(page);
                            }
                        }}
                        savedFactors={savedFactors}
                    />
                )}

                {activeTab === 'editor' && (
                    <FactorEditor 
                        onSave={onSave}
                        onTestFactor={onTestFactor}
                        savedFactors={savedFactors}
                        onNavigate={(page) => {
                            if (page === 'analysis') {
                                setActiveTab('analysis');
                            } else {
                                onNavigate(page);
                            }
                        }}
                        onAnalyzeFactor={navigateToAnalysis}
                    />
                )}

                {activeTab === 'analysis' && (
                    <FactorAnalysis 
                        factorData={selectedFactor}
                        onNavigate={onNavigate}
                        onBackToEditor={() => setActiveTab('editor')}
                        onBackToWorkshop={() => setActiveTab('workshop')}
                    />
                )}

                {activeTab === 'library' && (
                    <div className="factor-library-tab">
                        <div className="library-header">
                            <h2>📚 因子库管理</h2>
                            <div className="library-actions">
                                <button 
                                    className="btn-primary"
                                    onClick={() => setActiveTab('workshop')}
                                >
                                    🏭 创建新因子
                                </button>
                                <button 
                                    className="btn-secondary"
                                    onClick={() => setActiveTab('editor')}
                                >
                                    ✏️ 高级编辑
                                </button>
                            </div>
                        </div>

                        {savedFactors.length === 0 ? (
                            <div className="empty-library">
                                <div className="empty-icon">📭</div>
                                <div className="empty-title">暂无保存的因子</div>
                                <div className="empty-desc">
                                    开始使用因子工程或编辑器创建你的第一个因子
                                </div>
                                <div className="empty-actions">
                                    <button 
                                        className="btn-primary"
                                        onClick={() => setActiveTab('workshop')}
                                    >
                                        🏭 使用因子工程
                                    </button>
                                    <button 
                                        className="btn-secondary"
                                        onClick={() => setActiveTab('editor')}
                                    >
                                        ✏️ 高级编辑器
                                    </button>
                                </div>
                            </div>
                        ) : (
                            <div className="factor-library-grid">
                                {savedFactors.map(factor => (
                                    <div key={factor.id} className="factor-library-card">
                                        <div className="factor-card-header">
                                            <h3>{factor.name}</h3>
                                            <div className={`factor-status ${factor.status}`}>
                                                {factor.status === 'active' ? '活跃' : '停用'}
                                            </div>
                                        </div>
                                        
                                        <div className="factor-expression-preview">
                                            {factor.expression}
                                        </div>
                                        
                                        <div className="factor-description">
                                            {factor.description || '暂无描述'}
                                        </div>
                                        
                                        <div className="factor-meta">
                                            <div className="meta-item">
                                                <span className="meta-label">创建时间:</span>
                                                <span className="meta-value">{factor.createTime}</span>
                                            </div>
                                        </div>
                                        
                                        <div className="factor-card-actions">
                                            <button 
                                                className="btn-secondary btn-sm"
                                                onClick={() => {
                                                    // 编辑因子
                                                    setActiveTab('editor');
                                                }}
                                            >
                                                编辑
                                            </button>
                                            <button 
                                                className="btn-primary btn-sm"
                                                onClick={() => handleFactorSelect(factor)}
                                            >
                                                分析
                                            </button>
                                            <button 
                                                className="btn-success btn-sm"
                                                onClick={() => onNavigate('model')}
                                            >
                                                训练模型
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}
            </div>

            {/* 快速导航面板 */}
            <div className="quick-navigation">
                <h3>🚀 快速操作</h3>
                <div className="quick-nav-grid">
                    <button 
                        className="quick-nav-item"
                        onClick={() => setActiveTab('workshop')}
                    >
                        <div className="nav-icon">🏭</div>
                        <div className="nav-label">创建因子</div>
                        <div className="nav-desc">使用内置库和AI助手</div>
                    </button>
                    <button 
                        className="quick-nav-item"
                        onClick={() => setActiveTab('editor')}
                    >
                        <div className="nav-icon">✏️</div>
                        <div className="nav-label">编辑表达式</div>
                        <div className="nav-desc">高级编辑和测试</div>
                    </button>
                    <button 
                        className="quick-nav-item"
                        onClick={() => setActiveTab('analysis')}
                    >
                        <div className="nav-icon">📊</div>
                        <div className="nav-label">分析因子</div>
                        <div className="nav-desc">效果分析和可视化</div>
                    </button>
                    <button 
                        className="quick-nav-item"
                        onClick={() => onNavigate('model')}
                    >
                        <div className="nav-icon">🤖</div>
                        <div className="nav-label">训练模型</div>
                        <div className="nav-desc">使用因子训练模型</div>
                    </button>
                </div>
            </div>
        </div>
    );
};