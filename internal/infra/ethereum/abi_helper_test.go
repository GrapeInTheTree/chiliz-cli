package ethereum

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func TestParseCallSignature(t *testing.T) {
	tests := []struct {
		name       string
		sig        string
		wantInput  string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "with output type",
			sig:        "balanceOf(address)(uint256)",
			wantInput:  "balanceOf(address)",
			wantOutput: "(uint256)",
		},
		{
			name:       "no output type",
			sig:        "totalSupply()",
			wantInput:  "totalSupply()",
			wantOutput: "",
		},
		{
			name:       "multiple outputs",
			sig:        "getStatus(address)(address,uint8,uint256)",
			wantInput:  "getStatus(address)",
			wantOutput: "(address,uint8,uint256)",
		},
		{
			name:       "no args with output",
			sig:        "getValidators()(address[])",
			wantInput:  "getValidators()",
			wantOutput: "(address[])",
		},
		{
			name:       "multiple inputs and outputs",
			sig:        "swap(address,uint256)(bool)",
			wantInput:  "swap(address,uint256)",
			wantOutput: "(bool)",
		},
		{
			name:    "no parens",
			sig:     "noparens",
			wantErr: true,
		},
		{
			name:    "no function name",
			sig:     "(address)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, output, err := ParseCallSignature(tt.sig)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if input != tt.wantInput {
				t.Errorf("input: got %q, want %q", input, tt.wantInput)
			}
			if output != tt.wantOutput {
				t.Errorf("output: got %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestConvertArg(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		typStr  string
		check   func(interface{}) bool
		wantErr bool
	}{
		{
			name:   "address",
			input:  "0x0000000000000000000000000000000000001000",
			typStr: "address",
			check: func(v interface{}) bool {
				addr, ok := v.(common.Address)
				return ok && addr == common.HexToAddress("0x0000000000000000000000000000000000001000")
			},
		},
		{
			name:   "uint256 decimal",
			input:  "42",
			typStr: "uint256",
			check: func(v interface{}) bool {
				n, ok := v.(*big.Int)
				return ok && n.Cmp(big.NewInt(42)) == 0
			},
		},
		{
			name:   "uint256 hex",
			input:  "0x2a",
			typStr: "uint256",
			check: func(v interface{}) bool {
				n, ok := v.(*big.Int)
				return ok && n.Cmp(big.NewInt(42)) == 0
			},
		},
		{
			name:   "bool true",
			input:  "true",
			typStr: "bool",
			check: func(v interface{}) bool {
				b, ok := v.(bool)
				return ok && b
			},
		},
		{
			name:   "bool false",
			input:  "false",
			typStr: "bool",
			check: func(v interface{}) bool {
				b, ok := v.(bool)
				return ok && !b
			},
		},
		{
			name:   "string",
			input:  "hello",
			typStr: "string",
			check: func(v interface{}) bool {
				s, ok := v.(string)
				return ok && s == "hello"
			},
		},
		{
			name:   "uint8",
			input:  "18",
			typStr: "uint8",
			check: func(v interface{}) bool {
				n, ok := v.(uint8)
				return ok && n == 18
			},
		},
		{
			name:    "invalid bool",
			input:   "maybe",
			typStr:  "bool",
			wantErr: true,
		},
		{
			name:    "address without 0x",
			input:   "1234567890abcdef1234567890abcdef12345678",
			typStr:  "address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := abi.NewType(tt.typStr, "", nil)
			if err != nil {
				t.Fatalf("failed to create type %s: %v", tt.typStr, err)
			}

			result, err := ConvertArg(tt.input, typ)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.check(result) {
				t.Errorf("check failed for input %q (type %s): got %v (%T)", tt.input, tt.typStr, result, result)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"address", common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"), common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678").Hex()},
		{"big int", big.NewInt(123456789), "123456789"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"string", "hello world", "hello world"},
		{"uint8", uint8(42), "42"},
		{"uint64", uint64(1000000), "1000000"},
		{"int32", int32(-5), "-5"},
		{"bytes", []byte{0xab, 0xcd}, "0xabcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(tt.value)
			if got != tt.want {
				t.Errorf("FormatValue(%v): got %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestBuildCalldata(t *testing.T) {
	// totalSupply() should produce exactly 4 bytes (just the selector)
	data, err := BuildCalldata("totalSupply()", nil)
	if err != nil {
		t.Fatalf("BuildCalldata error: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("expected 4 bytes for no-arg call, got %d", len(data))
	}

	// balanceOf(address) should produce 4 + 32 bytes
	data, err = BuildCalldata("balanceOf(address)", []string{"0x0000000000000000000000000000000000001000"})
	if err != nil {
		t.Fatalf("BuildCalldata error: %v", err)
	}
	if len(data) != 36 {
		t.Errorf("expected 36 bytes for single address arg, got %d", len(data))
	}

	// Wrong arg count should error
	_, err = BuildCalldata("balanceOf(address)", nil)
	if err == nil {
		t.Errorf("expected error for wrong arg count, got nil")
	}
}
