package system

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"strconv"
)

// Distro 发行版信息
type Distro struct {
	Name    string // 发行版名称
	Version string // 发行版版本
	Arch    string // 系统架构
}

// Detector 系统检测器
type Detector struct {}

// NewDetector 创建系统检测器实例
func NewDetector() *Detector {
	return &Detector{}
}

// Detect 检测系统信息
func (d *Detector) Detect() (*Distro, error) {
	distro, err := d.detectDistro()
	if err != nil {
		return nil, err
	}

	arch, err := d.detectArch()
	if err != nil {
		return nil, err
	}

	return &Distro{
		Name:    distro.Name,
		Version: distro.Version,
		Arch:    arch,
	}, nil
}

// detectDistro 检测发行版信息
func (d *Detector) detectDistro() (*Distro, error) {
	distro := &Distro{}

	// 优先从 /etc/centos-release 文件获取版本信息，以支持 CentOS 7.9.2009 这样的格式
	if err := d.detectVersionFromOtherFiles(distro); err == nil {
		// 规范化发行版名称
		distro.Name = d.normalizeDistroName(distro.Name)
		return distro, nil
	}

	// 如果从 /etc/centos-release 获取失败，尝试从 /etc/os-release 文件获取
	content, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("read os-release failed: %v", err)
	}

	// 解析文件内容
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "NAME=") {
			distro.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			distro.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		}
	}

	// 规范化发行版名称
	distro.Name = d.normalizeDistroName(distro.Name)

	// 如果没有获取到版本号，返回错误
	if distro.Version == "" {
		return nil, fmt.Errorf("detect version failed")
	}

	return distro, nil
}

// detectVersionFromOtherFiles 从其他文件检测版本号
func (d *Detector) detectVersionFromOtherFiles(distro *Distro) error {
	// 尝试从 /etc/centos-release 文件获取，支持 CentOS 7.9.2009 这样的格式
	if content, err := ioutil.ReadFile("/etc/centos-release"); err == nil {
		// 匹配格式如 "CentOS Linux release 7.9.2009 (Core)" 或 "CentOS Stream release 8" 的版本号
		if match := regexp.MustCompile(`CentOS.*release\s+([0-9.]+(?:\s+\(Core\))?)`).FindStringSubmatch(string(content)); len(match) > 1 {
			// 提取版本号，去除括号部分
			version := strings.TrimSpace(strings.Replace(match[1], "(Core)", "", -1))
			distro.Version = version
			distro.Name = "CentOS"
			return nil
		}
	}

	// 尝试从 /etc/redhat-release 文件获取
	if content, err := ioutil.ReadFile("/etc/redhat-release"); err == nil {
		if match := regexp.MustCompile(`.*release\s+([0-9.]+)`).FindStringSubmatch(string(content)); len(match) > 1 {
			distro.Version = match[1]
			return nil
		}
	}

	return fmt.Errorf("detect version failed")
}

// detectArch 检测系统架构
func (d *Detector) detectArch() (string, error) {
	// 读取 /proc/version 文件
	content, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return "", fmt.Errorf("read proc/version failed: %v", err)
	}

	// 解析架构信息
	if strings.Contains(string(content), "x86_64") {
		return "x86_64", nil
	} else if strings.Contains(string(content), "i686") {
		return "i686", nil
	} else if strings.Contains(string(content), "aarch64") {
		return "aarch64", nil
	} else if strings.Contains(string(content), "armv7l") {
		return "armv7l", nil
	}

	return "", fmt.Errorf("detect arch failed")
}

// normalizeDistroName 规范化发行版名称
func (d *Detector) normalizeDistroName(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "centos"):
		return "centos"
	case strings.Contains(name, "rocky"):
		return "rockylinux"
	case strings.Contains(name, "alma"):
		return "almalinux"
	case strings.Contains(name, "fedora"):
		return "fedora"
	case strings.Contains(name, "rhel") || strings.Contains(name, "red hat"):
		return "rhel"
	default:
		return name
	}
}

// GetReleasever 获取 releasever 变量
func (d *Detector) GetReleasever() (string, error) {
	distro, err := d.Detect()
	if err != nil {
		return "", err
	}

	// 返回完整的版本号
	return distro.Version, nil
}

// GetMajorReleasever 获取主版本号
func (d *Detector) GetMajorReleasever() (string, error) {
	distro, err := d.Detect()
	if err != nil {
		return "", err
	}

	// 提取主版本号
	parts := strings.Split(distro.Version, ".")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("get major releasever failed")
}

// GetBasearch 获取 basearch 变量
func (d *Detector) GetBasearch() (string, error) {
	distro, err := d.Detect()
	if err != nil {
		return "", err
	}

	return distro.Arch, nil
}

// IsSupported 检查是否支持当前发行版
func (d *Detector) IsSupported() (bool, error) {
	distro, err := d.Detect()
	if err != nil {
		return false, err
	}

	supportedDistros := []string{"centos", "rockylinux", "almalinux", "fedora", "rhel"}
	for _, supported := range supportedDistros {
		if distro.Name == supported {
			return true, nil
		}
	}

	return false, nil
}

// IsVersionExpired 检查版本是否过期
func (d *Detector) IsVersionExpired() (bool, error) {
	distro, err := d.Detect()
	if err != nil {
		return false, err
	}

	// 获取主版本号
	parts := strings.Split(distro.Version, ".")
	if len(parts) == 0 {
		return false, fmt.Errorf("invalid version format: %s", distro.Version)
	}

	majorVersion, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// 定义各发行版的当前支持版本
	supportedVersions := map[string]int{
		"centos":     9,    // 当前支持 CentOS 9
		"rockylinux": 9,    // 当前支持 Rocky Linux 9
		"almalinux":  9,    // 当前支持 AlmaLinux 9
		"fedora":     40,   // 当前支持 Fedora 40
		"rhel":       9,    // 当前支持 RHEL 9
	}

	// 检查主版本号是否低于当前支持版本
	if supportedVersion, ok := supportedVersions[distro.Name]; ok {
		return majorVersion < supportedVersion, nil
	}

	// 对于未知发行版，默认认为版本未过期
	return false, nil
}

// GetDistroName 获取规范化的发行版名称
func (d *Detector) GetDistroName() (string, error) {
	distro, err := d.Detect()
	if err != nil {
		return "", err
	}

	return distro.Name, nil
}
