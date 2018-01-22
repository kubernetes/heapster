package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"reflect"
	"sort"

	"github.com/lestrrat/go-msgpack/internal/util"
	"github.com/pkg/errors"
)

func main() {
	if err := _main(); err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
}

func _main() error {
	if err := generateNumericDecoders(); err != nil {
		return errors.Wrap(err, `failed to generate numeric decoders`)
	}
	return nil
}

func generateNumericDecoders() error {

	var buf bytes.Buffer

	buf.WriteString("package msgpack")
	buf.WriteString("\n\n// Auto-generated by internal/cmd/gendecoder-numeric/gendecoder-numeric.go. DO NOT EDIT!")
	buf.WriteString("\n\nimport (")
	buf.WriteString("\n\"math\"")
	buf.WriteString("\n\n\"github.com/pkg/errors\"")
	buf.WriteString("\n)")

	if err := generateIntegerTypes(&buf); err != nil {
		return errors.Wrap(err, `failed to generate integer decoders`)
	}

	if err := generateFloatTypes(&buf); err != nil {
		return errors.Wrap(err, `failed to generate float decoders`)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		return err
	}

	var fn string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] != "-" {
			fn = os.Args[i]
			break
		}
	}

	dst, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, `failed to open file`)
	}
	defer dst.Close()

	dst.Write(formatted)
	return nil
}

func generateIntegerTypes(dst io.Writer) error {
	types := map[reflect.Kind]struct {
		Code string
		Bits int
	}{
		reflect.Int:    {Code: "Int64", Bits: 64},
		reflect.Int8:   {Code: "Int8", Bits: 8},
		reflect.Int16:  {Code: "Int16", Bits: 16},
		reflect.Int32:  {Code: "Int32", Bits: 32},
		reflect.Int64:  {Code: "Int64", Bits: 64},
		reflect.Uint:   {Code: "Uint64", Bits: 64},
		reflect.Uint8:  {Code: "Uint8", Bits: 8},
		reflect.Uint16: {Code: "Uint16", Bits: 16},
		reflect.Uint32: {Code: "Uint32", Bits: 32},
		reflect.Uint64: {Code: "Uint64", Bits: 64},
	}

	keys := make([]reflect.Kind, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})
	for _, typ := range keys {
		data := types[typ]
		fmt.Fprintf(dst, "\n\nfunc (d *Decoder) Decode%s(v *%s) error {", util.Ucfirst(typ.String()), typ)
		fmt.Fprintf(dst, "\ncode, err := d.src.ReadByte()")
		fmt.Fprintf(dst, "\nif err != nil {")
		fmt.Fprintf(dst, "\nreturn errors.Wrap(err, `msgpack: failed to read code for %s`)", data.Code)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\nif IsFixNumFamily(Code(code)) {")
		fmt.Fprintf(dst, "\n*v = %s(code)", typ)
		fmt.Fprintf(dst, "\nreturn nil")
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\n\nif code != %s.Byte() {", data.Code)
		fmt.Fprintf(dst, "\nreturn errors.Errorf(`msgpack: expected %s, got %%s`, Code(code))", data.Code)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\nx, err := d.src.ReadUint%d()", data.Bits)
		fmt.Fprintf(dst, "\nif err != nil {")
		fmt.Fprintf(dst, "\nreturn errors.Wrap(err, `msgpack: failed to read payload for %s`)", typ)
		fmt.Fprintf(dst, "\n}")
		switch typ {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Fprintf(dst, "\n\n*v = x")
		default:
			fmt.Fprintf(dst, "\n\n*v = %s(x)", typ)
		}
		fmt.Fprintf(dst, "\nreturn nil")
		fmt.Fprintf(dst, "\n}")
	}
	return nil
}

func generateFloatTypes(dst io.Writer) error {
	types := map[reflect.Kind]struct {
		Code string
		Bits int
	}{
		reflect.Float32: {Code: "Float", Bits: 32},
		reflect.Float64: {Code: "Double", Bits: 64},
	}

	keys := make([]reflect.Kind, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})
	for _, typ := range keys {
		data := types[typ]
		fmt.Fprintf(dst, "\n\nfunc (d *Decoder) Decode%s(v *%s) error {", util.Ucfirst(typ.String()), typ)
		fmt.Fprintf(dst, "\ncode, x, err := d.src.ReadByteUint%d()", data.Bits)
		fmt.Fprintf(dst, "\nif err != nil {")
		fmt.Fprintf(dst, "\nreturn errors.Wrap(err, `msgpack: failed to read %s`)", typ)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\n\nif code != %s.Byte() {", data.Code)
		fmt.Fprintf(dst, "\nreturn errors.Errorf(`msgpack: expected %s, got %%s`, Code(code))", data.Code)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\n\n*v = math.Float%dfrombits(x)", data.Bits)
		fmt.Fprintf(dst, "\nreturn nil")
		fmt.Fprintf(dst, "\n}")
	}
	return nil
}
