// Package main demonstrates the classic JWT alg:none vulnerability
// when verification code does not assert the algorithm.
//
// PoC E8: "JWT 库给你 API 让你验签，但如果不校验 alg，攻击者塞个 alg:none 就过关。"
//
// Run: go run main.go
package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// JWTHeader 简化的 JWT header
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTClaims 简化的 JWT payload
type JWTClaims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

// --- 漏洞代码：一个常见的 JWT 验证函数 ---

// verifyJWTUnsafe 看起来完整：解码 header、解码 payload、验签。
// 问题：**它没有声明"我接受哪些 alg"**——信任了 header 里传进来的 alg 值。
// 如果 header.Alg == "none"，这段代码直接相信并返回 claims。
// 这个模式在野外出现过非常多次，OWASP 专门给了条目。
func verifyJWTUnsafe(token string, secret []byte) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}

	// 这里是漏洞点：根据 header.Alg 选择验签方式
	// 开发者可能觉得"支持 alg:none 是协议允许的"——协议确实允许，
	// 但业务绝不该接受。
	switch header.Alg {
	case "HS256":
		// 正常 HMAC 验签流程（简化）
		if !verifyHMAC(parts[0]+"."+parts[1], parts[2], secret) {
			return nil, errors.New("HMAC signature invalid")
		}
	case "none":
		// "没签名就不用验"——最常见的坑
		// 不 return error，直接继续
	default:
		return nil, fmt.Errorf("unsupported alg: %s", header.Alg)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

// --- 修复版本：显式 allowlist ---

func verifyJWTSafe(token string, secret []byte, allowedAlgs []string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}

	// 显式白名单：不在列表里直接拒绝
	allowed := false
	for _, a := range allowedAlgs {
		if header.Alg == a {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("alg %s not allowed", header.Alg)
	}

	if header.Alg != "HS256" {
		return nil, errors.New("only HS256 supported")
	}
	if !verifyHMAC(parts[0]+"."+parts[1], parts[2], secret) {
		return nil, errors.New("HMAC signature invalid")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

// verifyHMAC 占位，真实实现用 crypto/hmac
func verifyHMAC(_ string, _ string, _ []byte) bool {
	// 模拟实现：假设签名校验通过
	return true
}

// --- 构造攻击 token ---

func buildUnsignedToken(sub, role string) string {
	h := JWTHeader{Alg: "none", Typ: "JWT"}
	c := JWTClaims{Sub: sub, Role: role, Exp: 9999999999}
	hB, _ := json.Marshal(h)
	cB, _ := json.Marshal(c)
	h64 := base64.RawURLEncoding.EncodeToString(hB)
	c64 := base64.RawURLEncoding.EncodeToString(cB)
	// 第三段留空——攻击 token 没有签名
	return h64 + "." + c64 + "."
}

func main() {
	secret := []byte("server-secret-used-for-hs256")

	// 攻击者构造一个声称自己是 admin 的 token，不签名
	evilToken := buildUnsignedToken("alice", "admin")
	fmt.Println("=== 攻击 token（alg:none，未签名） ===")
	fmt.Println(evilToken)
	fmt.Println()

	fmt.Println("=== 漏洞版本：verifyJWTUnsafe 接受 alg:none ===")
	claims, err := verifyJWTUnsafe(evilToken, secret)
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Printf("  ✗ 被接受！sub=%s, role=%s（攻击者获得 admin 权限）\n",
			claims.Sub, claims.Role)
	}
	fmt.Println()

	fmt.Println("=== 修复版本：verifyJWTSafe 白名单拒绝 ===")
	_, err = verifyJWTSafe(evilToken, secret, []string{"HS256"})
	if err != nil {
		fmt.Printf("  ✓ 按预期拒绝: %v\n", err)
	} else {
		fmt.Println("  ⚠️ 意外放行")
	}
	fmt.Println()

	fmt.Println("=== 关键观察 ===")
	fmt.Println("1. JWT 规范本身允许 alg:none——这是协议层面的决定")
	fmt.Println("2. 多数 JWT 库默认不接受 none，但**手写验证代码**非常容易踩坑")
	fmt.Println("3. 核心防线：永远用白名单（allowedAlgs），不信任 header 里的 alg 值")
	fmt.Println("4. 历史案例：2015-2018 年间多个主流 JWT 库爆出 alg-confusion 漏洞")
}
