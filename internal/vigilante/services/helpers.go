package services

import (
	"github.com/ahmetb/go-linq/v3"
	v1 "k8s.io/api/core/v1"
)

func FindNodeByName(nodes []*v1.Node, name string) *v1.Node {
	result := linq.From(nodes).WhereT(func(node *v1.Node) bool {
		return node.Name == name
	}).Single()

	if result != nil {
		return result.(*v1.Node)
	} else {
		return nil
	}
}
