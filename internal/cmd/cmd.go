package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"demo/internal/controller/algorithm"
	"demo/internal/controller/user"
	"demo/internal/service"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化数据库
			initDatabase(ctx)

			// 启动MQTT算法监听服务
			startMQTTAlgorithmService(ctx)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					user.NewV1(),
					algorithm.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)

// initDatabase 初始化数据库表结构
func initDatabase(ctx context.Context) {
	// 读取初始化SQL文件
	sqlFile := "data/init.sql"
	if !gfile.Exists(sqlFile) {
		g.Log().Warning(ctx, "SQL init file not found:", sqlFile)
		return
	}

	sqlContent := gfile.GetContents(sqlFile)
	if sqlContent == "" {
		g.Log().Warning(ctx, "SQL init file is empty:", sqlFile)
		return
	}

	// 执行SQL初始化
	db := g.DB()
	_, err := db.Exec(ctx, sqlContent)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to initialize database: %v", err)
		return
	}

	g.Log().Info(ctx, "Database initialized successfully")
}

// startMQTTAlgorithmService 启动MQTT算法处理服务
func startMQTTAlgorithmService(ctx context.Context) {
	// 从配置文件获取设备ID，如果没有配置则使用默认值
	deviceId := g.Cfg().MustGet(ctx, "device.id", "edge-device-001").String()

	// 获取MQTT服务实例
	mqttService := service.Mqtt()

	// 启动算法消息监听
	err := mqttService.StartAlgorithmMessageListener(deviceId)
	if err != nil {
		g.Log().Errorf(ctx, "启动MQTT算法监听服务失败: %v", err)
		return
	}

	g.Log().Info(ctx, "MQTT算法处理服务启动成功", g.Map{
		"deviceId": deviceId,
		"topic":    "/sys/i800/" + deviceId + "/request",
	})
}
