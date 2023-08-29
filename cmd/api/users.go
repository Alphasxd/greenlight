package main

import (
	"errors"
	"net/http"

	"github.com/Alphasxd/greenlight/internal/data"
	"github.com/Alphasxd/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// 将 JSON 解码到 input 结构体中
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 创建一个新的 User 记录，将 input 中的数据赋值给对应的字段
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false, // 默认情况下，新用户的激活状态为 false，显式指定有助于代码的可读性
	}
	// 使用 Set() 方法设置密码
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 将用户信息插入到数据库中
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 调用 background() 方法，将发送欢迎邮件的任务放入任务队列中
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// 将用户信息以 JSON 格式写入响应体中，并将状态码设为 201 Created
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
