package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	httprouter "github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
)

func RedirectAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	projectID := ps.ByName("id")
	provider := r.URL.Query().Get("provider")
	redirectToFinal := r.URL.Query().Get("redirect_to")

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
	)

	http.Redirect(w, r, url, http.StatusFound)
}

func CallbackAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	projectID := ps.ByName("id")
	redirectTo := r.URL.Query().Get("redirect_to")
	accessToken := r.URL.Query().Get("access_token")
	refreshToken := r.URL.Query().Get("refresh_token")
	error := r.URL.Query().Get("error")

	if accessToken == "" && error == "" {
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
		w.Write(
			[]byte(
				"<script>l=document.location;a=l.hash.slice(1);history.replaceState(null,null,' ');l.search+=(l.search?\"&\":\"\")+(a?a:\"error=unknown_error\")</script>",
			),
		)
		return
	}

	if accessToken == "" {
		// TODO: Handle errors
		http.Error(w, error, http.StatusInternalServerError)
		return
	}

	fmt.Println("Project ID:", projectID)
	fmt.Println("Access Token:", accessToken)
	fmt.Println("Refresh Token:", refreshToken)
	fmt.Println("Redirect To:", redirectTo)

	project, err := db.GetProjectByID(projectID)
	if err != nil || project == nil {
		http.Error(w, "project_retrieval_error", http.StatusInternalServerError)
		return
	}

	token, err := ExchangeToken(accessToken, *project, GetUserFromSupabaseToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Token:", token)

	projectRefreshToken := make([]byte, 32)
	_, err = rand.Read(projectRefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projectRefreshTokenString := strings.ReplaceAll(base64.StdEncoding.EncodeToString(projectRefreshToken), "/", "_")
	projectRefreshTokenString = strings.ReplaceAll(projectRefreshTokenString, "+", "-")
	projectRefreshTokenString = strings.ReplaceAll(projectRefreshTokenString, "=", "")

	fmt.Println("Project Refresh Token:", projectRefreshTokenString)

	err = db.CreateRefreshToken(projectRefreshTokenString, refreshToken, project.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectTo+"#access_token="+token+"&refresh_token="+projectRefreshTokenString, http.StatusFound)
}
