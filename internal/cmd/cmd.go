package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"demo/internal/controller/algorithm"
	"demo/internal/controller/user"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化数据库
			initDatabase(ctx)

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
