package main

import (
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/xhuliodo/todos_hat_stack/internal"
	"github.com/xhuliodo/todos_hat_stack/views/components"
	"github.com/xhuliodo/todos_hat_stack/views/pages"
)

func main() {
	internal.ConfigureLogger()
	internal.InitDB()

	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.HEAD("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// server static files
	e.Static("/static", "assets")

	e.GET("/", func(c echo.Context) error {
		return render(c, pages.Home(false, internal.User{}))
	})

	e.GET("/todos", func(c echo.Context) error {
		u := internal.GetUser(c)
		todos, err := internal.DB.Query("SELECT * FROM todos WHERE user_id = ? ORDER BY created_at DESC", u.Id)
		if err != nil {
			log.Println("Error fetching todos:", err)
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}
		defer todos.Close()

		resp := []internal.Todo{}
		for todos.Next() {
			var id int
			var content string
			var createdAt time.Time
			var updatedAt time.Time
			var completed bool
			err := todos.Scan(&id, &content, &createdAt, &updatedAt, &completed, &u.Id)
			if err != nil {
				log.Println("Error scanning todos:", err)
				return c.String(http.StatusInternalServerError, "Internal Server Error")
			}
			resp = append(resp, internal.Todo{Id: id, Content: content, CreatedAt: createdAt, UpdatedAt: updatedAt, Completed: completed})
		}

		return render(c, pages.Todos(true, u, resp))
	}, internal.AuthMiddleware)

	e.POST("/todos", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		u := internal.GetUser(c)

		content := c.FormValue("content")
		if content == "" {
			return c.JSON(http.StatusBadRequest, "todo's content cannot be empty")
		}
		if len(content) > 255 {
			return c.JSON(http.StatusBadRequest, "todo's content cannot be longer than 255 characters")
		}

		// save the todo
		res, err := internal.DB.Exec("INSERT INTO todos (content, user_id) VALUES (?, ?)", content, u.Id)
		if err != nil {
			log.Println("Error inserting todo:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
		rows, err := res.RowsAffected()
		if err != nil {
			log.Println("Error getting rows affected:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
		if rows != 1 {
			log.Println("Error inserting todo: no rows affected")
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		return render(c, components.Todo(internal.Todo{Content: content}))
	}, internal.AuthMiddleware)

	e.DELETE("/todos/:id", func(c echo.Context) error {
		u := internal.GetUser(c)

		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, "todo's id cannot be empty")
		}

		// delete the todo
		res, err := internal.DB.Exec("DELETE FROM todos WHERE id = ? and user_id = ?", id, u.Id)
		if err != nil {
			log.Println("Error deleting todo:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		rows, err := res.RowsAffected()
		if err != nil {
			log.Println("Error getting rows affected:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
		if rows != 1 {
			log.Println("Error deleting todo: no rows affected")
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		return c.HTML(http.StatusOK, "")
	}, internal.AuthMiddleware)

	e.PATCH("/todos/:id/toggle", func(c echo.Context) error {
		u := internal.GetUser(c)

		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, "todo's id cannot be empty")
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Println("Error converting id to int:", err)
			return c.JSON(http.StatusBadRequest, "todo's id must be an integer")
		}

		// toggle the todo
		res, err := internal.DB.Exec("UPDATE todos SET completed = NOT completed WHERE id = ? and user_id = ?", id, u.Id)
		if err != nil {
			log.Println("Error toggling todo:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		rows, err := res.RowsAffected()
		if err != nil {
			log.Println("Error getting rows affected:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
		if rows != 1 {
			log.Println("Error toggling todo: no rows affected")
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		todo := internal.DB.QueryRow("SELECT completed FROM todos WHERE id = ?", idInt)
		var completed bool
		err = todo.Scan(&completed)
		if err != nil {
			log.Println("Error scanning todo:", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		return render(c, components.TodoCheckbox(idInt, completed))
	}, internal.AuthMiddleware)

	e.GET("/sign-up", func(c echo.Context) error {
		f := internal.SignUpForm{}
		return render(c, pages.SignUp(false, internal.User{}, f))
	})

	e.POST("/sign-up", func(c echo.Context) error {
		f := internal.SignUpForm{}

		f.Name = c.FormValue("name")
		f.Email = c.FormValue("email")
		password := c.FormValue("password")

		if f.Name == "" || len(f.Name) > 255 {
			f.NameError = "Name cannot be empty or longer than 255 characters"
			f.HasError = true
		}
		if _, err := mail.ParseAddress(f.Email); err != nil {
			f.EmailError = "Email cannot be empty or invalid"
			f.HasError = true
		}
		if len(password) < 8 {
			f.PasswordError = "Password must be at least 8 characters long"
			f.HasError = true
		}

		if f.HasError {
			return render(c, pages.SignUp(false, internal.User{}, f))
		}

		// check if the user already exists
		user := internal.DB.QueryRow("SELECT id FROM users WHERE email = ?", f.Email)
		var id int
		err := user.Scan(&id)
		if err == nil {
			f.EmailError = "Email already exists. Please login or use another email."
			f.HasError = true
			return render(c, pages.SignUp(false, internal.User{}, f))
		}

		// save the user
		res, err := internal.DB.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", f.Name, f.Email, password)
		if err != nil {
			f.HasError = true
			f.Error = "Something went wrong. Please try again later."
			return render(c, pages.SignUp(false, internal.User{}, f))
		}
		rows, err := res.RowsAffected()
		if err != nil {
			f.Error = "Something went wrong. Please try again later."
			f.HasError = true
			return render(c, pages.SignUp(false, internal.User{}, f))
		}
		if rows != 1 {
			f.Error = "Something went wrong. Please try again later."
			f.HasError = true
			return render(c, pages.SignUp(false, internal.User{}, f))
		}

		f.Message = "User created successfully. Please login."
		f.Email = ""
		f.Name = ""
		return render(c, pages.SignUp(false, internal.User{}, f))
	})

	e.GET("/login", func(c echo.Context) error {
		f := internal.SignInForm{}
		return render(c, pages.Login(false, internal.User{}, f))
	})

	e.POST("/login", func(c echo.Context) error {
		f := internal.SignInForm{}
		f.Email = c.FormValue("email")
		password := c.FormValue("password")

		time.Sleep(10 * time.Second)

		// check if the user exists
		row := internal.DB.QueryRow("SELECT id, name, email, password FROM users WHERE email = ? and password = ?", f.Email, password)
		u := internal.User{}
		err := row.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
		if err != nil {
			f.Error = "Invalid email or password"
			f.HasError = true
			return render(c, pages.Login(false, u, f))
		}

		// set the jwt as a cookie
		token, err := internal.CreateJWT(u.Id, u.Name, u.Email)
		if err != nil {
			log.Println("Error creating JWT:", err)
			f.HasError = true
			f.Error = "Something went wrong. Please try again later."
			return render(c, pages.Login(false, u, f))
		}
		c.SetCookie(&http.Cookie{
			Name:     "jwt",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// redirect to todos via header
		c.Response().Header().Set("HX-Redirect", "/todos")
		return c.NoContent(http.StatusSeeOther)
	})

	e.POST("/logout", func(c echo.Context) error {
		c.SetCookie(&http.Cookie{
			Name:     "jwt",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// redirect to login via header
		c.Response().Header().Set("HX-Redirect", "/")
		return c.NoContent(http.StatusSeeOther)
	})

	// start the server
	if err := e.Start(":3000"); err != nil {
		log.Fatal("shutting down server with err:", err)
	}
}

func render(ctx echo.Context, component templ.Component) error {
	return component.Render(ctx.Request().Context(), ctx.Response())
}
