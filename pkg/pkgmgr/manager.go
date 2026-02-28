package pkgmgr

import (
	"os"
	"os/exec"
)

// Manager 包管理器
type Manager struct {
	UseYum bool // 是否使用 yum 命令
}

// NewManager 创建包管理器实例
func NewManager() *Manager {
	// 检测系统是否有 dnf 命令
	_, err := exec.LookPath("dnf")
	useYum := err != nil
	return &Manager{
		UseYum: useYum,
	}
}

// getCommand 获取包管理命令
func (m *Manager) getCommand() string {
	if m.UseYum {
		return "yum"
	}
	return "dnf"
}

// Install 安装包
func (m *Manager) Install(packages ...string) error {
	// 检查是否已经包含 -y 标志
	hasYFlag := false
	for _, pkg := range packages {
		if pkg == "-y" {
			hasYFlag = true
			break
		}
	}

	// 构建命令参数
	args := []string{"install"}
	if !hasYFlag {
		args = append(args, "-y")
	}
	args = append(args, packages...)

	cmd := exec.Command(m.getCommand(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Remove 卸载包
func (m *Manager) Remove(packages ...string) error {
	cmd := exec.Command(m.getCommand(), append([]string{"remove", "-y"}, packages...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Update 更新包
func (m *Manager) Update(packages ...string) error {
	args := []string{"update", "-y"}
	if len(packages) > 0 {
		args = append(args, packages...)
	}
	cmd := exec.Command(m.getCommand(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Upgrade 升级系统
func (m *Manager) Upgrade() error {
	cmd := exec.Command(m.getCommand(), "upgrade", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Erase 彻底卸载包
func (m *Manager) Erase(packages ...string) error {
	cmd := exec.Command(m.getCommand(), append([]string{"erase", "-y"}, packages...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Search 搜索包
func (m *Manager) Search(pattern string) error {
	cmd := exec.Command(m.getCommand(), "search", pattern)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// List 列出包
func (m *Manager) List(args ...string) error {
	cmd := exec.Command(m.getCommand(), append([]string{"list"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Info 查看包信息
func (m *Manager) Info(packages ...string) error {
	cmd := exec.Command(m.getCommand(), append([]string{"info"}, packages...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean 清理缓存
func (m *Manager) Clean() error {
	cmd := exec.Command(m.getCommand(), "clean", "all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// MakeCache 生成缓存
func (m *Manager) MakeCache() error {
	cmd := exec.Command(m.getCommand(), "makecache")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Downgrade 降级包
func (m *Manager) Downgrade(packageName string) error {
	cmd := exec.Command(m.getCommand(), "downgrade", "-y", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CheckUpdate 检查更新
func (m *Manager) CheckUpdate() error {
	cmd := exec.Command(m.getCommand(), "check-update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Provides 查找提供指定文件的包
func (m *Manager) Provides(filePath string) error {
	cmd := exec.Command(m.getCommand(), "provides", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WhatProvides 查找提供指定功能的包
func (m *Manager) WhatProvides(feature string) error {
	cmd := exec.Command(m.getCommand(), "whatprovides", feature)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Deplist 查看包依赖
func (m *Manager) Deplist(packageName string) error {
	cmd := exec.Command(m.getCommand(), "deplist", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// History 查看历史
func (m *Manager) History(args ...string) error {
	cmd := exec.Command(m.getCommand(), append([]string{"history"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
