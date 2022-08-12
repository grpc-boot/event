package base

import (
	"fmt"
	"testing"
)

func TestParams_GetString(t *testing.T) {
	var p = Params{
		"int":     45,
		"int64":   int64(45),
		"int32":   int32(45),
		"int16":   int16(45),
		"int8":    int8(45),
		"uint64":  uint64(45),
		"uint32":  uint32(45),
		"uint16":  uint16(45),
		"uint8":   uint8(45),
		"uint":    uint(45),
		"float64": float64(45),
		"float32": float32(45),
		"bytes":   []byte("45"),
		"string":  "45",
	}

	for key, _ := range p {
		val := p.GetString(key)
		if val != "45" {
			t.Fatalf("key: %s want 45, got %s\n", key, val)
		}
	}

	p["true"] = true

	val := p.GetString("true")
	if val != "true" {
		t.Fatalf("key: %s want true, got %s\n", "true", val)
	}

	p["false"] = false
	val = p.GetString("false")
	if val != "false" {
		t.Fatalf("key: %s want false, got %s\n", "false", val)
	}

	p["float64"] = 45.45435142134234
	val = p.GetString("float64")
	if val != "45.45435142134234" {
		t.Fatalf("key: %s want 45.45435142134234, got %s\n", "float64", val)
	}
}

func TestParams_GetFloat64(t *testing.T) {
	var p = Params{
		"int":     45,
		"int64":   int64(45),
		"int32":   int32(45),
		"int16":   int16(45),
		"int8":    int8(45),
		"uint64":  uint64(45),
		"uint32":  uint32(45),
		"uint16":  uint16(45),
		"uint8":   uint8(45),
		"uint":    uint(45),
		"float64": float64(45),
		"float32": float32(45),
		"bytes":   []byte("45"),
		"string":  "45",
	}

	for key, _ := range p {
		val := p.GetFloat64(key)
		if val != 45 {
			t.Fatalf("key: %s want 45, got %s\n", key, fmt.Sprint(val))
		}
	}
}

func TestParams_GetInt64(t *testing.T) {
	var p = Params{
		"int":     45,
		"int64":   int64(45),
		"int32":   int32(45),
		"int16":   int16(45),
		"int8":    int8(45),
		"uint64":  uint64(45),
		"uint32":  uint32(45),
		"uint16":  uint16(45),
		"uint8":   uint8(45),
		"uint":    uint(45),
		"float64": float64(45),
		"float32": float32(45),
		"bytes":   []byte("45"),
		"string":  "45",
	}

	for key, _ := range p {
		val := p.GetInt64(key)
		if val != 45 {
			t.Fatalf("key: %s want 45, got %d\n", key, val)
		}
	}
}
