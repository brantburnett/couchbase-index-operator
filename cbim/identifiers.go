package cbim

import (
	"errors"
	"fmt"
	"strings"

	couchbasev1beta1 "github.com/brantburnett/couchbase-index-operator/api/v1beta1"
)

const (
	defaultScopeName      = "_default"
	defaultCollectionName = "_default"
)

// Uniquely defines a global secondary index.
type GlobalSecondaryIndexIdentifier struct {
	// Name of the index
	Name string
	// Name of the index's scope
	ScopeName string
	// Name of the index's collection
	CollectionName string
}

func defaultedName(name *string) string {
	if name == nil {
		return defaultScopeName
	}

	return *name
}

func GetIndexIdentifier(index couchbasev1beta1.GlobalSecondaryIndex) GlobalSecondaryIndexIdentifier {
	return GlobalSecondaryIndexIdentifier{
		Name:           index.Name,
		ScopeName:      defaultedName(index.ScopeName),
		CollectionName: defaultedName(index.CollectionName),
	}
}

func (identifier GlobalSecondaryIndexIdentifier) IsDefaultCollection() bool {
	return identifier.ScopeName == defaultScopeName && identifier.CollectionName == defaultCollectionName
}

func (identifier GlobalSecondaryIndexIdentifier) ToString() string {
	if identifier.IsDefaultCollection() {
		return identifier.Name
	}

	return fmt.Sprintf("%s.%s.%s", identifier.ScopeName, identifier.CollectionName, identifier.Name)
}

func ParseIndexIdentifierString(identifier string) (GlobalSecondaryIndexIdentifier, error) {
	if identifier == "" {
		return GlobalSecondaryIndexIdentifier{}, errors.New("invalid index identifier")
	}

	split := strings.Split(identifier, ".")
	if len(split) == 1 {
		return GlobalSecondaryIndexIdentifier{
			ScopeName:      defaultScopeName,
			CollectionName: defaultCollectionName,
			Name:           split[0],
		}, nil
	}

	if len(split) != 3 || split[0] == "" || split[1] == "" || split[2] == "" {
		return GlobalSecondaryIndexIdentifier{}, errors.New("invalid index identifier")
	}

	return GlobalSecondaryIndexIdentifier{
		ScopeName:      split[0],
		CollectionName: split[1],
		Name:           split[2],
	}, nil
}
