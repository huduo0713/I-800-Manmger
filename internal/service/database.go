package service

import (
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

type sDatabase struct{}

var (
	dbService *sDatabase
	dbOnce    sync.Once
)

// Database 返回数据库服务单例
func Database() *sDatabase {
	dbOnce.Do(func() {
		dbService = &sDatabase{}
		dbService.initDatabase()
	})
	return dbService
}

// initDatabase 初始化数据库表结构
func (s *sDatabase) initDatabase() {
	ctx := gctx.New()

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
