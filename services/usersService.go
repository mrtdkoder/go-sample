package services

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	UID         int       `json:"UID"`
	EMail       string    `json:"EMail"`
	Password    string    `json:"Password"`
	FullName    string    `json:"FullName"`
	HomeDir     string    `json:"HomeDir"`
	SCode       int       `json:"SCode"`
	AuthToken   string    `json:"AuthToken"`
	IsActive    bool      `json:"IsActive"`
	LastLoginAt time.Time `json:"lastLogin"`
	CreatedAt   time.Time `json:"createdAt"`
	TokenInfo   AccessToken
}

func (u *User) GetTokenInfo() AccessToken {
	cc, _ := ParseJWT(u.AuthToken)
	return AccessToken{
		Token:     u.AuthToken,
		ExpiresAt: cc.ExpiresAt.Time,
		IssuedAt:  cc.IssuedAt.Time,
	}
}

type LoginCredentials struct {
	EMail    string `json:"email"`
	Password string `json:"password"`
}

type AccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	IssuedAt  time.Time `json:"issuedAt"`
}

// Define your secret key (keep this secure!)
var secretKey = []byte("abuzerkadayif")
var _issuer = "mrtCloud"
var SQLiteDB *sql.DB

// Custom claims structure
type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func OpenDB() *sql.DB {
	if SQLiteDB != nil {
		// Check if the connection is still alive
		if err := SQLiteDB.Ping(); err == nil {
			return SQLiteDB
		}
	}
	if runtime.GOOS == "windows" {
		SQLiteDB, _ = sql.Open("sqlite3", "data\\data.db")
	} else {
		SQLiteDB, _ = sql.Open("sqlite3", "data/data.db")
	}
	return SQLiteDB
}

func GenerateJWT(userID, email string) (AccessToken, error) {
	// Create claims
	tnow := time.Now()
	expAt := tnow.Add(time.Duration(24) * time.Hour)
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expAt), // 24 hours
			IssuedAt:  jwt.NewNumericDate(tnow),
			NotBefore: jwt.NewNumericDate(tnow),
			Issuer:    _issuer,
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return AccessToken{
			Token: "",
		}, err
	}

	return AccessToken{
		Token:     tokenString,
		IssuedAt:  tnow,
		ExpiresAt: expAt,
	}, nil
}

func ParseJWT(tokenString string) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashPassword hashes a password using SHA-1
func HashPassword(password string) string {
	h := sha256.New() //sha1.New()
	// Write the password to the hash
	h.Write([]byte(password))
	// Return the hexadecimal representation of the hash
	return fmt.Sprintf("%x", h.Sum(nil))
}

/************************** User Service *****************************/

// AddUser adds a new user to the database
func AddUser(user *User) (int64, error) {
	user.SCode = time.Now().Nanosecond()
	user.IsActive = true
	user.CreatedAt = time.Now()
	result, err := OpenDB().Exec(`INSERT INTO users(EMail, Password, FullName, HomeDir, SCode, AuthToken, IsActive, CreatedAt) VALUES(?, ?, ?, ?, ?, ?, ?,?)`,
		user.EMail, HashPassword(user.Password), user.FullName, user.HomeDir, user.SCode, user.AuthToken, user.IsActive, user.CreatedAt)
	if err != nil {
		log.Printf(" > Failed to add user to database: %s-%s", err, user.EMail)
		return 0, err
	}
	lastid, e := result.LastInsertId()
	if e != nil {
		log.Printf(" > Failed to retrieve last insert ID: %s", e)
		return 0, err
	}
	return lastid, err
}

// EditUser edits an existing user in the database
func EditUserById(user *User, uid int) (int64, error) {
	result, err := OpenDB().Exec(`UPDATE users SET EMail=?, FullName=?, HomeDir=?, SCode=?, AuthToken=?, IsActive=?, LastLoginAt=? WHERE UID=?`,
		user.EMail, user.FullName, user.HomeDir, user.SCode, user.AuthToken, user.IsActive, user.LastLoginAt, uid)
	if err != nil {
		log.Printf("Failed to edit user to database: %s", err)
		return 0, err
	}
	rowsAffected, e := result.RowsAffected()
	if e != nil {
		log.Printf("Failed to retrieve affected rows count: %s", e)
		return 0, err
	}
	return rowsAffected, err
}

func EditUserByEMail(user *User, email string) (int64, error) {
	result, err := OpenDB().Exec(`UPDATE users SET EMail=?, FullName=?, HomeDir=?, SCode=?, AuthToken=?, IsActive=?, LastLoginAt=? WHERE EMail=?`,
		user.EMail, user.FullName, user.HomeDir, user.SCode, user.AuthToken, user.IsActive, user.LastLoginAt, email)
	if err != nil {
		log.Fatal("Failed to edit user to database: ", err)
		return 0, err
	}
	rowsAffected, e := result.RowsAffected()
	if e != nil {
		log.Fatal("Failed to retrieve affected rows count: ", e)
		return 0, err
	}
	return rowsAffected, err
}

// change user password
func ChangePassword(uid int, newPassword string) (int64, error) {
	hashedPwd := HashPassword(newPassword)
	result, err := OpenDB().Exec(`UPDATE users SET Password=? WHERE UID=?`, hashedPwd, uid)
	if err != nil {
		log.Fatal("Failed to change user password: ", err)
		return 0, err
	}
	rowsAffected, e := result.RowsAffected()
	if e != nil {
		log.Fatal("Failed to retrieve affected rows count: ", e)
		return 0, err
	}
	return rowsAffected, e
}

// DeleteUser deletes a user from the database
func DeleteUserById(uid int) (int64, error) {
	result, err := OpenDB().Exec(`delete from users WHERE UID=?`, uid)
	if err != nil {
		log.Fatal("Failed to delete user from database: ", err)
		return 0, err
	}
	rowsAffected, e := result.RowsAffected()
	if e != nil {
		log.Fatal("Failed to retrieve affected rows count: ", e)
		return 0, err
	}
	return rowsAffected, err
}

// DeleteUser deletes a user from the database
func DeleteUserByEMail(email string) (int64, error) {
	result, err := OpenDB().Exec(`delete from users WHERE EMail=?`, email)
	if err != nil {
		log.Fatal("Failed to delete user from database: ", err)
		return 0, err
	}
	rowsAffected, e := result.RowsAffected()
	if e != nil {
		log.Fatal("Failed to retrieve affected rows count: ", e)
		return 0, err
	}
	return rowsAffected, err
}

// GetUser retrieves a user from the database based on credentials
func GetUserById(uid int) (*User, error) {
	result := &User{}
	row := OpenDB().QueryRow(`select UID, EMail, Password, FullName, HomeDir, SCode, AuthToken, IsActive, LastLoginAt, CreatedAt from users where UID = ?`, uid)

	err := row.Scan(&result.UID, &result.EMail, &result.Password, &result.FullName, &result.HomeDir, &result.SCode, &result.AuthToken, &result.IsActive, &result.LastLoginAt, &result.CreatedAt)
	return result, err
}

// GetUser retrieves a user from the database based on credentials
func GetUserByEMail(email string) (*User, error) {
	result := &User{}
	row := OpenDB().QueryRow(`select UID, EMail, Password, FullName, HomeDir, SCode, AuthToken, IsActive from users where EMail = ?`, email)

	err := row.Scan(&result.UID, &result.EMail, &result.Password, &result.FullName, &result.HomeDir, &result.SCode, &result.AuthToken, &result.IsActive)
	return result, err
}

// GetAllUsers retrieves all users from the database
func GetUsers(queryTxt string) ([]User, error) {
	sqlstr := `select UID, EMail, FullName, HomeDir, SCode, AuthToken, IsActive from users `
	if queryTxt != "" {
		sqlstr += fmt.Sprintf("where EMail like '%%%s%%' or FullName like '%%%s%%' ", queryTxt, queryTxt)
	}
	fmt.Println(" > SQL Query:", sqlstr)
	rows, err := OpenDB().Query(sqlstr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var _user User
		rows.Scan(&_user.UID, &_user.EMail, &_user.FullName, &_user.HomeDir, &_user.SCode, &_user.AuthToken, &_user.IsActive)
		users = append(users, _user)
	}
	return users, nil
}

// AuthenticateUser authenticates a user based on credentials
func AuthenticateUser(email, password string) (*User, error) {
	user, err := GetUserByEMail(email)
	if err != nil {
		return nil, err
	}

	if user != nil && user.Password == HashPassword(password) {
		user.TokenInfo, _ = GenerateAuthToken(user) // GenerateJWT(fmt.Sprint(user.UID), user.EMail) // GenerateAuthToken(user.UName)
		user.AuthToken = user.TokenInfo.Token
		user.LastLoginAt = time.Now()
		EditUserByEMail(user, email) // Update last login time or other info if needed
		return user, nil
	}
	return nil, fmt.Errorf("authentication failed")
}

func AuthenticateUserWithToken(token string) (*User, error) {
	claims, err := ParseJWT(token)
	if err != nil {
		return nil, err
	}
	user, err := GetUserByEMail(claims.Email)
	if err != nil {
		return nil, err
	}
	if user != nil && strconv.Itoa(user.UID) == claims.UserID {
		user.AuthToken = token
		return user, nil
	}
	return nil, fmt.Errorf("authentication failed")
}

// GenerateAuthToken generates an authentication token for a user
func GenerateAuthToken(user *User) (AccessToken, error) {
	accToken, err := GenerateJWT(fmt.Sprint(user.UID), user.EMail)
	if err != nil {
		return accToken, err
	}
	return accToken, nil
}
