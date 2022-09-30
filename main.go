package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/detail-project/{id}", blogDetail).Methods("GET")
	route.HandleFunc("/add-project", formAddBlog).Methods("GET")
	route.HandleFunc("/add-blog", middleware.UploadFile(addBlog)).Methods("POST")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-project/{id}", updateProject).Methods("POST")
	route.HandleFunc("/form-register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")
	route.HandleFunc("/form-login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server berjalan di port 8080")
	http.ListenAndServe("localhost:8080", route)
}

type Project struct {
	ID           int
	ProjectName  string
	Description  string
	StartDate    string
	EndDate      string
	Technologies []string
	Duration     string
	Image        string
	Author       string
	IsLogin      bool
}

type SessionData struct {
	IsLogin  bool
	Username string
}

var Data = SessionData{}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/index.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
	}

	if session.Values["IsLogin"] != true {
		data, _ := connection.Conn.Query(context.Background(), "SELECT tb_blog.id, tb_blog.name, description, technologies, duration, image FROM tb_blog ORDER BY id ASC")

		var result []Project
		for data.Next() {
			var each = Project{}

			err := data.Scan(&each.ID, &each.ProjectName, &each.Description, &each.Technologies, &each.Duration, &each.Image)
			if err != nil {
				w.Write([]byte(error.Error()))
				return
			}
			result = append(result, each)
		}
		resData := map[string]interface{}{
			"DataSession": Data,
			"Project":     result,
		}

		w.WriteHeader(http.StatusOK)
		template.Execute(w, resData)

	} else {

		sessionID := session.Values["ID"]

		data, _ := connection.Conn.Query(context.Background(), "SELECT tb_blog.id, tb_blog.name, description, technologies, duration, image FROM tb_blog WHERE tb_blog.author_id = $1 ORDER BY id ASC", sessionID)

		var result []Project
		for data.Next() {
			var each = Project{}

			err := data.Scan(&each.ID, &each.ProjectName, &each.Description, &each.Technologies, &each.Duration, &each.Image)
			if err != nil {
				w.Write([]byte(error.Error()))
				return
			}
			result = append(result, each)
		}
		resData := map[string]interface{}{
			"DataSession": Data,
			"Project":     result,
		}

		w.WriteHeader(http.StatusOK)
		template.Execute(w, resData)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/contact.html")

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
	}

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	data := map[string]interface{}{
		"DataSession": Data,
	}

	template.Execute(w, data)
}

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/detail-project.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
	}

	var BlogDetail = Project{}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	error = connection.Conn.QueryRow(context.Background(), "SELECT name, start_date, end_date, description, duration, technologies, image FROM tb_blog WHERE id=$1", id).Scan(&BlogDetail.ProjectName, &BlogDetail.StartDate, &BlogDetail.EndDate, &BlogDetail.Description, &BlogDetail.Duration, &BlogDetail.Technologies, &BlogDetail.Image)

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(error.Error()))
	}

	data := map[string]interface{}{
		"DataSession": Data,
		"BlogDetail":  BlogDetail,
	}

	template.Execute(w, data)
}

func formAddBlog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/add-project.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	data := map[string]interface{}{
		"DataSession": Data,
	}

	template.Execute(w, data)
}

func addBlog(w http.ResponseWriter, r *http.Request) {
	error := r.ParseForm()
	if error != nil {
		log.Fatal(error)
	}

	var duration string
	var projectName = r.PostForm.Get("projectName")
	var deskripsi = r.PostForm.Get("deskripsi")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var tech = r.Form["checkbox"]

	var layout = "2006-01-02"
	var startDateParse, _ = time.Parse(layout, startDate)
	var endDateParse, _ = time.Parse(layout, endDate)
	var startDateConvert = startDateParse.Format("02 Jan 2006")
	var endDateConvert = endDateParse.Format("02 Jan 2006")

	var hours = endDateParse.Sub(startDateParse).Hours()
	var days = hours / 24
	var weeks = math.Round(days / 7)
	var months = math.Round(days / 30)
	var years = math.Round(days / 365)

	if days >= 1 && days <= 6 {
		duration = strconv.Itoa(int(days)) + " day(s)"
	} else if days >= 7 && days <= 29 {
		duration = strconv.Itoa(int(weeks)) + " week(s)"
	} else if days >= 30 && days <= 364 {
		duration = strconv.Itoa(int(months)) + " month(s)"
	} else if days >= 365 {
		duration = strconv.Itoa(int(years)) + " year(s)"
	}

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	author := session.Values["ID"].(int)

	_, error = connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_blog(name, start_date, end_date, description, technologies, duration, image, author_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", projectName, startDateConvert, endDateConvert, deskripsi, tech, duration, image, author)

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(error.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_blog WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/edit-project.html")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var editProject = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, description FROM tb_blog WHERE id = $1", id).Scan(&editProject.ID, &editProject.ProjectName, &editProject.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	data := map[string]interface{}{
		"DataSession": Data,
		"EditProject": editProject,
	}

	tmpl.Execute(w, data)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var durationUpdate string
	var projectNameUpdate = r.PostForm.Get("projectName")
	var deskripsiUpdate = r.PostForm.Get("deskripsi")
	var startDateUpdate = r.PostForm.Get("startDate")
	var endDateUpdate = r.PostForm.Get("endDate")
	var techUpdate = r.Form["checkbox"]

	var layoutUpdate = "2006-01-02"
	var startDateParseUpdate, _ = time.Parse(layoutUpdate, startDateUpdate)
	var endDateParseUpdate, _ = time.Parse(layoutUpdate, endDateUpdate)
	var startDateConvert = startDateParseUpdate.Format("02 Jan 2006")
	var endDateConvertUpdate = endDateParseUpdate.Format("02 Jan 2006")

	var hoursUpdate = endDateParseUpdate.Sub(startDateParseUpdate).Hours()
	var daysUpdate = hoursUpdate / 24
	var weeksUpdate = math.Round(daysUpdate / 7)
	var monthsUpdate = math.Round(daysUpdate / 30)
	var yearsUpdate = math.Round(daysUpdate / 365)

	if daysUpdate >= 1 && daysUpdate <= 6 {
		durationUpdate = strconv.Itoa(int(daysUpdate)) + " day(s)"
	} else if daysUpdate >= 7 && daysUpdate <= 29 {
		durationUpdate = strconv.Itoa(int(weeksUpdate)) + " week(s)"
	} else if daysUpdate >= 30 && daysUpdate <= 364 {
		durationUpdate = strconv.Itoa(int(monthsUpdate)) + " month(s)"
	} else if daysUpdate >= 365 {
		durationUpdate = strconv.Itoa(int(yearsUpdate)) + " year(s)"
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	sqlStatement := `
	UPDATE public.tb_blog
	SET name=$7, description=$4, technologies=$5, duration=$6, start_date=$2, end_date=$3
	WHERE id=$1;
	`

	_, err = connection.Conn.Exec(context.Background(), sqlStatement, id, startDateConvert, endDateConvertUpdate, deskripsiUpdate, techUpdate, durationUpdate, projectNameUpdate)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-register.html")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	tmpl.Execute(w, Data)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var passwordRaw = r.PostForm.Get("inputPassword")

	password, _ := bcrypt.GenerateFromPassword([]byte(passwordRaw), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, password)
	if err != nil {
		http.Redirect(w, r, "/form-register", http.StatusMovedPermanently)
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-login.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(),
		"SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["ID"] = user.ID
	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 86400

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
