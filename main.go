package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	geminiAPIURLBase   = "https://generativelanguage.googleapis.com/v1beta/models"
	geminiAPIGenerate  = ":generateContent"
	ErrInvalidResponse = "geminiclient: invalid response format from Gemini API: no candidates or parts found"
	ErrEmptyInput      = "geminiclient: input text and fileURL cannot both be empty"
	ErrMissingAPIKey   = "geminiclient: API key is required"
	ErrMissingPrompt   = "geminiclient: system prompt is required"
)

type AISettings struct {
	Temperature float64
	TopK        int
	TopP        float64
	MaxTokens   int
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type TemplateData struct {
	CurrentPath     string
	Command         string
	CommandExecuted bool
	CommandName     string
	IsExecutable    bool
	FileType        string
	IsInteractive   bool
	Output          string
	Error           string
	ExitCode        int
	DirError        string
	Entries         []DirEntry
}

type DirEntry struct {
	Name         string
	IsDir        bool
	IsExecutable bool
	IsEditable   bool
	FileType     string
	FullPath     string
	ParentPath   string
}

var codeFileExtensions = map[string]bool{
	".go": true, ".sh": true, ".bash": true, ".js": true, ".ts": true,
	".html": true, ".htm": true, ".css": true, ".scss": true, ".less": true,
	".txt": true, ".log": true, ".md": true, ".json": true, ".yaml": true, ".yml": true,
	".xml": true, ".toml": true, ".env": true, ".sql": true, ".py": true,
	".c": true, ".cpp": true, ".h": true, ".hpp": true, ".java": true, ".rb": true,
	".php": true, ".pl": true, ".pm": true, ".config": true, ".conf": true,
	".gitignore": true, ".editorconfig": true, ".prettierrc": true,
}

var tmpl *template.Template

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	var errTemplate error
	tmpl, errTemplate = template.ParseFiles("templates/index.html.tmpl")
	if errTemplate != nil {
		fmt.Println("Error loading template:", errTemplate)
		os.Exit(1)
	}

	http.HandleFunc("/", handleMain)

	os.Setenv("GEMINI_API_KEY", "")
	fmt.Println("GEMINI_API_KEY set to:", os.Getenv("GEMINI_API_KEY"))

	http.HandleFunc("/api/gemini", handleGeminiAPI)
	http.HandleFunc("/api/get-file", handleGetFile)
	http.HandleFunc("/api/save-file", handleSaveFile)

	url := "http://localhost:76076"
	fmt.Printf("Click here to open in your browser: %s\n", url)

	http.ListenAndServe(":76076", nil)
}

func isInteractiveCommand(cmdStr string) bool {
	interactiveCommands := []string{
		"nano", "vim", "vi", "emacs", "less", "more", "top", "htop",
		"man", "ssh", "ftp", "telnet", "mysql", "psql", "mongo",
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return false
	}

	cmdName := filepath.Base(parts[0])
	for _, interactive := range interactiveCommands {
		if cmdName == interactive {
			return true
		}
	}
	return false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := info.Mode()
	return mode&0111 != 0
}

func getFileType(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "unknown"
	}

	content := string(buffer[:n])

	if strings.HasPrefix(content, "#!/bin/bash") || strings.HasPrefix(content, "#!/usr/bin/bash") {
		return "bash"
	}
	if strings.HasPrefix(content, "#!/bin/sh") || strings.HasPrefix(content, "#!/usr/bin/sh") {
		return "shell"
	}
	if strings.HasPrefix(content, "#!/usr/bin/python") || strings.HasPrefix(content, "#!/bin/python") {
		return "python"
	}
	if strings.HasPrefix(content, "#!/usr/bin/node") || strings.HasPrefix(content, "#!/bin/node") {
		return "node"
	}

	if len(buffer) >= 4 && buffer[0] == 0x7f && buffer[1] == 'E' && buffer[2] == 'L' && buffer[3] == 'F' {
		return "binary"
	}

	if strings.Contains(content, "Go build ID:") || strings.Contains(content, "runtime.") {
		return "go-binary"
	}

	return "executable"
}

func executeCommand(cmdStr, workDir string) (string, error, int, bool) {
	if isInteractiveCommand(cmdStr) {
		return "Interactive command detected: " + cmdStr + "\n\nNote: Interactive commands like nano, vim, ssh, etc. are not supported in this web terminal.\nUse non-interactive alternatives when possible.", nil, 0, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	var output bytes.Buffer
	var stderr bytes.Buffer

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil, 0, false
	}

	cmdName := parts[0]
	args := parts[1:]

	if strings.Contains(cmdName, "/") || strings.HasPrefix(cmdName, "./") {
		fullPath := cmdName
		if !filepath.IsAbs(cmdName) {
			fullPath = filepath.Join(workDir, cmdName)
		}

		if isExecutable(fullPath) {
			fileType := getFileType(fullPath)

			switch fileType {
			case "bash":
				cmd = exec.CommandContext(ctx, "bash", fullPath)
				if len(args) > 0 {
					cmd.Args = append(cmd.Args, args...)
				}
			case "shell":
				cmd = exec.CommandContext(ctx, "sh", fullPath)
				if len(args) > 0 {
					cmd.Args = append(cmd.Args, args...)
				}
			case "python":
				cmd = exec.CommandContext(ctx, "python3", fullPath)
				if len(args) > 0 {
					cmd.Args = append(cmd.Args, args...)
				}
			case "node":
				cmd = exec.CommandContext(ctx, "node", fullPath)
				if len(args) > 0 {
					cmd.Args = append(cmd.Args, args...)
				}
			default:
				cmd = exec.CommandContext(ctx, fullPath, args...)
			}
		} else {
			cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
		}
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	cmd.Dir = workDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err, 1, false
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err, 1, false
	}

	if err := cmd.Start(); err != nil {
		return "", err, 1, false
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			output.WriteString(scanner.Text() + "\n")
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			stderr.WriteString(scanner.Text() + "\n")
		}
	}()

	err = cmd.Wait()
	wg.Wait()

	combinedOutput := output.String()
	if stderr.Len() > 0 {
		if combinedOutput != "" {
			combinedOutput += "\n"
		}
		combinedOutput += stderr.String()
	}

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else if ctx.Err() == context.DeadlineExceeded {
			return combinedOutput + "\n[Command timed out after 30 seconds]", err, 124, false
		}
	}

	return combinedOutput, err, exitCode, false
}

func constructAPIURL(model string) string {
	return fmt.Sprintf("%s/%s%s", geminiAPIURLBase, model, geminiAPIGenerate)
}

func handleError(err error, message string) error {
	if err != nil {
		return fmt.Errorf("%s: %v", message, err)
	}
	return nil
}

func createTextPart(text string) map[string]interface{} {
	return map[string]interface{}{
		"text": text,
	}
}

func createFileDataPart(fileURLStr string) (map[string]interface{}, error) {
	_, err := url.ParseRequestURI(fileURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid file URL: %v", err)
	}

	resp, err := http.Get(fileURLStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file from URL %s: %v", fileURLStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch file: status code %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %v", err)
	}

	var mimeType string
	headerMime := resp.Header.Get("Content-Type")
	if parsedMime, _, err := mime.ParseMediaType(headerMime); err == nil && parsedMime != "" {
		mimeType = parsedMime
	}

	if mimeType == "" || mimeType == "application/octet-stream" {
		detectedMime := http.DetectContentType(fileData)
		if parsedMime, _, err := mime.ParseMediaType(detectedMime); err == nil && parsedMime != "" {
			mimeType = parsedMime
		} else if detectedMime != "" {
			mimeType = detectedMime
		}
	}

	if mimeType == "" || mimeType == "application/octet-stream" {
		parsedURL, _ := url.Parse(fileURLStr)
		ext := filepath.Ext(parsedURL.Path)
		guessedMime := mime.TypeByExtension(ext)
		if guessedMime != "" {
			mimeType = guessedMime
		}
	}

	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = "application/octet-stream"
	}

	encodedData := base64.StdEncoding.EncodeToString(fileData)

	return map[string]interface{}{
		"inline_data": map[string]interface{}{
			"mime_type": mimeType,
			"data":      encodedData,
		},
	}, nil
}

func sendToGeminiAPIInternal(apiKey, systemPrompt string, currentHistory []map[string]interface{}, settings AISettings) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf(ErrMissingAPIKey)
	}
	if systemPrompt == "" {
		return "", fmt.Errorf(ErrMissingPrompt)
	}
	if len(currentHistory) == 0 {
		return "", fmt.Errorf("geminiclient: cannot send empty history to Gemini API")
	}

	genConfig := map[string]interface{}{
		"temperature":     settings.Temperature,
		"maxOutputTokens": settings.MaxTokens,
	}
	if settings.TopK > 0 {
		genConfig["topK"] = settings.TopK
	}
	if settings.TopP > 0.0 {
		genConfig["topP"] = settings.TopP
	}

	payload := map[string]interface{}{
		"contents":         currentHistory,
		"generationConfig": genConfig,
		"system_instruction": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemPrompt},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", handleError(err, "geminiclient: failed to marshal payload")
	}

	requestURL := fmt.Sprintf("%s?key=%s", constructAPIURL("gemini-2.0-flash"), apiKey)

	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", handleError(err, "geminiclient: failed to send request to Gemini API")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", handleError(err, "geminiclient: failed to read response body")
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("geminiclient: Gemini API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var geminiResponse GeminiResponse
	if err := json.Unmarshal(responseBody, &geminiResponse); err != nil {
		return "", handleError(err, "geminiclient: failed to decode Gemini API response")
	}

	if len(geminiResponse.Candidates) > 0 &&
		len(geminiResponse.Candidates[0].Content.Parts) > 0 &&
		geminiResponse.Candidates[0].Content.Parts[0].Text != "" {
		return geminiResponse.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf(ErrInvalidResponse)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := TemplateData{
		CurrentPath: "/data/data/com.termux/files/home",
	}

	queryPath := r.URL.Query().Get("path")
	if queryPath != "" {
		data.CurrentPath = queryPath
	}

	data.CurrentPath = filepath.Clean(data.CurrentPath)

	cmdStr := strings.TrimSpace(r.URL.Query().Get("cmd"))

	if cmdStr != "" {
		data.Command = cmdStr
		data.CommandExecuted = true

		if strings.HasPrefix(cmdStr, "cd ") {
			target := strings.TrimSpace(cmdStr[3:])
			if !filepath.IsAbs(target) {
				target = filepath.Join(data.CurrentPath, target)
			}
			target = filepath.Clean(target)

			if fi, err := os.Stat(target); err == nil && fi.IsDir() {
				// A simple check to prevent 'cd ../../..' out of the base,
				// but this needs a defined base directory for robustness.
				// For now, just setting the path.
				data.CurrentPath = target // Update path if cd is successful
				data.Output = "Changed directory to: " + data.CurrentPath
				data.ExitCode = 0
			} else {
				data.Output = "cd: no such directory: " + target
				data.Error = fmt.Sprintf("cd: no such directory: %s", target)
				data.ExitCode = 1
			}
			data.IsInteractive = false

		} else {
			output, cmdErr, exitCode, isInteractive := executeCommand(cmdStr, data.CurrentPath)

			data.Output = output
			data.Error = ""
			if cmdErr != nil {
				data.Error = cmdErr.Error()
			}
			data.ExitCode = exitCode
			data.IsInteractive = isInteractive

			parts := strings.Fields(cmdStr)
			if len(parts) > 0 {
				cmdName := parts[0]
				if strings.Contains(cmdName, "/") || strings.HasPrefix(cmdName, "./") {
					fullPath := cmdName
					if !filepath.IsAbs(cmdName) {
						fullPath = filepath.Join(data.CurrentPath, cmdName)
					}
					fullPath = filepath.Clean(fullPath)

					if isExecutable(fullPath) {
						data.IsExecutable = true
						data.CommandName = cmdName
						data.FileType = getFileType(fullPath)
					}
				}
			}
		}
	}

	entries, err := ioutil.ReadDir(data.CurrentPath)
	if err != nil {
		data.DirError = err.Error()
	} else {
		data.Entries = make([]DirEntry, 0, len(entries))
		for _, e := range entries {
			name := e.Name()
			fullPath := filepath.Join(data.CurrentPath, name)
			fullPath = filepath.Clean(fullPath)

			entryData := DirEntry{
				Name:       name,
				IsDir:      e.IsDir(),
				FullPath:   fullPath,
				ParentPath: data.CurrentPath,
			}

			if !e.IsDir() {
				entryData.IsExecutable = isExecutable(fullPath)
				if entryData.IsExecutable {
					entryData.FileType = getFileType(fullPath)
				} else {
					ext := strings.ToLower(filepath.Ext(name))
					entryData.IsEditable = codeFileExtensions[ext]
				}
			}

			data.Entries = append(data.Entries, entryData)
		}
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func handleGetFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "File path is required",
		})
		return
	}

	filePath = filepath.Clean(filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "File not found or accessible: " + err.Error(),
		})
		return
	}

	if info.IsDir() {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot edit a directory",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if !codeFileExtensions[ext] {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Cannot edit binary or non-code file type: %s", ext),
		})
		return
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to read file: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content": string(content),
	})
}

type SaveFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func handleSaveFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	filePath := req.Path
	content := req.Content

	if filePath == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "File path is required",
		})
		return
	}

	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Absolute file path is required",
		})
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Error accessing file: " + err.Error(),
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"error": "File does not exist",
		})
		return
	}

	if info.IsDir() {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Cannot save to a directory",
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if !codeFileExtensions[ext] {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Cannot save binary or non-code file type: %s", ext),
		})
		return
	}

	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to write file: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

func handleGeminiAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "GEMINI_API_KEY environment variable not set. Please set it with your Google AI Studio API key.",
		})
		return
	}

	var req GeminiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
		})
		return
	}

	if len(req.Contents) == 0 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: 'contents' array is empty",
		})
		return
	}

	geminiReq := req

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to prepare request",
		})
		return
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to connect to Gemini API: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to read response",
		})
		return
	}

	if resp.StatusCode != 200 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Gemini API error: " + string(body),
		})
		return
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to parse Gemini response",
		})
		return
	}

	if len(geminiResp.Candidates) == 0 ||
		len(geminiResp.Candidates[0].Content.Parts) == 0 ||
		geminiResp.Candidates[0].Content.Parts[0].Text == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": ErrInvalidResponse,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"response": geminiResp.Candidates[0].Content.Parts[0].Text,
	})
}
