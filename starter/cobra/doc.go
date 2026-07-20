// Package cobra 提供 Cobra CLI 框架自动配置。
//
// Cobra 是 Go 语言最流行的 CLI 框架。
//
// 功能特性：
//   - 自动配置 Cobra CLI
//   - 子命令支持
//   - 标志位解析
//   - 自动生成帮助文档
//
// 配置示例：
//
//	{
//	  "cobra": {
//	    "enabled": true,
//	    "use": "mycli",
//	    "version": "1.0.0"
//	  }
//	}
//
// 使用示例：
//
//	rootCmd := core.MustGetBean[*cobra.Command](app.Container())
//	rootCmd.AddCommand(&cobra.Command{
//	    Use:   "start",
//	    Short: "Start the application",
//	    Run: func(cmd *cobra.Command, args []string) {
//	        fmt.Println("Starting...")
//	    },
//	})
package cobra
