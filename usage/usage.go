package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/victorgawk/java2json-go/java2json"
)

func main() {
	javaObjectBase64 := "rO0ABXQADEhlbGxvLCBXb3JsZA=="
	javaObjectBytes, err := base64.StdEncoding.DecodeString(javaObjectBase64)
	if err != nil {
		panic(err)
	}

	var obj interface{}

	// Usage with []byte
	obj, err = java2json.ParseJavaObject(javaObjectBytes)
	if err != nil {
		panic(err)
	}

	printJson(obj)

	// Usage with io.Reader
	reader := bytes.NewReader(javaObjectBytes)
	jop := java2json.NewJavaObjectParser(reader)

	jop.SetMaxDataBlockSize(1024) // (optional) set max data block size

	obj, err = jop.ParseJavaObject()
	if err != nil {
		panic(err)
	}

	printJson(obj)
}

func printJson(obj interface{}) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}
