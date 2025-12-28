package youtube

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type tokenFile struct {
	Token *oauth2.Token `json:"token"`
}

// LoadOAuthConfig reads a Google OAuth client JSON (downloaded from Cloud Console)
// and returns an oauth2.Config for YouTube readonly access.
func LoadOAuthConfig(clientSecretPath string) (*oauth2.Config, error) {
	f, err := os.Open(clientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("open client secret json: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read client secret json: %w", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/youtube.readonly")
	if err != nil {
		return nil, fmt.Errorf("parse oauth client json: %w", err)
	}

	return config, nil
}

// GetClient returns an authenticated HTTP client. It caches the oauth token in tokenPath.
//
// If tokenPath is empty, it defaults to <clientSecretPath>.token.json.
func GetClient(ctx context.Context, clientSecretPath, tokenPath string) (*http.Client, error) {
	config, err := LoadOAuthConfig(clientSecretPath)
	if err != nil {
		return nil, err
	}

	if tokenPath == "" {
		tokenPath = clientSecretPath + ".token.json"
	}

	if tok, err := readToken(tokenPath); err == nil {
		return config.Client(ctx, tok), nil
	}

	tok, err := authorizeViaLocalhost(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := writeToken(tokenPath, tok); err != nil {
		return nil, err
	}

	return config.Client(ctx, tok), nil
}

func readToken(tokenPath string) (*oauth2.Token, error) {
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tf tokenFile
	if err := json.NewDecoder(f).Decode(&tf); err != nil {
		return nil, err
	}
	if tf.Token == nil {
		return nil, fmt.Errorf("token file missing token")
	}

	return tf.Token, nil
}

func writeToken(tokenPath string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0o755); err != nil {
		// If tokenPath is in repo root, Dir() is "." and this is a no-op.
		if filepath.Dir(tokenPath) != "." {
			return fmt.Errorf("create token dir: %w", err)
		}
	}

	f, err := os.Create(tokenPath)
	if err != nil {
		return fmt.Errorf("create token file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(tokenFile{Token: tok}); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}

	return nil
}

func authorizeViaLocalhost(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen localhost: %w", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	cfg := *config
	cfg.RedirectURL = redirectURL

	state, err := randomState()
	if err != nil {
		return nil, err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("oauth state mismatch")
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("missing oauth code")
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		io.WriteString(w, "Authorization complete. You can close this tab.")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	defer func() {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxShutdown)
	}()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("open browser: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case code := <-codeCh:
		tok, err := cfg.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange oauth code: %w", err)
		}
		return tok, nil
	}
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
