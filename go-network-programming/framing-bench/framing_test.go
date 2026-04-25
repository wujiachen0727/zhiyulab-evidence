// 三种分帧策略 benchmark 对比
// length-prefix vs delimiter vs fixed-length
// [实测 Go 1.26.2 darwin/arm64]
package framing_bench

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// === Length-Prefix 编码/解码 ===

func encodeLengthPrefix(msg []byte) []byte {
	buf := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(msg)))
	copy(buf[4:], msg)
	return buf
}

func decodeLengthPrefix(data []byte) ([]byte, int) {
	if len(data) < 4 {
		return nil, 0
	}
	msgLen := int(binary.BigEndian.Uint32(data[:4]))
	if len(data) < 4+msgLen {
		return nil, 0
	}
	return data[4 : 4+msgLen], 4 + msgLen
}

// === Delimiter 编码/解码 ===

var delimiter = []byte("\n")

func encodeDelimiter(msg []byte) []byte {
	buf := make([]byte, len(msg)+1)
	copy(buf, msg)
	buf[len(msg)] = '\n'
	return buf
}

func decodeDelimiter(data []byte) ([]byte, int) {
	idx := bytes.IndexByte(data, '\n')
	if idx == -1 {
		return nil, 0
	}
	return data[:idx], idx + 1
}

// === Fixed-Length 编码/解码 ===

const fixedLen = 128

func encodeFixedLength(msg []byte) []byte {
	buf := make([]byte, fixedLen)
	copy(buf, msg)
	return buf
}

func decodeFixedLength(data []byte) ([]byte, int) {
	if len(data) < fixedLen {
		return nil, 0
	}
	// 找到实际内容的末尾（去除填充的零字节）
	end := fixedLen
	for end > 0 && data[end-1] == 0 {
		end--
	}
	return data[:end], fixedLen
}

// === Benchmark ===

var testMsg = []byte("hello world, this is a test message for framing benchmark")

// sink 变量防止编译器 DCE 优化
var sinkBytes []byte
var sinkInt int

func BenchmarkEncodeLengthPrefix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBytes = encodeLengthPrefix(testMsg)
	}
}

func BenchmarkEncodeDelimiter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBytes = encodeDelimiter(testMsg)
	}
}

func BenchmarkEncodeFixedLength(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBytes = encodeFixedLength(testMsg)
	}
}

func BenchmarkDecodeLengthPrefix(b *testing.B) {
	encoded := encodeLengthPrefix(testMsg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkBytes, sinkInt = decodeLengthPrefix(encoded)
	}
}

func BenchmarkDecodeDelimiter(b *testing.B) {
	encoded := encodeDelimiter(testMsg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkBytes, sinkInt = decodeDelimiter(encoded)
	}
}

func BenchmarkDecodeFixedLength(b *testing.B) {
	encoded := encodeFixedLength(testMsg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkBytes, sinkInt = decodeFixedLength(encoded)
	}
}
