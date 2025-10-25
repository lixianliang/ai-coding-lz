# imgagent AGENTS.md

## 概述
- 目的：实现小说内容理解，进行小说摘要总结；根据小说内容提取场景内容（人物、地点、事件以及简要对话和旁白）；根据生成的场景调用 llm 生产连环漫画
- 技术栈： 
  - 编程语言: Go 
  - HTTP框架: Gin 
  - 数据库: MySQL 8.4 
  - ORM: GORM 
  - 日志: Zap
  - 单元测试：Testify

## 代码规范
- 编写的代码必须遵循 google go 规范 https://google.github.io/styleguide/go/

## 日志规范
```go
    // 从 gin.Context 获取 log 句柄
	log := logger.FromGinContext(c)

	var args api.RetrieveDatasetsArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		// 参数解析失败
		log.Errorf("Invalid request body: %v", err)
		hutil.AbortError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Infof("Retrieve datasets, args: %v", args)
    // 后续的相关业务逻辑
	// 返回业务数据
	hutil.WriteData(c, result)
```

## 代码模式

## api 规范
- api 遵循 restful 
- 公共 body
    http code 为 200 表示服务接受请求并处理（非 200 不符合预期），具体业务是否处理成功需要查看 body 中的 code 字段，200 表示成功，非 200 表示失败，message 为失败信息；data 为具体业务返回的数据。
```
http code: 200
header: x-reqid
{
    "code": <int>,  // 成功为 200，失败未非 200
    "message": <string>, // 失败信息，成功是为空
    "reqid": <string>,
    "data": <object>
}
```
- 列取 api 的接口响应 data 字段必须为结构体嵌套数组，比如列取知识库（data.datasets）
```shell
GET /v1/datasets

Authorization: Bearer <token>

Response
200
{
    "code": 200,
    "data": {
      "datasets": [
        {
          "id": <string>,
          "name": <string>, 
          "description": <string>,
          "created_at": <string>, 
          "updated_at": <string>
        }
      ]
    }
}
```

## 错误处理

## 单元测试规范
- 必须使用 Testify 编写 ut

## 关键提示
- 所有数据库操作必须通过 gorm 进行
- 日志记录必须使用 zap console 格式化日志方式
