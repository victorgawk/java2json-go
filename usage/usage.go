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
		fmt.Printf("error decoding base64: %s\n", err.Error())
		return
	}

	var obj interface{}

	// Usage with []byte
	obj, err = java2json.ParseJavaObject(javaObjectBytes)
	if err != nil {
		fmt.Printf("error parsing java object: %s\n", err.Error())
		return
	}

	printJson(obj)

	// Usage with io.Reader
	reader := bytes.NewReader(javaObjectBytes)
	jop := java2json.NewJavaObjectParser(reader)

	jop.SetMaxDataBlockSize(2048)                 // (optional) set max data block size
	jop.SetCycleReferenceValue("cycle reference") // (optional) set cycle reference value

	obj, err = jop.ParseJavaObject()
	if err != nil {
		fmt.Printf("error parsing java object: %s\n", err.Error())
		return
	}

	printJson(obj)
}

func printJson(obj interface{}) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("error marshalling JSON: %s\n", err.Error())
		return
	}
	fmt.Println(string(bytes))
}
