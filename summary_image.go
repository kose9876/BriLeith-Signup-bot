package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type summaryImageRow struct {
	Player string
	Boss1  string
	Boss2  string
	Boss3A string
	Boss3B string
}

type summaryImageData struct {
	DateLabel string
	Boss1Sub  string
	Boss2Sub  string
	Boss3SubA string
	Boss3SubB string
	Rows      []summaryImageRow
}

func handleAdminTestSummaryImageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	dayOption := findOption(i.ApplicationCommandData().Options, "day")
	if dayOption == nil {
		respondEphemeral(s, i, "缺少 day 參數。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	dayKey := dayOption.StringValue()

	data := buildTestSummaryImageData(weekKey, dayKey)
	imageBytes, err := renderSummaryImagePNG(data)
	if err != nil {
		respondEphemeral(s, i, "產生表格圖片失敗："+err.Error())
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: getWeekRangeText(weekKey) + " " + dayLabels[dayKey] + " 測試表格圖片",
			Files: []*discordgo.File{
				{
					Name:        "test-summary-" + dayKey + ".png",
					ContentType: "image/png",
					Reader:      bytes.NewReader(imageBytes),
				},
			},
		},
	})
	if err != nil {
		fmt.Println("admin test summary image failed:", err)
	}
}

func handleAdminSummaryImageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	dayOption := findOption(i.ApplicationCommandData().Options, "day")
	if dayOption == nil {
		respondEphemeral(s, i, "缺少 day 參數。")
		return
	}

	weekKey := getManagedSignupWeekKey()
	dayKey := dayOption.StringValue()

	data := buildSummaryImageDataFromStore(weeklySignups, weekKey, dayKey)
	imageBytes, err := renderSummaryImagePNG(data)
	if err != nil {
		respondEphemeral(s, i, "產生表格圖片失敗："+err.Error())
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: getWeekRangeText(weekKey) + " " + dayLabels[dayKey] + " 表格圖片",
			Files: []*discordgo.File{
				{
					Name:        "summary-" + dayKey + ".png",
					ContentType: "image/png",
					Reader:      bytes.NewReader(imageBytes),
				},
			},
		},
	})
	if err != nil {
		fmt.Println("admin summary image failed:", err)
	}
}

func buildTestSummaryImageData(weekKey string, dayKey string) summaryImageData {
	return buildSummaryImageDataFromStore(testWeeklySignups, weekKey, dayKey)
}

func buildSummaryImageDataFromStore(signups map[string]map[string][]string, weekKey string, dayKey string) summaryImageData {
	assignment := buildWeekAssignmentFromStore(signups, weekKey)
	day := assignment.Days[dayKey]
	boss1 := assignBoss1(day, map[string][]string{})
	boss2 := assignBoss2Group(day, map[string][]string{})
	boss3 := assignBoss3(day, map[string][]string{})

	boss3Columns := splitBoss3AssignmentsForImage(boss3)
	rows := buildSummaryImageRows(day, boss1, boss2, boss3Columns)

	return summaryImageData{
		DateLabel: buildSummaryImageDateLabel(weekKey, dayKey),
		Boss1Sub:  "一王分配",
		Boss2Sub:  "二王站位",
		Boss3SubA: "80%/40%",
		Boss3SubB: "60%",
		Rows:      rows,
	}
}

func buildSummaryImageDateLabel(weekKey string, dayKey string) string {
	weekdayMark := map[string]string{
		"day_mon": "(一)",
		"day_tue": "(二)",
		"day_wed": "(三)",
		"day_thu": "(四)",
		"day_fri": "(五)",
		"day_sat": "(六)",
		"day_sun": "(日)",
	}

	dateText := strings.ReplaceAll(getDayDateText(weekKey, dayKey), "/", "")
	return dateText + weekdayMark[dayKey]
}

func buildSummaryImageRows(day DayAssignment, boss1 []WorkAssignment, boss2 []GroupAssignment, boss3Columns [2][]WorkAssignment) []summaryImageRow {
	playerOrder := collectSummaryImagePlayers(day)
	rows := make([]summaryImageRow, 0, len(playerOrder))

	for _, userID := range playerOrder {
		rows = append(rows, summaryImageRow{
			Player: getDisplayName(userID),
			Boss1:  collectBoss1AssignmentsForUser(boss1, userID),
			Boss2:  collectBoss2AssignmentsForUser(boss2, userID),
			Boss3A: collectBoss3AssignmentsForUser(boss3Columns[0], userID),
			Boss3B: collectBoss3AssignmentsForUser(boss3Columns[1], userID),
		})
	}

	if len(rows) == 0 {
		rows = append(rows, summaryImageRow{Player: "目前無資料"})
	}

	return rows
}

func collectSummaryImagePlayers(day DayAssignment) []string {
	seen := map[string]bool{}
	var players []string

	add := func(userID string) {
		if userID == "" || userID == "缺坦" || userID == "缺補" || userID == "缺人" || seen[userID] {
			return
		}
		seen[userID] = true
		players = append(players, userID)
	}

	add(day.Tank)
	add(day.Healer)
	for _, userID := range day.DPS {
		add(userID)
	}

	return players
}

func collectBoss1AssignmentsForUser(assignments []WorkAssignment, userID string) string {
	var labels []string
	for _, assignment := range assignments {
		if assignment.UserID == userID {
			labels = append(labels, assignment.Label)
		}
	}
	return strings.Join(labels, "\n")
}

func collectBoss2AssignmentsForUser(groups []GroupAssignment, userID string) string {
	var labels []string
	for _, group := range groups {
		for _, id := range group.UserIDs {
			if id == userID {
				labels = append(labels, group.Label)
				break
			}
		}
	}
	return strings.Join(labels, "\n")
}

func collectBoss3AssignmentsForUser(assignments []WorkAssignment, userID string) string {
	var labels []string
	for _, assignment := range assignments {
		if assignment.UserID == userID {
			labels = append(labels, formatBoss3ImageLabel(assignment.Label))
		}
	}
	return strings.Join(labels, "\n")
}

func formatBoss3ImageLabel(label string) string {
	switch label {
	case "60%坦克工作":
		return "刻印"
	default:
		return label
	}
}

func splitBoss3AssignmentsForImage(assignments []WorkAssignment) [2][]WorkAssignment {
	var columns [2][]WorkAssignment
	for _, assignment := range assignments {
		if isBoss3SixtyPercentTask(assignment.Label) {
			columns[1] = append(columns[1], assignment)
			continue
		}
		columns[0] = append(columns[0], assignment)
	}
	return columns
}

func isBoss3SixtyPercentTask(label string) bool {
	switch label {
	case "60%坦克工作":
		return true
	default:
		return false
	}
}

func renderSummaryImagePNG(data summaryImageData) ([]byte, error) {
	face, err := loadSummaryImageFontFace(14)
	if err != nil {
		return nil, err
	}
	defer face.Close()

	headerFace, err := loadSummaryImageFontFace(17)
	if err != nil {
		return nil, err
	}
	defer headerFace.Close()

	subHeaderFace, err := loadSummaryImageFontFace(13)
	if err != nil {
		return nil, err
	}
	defer subHeaderFace.Close()

	columnWidths := []int{150, 110, 110, 235, 235}
	topHeaderHeight := 30
	subHeaderHeight := 40
	rowHeight := 35
	padding := 6
	width := sumInts(columnWidths)
	height := topHeaderHeight + subHeaderHeight + len(data.Rows)*rowHeight

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fillRect(img, 0, 0, width, height, color.RGBA{245, 245, 245, 255})

	playerHeaderColor := color.RGBA{236, 236, 236, 255}
	boss1HeaderColor := color.RGBA{237, 214, 214, 255}
	boss2HeaderColor := color.RGBA{242, 245, 219, 255}
	boss3HeaderColor := color.RGBA{219, 236, 245, 255}
	rowAltColor := color.RGBA{229, 229, 229, 255}
	borderColor := color.RGBA{160, 160, 160, 255}
	textColor := color.RGBA{20, 20, 20, 255}

	x := 0
	fillRect(img, x, 0, columnWidths[0], topHeaderHeight+subHeaderHeight, playerHeaderColor)
	drawCenteredText(img, face, data.DateLabel, image.Rect(0, 0, columnWidths[0], topHeaderHeight+subHeaderHeight), textColor)

	x += columnWidths[0]
	fillRect(img, x, 0, columnWidths[1], topHeaderHeight, boss1HeaderColor)
	drawCenteredText(img, headerFace, "一王", image.Rect(x, 0, x+columnWidths[1], topHeaderHeight), textColor)
	fillRect(img, x, topHeaderHeight, columnWidths[1], subHeaderHeight, boss1HeaderColor)
	drawWrappedCenteredText(img, subHeaderFace, data.Boss1Sub, image.Rect(x+padding, topHeaderHeight+padding/2, x+columnWidths[1]-padding, topHeaderHeight+subHeaderHeight-padding/2), textColor)

	x += columnWidths[1]
	fillRect(img, x, 0, columnWidths[2], topHeaderHeight, boss2HeaderColor)
	drawCenteredText(img, headerFace, "二王", image.Rect(x, 0, x+columnWidths[2], topHeaderHeight), textColor)
	fillRect(img, x, topHeaderHeight, columnWidths[2], subHeaderHeight, boss2HeaderColor)
	drawWrappedCenteredText(img, subHeaderFace, data.Boss2Sub, image.Rect(x+padding, topHeaderHeight+padding/2, x+columnWidths[2]-padding, topHeaderHeight+subHeaderHeight-padding/2), textColor)

	x += columnWidths[2]
	boss3Width := columnWidths[3] + columnWidths[4]
	fillRect(img, x, 0, boss3Width, topHeaderHeight, boss3HeaderColor)
	drawCenteredText(img, headerFace, "三王", image.Rect(x, 0, x+boss3Width, topHeaderHeight), textColor)

	subHeaders := []string{data.Boss3SubA, data.Boss3SubB}
	for index, subHeader := range subHeaders {
		colX := x + sumInts(columnWidths[3:3+index])
		fillRect(img, colX, topHeaderHeight, columnWidths[3+index], subHeaderHeight, boss3HeaderColor)
		drawWrappedCenteredText(img, subHeaderFace, subHeader, image.Rect(colX+padding, topHeaderHeight+padding/2, colX+columnWidths[3+index]-padding, topHeaderHeight+subHeaderHeight-padding/2), textColor)
	}

	for rowIndex, row := range data.Rows {
		y := topHeaderHeight + subHeaderHeight + rowIndex*rowHeight
		if rowIndex%2 == 1 {
			fillRect(img, 0, y, width, rowHeight, rowAltColor)
		}

		values := []string{row.Player, row.Boss1, row.Boss2, row.Boss3A, row.Boss3B}
		cellX := 0
		for colIndex, value := range values {
			drawWrappedCenteredText(img, face, value, image.Rect(cellX+padding, y+padding/2, cellX+columnWidths[colIndex]-padding, y+rowHeight-padding/2), textColor)
			cellX += columnWidths[colIndex]
		}
	}

	drawVerticalGridLines(img, columnWidths, topHeaderHeight, height, borderColor)
	drawHorizontalGridLines(img, columnWidths[0], topHeaderHeight, subHeaderHeight, rowHeight, len(data.Rows), width, borderColor)

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func loadSummaryImageFontFace(size float64) (font.Face, error) {
	fontData, err := os.ReadFile("C:\\Windows\\Fonts\\msjh.ttc")
	if err != nil {
		return nil, err
	}

	collection, err := opentype.ParseCollection(fontData)
	if err != nil {
		return nil, err
	}

	summaryFont, err := collection.Font(0)
	if err != nil {
		return nil, err
	}

	return opentype.NewFace(summaryFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func fillRect(img *image.RGBA, x int, y int, width int, height int, c color.Color) {
	draw.Draw(img, image.Rect(x, y, x+width, y+height), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawCenteredText(img *image.RGBA, face font.Face, text string, rect image.Rectangle, textColor color.Color) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	textWidth := drawer.MeasureString(text).Round()
	metrics := face.Metrics()
	textHeight := (metrics.Ascent + metrics.Descent).Round()
	x := rect.Min.X + (rect.Dx()-textWidth)/2
	y := rect.Min.Y + (rect.Dy()-textHeight)/2 + metrics.Ascent.Round()
	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(text)
}

func drawWrappedText(img *image.RGBA, face font.Face, text string, rect image.Rectangle, textColor color.Color) {
	if text == "" {
		return
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	lines := wrapTextToWidth(drawer, text, rect.Dx())
	lineHeight := (face.Metrics().Ascent + face.Metrics().Descent).Round()
	y := rect.Min.Y + face.Metrics().Ascent.Round()
	for _, line := range lines {
		if y > rect.Max.Y {
			break
		}
		drawer.Dot = fixed.P(rect.Min.X, y)
		drawer.DrawString(line)
		y += lineHeight
	}
}

func drawWrappedCenteredText(img *image.RGBA, face font.Face, text string, rect image.Rectangle, textColor color.Color) {
	if text == "" {
		return
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	lines := wrapTextToWidth(drawer, text, rect.Dx())
	lineHeight := (face.Metrics().Ascent + face.Metrics().Descent).Round()
	totalHeight := len(lines) * lineHeight
	startY := rect.Min.Y + (rect.Dy()-totalHeight)/2 + face.Metrics().Ascent.Round()

	for index, line := range lines {
		y := startY + index*lineHeight
		if y > rect.Max.Y {
			break
		}
		lineWidth := drawer.MeasureString(line).Round()
		x := rect.Min.X + (rect.Dx()-lineWidth)/2
		drawer.Dot = fixed.P(x, y)
		drawer.DrawString(line)
	}
}

func wrapTextToWidth(drawer *font.Drawer, text string, maxWidth int) []string {
	paragraphs := strings.Split(text, "\n")
	var lines []string

	for _, paragraph := range paragraphs {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		var current []rune
		for _, r := range []rune(paragraph) {
			candidate := string(append(current, r))
			if drawer.MeasureString(candidate).Round() <= maxWidth || len(current) == 0 {
				current = append(current, r)
				continue
			}
			lines = append(lines, string(current))
			current = []rune{r}
		}
		if len(current) > 0 {
			lines = append(lines, string(current))
		}
	}

	return lines
}

func drawVerticalGridLines(img *image.RGBA, columnWidths []int, topHeaderHeight int, height int, c color.Color) {
	x := 0
	for index, width := range columnWidths {
		startY := 0
		if index == len(columnWidths)-2 {
			startY = topHeaderHeight
		}
		drawLine(img, x, startY, x, height, c)
		x += width
	}
	drawLine(img, x-1, 0, x-1, height, c)
}

func drawHorizontalGridLines(img *image.RGBA, firstColumnWidth int, topHeaderHeight int, subHeaderHeight int, rowHeight int, rowCount int, width int, c color.Color) {
	drawLine(img, 0, 0, width, 0, c)
	drawLine(img, firstColumnWidth, topHeaderHeight, width, topHeaderHeight, c)
	drawLine(img, 0, topHeaderHeight+subHeaderHeight, width, topHeaderHeight+subHeaderHeight, c)
	for rowIndex := 1; rowIndex <= rowCount; rowIndex++ {
		y := topHeaderHeight + subHeaderHeight + rowIndex*rowHeight
		drawLine(img, 0, y, width, y, c)
	}
}

func drawLine(img *image.RGBA, x1 int, y1 int, x2 int, y2 int, c color.Color) {
	if x1 == x2 {
		if y2 < y1 {
			y1, y2 = y2, y1
		}
		for y := y1; y < y2; y++ {
			img.Set(x1, y, c)
		}
		return
	}

	if y1 == y2 {
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		for x := x1; x < x2; x++ {
			img.Set(x, y1, c)
		}
	}
}

func sumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
