package app

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Yu-Dojin/web16/model"
	sessions "github.com/gorilla/Sessions"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

//확인용 테스트2
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY"))) //os의 환경변수에서 SESSION_KEY 라는 환경변수를 가져와서 암호화 한다.
var rd *render.Render = render.New()

type AppHandler struct {
	http.Handler //포함타입. 암시적으로 멤버 베리어블의 이름을 안써준다.
	db           model.DBHandler
}

// 펑션 포인터를 값으로 받는 variable 이 된다. 사용은 func 일때와 다를게 없다. 함수콜하듯이 쓰면 됨. 테스트코드를 위해 변경.
var getSessionID = func(r *http.Request) string { //쿠키는 리쿼스트에 들어있기 때문에 리쿼스트를 받아와야 한다.
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}

	// Set some session values.
	val := session.Values["id"] //비어있는 경우 nil 이 된다.
	if val == nil {
		return ""
	}
	return val.(string) //val 을 string 으로 바꿔서 리턴한다.
}

func (a *AppHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todo.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) getTodoListHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r) // getTodoListHandler 가 불렸다는건 이미 CheckSignin 데코레이터를 통과 했다는 것이다. 이미 로그인이 된 상태라는 뜻. sessionId 가 nil 인 경우는 없다.
	list := a.db.GetTodos(sessionId)
	rd.JSON(w, http.StatusOK, list)
}

func (a *AppHandler) addTodoHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	name := r.FormValue("name")
	todo := a.db.AddTodo(name, sessionId)
	rd.JSON(w, http.StatusCreated, todo)
}

type Success struct {
	Success bool `json:"success"`
}

func (a *AppHandler) removeTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	ok := a.db.RemoveTodo(id)
	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) completeTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	complete := r.FormValue("complete") == "true"
	ok := a.db.CompleteTodo(id, complete)
	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) Close() {
	a.db.Close()
}

func CheckSignin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// if request URL is /signin.html, then next(). /signin.html 로 들어왔을때 계속 리다이렉트되는 무한루프 방지를 위해서.
	// (수정) /signin.html 로 하면 signin.css 가 걸려서 제대로 동작이 안되므로 /signin 으로 수정.
	if strings.Contains(r.URL.Path, "/signin") ||
		strings.Contains(r.URL.Path, "/auth") {
		next(w, r)
		return
	}

	// if user already signed in
	sessionID := getSessionID(r)
	if sessionID != "" {
		next(w, r) // 다음핸들로러 넘어간다.
		return
	}

	// if not user sign in
	// redirect signin.html
	http.Redirect(w, r, "/signin.html", http.StatusTemporaryRedirect)
}

//todoMap = make(map[int]*Todo)
func MakeHandler(filepath string) *AppHandler {
	r := mux.NewRouter()
	// negroni : 웹 핸들러에 로그등 기본적인 부가기능들을 데코레이팅 해주는 미들웨어. 핸들러를 래핑하여 여러 부가기능을 제공한다. 이경우 http 호출->negroni 호출->mux 호출.
	// signin 체크 때문에 main에서 여기로 옮김.
	// 여기선 Classic() 을 그대로 안쓰고 가운데 Signin 을 추가했다. 체인으로 되어있기 때문에 CheckSignin 데코레이터에서 걸리면 다음으로 못넘어가게 된다. 라우터로 들어가지 못하게 된다.
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.HandlerFunc(CheckSignin), negroni.NewStatic(http.Dir("public")))
	n.UseHandler(r)

	a := &AppHandler{
		Handler: n,
		db:      model.NewDBHandler(filepath),
	}

	r.HandleFunc("/todos", a.getTodoListHandler).Methods("GET")
	r.HandleFunc("/todos", a.addTodoHandler).Methods("POST")
	r.HandleFunc("/todos/{id:[0-9]+}", a.removeTodoHandler).Methods("DELETE")
	r.HandleFunc("/complete-todo/{id:[0-9]+}", a.completeTodoHandler).Methods("GET")
	r.HandleFunc("/auth/google/login", googleLoginHandler)
	r.HandleFunc("/auth/google/callback", googleAuthCallback)
	r.HandleFunc("/", a.indexHandler)

	return a
}
