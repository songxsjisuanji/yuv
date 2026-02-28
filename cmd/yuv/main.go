package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	pkg "yuv/pkg/pkgmgr"
	"yuv/pkg/repo"
	"yuv/pkg/system"
)

var (
	detector  = system.NewDetector()
	repoMgr   = repo.NewManager()
	packageMgr = pkg.NewManager()
)

func main() {
	// 检测系统是否支持
	supported, err := detector.IsSupported()
	if err != nil {
		log.Fatalf("检测系统失败: %v", err)
	}
	if !supported {
		log.Fatalf("当前系统不支持")
	}

	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "yuv",
		Short: "轻量级高性能 YUM/DNF 增强工具",
		Long: `yuv (yum+uv) 是一款轻量级、高性能、零负担的 Linux RPM 包管理器增强工具。
它提供一键式 yum 源管理和快速包安装功能，兼容 yum/dnf 命令。`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// 源管理命令组
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "管理 YUM 仓库",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// use 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "use [repo]",
		Short: "切换到指定的公共仓库",
		Example: "yuv repo use aliyun",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			releasever, err := detector.GetReleasever()
			if err != nil {
				log.Fatalf("获取 releasever 失败: %v", err)
			}
			basearch, err := detector.GetBasearch()
			if err != nil {
				log.Fatalf("获取 basearch 失败: %v", err)
			}

			repoName := args[0]
			if err := repoMgr.Use(repoName, releasever, basearch); err != nil {
				log.Fatalf("使用仓库失败: %v", err)
			}
			fmt.Printf("成功切换到 %s 仓库\n", repoName)
		},
	})

	// add 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "add [repo]",
		Short: "添加指定的仓库",
		Example: "yuv repo add mysql8",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			releasever, err := detector.GetReleasever()
			if err != nil {
				log.Fatalf("获取 releasever 失败: %v", err)
			}
			basearch, err := detector.GetBasearch()
			if err != nil {
				log.Fatalf("获取 basearch 失败: %v", err)
			}

			repoName := args[0]
			if err := repoMgr.Add(repoName, releasever, basearch); err != nil {
				log.Fatalf("添加仓库失败: %v", err)
			}
			fmt.Printf("成功添加 %s 仓库\n", repoName)
		},
	})

	// remove 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "remove [repo]",
		Short: "移除指定的仓库",
		Example: "yuv repo remove mysql8",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoName := args[0]
			if err := repoMgr.Remove(repoName); err != nil {
				log.Fatalf("移除仓库失败: %v", err)
			}
			fmt.Printf("成功移除 %s 仓库\n", repoName)
		},
	})

	// list 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "列出所有仓库",
		Run: func(cmd *cobra.Command, args []string) {
			repos, err := repoMgr.List()
			if err != nil {
				log.Fatalf("列出仓库失败: %v", err)
			}
			fmt.Println("当前仓库:")
			for _, r := range repos {
				fmt.Printf("  - %s\n", r)
			}
		},
	})

	// backup 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "backup",
		Short: "备份所有仓库",
		Run: func(cmd *cobra.Command, args []string) {
			if err := repoMgr.Backup(); err != nil {
				log.Fatalf("备份仓库失败: %v", err)
			}
			fmt.Println("成功备份所有仓库")
		},
	})

	// restore 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "restore",
		Short: "从备份恢复仓库",
		Run: func(cmd *cobra.Command, args []string) {
			if err := repoMgr.Restore(); err != nil {
				log.Fatalf("恢复仓库失败: %v", err)
			}
			fmt.Println("成功从备份恢复仓库")
		},
	})

	// enable 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "enable [repo]",
		Short: "启用指定的仓库",
		Example: "yuv repo enable mysql8",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoName := args[0]
			if err := repoMgr.Enable(repoName); err != nil {
				log.Fatalf("启用仓库失败: %v", err)
			}
			fmt.Printf("成功启用 %s 仓库\n", repoName)
		},
	})

	// disable 命令
	repoCmd.AddCommand(&cobra.Command{
		Use:   "disable [repo]",
		Short: "禁用指定的仓库",
		Example: "yuv repo disable mysql8",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoName := args[0]
			if err := repoMgr.Disable(repoName); err != nil {
				log.Fatalf("禁用仓库失败: %v", err)
			}
			fmt.Printf("成功禁用 %s 仓库\n", repoName)
		},
	})

	// 包管理命令
	installCmd := &cobra.Command{
		Use:   "install [packages...]",
		Short: "安装软件包",
		Example: "yuv install nginx",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Install(args...); err != nil {
				log.Fatalf("安装软件包失败: %v", err)
			}
		},
	}
	installCmd.Flags().BoolP("yes", "y", false, "对所有问题回答是")
	rootCmd.AddCommand(installCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "remove [packages...]",
		Short: "移除软件包",
		Example: "yuv remove nginx",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Remove(args...); err != nil {
				log.Fatalf("移除软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "update [packages...]",
		Short: "更新软件包",
		Example: "yuv update nginx",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Update(args...); err != nil {
				log.Fatalf("更新软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "upgrade",
		Short: "升级系统",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Upgrade(); err != nil {
				log.Fatalf("升级系统失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "erase [packages...]",
		Short: "删除软件包",
		Example: "yuv erase nginx",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Erase(args...); err != nil {
				log.Fatalf("删除软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [pattern]",
		Short: "搜索软件包",
		Example: "yuv search nginx",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Search(args[0]); err != nil {
				log.Fatalf("搜索软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list [args...]",
		Short: "列出软件包",
		Example: "yuv list installed",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.List(args...); err != nil {
				log.Fatalf("列出软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "info [packages...]",
		Short: "显示软件包信息",
		Example: "yuv info nginx",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Info(args...); err != nil {
				log.Fatalf("显示软件包信息失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "清理缓存",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Clean(); err != nil {
				log.Fatalf("清理缓存失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "makecache",
		Short: "生成缓存",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.MakeCache(); err != nil {
				log.Fatalf("生成缓存失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "downgrade [package]",
		Short: "降级软件包",
		Example: "yuv downgrade nginx",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Downgrade(args[0]); err != nil {
				log.Fatalf("降级软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "check-update",
		Short: "检查更新",
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.CheckUpdate(); err != nil {
				log.Fatalf("检查更新失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "provides [file]",
		Short: "查找提供文件的软件包",
		Example: "yuv provides /usr/sbin/nginx",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Provides(args[0]); err != nil {
				log.Fatalf("查找提供文件的软件包失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "deplist [package]",
		Short: "列出软件包依赖",
		Example: "yuv deplist nginx",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := packageMgr.Deplist(args[0]); err != nil {
				log.Fatalf("列出软件包依赖失败: %v", err)
			}
		},
	})

	// 添加 repo 命令组到根命令
	rootCmd.AddCommand(repoCmd)

	// 直接添加中文的 completion 命令，覆盖默认的
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion",
		Short: "生成指定 shell 的自动补全脚本",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("使用方法: yuv completion [shell]")
			fmt.Println("支持的 shell: bash, zsh, fish, powershell")
		},
	})

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("执行命令失败: %v", err)
	}
}
