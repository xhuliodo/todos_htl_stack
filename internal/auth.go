package internal

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func CreateJWT(id int, name, email string) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"iat":   time.Now().Unix(),
		"sub":   strconv.Itoa(id),
		"name":  name,
		"email": email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("jwt")
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		claims, err := ParseJWT(cookie.Value)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		c.Set("user", claims)
		return next(c)
	}
}

func GetUser(c echo.Context) User {
	claims := c.Get("user").(jwt.MapClaims)
	idString := claims["sub"].(string)
	id, _ := strconv.Atoi(idString)

	return User{
		Id:    id,
		Name:  claims["name"].(string),
		Email: claims["email"].(string),
	}
}
