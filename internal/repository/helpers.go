package repository

import (
	"database/sql"
	"errors"
)

// HandleNotFound processes a database query result, converting sql.ErrNoRows
// to a nil result without error. This is a common pattern for Find* operations
// where a missing row is not an error condition.
//
// Usage:
//
//	var item model.Item
//	err := r.db.GetContext(ctx, &item, query, args...)
//	return HandleNotFound(&item, err)
func HandleNotFound[T any](result *T, err error) (*T, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
