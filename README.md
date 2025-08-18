## 設計概念
這個專案是一個簡單的文章推薦系統，主要功能包括文章的 CRUD (建立、讀取、更新、刪除)、使用者的註冊和登入、文章的評分和推薦系統。專案使用 Go 語言實作，並遵循 Clean Architecture 和 Domain-Driven Design (DDD) 的原則。
*   **使用 Restful API**
    *   使用 Gin 框架 實作 RESTful API 介面
    *   大部分 API 都具有 冪等性

*   **容器化 (Containerization):**
    *   可容器化部署，便於環境搭建和擴展。

*   **資料庫管理與優化:**
    *   **PostgreSQL:** 選用 PostgreSQL 作為主要資料庫。
    *   **SQL Migrations:** 使用 `migrations` 目錄下的 SQL 腳本管理資料庫 Schema 變更，確保資料庫版本控制的穩定性。
    *   **Materialized Views (物化視圖):** 為了優化推薦系統的效能，利用物化視圖在離峰時預先計算文章的平均評分，提升查詢效率。
    *   **UUID 優化:** 使用 UUIDv7 作為文章的唯一識別符，避免了傳統 UUID 的效率問題，並且能夠在分布式系統中保持唯一性。
    *   **排程任務:** 在應用程式內部實作排程任務，例如定期刷新物化視圖和執行背景爬取工作，之後也能設定執行時間，讓計算量大的任務在離峰時執行，或者是透過選舉機制，確保多個 pod 的情況下，也只會有一個排程任務在執行，不會有多個排程任務在執行。

*  **文章 Metadata 抓取:**
    *   用 URL 做 ID，即便多位使用者同時提交相同的文章 URL，也只需要處理一次，有效緩解熱點文章重複抓取的問題。
    *   抓取文章的 Metadata (如 title、description、image_url)，並儲存到資料庫中。
    *   當抓取失敗時，會將錯誤記錄到資料庫，並透過背景工作重試機制進行重試。
    *   非同步抓取 Metadata
        *   正式環境應使用 Message Queue (如 Pubsub) 來處理非同步任務，但在這個專案中，為了簡化實作，直接在應用程式內部進行排程和非同步處理。
        *   有些動態網站需要 JavaScript 渲染，這種情況下可以使用 headless browser 來抓取文章，但在這個專案中，為了簡化實作，直接進行靜態抓取。
    *   (未實作) 遇到使用者惡意提交不存在的 URL 時，可以考慮使用 bloom filter 做過濾，避免不必要的抓取。

*  **文章推薦系統:**
    *   正式環境中會有專門的服務來處理推薦系統，為了簡化，推薦邏輯直接簡單的實現在應用程式中。
        *   演算法是 推薦使用者全站平均最高分的文章，並排除已經收藏的文章。
        *   使用 postgresql 物化視圖預先計算儲存文章的平均評分，並在推薦時查詢這些資料，這項作業可以在離峰時非同步執行，加快推薦的效率。

*   **認證與授權 (Authentication & Authorization):**
    *   實作了基於 JWT (JSON Web Token) 的認證系統，支援使用者註冊、登入和個人資訊查詢。
    *   密碼經過雜湊處理，確保安全性。

*   **Clean Architecture (整潔架構):**
    *   遵循 Clean Architecture 原則，分成 `domain` (核心業務實體與規則)、`app` (應用層業務邏輯/Use Cases)、`adapter` (外部介面實作，如資料庫、外部服務) 和 `router` (API 介面) 等層次，達到關注點分離
    *   這使得業務邏輯與外部框架、資料庫等細節解耦，提高了程式碼的可測試性、可維護性和可擴展性

*   **Domain-Driven Design (領域驅動設計 - DDD):**
    *   `internal/domain` 包含 `article` 和 `user` 等子領域，每個子領域包含其核心實體和業務規則。這有助於更好地建模複雜的業務領域，並確保程式碼與業務語言的一致性。


## DB 設計

**主要表格：**

*   **`users`**：儲存使用者資訊，包括 `id` 、 `username`、`email` 和 `password_hash`。
*   **`articles`**：儲存文章詳細資訊，例如 `url`、`title`、`description`、`image_url` 和 `metadata` (JSONB)。`url` 是唯一的，並作為文章的ID。
*   **`user_articles`**：一個連接 `users` 和 `articles` 的聯結表，記錄使用者對文章的收藏和 `rate` (0-5 分)。 使用 `user_id` 和 `article_id` 做 UniqueID，以防止重複條目。
*   **`metadata_fetch_retries`**：管理文章抓取的重試機制，包含 `retry_count`、`last_attempt_at`、`next_attempt_at`、`status` 和 `error_message`。


## 資料夾結構
*   `cmd/`: 應用程式的進入點
    *   `app/`: 包含主要的應用程式啟動邏輯，如 `main.go`
*   `internal/`: 按照 Clean Architecture 分層
    *   `adapter/`: 外部介面實作層。
        *   `repository/`: 資料庫儲存庫的實作，目前主要為 `postgres` (PostgreSQL)。
    *   `app/`: 應用層，包含應用程式特定的業務規則和 Use Cases。
        *   `application.go`: 應用程式的初始化和依賴注入。
        *   `service/`: 核心業務服務的定義和實作，如 `article` 和 `user` 
    *   `domain/`: 領域層，包含企業級的業務規則和實體。
        *   `article/`: 文章領域的實體和行為。
        *   `common/`: 通用領域定義，如錯誤碼。
        *   `user/`: 使用者領域的實體和行為。
    *   `router/`: 路由層，處理 HTTP 請求的路由和處理器 (Handler)。
*   `migrations/`: 資料庫遷移腳本，用於管理資料庫 Schema 的版本演進。


## API
- `api.md`: API 文檔
- `postman.json`: Postman 的 API 集合檔案

## Setup
`docker compose up -d`
port `9000`

