// 布局组件
const { useState, useEffect, useCallback } = React;

const Layout = ({ children, currentPage, onNavigate }) => {
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    const [notifications, setNotifications] = useState([
        { id: 1, type: 'success', message: '模型训练完成', time: '2分钟前' },
        { id: 2, type: 'info', message: '数据集更新', time: '15分钟前' }
    ]);
    const [showNotifications, setShowNotifications] = useState(false);

    const menuItems = [
        { key: 'dashboard', label: '总览', icon: '🏠', desc: '系统概览和快速操作' },
        { key: 'data', label: '数据管理', icon: '💾', desc: 'Qlib数据集和数据源管理' },
        { key: 'factor', label: '因子研究', icon: '🧮', desc: '因子开发、编辑和分析' },
        { key: 'workflow', label: 'qlib量化研究工作流', icon: '⚙️', desc: '基于qlib的端到端量化投资研究流程' },
        { key: 'results', label: '结果分析', icon: '📊', desc: '模型和策略分析报告' },
    ];

    const systemStatus = {
        cpu: 65,
        memory: 78,
        disk: 45,
        network: 'online'
    };

    const clearNotification = (id) => {
        setNotifications(prev => prev.filter(n => n.id !== id));
    };

    return (
        <div className="layout">
            {/* 顶部导航栏 */}
            <header className="header">
                <div className="header-left">
                    <button 
                        className="sidebar-toggle"
                        onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                    >
                        ☰
                    </button>
                    <div className="logo">
                        <span className="logo-icon">📊</span>
                        <span className="logo-text">Qlib量化平台</span>
                    </div>
                </div>
                
                <div className="header-center">
                    <div className="breadcrumb">
                        <span className="breadcrumb-item">
                            {menuItems.find(item => item.key === currentPage)?.icon} {' '}
                            {menuItems.find(item => item.key === currentPage)?.label}
                        </span>
                    </div>
                </div>
                
                <div className="header-right">
                    <div className="system-indicators">
                        <div className="indicator">
                            <span className="indicator-label">CPU</span>
                            <div className="indicator-bar">
                                <div className="indicator-fill" style={{width: `${systemStatus.cpu}%`}}></div>
                            </div>
                            <span className="indicator-value">{systemStatus.cpu}%</span>
                        </div>
                        <div className="indicator">
                            <span className="indicator-label">内存</span>
                            <div className="indicator-bar">
                                <div className="indicator-fill" style={{width: `${systemStatus.memory}%`}}></div>
                            </div>
                            <span className="indicator-value">{systemStatus.memory}%</span>
                        </div>
                    </div>
                    
                    <div className="header-actions">
                        <div className="notification-bell" onClick={() => setShowNotifications(!showNotifications)}>
                            🔔
                            {notifications.length > 0 && (
                                <span className="notification-badge">{notifications.length}</span>
                            )}
                        </div>
                        <div className="user-menu">
                            <div className="user-avatar">👤</div>
                            <div className="user-name">管理员</div>
                        </div>
                    </div>
                </div>
            </header>

            <div className="layout-body">
                {/* 侧边栏 */}
                <aside className={`sidebar ${sidebarCollapsed ? 'collapsed' : ''}`}>
                    <nav className="sidebar-nav">
                        {menuItems.map(item => (
                            <div
                                key={item.key}
                                className={`nav-item ${currentPage === item.key ? 'active' : ''}`}
                                onClick={() => onNavigate(item.key)}
                                title={sidebarCollapsed ? `${item.label} - ${item.desc}` : ''}
                            >
                                <span className="nav-icon">{item.icon}</span>
                                {!sidebarCollapsed && (
                                    <div className="nav-content">
                                        <span className="nav-label">{item.label}</span>
                                        <span className="nav-desc">{item.desc}</span>
                                    </div>
                                )}
                            </div>
                        ))}
                    </nav>
                    
                    {!sidebarCollapsed && (
                        <div className="sidebar-footer">
                            <div className="system-info">
                                <h3>系统信息</h3>
                                <div className="info-item">
                                    <span>状态:</span>
                                    <span className="status-online">正常运行</span>
                                </div>
                                <div className="info-item">
                                    <span>版本:</span>
                                    <span>v1.0.0</span>
                                </div>
                                <div className="info-item">
                                    <span>运行时间:</span>
                                    <span>2天3小时</span>
                                </div>
                            </div>
                        </div>
                    )}
                </aside>

                {/* 主内容区 */}
                <main className="main-content">
                    {children}
                </main>
            </div>

            {/* 状态栏 */}
            <footer className="status-bar">
                <div className="status-left">
                    <span className="status-item">
                        🟢 系统正常
                    </span>
                    <span className="status-item">
                        📊 实时数据已连接
                    </span>
                    <span className="status-item">
                        ⏰ {new Date().toLocaleString()}
                    </span>
                </div>
                <div className="status-right">
                    <span className="status-item">
                        磁盘: {systemStatus.disk}%
                    </span>
                    <span className="status-item">
                        网络: {systemStatus.network === 'online' ? '🟢 在线' : '🔴 离线'}
                    </span>
                </div>
            </footer>

            {/* 通知面板 */}
            {showNotifications && (
                <div className="notification-panel">
                    <div className="notification-header">
                        <h3>通知</h3>
                        <button onClick={() => setShowNotifications(false)}>×</button>
                    </div>
                    <div className="notification-list">
                        {notifications.length === 0 ? (
                            <div className="no-notifications">暂无通知</div>
                        ) : (
                            notifications.map(notification => (
                                <div key={notification.id} className={`notification-item ${notification.type}`}>
                                    <div className="notification-content">
                                        <div className="notification-message">{notification.message}</div>
                                        <div className="notification-time">{notification.time}</div>
                                    </div>
                                    <button 
                                        className="notification-close"
                                        onClick={() => clearNotification(notification.id)}
                                    >
                                        ×
                                    </button>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            )}
        </div>
    );
};