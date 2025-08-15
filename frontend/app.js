// 主应用入口

// 主应用组件
const App = () => {
    const [currentPage, setCurrentPage] = useState('dashboard');
    const [tasks, setTasks] = useState([]);
    const [datasets, setDatasets] = useState([
        { id: 'ds1', name: '沪深300日线数据', status: 'ready', samples: 50000, features: 20, dateRange: '2020-01-01 至 2023-12-31' },
        { id: 'ds2', name: '中证500日线数据', status: 'ready', samples: 45000, features: 18, dateRange: '2020-01-01 至 2023-12-31' }
    ]);
    const [models, setModels] = useState([
        { id: 'model1', name: 'XGBoost-v1.0', type: 'XGBoost', ic: 0.045, sharpe: 1.23, status: 'trained', trainTime: '2024-01-10' },
        { id: 'model2', name: 'LightGBM-v2.0', type: 'LightGBM', ic: 0.052, sharpe: 1.35, status: 'trained', trainTime: '2024-01-12' }
    ]);
    const [savedFactors, setSavedFactors] = useState([
        { id: 'factor1', name: '动量因子V1', expression: 'Rank($close / Ref($close, 20) - 1)', description: '20日价格动量因子', createTime: '2024-01-15 10:30:00', status: 'active' },
        { id: 'factor2', name: '反转因子V1', expression: '-Rank(Sum($close / Ref($close, 1) - 1, 5))', description: '5日收益率反转因子', createTime: '2024-01-14 15:20:00', status: 'active' }
    ]);
    const [backtestResults, setBacktestResults] = useState([
        { id: 'bt1', name: '动量TopK策略', status: 'completed', createTime: '2024-01-16 09:15:00', totalReturn: '23.5%', sharpe: 1.85, maxDrawdown: '-8.5%' }
    ]);

    // 添加任务
    const addTask = (task) => {
        const newTask = {
            id: `task_${Date.now()}`,
            ...task,
            startTime: new Date().toLocaleString(),
            status: 'running',
            progress: 0
        };
        setTasks(prev => [...prev, newTask]);
        
        // 模拟任务进度
        simulateTaskProgress(newTask.id);
        return newTask.id;
    };

    // 模拟任务进度
    const simulateTaskProgress = (taskId) => {
        let progress = 0;
        const interval = setInterval(() => {
            progress += Math.random() * 15;
            if (progress >= 100) {
                progress = 100;
                setTasks(prev => prev.map(t => 
                    t.id === taskId ? { ...t, progress: 100, status: 'completed' } : t
                ));
                clearInterval(interval);
            } else {
                setTasks(prev => prev.map(t => 
                    t.id === taskId ? { ...t, progress: Math.min(progress, 99) } : t
                ));
            }
        }, 500);
    };

    // 添加数据集
    const addDataset = (dataset) => {
        const newDataset = {
            id: `ds_${Date.now()}`,
            ...dataset,
            status: 'preparing',
            samples: Math.floor(Math.random() * 50000) + 10000,
            features: Math.floor(Math.random() * 30) + 10
        };
        
        setDatasets(prev => [...prev, newDataset]);
        
        // 模拟数据准备完成
        setTimeout(() => {
            setDatasets(prev => prev.map(ds => 
                ds.id === newDataset.id ? { ...ds, status: 'ready' } : ds
            ));
        }, 3000);
        
        return newDataset.id;
    };

    // 添加模型
    const addModel = (model) => {
        const newModel = {
            id: `model_${Date.now()}`,
            ...model,
            status: 'training',
            ic: 0,
            sharpe: 0,
            trainTime: new Date().toLocaleString()
        };
        
        setModels(prev => [...prev, newModel]);
        
        // 模拟训练完成
        setTimeout(() => {
            setModels(prev => prev.map(m => 
                m.id === newModel.id ? { 
                    ...m, 
                    status: 'trained',
                    ic: (Math.random() * 0.05 + 0.02).toFixed(3),
                    sharpe: (Math.random() * 0.8 + 0.8).toFixed(2)
                } : m
            ));
        }, 5000);
        
        return newModel.id;
    };

    // 添加因子
    const addFactor = (factor) => {
        const newFactor = {
            id: `factor_${Date.now()}`,
            ...factor,
            createTime: new Date().toLocaleString(),
            status: 'active'
        };
        setSavedFactors(prev => [...prev, newFactor]);
        return newFactor.id;
    };

    // 测试因子
    const testFactor = async (expression, config) => {
        // 模拟因子测试API调用
        return new Promise((resolve) => {
            setTimeout(() => {
                const mockResult = {
                    ic: (Math.random() * 0.1 - 0.05).toFixed(4),
                    icIR: (Math.random() * 2 + 0.5).toFixed(2),
                    rank_ic: (Math.random() * 0.2 - 0.1).toFixed(4),
                    rank_icIR: (Math.random() * 2 + 0.5).toFixed(2),
                    turnover: (Math.random() * 0.3 + 0.1).toFixed(3),
                    coverage: (Math.random() * 0.2 + 0.8).toFixed(3),
                    validPeriods: Math.floor(Math.random() * 50 + 200)
                };
                resolve(mockResult);
            }, 2000);
        });
    };

    // 添加回测结果
    const addBacktestResult = (result) => {
        const newResult = {
            id: `bt_${Date.now()}`,
            ...result,
            createTime: new Date().toLocaleString()
        };
        setBacktestResults(prev => [...prev, newResult]);
        return newResult.id;
    };

    // 渲染当前页面
    const renderPage = () => {
        switch(currentPage) {
            case 'dashboard':
                return <Dashboard 
                    tasks={tasks} 
                    datasets={datasets} 
                    models={models}
                    savedFactors={savedFactors}
                    onNavigate={setCurrentPage}
                />;
            case 'data':
                return <DataManagement 
                    datasets={datasets} 
                    onAddDataset={addDataset}
                    onAddTask={addTask}
                />;
            case 'factor':
                return <FactorResearch 
                    onSave={addFactor}
                    onTestFactor={testFactor}
                    savedFactors={savedFactors}
                    onNavigate={setCurrentPage}
                />;
            case 'workflow':
                return <QlibWorkflow 
                    onNavigate={setCurrentPage}
                    onAddTask={addTask}
                    savedFactors={savedFactors}
                    models={models}
                    onAddModel={addModel}
                    datasets={datasets}
                />;
            case 'results':
                return <ResultsAnalysis 
                    tasks={tasks}
                    models={models}
                    savedFactors={savedFactors}
                    onNavigate={setCurrentPage}
                />;
            default:
                return <Dashboard 
                    tasks={tasks} 
                    datasets={datasets} 
                    models={models}
                    savedFactors={savedFactors}
                    onNavigate={setCurrentPage}
                />;
        }
    };

    return (
        <div className="app">
            <Layout currentPage={currentPage} onNavigate={setCurrentPage}>
                {renderPage()}
            </Layout>
            
            {/* 任务悬浮窗 */}
            <div className="task-float-panel">
                <div className="task-float-header">
                    任务队列 ({tasks.filter(t => t.status === 'running').length})
                </div>
                <div className="task-float-content">
                    {tasks.slice(-3).reverse().map(task => (
                        <div key={task.id} className="task-item">
                            <div className="task-name">{task.name}</div>
                            <div className="task-progress">
                                <div 
                                    className="task-progress-bar" 
                                    style={{width: `${task.progress}%`}}
                                ></div>
                            </div>
                            <div className="task-status">
                                {task.status === 'running' ? `${Math.floor(task.progress)}%` : task.status}
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

// 渲染应用
ReactDOM.render(<App />, document.getElementById('root'));