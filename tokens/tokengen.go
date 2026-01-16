package tokens

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var SECRET_KEY string = os.Getenv("SECRET_KEY")

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	User_ID    string
	jwt.RegisteredClaims
}

func TokenGenerator(email, firstName, lastName, userID string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_Name: firstName,
		Last_Name:  lastName,
		User_ID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24))),
		},
	}

	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(168))), // 7 days
		},
	}

	if SECRET_KEY == "" {
		SECRET_KEY = "secret"
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, err
}

// ValidateToken validates and parses a JWT token
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	if SECRET_KEY == "" {
		SECRET_KEY = "secret"
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is invalid"
		return
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now().Local()) {
		msg = "token is expired"
		return
	}

	return claims, msg
}

func UpdateAllTokens(signedToken, signedRefreshToken, userID string, userCollection *mongo.Collection, ctx context.Context) error {
	var updateObj bson.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	_, err := userCollection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.D{
			{Key: "$set", Value: updateObj},
		},
	)
	return err
}
