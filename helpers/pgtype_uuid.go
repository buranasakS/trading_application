package helpers

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)
func PgtypeUUID(t *testing.T, uuidStr string) pgtype.UUID {
	id := pgtype.UUID{}
	err := id.Scan(uuidStr)
	if err != nil {
		id.Valid = false
		return id
	}
	id.Valid = true
	return id
}


