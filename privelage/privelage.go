package privelage

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//Query contains data to search for a subject roles / cluster roles and result formating (JSON or YAML)
type Query struct {
	client       kubernetes.Interface
	subjects     []string
	resultFormat string
}

//NewQuery query constructor
func NewQuery() *Query {
	return &Query{}
}

//Client configures the k8s client
func (q *Query) Client(c kubernetes.Interface) *Query {
	q.client = c
	return q
}

//Subjects configures the query subjects
func (q *Query) Subjects(s []string) *Query {
	q.subjects = s
	return q
}

//ResultFormat configures the query result format
func (q *Query) ResultFormat(f string) *Query {
	q.resultFormat = f
	return q
}

type hit struct {
	Roles        []*v1.Role
	ClusterRoles []*v1.ClusterRole
}

type result struct {
	Subject      string            `json:"subject"`
	Roles        []*v1.Role        `json:"roles"`
	ClusterRoles []*v1.ClusterRole `json:"clusterroles"`
}

func collectRoleHits(hits map[string]*hit, client kubernetes.Interface, rbindings *v1.RoleBindingList, querySubjects []string) {

	for _, item := range rbindings.Items {
		for _, sub := range item.Subjects {
			for _, qSub := range querySubjects {
				match, _ := regexp.MatchString(qSub, sub.Name)
				if match {
					role, err := client.RbacV1().Roles("default").Get(item.RoleRef.Name, metav1.GetOptions{})
					if err == nil {
						if _, ok := hits[qSub]; !ok {
							hits[qSub] = &hit{}
						}
						hits[qSub].Roles = append(hits[qSub].Roles, role)
					}
				}
			}
		}
	}
}

func calculateResults(hits map[string]*hit, querySubjects []string) []result {
	var results []result

	sort.Strings(querySubjects)

	for _, qSub := range querySubjects {
		if hit, ok := hits[qSub]; ok {
			results = append(results, result{
				Subject:      qSub,
				Roles:        hit.Roles,
				ClusterRoles: hit.ClusterRoles,
			})
		} else {
			results = append(results, result{
				Subject: qSub,
			})
		}
	}

	return results
}

//Do executes the query and formats the result
func (q *Query) Do() ([]byte, error) {
	rbindings, err := q.client.RbacV1().RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hits := make(map[string]*hit)

	collectRoleHits(hits, q.client, rbindings, q.subjects)

	results := calculateResults(hits, q.subjects)

	switch strings.ToLower(q.resultFormat) {
	case "yaml", "yml":
		return yaml.Marshal(results)
	default:
		return json.MarshalIndent(results, "", "  ")
	}
}
