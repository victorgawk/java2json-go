package java2json

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestDate(t *testing.T) {
	input := "rO0ABXNyAA5qYXZhLnV0aWwuRGF0ZWhqgQFLWXQZAwAAeHB3CAAAAX/a+xS+eA=="
	expected := `"2022-03-30T10:19:22.302-03:00"`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestHashtable(t *testing.T) {
	input := "rO0ABXNyABNqYXZhLnV0aWwuSGFzaHRhYmxlE7sPJSFK5LgDAAJGAApsb2FkRmFjdG9ySQAJdGhyZXNob2xkeHA/QAAAAAAACHcIAAAACwAAAAN0AARrZXkzdAAEdmFsM3QABGtleTJ0AAR2YWwydAAEa2V5MXQABHZhbDF4"
	expected := `{"key1":"val1","key2":"val2","key3":"val3"}`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestHashMap(t *testing.T) {
	input := "rO0ABXNyABFqYXZhLnV0aWwuSGFzaE1hcAUH2sHDFmDRAwACRgAKbG9hZEZhY3RvckkACXRocmVzaG9sZHhwP0AAAAAAAAx3CAAAABAAAAADdAAEa2V5MXQABHZhbDF0AARrZXkydAAEdmFsMnQABGtleTN0AAR2YWwzeA=="
	expected := `{"key1":"val1","key2":"val2","key3":"val3"}`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestEnumMap(t *testing.T) {
	input := "rO0ABXNyABFqYXZhLnV0aWwuRW51bU1hcAZdffe+kHyhAwABTAAHa2V5VHlwZXQAEUxqYXZhL2xhbmcvQ2xhc3M7eHB2cgAWQmFzZTY0RW5jb2RlciRFbnVtVHlwZQAAAAAAAAAAEgAAeHIADmphdmEubGFuZy5FbnVtAAAAAAAAAAASAAB4cHcEAAAAA35xAH4AA3QABkVOVU1fQXQABHZhbDF+cQB+AAN0AAZFTlVNX0J0AAR2YWwyfnEAfgADdAAGRU5VTV9DdAAEdmFsM3g="
	expected := `{"ENUM_A":"val1","ENUM_B":"val2","ENUM_C":"val3"}`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestHashSet(t *testing.T) {
	input := "rO0ABXNyABFqYXZhLnV0aWwuSGFzaFNldLpEhZWWuLc0AwAAeHB3DAAAABA/QAAAAAAAA3QABGhzZTF0AARoc2UzdAAEaHNlMng="
	expected := `["hse1","hse3","hse2"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestCalendar(t *testing.T) {
	input := "rO0ABXNyABtqYXZhLnV0aWwuR3JlZ29yaWFuQ2FsZW5kYXKPPdfW5bDQwQIAAUoAEGdyZWdvcmlhbkN1dG92ZXJ4cgASamF2YS51dGlsLkNhbGVuZGFy5upNHsjcW44DAAtaAAxhcmVGaWVsZHNTZXRJAA5maXJzdERheU9mV2Vla1oACWlzVGltZVNldFoAB2xlbmllbnRJABZtaW5pbWFsRGF5c0luRmlyc3RXZWVrSQAJbmV4dFN0YW1wSQAVc2VyaWFsVmVyc2lvbk9uU3RyZWFtSgAEdGltZVsABmZpZWxkc3QAAltJWwAFaXNTZXR0AAJbWkwABHpvbmV0ABRMamF2YS91dGlsL1RpbWVab25lO3hwAQAAAAEBAQAAAAEAAAACAAAAAQAAAX/bR4RDdXIAAltJTbpgJnbqsqUCAAB4cAAAABEAAAABAAAH5gAAAAIAAAAOAAAABQAAAB4AAABZAAAABAAAAAUAAAAAAAAACwAAAAsAAAAqAAAAMwAAAkv/WzSAAAAAAHVyAAJbWlePIDkUuF3iAgAAeHAAAAARAQEBAQEBAQEBAQEBAQEBAQFzcgAYamF2YS51dGlsLlNpbXBsZVRpbWVab25l+mddYNFe9aYDABJJAApkc3RTYXZpbmdzSQAGZW5kRGF5SQAMZW5kRGF5T2ZXZWVrSQAHZW5kTW9kZUkACGVuZE1vbnRoSQAHZW5kVGltZUkAC2VuZFRpbWVNb2RlSQAJcmF3T2Zmc2V0SQAVc2VyaWFsVmVyc2lvbk9uU3RyZWFtSQAIc3RhcnREYXlJAA5zdGFydERheU9mV2Vla0kACXN0YXJ0TW9kZUkACnN0YXJ0TW9udGhJAAlzdGFydFRpbWVJAA1zdGFydFRpbWVNb2RlSQAJc3RhcnRZZWFyWgALdXNlRGF5bGlnaHRbAAttb250aExlbmd0aHQAAltCeHIAEmphdmEudXRpbC5UaW1lWm9uZTGz6fV3RKyhAgABTAACSUR0ABJMamF2YS9sYW5nL1N0cmluZzt4cHQAEUFtZXJpY2EvU2FvX1BhdWxvADbugAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAP9bNIAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB1cgACW0Ks8xf4BghU4AIAAHhwAAAADB8cHx4fHh8fHh8eH3cKAAAABgAAAAAAAHVxAH4ABgAAAAIAAAAAAAAAAHhzcgAac3VuLnV0aWwuY2FsZW5kYXIuWm9uZUluZm8k0dPOAB1xmwIACEkACGNoZWNrc3VtSQAKZHN0U2F2aW5nc0kACXJhd09mZnNldEkADXJhd09mZnNldERpZmZaABN3aWxsR01UT2Zmc2V0Q2hhbmdlWwAHb2Zmc2V0c3EAfgACWwAUc2ltcGxlVGltZVpvbmVQYXJhbXNxAH4AAlsAC3RyYW5zaXRpb25zdAACW0p4cQB+AAxxAH4AD7jHWBgAAAAA/1s0gAAAAAAAdXEAfgAGAAAABP9bNID/VUjg/5IjAAA27oBwdXIAAltKeCAEtRKxdZMCAAB4cAAAAF3/39rgHcAAAf/mSJ0A8gAA/+5vu4kwADL/7qnURxAAAP/u5WM9uAAy/+8fT1nQAAD/9sbWhrgAMv/28pyUuAAA//c8UZl4ADL/92NAQlAAAP/3scysOAAy//fZDbrQAAD/+CeaJLgAMv/4RI57UAAA//0n+z44ADL//VHPetAAAP/9vfh1uAAy//3Q8noQAAD//h/RSbgAMv/+PMWgUAAA//6LpG/4ADL//rJAsxAAAP//AR+CuAAy//8oDiuQAAAAB0W1NrgAMgAHcICkkAAAAAe4nRt4ADIAB9ymMJAAAAAILhguOAAyAAhP4HsQAAAACKEAEvgAMgAIwshf0AAAAAkWKL/4ADIACTxynVAAAAAJjZI1OAAyAAmz3BKQAAAACgK64jgAMgAKJsP3UAAAAAp6JFd4ADIACpmr3BAAAAAK7Qw8OAAyAAsVluHQAAAAC2I06TgAMgALir+O0AAAAAvXXZY4ADIAC/2nc5AAAAAMSkV6+AAyAAx1EOjQAAAADL/AjbgAMgAM7rsmUAAAAA021504ADIADWGjCxAAAAANqb+B+AAyAA3ZDIBQAAAADiEo9zgAMgAOS/RlEAAAAA6Ykmx4ADIADsEdEhAAAAAPFH1yOAAyAA82Rb8QAAAAD4UkjrgAMgAPq25sEAAAAA//c5e4ADIAECLX4VAAAAAQb3XouAAyABCYAI5QAAAAEOtg7ngAMgARD2oDkAAAABFZx0K4ADIAEYJR6FAAAAAR0TC3+AAyABH3epVQAAAAEkZZZPgAMgASbuQKkAAAABK7ghH4ADIAEuQMt5AAAAATMKq++AAyABNbdizQAAAAE6gUNDgAMgATzl4RkAAAABQdPOE4ADIAFEOGvpAAAAAUkmWOOAAyABS68DPQAAAAFQeOOzgAMgAVMBjg0AAAABV8tug4ADIAFaVBjdAAAAAV8d+VOAAyABYaajrQAAAAFm3KmvgAMgAWj5Ln0AAAAB7EuPa4AAB4///04vlkrAA="
	expected := `"2022-03-30T11:42:51.587-03:00"`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestArraysArrayList(t *testing.T) {
	input := "rO0ABXNyABpqYXZhLnV0aWwuQXJyYXlzJEFycmF5TGlzdNmkPL7NiAbSAgABWwABYXQAE1tMamF2YS9sYW5nL09iamVjdDt4cHVyABNbTGphdmEubGFuZy5TdHJpbmc7rdJW5+kde0cCAAB4cAAAAAN0AAVlbGVtMXQABWVsZW0ydAAFZWxlbTM="
	expected := `["elem1","elem2","elem3"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestArrayList(t *testing.T) {
	input := "rO0ABXNyABNqYXZhLnV0aWwuQXJyYXlMaXN0eIHSHZnHYZ0DAAFJAARzaXpleHAAAAADdwQAAAADdAAFZWxlbTF0AAVlbGVtMnQABWVsZW0zeA=="
	expected := `["elem1","elem2","elem3"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestArrayDeque(t *testing.T) {
	input := "rO0ABXNyABRqYXZhLnV0aWwuQXJyYXlEZXF1ZSB82i4kDaCLAwAAeHB3BAAAAAN0AAJlMXQAAmUydAACZTN4"
	expected := `["e1","e2","e3"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestArray(t *testing.T) {
	input := "rO0ABXVyABNbTGphdmEubGFuZy5PYmplY3Q7kM5YnxBzKWwCAAB4cAAAAAN0AAVlbGVtMXQABWVsZW0ydAAFZWxlbTM="
	expected := `["elem1","elem2","elem3"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestCollSer(t *testing.T) {
	input := "rO0ABXNyABFqYXZhLnV0aWwuQ29sbFNlcleOq7Y6G6gRAwABSQADdGFneHAAAAABdwQAAAADdAAFZWxlbTF0AAVlbGVtMnQABWVsZW0zeA=="
	expected := `["elem1","elem2","elem3"]`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestCompose1(t *testing.T) {
	input := "rO0ABXNyABlCYXNlNjRFbmNvZGVyJDFPYmpldG9KYXZhA2D37c6rQAoCAARJAA1udW1iZXJFeGFtcGxlWwAMYXJyYXlFeGFtcGxldAATW0xqYXZhL2xhbmcvT2JqZWN0O0wAC2RhdGFFeGFtcGxldAAQTGphdmEvdXRpbC9EYXRlO0wADXN0cmluZ0V4YW1wbGV0ABJMamF2YS9sYW5nL1N0cmluZzt4cAAAAHt1cgATW0xqYXZhLmxhbmcuT2JqZWN0O5DOWJ8QcylsAgAAeHAAAAADdAAGYXJyIGUxdAAGYXJyIGUydAAGYXJyIGUzc3IADmphdmEudXRpbC5EYXRlaGqBAUtZdBkDAAB4cHcIAAABf9snj5t4dAAMc3RyaW5nIHZhbHVl"
	expected := `{"arrayExample":["arr e1","arr e2","arr e3"],"dataExample":"2022-03-30T11:07:57.339-03:00","numberExample":123,"stringExample":"string value"}`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func TestCompose2(t *testing.T) {
	input := "rO0ABXNyABJCYXNlNjRFbmNvZGVyJDFPYmqIcPwzv07pKgIAAUwABG1hcGF0AA9MamF2YS91dGlsL01hcDt4cHNyABFqYXZhLnV0aWwuSGFzaE1hcAUH2sHDFmDRAwACRgAKbG9hZEZhY3RvckkACXRocmVzaG9sZHhwP0AAAAAAAAx3CAAAABAAAAAGfnIAF0Jhc2U2NEVuY29kZXIkUEFSQU1FVEVSAAAAAAAAAAASAAB4cgAOamF2YS5sYW5nLkVudW0AAAAAAAAAABIAAHhwdAAOT1NfRVhURVJOQUxfSTNzcgATamF2YS51dGlsLkFycmF5TGlzdHiB0h2Zx2GdAwABSQAEc2l6ZXhwAAAAAHcEAAAAAHh+cQB+AAV0AA5PU19FWFRFUk5BTF9JNnVyABNbTGphdmEubGFuZy5PYmplY3Q7kM5YnxBzKWwCAAB4cAAAAAJzcgARamF2YS5sYW5nLkludGVnZXIS4qCk94GHOAIAAUkABXZhbHVleHIAEGphdmEubGFuZy5OdW1iZXKGrJUdC5TgiwIAAHhwAAAByHQAA1NUUn5xAH4ABXQADk9TX0VYVEVSTkFMX0k1dXEAfgANAAAAAH5xAH4ABXQADk9TX0VYVEVSTkFMX0kxc3EAfgAJAAAAAXcEAAAAAXQABkkxIHN0cnh+cQB+AAV0AA5PU19FWFRFUk5BTF9JMnNyABFqYXZhLnV0aWwuSGFzaFNldLpEhZWWuLc0AwAAeHB3DAAAABA/QAAAAAAAAXNxAH4ADwAAAHt4fnEAfgAFdAAOT1NfRVhURVJOQUxfSTRzcQB+ABx3DAAAABA/QAAAAAAAAHh4"
	expected := `{"mapa":{"OS_EXTERNAL_I1":["I1 str"],"OS_EXTERNAL_I2":[123],"OS_EXTERNAL_I3":[],"OS_EXTERNAL_I4":[],"OS_EXTERNAL_I5":[],"OS_EXTERNAL_I6":[456,"STR"]}}`
	output := base64Java2Json(input)
	if output != expected {
		t.Errorf("%s != %s", output, expected)
	}
}

func base64Java2Json(b64str string) string {
	bytes, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		panic(err)
	}

	obj, err := ParseJavaObject(bytes)
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(data)
}
