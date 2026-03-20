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
- `boss3_assignments.json`
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

- `/setrole name:<遊戲名稱>`：設定自己的遊戲名稱、主職、副職、群組與破袍
- `/signup`：開啟每週報名面板
- `/summary`：查看本週分配摘要
- `/whatrole user:@玩家`：查看指定玩家的職業資料
- `/help`：查看指令說明

## 管理員指令

- `/a_list`：查看管理總覽
- `/a_list_players`：列出所有已註冊玩家
- `/a_profile player:<玩家>`：查看玩家 profile 與報名資訊
- `/a_addplayer user_id:<Discord ID 或 @mention> game_name:<遊戲名稱> [display_name:<顯示名>]`：建立新玩家資料
- `/a_setrole player:<玩家>`：編輯既有玩家的遊戲名稱、主副職、群組、破袍
- `/a_setgamename player:<玩家> game_name:<遊戲名稱>`：只修改既有玩家的遊戲名稱
- `/a_removeplayer player:<玩家>`：移除玩家，可選是否刪除本週報名
- `/a_signup player:<玩家> day:<日期>`：手動幫玩家報名某一天
- `/a_unsignup player:<玩家> day:<日期>`：手動取消玩家某一天報名
- `/a_boss3_assign day:<日期> task:<工作> mode:<換位|追加兼任> player:<玩家>`：手動調整正式版三王工作分配
- `/a_boss3_clear day:<日期> task:<工作>`：清除正式版三王工作分配覆寫
- `/a_signup_access player:<玩家> blocked:<true|false>`：設定玩家能否自行報名
- `/a_grant player:<玩家>`：將已註冊玩家加入管理員
- `/a_grant_tester player:<玩家>`：將已註冊玩家加入測試員
- `/a_revoke player:<玩家>`：移除動態管理員權限
- `/a_revoke_tester player:<玩家>`：移除測試員權限
- `/a_summary_image day:<日期>`：輸出正式報名的表格圖片

## 測試員指令

- `/a_profile player:<玩家>`：查看玩家 profile 與報名資訊
- `/a_list_players`：列出所有已註冊玩家
- `/t_signup player:<玩家> day:<日期>`：手動幫玩家加入測試報名某一天
- `/t_unsignup player:<玩家> day:<日期>`：手動取消玩家某一天的測試報名
- `/t_signup_post`：手動發送測試用報名表
- `/t_boss3_assign day:<日期> task:<工作> mode:<換位|追加兼任> player:<玩家>`：手動調整測試版三王工作分配
- `/t_boss3_clear day:<日期> task:<工作>`：清除測試版三王工作分配覆寫
- `/t_summary`：查看測試報名與分配輸出
- `/t_summary_image day:<日期>`：輸出測試報名的表格圖片

### 玩家查找規則

- `player` 參數可接受 Discord `user ID`、`@mention`、`遊戲名稱`、`顯示名稱`
- 若名稱重複，系統會提示候選名單，這時請改用 `user ID`
- `a_addplayer` 的 `user_id` 也可直接填 Discord `user ID` 或 `@mention`

### 離開 Server 的玩家

- 管理員指令不再依賴目前 server 成員清單
- 只要玩家已存在於 `profiles.json`，即使已離開 server，仍可用 `/a_profile`、`/a_signup`、`/a_unsignup`、`/a_setrole` 等指令管理
- 若玩家已離開 server、但尚未建檔，可先用 `/a_addplayer user_id:<Discord ID> game_name:<遊戲名稱> display_name:<顯示名>` 建立資料，再進行報名操作

### 測試版三王手動分配

- 這一版只先套用在測試資料，不影響正式報名
- `/t_summary` 與 `/t_summary_image` 會讀取同一份測試版三王工作覆寫紀錄
- 手動指定時，玩家必須已經在該日期的測試報名名單內
- `mode:換位` 會與原本工作的人交換
- `mode:追加兼任` 會讓同一位玩家同時兼任多個三王工作
- 覆寫資料會存到 `test_boss3_assignments.json`

### 正式版三王手動分配

- `/a_boss3_assign` 與 `/a_boss3_clear` 會直接影響正式版摘要與正式版圖片
- `/summary` 與 `/a_summary_image` 會讀取同一份正式版三王工作覆寫紀錄
- 若玩家已不在該日期正式名單，過期覆寫會自動失效並清除
- 覆寫資料會存到 `boss3_assignments.json`

## 注意事項
- 若你用 Windows 編輯 JSON，專案目前可容忍 UTF-8 BOM，不會因此讀取失敗。
- 如果 Slash Command 沒有更新，請確認 Bot 有在 `guild_ids` 指定的伺服器內，並重新啟動程式。
