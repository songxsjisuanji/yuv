package repo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"yuv/pkg/system"
)

const (
	// RepoDir YUM源配置目录
	RepoDir = "/etc/yum.repos.d"
	// BackupDir 备份目录
	BackupDir = "/etc/yum.repos.d.bak"
)

// Manager 源管理器
type Manager struct {
	RepoDir   string
	BackupDir string
}

// NewManager 创建源管理器实例
func NewManager() *Manager {
	return &Manager{
		RepoDir:   RepoDir,
		BackupDir: BackupDir,
	}
}

// Backup 备份所有源配置文件
func (m *Manager) Backup() error {
	// 创建备份目录
	if err := os.MkdirAll(m.BackupDir, 0755); err != nil {
		return fmt.Errorf("create backup directory failed: %v", err)
	}

	// 遍历源配置文件
	files, err := ioutil.ReadDir(m.RepoDir)
	if err != nil {
		return fmt.Errorf("read repo directory failed: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".repo") {
			continue
		}

		// 移动文件到备份目录
		src := filepath.Join(m.RepoDir, file.Name())
		dst := filepath.Join(m.BackupDir, file.Name())
		
		// 如果备份目录中已存在同名文件，先删除
		if _, err := os.Stat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("remove existing backup file failed: %v", err)
			}
		}
		
		// 移动文件
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("move repo file to backup failed: %v", err)
		}
	}

	return nil
}

// Restore 恢复源配置文件
func (m *Manager) Restore() error {
	// 检查备份目录是否存在
	if _, err := os.Stat(m.BackupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory not found")
	}

	// 遍历备份文件
	files, err := ioutil.ReadDir(m.BackupDir)
	if err != nil {
		return fmt.Errorf("read backup directory failed: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".repo") {
			continue
		}

		// 复制文件到源配置目录
		src := filepath.Join(m.BackupDir, file.Name())
		dst := filepath.Join(m.RepoDir, file.Name())
		content, err := ioutil.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read backup file failed: %v", err)
		}
		if err := ioutil.WriteFile(dst, content, file.Mode()); err != nil {
			return fmt.Errorf("write repo file failed: %v", err)
		}
	}

	return nil
}

// Use 使用指定的公共镜像源
func (m *Manager) Use(repoName, releasever, basearch string) error {
	// 备份当前源
	if err := m.Backup(); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	// 获取源配置
	repo, err := GetRepoByName(repoName)
	if err != nil {
		return err
	}

	// 清理当前源配置
	if err := m.cleanRepoDir(); err != nil {
		return fmt.Errorf("clean repo directory failed: %v", err)
	}

	// 创建系统检测器实例
	detector := system.NewDetector()

	// 获取发行版名称
	distro, err := detector.GetDistroName()
	if err != nil {
		return fmt.Errorf("get distro name failed: %v", err)
	}

	// 对于 CentOS 系统，使用阿里云的预配置 repo 文件
	if distro == "centos" {
		// 提取主版本号
		parts := strings.Split(releasever, ".")
		if len(parts) == 0 {
			return fmt.Errorf("invalid version format: %s", releasever)
		}
		majorVersion := parts[0]

		// 根据主版本号选择不同的 repo 文件
		var repoURL string
		if majorVersion == "7" {
			repoURL = "https://mirrors.aliyun.com/repo/Centos-7.repo"
		} else if majorVersion == "8" {
			repoURL = "https://mirrors.aliyun.com/repo/Centos-vault-8.5.2111.repo"
		} else {
			// 对于其他版本，使用默认的生成方式
			goto defaultRepo
		}

		// 下载 repo 文件
		repoFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s.repo", repoName))
		resp, err := http.Get(repoURL)
		if err != nil {
			return fmt.Errorf("download repo file failed: %v", err)
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read repo file failed: %v", err)
		}

		if err := ioutil.WriteFile(repoFile, content, 0644); err != nil {
			return fmt.Errorf("write repo file failed: %v", err)
		}

		return nil
	}

defaultRepo:
	// 检测版本是否过期
	expired, err := detector.IsVersionExpired()
	if err != nil {
		return fmt.Errorf("check version expired failed: %v", err)
	}

	// 为 Rocky Linux 和 AlmaLinux 添加多个源配置
	if distro == "rockylinux" || distro == "almalinux" {
		// 创建 BaseOS 源
		baseosContent, err := m.generateRepoContentForRepoType(repo, "BaseOS", releasever, basearch, expired)
		if err != nil {
			return err
		}
		baseosFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s-baseos.repo", repoName))
		if err := ioutil.WriteFile(baseosFile, []byte(baseosContent), 0644); err != nil {
			return fmt.Errorf("write baseos repo file failed: %v", err)
		}

		// 创建 AppStream 源
		appstreamContent, err := m.generateRepoContentForRepoType(repo, "AppStream", releasever, basearch, expired)
		if err != nil {
			return err
		}
		appstreamFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s-appstream.repo", repoName))
		if err := ioutil.WriteFile(appstreamFile, []byte(appstreamContent), 0644); err != nil {
			return fmt.Errorf("write appstream repo file failed: %v", err)
		}
	} else {
		// 对于其他发行版，使用单一源配置
		repoFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s.repo", repoName))
		content, err := m.generateRepoContent(repo, releasever, basearch)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(repoFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("write repo file failed: %v", err)
		}
	}

	return nil
}

// generateRepoContentForRepoType 为指定类型的源生成配置内容
func (m *Manager) generateRepoContentForRepoType(repo *Repo, repoType, releasever, basearch string, expired bool) (string, error) {
	// 创建系统检测器实例
	detector := system.NewDetector()

	// 获取发行版名称
	distro, err := detector.GetDistroName()
	if err != nil {
		return "", fmt.Errorf("get distro name failed: %v", err)
	}

	// 根据过期状态选择正确的 URL
	var url string
	if expired && repo.VaultURL != "" {
		url = strings.ReplaceAll(repo.VaultURL, "$distro", distro)
		url = strings.ReplaceAll(url, "$releasever", releasever)
		url = strings.ReplaceAll(url, "$basearch", basearch)
		url = strings.ReplaceAll(url, "AppStream", repoType)
	} else {
		url = strings.ReplaceAll(repo.URL, "$distro", distro)
		url = strings.ReplaceAll(url, "$releasever", releasever)
		url = strings.ReplaceAll(url, "$basearch", basearch)
		url = strings.ReplaceAll(url, "AppStream", repoType)
	}

	// 对于 CentOS 7，调整 URL 格式，移除 AppStream 部分
	if distro == "centos" {
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			majorVersion := parts[0]
			if majorVersion == "7" {
				// CentOS 7 使用 /os/x86_64/ 格式
				url = strings.Replace(url, "/AppStream/x86_64/os/", "/os/x86_64/", -1)
				url = strings.Replace(url, "/BaseOS/x86_64/os/", "/os/x86_64/", -1)
			}
		}
	}

	// 获取 GPG 密钥 URL
	gpgKey := strings.ReplaceAll(repo.GPGKey, "$distro", distro)
	gpgKey = strings.ReplaceAll(gpgKey, "$releasever", releasever)
	gpgKey = strings.ReplaceAll(gpgKey, "$basearch", basearch)

	// 生成源配置内容
	return fmt.Sprintf(`[%s-%s]
name=%s %s Repository
baseurl=%s
enabled=1
gpgcheck=1
gpgkey=%s
priority=%d
`,
		repo.Name, repoType, repo.Name, repoType, url, gpgKey, repo.Priority), nil
}

// Add 添加指定的源
func (m *Manager) Add(repoName, releasever, basearch string) error {
	// 获取源配置
	repo, err := GetRepoByName(repoName)
	if err != nil {
		return err
	}

	// 对 MySQL 仓库使用 RPM 包安装的方式
	if strings.Contains(repoName, "mysql") {
		// 提取主版本号和小版本号
		var rpmURL string
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			elVersion := parts[0]
			// 默认为9，如果有小版本号则使用小版本号
			subVersion := "9"
			if len(parts) > 1 {
				subVersion = parts[1]
			}
			// 构建 RPM 包 URL
			if strings.Contains(repoName, "57") {
				rpmURL = fmt.Sprintf("https://repo.mysql.com/mysql57-community-release-el%s-%s.noarch.rpm", elVersion, subVersion)
			} else {
				rpmURL = fmt.Sprintf("https://repo.mysql.com/mysql80-community-release-el%s-%s.noarch.rpm", elVersion, subVersion)
			}
		} else {
			return fmt.Errorf("invalid releasever format: %s", releasever)
		}

		// 禁用 MySQL 模块
		cmd := exec.Command("yum", "module", "-y", "disable", "mysql")
		if err := cmd.Run(); err != nil {
			// 忽略错误，因为在某些系统上可能没有 mysql 模块
		}

		// 下载并安装 RPM 包
		rpmFile := filepath.Join("/tmp", fmt.Sprintf("%s-release.rpm", repoName))
		cmd = exec.Command("curl", "-o", rpmFile, rpmURL)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("download rpm failed: %v", err)
		}

		// 尝试安装 RPM 包，如果已安装则更新
		cmd = exec.Command("rpm", "-Uvh", "--force", rpmFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("install rpm failed: %v", err)
		}

		// 清理临时文件
		os.Remove(rpmFile)

		return nil
	}

	// 对 Docker 仓库使用 yum-config-manager 安装方式（除了 AlmaLinux 和 Fedora）
	if repoName == "docker" {
		// 创建系统检测器实例
		detector := system.NewDetector()
		// 获取发行版名称
		distro, err := detector.GetDistroName()
		if err != nil {
			return fmt.Errorf("get distro name failed: %v", err)
		}

		// 排除 AlmaLinux 和 Fedora
		if distro != "almalinux" && distro != "fedora" {
			// 检查并安装 yum-utils
			_, err := exec.LookPath("yum-config-manager")
			if err != nil {
				// 安装 yum-utils
				cmd := exec.Command("yum", "install", "-y", "yum-utils")
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("install yum-utils failed: %v", err)
				}
			}

			// 使用 yum-config-manager 添加阿里云 Docker 源
			cmd := exec.Command("yum-config-manager", "--add-repo", "https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("add docker repo failed: %v", err)
			}

			return nil
		}
	}

	// 对 Kubernetes 仓库使用阿里云源（除了 AlmaLinux 和 Fedora）
	if repoName == "kubernetes" {
		// 创建系统检测器实例
		detector := system.NewDetector()
		// 获取发行版名称
		distro, err := detector.GetDistroName()
		if err != nil {
			return fmt.Errorf("get distro name failed: %v", err)
		}

		// 排除 AlmaLinux 和 Fedora
		if distro != "almalinux" && distro != "fedora" {
			// 构建 Kubernetes 源配置
			repoFile := filepath.Join(m.RepoDir, "kubernetes.repo")
			repoContent := `[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/
enabled=1
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes-new/core/stable/v1.28/rpm/repodata/repomd.xml.key
`

			// 写入配置文件
			if err := ioutil.WriteFile(repoFile, []byte(repoContent), 0644); err != nil {
				return fmt.Errorf("write kubernetes repo file failed: %v", err)
			}

			return nil
		}
	}

	// 其他仓库使用传统方式
	repoFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s.repo", repoName))
	content, err := m.generateRepoContent(repo, releasever, basearch)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(repoFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write repo file failed: %v", err)
	}

	return nil
}

// Remove 删除指定的源
func (m *Manager) Remove(repoName string) error {
	repoFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s.repo", repoName))
	if err := os.Remove(repoFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo %s not found", repoName)
		}
		return fmt.Errorf("remove repo file failed: %v", err)
	}

	return nil
}

// List 列出所有源
func (m *Manager) List() ([]string, error) {
	files, err := ioutil.ReadDir(m.RepoDir)
	if err != nil {
		return nil, fmt.Errorf("read repo directory failed: %v", err)
	}

	var repos []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".repo") {
			repos = append(repos, strings.TrimSuffix(file.Name(), ".repo"))
		}
	}

	return repos, nil
}

// Enable 启用指定的源
func (m *Manager) Enable(repoName string) error {
	return m.setRepoEnabled(repoName, true)
}

// Disable 禁用指定的源
func (m *Manager) Disable(repoName string) error {
	return m.setRepoEnabled(repoName, false)
}

// setRepoEnabled 设置源的启用状态
func (m *Manager) setRepoEnabled(repoName string, enabled bool) error {
	repoFile := filepath.Join(m.RepoDir, fmt.Sprintf("%s.repo", repoName))
	content, err := ioutil.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("read repo file failed: %v", err)
	}

	// 替换 enabled 配置
	newContent := string(content)
	if enabled {
		newContent = strings.Replace(newContent, "enabled=0", "enabled=1", -1)
	} else {
		newContent = strings.Replace(newContent, "enabled=1", "enabled=0", -1)
	}

	// 写入文件
	if err := ioutil.WriteFile(repoFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write repo file failed: %v", err)
	}

	return nil
}

// cleanRepoDir 清理源配置目录
func (m *Manager) cleanRepoDir() error {
	files, err := ioutil.ReadDir(m.RepoDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".repo") {
			if err := os.Remove(filepath.Join(m.RepoDir, file.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateRepoContent 生成源配置文件内容
func (m *Manager) generateRepoContent(repo *Repo, releasever, basearch string) (string, error) {
	// 创建系统检测器实例
	detector := system.NewDetector()

	// 获取发行版名称
	distro, err := detector.GetDistroName()
	if err != nil {
		return "", fmt.Errorf("get distro name failed: %v", err)
	}

	// 检测版本是否过期
	expired, err := detector.IsVersionExpired()
	if err != nil {
		return "", fmt.Errorf("check version expired failed: %v", err)
	}

	// 对 MySQL 仓库特殊处理，只使用主版本号
	mysqlReleasever := releasever
	if strings.Contains(repo.Name, "mysql") {
		// 提取主版本号（如从 8.9 提取 8）
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			mysqlReleasever = parts[0]
		}
	}

	// 根据过期状态选择正确的 URL
	var url string
	if expired && repo.VaultURL != "" {
		// 对于非 MySQL 仓库，使用完整的版本号，以支持 CentOS 7.9.2009 这样的格式
		if strings.Contains(repo.Name, "mysql") {
			url = repo.GetVaultURL(distro, mysqlReleasever, basearch)
		} else {
			url = repo.GetVaultURL(distro, releasever, basearch)
		}
	} else {
		// 对于非 MySQL 仓库，使用完整的版本号
		if strings.Contains(repo.Name, "mysql") {
			url = repo.ReplaceVariables(distro, mysqlReleasever, basearch)
		} else {
			url = repo.ReplaceVariables(distro, releasever, basearch)
		}
	}
	
	// 对于 CentOS 7，调整 URL 格式，移除 AppStream 部分
	if distro == "centos" {
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			majorVersion := parts[0]
			if majorVersion == "7" {
				// CentOS 7 使用 /os/x86_64/ 格式
				url = strings.Replace(url, "/AppStream/x86_64/os/", "/os/x86_64/", -1)
			}
		}
	}

	// 获取 GPG 密钥 URL
	gpgKey := repo.GetGPGKeyURL(distro, releasever, basearch)

	return fmt.Sprintf(`[%s]
name=%s Repository
baseurl=%s
enabled=%d
gpgcheck=1
gpgkey=%s
priority=%d
`,
		repo.Name, repo.Name, url, boolToInt(repo.Enabled), gpgKey, repo.Priority), nil
}

// boolToInt 将布尔值转换为整数
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
