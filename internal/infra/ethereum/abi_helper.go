package ethereum

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ParseCallSignature splits "balanceOf(address)(uint256)" into input sig and output types.
// Returns inputSig="balanceOf(address)", outputTypes="(uint256)".
// If no output types: outputTypes="".
func ParseCallSignature(sig string) (string, string, error) {
	// Find the end of the first balanced parenthesized group
	openIdx := strings.IndexByte(sig, '(')
	if openIdx < 0 {
		return "", "", fmt.Errorf("no opening parenthesis in signature")
	}
	if openIdx == 0 {
		return "", "", fmt.Errorf("missing function name before parenthesis")
	}

	depth := 0
	closeIdx := -1
	for i := openIdx; i < len(sig); i++ {
		switch sig[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				closeIdx = i
				goto found
			}
		}
	}
	return "", "", fmt.Errorf("unmatched parenthesis in signature")

found:
	inputSig := sig[:closeIdx+1]
	rest := sig[closeIdx+1:]

	if rest == "" {
		return inputSig, "", nil
	}

	// Validate output types format
	if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, ")") {
		return "", "", fmt.Errorf("invalid output types: %q", rest)
	}

	return inputSig, rest, nil
}

// BuildCalldata encodes a function call from a signature string and CLI arguments.
// Returns 4-byte selector + ABI-encoded arguments.
func BuildCalldata(inputSig string, args []string) ([]byte, error) {
	sel, err := abi.ParseSelector(inputSig)
	if err != nil {
		return nil, fmt.Errorf("parse signature: %w", err)
	}

	inputs, err := argumentsFromMarshalings(sel.Inputs)
	if err != nil {
		return nil, fmt.Errorf("build input types: %w", err)
	}

	if len(args) != len(inputs) {
		return nil, fmt.Errorf("argument count mismatch: got %d, want %d", len(args), len(inputs))
	}

	// Convert CLI string args to Go types
	goArgs := make([]interface{}, len(args))
	for i, argStr := range args {
		goArgs[i], err = ConvertArg(argStr, inputs[i].Type)
		if err != nil {
			return nil, fmt.Errorf("argument %d (%s): %w", i, inputs[i].Type.String(), err)
		}
	}

	// Build method to get 4-byte selector
	method := abi.NewMethod(sel.Name, sel.Name, abi.Function, "view", true, false, inputs, abi.Arguments{})

	// Pack arguments
	packed, err := inputs.Pack(goArgs...)
	if err != nil {
		return nil, fmt.Errorf("abi pack: %w", err)
	}

	return append(method.ID, packed...), nil
}

// ConvertArg converts a CLI string argument to the Go type expected by abi.Arguments.Pack.
func ConvertArg(s string, typ abi.Type) (interface{}, error) {
	switch typ.T {
	case abi.AddressTy:
		if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
			return nil, fmt.Errorf("address must start with 0x")
		}
		return common.HexToAddress(s), nil

	case abi.UintTy:
		return parseUint(s, typ.Size)

	case abi.IntTy:
		return parseInt(s, typ.Size)

	case abi.BoolTy:
		switch strings.ToLower(s) {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool: %q (use true/false)", s)
		}

	case abi.StringTy:
		return s, nil

	case abi.BytesTy:
		return hexDecode(s)

	case abi.FixedBytesTy:
		b, err := hexDecode(s)
		if err != nil {
			return nil, err
		}
		return toFixedBytes(b, typ.Size)

	default:
		return nil, fmt.Errorf("unsupported input type: %s", typ.String())
	}
}

// DecodeOutputs decodes raw eth_call return data using the output type string.
// outputTypes is e.g. "(uint256)" or "(address,uint8,uint256)".
// Returns nil if outputTypes is empty.
func DecodeOutputs(outputTypes string, data []byte) ([]string, error) {
	if outputTypes == "" {
		return nil, nil
	}

	// Use ParseSelector with a dummy function name to reuse the parser
	sel, err := abi.ParseSelector("x" + outputTypes)
	if err != nil {
		return nil, fmt.Errorf("parse output types: %w", err)
	}

	outputs, err := argumentsFromMarshalings(sel.Inputs)
	if err != nil {
		return nil, fmt.Errorf("build output types: %w", err)
	}

	values, err := outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("unpack: %w", err)
	}

	result := make([]string, len(values))
	for i, v := range values {
		result[i] = FormatValue(v)
	}
	return result, nil
}

// DecodeRawOutputs is like DecodeOutputs but returns raw Go values instead of formatted strings.
// Useful when callers need to work with the actual types (e.g., []common.Address).
func DecodeRawOutputs(outputTypes string, data []byte) ([]interface{}, error) {
	if outputTypes == "" {
		return nil, nil
	}

	sel, err := abi.ParseSelector("x" + outputTypes)
	if err != nil {
		return nil, fmt.Errorf("parse output types: %w", err)
	}

	outputs, err := argumentsFromMarshalings(sel.Inputs)
	if err != nil {
		return nil, fmt.Errorf("build output types: %w", err)
	}

	values, err := outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("unpack: %w", err)
	}

	return values, nil
}

// argumentsFromMarshalings converts []abi.ArgumentMarshaling to abi.Arguments.
func argumentsFromMarshalings(marshalings []abi.ArgumentMarshaling) (abi.Arguments, error) {
	args := make(abi.Arguments, len(marshalings))
	for i, am := range marshalings {
		typ, err := abi.NewType(am.Type, am.InternalType, am.Components)
		if err != nil {
			return nil, fmt.Errorf("type %q: %w", am.Type, err)
		}
		args[i] = abi.Argument{
			Name: am.Name,
			Type: typ,
		}
	}
	return args, nil
}

// FormatValue converts a Go value from abi.Unpack to a human-readable string.
func FormatValue(v interface{}) string {
	switch val := v.(type) {
	case common.Address:
		return val.Hex()
	case *big.Int:
		return val.String()
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return val
	case []byte:
		return "0x" + hex.EncodeToString(val)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		// Handle slices and arrays via reflection
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			items := make([]string, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				items[i] = FormatValue(rv.Index(i).Interface())
			}
			return "[" + strings.Join(items, ", ") + "]"
		}
		return fmt.Sprintf("%v", v)
	}
}

// parseUint parses a string to the correct uint type based on bit size.
func parseUint(s string, bitSize int) (interface{}, error) {
	n := new(big.Int)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, ok := n.SetString(s[2:], 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex number: %s", s)
		}
	} else {
		_, ok := n.SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("invalid number: %s", s)
		}
	}

	switch {
	case bitSize <= 8:
		return uint8(n.Uint64()), nil
	case bitSize <= 16:
		return uint16(n.Uint64()), nil
	case bitSize <= 32:
		return uint32(n.Uint64()), nil
	case bitSize <= 64:
		return n.Uint64(), nil
	default:
		return n, nil
	}
}

// parseInt parses a string to the correct int type based on bit size.
func parseInt(s string, bitSize int) (interface{}, error) {
	n := new(big.Int)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, ok := n.SetString(s[2:], 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex number: %s", s)
		}
	} else {
		_, ok := n.SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("invalid number: %s", s)
		}
	}

	switch {
	case bitSize <= 8:
		return int8(n.Int64()), nil
	case bitSize <= 16:
		return int16(n.Int64()), nil
	case bitSize <= 32:
		return int32(n.Int64()), nil
	case bitSize <= 64:
		return n.Int64(), nil
	default:
		return n, nil
	}
}

// hexDecode decodes a hex string (with or without 0x prefix) to bytes.
func hexDecode(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	return hex.DecodeString(s)
}

// toFixedBytes converts a byte slice to a fixed-size byte array using reflection.
func toFixedBytes(b []byte, size int) (interface{}, error) {
	if len(b) > size {
		return nil, fmt.Errorf("bytes too long: got %d, max %d", len(b), size)
	}
	// Create a fixed-size array via reflection
	arr := reflect.New(reflect.ArrayOf(size, reflect.TypeOf(byte(0)))).Elem()
	for i := 0; i < len(b) && i < size; i++ {
		arr.Index(i).Set(reflect.ValueOf(b[i]))
	}
	return arr.Interface(), nil
}
