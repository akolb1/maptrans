package maptrans

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MissingKeys returns list of fields missing from common.JMap
func MissingKeys(src map[string]interface{},
	fields []string) []string {
	result := []string{}
	for _, v := range fields {
		if _, found := src[v]; !found {
			result = append(result, v)
		}
	}
	return result
}

func TestMapMap(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"A1": "a1",
	}
	src := map[string]interface{}{
		"A1": "foo",
		"C1": "missing",
	}

	verifier := map[string]interface{}{
		"a1": "A1",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Nil(t, dst["c1"])
	assert.Equal(t, dst["a1"], "foo")
	_, err = IsSimilar(dst, src, verifier)
	assert.NoError(t, err)
}

func TestMapTranslation(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"E1": Description{
			TargetName: "e1",
			Type:       MapTranslation,
			SubTranslation: map[string]interface{}{
				"E11": "e11",
				"E12": "e12",
			},
		},
	}
	src := map[string]interface{}{
		"E1": map[string]interface{}{"E11": "is_e11", "E12": "is_e12"},
	}

	verifier := map[string]interface{}{
		"e1": Description{
			TargetName: "E1",
			Type:       MapTranslation,
			SubTranslation: map[string]interface{}{
				"e11": "E11",
				"e12": "E12",
			},
		},
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	subObj, ok := dst["e1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, subObj["e11"], "is_e11")
	assert.Equal(t, subObj["e12"], "is_e12")
	_, err = IsSimilar(dst, src, verifier)
	assert.NoError(t, err)
}

func TestMapArrayTranslation(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"M": Description{TargetName: "m",
			Type: MapArrayTranslation,
			SubTranslation: map[string]interface{}{
				"AA": "a",
				"BB": "b",
			},
		},
	}
	src := map[string]interface{}{
		"M": []map[string]interface{}{
			map[string]interface{}{"AA": "1", "BB": "1"},
			map[string]interface{}{"AA": "2", "BB": "3"},
		},
	}

	verifier := map[string]interface{}{
		"m": Description{
			TargetName: "M",
			Type:       MapArrayTranslation,
			SubTranslation: map[string]interface{}{
				"AA": "aa",
				"BB": "b",
			},
		},
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = IsSimilar(dst, src, verifier)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	subObj, ok := dst["m"].([]map[string]interface{})
	if !assert.True(t, ok) {
		t.FailNow()
	}
	if !assert.Equal(t, len(subObj), 2) {
		t.FailNow()
	}
	v0, ok := subObj[0]["a"].(string)
	assert.True(t, ok)
	assert.Equal(t, "1", v0)
	v1, ok := subObj[0]["b"].(string)
	assert.True(t, ok)
	assert.Equal(t, "1", v1)
	v2, ok := subObj[1]["a"].(string)
	assert.True(t, ok)
	assert.Equal(t, "2", v2)
	v3, ok := subObj[1]["b"].(string)
	assert.True(t, ok)
	assert.Equal(t, "3", v3)
}

func TestIdMapTranslation(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"G1": Description{TargetName: "g1", MapFunc: IDMap},
	}
	src := map[string]interface{}{
		"G1": map[string]interface{}{"a": "b"},
	}

	verifier := map[string]interface{}{
		"g1": Description{
			TargetName: "G1",
			Type:       MapTranslation,
			SubTranslation: map[string]interface{}{
				"a": "a",
			},
		},
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	_, err = IsSimilar(dst, src, verifier)
	assert.NoError(t, err)
	subObj, ok := dst["g1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, subObj["a"], "b")
}

func TestStringToLowerMapTranslation(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"G1": Description{
			TargetName: "g1",
			MapFunc:    StringToLowerMap,
		},
	}
	src := map[string]interface{}{
		"G1": "aBcD01",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, "abcd01", dst["g1"].(string))
}

func TestStringToUpperMapTranslation(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"G1": Description{
			TargetName: "g1",
			MapFunc:    StringToUpperMap,
		},
	}
	src := map[string]interface{}{
		"G1": "aBcD01",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, "ABCD01", dst["g1"].(string))
}

func TestModifyTranslation(t *testing.T) {
	t.Parallel()
	modFunc := func(_, o map[string]interface{}, v interface{}) error {
		x, ok := v.(int)
		assert.True(t, ok)
		o["foo"] = 2 * x
		return nil
	}
	descr := map[string]interface{}{
		"A": Description{
			TargetName: "a",
			Type:       ModifyTranslation,
			ModFunc:    modFunc,
		},
	}
	src := map[string]interface{}{"A": 2}
	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	x, ok := dst["foo"].(int)
	assert.True(t, ok)
	assert.Equal(t, 4, x)
}

func TestInsertTranslation(t *testing.T) {
	t.Parallel()
	const value = "Hello"
	insFunc := func(_, o map[string]interface{}, v string) (interface{}, error) {
		o["A"] = v
		return o, nil
	}
	descr := map[string]interface{}{
		value: Description{
			TargetName: "a",
			Type:       InsertTranslation,
			InsertFunc: insFunc,
		},
	}
	src := map[string]interface{}{"B": 2}
	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	x, ok := dst["A"].(string)
	assert.True(t, ok)
	assert.Equal(t, value, x)
}

// Test mapping with invalid value type
func TestMapMapBad(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"A1": "a1",
		"B1": Description{TargetName: "b1", MapFunc: StringMap},
	}
	src := map[string]interface{}{"A1": "foo", "B1": 1, "C1": "missing"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}

func TestBoolMap(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"A": Description{TargetName: "a", MapFunc: BoolMap},
		"B": Description{TargetName: "b", MapFunc: BoolMap},
		"C": Description{TargetName: "c", MapFunc: BoolMap},
		"D": Description{TargetName: "d", MapFunc: BoolMap},
		"E": Description{TargetName: "e", MapFunc: BoolMap},
		"F": Description{TargetName: "f", MapFunc: BoolMap},
	}
	src := map[string]interface{}{
		"A": "T",
		"B": false,
		"C": "F",
		"D": true,
		"E": "False",
		"F": "True",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, dst["a"].(bool))
	assert.True(t, dst["f"].(bool))
	assert.True(t, dst["d"].(bool))
	assert.False(t, dst["b"].(bool))
	assert.False(t, dst["c"].(bool))
	assert.False(t, dst["e"].(bool))
}

func TestBoolToStringMap(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"A": Description{TargetName: "a", MapFunc: BoolToStrMap},
		"B": Description{TargetName: "b", MapFunc: BoolToStrMap},
		"C": Description{TargetName: "c", MapFunc: BoolToStrMap},
		"D": Description{TargetName: "d", MapFunc: BoolToStrMap},
		"E": Description{TargetName: "e", MapFunc: BoolToStrMap},
		"F": Description{TargetName: "f", MapFunc: BoolToStrMap},
	}
	src := map[string]interface{}{
		"A": "T",
		"B": false,
		"C": "F",
		"D": true,
		"E": "False",
		"F": "True",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, "True", dst["a"].(string))
	assert.Equal(t, "False", dst["b"].(string))
	assert.Equal(t, "False", dst["c"].(string))
	assert.Equal(t, "True", dst["d"].(string))
	assert.Equal(t, "False", dst["e"].(string))
	assert.Equal(t, "True", dst["f"].(string))
}

func TestStringArray(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"F1": Description{TargetName: "f1", MapFunc: StringArrayMap},
	}
	src := map[string]interface{}{
		"F1": []string{"a", "b", "c"},
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, len(dst), 1)
	arrayObj, ok := dst["f1"].([]string)
	assert.True(t, ok)
	assert.Equal(t, len(arrayObj), 3)
	assert.Equal(t, arrayObj[0], "a")
	assert.Equal(t, arrayObj[1], "b")
	assert.Equal(t, arrayObj[2], "c")
}

func TestNullStringArray(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"F": Description{
			TargetName: "f",
			MapFunc:    StringArrayMap,
		},
	}
	src := map[string]interface{}{
		"F": nil,
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, len(dst), 1)
	arrayObj, ok := dst["f"]
	assert.True(t, ok)
	assert.Nil(t, arrayObj)
}

func TestIdentifier(t *testing.T) {
	t.Parallel()
	const m = "Hello0World"
	// Create description
	descr := map[string]interface{}{
		"H1": Description{TargetName: "h1", MapFunc: IdentifierMap},
	}
	src := map[string]interface{}{
		"H1": m,
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	val, ok := dst["h1"].(string)
	assert.True(t, ok)
	assert.Equal(t, m, val)
}

func TestUUID(t *testing.T) {
	t.Parallel()
	const m = "fc62e0eb-7969-5c24-b83f-955bf7f4ad0b"
	// Create description
	descr := map[string]interface{}{
		"A": Description{TargetName: "a", MapFunc: UUIDMap},
	}
	src := map[string]interface{}{
		"A": m,
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	val, ok := dst["a"].(string)
	assert.True(t, ok)
	assert.Equal(t, m, val)
}

func TestIp(t *testing.T) {
	t.Parallel()
	const a = "1.2.3.4"
	// Create description
	descr := map[string]interface{}{
		"J": Description{TargetName: "j", MapFunc: IPAddrMap},
	}
	src := map[string]interface{}{
		"J": a,
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	val, ok := dst["j"].(string)
	assert.True(t, ok)
	assert.Equal(t, a, val)
}

func TestCidr(t *testing.T) {
	t.Parallel()
	const a = "1.2.3.4/24"
	// Create description
	descr := map[string]interface{}{
		"A": Description{TargetName: "a", MapFunc: CIDRMap},
	}
	src := map[string]interface{}{
		"A": a,
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	val, ok := dst["a"].(string)
	assert.True(t, ok)
	assert.Equal(t, a, val)
}

func TestInteger(t *testing.T) {
	t.Parallel()
	// Create description
	descr := map[string]interface{}{
		"A": Description{TargetName: "a", MapFunc: IntegerMap},
		"B": Description{TargetName: "b", MapFunc: IntegerMap},
	}
	src := map[string]interface{}{
		"A": 1024,
		"B": "1024",
	}

	dst, err := Translate(src, descr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	val, ok := dst["a"].(string)
	assert.True(t, ok)
	assert.Equal(t, "1024", val)
	val, ok = dst["b"].(string)
	assert.True(t, ok)
	assert.Equal(t, "1024", val)
}

func TestMissingValues(t *testing.T) {
	t.Parallel()
	s := map[string]interface{}{"a1": 1, "b1": 2}
	expected := []string{"b1", "c", "d"}
	missing := MissingKeys(s, expected)
	assert.Equal(t, len(missing), 2)
	sort.Strings(missing)
	assert.Equal(t, missing[0], "c")
	assert.Equal(t, missing[1], "d")
}

func TestBadId(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"A1": Description{TargetName: "b1", MapFunc: IdentifierMap},
	}
	src := map[string]interface{}{"A1": "a$"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}

func TestBadIP(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"A1": Description{TargetName: "a1", MapFunc: IPAddrMap},
	}
	src := map[string]interface{}{"A1": "1.2.3.4/24"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}

func TestBadCIDR(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"A1": Description{TargetName: "a1", MapFunc: CIDRMap},
	}
	src := map[string]interface{}{"A1": "1.2.3.4/245"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}

func TestMandatoryOption(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"X": "Y",
		"A1": Description{TargetName: "a1",
			MapFunc:   StringMap,
			Mandatory: true,
		},
	}
	src := map[string]interface{}{"X": "X"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}

func TestInvalidNumbers(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"A": Description{TargetName: "a",
			MapFunc:   IntegerMap,
			Mandatory: true,
		},
	}
	src := map[string]interface{}{"A": "foo"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
	src1 := map[string]interface{}{"A": -1}
	_, err1 := Translate(src1, descr)
	assert.Error(t, err1, "Error expected")
}

func TestInvalidUUID(t *testing.T) {
	t.Parallel()
	descr := map[string]interface{}{
		"A": Description{TargetName: "a",
			MapFunc: UUIDMap,
		},
	}
	src := map[string]interface{}{"A": "cb89a4a9-7a7e-59ea-a0f2"}
	_, err := Translate(src, descr)
	assert.Error(t, err, "Error expected")
}
