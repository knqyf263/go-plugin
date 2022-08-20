// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package defval marshals and unmarshals textual forms of default values.
//
// This package handles both the form historically used in Go struct field tags
// and also the form used by google.protobuf.FieldDescriptorProto.default_value
// since they differ in superficial ways.
package defval

import (
	"fmt"
	"math"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Format is the serialization format used to represent the default value.
type Format int

const (
	_ Format = iota

	// GoTag uses the historical serialization format in Go struct field tags.
	GoTag
)

// Marshal serializes v as the default string according to the given kind k.
// When specifying the Descriptor format for an enum kind, the associated
// enum value descriptor must be provided.
func Marshal(v protoreflect.Value, ev protoreflect.EnumValueDescriptor, k protoreflect.Kind, f Format) (string, error) {
	switch k {
	case protoreflect.BoolKind:
		if f == GoTag {
			if v.Bool() {
				return "1", nil
			} else {
				return "0", nil
			}
		} else {
			if v.Bool() {
				return "true", nil
			} else {
				return "false", nil
			}
		}
	case protoreflect.EnumKind:
		if f == GoTag {
			return strconv.FormatInt(int64(v.Enum()), 10), nil
		} else {
			return string(ev.Name()), nil
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(v.Int(), 10), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(v.Uint(), 10), nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		f := v.Float()
		switch {
		case math.IsInf(f, -1):
			return "-inf", nil
		case math.IsInf(f, +1):
			return "inf", nil
		case math.IsNaN(f):
			return "nan", nil
		default:
			if k == protoreflect.FloatKind {
				return strconv.FormatFloat(f, 'g', -1, 32), nil
			} else {
				return strconv.FormatFloat(f, 'g', -1, 64), nil
			}
		}
	case protoreflect.StringKind:
		// String values are serialized as is without any escaping.
		return v.String(), nil
	case protoreflect.BytesKind:
		if s, ok := marshalBytes(v.Bytes()); ok {
			return s, nil
		}
	}
	return "", fmt.Errorf("could not format value for %v: %v", k, v)
}

// marshalBytes serializes bytes by using C escaping.
// To match the exact output of protoc, this is identical to the
// CEscape function in strutil.cc of the protoc source code.
func marshalBytes(b []byte) (string, bool) {
	var s []byte
	for _, c := range b {
		switch c {
		case '\n':
			s = append(s, `\n`...)
		case '\r':
			s = append(s, `\r`...)
		case '\t':
			s = append(s, `\t`...)
		case '"':
			s = append(s, `\"`...)
		case '\'':
			s = append(s, `\'`...)
		case '\\':
			s = append(s, `\\`...)
		default:
			if printableASCII := c >= 0x20 && c <= 0x7e; printableASCII {
				s = append(s, c)
			} else {
				s = append(s, fmt.Sprintf(`\%03o`, c)...)
			}
		}
	}
	return string(s), true
}
