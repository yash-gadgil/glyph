package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func GetGrpcClient(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return conn
}

func CreateToken(id string, exp time.Time, keyPair *KeyPair) (string, error) {
	return CreateTokenWithClaims(map[string]interface{}{"userid": id}, exp, keyPair)
}

func CreateTokenWithClaims(claims map[string]interface{}, exp time.Time, keyPair *KeyPair) (string, error) {
	mc := jwt.MapClaims{}
	for k, v := range claims {
		mc[k] = v
	}
	mc["exp"] = exp.Unix()
	mc["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mc)
	token.Header["kid"] = keyPair.ID

	tokenStr, err := token.SignedString(keyPair.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}

func parseWithKeystore(tokenString string, keyStore *KeyStore) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v (expected RS256)", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid (key ID) not found in token header")
		}

		publicKey, err := keyStore.GetPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
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

func VerifyToken(tokenString string, keyStore *KeyStore) (string, error) {
	claims, err := parseWithKeystore(tokenString, keyStore)
	if err != nil {
		return "", err
	}
	userID, ok := claims["userid"].(string)
	if !ok {
		return "", fmt.Errorf("userid claim not found or invalid")
	}
	return userID, nil
}

func ParseTokenClaims(tokenString string, keyStore *KeyStore) (jwt.MapClaims, error) {
	return parseWithKeystore(tokenString, keyStore)
}

func GetUserIDFromToken(tokenString string, keyStore *KeyStore) (string, error) {
	return VerifyToken(tokenString, keyStore)
}

func SendEmail(toEmail, subject, body string) error {
	if err := ValidateEmail(toEmail); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	from := "glyph.platform@gmail.com"
	password := os.Getenv("GLYPH_EMAIL_PASSWORD")

	msg := []byte(
		"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			body + "\r\n")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

const (
	minPasswordLen = 8
	maxPasswordLen = 16
	minNameLen     = 1
	maxNameLen     = 20
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if email == "" {
		return status.Errorf(codes.InvalidArgument, "email is required")
	}

	email = strings.ToLower(strings.TrimSpace(email))

	if len(email) > 254 {
		return status.Errorf(codes.InvalidArgument, "email is too long (max 254 characters)")
	}

	if !emailRegex.MatchString(email) {
		return status.Errorf(codes.InvalidArgument, "email format is invalid")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return status.Errorf(codes.InvalidArgument, "email format is invalid")
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) == 0 || len(localPart) > 64 {
		return status.Errorf(codes.InvalidArgument, "email local part must be 1-64 characters")
	}

	if len(domainPart) == 0 || len(domainPart) > 253 {
		return status.Errorf(codes.InvalidArgument, "email domain is invalid")
	}

	if !strings.Contains(domainPart, ".") {
		return status.Errorf(codes.InvalidArgument, "email domain must contain a dot")
	}

	return nil
}

func ValidatePassword(pw string) error {
	charCount := utf8.RuneCountInString(pw)
	if charCount < minPasswordLen {
		return status.Errorf(codes.InvalidArgument, "password must be at least %d characters", minPasswordLen)
	}
	if charCount > maxPasswordLen {
		return status.Errorf(codes.InvalidArgument, "password must be no more than %d characters", maxPasswordLen)
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, ch := range pw {
		switch {
		case ch == ' ':
			return status.Errorf(codes.InvalidArgument, "password must not contain spaces")
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch == '_':
			return status.Errorf(codes.InvalidArgument, "password contains invalid character '_'")
		case (ch >= '!' && ch <= '/') || (ch >= ':' && ch <= '@') || (ch >= '[' && ch <= '`') || (ch >= '{' && ch <= '~'):
			hasSpecial = true
		default:
			return status.Errorf(codes.InvalidArgument, "password contains invalid character")
		}
	}

	if !hasLower {
		return status.Errorf(codes.InvalidArgument, "password must include at least one lowercase letter")
	}
	if !hasUpper {
		return status.Errorf(codes.InvalidArgument, "password must include at least one uppercase letter")
	}
	if !hasDigit {
		return status.Errorf(codes.InvalidArgument, "password must include at least one digit")
	}
	if !hasSpecial {
		return status.Errorf(codes.InvalidArgument, "password must include at least one special character")
	}

	return nil
}

func ValidateName(name string) error {
	charCount := utf8.RuneCountInString(name)
	if charCount < minNameLen {
		return status.Errorf(codes.InvalidArgument, "name must be at least %d characters", minNameLen)
	}
	if charCount > maxNameLen {
		return status.Errorf(codes.InvalidArgument, "name must be no more than %d characters", maxNameLen)
	}

	return nil
}
