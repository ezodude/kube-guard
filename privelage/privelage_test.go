package privelage_test

import (
	"path/filepath"
	"testing"
	"log"
	"io/ioutil"

	"github.com/ezodude/kube-guard/privelage"
	"github.com/ezodude/kube-guard/testutil"
	"k8s.io/client-go/kubernetes/fake"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func mustPrepRoleAndBinding(fakeK8s *fake.Clientset, rolePath, bindingPath string) {
	if err := testutil.ImportRoleBinding(fakeK8s, bindingPath); err != nil {
		panic(err)
	}
	if err := testutil.ImportRole(fakeK8s, rolePath); err != nil {
		panic(err)
	}
}

func TestSubjectHasRoleJsonFormatted(t *testing.T) {
	fakeK8s := fake.NewSimpleClientset()
	rolePath := filepath.Join("..", "testdata", "import-role.json")
	bindingPath := filepath.Join("..", "testdata", "import-rolebindings.json")
	mustPrepRoleAndBinding(fakeK8s, rolePath, bindingPath)

	cases := []struct {
		subjects     []string
		expectedPath string
	}{
		{subjects: []string{"developer"}, expectedPath: "dev-roles-res.json"},
		{subjects: []string{"deve*"}, expectedPath: "dev-roles-regexp-res.json"},
	}

	for _, tc := range cases {
		q := privelage.NewQuery().Client(fakeK8s).Subjects(tc.subjects).ResultFormat("JSON")

		expected := testutil.MustReadFile(t, filepath.Join("..", "testdata", tc.expectedPath))
		actual, err := q.Do()
		if err != nil {
			t.Fatalf("Could not execute query, err: %v\n", err)
		}

		testutil.AssertBytes(t, expected, actual, "Content did not match.")
	}
}

func TestSubjectHasRoleYamlFormatted(t *testing.T) {
	fakeK8s := fake.NewSimpleClientset()
	rolePath := filepath.Join("..", "testdata", "import-role.json")
	bindingPath := filepath.Join("..", "testdata", "import-rolebindings.json")
	mustPrepRoleAndBinding(fakeK8s, rolePath, bindingPath)

	subjects := []string{"developer"}
	q := privelage.NewQuery().Client(fakeK8s).Subjects(subjects).ResultFormat("YAML")

	expected := testutil.MustReadFile(t, filepath.Join("..", "testdata", "dev-roles-res.yaml"))
	actual, err := q.Do()
	if err != nil {
		t.Fatalf("Could not execute query, err: %v\n", err)
	}

	testutil.AssertBytes(t, expected, actual, "Content did not match.")
}
