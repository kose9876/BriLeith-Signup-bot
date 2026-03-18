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

3. 編輯 `config.json`，填入你的 Bot 設定：

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

如果首次執行缺少檔案，程式會自動建立以下 JSON：

- `profiles.json`
- `signups.json`
- `test_signups.json`
- `admin_state.json`
- `signup_schedule_state.json`

所有自動產生的 JSON 檔都使用 UTF-8 編碼與 2 空白縮排。

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
- `/admin_profile user:@玩家`：查看玩家 profile 與報名資訊
- `/admin_addplayer`：建立新玩家資料
- `/admin_setrole`：編輯既有玩家的遊戲名稱、主副職、群組、破袍
- `/admin_setgamename`：只修改既有玩家的遊戲名稱
- `/admin_removeplayer`：移除玩家，可選是否刪除本週報名
- `/admin_signup`：手動幫玩家報名某一天
- `/admin_unsignup`：手動取消玩家某一天報名
- `/admin_signup_access`：設定玩家能否自行報名
- `/admin_grant`：將已註冊玩家加入管理員
- `/admin_revoke`：移除動態管理員權限
- `/admin_test_signup_post`：手動發送測試用報名表
- `/admin_test_summary`：查看測試報名與分配輸出

## 注意事項

- `config.json` 不應提交到 GitHub。
- 若你用 Windows 編輯 JSON，專案目前可容忍 UTF-8 BOM，不會因此讀取失敗。
- 如果 Slash Command 沒有更新，請確認 Bot 有在 `guild_ids` 指定的伺服器內，並重新啟動程式。
