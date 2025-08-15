// Qlib API 接口封装 - 与Golang后端通信
class QlibAPI {
    constructor(baseURL = 'http://localhost:8080/api/v1') {
        this.baseURL = baseURL;
        this.token = localStorage.getItem('qlib_token');
    }

    // 通用请求方法
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': this.token ? `Bearer ${this.token}` : '',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || `HTTP ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('API Request Error:', error);
            throw error;
        }
    }

    // GET 请求
    async get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(url, { method: 'GET' });
    }

    // POST 请求
    async post(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    // PUT 请求
    async put(endpoint, data = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    // DELETE 请求
    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // ============ 数据管理接口 ============
    
    // 获取数据集列表
    async getDatasets(page = 1, limit = 20, status = '') {
        return this.get('/datasets', { page, limit, status });
    }

    // 创建数据集
    async createDataset(datasetConfig) {
        return this.post('/datasets', datasetConfig);
    }

    // 获取数据集详情
    async getDatasetDetail(datasetId) {
        return this.get(`/datasets/${datasetId}`);
    }

    // 更新数据集
    async updateDataset(datasetId, config) {
        return this.put(`/datasets/${datasetId}`, config);
    }

    // 删除数据集
    async deleteDataset(datasetId) {
        return this.delete(`/datasets/${datasetId}`);
    }

    // 获取股票池列表
    async getUniverses() {
        return this.get('/universes');
    }

    // 获取数据字段信息
    async getDataFields() {
        return this.get('/data/fields');
    }

    // 数据预览
    async previewData(datasetId, limit = 100) {
        return this.get(`/datasets/${datasetId}/preview`, { limit });
    }

    // ============ 因子管理接口 ============
    
    // 获取因子列表
    async getFactors(page = 1, limit = 20, category = '') {
        return this.get('/factors', { page, limit, category });
    }

    // 创建因子
    async createFactor(factorConfig) {
        return this.post('/factors', factorConfig);
    }

    // 测试因子表达式
    async testFactor(expression, datasetId, startDate, endDate) {
        return this.post('/factors/test', {
            expression,
            dataset_id: datasetId,
            start_date: startDate,
            end_date: endDate
        });
    }

    // 获取因子详情
    async getFactorDetail(factorId) {
        return this.get(`/factors/${factorId}`);
    }

    // 计算因子值
    async calculateFactor(factorId, datasetId, startDate, endDate) {
        return this.post('/factors/calculate', {
            factor_id: factorId,
            dataset_id: datasetId,
            start_date: startDate,
            end_date: endDate
        });
    }

    // 因子分析
    async analyzeFactor(factorId, analysisConfig) {
        return this.post(`/factors/${factorId}/analyze`, analysisConfig);
    }

    // 获取因子分析结果
    async getFactorAnalysis(analysisId) {
        return this.get(`/factor-analysis/${analysisId}`);
    }

    // 验证因子表达式语法
    async validateFactorExpression(expression) {
        return this.post('/factors/validate', { expression });
    }

    // 获取因子操作符和函数库
    async getFactorLibrary() {
        return this.get('/factors/library');
    }

    // ============ 模型管理接口 ============
    
    // 获取模型列表
    async getModels(page = 1, limit = 20, status = '') {
        return this.get('/models', { page, limit, status });
    }

    // 创建模型
    async createModel(modelConfig) {
        return this.post('/models', modelConfig);
    }

    // 训练模型
    async trainModel(modelId, trainConfig) {
        return this.post(`/models/${modelId}/train`, trainConfig);
    }

    // 获取模型详情
    async getModelDetail(modelId) {
        return this.get(`/models/${modelId}`);
    }

    // 获取模型训练进度
    async getTrainingProgress(taskId) {
        return this.get(`/training-tasks/${taskId}`);
    }

    // 模型预测
    async predictModel(modelId, predictConfig) {
        return this.post(`/models/${modelId}/predict`, predictConfig);
    }

    // 获取模型评估结果
    async getModelEvaluation(modelId) {
        return this.get(`/models/${modelId}/evaluation`);
    }

    // 删除模型
    async deleteModel(modelId) {
        return this.delete(`/models/${modelId}`);
    }

    // 导出模型
    async exportModel(modelId, format = 'pickle') {
        return this.get(`/models/${modelId}/export`, { format });
    }

    // ============ 策略回测接口 ============
    
    // 获取回测列表
    async getBacktests(page = 1, limit = 20, status = '') {
        return this.get('/backtests', { page, limit, status });
    }

    // 创建回测任务
    async createBacktest(backtestConfig) {
        return this.post('/backtests', backtestConfig);
    }

    // 运行回测
    async runBacktest(backtestId) {
        return this.post(`/backtests/${backtestId}/run`);
    }

    // 获取回测详情
    async getBacktestDetail(backtestId) {
        return this.get(`/backtests/${backtestId}`);
    }

    // 获取回测结果
    async getBacktestResults(backtestId) {
        return this.get(`/backtests/${backtestId}/results`);
    }

    // 获取回测进度
    async getBacktestProgress(taskId) {
        return this.get(`/backtest-tasks/${taskId}`);
    }

    // 对比多个回测结果
    async compareBacktests(backtestIds) {
        return this.post('/backtests/compare', { backtest_ids: backtestIds });
    }

    // 获取策略净值数据
    async getStrategyNetValue(backtestId, startDate, endDate) {
        return this.get(`/backtests/${backtestId}/nav`, { 
            start_date: startDate, 
            end_date: endDate 
        });
    }

    // 获取策略持仓数据
    async getStrategyPositions(backtestId, date) {
        return this.get(`/backtests/${backtestId}/positions`, { date });
    }

    // 获取策略交易记录
    async getStrategyTrades(backtestId, page = 1, limit = 100) {
        return this.get(`/backtests/${backtestId}/trades`, { page, limit });
    }

    // 导出回测报告
    async exportBacktestReport(backtestId, format = 'pdf') {
        return this.get(`/backtests/${backtestId}/export`, { format });
    }

    // ============ 任务管理接口 ============
    
    // 获取任务列表
    async getTasks(page = 1, limit = 20, status = '', type = '') {
        return this.get('/tasks', { page, limit, status, type });
    }

    // 获取任务详情
    async getTaskDetail(taskId) {
        return this.get(`/tasks/${taskId}`);
    }

    // 取消任务
    async cancelTask(taskId) {
        return this.post(`/tasks/${taskId}/cancel`);
    }

    // 重试任务
    async retryTask(taskId) {
        return this.post(`/tasks/${taskId}/retry`);
    }

    // 删除任务
    async deleteTask(taskId) {
        return this.delete(`/tasks/${taskId}`);
    }

    // 获取任务日志
    async getTaskLogs(taskId, page = 1, limit = 100) {
        return this.get(`/tasks/${taskId}/logs`, { page, limit });
    }

    // ============ 系统管理接口 ============
    
    // 获取系统状态
    async getSystemStatus() {
        return this.get('/system/status');
    }

    // 获取系统配置
    async getSystemConfig() {
        return this.get('/system/config');
    }

    // 更新系统配置
    async updateSystemConfig(config) {
        return this.put('/system/config', config);
    }

    // 获取系统资源使用情况
    async getSystemResources() {
        return this.get('/system/resources');
    }

    // 获取Qlib状态
    async getQlibStatus() {
        return this.get('/qlib/status');
    }

    // 重启Qlib服务
    async restartQlib() {
        return this.post('/qlib/restart');
    }

    // 获取数据源状态
    async getDataSourceStatus() {
        return this.get('/data-sources/status');
    }

    // 同步数据源
    async syncDataSource(sourceName) {
        return this.post(`/data-sources/${sourceName}/sync`);
    }

    // ============ 用户管理接口 ============
    
    // 用户登录
    async login(username, password) {
        const response = await this.post('/auth/login', { username, password });
        if (response.token) {
            this.token = response.token;
            localStorage.setItem('qlib_token', response.token);
        }
        return response;
    }

    // 用户登出
    async logout() {
        try {
            await this.post('/auth/logout');
        } finally {
            this.token = null;
            localStorage.removeItem('qlib_token');
        }
    }

    // 获取用户信息
    async getUserInfo() {
        return this.get('/auth/user');
    }

    // 刷新token
    async refreshToken() {
        const response = await this.post('/auth/refresh');
        if (response.token) {
            this.token = response.token;
            localStorage.setItem('qlib_token', response.token);
        }
        return response;
    }

    // ============ WebSocket 连接（实时数据）============
    
    // 创建WebSocket连接
    connectWebSocket(endpoint, onMessage, onError, onClose) {
        const wsUrl = this.baseURL.replace('http', 'ws') + endpoint;
        const ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('WebSocket connected to:', endpoint);
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                onMessage(data);
            } catch (error) {
                console.error('WebSocket message parse error:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (onError) onError(error);
        };

        ws.onclose = (event) => {
            console.log('WebSocket closed:', event);
            if (onClose) onClose(event);
        };

        return ws;
    }

    // 订阅任务进度更新
    subscribeTaskProgress(taskId, onProgress) {
        return this.connectWebSocket(
            `/ws/tasks/${taskId}/progress`,
            onProgress,
            (error) => console.error('Task progress subscription error:', error)
        );
    }

    // 订阅系统状态更新
    subscribeSystemStatus(onStatusUpdate) {
        return this.connectWebSocket(
            '/ws/system/status',
            onStatusUpdate,
            (error) => console.error('System status subscription error:', error)
        );
    }

    // 订阅实时日志
    subscribeTaskLogs(taskId, onLogUpdate) {
        return this.connectWebSocket(
            `/ws/tasks/${taskId}/logs`,
            onLogUpdate,
            (error) => console.error('Task logs subscription error:', error)
        );
    }
}

// API实例
const qlib = new QlibAPI();

// 导出API实例和类
window.QlibAPI = QlibAPI;
window.qlib = qlib;

// 为React组件提供API hooks
const useQlibAPI = () => {
    return qlib;
};

window.useQlibAPI = useQlibAPI;