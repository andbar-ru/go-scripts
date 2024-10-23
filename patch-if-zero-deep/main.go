package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
)

type Tree map[string]FirstLevel

type FirstLevel struct {
	StructKey1 SecondLevel `json:"structKey1"`
	StructKey2 SecondLevel `json:"structKey2"`
	StructKey3 SecondLevel `json:"structKey3"`
	BoolKey    bool        `json:"boolKey"`
	IntKey     int         `json:"intKey"`
	StringKey  string      `json:"stringKey"`
}

type SecondLevel struct {
	StructKey1 ThirdLevel `json:"structKey1"`
	StructKey2 ThirdLevel `json:"structKey2"`
	StructKey3 ThirdLevel `json:"structKey3"`
	BoolKey    bool       `json:"boolKey"`
	IntKey     int        `json:"intKey"`
	StringKey  string     `json:"stringKey"`
}

type ThirdLevel struct {
	BoolKey   bool   `json:"boolKey"`
	IntKey    int    `json:"intKey"`
	StringKey string `json:"stringKey"`
}

func randomBool(nonEmpty bool) bool {
	if !nonEmpty && rand.Float64() < 0.5 {
		return false
	}
	return true
}

func randomInt(nonEmpty bool) int {
	var zero int
	if !nonEmpty && rand.Float64() < 0.5 {
		return zero
	}
	return rand.Intn(10) + 1
}

func randomString(nonEmpty bool) string {
	var zero string
	if !nonEmpty && rand.Float64() < 0.5 {
		return zero
	}
	return string(rand.Intn(26) + 97)
}

func genThirdLevel(nonEmpty bool) ThirdLevel {
	return ThirdLevel{
		BoolKey:   randomBool(nonEmpty),
		IntKey:    randomInt(nonEmpty),
		StringKey: randomString(nonEmpty),
	}
}

func genSecondLevel(nonEmpty bool) SecondLevel {
	return SecondLevel{
		StructKey1: genThirdLevel(nonEmpty),
		StructKey2: genThirdLevel(nonEmpty),
		StructKey3: genThirdLevel(nonEmpty),
		BoolKey:    randomBool(nonEmpty),
		IntKey:     randomInt(nonEmpty),
		StringKey:  randomString(nonEmpty),
	}
}

func genFirstLevel(nonEmpty bool) FirstLevel {
	return FirstLevel{
		StructKey1: genSecondLevel(nonEmpty),
		StructKey2: genSecondLevel(nonEmpty),
		StructKey3: genSecondLevel(nonEmpty),
		BoolKey:    randomBool(nonEmpty),
		IntKey:     randomInt(nonEmpty),
		StringKey:  randomString(nonEmpty),
	}
}

func genTree() Tree {
	tree := make(Tree, 26)
	for i := 97; i <= 122; i++ {
		key := string(i)
		tree[key] = genFirstLevel(false)
	}
	return tree
}

func patch(value reflect.Value, key string, defaults map[string]any) {
	if value.Kind() != reflect.Struct {
		panic(fmt.Sprintf("path: %s, %v is not a struct", key, value))
	}

	for i := 0; i < value.NumField(); i++ {
		name := value.Type().Field(i).Name
		field := value.Field(i)
		newKey := key + "." + name
		if field.Kind() == reflect.Struct {
			patch(field, newKey, defaults)
		} else if field.IsZero() {
			v := defaults[newKey]
			field.Set(reflect.ValueOf(v))
		}
	}
}

func patchTree(tree Tree, defaults map[string]any) {
	for key, value := range tree {
		patch(reflect.ValueOf(&value).Elem(), "", defaults)
		tree[key] = value
	}
}

func flatten(value reflect.Value, key string, m map[string]any) {
	if value.Kind() != reflect.Struct {
		panic(fmt.Sprintf("path: %s, %v is not a struct", key, value))
	}

	for i := 0; i < value.NumField(); i++ {
		name := value.Type().Field(i).Name
		field := value.Field(i)
		newKey := key + "." + name
		if field.Kind() == reflect.Struct {
			flatten(field, newKey, m)
		} else {
			m[newKey] = field.Interface()
		}
	}
}

func main() {
	tree := genTree()
	defaults := genFirstLevel(true)
	defaultsMap := make(map[string]any)
	flatten(reflect.ValueOf(defaults), "", defaultsMap)

	// for key, value := range defaultsMap {
	// 	fmt.Printf("key: %s, type: %T, value: %v\n", key, value, value)
	// }

	patchTree(tree, defaultsMap)

	tree["_defaults"] = defaults
	jsonData, err := json.Marshal(tree)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
