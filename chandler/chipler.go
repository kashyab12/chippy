package chandler

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kashyab12/chippy/internal/database"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func DecodeRequestBody[J *BodyJson | *UserJson | *WebhookJson](r *http.Request, bodyStructure J) (J, error) {
	decoder := json.NewDecoder(r.Body)
	defer CloseIoReadCloserStream(r.Body)
	err := decoder.Decode(bodyStructure)
	return bodyStructure, err
}

func EncodeErrorResponse(responseBody ErrorJson) ([]byte, error) {
	encodedBody, encodingErr := json.Marshal(responseBody)
	return encodedBody, encodingErr
}

func invalidChippyRequestStruct(w http.ResponseWriter) {
	log.Printf("Error decoding body JSON params!")
	if encodedErrJson, encodingErr := EncodeErrorResponse(ErrorJson{ErrMsg: "Something went wrong"}); encodingErr != nil {
		log.Println("Inception wtf!")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write(encodedErrJson)
		if writeErr != nil {
			log.Println("Stopping this right now lol")
		}
	}
}

func chippyTooLong(w http.ResponseWriter) {
	log.Printf("Chippy too damn long!")
	if encodedErrJson, encodingErr := json.Marshal(ErrorJson{ErrMsg: "Chirp is too long"}); encodingErr != nil {
		log.Println("Inception wtf!")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write(encodedErrJson)
		if writeErr != nil {
			log.Println("Stopping this right now lol")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (config *ApiConfig) PostChirp(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating Chippy!")
	if jsonBody, decodeErr := DecodeRequestBody(r, &BodyJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else if len(jsonBody.Body) > MaxChippyLen {
		chippyTooLong(w)
	} else if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		extractedJwtToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
		registeredClaims := jwt.RegisteredClaims{}
		if token, parseErr := jwt.ParseWithClaims(extractedJwtToken, &registeredClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSecret), nil
		}); parseErr != nil {
			log.Println("Invalid JWT, token is invalid or expired.")
			w.WriteHeader(http.StatusUnauthorized)
		} else if userId, subjectErr := token.Claims.GetSubject(); subjectErr != nil {
			log.Println("Unable to extract user id via the subject info within JWT")
			w.WriteHeader(http.StatusInternalServerError)
		} else if issuer, _ := token.Claims.GetIssuer(); issuer == RefreshTokenIssuer {
			log.Println("Can't use RefreshToken to post a chirp!")
			w.WriteHeader(http.StatusUnauthorized)
		} else if userIdInt, convErr := strconv.Atoi(userId); convErr != nil {
			log.Println(convErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chirp, createErr := chibeDb.CreateChirp(jsonBody.Body, userIdInt); createErr != nil {
			log.Printf("Error while creating the chirp: %v\n", createErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJson, encodeErr := json.Marshal(chirp); encodeErr != nil {
			log.Printf("Error while encoding the chirp to raw json %v: %v\n", rawJson, encodeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Println("Successfully encoded chirp and saved within CHIBE")
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJson)
			if err != nil {
				return
			}
		}
	}
}

func GetChirp(w http.ResponseWriter, r *http.Request) {
	if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
		log.Printf("Error while creating the database: %v\n", newDbErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if authorIDFilter := r.URL.Query().Get("author_id"); authorIDFilter != "" {
		if authorID, convErr := strconv.Atoi(authorIDFilter); convErr != nil {
			log.Println(convErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if filteredChirps, getFilteredChirpsErr := chibeDb.GetChirpsByAuthorID(authorID); getFilteredChirpsErr != nil {
			log.Printf("Error while obtaining chibe entries: %v\n", filteredChirps)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJsonList, encodingErr := json.Marshal(filteredChirps); encodingErr != nil {
			log.Printf("Error while encoding chibe entries: %v\n", encodingErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJsonList)
			if err != nil {
				return
			}
		}
	} else if chirps, getChirpsErr := chibeDb.GetChirps(); getChirpsErr != nil {
		log.Printf("Error while obtaining chibe entries: %v\n", getChirpsErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if rawJsonList, encodingErr := json.Marshal(chirps); encodingErr != nil {
		log.Printf("Error while encoding chibe entries: %v\n", encodingErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err := w.Write(rawJsonList)
		if err != nil {
			return
		}
	}
}

func GetSingleChirp(w http.ResponseWriter, r *http.Request) {
	if targetChirpIdStr := r.PathValue("chirpID"); len(targetChirpIdStr) < 1 {
		log.Printf("Unable to match to chirp id based on provided path\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if targetChirpId, convErr := strconv.Atoi(targetChirpIdStr); convErr != nil {
		log.Printf("Conversion error of chirp id from string to integer\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
		log.Printf("Error while creating the database: %v\n", newDbErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if chirps, getChirpsErr := chibeDb.GetChirps(); getChirpsErr != nil {
		log.Printf("Error while obtaining chibe entries: %v\n", getChirpsErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if targetIdx := slices.IndexFunc(chirps, func(ch database.Chirp) bool {
		return ch.Uid == targetChirpId
	}); targetIdx == -1 {
		log.Printf("Chirp ID not present within chibe the db\n")
		w.WriteHeader(http.StatusNotFound)
	} else if rawData, encodingErr := json.Marshal(chirps[targetIdx]); encodingErr != nil {
		log.Printf("Error while encoding target chibe\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err := w.Write(rawData)
		if err != nil {
			return
		}
	}
}

func (config *ApiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	if targetChirpIdStr := r.PathValue("chirpID"); len(targetChirpIdStr) < 1 {
		log.Printf("Unable to match to chirp id based on provided path\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if targetChirpId, convErr := strconv.Atoi(targetChirpIdStr); convErr != nil {
		log.Printf("Conversion error of chirp id from string to integer\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		extractedJwtToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
		registeredClaims := jwt.RegisteredClaims{}
		if token, parseErr := jwt.ParseWithClaims(extractedJwtToken, &registeredClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSecret), nil
		}); parseErr != nil {
			log.Println("Invalid JWT, token is invalid or expired.")
			w.WriteHeader(http.StatusUnauthorized)
		} else if expectedAuthorIdStr, subjectErr := token.Claims.GetSubject(); subjectErr != nil {
			log.Println("Unable to extract user id via the subject info within JWT")
			w.WriteHeader(http.StatusInternalServerError)
		} else if issuer, _ := token.Claims.GetIssuer(); issuer == RefreshTokenIssuer {
			log.Println("Can't use RefreshToken to post a chirp!")
			w.WriteHeader(http.StatusUnauthorized)
		} else if expectedAuthorId, idConvErr := strconv.Atoi(expectedAuthorIdStr); idConvErr != nil {
			log.Println(convErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chirps, getChirpsErr := chibeDb.GetChirps(); getChirpsErr != nil {
			log.Printf("Error while obtaining chibe entries: %v\n", getChirpsErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if targetIdx := slices.IndexFunc(chirps, func(ch database.Chirp) bool {
			return ch.Uid == targetChirpId
		}); targetIdx == -1 {
			log.Printf("Chirp ID not present within chibe the db\n")
			w.WriteHeader(http.StatusNotFound)
		} else if expectedAuthorId != chirps[targetIdx].AuthorID {
			log.Printf("The expected author id '%v' is not equal to the actual author id (via JWT) '%v'", expectedAuthorId, chirps[targetIdx].AuthorID)
			w.WriteHeader(http.StatusForbidden)
		} else if deletionErr := chibeDb.DeleteChirp(targetChirpId); deletionErr != nil {
			log.Println(deletionErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func PostUsers(w http.ResponseWriter, r *http.Request) {
	if jsonBody, decodeErr := DecodeRequestBody(r, &UserJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else {
		if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if user, createErr := chibeDb.CreateUser(jsonBody.Email, jsonBody.Password); createErr != nil {
			log.Printf("Error while creating the user: %v\n", createErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJson, encodeErr := json.Marshal(UserReturnJson{
			ID:          user.Uid,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		}); encodeErr != nil {
			log.Printf("Error while encoding the user to raw json %v: %v\n", rawJson, encodeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Println("Successfully encoded user and saved within CHIBE")
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJson)
			if err != nil {
				return
			}
		}
	}
}

func (config *ApiConfig) PutUser(w http.ResponseWriter, r *http.Request) {
	// Get Auth Headers
	if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		extractedJwtToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
		registeredClaims := jwt.RegisteredClaims{}
		if token, parseErr := jwt.ParseWithClaims(extractedJwtToken, &registeredClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSecret), nil
		}); parseErr != nil {
			log.Println("Invalid JWT, token is invalid or expired.")
			w.WriteHeader(http.StatusUnauthorized)
		} else if userId, subjectErr := token.Claims.GetSubject(); subjectErr != nil {
			log.Println("Unable to extract user id via the subject info within JWT")
			w.WriteHeader(http.StatusInternalServerError)
		} else if issuer, _ := token.Claims.GetIssuer(); issuer == RefreshTokenIssuer {
			log.Println("Can't use RefreshToken to update user info!")
			w.WriteHeader(http.StatusUnauthorized)
		} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if jsonBody, decodeErr := DecodeRequestBody(r, &UserJson{}); decodeErr != nil {
			invalidChippyRequestStruct(w)
		} else if userIdInt, convErr := strconv.Atoi(userId); convErr != nil {
			log.Println(convErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if updatedUser, updateErr := chibeDb.UpdateUser(userIdInt, jsonBody.Email, jsonBody.Password); updateErr != nil {
			log.Println(updateErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJson, encodeErr := json.Marshal(UserReturnJson{
			ID:          updatedUser.Uid,
			Email:       updatedUser.Email,
			IsChirpyRed: updatedUser.IsChirpyRed,
		}); encodeErr != nil {
			log.Printf("Error while encoding the user to raw json %v: %v\n", rawJson, encodeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJson)
			if err != nil {
				return
			}
		}
	}
}

func (config *ApiConfig) PostLogin(w http.ResponseWriter, r *http.Request) {
	if jsonBody, decodeErr := DecodeRequestBody(r, &UserJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else {
		if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if presentUser, userError := chibeDb.AuthUser(jsonBody.Email, jsonBody.Password); userError != nil {
			if userError.Error() == "password does not match" {
				log.Println("The password provided does not match the stored password within chibe!")
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				log.Println(userError)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			accessTokenIssuedAt := jwt.NewNumericDate(time.Now().UTC())
			accessTokenExpiresAt := jwt.NewNumericDate(accessTokenIssuedAt.Add(AccessTokenExpiry))
			refreshTokenExpiresAt := jwt.NewNumericDate(accessTokenIssuedAt.Add(RefreshTokenExpiry))
			if accessToken, accessTokenSigningErr := createJwt(AccessTokenIssuer, strconv.Itoa(presentUser.Uid), config.JwtSecret, accessTokenExpiresAt, accessTokenIssuedAt); accessTokenSigningErr != nil {
				log.Println(accessTokenSigningErr)
				w.WriteHeader(http.StatusInternalServerError)
			} else if refreshToken, refreshTokenSigningError := createJwt(RefreshTokenIssuer, strconv.Itoa(presentUser.Uid), config.JwtSecret, refreshTokenExpiresAt, accessTokenIssuedAt); refreshTokenSigningError != nil {
				log.Println(refreshTokenSigningError)
				w.WriteHeader(http.StatusInternalServerError)
			} else if addToSessionStoreErr := chibeDb.AddRefreshTokenToSessionStore(refreshToken); addToSessionStoreErr != nil {
				log.Println(addToSessionStoreErr)
				w.WriteHeader(http.StatusInternalServerError)
			} else if rawJson, encodeErr := json.Marshal(UserReturnJson{
				ID:           presentUser.Uid,
				Email:        presentUser.Email,
				IsChirpyRed:  presentUser.IsChirpyRed,
				Token:        accessToken,
				RefreshToken: refreshToken,
			}); encodeErr != nil {
				log.Printf("Error while encoding the user to raw json %v: %v\n", rawJson, encodeErr)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_, err := w.Write(rawJson)
				if err != nil {
					return
				}
			}
		}
	}
}

func (config *ApiConfig) PostRefresh(w http.ResponseWriter, r *http.Request) {
	// Get Auth Headers
	if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		extractedJwtToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
		registeredClaims := jwt.RegisteredClaims{}
		if token, parseErr := jwt.ParseWithClaims(extractedJwtToken, &registeredClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSecret), nil
		}); parseErr != nil {
			log.Println("Invalid JWT, token is invalid or expired.")
			w.WriteHeader(http.StatusUnauthorized)
		} else if issuer, _ := token.Claims.GetIssuer(); issuer != RefreshTokenIssuer {
			// Make sure it's a refresh token
			log.Println("RefreshToken required for the refresh endpoint!")
			w.WriteHeader(http.StatusUnauthorized)
		} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if userId, fetchSubjectErr := token.Claims.GetSubject(); fetchSubjectErr != nil {
			log.Println(fetchSubjectErr)
			w.WriteHeader(http.StatusUnauthorized)
		} else if isRevoked, checkStoreErr := chibeDb.IsRefreshTokenRevoked(extractedJwtToken); checkStoreErr != nil {
			log.Println(checkStoreErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if isRevoked {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			accessTokenIssuedAt := jwt.NewNumericDate(time.Now().UTC())
			accessTokenExpiresAt := jwt.NewNumericDate(accessTokenIssuedAt.Add(AccessTokenExpiry))
			if accessToken, accessTokenSigningErr := createJwt(AccessTokenIssuer, userId, config.JwtSecret, accessTokenExpiresAt, accessTokenIssuedAt); accessTokenSigningErr != nil {
				log.Println(accessTokenSigningErr)
				w.WriteHeader(http.StatusInternalServerError)
			} else if rawJson, encodeErr := json.Marshal(TokenRefreshReturnJson{
				Token: accessToken,
			}); encodeErr != nil {
				log.Println(encodeErr)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_, err := w.Write(rawJson)
				if err != nil {
					return
				}
			}
		}
	}
}

func (config *ApiConfig) PostRevoke(w http.ResponseWriter, r *http.Request) {
	// Get Auth Headers
	if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		extractedJwtToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
		registeredClaims := jwt.RegisteredClaims{}
		if token, parseErr := jwt.ParseWithClaims(extractedJwtToken, &registeredClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.JwtSecret), nil
		}); parseErr != nil {
			log.Println("Invalid JWT, token is invalid or expired.")
			w.WriteHeader(http.StatusUnauthorized)
		} else if issuer, _ := token.Claims.GetIssuer(); issuer != RefreshTokenIssuer {
			// Make sure it's a refresh token
			log.Println("RefreshToken required for the revoke endpoint!")
			w.WriteHeader(http.StatusUnauthorized)
		} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if revokeErr := chibeDb.RevokeToken(extractedJwtToken); revokeErr != nil {
			log.Println(revokeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (config *ApiConfig) ChirpyRedWebhook(w http.ResponseWriter, r *http.Request) {
	const ChirpyRedReq = "user.upgraded"
	if jsonBody, decodeErr := DecodeRequestBody(r, &WebhookJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else if jsonBody.Event != ChirpyRedReq {
		w.WriteHeader(http.StatusOK)
	} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
		log.Printf("Error while creating the database: %v\n", newDbErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if userId, isPresent := jsonBody.Data["user_id"]; !isPresent {
		log.Println("User ID not provided in webhook")
		w.WriteHeader(http.StatusBadRequest)
	} else if upgradeErr := chibeDb.UpgradeUser(userId); upgradeErr != nil {
		if upgradeErr.Error() == "user not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Header.Get("Authorization") == "" {
		log.Println("Authorization header not provided")
		w.WriteHeader(http.StatusUnauthorized)
	} else if extractedApiKey := strings.Split(r.Header.Get("Authorization"), "ApiKey ")[1]; extractedApiKey != config.PolkaKey {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func createJwt(issuer, subject, secretKey string, expiresAt, issuedAt *jwt.NumericDate) (jwtToken string, signingError error) {
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   subject,
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims)
	if jwtToken, signingError = token.SignedString([]byte(secretKey)); signingError != nil {
		return jwtToken, signingError
	}
	return jwtToken, nil
}
