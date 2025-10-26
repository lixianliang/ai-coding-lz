# AGENTS.md（根目录）

## 项目概述
   ai-coding-lz 是一个连环动漫智能体，可根据用户输入一篇小说，自动理解小说内容摘要、人物特征、风格情节；根据小说内的场景生成动漫连环画；每个场景包含多张图片，每张图片配备对应文字，文字就是场景中的对话或者旁白，语音就是文字转换为 tts 的结果。

代码库分为：
- 前端 web: 使用 vue3 构建
- 后端 imgagent: 使用 go 语言实现，对应的说明在 imgagent/AGENTS.md
- books 目录为小说数据

## 前端 web
``` shell
cd web
npm run dev
```

## 后端 imgagent

```shell
cd imgagent
go build main.go
./imgagent -f ./imgagent.json
```

## 通用实践
- 优先编辑现有文件；仅在需要时添加新文档

