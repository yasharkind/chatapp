package middlewares

import (
	"chatapp/objects"
	"chatapp/services"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Validate(cookie string, secret string, userservice service.UserService) *objects.User {
		if cookie == "" {
			return nil
		}
		token, err := jwt.Parse(cookie, func(token *jwt.Token) (any, error) {
			if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
			}
	
			return secret, nil;
		})
		if token == nil {
			return nil
		}
		claims, _ := token.Claims.(jwt.MapClaims)

		if claims["sub"] == nil || claims["exp"] == nil {
			fmt.Println("token invalid", token)
			return nil
		}
	
		user, err := userservice.FindById(int(claims["sub"].(float64)))
		t2 := time.Unix(int64(claims["exp"].(float64)), 0)
		if err != nil {
			log.Printf("validate error %s", err)
		}
		expired := t2.Unix() <= time.Now().Unix() 
		if user != nil && !expired {
			return user
		}
		println(expired)
		return nil
	}
