package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/ezodude/kube-guard/testutil"
)

var a app

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestMain(m *testing.M) {
	a = app{}
	fakeK8s := fake.NewSimpleClientset()

	a.k8s = fakeK8s
	a.initialize()

	code := m.Run()
	os.Exit(code)
}

func TestSubjectHasNoRoles(t *testing.T) {
	expected := `[
  {
    "subject": "unknown",
    "roles": null,
    "clusterroles": null
  }
]`

	rr := testutil.ServeFakeHTTP(
		a.router,
		"GET",
		"/api/v0.1/privilege/search",
		[]string{"Content-Type: application/json"},
		`{ "subjects": ["unknown"], "format": "json" }`,
	)
	testutil.AssertStatus(t, rr.Code, http.StatusOK)
	testutil.AssertEqual(t, rr.Body.String(), expected, "Content did not match.")
}
