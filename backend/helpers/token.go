package helpers

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/shjung-dev/ChatApplication/backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type`
	jwt.StandardClaims
}

var jwtKey []byte

func SetJWTKey(key string) {
	jwtKey = []byte(key)
}

func GetJWTKey() []byte {
	return []byte(jwtKey)
}

func GenerateToken(userID string, username string) (string, string) {
	tokenExpiry := time.Now().Add(24 * time.Hour).Unix()
		
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	//To create a new token we need to a new claim for respective token
	//For access token, userID is needed to identify who is making the request
	//We need to indicate the expiry date for each token
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpiry,
		},
	}

	refreshClaims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshTokenExpiry,
		},
	}

	//Generating tokens
	//Token is in the form -> <header> <payload> <signature>
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //now contains <header> <payload>
	signedAccessToken, err := accessToken.SignedString(jwtKey)       //now contains <header> <payload> <signature>
	if err != nil {
		panic(err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		panic(err)
	}

	return signedAccessToken, signedRefreshToken
}

func UpdateAllToken(signedToken string, signedRefreshToken string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userCollection := config.OpenCollection("user")

	updateObj := bson.D{
		{"$set", bson.D{
			{"token", signedToken},
			{"refresh_token", signedRefreshToken},
			{"updated_at", time.Now()},
		}},
	}

	filter := bson.M{"user_id": userID}

	_, err := userCollection.UpdateOne(ctx, filter, updateObj)

	return err
}

func HashPassword(password *string) *string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	hashedPwd := string(bytes)

	return &hashedPwd

}

func VerifyPassword(foundPwd string, pwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(foundPwd), []byte(pwd))
	return err == nil, err
}

var ErrTokenExpired = errors.New("token expired")

func ValidateToken(tokenString string) (*Claims, error) {
	secretKey := GetJWTKey()

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				//JWT validation error includes the ‘expired token’ flag
				//If the token isn't expired the bitwise operation will give 0
				return nil, ErrTokenExpired
			}
		}
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
