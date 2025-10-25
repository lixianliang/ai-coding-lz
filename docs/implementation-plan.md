# 实施计划文档

## 一、开发阶段划分

### 阶段 1: 后端基础设施（预计 2-3 天）

#### 1.1 数据库层扩展
**任务：**
- 修复 Document 表状态常量（DocumentStatusChapterReady）
- 添加 Role 和 Scene 表的 AutoMigrate
- 实现 Role DAO 方法（CreateRoles, ListRoles）
- 实现 Scene DAO 方法（CreateScenes, ListScenes, UpdateSceneImageURL 等）
- 实现 Document 新增查询方法（ListChapterReadyDocuments, ListSceneReadyDocuments）
- 实现 Chapter 更新方法（UpdateChapterSceneIDs）

**验收标准：**
- 所有 DAO 方法通过单元测试
- 数据库表结构正确创建
- 索引正确建立

#### 1.2 阿里云百炼客户端
**任务：**
- 创建 bailian 包结构
- 实现 Client 结构和配置
- 实现文件上传方法（UploadFile）
- 实现角色提取方法（ExtractRoles）
- 实现场景生成方法（GenerateScenes）
- 实现图片生成方法（GenerateImage）
- 编写默认 Prompt
- 单元测试（使用 mock）

**验收标准：**
- 所有方法实现完成
- 错误处理完善
- 日志记录完整
- 单元测试覆盖

#### 1.3 配置和服务初始化
**任务：**
- 修改 imgagent.json 添加 bailian 和 document_mgr 配置
- 修改 main.go Config 结构
- 修改 svr.go Config 和 Service 结构
- 初始化 bailian Client
- 初始化 DocumentMgr

**验收标准：**
- 配置文件格式正确
- 服务启动成功
- 配置加载正确

### 阶段 2: 后端异步任务（预计 2-3 天）

#### 2.1 完善 DocumentMgr
**任务：**
- 添加 bailianClient 字段
- 完善 HandleDocumentScene 方法：
  - 上传文件逻辑
  - 角色提取逻辑
  - 场景生成逻辑
  - 错误处理
- 实现 HandleImageGenTasks 和 loopHandleImageGenTasks
- 实现 HandleDocumentImageGen 方法
- 在 main.go 中启动 DocumentMgr

**验收标准：**
- Worker 1 和 Worker 2 正常运行
- 轮询机制正常工作
- 状态流转正确
- 错误恢复机制有效

#### 2.2 集成测试
**任务：**
- 准备测试文件（小说）
- 测试完整流程：上传 → 场景提取 → 图片生成
- 测试错误恢复：模拟 API 失败
- 测试并发：多个文档同时处理
- 性能测试：大文件、多章节

**验收标准：**
- 完整流程正常运行
- 错误处理符合预期
- 性能满足要求

### 阶段 3: 后端 API 扩展（预计 1 天）

#### 3.1 修改现有接口
**任务：**
- 修改 HandleCreateDocument：
  - 不调用 bailian 上传
  - 保存文件到 temp（使用 docID 命名）
  - 设置状态为 chapterReady

**验收标准：**
- 文档上传正常
- 文件保存正确
- 状态设置正确

#### 3.2 新增接口
**任务：**
- 实现 HandleGetRoles
- 实现 HandleListScenesByDocument
- 实现 HandleListScenesByChapter
- 添加 API 类型定义（Role, Scene）
- 注册路由

**验收标准：**
- 所有接口正常工作
- 返回数据格式正确
- 错误处理完善

#### 3.3 更新 API 文档
**任务：**
- 更新 docs/API.md
- 添加新接口文档
- 更新状态说明

### 阶段 4: 前端项目初始化（预计 1 天）

#### 4.1 项目创建
**任务：**
- 使用 Vite 创建 Vue3 + TypeScript 项目
- 安装依赖：vue-router, pinia, element-plus, axios, sass
- 配置 vite.config.ts（alias, proxy）
- 配置 tsconfig.json
- 创建目录结构

**验收标准：**
- 项目可以正常启动
- 依赖安装成功
- 配置正确

#### 4.2 基础配置
**任务：**
- 创建 router/index.ts
- 创建 stores/index.ts
- 创建 styles/variables.scss
- 创建 styles/global.scss
- 配置 Element Plus

**验收标准：**
- 路由工作正常
- 状态管理可用
- 样式系统就位

### 阶段 5: 前端核心功能（预计 3-4 天）

#### 5.1 API 层实现
**任务：**
- 实现 apis/types/common.ts
- 实现 apis/types/document.ts
- 实现 apis/request.ts
- 实现 apis/modules/document.ts

**验收标准：**
- 类型定义完整
- 请求封装正确
- API 调用成功

#### 5.2 状态管理
**任务：**
- 实现 documentStore
- 实现所有状态和方法
- 实现轮询机制

**验收标准：**
- 状态管理正常
- 轮询机制工作正常

#### 5.3 基础组件
**任务：**
- 实现 BaseButton.vue
- 实现 BaseCard.vue
- 实现 StatusTag.vue

**验收标准：**
- 组件可复用
- 样式统一

#### 5.4 业务组件
**任务：**
- 实现 DocumentCard.vue
- 实现 RoleCard.vue
- 实现 SceneCard.vue

**验收标准：**
- 组件功能完整
- 交互正常

#### 5.5 页面实现
**任务：**
- 实现 DocumentList.vue：
  - 文档列表展示
  - 上传功能
  - 删除功能
- 实现 DocumentDetail.vue：
  - 文档信息展示
  - 状态轮询
  - 角色列表
  - 章节列表
- 实现 SceneView.vue：
  - 场景列表
  - 图片展示
  - 图片预览

**验收标准：**
- 所有页面功能完整
- 交互流畅
- 样式美观

### 阶段 6: 集成测试和优化（预计 1-2 天）

#### 6.1 端到端测试
**任务：**
- 测试完整用户流程
- 测试各种边界情况
- 测试错误处理
- 浏览器兼容性测试

**验收标准：**
- 所有功能正常
- 无明显 bug

#### 6.2 性能优化
**任务：**
- 前端代码优化
- 图片懒加载
- 请求优化
- 打包优化

**验收标准：**
- 页面加载速度快
- 交互响应及时

#### 6.3 文档完善
**任务：**
- 完善 README
- 编写部署文档
- 编写使用手册

## 二、依赖关系

```
后端数据库层
    ↓
后端 bailian 客户端
    ↓
后端配置和初始化
    ↓
后端异步任务 ← 依赖数据库层和 bailian
    ↓
后端 API 扩展 ← 依赖数据库层
    ↓
前端项目初始化（独立）
    ↓
前端 API 层 ← 依赖后端 API
    ↓
前端组件和页面 ← 依赖 API 层
    ↓
集成测试
```

## 三、测试计划

### 3.1 单元测试

**后端：**
- DAO 层测试（使用 testify）
- bailian 客户端测试（使用 mock）
- Worker 逻辑测试（使用 mock）

**前端：**
- 工具函数测试
- 组件测试（可选）

### 3.2 集成测试

**后端：**
- 完整流程测试
- 错误恢复测试
- 并发测试

**前端：**
- API 调用测试
- 状态管理测试

### 3.3 端到端测试

**场景：**
1. 上传小说 → 查看详情 → 等待处理 → 查看场景
2. 上传多个文档 → 并发处理
3. 删除文档
4. 错误场景：无效文件、网络错误等

### 3.4 性能测试

**指标：**
- 文档上传时间
- 章节分割时间
- 场景提取时间（单个文档）
- 图片生成时间（单个场景）
- 前端页面加载时间

**目标：**
- 10 章节文档上传 < 10s
- 场景提取 < 5min
- 图片生成 < 3min
- 页面加载 < 2s

## 四、部署方案

### 4.1 后端部署

**环境要求：**
- Go 1.21+
- MySQL 8.4+
- Redis 6.0+

**部署步骤：**
1. 编译：`go build main.go`
2. 配置 imgagent.json
3. 初始化数据库
4. 启动服务：`./imgagent -f imgagent.json`

**进程管理：**
- 使用 systemd 或 supervisor
- 配置自动重启
- 配置日志轮转

### 4.2 前端部署

**构建：**
```bash
cd web
npm install
npm run build
```

**部署方式：**
1. Nginx 静态文件服务
2. 配置反向代理到后端

**Nginx 配置示例：**
```nginx
server {
    listen 80;
    server_name example.com;
    
    root /path/to/web/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /v1 {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4.3 环境配置

**生产环境配置：**
- 数据库：使用独立 MySQL 服务器
- Redis：使用 Redis 集群
- 日志：配置文件日志和轮转
- 监控：配置监控告警

**配置检查清单：**
- [ ] 数据库连接配置
- [ ] Redis 连接配置
- [ ] 阿里云 API Key
- [ ] 日志路径和级别
- [ ] Worker 轮询间隔
- [ ] Nginx 配置
- [ ] 防火墙规则
- [ ] SSL 证书（如需要）

## 五、风险和应对

### 5.1 技术风险

**风险 1: 阿里云 API 稳定性**
- 影响：处理失败或延迟
- 应对：实现重试机制，记录详细日志

**风险 2: LLM 返回格式不符**
- 影响：解析失败
- 应对：增强 Prompt，容错解析

**风险 3: 大文件处理**
- 影响：内存溢出，处理超时
- 应对：限制文件大小，优化内存使用

### 5.2 进度风险

**风险：开发时间超出预期**
- 应对：
  - 优先实现核心功能
  - 非关键功能可以后续迭代
  - 增加人力投入

### 5.3 质量风险

**风险：测试不充分导致线上问题**
- 应对：
  - 完善单元测试
  - 充分的集成测试
  - 灰度发布

## 六、后续优化方向

### 6.1 功能优化

- 支持更多文件格式
- TTS 语音合成
- 场景编辑功能
- 导出功能（PDF、视频）

### 6.2 性能优化

- Worker 并发处理
- 图片 CDN 加速
- 前端缓存优化
- 数据库查询优化

### 6.3 用户体验

- 更丰富的状态提示
- 处理进度百分比
- 实时通知（WebSocket）
- 移动端适配

## 七、时间估算

**总计：约 10-15 个工作日**

- 后端基础设施：2-3 天
- 后端异步任务：2-3 天
- 后端 API：1 天
- 前端初始化：1 天
- 前端开发：3-4 天
- 集成测试：1-2 天
- 文档和部署：1 天

**注：以上为单人开发估算，实际时间可能根据经验和具体情况调整**

