// [实测结构，代码逻辑为演示]
// E2: MySQL 实现 —— UserRepository 接口的具体实现
package repository

import (
	"database/sql"

	"go-web-demo/model"
)

// MySQLUserRepo MySQL 实现
// 所有 SQL 集中在这一个文件 —— 改表结构只改这里
type MySQLUserRepo struct {
	db *sql.DB
}

func NewMySQLUserRepo(db *sql.DB) *MySQLUserRepo {
	return &MySQLUserRepo{db: db}
}

func (r *MySQLUserRepo) List() (users []model.User, err error) {
	rows, err := r.db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u model.User
		if err = rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return
		}
		users = append(users, u)
	}
	return
}

func (r *MySQLUserRepo) GetByID(id int64) (user model.User, err error) {
	err = r.db.QueryRow(
		"SELECT id, name, email FROM users WHERE id = ?", id,
	).Scan(&user.ID, &user.Name, &user.Email)
	return
}

func (r *MySQLUserRepo) Create(user *model.User) (err error) {
	result, err := r.db.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		user.Name, user.Email,
	)
	if err != nil {
		return
	}
	user.ID, err = result.LastInsertId()
	return
}

func (r *MySQLUserRepo) Update(user *model.User) (err error) {
	_, err = r.db.Exec(
		"UPDATE users SET name = ?, email = ? WHERE id = ?",
		user.Name, user.Email, user.ID,
	)
	return
}

func (r *MySQLUserRepo) Delete(id int64) (err error) {
	_, err = r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return
}
