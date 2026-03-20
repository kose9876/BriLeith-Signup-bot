# bri 安裝與指令說明

## 環境需求

- Go 1.22 以上
- 一個已建立的 Discord Bot
- Bot 已加入目標伺服器，且有建立 Slash Commands 的權限

## 安裝步驟

1. 下載專案：

```powershell
git clone <你的-repo-url>
cd bri
```

2. 建立設定檔：

```powershell
Copy-Item config.example.json config.json
```

3. 你可以直接手動編輯 `config.json`，或直接執行程式後依照 CLI 提示完成初始化：

```json
{
  "token": "你的-bot-token",
  "application_id": "你的-application-id",
  "guild_id": "主要-guild-id",
  "guild_ids": ["guild-id-1"],
  "admin_user_ids": ["你的-discord-user-id"]
}
```

## 啟動方式

在專案目錄執行：

```powershell
go run .
```

如果 `config.json` 尚未填完整，程式會在 CLI 逐步要求輸入：

- `Discord Bot Token`
- `Discord Application ID`
- `主要 Guild ID`
- `管理員 Discord User ID`

如果首次執行缺少檔案，程式也會自動建立以下 JSON：

- `profiles.json`
- `signups.json`
- `test_signups.json`
- `admin_state.json`
- `signup_schedule_state.json`

所有自動產生的 JSON 檔都使用 UTF-8 編碼與 2 空白縮排。

程式執行中，若有人使用 slash command 或按下按鈕，CLI 也會顯示簡單紀錄，方便確認目前是哪位使用者觸發了哪個操作。

## 設定欄位說明

- `token`：Discord Bot Token
- `application_id`：Discord Application ID
- `guild_id`：單一伺服器 ID，舊格式相容用
- `guild_ids`：可註冊指令的伺服器 ID 陣列
- `admin_user_ids`：固定管理員的 Discord User ID 陣列

`guild_id` 與 `guild_ids` 至少要有一個。

## 一般玩家指令

- `/setgamename name:<遊戲名稱>`：設定自己的遊戲名稱
- `/setrole`：設定自己的主職、副職、群組與破袍
- `/signup`：開啟每週報名面板
- `/summary`：查看本週分配摘要
- `/whatrole user:@玩家`：查看指定玩家的職業資料
- `/help`：查看指令說明

## 管理員指令

- `/admin_list`：查看管理總覽
- `/admin_list_players`：列出所有已註冊玩家
- `/admin_profile player:<玩家>`：查看玩家 profile 與報名資訊
- `/admin_addplayer user_id:<Discord ID 或 @mention> game_name:<遊戲名稱> [display_name:<顯示名>]`：建立新玩家資料
- `/admin_setrole player:<玩家>`：編輯既有玩家的遊戲名稱、主副職、群組、破袍
- `/admin_setgamename player:<玩家> game_name:<遊戲名稱>`：只修改既有玩家的遊戲名稱
- `/admin_removeplayer player:<玩家>`：移除玩家，可選是否刪除本週報名
- `/admin_signup player:<玩家> day:<日期>`：手動幫玩家報名某一天
- `/admin_unsignup player:<玩家> day:<日期>`：手動取消玩家某一天報名
- `/admin_test_signup player:<玩家> day:<日期>`：手動幫玩家加入測試報名某一天
- `/admin_test_unsignup player:<玩家> day:<日期>`：手動取消玩家某一天的測試報名
- `/admin_signup_access player:<玩家> blocked:<true|false>`：設定玩家能否自行報名
- `/admin_grant player:<玩家>`：將已註冊玩家加入管理員
- `/admin_revoke player:<玩家>`：移除動態管理員權限
- `/admin_test_signup_post`：手動發送測試用報名表
- `/admin_test_summary`：查看測試報名與分配輸出
- `/admin_summary_image day:<日期>`：輸出正式報名的表格圖片
- `/admin_test_summary_image day:<日期>`：輸出測試報名的表格圖片

### 玩家查找規則

- `player` 參數可接受 Discord `user ID`、`@mention`、`遊戲名稱`、`顯示名稱`
- 若名稱重複，系統會提示候選名單，這時請改用 `user ID`
- `admin_addplayer` 的 `user_id` 也可直接填 Discord `user ID` 或 `@mention`

### 離開 Server 的玩家

- 管理員指令不再依賴目前 server 成員清單
- 只要玩家已存在於 `profiles.json`，即使已離開 server，仍可用 `/admin_profile`、`/admin_signup`、`/admin_unsignup`、`/admin_setrole` 等指令管理
- 若玩家已離開 server、但尚未建檔，可先用 `/admin_addplayer user_id:<Discord ID> game_name:<遊戲名稱> display_name:<顯示名>` 建立資料，再進行報名操作

## 注意事項
- 若你用 Windows 編輯 JSON，專案目前可容忍 UTF-8 BOM，不會因此讀取失敗。
- 如果 Slash Command 沒有更新，請確認 Bot 有在 `guild_ids` 指定的伺服器內，並重新啟動程式。
