package model

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3" //명시적으로 사용하진 않지만 위에 임포트해둔 sql 사용시 이것을 사용하겠다는 표시라고 할 수 있다.
)

type sqliteHandler struct {
	db *sql.DB
}

func (s *sqliteHandler) GetTodos(sessionId string) []*Todo {
	todos := []*Todo{}
	rows, err := s.db.Query("SELECT id, name, completed, createdAt FROM todos WHERE sessionId=?", sessionId) //sessionId 를 이용시 모든레코드를 검사하기 때문에 테이블 생성시 해당 항목에 인덱스도 같이 생성.
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() { //다음행이 있으면 계속 돌면서 레코드를 읽어온다.
		var todo Todo
		rows.Scan(&todo.ID, &todo.Name, &todo.Completed, &todo.CreatedAt)
		todos = append(todos, &todo)
	}
	return todos
}

func (s *sqliteHandler) AddTodo(name string, sessionId string) *Todo {
	stmt, err := s.db.Prepare("INSERT INTO todos (sessionId, name, completed, createdAt) VALUES (?, ?, ?, datetime('now'))") //스테이트먼트 작성.
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(sessionId, name, false) //위 쿼리문에서 ? 로 표시한 아규먼트를 적어준다.
	if err != nil {
		panic(err)
	}
	id, _ := rst.LastInsertId()
	var todo Todo
	todo.ID = int(id)
	todo.Name = name
	todo.Completed = false
	todo.CreatedAt = time.Now()
	return &todo
}

func (s *sqliteHandler) RemoveTodo(id int) bool {
	stmt, err := s.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(id)
	if err != nil {
		panic(err)
	}
	cnt, _ := rst.RowsAffected()
	return cnt > 0
}

func (s *sqliteHandler) CompleteTodo(id int, complete bool) bool {
	stmt, err := s.db.Prepare("UPDATE todos SET completed=? WHERE id=?")
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(complete, id)
	if err != nil {
		panic(err)
	}
	cnt, _ := rst.RowsAffected()
	return cnt > 0
}

func (s *sqliteHandler) Close() {
	s.db.Close()
}

func newSqliteHandler(filepath string) DBHandler {
	//콘피규레이션에 관련된 항목을 코드에 밖아넣는 것은 별로 좋은것이 아니다. 그래서 filepath 로 변수를 넣고 핸들러를 부르는 여러단계를 거슬러 메인에 이름을 박았다.
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}

	statement, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id	INTEGER PRIMARY KEY AUTOINCREMENT,
			sessionId STRING,
			name TEXT,
			completed BOOLEAN,
			createdAt DATETIME
		);
		CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos ON todos (
			sessionId ASC
		);`)
	statement.Exec()
	return &sqliteHandler{db: database}
}
