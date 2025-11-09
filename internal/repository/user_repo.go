package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"azret/internal/model"
)

const getUserSQL = `SELECT user_id, deeplink, promo_message FROM users WHERE user_id=$1`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetByID fetches a user by ID.
// Returns (user, true, nil) when found; (_, false, nil) when not found.
func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, bool, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, getUserSQL, id).Scan(&u.UserID, &u.Deeplink, &u.PromoMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.User{}, false, nil
		}
		return model.User{}, false, err
	}
	return u, true, nil
}
