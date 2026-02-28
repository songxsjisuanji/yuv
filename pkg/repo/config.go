package repo

import (
	"fmt"
	"strings"
)

// RepoType 源类型
type RepoType string

const (
	TypePublic   RepoType = "public"   // 公共镜像源
	TypeThird    RepoType = "third"    // 第三方官方源
	TypeCustom   RepoType = "custom"   // 自定义源
)

// Repo 源配置结构体
type Repo struct {
	Name       string   // 源名称
	Type       RepoType // 源类型
	URL        string   // 源URL
	VaultURL   string   // 过期源URL
	GPGKey     string   // GPG密钥URL
	Enabled    bool     // 是否启用
	Priority   int      // 优先级
	Releasever string   // 发行版版本变量
	Basearch   string   // 架构变量
}

// PublicRepos 预置公共镜像源
var PublicRepos = map[string]*Repo{
	"aliyun": {
		Name:     "aliyun",
		Type:     TypePublic,
		URL:      "https://mirrors.aliyun.com/$distro/$releasever/AppStream/$basearch/os/",
		VaultURL: "https://mirrors.aliyun.com/$distro-vault/$releasever/AppStream/$basearch/os/",
		GPGKey:   "https://mirrors.aliyun.com/$distro/RPM-GPG-KEY-$distro-$releasever",
		Enabled:  true,
		Priority: 1,
	},
}

// ThirdRepos 预置第三方官方源
var ThirdRepos = map[string]*Repo{
	"mysql57": {
		Name:     "mysql57",
		Type:     TypeThird,
		URL:      "https://repo.mysql.com/yum/mysql-5.7-community/el/$releasever/$basearch/",
		GPGKey:   "https://repo.mysql.com/RPM-GPG-KEY-mysql",
		Enabled:  true,
		Priority: 5,
	},
	"redis": {
		Name:     "redis",
		Type:     TypeThird,
		URL:      "https://rpms.remirepo.net/enterprise/$releasever/redis/$basearch/",
		GPGKey:   "https://rpms.remirepo.net/RPM-GPG-KEY-remi",
		Enabled:  true,
		Priority: 5,
	},
	"nginx": {
		Name:     "nginx",
		Type:     TypeThird,
		URL:      "https://nginx.org/packages/centos/$releasever/$basearch/",
		GPGKey:   "https://nginx.org/keys/nginx_signing.key",
		Enabled:  true,
		Priority: 5,
	},
	"docker": {
		Name:     "docker",
		Type:     TypeThird,
		URL:      "https://download.docker.com/linux/centos/$releasever/$basearch/stable",
		GPGKey:   "https://download.docker.com/linux/centos/gpg",
		Enabled:  true,
		Priority: 5,
	},
	"k8s": {
		Name:     "k8s",
		Type:     TypeThird,
		URL:      "https://packages.cloud.google.com/yum/repos/kubernetes-el7-$basearch",
		GPGKey:   "https://packages.cloud.google.com/yum/doc/yum-key.gpg",
		Enabled:  true,
		Priority: 5,
	},
	"php7": {
		Name:     "php7",
		Type:     TypeThird,
		URL:      "https://rpms.remirepo.net/enterprise/$releasever/php74/$basearch/",
		GPGKey:   "https://rpms.remirepo.net/RPM-GPG-KEY-remi",
		Enabled:  true,
		Priority: 5,
	},
	"php8": {
		Name:     "php8",
		Type:     TypeThird,
		URL:      "https://rpms.remirepo.net/enterprise/$releasever/php80/$basearch/",
		GPGKey:   "https://rpms.remirepo.net/RPM-GPG-KEY-remi",
		Enabled:  true,
		Priority: 5,
	},
	"nodejs": {
		Name:     "nodejs",
		Type:     TypeThird,
		URL:      "https://rpm.nodesource.com/pub_16.x/el/$releasever/$basearch/",
		GPGKey:   "https://rpm.nodesource.com/pub/el/NODESOURCE-GPG-SIGNING-KEY-EL",
		Enabled:  true,
		Priority: 5,
	},
}

// GetRepoByName 根据名称获取源配置
func GetRepoByName(name string) (*Repo, error) {
	// 处理 kubernetes 别名
	if name == "kubernetes" {
		name = "k8s"
	}
	// 处理 mysql8 别名
	if name == "mysql8" {
		name = "mysql57"
	}
	// 先查找公共源
	if repo, ok := PublicRepos[name]; ok {
		return repo, nil
	}
	// 再查找第三方源
	if repo, ok := ThirdRepos[name]; ok {
		return repo, nil
	}
	return nil, fmt.Errorf("repo %s not found", name)
}

// GetReposByType 根据类型获取源列表
func GetReposByType(repoType RepoType) map[string]*Repo {
	switch repoType {
	case TypePublic:
		return PublicRepos
	case TypeThird:
		return ThirdRepos
	default:
		return nil
	}
}

// ReplaceVariables 替换源URL中的变量
func (r *Repo) ReplaceVariables(distro, releasever, basearch string) string {
	url := r.URL
	url = strings.ReplaceAll(url, "$distro", distro)
	
	// 对 MySQL 仓库特殊处理，只使用主版本号
	if strings.Contains(r.Name, "mysql") {
		// 提取主版本号（如从 8.9 提取 8）
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			releasever = parts[0]
		}
	}
	
	url = strings.ReplaceAll(url, "$releasever", releasever)
	url = strings.ReplaceAll(url, "$basearch", basearch)
	return url
}

// GetVaultURL 获取过期源URL（替换变量）
func (r *Repo) GetVaultURL(distro, releasever, basearch string) string {
	url := r.VaultURL
	url = strings.ReplaceAll(url, "$distro", distro)
	
	// 对 MySQL 仓库特殊处理，只使用主版本号
	if strings.Contains(r.Name, "mysql") {
		// 提取主版本号（如从 8.9 提取 8）
		parts := strings.Split(releasever, ".")
		if len(parts) > 0 {
			releasever = parts[0]
		}
	}
	
	url = strings.ReplaceAll(url, "$releasever", releasever)
	url = strings.ReplaceAll(url, "$basearch", basearch)
	return url
}

// GetGPGKeyURL 获取GPG密钥URL（替换变量）
func (r *Repo) GetGPGKeyURL(distro, releasever, basearch string) string {
	if r.GPGKey == "" {
		return ""
	}
	gpgKey := r.GPGKey
	gpgKey = strings.ReplaceAll(gpgKey, "$distro", distro)
	gpgKey = strings.ReplaceAll(gpgKey, "$releasever", releasever)
	gpgKey = strings.ReplaceAll(gpgKey, "$basearch", basearch)
	return gpgKey
}
