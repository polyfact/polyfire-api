package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	httprouter "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func checkAuthorizedDomains(project *database.Project, redirectURI string) bool {
	if len(project.AuthorizedDomains) == 0 {
		return true
	}

	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}
	origin := redirectURL.Hostname()

	if redirectURL.Port() != "" {
		origin = origin + ":" + redirectURL.Port()
	}

	fmt.Println(project.AuthorizedDomains)
	for _, domain := range project.AuthorizedDomains {
		if domain == origin {
			return true
		}
	}

	return false
}

func RedirectAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	projectID := ps.ByName("id")
	provider := r.URL.Query().Get("provider")
	redirectToFinal := r.URL.Query().Get("redirect_to")

	project, err := db.GetProjectByID(projectID)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(
				"The project \"%s\" was not found.\n\nIf you're not the app developer, please contact the developer.\n\nIf you are the developer, please check your project-id. If you're using React, it should be in the PolyfireProvider \"project\" props. You can find it on https://beta.polyfire.com/.\n",
				projectID,
			),
			http.StatusBadRequest,
		)
		return
	}

	if !checkAuthorizedDomains(project, redirectToFinal) {
		http.Error(w, "Unauthorized redirect URI", http.StatusUnauthorized)
		return
	}

	redirectToAPI := fmt.Sprintf(
		"%s/project/%s/auth/provider/callback?provider=%s&redirect_to=%s",
		os.Getenv("API_URL"),
		projectID,
		provider,
		url.QueryEscape(redirectToFinal),
	)

	url := os.Getenv(
		"SUPABASE_URL",
	) + "/auth/v1/authorize?provider=" + provider + "&redirect_to=" + url.QueryEscape(
		redirectToAPI,
	) + "&scopes=email"

	http.Redirect(w, r, url, http.StatusFound)
}

func wrapSupabaseRefreshToken(
	db database.Database,
	refreshToken string,
	projectID string,
) (string, error) {
	wrappedRefreshToken := make([]byte, 32)
	_, err := rand.Read(wrappedRefreshToken)
	if err != nil {
		return "", err
	}

	wrappedRefreshTokenString := strings.ReplaceAll(
		base64.StdEncoding.EncodeToString(wrappedRefreshToken),
		"/",
		"_",
	)
	wrappedRefreshTokenString = strings.ReplaceAll(wrappedRefreshTokenString, "+", "-")
	wrappedRefreshTokenString = strings.ReplaceAll(wrappedRefreshTokenString, "=", "")

	err = db.CreateRefreshToken(wrappedRefreshTokenString, refreshToken, projectID)
	if err != nil {
		return "", err
	}

	return wrappedRefreshTokenString, nil
}

func CallbackAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	projectID := ps.ByName("id")
	redirectTo := r.URL.Query().Get("redirect_to")
	accessToken := r.URL.Query().Get("access_token")
	refreshToken := r.URL.Query().Get("refresh_token")
	errorField := r.URL.Query().Get("error")

	if accessToken == "" && errorField == "" {
		// Supabase doesn't return the access_token in the query params but in the hash
		// to avoid leaking the token to the history. In our case we we'll do a redirect
		// with the token in the hash and therefore the token cannot leak.
		// They recommend to use the PKCE method when using server-side auth but I can't
		// manage to use it so we'll stick with Implicit flow and get the access_token to
		// the backend from a simple frontend page.
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		// TODO: Add a no-js fallback/a message in case the user get stuck on this page
		// 			 for unexpected reasons
		_, _ = w.Write(
			[]byte(
				"<script>l=document.location;a=l.hash.slice(1);history.replaceState(null,null,' ');l.search+=(l.search?\"&\":\"\")+(a?a:\"error=unknown_error\")</script>",
			),
		)
		return
	}

	if accessToken == "" {
		// TODO: Handle errors
		http.Error(w, errorField, http.StatusInternalServerError)
		return
	}

	project, err := db.GetProjectByID(projectID)
	if err != nil || project == nil {
		http.Error(w, "project_retrieval_error", http.StatusInternalServerError)
		return
	}

	if !checkAuthorizedDomains(project, redirectTo) {
		http.Error(w, "Unauthorized redirect URI", http.StatusUnauthorized)
		return
	}

	token, err := ExchangeToken(r.Context(), accessToken, *project, GetUserFromSupabaseToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrappedRefreshToken, err := wrapSupabaseRefreshToken(db, refreshToken, project.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(
		w,
		r,
		redirectTo+"#access_token="+token+"&refresh_token="+wrappedRefreshToken,
		http.StatusFound,
	)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenSupabaseResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	projectID := ps.ByName("id")

	var refreshTokenRequest RefreshTokenRequest

	err := json.NewDecoder(r.Body).Decode(&refreshTokenRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := db.GetProjectByID(projectID)
	if err != nil || project == nil {
		http.Error(w, "project_retrieval_error", http.StatusInternalServerError)
		return
	}

	refreshTokenFromDB, err := db.GetAndDeleteRefreshToken(refreshTokenRequest.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if (*refreshTokenFromDB).ProjectID != project.ID {
		http.Error(w, "refresh_token_project_mismatch", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	reqURL := os.Getenv("SUPABASE_URL") + "/auth/v1/token?grant_type=refresh_token"
	reqBody, err := json.Marshal(map[string]string{
		"refresh_token": refreshTokenFromDB.RefreshTokenSupabase,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(reqBody)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", os.Getenv("SUPABASE_KEY"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var refreshTokenResponse RefreshTokenSupabaseResponse
	err = json.NewDecoder(res.Body).Decode(&refreshTokenResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wrappedRefreshToken, err := wrapSupabaseRefreshToken(
		db,
		refreshTokenResponse.RefreshToken,
		project.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := ExchangeToken(
		r.Context(),
		refreshTokenResponse.AccessToken,
		*project,
		GetUserFromSupabaseToken,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"refresh_token": wrappedRefreshToken,
		"access_token":  token,
	})
}
