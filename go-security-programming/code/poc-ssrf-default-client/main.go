// Package main demonstrates SSRF via default http.Client + user-controlled URL.
//
// PoC E6: "http.Get(userURL) 没有任何信任边界。内网、云 metadata、loopback 全都能访问。"
//
// Run: go run main.go
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// --- 漏洞代码：常见的"图片代理"/"URL 预览"业务 ---

// fetchPreview 看起来完全合理：业务让用户贴 URL，服务端抓取预览。
// 问题：默认 http.Client 对任何 URL 一视同仁——
// 公网、内网（10.x/172.16.x/192.168.x）、loopback（127.0.0.1）、
// 云厂商 metadata（169.254.169.254）全都能访问。
func fetchPreview(userURL string) (string, error) {
	resp, err := http.Get(userURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return string(body), nil
}

// --- 修复版本：预检 URL 的 host ---

// fetchPreviewSafe 在请求前把主机名解析成 IP，拒绝私有段/loopback/link-local。
// 这只是最基本的防线——生产环境还要处理 DNS rebinding、重定向链等。
func fetchPreviewSafe(userURL string) (string, error) {
	u, err := url.Parse(userURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return "", fmt.Errorf("refused: %s -> %s is private/loopback", host, ip)
		}
	}

	// 另外还要禁止 http.Client 自动跟随 302 重定向到不安全地址
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", userURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return string(body), nil
}

func main() {
	// 模拟一个"云 metadata"服务监听 loopback
	// 真实的云 metadata 地址是 169.254.169.254，但为了在任何环境能跑，
	// 这里用本地 httptest 伪装一个"敏感内网服务"。
	fakeMetadata := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"instance-role":"web-prod","aws_access_key":"AKIA...EXPOSED",`+
			`"secret":"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}`)
	}))
	defer fakeMetadata.Close()

	metadataURL := fakeMetadata.URL // 形如 http://127.0.0.1:XXXXX

	fmt.Println("=== 漏洞版本：fetchPreview 默认信任任何 URL ===")
	body, err := fetchPreview(metadataURL)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  泄漏内容: %s", strings.TrimSpace(body))
		fmt.Println()
	}
	fmt.Println()

	fmt.Println("=== 修复版本：fetchPreviewSafe 拒绝内网/loopback ===")
	_, err = fetchPreviewSafe(metadataURL)
	if err != nil {
		fmt.Printf("  按预期拒绝: %v\n", err)
	} else {
		fmt.Println("  ⚠️ 意外：本应拒绝却放行了")
	}
	fmt.Println()

	fmt.Println("=== 关键观察 ===")
	fmt.Println("1. Go 标准库 http.Client 对 URL 的 host 不做任何安全过滤")
	fmt.Println("2. http.Get(userInput) 在业务代码里非常常见——URL 预览、图片代理、Webhook")
	fmt.Println("3. 真实威胁目标：")
	fmt.Println("   - AWS/GCP/Alibaba 云 metadata (169.254.169.254) → 临时凭证")
	fmt.Println("   - 内网服务（Redis/ES/Consul 没认证的管理接口）")
	fmt.Println("   - 本机 loopback（file://、gopher://、dict:// 如果库支持）")
	fmt.Println("4. 防御要点：IP 白名单 + 自定义 DialContext + 禁止重定向 + 处理 DNS rebinding")
}
