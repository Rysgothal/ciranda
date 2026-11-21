package http

import (
	"ciranda/internal/password"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,max=100"`
	Login    string `json:"login" validate:"required,max=50"`
	Email    string `json:"email" validate:"omitempty,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type DeleteUserRequest struct {
	Login    string `json:"login" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type UserResponse struct {
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
	ID    int64  `json:"id"`
}
type UpdateUserRequest struct {
	Name     *string `json:"name" validate:"omitempty,required,max=100"`
	Login    *string `json:"login" validate:"omitempty,required,max=50"`
	Email    *string `json:"email" validate:"omitempty,email,max=254"`
	Password *string `json:"password" validate:"omitempty,required,min=8,max=128"`
}

func (s *Server) CreateUser(c *gin.Context) {
	var body CreateUserRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Login = strings.TrimSpace(body.Login)
	body.Email = strings.TrimSpace(body.Email)

	if isPropJSONEmpty(c, "name", body.Name) ||
		isPropJSONEmpty(c, "login", body.Login) ||
		isPropJSONEmpty(c, "password", body.Password) {
		return
	}

	if err := validator.New().Struct(body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	hashedPassword, err := password.Hash(body.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to create user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return
	}

	var result sql.Result
	result, err = s.db.ExecContext(c.Request.Context(), `
	    INSERT INTO users (
			name, email, login, password
		) VALUES (
		    ?, ?, ?, ?
		)`,
		body.Name, body.Email, body.Login, hashedPassword)

	if err != nil {
		writeDBValidationError(c, err)
		return
	}

	var lastInsertID int64
	lastInsertID, err = result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to create user",
			"errors":  gin.H{"internal": "failed to get last insert id"},
		})
		return
	}

	responseUser := &UserResponse{Name: body.Name, Login: body.Login, Email: body.Email, ID: lastInsertID}
	c.Header("Location", fmt.Sprintf("/users/%d", lastInsertID))
	c.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "user created successfully",
		"user":    responseUser,
	})
}

func (s *Server) DeleteUser(c *gin.Context) {
	isValidID, userID := isValidParamID(c)
	if !isValidID {
		return
	}

	var body DeleteUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	body.Login = strings.TrimSpace(body.Login)
	if isPropJSONEmpty(c, "login", body.Login) || isPropJSONEmpty(c, "password", body.Password) {
		return
	}

	if err := validator.New().Struct(body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	var userLogin, userPassword string
	err := s.db.QueryRowContext(c.Request.Context(), "SELECT LOGIN, PASSWORD FROM USERS WHERE ID = ?", userID).Scan(&userLogin, &userPassword)

	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "user not found",
			"errors":  gin.H{"not_found": "user not found"},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to delete user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return
	}

	if userLogin != body.Login {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "wrong login",
			"errors":  gin.H{"login": "wrong login"},
		})
		return
	}

	if !password.Verify(body.Password, userPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "wrong password",
			"errors":  gin.H{"password": "wrong password"},
		})
		return
	}

	var result sql.Result
	result, err = s.db.ExecContext(c.Request.Context(), "DELETE FROM USERS WHERE ID = ?", userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to delete user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return
	}

	if writeDBUserNotFound(c, result, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "user deleted successfully",
	})
}

func (s *Server) UpdateUser(c *gin.Context) {
	isValidID, userID := isValidParamID(c)
	if !isValidID {
		return
	}

	var body UpdateUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	if body.Name == nil && body.Login == nil && body.Email == nil && body.Password == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "validation failed",
			"errors":  gin.H{"body": "at least one field is required"},
		})
		return
	}

	if body.Name != nil {
		value := strings.TrimSpace(*body.Name)
		body.Name = &value
	}
	if body.Login != nil {
		value := strings.TrimSpace(*body.Login)
		body.Login = &value
	}
	if body.Email != nil {
		value := strings.TrimSpace(*body.Email)
		body.Email = &value
	}

	if (body.Name != nil && *body.Name == "") ||
		(body.Login != nil && *body.Login == "") ||
		(body.Password != nil && strings.TrimSpace(*body.Password) == "") {
		field := "name"
		if body.Login != nil && *body.Login == "" {
			field = "login"
		} else if body.Password != nil && strings.TrimSpace(*body.Password) == "" {
			field = "password"
		}
		isPropJSONEmpty(c, field, "")
		return
	}

	if err := validator.New().Struct(body); err != nil {
		writeRequestValidationError(c, err)
		return
	}

	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 5)
	if body.Name != nil {
		setClauses = append(setClauses, "NAME = ?")
		args = append(args, *body.Name)
	}
	if body.Login != nil {
		setClauses = append(setClauses, "LOGIN = ?")
		args = append(args, *body.Login)
	}
	if body.Email != nil {
		setClauses = append(setClauses, "EMAIL = ?")
		args = append(args, *body.Email)
	}
	if body.Password != nil {
		hashedPassword, err := password.Hash(*body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "failed to update user",
				"errors":  gin.H{"internal": "internal error"},
			})
			return
		}
		setClauses = append(setClauses, "PASSWORD = ?")
		args = append(args, hashedPassword)
	}

	args = append(args, userID)
	result, err := s.db.ExecContext(c.Request.Context(), "UPDATE USERS SET "+strings.Join(setClauses, ", ")+" WHERE ID = ?", args...)
	if err != nil {
		if sqliteError, ok := errors.AsType[*sqlite.Error](err); ok && sqliteError.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			c.JSON(http.StatusConflict, gin.H{
				"code":    http.StatusConflict,
				"message": "failed to update user",
				"errors":  gin.H{"login": "user with this login already exists"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to update user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return
	}

	if writeDBUserNotFound(c, result, err) {
		return
	}

	c.Status(http.StatusNoContent)
}

func writeRequestValidationError(c *gin.Context, err error) {
	validationErrs, ok := errors.AsType[validator.ValidationErrors](err)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "invalid json",
			"errors":  gin.H{"json": "invalid"},
		})
		return
	}

	errs := make(map[string]string)

	for _, fieldErr := range validationErrs {
		switch fieldErr.Field() {
		case "Password":

			switch fieldErr.Tag() {
			case "min":
				errs["password"] = "password must be at least 8 characters"
			case "max":
				errs["password"] = "password must be at most 128 characters"
			default:
				errs["password"] = "password is required"
			}

		case "Email":
			errs["email"] = "invalid email"

			if fieldErr.Tag() == "max" {
				errs["email"] = "email must be at most 254 characters"
			}
		case "Name":
			errs["name"] = "name is required"

			if fieldErr.Tag() == "max" {
				errs["name"] = "name must be at most 100 characters"
			}
		case "Login":
			errs["login"] = "login is required"

			if fieldErr.Tag() == "max" {
				errs["login"] = "login must be at most 50 characters"
			}
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"code":    http.StatusBadRequest,
		"message": "validation failed",
		"errors":  errs,
	})
}

func writeDBValidationError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	sqliteError, ok := errors.AsType[*sqlite.Error](err)

	if !ok {
		c.JSON(status, gin.H{
			"code":    status,
			"message": "failed to create user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return
	}

	var errorMessage gin.H

	switch sqliteError.Code() {
	case sqlite3.SQLITE_CONSTRAINT_UNIQUE:
		status = http.StatusConflict
		errorMessage = gin.H{"login": "user with this login already exists"}

	case sqlite3.SQLITE_CONSTRAINT_CHECK, sqlite3.SQLITE_CONSTRAINT_NOTNULL:
		status = http.StatusBadRequest
		errorMessage = gin.H{"database": "required field"}

	default:
		errorMessage = gin.H{"database": "unknown error"}
	}

	c.JSON(status, gin.H{
		"code":    status,
		"message": "failed to create user",
		"errors":  errorMessage,
	})
}

func writeDBUserNotFound(c *gin.Context, qry sql.Result, err error) bool {
	var affected int64
	affected, err = qry.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "failed to delete user",
			"errors":  gin.H{"internal": "internal error"},
		})
		return true
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "user not found",
			"errors":  gin.H{"not_found": "user not found"},
		})
		return true
	}

	return false
}

func isPropJSONEmpty(c *gin.Context, propName string, propValue string) bool {
	if strings.TrimSpace(propValue) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "validation failed",
			"errors":  gin.H{propName: fmt.Sprintf("%s cannot be empty", propName)},
		})

		return true
	}

	return false
}

func isValidParamID(c *gin.Context) (bool, int) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "invalid user id",
			"errors":  gin.H{"invalid": "invalid user id"},
		})

		return false, 0
	}

	return true, userID
}
