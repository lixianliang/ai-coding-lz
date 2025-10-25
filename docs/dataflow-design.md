# 数据流转设计文档

## 一、完整业务流程

### 1.1 总览流程图（文字描述）

```
用户上传文档
    ↓
前端: POST /v1/documents (name + file)
    ↓
后端: 
  - 保存文件到 temp 目录
  - 章节分割（spliter.Split）
  - 创建 Document 记录（status: chapterReady）
  - 创建 Chapter 记录
    ↓
前端: 
  - 接收文档信息
  - 跳转到详情页
  - 开始轮询状态（每5秒）
    ↓
后端 Worker 1 (场景提取):
  - 轮询 chapterReady 文档（每30秒）
  - 上传文件到阿里云百炼 → 获取 fileID
  - 调用 qwen-long 提取角色 → 保存到 Role 表
  - 遍历章节，调用 qwen-long 生成场景 → 保存到 Scene 表
  - 更新 Chapter.SceneIDs
  - 更新 Document.status = sceneReady
    ↓
后端 Worker 2 (图片生成):
  - 轮询 sceneReady 文档（每30秒）
  - 获取所有未生成图片的场景
  - 遍历场景，调用 qwen-image-plus 生成图片
  - 更新 Scene.ImageURL
  - 所有场景完成后，更新 Document.status = imgReady
    ↓
前端:
  - 轮询检测到 status = imgReady
  - 停止轮询
  - 加载场景列表
  - 显示完成状态
```

### 1.2 详细阶段说明

#### 阶段 1: 文档上传（同步）
**时长：** 几秒到几十秒（取决于文件大小和章节数）

**前端操作：**
1. 用户填写文档名称
2. 选择本地文件
3. 点击上传按钮
4. 显示上传进度
5. 上传成功后跳转到详情页

**后端操作：**
1. 接收文件和名称
2. 验证文件类型和名称唯一性
3. 保存文件到 temp 目录（命名：{docID}.{ext}）
4. 调用 spliter.Split 进行章节分割
5. 创建 Document 记录（status: chapterReady）
6. 批量创建 Chapter 记录
7. 返回文档信息

**数据变化：**
- 新增 1 条 Document 记录
- 新增 N 条 Chapter 记录（N = 章节数）
- temp 目录新增文件

#### 阶段 2: 场景提取（异步）
**时长：** 几分钟到几十分钟（取决于章节数和 LLM 响应速度）

**Worker 1 处理流程：**

**步骤 1: 上传文件到百炼**
- 检查 Document.FileID 是否为空
- 如果为空，从 temp 读取文件并上传
- 更新 Document.FileID

**步骤 2: 提取角色**
- 构造角色提取 Prompt
- 调用 qwen-long（使用 fileID）
- 解析返回的 JSON 数组
- 批量创建 Role 记录

**步骤 3: 生成场景**
- 遍历所有章节
- 为每个章节调用 qwen-long 生成场景（0-3个）
- 为每个场景创建 Scene 记录
- 更新 Chapter.SceneIDs

**步骤 4: 更新状态**
- 更新 Document.status = sceneReady

**数据变化：**
- Document.FileID 更新
- 新增 M 条 Role 记录（M = 角色数，通常 3-10）
- 新增 K 条 Scene 记录（K = 场景总数，通常章节数 * 1-2）
- Chapter.SceneIDs 更新
- Document.status 更新

#### 阶段 3: 图片生成（异步）
**时长：** 几分钟到几十分钟（取决于场景数和图片生成速度）

**Worker 2 处理流程：**

**步骤 1: 获取待处理场景**
- 查询所有 ImageURL 为空的场景

**步骤 2: 生成图片**
- 遍历每个场景
- 调用 qwen-image-plus 生成图片
- 更新 Scene.ImageURL

**步骤 3: 更新状态**
- 检查是否所有场景都已生成图片
- 更新 Document.status = imgReady

**数据变化：**
- Scene.ImageURL 更新
- Document.status 更新为 imgReady

#### 阶段 4: 完成展示（前端）
**前端操作：**
1. 轮询检测到 status = imgReady
2. 停止轮询
3. 加载场景列表
4. 显示所有场景图片
5. 用户可以查看、预览图片

## 二、状态机设计

### 2.1 Document 状态流转

```
[chapterReady] ────────────> [sceneReady] ────────────> [imgReady]
   (初始状态)      Worker 1        (中间状态)      Worker 2      (最终状态)
       ↑                                                              
       │                                                              
       └────────────── 错误重试保持状态 ───────────────┘
```

### 2.2 状态定义

**chapterReady（章节就绪）**
- 含义：文档上传成功，章节分割完成
- 触发条件：文档上传 API 成功返回
- 下一步：等待 Worker 1 提取角色和场景

**sceneReady（场景就绪）**
- 含义：角色和场景提取完成
- 触发条件：Worker 1 处理成功
- 下一步：等待 Worker 2 生成图片

**imgReady（图片就绪）**
- 含义：所有场景图片生成完成
- 触发条件：Worker 2 处理成功
- 下一步：无，最终状态

### 2.3 状态转换条件

**chapterReady → sceneReady:**
- 前置条件：
  - Document.status = chapterReady
  - 至少有一个 Chapter 记录
- 处理步骤：
  - 上传文件到百炼（如果 FileID 为空）
  - 提取角色信息
  - 为每个章节生成场景
- 后置条件：
  - Role 记录已创建
  - Scene 记录已创建
  - Chapter.SceneIDs 已更新

**sceneReady → imgReady:**
- 前置条件：
  - Document.status = sceneReady
  - 至少有一个 Scene 记录
- 处理步骤：
  - 为每个场景生成图片
- 后置条件：
  - 所有 Scene.ImageURL 非空

### 2.4 异常状态处理

**原则：失败保持当前状态，下次轮询继续处理**

**场景 1: Worker 1 处理失败**
- 情况：API 调用失败、解析错误等
- 处理：记录错误日志，保持 chapterReady 状态
- 恢复：下次轮询重新处理

**场景 2: Worker 2 处理失败**
- 情况：图片生成失败
- 处理：记录错误日志，保持 sceneReady 状态
- 恢复：下次轮询继续处理未完成的场景

**场景 3: 部分成功**
- Worker 1: 如果角色提取成功但场景生成部分失败，下次重新处理
- Worker 2: 只处理 ImageURL 为空的场景，已成功的不重复处理

## 三、轮询机制设计

### 3.1 后端 Worker 轮询

#### Worker 1: 场景提取轮询

**配置：**
- 轮询间隔：30秒（可配置）
- 并发控制：串行处理，一次处理一个文档

**轮询逻辑：**
```
每 30 秒：
  1. 查询 status = chapterReady 的文档
  2. 逐个处理：
     - 调用 HandleDocumentScene
     - 成功：更新状态为 sceneReady
     - 失败：记录日志，跳过，下次继续
```

**优化考虑：**
- 可以按创建时间排序，优先处理旧文档
- 可以限制每次处理的文档数量
- 可以记录处理耗时，监控性能

#### Worker 2: 图片生成轮询

**配置：**
- 轮询间隔：30秒（可配置）
- 并发控制：串行处理

**轮询逻辑：**
```
每 30 秒：
  1. 查询 status = sceneReady 的文档
  2. 逐个处理：
     - 获取未生成图片的场景
     - 逐个生成图片
     - 成功：更新状态为 imgReady
     - 失败：记录日志，跳过，下次继续
```

### 3.2 前端状态轮询

**触发时机：**
- 文档详情页加载时
- 检查 currentDocument.status !== 'imgReady'
- 开始轮询

**轮询配置：**
- 轮询间隔：5秒
- 超时时间：30分钟（360次轮询）
- 停止条件：
  - 文档状态变为 imgReady
  - 页面卸载（onUnmounted）
  - 超时（可选）

**轮询实现：**
```typescript
const startPolling = (docId: string) => {
  stopPolling() // 清除之前的定时器
  
  let count = 0
  const maxCount = 360 // 30分钟超时
  
  pollingTimer.value = setInterval(async () => {
    count++
    
    // 超时检查
    if (count > maxCount) {
      stopPolling()
      ElMessage.warning('处理超时，请稍后刷新查看')
      return
    }
    
    try {
      const doc = await documentApi.get(docId)
      currentDocument.value = doc
      
      // 状态变化通知
      if (doc.status === 'sceneReady') {
        // 可以加载角色信息
        await fetchRoles(docId)
      } else if (doc.status === 'imgReady') {
        // 加载场景列表
        await fetchScenes(docId)
        stopPolling()
        ElMessage.success('处理完成！')
      }
    } catch (error) {
      console.error('Polling error:', error)
    }
  }, 5000)
}
```

## 四、数据一致性保证

### 4.1 幂等性处理

**问题：** Worker 重复处理同一个文档

**解决方案：**

**角色提取：**
- 每次处理前先删除该文档的所有角色
- 重新提取并创建

**场景生成：**
- 检查 Chapter.SceneIDs 是否为空
- 如果已有场景，跳过该章节
- 或者删除旧场景，重新生成

**图片生成：**
- 只查询 ImageURL 为空的场景
- 已生成的不重复处理

### 4.2 事务处理

**原则：** 尽量使用数据库事务保证一致性

**说明：** 由于服务端为单节点部署，无并发冲突，可以简化事务处理

**场景生成事务：**
```go
// 伪代码
tx.Begin()
  - 创建 Scene 记录（批量）
  - 更新 Chapter.SceneIDs
tx.Commit()
```

**简化方案：**
- 对于批量操作，可以使用 GORM 的批量方法
- 对于关键操作（如状态更新），使用事务保证原子性
- 无需考虑并发冲突和锁机制

### 4.3 级联删除

**删除文档时：**
1. 删除所有 Chapter
2. 删除所有 Scene
3. 删除所有 Role
4. 删除 Document
5. 删除 temp 文件（可选）

## 五、性能优化

### 5.1 批量操作

- 批量创建 Chapter（已实现）
- 批量创建 Scene
- 批量创建 Role

### 5.2 并发控制

**当前：串行处理**
- 优点：简单可靠，避免冲突
- 缺点：处理速度慢

**未来优化：**
- Worker 2 可以并发生成图片（多个场景同时生成）
- 使用 goroutine pool 控制并发数
- 需要考虑 API 速率限制

### 5.3 缓存策略

**前端缓存：**
- 文档列表缓存 5 分钟
- 文档详情根据状态决定是否缓存

**后端缓存：**
- 文档状态可以放入 Redis
- 减少数据库查询

## 六、监控和日志

### 6.1 关键指标

**处理耗时：**
- 文档上传耗时
- 场景提取耗时（单个文档）
- 图片生成耗时（单个场景）

**队列长度：**
- chapterReady 文档数
- sceneReady 文档数

**成功率：**
- Worker 1 处理成功率
- Worker 2 处理成功率

### 6.2 日志记录

**级别：**
- Info: 处理开始、完成
- Warn: 重试、超时
- Error: 处理失败、异常

**内容：**
- 文档ID、章节ID、场景ID
- 操作类型
- 耗时
- 错误详情

## 七、容错和恢复

### 7.1 服务重启

**问题：** 服务重启时，正在处理的文档如何恢复？

**解决：**
- 所有状态存储在数据库
- Worker 重启后自动从数据库恢复
- 继续处理未完成的文档

### 7.2 网络故障

**问题：** 阿里云 API 调用失败

**解决：**
- 不在客户端层重试
- Worker 下次轮询自动重试
- 记录错误日志

### 7.3 数据恢复

**场景：** 意外删除或数据损坏

**预防：**
- 定期数据库备份
- 保留原始文件（temp 目录）
- 可以重新处理

