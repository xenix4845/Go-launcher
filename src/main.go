package main

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// Constants for URLs and file paths
const (
    updateURL      = "https://stdmenu.kro.kr/launcher"
    updateFileName = "updater.exe"
)

func main() {
    // Create a new Fyne application
    a := app.New()
    w := a.NewWindow("Updater")

    // Create a text widget to display console output
    output := widget.NewMultiLineEntry()
    output.SetReadOnly(true)

    // Create an entry widget to accept user input
    input := widget.NewEntry()

    // Create a container to hold the output and input widgets
    content := container.NewVBox(output, input)
    w.SetContent(content)

    // Function to append text to the output widget
    appendOutput := func(text string) {
        output.SetText(output.Text + text + "\n")
    }

    // Get the AppData path
    appDataPath, err := os.UserConfigDir()
    if err != nil {
        appendOutput(fmt.Sprintf("시스템 폴더를 찾는데 실패했습니다: %v", err))
        return
    }

    // Define the path to save the updater
    stdMenuPath := filepath.Join(appDataPath, "STDMENU")
    if _, err := os.Stat(stdMenuPath); os.IsNotExist(err) {
        if err := os.MkdirAll(stdMenuPath, os.ModePerm); err != nil {
            appendOutput(fmt.Sprintf("디렉토리를 생성하는데 실패했습니다: %v", err))
            return
        }
    }
    updateFilePath := filepath.Join(stdMenuPath, updateFileName)

    // Download the updater
    resp, err := http.Get(updateURL)
    if err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 다운로드하는데 실패했습니다: %v", err))
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        appendOutput(fmt.Sprintf("서버 응답 오류: %d", resp.StatusCode))
        return
    }

    out, err := os.Create(updateFilePath)
    if err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 생성하는데 실패했습니다: %v", err))
        return
    }

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 저장하는데 실패했습니다: %v", err))
        out.Close()
        return
    }

    // Close the file before executing
    out.Close()

    // Execute the updater and capture its output
    cmd := exec.Command(updateFilePath)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 실행하는데 실패했습니다: %v", err))
        return
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 실행하는데 실패했습니다: %v", err))
        return
    }

    if err := cmd.Start(); err != nil {
        appendOutput(fmt.Sprintf("업데이트 파일을 실행하는데 실패했습니다: %v", err))
        return
    }

    // Function to read and display command output
    go func() {
        reader := bufio.NewReader(io.MultiReader(stdout, stderr))
        for {
            line, err := reader.ReadString('\n')
            if err != nil {
                if err != io.EOF {
                    appendOutput(fmt.Sprintf("출력을 읽는 중 오류가 발생했습니다: %v", err))
                }
                break
            }
            appendOutput(strings.TrimSpace(line))
        }
    }()

    // Function to handle user input
    input.OnSubmitted = func(text string) {
        if _, err := cmd.StdinPipe().Write([]byte(text + "\n")); err != nil {
            appendOutput(fmt.Sprintf("입력을 처리하는 중 오류가 발생했습니다: %v", err))
        }
        input.SetText("")
    }

    // Wait for the updater to finish
    go func() {
        if err := cmd.Wait(); err != nil {
            appendOutput(fmt.Sprintf("업데이트 파일 실행 중 오류가 발생했습니다: %v", err))
        } else {
            appendOutput("업데이터가 완료되었습니다. 프로그램을 종료합니다.")
        }
    }()

    // Show the window and run the application
    w.ShowAndRun()
}