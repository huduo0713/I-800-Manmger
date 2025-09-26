package cmd

import (
	"context"
	"runtime"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"demo/internal/consts"
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
			// 显示版本信息
			printVersionInfo(ctx)

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
	g.Log().Info(ctx, "🗄️ 开始数据库初始化...")

	// 测试数据库连接
	db := g.DB()
	if db == nil {
		g.Log().Error(ctx, "❌ 数据库连接失败: 无法获取数据库实例")
		return
	}

	// 测试数据库连通性
	if err := db.PingMaster(); err != nil {
		g.Log().Error(ctx, "❌ 数据库连接测试失败", g.Map{
			"error": err,
		})
		return
	}

	g.Log().Info(ctx, "✅ 数据库连接测试成功")

	// 获取数据库配置信息
	config := db.GetConfig()
	g.Log().Info(ctx, "📋 数据库配置信息", g.Map{
		"type": config.Type,
		"name": config.Name,
		"host": config.Host,
		"port": config.Port,
	})

	// 读取初始化SQL文件
	sqlFile := "data/init.sql"
	if !gfile.Exists(sqlFile) {
		g.Log().Warning(ctx, "⚠️ SQL初始化文件不存在", g.Map{
			"file": sqlFile,
		})
		return
	}

	sqlContent := gfile.GetContents(sqlFile)
	if sqlContent == "" {
		g.Log().Warning(ctx, "⚠️ SQL初始化文件为空", g.Map{
			"file": sqlFile,
		})
		return
	}

	g.Log().Info(ctx, "📄 SQL初始化文件读取成功", g.Map{
		"file": sqlFile,
		"size": len(sqlContent),
	})

	// 执行SQL初始化
	_, err := db.Exec(ctx, sqlContent)
	if err != nil {
		g.Log().Error(ctx, "❌ 数据库初始化失败", g.Map{
			"error": err,
			"file":  sqlFile,
		})
		return
	}

	// 验证表结构
	tables, err := db.Tables(ctx)
	if err != nil {
		g.Log().Warning(ctx, "⚠️ 无法获取表列表", g.Map{
			"error": err,
		})
	} else {
		g.Log().Info(ctx, "📊 数据库表结构", g.Map{
			"tables": tables,
			"count":  len(tables),
		})
	}

	g.Log().Info(ctx, "✅ 数据库初始化完成")
}

// startMQTTAlgorithmService 启动MQTT算法处理服务
func startMQTTAlgorithmService(ctx context.Context) {
	// 异步启动MQTT算法监听服务，避免阻塞主程序
	go func() {
		// 获取MQTT服务实例
		mqttService := service.Mqtt()

		// 使用MQTT服务中动态检测到的设备ID（MAC地址）
		deviceId := mqttService.GetDeviceId()

		// 等待MQTT连接建立，但不无限等待
		maxWaitTime := 60 // 最多等待60秒
		waitInterval := 2 // 每2秒检查一次
		waited := 0

		for waited < maxWaitTime {
			if mqttService.IsConnected() {
				// MQTT已连接，启动算法消息监听
				err := mqttService.StartAlgorithmMessageListener(deviceId)
				if err != nil {
					g.Log().Error(ctx, "❌ 启动MQTT算法监听服务失败", g.Map{
						"error":    err.Error(),
						"deviceId": deviceId,
						"action":   "将在MQTT重连成功后自动启动",
					})
					return
				}

				g.Log().Info(ctx, "✅ MQTT算法处理服务启动成功", g.Map{
					"deviceId": deviceId,
					"topic":    "/sys/i800/" + deviceId + "/request",
				})
				return
			}

			// MQTT未连接，继续等待
			g.Log().Debug(ctx, "⏳ 等待MQTT连接建立...", g.Map{
				"waited": waited,
				"max":    maxWaitTime,
			})

			time.Sleep(time.Duration(waitInterval) * time.Second)
			waited += waitInterval
		}

		// 超时未连接
		g.Log().Warning(ctx, "⚠️ MQTT连接超时，算法监听服务将在连接成功后自动启动", g.Map{
			"waitedTime": waited,
			"deviceId":   deviceId,
		})
	}()
}

// printVersionInfo 显示版本信息
func printVersionInfo(ctx context.Context) {
	g.Log().Info(ctx, "🚀 "+consts.AppName+" 启动", g.Map{
		"version":   consts.AppVersion,
		"buildTime": consts.BuildTime,
		"gitCommit": consts.GitCommit,
		"gitBranch": consts.GitBranch,
		"goVersion": runtime.Version(),
		"author":    consts.Author,
	})
}
