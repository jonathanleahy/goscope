package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Determine web directory
	webDir := "web/public"
	if len(os.Args) > 1 {
		webDir = os.Args[1]
	}

	// Make path absolute
	absPath, err := filepath.Abs(webDir)
	if err != nil {
		log.Fatal(err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absPath)
	}

	// Create file server
	fs := http.FileServer(http.Dir(absPath))
	http.Handle("/", fs)

	port := "8080"
	fmt.Printf("\n🔍 Go Scope Visualizer Server\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📁 Serving: %s\n", absPath)
	fmt.Printf("🌐 URL: http://localhost:%s\n", port)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Printf("Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
