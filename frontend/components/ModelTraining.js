// 模型训练组件
const { useState, useEffect, useCallback } = React;

const ModelTraining = ({ datasets, models, onAddModel, onAddTask }) => {
    const [showForm, setShowForm] = useState(false);
    const [selectedModel, setSelectedModel] = useState(null);
    const [activeTab, setActiveTab] = useState('models');
    const [trainingStatus, setTrainingStatus] = useState(null);
    const [formData, setFormData] = useState({
        name: '',
        type: 'xgboost',
        dataset: '',
        learningRate: 0.01,
        nEstimators: 100,
        maxDepth: 6,
        validationSplit: 0.2,
        earlyStop: true
    });

    // 模拟训练指标数据
    const [trainingMetrics, setTrainingMetrics] = useState([]);
    
    // 实时训练任务
    const [activeTasks, setActiveTasks] = useState([]);
    
    const handleTrain = () => {
        const taskId = `task_${Date.now()}`;
        const newTask = {
            id: taskId,
            name: formData.name || `${formData.type}-模型`,
            type: formData.type,
            dataset: formData.dataset,
            status: 'running',
            progress: 0,
            startTime: new Date().toLocaleString(),
            epochs: formData.nEstimators,
            currentEpoch: 0,
            metrics: {
                loss: 0,
                ic: 0,
                sharpe: 0
            }
        };
        
        setActiveTasks(prev => [...prev, newTask]);
        setTrainingStatus(taskId);
        setActiveTab('training');
        
        const modelId = onAddModel({
            name: formData.name || `${formData.type}-模型`,
            type: formData.type,
            dataset: formData.dataset
        });
        
        onAddTask({
            name: `训练模型: ${formData.name || formData.type}`,
            type: 'model_training'
        });
        
        // 模拟实时训练过程
        simulateTraining(taskId);
        
        setShowForm(false);
    };

    const simulateTraining = (taskId) => {
        let epoch = 0;
        let metrics = [];
        
        const interval = setInterval(() => {
            epoch++;
            const loss = Math.max(0.5 - epoch * 0.04 + Math.random() * 0.02, 0.1);
            const ic = Math.min(0.01 + epoch * 0.004 + Math.random() * 0.01, 0.06);
            const sharpe = Math.min(0.5 + epoch * 0.08 + Math.random() * 0.05, 1.8);
            
            metrics.push({ epoch, loss, ic, sharpe });
            setTrainingMetrics([...metrics]);
            
            setActiveTasks(prev => prev.map(task => 
                task.id === taskId ? {
                    ...task,
                    progress: (epoch / task.epochs) * 100,
                    currentEpoch: epoch,
                    metrics: { loss, ic, sharpe }
                } : task
            ));
            
            if (epoch >= 10) {
                setActiveTasks(prev => prev.map(task => 
                    task.id === taskId ? { ...task, status: 'completed' } : task
                ));
                clearInterval(interval);
            }
        }, 1000);
    };

    const stopTraining = (taskId) => {
        setActiveTasks(prev => prev.map(task => 
            task.id === taskId ? { ...task, status: 'stopped' } : task
        ));
    };

    const modelTypes = [
        { value: 'xgboost', label: 'XGBoost', desc: '梯度提升树模型' },
        { value: 'lightgbm', label: 'LightGBM', desc: '轻量级梯度提升' },
        { value: 'mlp', label: 'MLP', desc: '多层感知机' },
        { value: 'lstm', label: 'LSTM', desc: '长短期记忆网络' },
        { value: 'transformer', label: 'Transformer', desc: '注意力机制模型' }
    ];

    return (
        <div className="model-training">
            <div className="page-header">
                <h1>🤖 模型训练</h1>
                <div className="header-actions">
                    <button className="btn-secondary" onClick={() => setActiveTab('hyperparameters')}>
                        ⚙️ 超参优化
                    </button>
                    <button className="btn-primary" onClick={() => setShowForm(true)}>
                        + 训练新模型
                    </button>
                </div>
            </div>

            {/* 标签页导航 */}
            <div className="tab-navigation">
                <button 
                    className={`tab-btn ${activeTab === 'models' ? 'active' : ''}`}
                    onClick={() => setActiveTab('models')}
                >
                    📚 模型仓库
                    <span className="tab-badge">{models.length}</span>
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'training' ? 'active' : ''}`}
                    onClick={() => setActiveTab('training')}
                >
                    ⏳ 训练监控
                    <span className="tab-badge">{activeTasks.filter(t => t.status === 'running').length}</span>
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'evaluation' ? 'active' : ''}`}
                    onClick={() => setActiveTab('evaluation')}
                >
                    📊 模型对比
                </button>
                <button 
                    className={`tab-btn ${activeTab === 'deployment' ? 'active' : ''}`}
                    onClick={() => setActiveTab('deployment')}
                >
                    🚀 模型部署
                </button>
            </div>

            {/* 标签页内容 */}
            <div className="tab-content">
                {activeTab === 'models' && (
                    <div className="models-content">
                        <div className="models-stats">
                            <div className="stat-card">
                                <div className="stat-number">{models.length}</div>
                                <div className="stat-label">总模型数</div>
                            </div>
                            <div className="stat-card">
                                <div className="stat-number">{models.filter(m => m.status === 'trained').length}</div>
                                <div className="stat-label">已完成</div>
                            </div>
                            <div className="stat-card">
                                <div className="stat-number">{models.length > 0 ? Math.max(...models.map(m => parseFloat(m.ic) || 0)).toFixed(3) : '0.000'}</div>
                                <div className="stat-label">最佳IC</div>
                            </div>
                        </div>
                        
                        <div className="model-list">
                            <table className="data-table">
                                <thead>
                                    <tr>
                                        <th>模型名称</th>
                                        <th>类型</th>
                                        <th>IC</th>
                                        <th>夏普比率</th>
                                        <th>状态</th>
                                        <th>训练时间</th>
                                        <th>操作</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {models.map(model => (
                                        <tr key={model.id}>
                                            <td>
                                                <div className="model-info">
                                                    <div className="model-name">{model.name}</div>
                                                </div>
                                            </td>
                                            <td>
                                                <span className="model-type-badge">{model.type}</span>
                                            </td>
                                            <td className="metric-value">
                                                <span className={parseFloat(model.ic) > 0.04 ? 'good' : 'normal'}>
                                                    {model.ic}
                                                </span>
                                            </td>
                                            <td className="metric-value">
                                                <span className={parseFloat(model.sharpe) > 1.5 ? 'good' : 'normal'}>
                                                    {model.sharpe}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`status-badge ${model.status}`}>
                                                    {model.status === 'trained' ? '✅ 已完成' : '⏳ 训练中'}
                                                </span>
                                            </td>
                                            <td>{model.trainTime}</td>
                                            <td>
                                                <div className="action-buttons">
                                                    <button 
                                                        className="btn-text"
                                                        onClick={() => setSelectedModel(model)}
                                                    >
                                                        👁️ 详情
                                                    </button>
                                                    <button className="btn-text">📊 对比</button>
                                                    <button className="btn-text">🚀 部署</button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}

                {activeTab === 'training' && (
                    <div className="training-content">
                        {activeTasks.length === 0 ? (
                            <div className="empty-state">
                                <div className="empty-icon">⏳</div>
                                <div className="empty-text">暂无训练任务</div>
                                <div className="empty-sub">开始训练新模型以监控进度</div>
                            </div>
                        ) : (
                            activeTasks.map(task => (
                                <div key={task.id} className="training-task">
                                    <div className="task-header">
                                        <div className="task-info">
                                            <h3>{task.name}</h3>
                                            <div className="task-meta">
                                                {task.type} • {task.startTime} • Epoch {task.currentEpoch}/{task.epochs}
                                            </div>
                                        </div>
                                        <div className="task-controls">
                                            <div className={`task-status ${task.status}`}>
                                                {task.status === 'running' ? '🔄 运行中' : 
                                                 task.status === 'completed' ? '✅ 已完成' : '⏹️ 已停止'}
                                            </div>
                                            {task.status === 'running' && (
                                                <button 
                                                    className="btn-secondary"
                                                    onClick={() => stopTraining(task.id)}
                                                >
                                                    ⏹️ 停止
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                    
                                    <div className="task-progress">
                                        <div className="progress-bar-container">
                                            <div className="progress-bar" style={{width: `${task.progress}%`}}></div>
                                        </div>
                                        <div className="progress-text">{Math.floor(task.progress)}%</div>
                                    </div>
                                    
                                    <div className="task-metrics">
                                        <div className="metric-item">
                                            <span className="metric-label">Loss:</span>
                                            <span className="metric-value">{task.metrics.loss.toFixed(4)}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">IC:</span>
                                            <span className="metric-value">{task.metrics.ic.toFixed(4)}</span>
                                        </div>
                                        <div className="metric-item">
                                            <span className="metric-label">Sharpe:</span>
                                            <span className="metric-value">{task.metrics.sharpe.toFixed(4)}</span>
                                        </div>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                )}

                {activeTab === 'evaluation' && (
                    <div className="evaluation-content">
                        <div className="evaluation-placeholder">
                            <div className="placeholder-icon">📊</div>
                            <h3>模型对比分析</h3>
                            <p>选择多个模型进行性能对比和分析</p>
                            <div className="comparison-features">
                                <div className="feature-item">📈 性能指标对比</div>
                                <div className="feature-item">⏱️ 训练时间分析</div>
                                <div className="feature-item">🎯 预测准确性</div>
                                <div className="feature-item">📋 详细报告导出</div>
                            </div>
                        </div>
                    </div>
                )}

                {activeTab === 'deployment' && (
                    <div className="deployment-content">
                        <div className="deployment-placeholder">
                            <div className="placeholder-icon">🚀</div>
                            <h3>模型部署</h3>
                            <p>将训练好的模型部署到生产环境</p>
                            <div className="deployment-options">
                                <div className="deploy-option">
                                    <div className="option-icon">☁️</div>
                                    <div className="option-name">云端部署</div>
                                    <div className="option-desc">部署到云服务器</div>
                                </div>
                                <div className="deploy-option">
                                    <div className="option-icon">🖥️</div>
                                    <div className="option-name">本地部署</div>
                                    <div className="option-desc">部署到本地服务器</div>
                                </div>
                                <div className="deploy-option">
                                    <div className="option-icon">🔗</div>
                                    <div className="option-name">API服务</div>
                                    <div className="option-desc">提供REST API接口</div>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>


            {/* 训练表单 */}
            {showForm && (
                <div className="modal-overlay" onClick={() => setShowForm(false)}>
                    <div className="modal large" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>配置并训练模型</h2>
                            <button className="close-btn" onClick={() => setShowForm(false)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="form-group">
                                <label>模型名称</label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={e => setFormData({...formData, name: e.target.value})}
                                    placeholder="输入模型名称"
                                />
                            </div>

                            <div className="form-group">
                                <label>选择模型类型</label>
                                <div className="model-type-grid">
                                    {modelTypes.map(type => (
                                        <div
                                            key={type.value}
                                            className={`model-type-card ${formData.type === type.value ? 'selected' : ''}`}
                                            onClick={() => setFormData({...formData, type: type.value})}
                                        >
                                            <div className="model-type-name">{type.label}</div>
                                            <div className="model-type-desc">{type.desc}</div>
                                        </div>
                                    ))}
                                </div>
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
                                            {dataset.name} ({dataset.samples} 样本)
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div className="form-group">
                                <label>模型参数</label>
                                <div className="params-grid">
                                    <div className="param-item">
                                        <label>学习率</label>
                                        <input
                                            type="number"
                                            step="0.001"
                                            min="0.001"
                                            max="1"
                                            value={formData.learningRate}
                                            onChange={e => setFormData({...formData, learningRate: e.target.value})}
                                        />
                                        <span className="param-hint">建议: 0.01-0.1</span>
                                    </div>
                                    <div className="param-item">
                                        <label>迭代次数</label>
                                        <input
                                            type="number"
                                            min="10"
                                            max="1000"
                                            value={formData.nEstimators}
                                            onChange={e => setFormData({...formData, nEstimators: e.target.value})}
                                        />
                                        <span className="param-hint">建议: 100-500</span>
                                    </div>
                                    <div className="param-item">
                                        <label>树深度</label>
                                        <input
                                            type="number"
                                            min="1"
                                            max="20"
                                            value={formData.maxDepth}
                                            onChange={e => setFormData({...formData, maxDepth: e.target.value})}
                                        />
                                        <span className="param-hint">建议: 6-10</span>
                                    </div>
                                </div>
                            </div>

                            <div className="form-group">
                                <label>训练配置</label>
                                <div className="training-config">
                                    <div className="config-item">
                                        <label>验证集比例</label>
                                        <input
                                            type="range"
                                            min="0.1"
                                            max="0.5"
                                            step="0.05"
                                            value={formData.validationSplit}
                                            onChange={e => setFormData({...formData, validationSplit: e.target.value})}
                                        />
                                        <span className="range-value">{(formData.validationSplit * 100).toFixed(0)}%</span>
                                    </div>
                                    
                                    <div className="config-item">
                                        <label>
                                            <input
                                                type="checkbox"
                                                checked={formData.earlyStop}
                                                onChange={e => setFormData({...formData, earlyStop: e.target.checked})}
                                            />
                                            启用早停机制
                                        </label>
                                        <span className="config-hint">防止过拟合</span>
                                    </div>
                                </div>
                            </div>

                            <div className="form-group">
                                <label>预计训练时间</label>
                                <div className="time-estimate">
                                    <div className="estimate-item">
                                        <span className="estimate-label">预计时长:</span>
                                        <span className="estimate-value">~{Math.ceil(formData.nEstimators / 10)} 分钟</span>
                                    </div>
                                    <div className="estimate-item">
                                        <span className="estimate-label">资源需求:</span>
                                        <span className="estimate-value">中等</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setShowForm(false)}>
                                取消
                            </button>
                            <button className="btn-primary" onClick={handleTrain}>
                                开始训练
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 模型详情 */}
            {selectedModel && (
                <div className="modal-overlay" onClick={() => setSelectedModel(null)}>
                    <div className="modal" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <h2>模型详情: {selectedModel.name}</h2>
                            <button className="close-btn" onClick={() => setSelectedModel(null)}>×</button>
                        </div>
                        <div className="modal-body">
                            <div className="detail-grid">
                                <div className="detail-item">
                                    <label>模型类型</label>
                                    <span>{selectedModel.type}</span>
                                </div>
                                <div className="detail-item">
                                    <label>IC值</label>
                                    <span>{selectedModel.ic}</span>
                                </div>
                                <div className="detail-item">
                                    <label>夏普比率</label>
                                    <span>{selectedModel.sharpe}</span>
                                </div>
                                <div className="detail-item">
                                    <label>训练时间</label>
                                    <span>{selectedModel.trainTime}</span>
                                </div>
                            </div>
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setSelectedModel(null)}>
                                关闭
                            </button>
                            <button className="btn-primary">
                                部署模型
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};