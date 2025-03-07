# 後勤平台管理系統 - Socket 傳輸

## Docker 部屬

1. 需先安裝 Docker
2. 終端指令建置

   ```shell
   make docker-build
   ```

3. 終端指令運行

   ```shell
   make docker-up
   ```

## 建置工具

Windows 系統使用 `build.bat`，Linux 和 MacOS 使用 `Makefile`檔案，各自皆可使用 `Help` 查詢 command 指令。

## 用戶端

### DLL 建置

建置成動態連結庫(DLL)後提供給各軟體使用。針對 32 和 64 位元可建置不同的 DLL。

需要設定 go env 編譯不同平台:

Windows/x86:

```shell
go env -w GOOS=windows
go env -w GOARCH=386
```

Linux/amd64:

```shell
go env -w GOOS=linux
go env -w GOARCH=amd64
```

- 32 位元使用 `client32` 指令建置，建置後檔案會在 `/build/client/386`
- 64 位元使用 `client64` 指令建置，建置後檔案會在 `/build/client/amd64`

以上需注意電腦系統環境變數是否支援 32 或 64 位元，可使用 `msys64` 下載 `mingw32` 及 `mingw64` 後，添加到系統變數的 Path 內，新增到 /bin 位置即可。
目前在建置 32 或 64 位元時，都需要調整 Path 內上下位置才可建置相對應位元的 DLL。

### 主程式

`cmd/client/main.go` 為建置用戶端主程式，須注意:

1. 最上層多行註解的部分為定義 `c/c++` 程式碼，`cgo` 允許在 `.go` 原始碼撰寫
2. 基本 `import "C"` 需要在緊接在註解區塊之下，有時候引用到原未使用過的 C 的型態，整體程式碼為很多錯誤，可以刪除 `import "C"` 再次引入會恢復正常。
3. 在需要匯出的 function 上要註解 `//export your_function`，函式才會被匯出。
