package spaces

import (
	"context"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

// Of is GET /spaces for userID: the spaces they belong to, newest first.
func (s *Service) Of(ctx context.Context, userID int64) ([]schemas.Space, error) {
	rows, err := s.listSpaces(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, err
	}
	spaces := make([]schemas.Space, len(rows))
	for i, row := range rows {
		spaces[i] = row.Space
	}
	return spaces, nil
}
