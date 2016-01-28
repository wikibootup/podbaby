package store

import (
	"github.com/danjac/podbaby/models"
	"github.com/jmoiron/sqlx"
)

type UserWriter interface {
	Create(DataHandler, *models.User) error
	UpdateEmail(DataHandler, string, int64) error
	UpdatePassword(DataHandler, string, int64) error
	DeleteUser(DataHandler, int64) error
}

type UserReader interface {
	GetByID(DataHandler, *models.User, int64) error
	GetByNameOrEmail(DataHandler, *models.User, string) error
	IsName(DataHandler, string) (bool, error)
	IsEmail(DataHandler, string, int64) (bool, error)
}

type UserStore interface {
	UserReader
	UserWriter
}

// UserStore handles all user queries
type userSqlStore struct {
	UserReader
	UserWriter
}

func newUserStore() UserStore {
	return &userSqlStore{
		UserWriter: &userSqlWriter{},
		UserReader: &userSqlReader{},
	}
}

type userSqlWriter struct{}

func (w *userSqlWriter) UpdateEmail(dh DataHandler, email string, userID int64) error {
	q := "UPDATE users SET email=$1 WHERE id=$2"
	_, err := dh.Exec(q, email, userID)
	return err
}

func (w *userSqlWriter) UpdatePassword(dh DataHandler, password string, userID int64) error {
	q := "UPDATE users SET password=$1 WHERE id=$2"
	_, err := dh.Exec(q, password, userID)
	return err
}
func (w *userSqlWriter) Create(dh DataHandler, user *models.User) error {
	q := "INSERT INTO users(name, email, password) VALUES (:name, :email, :password) RETURNING id"
	q, args, err := sqlx.Named(q, user)
	if err != nil {
		return err
	}
	return dh.QueryRowx(dh.Rebind(q), args...).Scan(&user.ID)
}

func (w *userSqlWriter) DeleteUser(dh DataHandler, userID int64) error {
	q := "DELETE FROM users WHERE id=$1"
	_, err := dh.Exec(q, userID)
	return err
}

type userSqlReader struct{}

func (r *userSqlReader) GetByID(dh DataHandler, user *models.User, id int64) error {
	q := "SELECT * FROM users WHERE id=$1"
	return sqlx.Get(dh, user, q, id)
}

func (r *userSqlReader) GetByNameOrEmail(dh DataHandler, user *models.User, identifier string) error {
	q := "SELECT * FROM users WHERE email=$1 or name=$1"
	return sqlx.Get(dh, user, q, identifier)
}

func (r *userSqlReader) IsName(dh DataHandler, name string) (bool, error) {
	q := "SELECT COUNT(id) FROM users WHERE name=$1"
	var count int64
	if err := dh.QueryRowx(q, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil

}

func (r *userSqlReader) IsEmail(dh DataHandler, email string, userID int64) (bool, error) {

	q := "SELECT COUNT(id) FROM users WHERE email ILIKE $1"
	args := []interface{}{email}

	if userID != 0 {
		q += " AND id != $2"
		args = append(args, userID)
	}

	var count int64

	if err := dh.QueryRowx(q, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
