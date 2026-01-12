package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"

	"hytale-downloader/internal/buildinfo"
	"hytale-downloader/internal/ioutil"
	"hytale-downloader/internal/oauth"
	"hytale-downloader/internal/version"
)

var (
	printVersion    = flag.Bool("print-version", false, "Print available game version and exit")
	showBuildVersion = flag.Bool("version", false, "Print hytale-downloader version and exit")
	skipUpdateCheck = flag.Bool("skip-update-check", false, "Skip checking for hytale-downloader updates")
	checkUpdateOnly = flag.Bool("check-update", false, "Check for hytale-downloader updates and exit")
	downloadPath    = flag.String("download-path", "", "Path to download zip to")
	patchlineName   = flag.String("patchline", "release", "Patchline to download from")
	credentialsPath = flag.String("credentials-path", "", "Path to credentials file")
)

func main() {
	flag.Parse()

	// Handle -version flag
	if *showBuildVersion {
		fmt.Println(buildinfo.Version)
		return
	}

	// Handle -check-update flag
	if *checkUpdateOnly {
		checkForUpdates()
		return
	}

	// Check for updates unless skipped
	if !*skipUpdateCheck {
		checkForUpdates()
	}

	// Determine credentials path
	cPath := *credentialsPath
	if cPath == "" {
		cPath = ".hytale-downloader-credentials.json"
	}

	// Load or create OAuth session
	tok, err := getSavedSession(cPath)
	if err != nil {
		tok, err = createNewSession(cPath)
		if err != nil {
			log.Fatalf("error obtaining token: %v", err)
		}
	}

	// Save session callback
	saveSession := func(t *oauth.Token) {
		if err := oauth.SaveCredentials(cPath, t); err != nil {
			log.Printf("error saving session: %v", err)
		}
	}

	// Create token source that watches for refreshes
	ts := oauth.WatchTokenSource(cPath, tok, saveSession)
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: ts,
		},
	}

	// Handle -print-version flag
	if *printVersion {
		printVersionInfo(client)
		return
	}

	// Download the game
	download(client)
}

func getSavedSession(path string) (*oauth.Token, error) {
	return oauth.LoadCredentials(path)
}

func createNewSession(path string) (*oauth.Token, error) {
	ctx := context.Background()

	// Request device code
	da, err := oauth.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error requesting device code: %v", err)
	}

	// Display authorization instructions to user
	fmt.Printf("Please visit the following URL to authenticate:\n%s\n\n", da.VerificationURI)
	if da.VerificationURIComplete != "" {
		fmt.Printf("Or visit the following URL and enter the code:\n%s\n\n", da.VerificationURIComplete)
	}
	fmt.Printf("Authorization code: %s\n\n", da.UserCode)

	// Poll for token
	tok, err := oauth.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, err
	}

	// Save the token
	if err := oauth.SaveCredentials(path, tok); err != nil {
		return nil, fmt.Errorf("error saving session: %v", err)
	}

	return tok, nil
}

func checkForUpdates() {
	// Skip update check for dev builds
	if buildinfo.Version == "dev" {
		fmt.Println("skipping update check for dev build")
		return
	}

	info, err := version.CheckForUpdates()
	if err != nil {
		fmt.Printf("warning: failed to check for updates: %v\n", err)
		return
	}

	if info.Latest != buildinfo.Version {
		fmt.Printf("A new version of hytale-downloader is available: %s (current: %s)\n", info.Latest, buildinfo.Version)
		fmt.Printf("Download it from: %s\n", version.GetDownloaderURL())
	}
}

func printVersionInfo(client *http.Client) {
	manifest, err := version.Get(client, *patchlineName)
	if err != nil {
		log.Fatalf("error printing version: %v", err)
	}
	fmt.Println(manifest.Version)
}

func download(client *http.Client) {
	// Get manifest for version info
	manifest, err := version.Get(client, *patchlineName)
	if err != nil {
		log.Fatalf("error getting version: %v", err)
	}

	// Get signed download URL
	signedURL, err := version.GetSignedURL(client, *patchlineName)
	if err != nil {
		log.Fatalf("error getting download URL: %v", err)
	}

	// Determine download path
	path := *downloadPath
	if path == "" {
		path = fmt.Sprintf("hytale-%s-%s.zip", *patchlineName, manifest.Version)
	}

	// Ensure path ends with .zip
	if !strings.HasSuffix(strings.ToLower(path), ".zip") {
		path = path + ".zip"
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("error resolving path: %v", err)
	}

	fmt.Printf("downloading latest (%q patchline) to %q\n", *patchlineName, absPath)

	// Download the file
	if err := ioutil.Download(client, signedURL, absPath); err != nil {
		log.Fatalf("error downloading: %v", err)
	}

	// Validate checksum
	fmt.Println("validating checksum...")
	if err := ioutil.ValidateSHA256(absPath, manifest.SHA256); err != nil {
		// Remove the file if checksum fails
		os.Remove(absPath)
		log.Fatalf("%v", err)
	}

	fmt.Printf("successfully downloaded %q patchline (version %s)\n", *patchlineName, manifest.Version)
}
