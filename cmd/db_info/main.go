package main

import (
	"fmt"
	"log"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()

	// 连接数据库
	db := g.DB()

	// 查询所有表
	tables, err := db.GetAll(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		log.Fatal("Error querying tables:", err)
	}

	fmt.Println("=== SQLite数据库表信息 ===")
	for _, table := range tables {
		tableName := table["name"].String()
		if tableName == "sqlite_sequence" {
			continue // 跳过SQLite内部表
		}

		fmt.Printf("表名: %s\n", tableName)

		// 查询表结构
		schema, err := db.GetAll(ctx, "PRAGMA table_info("+tableName+")")
		if err != nil {
			fmt.Printf("  获取表结构失败: %v\n", err)
			continue
		}

		fmt.Println("  字段信息:")
		for _, field := range schema {
			fmt.Printf("    - %s (%s)\n", field["name"], field["type"])
		}

		// 查询数据条数
		count, err := db.GetValue(ctx, "SELECT COUNT(*) FROM "+tableName)
		if err != nil {
			fmt.Printf("  获取记录数失败: %v\n", err)
		} else {
			fmt.Printf("  记录数: %d\n", count.Int())
		}

		// 显示前几条数据
		if count.Int() > 0 {
			data, err := db.GetAll(ctx, "SELECT * FROM "+tableName+" LIMIT 5")
			if err != nil {
				fmt.Printf("  获取数据失败: %v\n", err)
			} else {
				fmt.Println("  数据示例:")
				for _, row := range data {
					fmt.Printf("    %v\n", row.Map())
				}
			}
		}

		fmt.Println()
	}
}
