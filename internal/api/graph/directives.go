package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

func LengthDirective(ctx context.Context, obj interface{}, next graphql.Resolver, min *int, max *int) (res interface{}, err error) {

	res, err = next(ctx)

	str, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("length directive can only be applied to strings")
	}

	if max != nil && len([]rune(str)) > *max {
		return nil, fmt.Errorf("field exceeds the maximum length of %d", max)
	}
	if min != nil && len([]rune(str)) < *min {
		return nil, fmt.Errorf("field is below the minimum length of %d", min)
	}

	return
}
