package database

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
	GetByID(DataHandler, int64) (*models.User, error)
	GetByNameOrEmail(DataHandler, string) (*models.User, error)
	IsName(DataHandler, string) (bool, error)
	IsEmail(DataHandler, string, int64) (bool, error)
}

type UserDB interface {
	UserReader
	UserWriter
}

// UserDB handles all user queries
type UserSqlDB struct {
	UserReader
	UserWriter
}

func newUserDB() UserDB {
	return &UserSqlDB{
		UserWriter: &UserSqlWriter{},
		UserReader: &UserSqlReader{},
	}
}

type UserSqlWriter struct{}

func (w *UserSqlWriter) UpdateEmail(dh DataHandler, email string, userID int64) error {
	q := "UPDATE users SET email=$1 WHERE id=$2"
	_, err := dh.Exec(q, email, userID)
	return dbErr(err, q)
}

func (w *UserSqlWriter) UpdatePassword(dh DataHandler, password string, userID int64) error {
	q := "UPDATE users SET password=$1 WHERE id=$2"
	_, err := dh.Exec(q, password, userID)
	return dbErr(err, q)
}
func (w *UserSqlWriter) Create(dh DataHandler, user *models.User) error {
	q := "INSERT INTO users(name, email, password) VALUES (:name, :email, :password) RETURNING id"
	q, args, err := sqlx.Named(q, user)
	if err != nil {
		return dbErr(err, q)
	}
	return dbErr(dh.QueryRowx(dh.Rebind(q), args...).Scan(&user.ID), q)
}

func (w *UserSqlWriter) DeleteUser(dh DataHandler, userID int64) error {
	q := "DELETE FROM users WHERE id=$1"
	_, err := dh.Exec(q, userID)
	return dbErr(err, q)
}

type UserSqlReader struct{}

func (r *UserSqlReader) GetByID(dh DataHandler, id int64) (*models.User, error) {
	q := "SELECT * FROM users WHERE id=$1"
	user := &models.User{}
	err := sqlx.Get(dh, user, q, id)
	return user, dbErr(err, q)
}

func (r *UserSqlReader) GetByNameOrEmail(dh DataHandler, identifier string) (*models.User, error) {
	q := "SELECT * FROM users WHERE email=$1 or name=$1"
	user := &models.User{}
	err := sqlx.Get(dh, user, q, identifier)
	return user, dbErr(err, q)
}

func (r *UserSqlReader) IsName(dh DataHandler, name string) (bool, error) {
	q := "SELECT COUNT(id) FROM users WHERE name=$1"
	var count int64
	if err := dh.QueryRowx(q, name).Scan(&count); err != nil {
		return false, dbErr(err, q)
	}
	return count > 0, nil

}

func (r *UserSqlReader) IsEmail(dh DataHandler, email string, userID int64) (bool, error) {

	q := "SELECT COUNT(id) FROM users WHERE email ILIKE $1"
	args := []interface{}{email}

	if userID != 0 {
		q += " AND id != $2"
		args = append(args, userID)
	}

	var count int64

	if err := dh.QueryRowx(q, args...).Scan(&count); err != nil {
		return false, dbErr(err, q)
	}
	return count > 0, nil
}
