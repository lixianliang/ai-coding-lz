# ai-coding-lz

    根据小说内容，自动提取场景实现连环漫画制作。

# 程序运行说明

1.先安装mysql(参考官方) ,安装后，创建DataBase,名称为 imgagent:
  CREATE DATABASE imgagent;

2.把代码clone下来:
    git clone  --depth=1  github.com:lixianliang/ai-coding-lz.git

3.修改后端配置 imgagent.json :
   配置 db.password 为mysql密码,
   配置 bailian.api_key 为阿里云百炼AI大模型APIKey

3.编译并启动后端服务(需安装golang)
  编译：cd imgagent && go build main.go
  运行：./main -f imgagent.json

4.启动前端服务（需安装npm)
  cd web && npm run dev  

5.浏览器里访问 "<http://localhost:3000>"  就可以
