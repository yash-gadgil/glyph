package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetGrpcClient(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	return conn
}

func CreateTokenWithClaims(claims map[string]interface{}, exp time.Time) (string, error) {
	mc := jwt.MapClaims{}
	for k, v := range claims {
		mc[k] = v
	}
	mc["exp"] = exp.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	tokenStr, err := token.SignedString([]byte(os.Getenv("GLYPH_SECRET_KEY")))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func CreateToken(id string, exp time.Time) (string, error) {
	return CreateTokenWithClaims(map[string]interface{}{"userid": id}, exp)
}

func VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("GLYPH_SECRET_KEY")), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}

	userID, ok := claims["userid"].(string)
	if !ok {
		return "", fmt.Errorf("userid claim not found or invalid")
	}

	return userID, nil
}

func ParseTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("GLYPH_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}
	return claims, nil
}

func GetUserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("GLYPH_SECRET_KEY")), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse claims")
	}

	userID, ok := claims["userid"].(string)
	if !ok {
		return "", fmt.Errorf("userid claim not found or invalid")
	}

	return userID, nil
}

func SendEmail(toEmail, subject, body string) error {

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	from := "glyph.platform@gmail.com"
	password := os.Getenv("GLYPH_EMAIL_PASSWORD")

	token, err := CreateTokenWithClaims(map[string]interface{}{
		"email": toEmail,
	}, time.Now().Add(time.Minute*30))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	baseURL := os.Getenv("EMAIL_VERIFICATION_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost%s", os.Getenv("GATEWAY_SVC_PORT"))
	}

	verificationURL := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

	msg := []byte(
		"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			"<html><body>" +
			"<p>" + body + "</p>" +
			"<p><a href=\"" + verificationURL + "\">Click here to verify your email</a></p>" +
			"<p>Or copy and paste this link: " + verificationURL + "</p>" +
			"</body></html>\r\n")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
