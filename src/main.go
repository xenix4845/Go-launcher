package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
)

const (
    updateURL      = "https://stdmenu.kro.kr/launcher"
    updateFileName = "updater.exe"
)

func main() {
    // 1. %appdata%/STDMENU 디렉토리 경로 설정
    appDataPath, err := os.UserConfigDir()
    if err != nil {
        fmt.Println("시스템 폴더를 찾는데 실패했습니다:", err)
        os.Exit(1)
    }
    stdMenuPath := filepath.Join(appDataPath, "STDMENU")
    if _, err := os.Stat(stdMenuPath); os.IsNotExist(err) {
        if err := os.MkdirAll(stdMenuPath, 0755); err != nil {
            fmt.Println("디렉토리를 생성하는데 실패했습니다:", err)
            os.Exit(1)
        }
    }
    updateFilePath := filepath.Join(stdMenuPath, updateFileName)

    // 2. 업데이트 파일 다운로드
    resp, err := http.Get(updateURL)
    if err != nil {
        fmt.Println("업데이트 파일을 다운로드하는데 실패했습니다:", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("서버 응답 오류: %d\n", resp.StatusCode)
        os.Exit(1)
    }

    out, err := os.Create(updateFilePath)
    if err != nil {
        fmt.Println("업데이트 파일을 생성하는데 실패했습니다:", err)
        os.Exit(1)
    }

    _, err = io.Copy(out, resp.Body)
    if err != nil {
        fmt.Println("업데이트 파일을 저장하는데 실패했습니다:", err)
        out.Close()
        os.Exit(1)
    }

    // 파일을 닫기
    out.Close()

    fmt.Println("업데이트 다운로드가 완료되었습니다. 업데이터를 실행합니다.")

    // 3. cmd.exe를 사용하여 updater.exe를 새 콘솔 창에서 실행
    cmd := exec.Command("cmd", "/c", "start", updateFilePath)

    err = cmd.Start()
    if err != nil {
        fmt.Println("업데이트 파일을 실행하는데 실패했습니다:", err)
        os.Exit(1)
    }

    // 프로그램 종료
    os.Exit(0)
}
