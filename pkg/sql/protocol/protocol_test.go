// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/ring/bitand"
	"github.com/matrixorigin/matrixone/pkg/container/ring/bitor"
	"github.com/matrixorigin/matrixone/pkg/container/ring/bitxor"
	"github.com/matrixorigin/matrixone/pkg/container/ring/variance"

	"github.com/axiomhq/hyperloglog"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/ring"
	"github.com/matrixorigin/matrixone/pkg/container/ring/approxcd"
	"github.com/matrixorigin/matrixone/pkg/container/ring/avg"
	"github.com/matrixorigin/matrixone/pkg/container/ring/count"
	"github.com/matrixorigin/matrixone/pkg/container/ring/max"
	"github.com/matrixorigin/matrixone/pkg/container/ring/min"
	"github.com/matrixorigin/matrixone/pkg/container/ring/starcount"
	"github.com/matrixorigin/matrixone/pkg/container/ring/sum"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/encoding"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/connector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/dedup"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/extend"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/limit"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/merge"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/offset"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/order"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/output"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/projection"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/restrict"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/top"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/join"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/oplus"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/plus"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/transform"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/transformer"
	"github.com/matrixorigin/matrixone/pkg/sql/viewexec/untransform"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/stretchr/testify/require"
)

func TestTransform(t *testing.T) {
	var buf bytes.Buffer
	ins := vm.Instruction{
		Op:  vm.Transform,
		Arg: &transform.Argument{},
	}
	err := EncodeInstruction(ins, &buf)
	require.NoError(t, err)
	resultIns, _, err := DecodeInstruction(buf.Bytes())
	require.NoError(t, err)
	// Op
	if resultIns.Op != ins.Op {
		t.Errorf("Decode resultIns.Op failed. \nExpected/Got:\n%v\n%v", resultIns.Op, ins.Op)
		return
	}
	fmt.Println(resultIns)
}

func TestInstruction(t *testing.T) {
	insArray := []vm.Instruction{
		{
			Op: vm.Top,
			Arg: &top.Argument{
				Limit: 123,
				Fs:    []top.Field{{Attr: "hello", Type: top.Ascending}},
			},
		},
		{
			Op: vm.Plus,
			Arg: &plus.Argument{
				Typ: 123897,
			},
		},
		{
			Op: vm.Limit,
			Arg: &limit.Argument{
				Seen:  12893792345,
				Limit: 89757435634,
			},
		},
		{
			Op: vm.Join,
			Arg: &join.Argument{
				Vars: [][]string{{"x"}, {"y"}},
			},
		},
		{
			Op:  vm.Merge,
			Arg: &merge.Argument{},
		},
		{
			Op:  vm.Dedup,
			Arg: &dedup.Argument{},
		},
		{
			Op: vm.Order,
			Arg: &order.Argument{
				Fs: []order.Field{
					{
						Attr: "order field attr",
						Type: 1,
					},
				},
			},
		},
		{
			Op:  vm.Oplus,
			Arg: &oplus.Argument{Typ: 12312423},
		},
		{
			Op:  vm.Output,
			Arg: &output.Argument{Attrs: []string{"the", "first", "loop"}},
		},
		{
			Op: vm.Offset,
			Arg: &offset.Argument{
				Seen:   1231245,
				Offset: 65784654,
			},
		},
		{
			Op: vm.Restrict,
			Arg: &restrict.Argument{
				Attrs: []string{"the", "first", "loop"},
				E:     &extend.StarExtend{},
			},
		},
		{
			Op:  vm.Connector,
			Arg: &connector.Argument{},
		},
		{
			Op: vm.Connector,
			Arg: &transform.Argument{
				Typ:      12312312,
				IsMerge:  false,
				FreeVars: []string{"vm", "times", "test"},
				BoundVars: []transformer.Transformer{
					{
						Op:    1231,
						Ref:   897897,
						Name:  "happening",
						Alias: "method",
					},
				},
			},
		},
		{
			Op: vm.Projection,
			Arg: &projection.Argument{
				Rs: []uint64{123, 98789, 3456456},
				As: []string{"vm", "times", "test"},
				Es: []extend.Extend{
					&extend.StarExtend{},
				},
			},
		},
		{
			Op: vm.UnTransform,
			Arg: &untransform.Argument{
				FreeVars: []string{"the", "first", "loop"},
				Type:     1231237,
			},
		},
	}
	for _, ins := range insArray {
		var buf bytes.Buffer
		err := EncodeInstruction(ins, &buf)
		require.NoError(t, err)
		resultIns, _, err := DecodeInstruction(buf.Bytes())
		require.NoError(t, err)
		switch resultIns.Op {
		case vm.Top:
			expectArg := resultIns.Arg.(*top.Argument)
			actualArg := ins.Arg.(*top.Argument)
			// Limit
			if expectArg.Limit != actualArg.Limit {
				t.Errorf("Decode arg limit failed.")
				return
			}
			// Field Attr
			if expectArg.Fs[0].Attr != actualArg.Fs[0].Attr {
				t.Errorf("Decode arg Attr failed.")
				return
			}
			// Field Type
			if expectArg.Fs[0].Type != actualArg.Fs[0].Type {
				t.Errorf("Decode arg Type failed.")
				return
			}
		case vm.Plus:
			expectArg := resultIns.Arg.(*plus.Argument)
			actualArg := ins.Arg.(*plus.Argument)
			// Typ
			if expectArg.Typ != actualArg.Typ {
				t.Errorf("Decode arg Typ failed.")
				return
			}
		case vm.Limit:
			expectArg := resultIns.Arg.(*limit.Argument)
			actualArg := ins.Arg.(*limit.Argument)
			// Seen
			if expectArg.Seen != actualArg.Seen {
				t.Errorf("Decode arg Typ failed.")
				return
			}
			// Limit
			if expectArg.Limit != actualArg.Limit {
				t.Errorf("Decode arg Typ failed.")
				return
			}
		case vm.Join:
			expectArg := resultIns.Arg.(*join.Argument)
			actualArg := ins.Arg.(*join.Argument)
			// Vars
			for i, ar := range actualArg.Vars {
				as := expectArg.Vars[i]
				for j, a := range ar {
					if a != as[j] {
						t.Errorf("Decode arg vars failed.")
						return
					}
				}
			}
		case vm.Merge:
			if ins.Op != resultIns.Op {
				t.Errorf("Decode instruction Op failed.")
				return
			}
		case vm.Dedup:
			if ins.Op != resultIns.Op {
				t.Errorf("Decode instruction Op failed.")
				return
			}
		case vm.Order:
			expectArg := resultIns.Arg.(*order.Argument)
			actualArg := ins.Arg.(*order.Argument)
			// Field Attr
			if expectArg.Fs[0].Attr != actualArg.Fs[0].Attr {
				t.Errorf("Decode arg Attr failed.")
				return
			}
			// Field Type
			if expectArg.Fs[0].Type != actualArg.Fs[0].Type {
				t.Errorf("Decode arg Type failed.")
				return
			}
		case vm.Oplus:
			expectArg := resultIns.Arg.(*oplus.Argument)
			actualArg := ins.Arg.(*oplus.Argument)
			// Typ
			if expectArg.Typ != actualArg.Typ {
				t.Errorf("Decode arg Typ failed.")
				return
			}
		case vm.Output:
			expectArg := resultIns.Arg.(*output.Argument)
			actualArg := ins.Arg.(*output.Argument)
			// Attrs
			for i, ar := range actualArg.Attrs {
				if expectArg.Attrs[i] != ar {
					t.Errorf("Decode arg Attrs failed.")
					return
				}
			}
		case vm.Offset:
			expectArg := resultIns.Arg.(*offset.Argument)
			actualArg := ins.Arg.(*offset.Argument)
			// Seen
			if expectArg.Seen != actualArg.Seen {
				t.Errorf("Decode arg Seen failed.")
				return
			}
			// Offset
			if expectArg.Offset != actualArg.Offset {
				t.Errorf("Decode arg Offset failed.")
				return
			}
		case vm.Restrict:
			expectArg := resultIns.Arg.(*restrict.Argument)
			actualArg := ins.Arg.(*restrict.Argument)
			// Attrs
			for i, ar := range actualArg.Attrs {
				if expectArg.Attrs[i] != ar {
					t.Errorf("Decode arg Attrs failed.")
					return
				}
			}
		case vm.Connector:
			if ins.Op != resultIns.Op {
				t.Errorf("Decode instruction Op failed.")
				return
			}
		case vm.Transform:
			expectArg := resultIns.Arg.(*transform.Argument)
			actualArg := ins.Arg.(*transform.Argument)
			// Arg
			if expectArg.BoundVars[0].Alias != actualArg.BoundVars[0].Alias {
				t.Errorf("Decode arg Arg failed.")
				return
			}
		case vm.Projection:
			expectArg := resultIns.Arg.(*projection.Argument)
			actualArg := ins.Arg.(*projection.Argument)
			// Rs
			for i, ar := range actualArg.Rs {
				if expectArg.Rs[i] != ar {
					t.Errorf("Decode arg Rs failed.")
					return
				}
			}
			// As
			for i, ar := range actualArg.As {
				if expectArg.As[i] != ar {
					t.Errorf("Decode arg As failed.")
					return
				}
			}
			// extend
		case vm.UnTransform:
			expectArg := resultIns.Arg.(*untransform.Argument)
			actualArg := ins.Arg.(*untransform.Argument)
			// Typ
			if expectArg.Type != actualArg.Type {
				t.Errorf("Decode arg Typ failed.")
				return
			}
			// FreeVars
			for i, ar := range actualArg.FreeVars {
				if expectArg.FreeVars[i] != ar {
					t.Errorf("Decode arg FreeVars failed.")
					return
				}
			}
		}
	}
}

func TestExtend(t *testing.T) {
	extendArray := []extend.Extend{
		&extend.UnaryExtend{
			Op: 123123,
			E:  &extend.ValueExtend{V: NewFloatVector(1.2)},
		},
		&extend.BinaryExtend{
			Op:    574635,
			Left:  &extend.ValueExtend{V: NewFloatVector(1.2)},
			Right: &extend.FuncExtend{Name: "binary Extend"},
		},
		&extend.MultiExtend{
			Op: 123123,
			Args: []extend.Extend{
				&extend.FuncExtend{Name: "Multi Extend"},
				&extend.ValueExtend{V: NewFloatVector(1.2)},
			},
		},
		&extend.ParenExtend{
			E: &extend.FuncExtend{Name: "Paren Extend"},
		},
		&extend.FuncExtend{
			Name: "Func Extend",
			Args: []extend.Extend{
				&extend.FuncExtend{Name: "function args Extend"},
				&extend.ValueExtend{V: NewFloatVector(1.2)},
			},
		},
		&extend.StarExtend{},
		&extend.ValueExtend{
			V: NewInt32Vector(123123),
		},
		&extend.Attribute{
			Name: "attribute",
			Type: types.T_varchar,
		},
	}
	for _, e := range extendArray {
		var buf bytes.Buffer
		err := EncodeExtend(e, &buf)
		require.NoError(t, err)
		resultE, _, err := DecodeExtend(buf.Bytes())
		require.NoError(t, err)
		switch expectE := resultE.(type) {
		case *extend.UnaryExtend:
			actualE := e.(*extend.UnaryExtend)
			// Op
			if expectE.Op != actualE.Op {
				t.Errorf("Decode exetend Op failed. \nExpected/Got:\n%v\n%v", actualE.Op, expectE.Op)
				return
			}
		case *extend.BinaryExtend:
			actualE := e.(*extend.BinaryExtend)
			// Op
			if expectE.Op != actualE.Op {
				t.Errorf("Decode exetend Op failed. \nExpected/Got:\n%v\n%v", actualE.Op, expectE.Op)
				return
			}
			// Name
			if expectE.Right.(*extend.FuncExtend).Name != actualE.Right.(*extend.FuncExtend).Name {
				t.Error("Decode extend Name failed.")
				return
			}
		case *extend.MultiExtend:
			actualE := e.(*extend.MultiExtend)
			// Op
			if expectE.Op != actualE.Op {
				t.Errorf("Decode exetend Op failed. \nExpected/Got:\n%v\n%v", actualE.Op, expectE.Op)
				return
			}
			// Name
			if expectE.Args[0].(*extend.FuncExtend).Name != actualE.Args[0].(*extend.FuncExtend).Name {
				t.Error("Decode extend Name failed.")
				return
			}
		case *extend.ParenExtend:
			actualE := e.(*extend.ParenExtend)
			// E
			if expectE.E.(*extend.FuncExtend).Name != actualE.E.(*extend.FuncExtend).Name {
				t.Error("Decode extend E failed.")
				return
			}
		case *extend.FuncExtend:
			actualE := e.(*extend.FuncExtend)
			if expectE.Name != actualE.Name {
				t.Error("Decode extend Name failed.")
				return
			}
			// Args
			if expectE.Args[0].(*extend.FuncExtend).Name != actualE.Args[0].(*extend.FuncExtend).Name {
				t.Error("Decode extend Args failed.")
				return
			}
		case *extend.StarExtend:
			actualE := e.(*extend.StarExtend)
			fmt.Println(actualE)
		case *extend.ValueExtend:
			actualE := e.(*extend.ValueExtend)
			if expectE.V.Ref != actualE.V.Ref {
				t.Error("Decode extend Args failed.")
				return
			}
		case *extend.Attribute:
			actualE := e.(*extend.Attribute)
			if expectE.Name != actualE.Name {
				t.Error("Decode extend Name failed.")
				return
			}
			// Args
			if expectE.Type != actualE.Type {
				t.Error("Decode extend Type failed.")
				return
			}
		}
	}
}

func NewStrVector(v []byte) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_varchar), Size: 24})
	vector.Append(vec, [][]byte{v, v, v})
	return vec
}

func NewFloatVector(v float64) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_float64), Size: 8})
	vector.Append(vec, []float64{v, v, v})
	return vec
}

func NewInt8Vector(v int8) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_int8), Size: 8})
	vector.Append(vec, []int8{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewInt16Vector(v int16) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_int16), Size: 8})
	vector.Append(vec, []int16{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewInt32Vector(v int32) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_int32), Size: 8})
	vector.Append(vec, []int32{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewInt64Vector(v int64) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_int64), Size: 8})
	vector.Append(vec, []int64{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewUInt8Vector(v uint8) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_uint8), Size: 8})
	vector.Append(vec, []uint8{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewUInt16Vector(v uint16) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_uint16), Size: 8})
	vector.Append(vec, []uint16{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewUInt32Vector(v uint32) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_uint32), Size: 8})
	vector.Append(vec, []uint32{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewUInt64Vector(v uint64) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_uint64), Size: 8})
	vector.Append(vec, []uint64{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewFloat32Vector(v float32) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_float32), Size: 8})
	vector.Append(vec, []float32{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewFloat64Vector(v float64) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_float64), Size: 8})
	vector.Append(vec, []float64{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewCharVector(v []byte) *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_varchar), Size: 24})
	vector.Append(vec, [][]byte{v, v, v})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewTupleVector() *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_tuple), Size: 24})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewDateVector() *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_date), Size: 24})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func NewDatetimeVector() *vector.Vector {
	vec := vector.New(types.Type{Oid: types.T(types.T_datetime), Size: 24})
	vec.Ref = 12839791322
	vec.Link = 123908123
	vec.Data = []byte("Unless required by applicable law or agreed to in writing")
	return vec
}

func TestVector(t *testing.T) {
	vecArray := []*vector.Vector{
		NewInt8Vector(1),
		NewInt16Vector(12),
		NewInt32Vector(45),
		NewInt64Vector(134),
		NewUInt8Vector(1),
		NewUInt16Vector(12),
		NewUInt32Vector(45),
		NewUInt64Vector(134),
		NewFloat32Vector(1.5),
		NewFloat64Vector(1.2),
		NewCharVector([]byte("the next method of its source")),
		NewTupleVector(),
		NewDateVector(),
		NewDatetimeVector(),
	}
	for _, vec := range vecArray {
		var buf bytes.Buffer
		err := EncodeVector(vec, &buf)
		require.NoError(t, err)
		resultVec, _, err := DecodeVector(buf.Bytes())
		require.NoError(t, err)
		// Or
		if resultVec.Or != true {
			t.Errorf("Decode resultVec.Or failed. \nExpected/Got:\n%v\n%v", true, resultVec.Or)
			return
		}
		// Ref
		if resultVec.Ref != vec.Ref {
			t.Errorf("Decode resultVec.Ref failed. \nExpected/Got:\n%v\n%v", vec.Ref, resultVec.Ref)
			return
		}
		// Link
		if resultVec.Link != vec.Link {
			t.Errorf("Decode resultVec.Link failed. \nExpected/Got:\n%v\n%v", vec.Link, resultVec.Link)
			return
		}
		// Data
		if string(resultVec.Data) != string(vec.Data) {
			t.Errorf("Decode resultVec.Data failed. \nExpected/Got:\n%v\n%v", string(vec.Data), string(resultVec.Data))
			return
		}
		// Typ
		if resultVec.Typ.Oid != vec.Typ.Oid {
			t.Errorf("Decode resultVec.Typ failed. \nExpected/Got:\n%v\n%v", resultVec.Typ.Oid, vec.Typ.Oid)
			return
		}
		// Col
		// if  resultVec.Col.([]float64)[0] != vec.Col.([]float64)[0] {
		//	t.Errorf("Decode Batch.Vecs failed. \nExpected/Got:\n%v\n%v", vec.Col.([]float64)[0], resultVec.Col.([]float64)[0])
		//	return
		//}
	}
}

func TestRing(t *testing.T) {
	sk := hyperloglog.New()
	sk.Insert([]byte{0, 0, 0, 1})
	sk.Insert([]byte{0, 1, 0, 1})
	sk.Insert([]byte{0, 1, 2, 1})
	sk2 := hyperloglog.New()
	sk2.Insert([]byte{4, 0, 0, 1})
	sk2.Insert([]byte{0, 1, 0, 1})
	ringArray := []ring.Ring{
		&avg.AvgRing{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []float64{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&count.CountRing{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []int64{12312312, 34534345, 234123345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&starcount.CountRing{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []int64{12312312, 34534345, 234123345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&approxcd.ApproxCountDistinctRing{
			Vs:  []uint64{3, 2, 0},
			Sk:  []*hyperloglog.Sketch{sk, sk2, hyperloglog.New()},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Int8Ring{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []int8{6, 6, 8, 0},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Int16Ring{
			Ns:  []int64{567567, 123123908950, 9089374534},
			Vs:  []int16{62, 62, 8, 01},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Int32Ring{
			Ns:  []int64{789789, 123123908950, 9089374534},
			Vs:  []int32{612, 632, 81, 0423},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Int64Ring{
			Ns:  []int64{178923123, 123123908950, 9089374534},
			Vs:  []int64{6123, 123126, 2323328, 02342342},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.UInt8Ring{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []uint8{6, 6, 8, 0},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.UInt16Ring{
			Ns:  []int64{45634564, 123123908950, 9089374534},
			Vs:  []uint16{6123, 1236, 8123, 12310},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.UInt32Ring{
			Ns:  []int64{56784567, 123123908950, 9089374534},
			Vs:  []uint32{6123, 3454346, 345348, 345340},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.UInt64Ring{
			Ns:  []int64{8902345, 123123908950, 9089374534},
			Vs:  []uint64{6112323, 34542345346, 234, 23412312},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Float32Ring{
			Ns:  []int64{3246457, 123123908950, 9089374534},
			Vs:  []float32{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.Float64Ring{
			Ns:  []int64{996674, 123123908950, 9089374534},
			Vs:  []float64{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.StrRing{
			Ns:  []int64{1231245234, 123123123908950, 123},
			Vs:  [][]byte{[]byte("test1"), []byte("mysql1"), []byte("postgresql1")},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&max.DateRing{
			Ns:  []int64{789789, 123123908950, 9089374534},
			Vs:  []types.Date{612, 632, 81, 0423},
			Typ: types.Type{Oid: types.T(types.T_date), Size: 4},
		},
		&max.DatetimeRing{
			Ns:  []int64{178923123, 123123908950, 9089374534},
			Vs:  []types.Datetime{6123, 123126, 2323328, 02342342},
			Typ: types.Type{Oid: types.T(types.T_datetime), Size: 8},
		},
		&min.Int8Ring{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []int8{6, 6, 8, 0},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.Int16Ring{
			Ns:  []int64{567567, 123123908950, 9089374534},
			Vs:  []int16{62, 62, 8, 01},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.Int32Ring{
			Ns:  []int64{789789, 123123908950, 9089374534},
			Vs:  []int32{612, 632, 81, 0423},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.Int64Ring{
			Ns:  []int64{178923123, 123123908950, 9089374534},
			Vs:  []int64{6123, 123126, 2323328, 02342342},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.UInt8Ring{
			Ns:  []int64{123123123, 123123908950, 9089374534},
			Vs:  []uint8{6, 6, 8, 0},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.UInt16Ring{
			Ns:  []int64{45634564, 123123908950, 9089374534},
			Vs:  []uint16{6123, 1236, 8123, 12310},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.UInt32Ring{
			Ns:  []int64{56784567, 123123908950, 9089374534},
			Vs:  []uint32{6123, 3454346, 345348, 345340},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.UInt64Ring{
			Ns:  []int64{8902345, 123123908950, 9089374534},
			Vs:  []uint64{6112323, 34542345346, 234, 23412312},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.Float32Ring{
			Ns:  []int64{3246457, 123123908950, 9089374534},
			Vs:  []float32{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.Float64Ring{
			Ns:  []int64{996674, 123123908950, 9089374534},
			Vs:  []float64{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.StrRing{
			Ns:  []int64{1231245234, 123123123908950, 123},
			Vs:  [][]byte{[]byte("test1"), []byte("mysql1"), []byte("postgresql1")},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&min.DateRing{
			Ns:  []int64{789789, 123123908950, 9089374534},
			Vs:  []types.Date{612, 632, 81, 0423},
			Typ: types.Type{Oid: types.T(types.T_date), Size: 4},
		},
		&min.DatetimeRing{
			Ns:  []int64{178923123, 123123908950, 9089374534},
			Vs:  []types.Datetime{6123, 123126, 2323328, 02342342},
			Typ: types.Type{Oid: types.T(types.T_datetime), Size: 8},
		},
		&sum.IntRing{
			Ns:  []int64{178923123, 123123908950, 9089374534},
			Vs:  []int64{6123, 123126, 2323328, 02342342},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&sum.UIntRing{
			Ns:  []int64{8902345, 123123908950, 9089374534},
			Vs:  []uint64{6112323, 34542345346, 234, 23412312},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&sum.FloatRing{
			Ns:  []int64{996674, 123123908950, 9089374534},
			Vs:  []float64{123.123, 34534.345, 234123.345345},
			Typ: types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&variance.VarRing{
			NullCounts: []int64{1, 2, 3},
			SumX:       []float64{4, 9, 13},
			SumX2:      []float64{16, 81, 169},
			Typ:        types.Type{Oid: types.T(types.T_float64), Size: 8},
		},
		&bitand.NumericRing{
			NullCnt:      []int64{123123123, 123123908950, 9089374534},
			BitAndResult: []uint64{6, 6, 8, 0},
			Typ:          types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&bitand.NumericRing{
			NullCnt:      []int64{45634564, 123123908950, 9089374534},
			BitAndResult: []uint64{6123, 1236, 8123, 12310},
			Typ:          types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&bitand.NumericRing{
			NullCnt:      []int64{56784567, 123123908950, 9089374534},
			BitAndResult: []uint64{6123, 3454346, 345348, 345340},
			Typ:          types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&bitand.NumericRing{
			NullCnt:      []int64{8902345, 123123908950, 9089374534},
			BitAndResult: []uint64{6112323, 34542345346, 234, 23412312},
			Typ:          types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&bitor.BitOrRing{
			NullCounts: []int64{8902345, 123123908950, 9089374534},
			Values:     []uint64{6112323, 34542345346, 234, 23412312},
			Typ:        types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
		&bitxor.BitXorRing{
			Values:     []uint64{5234232, 6345123, 7345312, 878956},
			NullCounts: []int64{56784567, 123123908950, 9089374534},
			Typ:        types.Type{Oid: types.T(types.T_varchar), Size: 24},
		},
	}
	for _, r := range ringArray {
		var buf bytes.Buffer
		err := EncodeRing(r, &buf)
		if err != nil {
			t.Errorf("Encode err: %v", err)
			return
		}
		resultRing, _, err := DecodeRing(buf.Bytes())
		if err != nil {
			t.Errorf("Decode ring error: %v", err)
			return
		}

		switch ExpectRing := resultRing.(type) {
		case *bitxor.BitXorRing:
			oriRing := r.(*bitxor.BitXorRing)
			// Data
			if string(ExpectRing.Data) != string(encoding.EncodeUint64Slice(oriRing.Values)) {
				t.Errorf("Decode bit_xor numeric ring data failed")
				return
			}
			// NullCnt
			for i, n := range oriRing.NullCounts {
				if ExpectRing.NullCounts[i] != n {
					t.Errorf("Decode bit_xor int ring NullCounts failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.NullCounts[i])
					return
				}
			}
			// BitAndResult
			for i, v := range oriRing.Values {
				if ExpectRing.Values[i] != v {
					t.Errorf("Decode bit_xor int ring BitAndResult failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Values[i])
					return
				}
			}
		case *bitand.NumericRing:
			oriRing := r.(*bitand.NumericRing)
			// Data
			if string(ExpectRing.Data) != string(encoding.EncodeUint64Slice(oriRing.BitAndResult)) {
				t.Errorf("Decode bit_and numeric ring data failed")
				return
			}
			// NullCnt
			for i, n := range oriRing.NullCnt {
				if ExpectRing.NullCnt[i] != n {
					t.Errorf("Decode bit_and int ring NullCnt failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.NullCnt[i])
					return
				}
			}
			// BitAndResult
			for i, v := range oriRing.BitAndResult {
				if ExpectRing.BitAndResult[i] != v {
					t.Errorf("Decode bit_and int ring BitAndResult failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.BitAndResult[i])
					return
				}
			}
		case *avg.AvgRing:
			oriRing := r.(*avg.AvgRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *count.CountRing:
			oriRing := r.(*count.CountRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *starcount.CountRing:
			oriRing := r.(*starcount.CountRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *approxcd.ApproxCountDistinctRing:
			oriRing := r.(*approxcd.ApproxCountDistinctRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
			for i, sk := range oriRing.Sk {
				if expect, got := sk.Estimate(), ExpectRing.Sk[i].Estimate(); expect != got {
					t.Errorf("Decode ring failed. \nExpected/Got:\n%v\n%v", expect, got)
				}
			}
		case *max.Int8Ring:
			oriRing := r.(*max.Int8Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt8Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.Int16Ring:
			oriRing := r.(*max.Int16Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt16Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.Int32Ring:
			oriRing := r.(*max.Int32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.Int64Ring:
			oriRing := r.(*max.Int64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.UInt8Ring:
			oriRing := r.(*max.UInt8Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint8Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.UInt16Ring:
			oriRing := r.(*max.UInt16Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint16Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.UInt32Ring:
			oriRing := r.(*max.UInt32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.UInt64Ring:
			oriRing := r.(*max.UInt64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.Float32Ring:
			oriRing := r.(*max.Float32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.Float64Ring:
			oriRing := r.(*max.Float64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.StrRing:
			oriRing := r.(*max.StrRing)
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if string(ExpectRing.Vs[i]) != string(v) {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", string(v), string(ExpectRing.Vs[i]))
					return
				}
			}
		case *max.DateRing:
			oriRing := r.(*max.DateRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeDateSlice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *max.DatetimeRing:
			oriRing := r.(*max.DatetimeRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeDatetimeSlice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Int8Ring:
			oriRing := r.(*min.Int8Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt8Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Int16Ring:
			oriRing := r.(*min.Int16Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt16Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Int32Ring:
			oriRing := r.(*min.Int32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Int64Ring:
			oriRing := r.(*min.Int64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.UInt8Ring:
			oriRing := r.(*min.UInt8Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint8Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.UInt16Ring:
			oriRing := r.(*min.UInt16Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint16Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.UInt32Ring:
			oriRing := r.(*min.UInt32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.UInt64Ring:
			oriRing := r.(*min.UInt64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Float32Ring:
			oriRing := r.(*min.Float32Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat32Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.Float64Ring:
			oriRing := r.(*min.Float64Ring)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.StrRing:
			oriRing := r.(*min.StrRing)
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if string(ExpectRing.Vs[i]) != string(v) {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", string(v), string(ExpectRing.Vs[i]))
					return
				}
			}
		case *min.DateRing:
			oriRing := r.(*min.DateRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeDateSlice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *min.DatetimeRing:
			oriRing := r.(*min.DatetimeRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeDatetimeSlice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *sum.IntRing:
			oriRing := r.(*sum.IntRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeInt64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *sum.UIntRing:
			oriRing := r.(*sum.UIntRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeUint64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *sum.FloatRing:
			oriRing := r.(*sum.FloatRing)
			// Da
			if string(ExpectRing.Da) != string(encoding.EncodeFloat64Slice(oriRing.Vs)) {
				t.Errorf("Decode ring Da failed.")
				return
			}
			// Ns
			for i, n := range oriRing.Ns {
				if ExpectRing.Ns[i] != n {
					t.Errorf("Decode ring Ns failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.Ns[i])
					return
				}
			}
			// Vs
			for i, v := range oriRing.Vs {
				if ExpectRing.Vs[i] != v {
					t.Errorf("Decode ring Vs failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Vs[i])
					return
				}
			}
		case *variance.VarRing:
			oriRing := r.(*variance.VarRing)
			// Sumx
			if string(ExpectRing.Data) != string(encoding.EncodeFloat64Slice(oriRing.SumX)) {
				t.Errorf("Decode varRing Sums failed.")
				return
			}
			// NullCounts
			for i, n := range oriRing.NullCounts {
				if ExpectRing.NullCounts[i] != n {
					t.Errorf("Decode varRing NullCounts failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.NullCounts[i])
					return
				}
			}
			// Sumx2
			for i, v := range oriRing.SumX2 {
				if !reflect.DeepEqual(ExpectRing.SumX2[i], v) {
					t.Errorf("Decode varRing Values failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.SumX2[i])
					return
				}
			}
		case *bitor.BitOrRing:
			oriRing := r.(*bitor.BitOrRing)
			// Data
			if string(ExpectRing.Data) != string(encoding.EncodeUint64Slice(oriRing.Values)) {
				t.Errorf("Decode ring Data failed.")
				return
			}
			// NullCounts
			for i, n := range oriRing.NullCounts {
				if ExpectRing.NullCounts[i] != n {
					t.Errorf("Decode ring NullCounts failed. \nExpected/Got:\n%v\n%v", n, ExpectRing.NullCounts[i])
					return
				}
			}
			// Values
			for i, v := range oriRing.Values {
				if ExpectRing.Values[i] != v {
					t.Errorf("Decode ring Values failed. \nExpected/Got:\n%v\n%v", v, ExpectRing.Values[i])
					return
				}
			}
		default:
			t.Error("Can not find the type of ring")
		}
	}
}

func TestBatch(t *testing.T) {
	var buf bytes.Buffer

	bat := batch.New(true, []string{"a", "b", "c"})
	bat.Vecs[0] = NewFloatVector(1.2)
	bat.Vecs[1] = NewFloatVector(2.1)
	bat.Vecs[2] = NewStrVector([]byte("x"))
	bat.SelsData = []byte("hello")
	bat.Sels = []int64{123, 456, 678}
	bat.Zs = []int64{908, 46, 1234, 23412}
	bat.Refs = []uint64{90123123128, 41231231236, 11231231234, 23412123123123}
	bat.As = []string{"hello", "world", "big", "small"}
	// encode
	err := EncodeBatch(bat, &buf)
	if err != nil {
		t.Errorf("Encode err: %v", err)
		return
	}
	// decode
	resultBat, _, err := DecodeBatch(buf.Bytes())
	if err != nil {
		t.Errorf("Decode err: %v", err)
		return
	}
	// result verify
	// Attrs
	for i, v := range bat.Attrs {
		if resultBat.Attrs[i] != v {
			t.Errorf("Decode Batch.Attrs failed. \nExpected/Got:\n%s\n%s", v, resultBat.Attrs[1])
			return
		}
	}
	// Vecs
	if resultBat.Vecs[0].Col.([]float64)[0] != bat.Vecs[0].Col.([]float64)[0] {
		t.Errorf("Decode Batch.Vecs failed. \nExpected/Got:\n%v\n%v", bat.Vecs[0].Col.([]float64)[0], resultBat.Vecs[0].Col.([]float64)[0])
		return
	}
	if resultBat.Vecs[1].Col.([]float64)[0] != bat.Vecs[1].Col.([]float64)[0] {
		t.Errorf("Decode Batch.Vecs failed. \nExpected/Got:\n%v\n%v", bat.Vecs[1].Col.([]float64)[0], resultBat.Vecs[1].Col.([]float64)[0])
		return
	}
	// SelsData
	if string(resultBat.SelsData) != string(bat.SelsData) {
		t.Errorf("Decode Batch.SelsData failed. \nExpected/Got:\n%v\\n%v", string(bat.SelsData), string(resultBat.SelsData))
		return
	}
	// Sels
	for i, v := range bat.Sels {
		if resultBat.Sels[i] != v {
			t.Errorf("Decode Batch.Sels failed. \nExpected/Got:\n%v\n%v", v, resultBat.Sels[i])
			return
		}
	}
	// Zs
	for i, v := range bat.Zs {
		if resultBat.Zs[i] != v {
			t.Errorf("Decode Batch.Zs failed. \nExpected/Got:\n%v\n%v", v, resultBat.Zs[i])
			return
		}
	}
	// Refs
	for i, v := range bat.Refs {
		if resultBat.Refs[i] != v {
			t.Errorf("Decode Batch.Refs failed. \nExpected/Got:\n%v\n%v", v, resultBat.Refs[i])
			return
		}
	}
	// As
	for i, v := range bat.As {
		if resultBat.As[i] != v {
			t.Errorf("Decode Batch.As failed. \nExpected/Got:\n%v\n%v", v, resultBat.As[i])
			return
		}
	}
}
