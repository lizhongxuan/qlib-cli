// 数据管理组件
const { useState, useEffect, useCallback } = React;

const DataManagement = ({ datasets, onAddDataset, onAddTask }) => {
    const [showForm, setShowForm] = useState(false);
    const [showSourceForm, setShowSourceForm] = useState(false);
    const [activeTab, setActiveTab] = useState('datasets');
    const [dataSources, setDataSources] = useState([
        { id: 'yahoo', name: 'Yahoo Finance', type: 'API', status: '在线', lastUpdate: '实时', description: '免费股票数据' },
        { id: 'local', name: '本地CSV文件', type: 'File', status: '离线', lastUpdate: '手动', description: '上传的本地数据' },
        { id: 'qlib_cn', name: 'Qlib中国数据', type: 'Built-in', status: '在线', lastUpdate: '每日', description: '内置A股数据' }
    ]);
    
    const [formData, setFormData] = useState({
        name: '',
        dataset: 'csi300',
        startDate: '2020-01-01',
        endDate: '2023-12-31',
        features: []
    });

    const [sourceFormData, setSourceFormData] = useState({
        name: '',
        type: 'api',
        url: '',
        description: ''
    });

    const availableFeatures = [
        { value: 'open', label: '开盘价' },
        { value: 'close', label: '收盘价' },
        { value: 'high', label: '最高价' },
        { value: 'low', label: '最低价' },
        { value: 'volume', label: '成交量' },
        { value: 'amount', label: '成交额' },
        { value: 'rsi', label: 'RSI指标' },
        { value: 'macd', label: 'MACD指标' },
        { value: 'ma5', label: '5日均线' },
        { value: 'ma20', label: '20日均线' }
    ];

    const handleSubmit = () => {
        const datasetId = onAddDataset({
            name: formData.name || `${formData.dataset}-数据集`,
            dateRange: `${formData.startDate} 至 ${formData.endDate}`,
            features: formData.features.length
        });
        
        onAddTask({
            name: `准备数据集: ${formData.name || formData.dataset}`,
            type: 'data_prepare'
        });
        
        setShowForm(false);
        setFormData({
            name: '',
            dataset: 'csi300',
            startDate: '2020-01-01',
            endDate: '2023-12-31',
            features: []
        });
    };

    const toggleFeature = (feature) => {
        setFormData(prev => ({
            ...prev,
            features: prev.features.includes(feature)
                ? prev.features.filter(f => f !== feature)
                : [...prev.features, feature]
        }));
    };

    const addDataSource = () => {
        const newSource = {
            id: `source_${Date.now()}`,
            ...sourceFormData,
            status: '配置中',
            lastUpdate: '未更新'
        };
        setDataSources(prev => [...prev, newSource]);
        setShowSourceForm(false);
        setSourceFormData({ name: '', type: 'api', url: '', description: '' });
    };

    const testConnection = (sourceId) => {
        setDataSources(prev => prev.map(source => 
            source.id === sourceId 
                ? { ...source, status: '测试中...' }
                : source
        ));
        
        // 模拟连接测试
        setTimeout(() => {
            setDataSources(prev => prev.map(source => 
                source.id === sourceId 
                    ? { ...source, status: Math.random() > 0.2 ? '在线' : '连接失败' }
                    : source
            ));
        }, 2000);
    };

    return (
        <div className="data-management">
            <div className="page-header">
                <h1>💾 数据管理</h1>
                <div className="header-actions">
                    {activeTab === 'datasets' ? (
                        <button className="btn-primary" onClick={() => setShowForm(true)}>
                            + 创建数据集
                        </button>
                    ) : (
                        <button className="btn-primary" onClick={() => setShowSourceForm(true)}>
                            + 添加数据源
                        </button>
                    )}
                </div>
            </div>

            {/* 标签页导航 */}
            <div className="tab-navigation">
                <button 
                    className={`tab-btn ${activeTab === 'datasets' ? 'active' : ''}`}
                    onClick={() => setActiveTab('datasets')}
                >
                    📊 数据集管理
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'sources' ? 'active' : ''}`}
                    onClick={() => setActiveTab('sources')}
                >
                    🔗 数据源管理
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'explorer' ? 'active' : ''}`}
                    onClick={() => setActiveTab('explorer')}
                >
                    🔍 数据探索
                </button>
            </div>

            {/* 标签页内容 */}
            <div className="tab-content">
                {activeTab === 'datasets' && (
                    <div className="datasets-content">
                        <div className="content-header">
                            <div className="search-filter">
                                <input type="text" placeholder="🔍 搜索数据集..." className="search-input" />
                                <select className="filter-select">
                                    <option value="all">全部状态</option>
                                    <option value="ready">可用</option>
                                    <option value="preparing">准备中</option>
                                </select>
                            </div>
                            <div className="view-toggle">
                                <button className="view-btn active">📱</button>
                                <button className="view-btn">📋</button>
                            </div>
                        </div>
                        
                        <div className="dataset-grid">
                            {datasets.map(dataset => (
                                <div key={dataset.id} className="dataset-card">
                                    <div className="dataset-status">
                                        <span className={`status-badge ${dataset.status}`}>
                                            {dataset.status === 'ready' ? '✅ 可用' : '⏳ 准备中'}
                                        </span>
                                    </div>
                                    <h3>{dataset.name}</h3>
                                    <div className="dataset-info">
                                        <div className="info-item">
                                            <span className="info-label">📊 样本数:</span>
                                            <span className="info-value">{dataset.samples?.toLocaleString()}</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">🔧 特征数:</span>
                                            <span className="info-value">{dataset.features}</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">📅 时间范围:</span>
                                            <span className="info-value">{dataset.dateRange}</span>
                                        </div>
                                    </div>
                                    <div className="dataset-actions">
                                        <button className="btn-text">👁️ 查看</button>
                                        <button className="btn-text">📈 探索</button>
                                        <button className="btn-text">📋 导出</button>
                                        <button className="btn-text danger">🗑️ 删除</button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {activeTab === 'sources' && (
                    <div className="sources-content">
                        <div className="sources-stats">
                            <div className="stat-item">
                                <span className="stat-number">{dataSources.filter(s => s.status === '在线').length}</span>
                                <span className="stat-label">在线数据源</span>
                            </div>
                            <div className="stat-item">
                                <span className="stat-number">{dataSources.length}</span>
                                <span className="stat-label">总数据源</span>
                            </div>
                        </div>
                        
                        <div className="sources-table">
                            <table className="data-table">
                                <thead>
                                    <tr>
                                        <th>数据源名称</th>
                                        <th>类型</th>
                                        <th>状态</th>
                                        <th>最后更新</th>
                                        <th>说明</th>
                                        <th>操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {dataSources.map(source => (
                                        <tr key={source.id}>
                                            <td>
                                                <div className="source-name">
                                                    <div className="name">{source.name}</div>
                                                </div>
                                            </td>
                                            <td>
                                                <span className={`type-badge ${source.type.toLowerCase()}`}>
                                                    {source.type}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`status-badge ${
                                                    source.status === '在线' ? 'online' : 
                                                    source.status === '离线' ? 'offline' : 
                                                    source.status === '连接失败' ? 'error' : 'pending'
                                                }`}>
                                                    {source.status}
                                                </span>
                                            </td>
                                            <td>{source.lastUpdate}</td>
                                            <td>{source.description}</td>
                                            <td>
                                                <div className="action-buttons">
                                                    <button 
                                                        className="btn-text"
                                                        onClick={() => testConnection(source.id)}
                                                        disabled={source.status === '测试中...'}
                                                    >
                                                        🔍 测试
                                                    </button>
                                                    <button className="btn-text">⚙️ 配置</button>
                                                    <button className="btn-text danger">🗑️ 删除</button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {activeTab === 'explorer' && (
                    <div className="explorer-content">
                        <div className="explorer-placeholder">
                            <div className="placeholder-icon">🔍</div>
                            <h3>数据探索器</h3>
                            <p>选择数据集进行可视化分析和统计探索</p>
                            <div className="explorer-features">
                                <div className="feature-item">📊 数据分布图</div>
                                <div className="feature-item">📈 相关性分析</div>
                                <div className="feature-item">📋 缺失值检查</div>
                                <div className="feature-item">🎯 异常值检测</div>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* 创建数据集表单 */}
            {showForm && (
                <div className="modal-overlay" onClick={() => setShowForm(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>创建新数据集</h2>
                            <button className="close-btn" onClick={() => setShowForm(false)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="form-group">
                                <label>数据集名称</label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={e => setFormData({...formData, name: e.target.value})}
                                    placeholder="输入数据集名称"
                                />
                            </div>

                            <div className="form-group">
                                <label>股票池</label>
                                <select
                                    value={formData.dataset}
                                    onChange={e => setFormData({...formData, dataset: e.target.value})}
                                >
                                    <option value="csi300">沪深300</option>
                                    <option value="csi500">中证500</option>
                                    <option value="csi1000">中证1000</option>
                                    <option value="all">全市场</option>
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

                            <div className="form-group">
                                <label>选择特征</label>
                                <div className="feature-grid">
                                    {availableFeatures.map(feature => (
                                        <div
                                            key={feature.value}
                                            className={`feature-item ${formData.features.includes(feature.value) ? 'selected' : ''}`}
                                            onClick={() => toggleFeature(feature.value)}
                                        >
                                            {feature.label}
                                        </div>
                                    ))}
                                </div>
                            </div>

                            <div className="form-group">
                                <label>标签定义</label>
                                <input
                                    type="text"
                                    placeholder="Ref($close, -1) / $close - 1"
                                    defaultValue="Ref($close, -1) / $close - 1"
                                />
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setShowForm(false)}>
                                取消
                            </button>
                            <button className="btn-primary" onClick={handleSubmit}>
                                创建数据集
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 添加数据源表单 */}
            {showSourceForm && (
                <div className="modal-overlay" onClick={() => setShowSourceForm(false)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>🔗 添加新数据源</h2>
                            <button className="close-btn" onClick={() => setShowSourceForm(false)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="form-group">
                                <label>数据源名称</label>
                                <input
                                    type="text"
                                    value={sourceFormData.name}
                                    onChange={e => setSourceFormData({...sourceFormData, name: e.target.value})}
                                    placeholder="输入数据源名称"
                                />
                            </div>

                            <div className="form-group">
                                <label>数据源类型</label>
                                <select
                                    value={sourceFormData.type}
                                    onChange={e => setSourceFormData({...sourceFormData, type: e.target.value})}
                                >
                                    <option value="api">API接口</option>
                                    <option value="database">数据库</option>
                                    <option value="file">文件上传</option>
                                    <option value="ftp">FTP服务器</option>
                                </select>
                            </div>

                            {sourceFormData.type === 'api' && (
                                <div className="form-group">
                                    <label>API地址</label>
                                    <input
                                        type="url"
                                        value={sourceFormData.url}
                                        onChange={e => setSourceFormData({...sourceFormData, url: e.target.value})}
                                        placeholder="https://api.example.com/data"
                                    />
                                </div>
                            )}

                            {sourceFormData.type === 'database' && (
                                <div className="form-group">
                                    <label>连接字符串</label>
                                    <input
                                        type="text"
                                        value={sourceFormData.url}
                                        onChange={e => setSourceFormData({...sourceFormData, url: e.target.value})}
                                        placeholder="mysql://user:password@host:port/database"
                                    />
                                </div>
                            )}

                            {sourceFormData.type === 'file' && (
                                <div className="form-group">
                                    <label>文件上传</label>
                                    <div className="file-upload-area">
                                        <div className="upload-icon">📁</div>
                                        <div className="upload-text">拖拽文件到此处或点击上传</div>
                                        <input type="file" accept=".csv,.xlsx,.json" style={{display: 'none'}} />
                                    </div>
                                </div>
                            )}

                            <div className="form-group">
                                <label>描述</label>
                                <textarea
                                    value={sourceFormData.description}
                                    onChange={e => setSourceFormData({...sourceFormData, description: e.target.value})}
                                    placeholder="描述数据源的内容和用途"
                                    rows="3"
                                />
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setShowSourceForm(false)}>
                                取消
                            </button>
                            <button className="btn-primary" onClick={addDataSource}>
                                添加数据源
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};