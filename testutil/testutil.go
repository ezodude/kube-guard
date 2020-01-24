package testutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func k8sSchemeDecode(path string) (interface{}, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(content), nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// ImportRoleBinding imports rolebinding into a fake k8s instance
func ImportRoleBinding(k8s *fake.Clientset, path string) error {
	obj, err := k8sSchemeDecode(path)
	if err != nil {
		return err
	}

	if _, err := k8s.RbacV1().RoleBindings("default").Create(obj.(*v1.RoleBinding)); err != nil {
		return err
	}

	return nil
}

// ImportRole imports a role into a fake k8s instance
func ImportRole(k8s *fake.Clientset, path string) error {
	obj, err := k8sSchemeDecode(path)
	if err != nil {
		return err
	}

	if _, err := k8s.RbacV1().Roles("default").Create(obj.(*v1.Role)); err != nil {
		return err
	}

	return nil
}

// ServeFakeHTTP creates a fake request and records the response from a relevant route.
func ServeFakeHTTP(router *mux.Router, method, url string, headers []string, body ...string) *httptest.ResponseRecorder {
	var b string
	if len(body) > 0 {
		b = body[0]
	}

	req, err := http.NewRequest(method, url, strings.NewReader(b))
	if err != nil {
		panic(err)
	}

	for _, header := range headers {
		parts := strings.Split(header, ":")
		req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// MustReadFile fails the test if a file cannot be read
func MustReadFile(tb testing.TB, filename string, v ...interface{}) []byte {
	result, err := ioutil.ReadFile(filename)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf("Cannot read filename[%s]\n", filename)
		tb.Fatalf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
	}
	return result
}

// AssertBytes asserts that 2 []byte are equal
func AssertBytes(tb testing.TB, expected []byte, actual []byte, msg string, v ...interface{}) {
	e := strings.Split(strings.TrimSpace(string(expected)), "")
	a := strings.Split(strings.TrimSpace(string(actual)), "")

	sort.Strings(e)
	sort.Strings(a)

	condition := reflect.DeepEqual(e, a)

	if !condition {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("\033%s:%d: "+msg+"\033\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
	}
}

// AssertStatus asserts that the HTTP status code is as expected
func AssertStatus(tb testing.TB, actual int, expected int, v ...interface{}) {
	if actual != expected {
		_, file, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf("wrong status code: got %v want %v\n", actual, expected)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// AssertEqual asserts that 2 strings are equal
func AssertEqual(tb testing.TB, actual, expected, msg string, v ...interface{}) {
	if !strings.EqualFold(actual, expected) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}
