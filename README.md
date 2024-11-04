# My Go Web App

這是一個用 Go 語言製作的網頁應用程式，提供基本的 CRUD 操作、使用者驗證，以及使用 WebSocket 實現的即時更新功能。

## 功能
- 使用者登入和註冊
- 使用 WebSocket 即時顯示時間更新
- 使用者資料的 CRUD 操作

## 安裝與運行
下載並運行此應用程式：
```bash
docker pull kaiji425/myapp
docker run -p 8080:8080 kaiji425/myapp
