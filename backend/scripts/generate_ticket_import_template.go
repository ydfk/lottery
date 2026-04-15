package main

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	file := excelize.NewFile()
	defer file.Close()

	createImportSheet(file)
	createReadmeSheet(file)

	file.DeleteSheet("Sheet1")
	sheetIndex, err := file.GetSheetIndex("批量导入模板")
	if err != nil {
		panic(err)
	}
	file.SetActiveSheet(sheetIndex)

	outputPath := filepath.Join("..", "docs", "ticket-import-template.xlsx")
	if err := file.SaveAs(outputPath); err != nil {
		panic(err)
	}

	fmt.Println(outputPath)
}

func createImportSheet(file *excelize.File) {
	sheet := "批量导入模板"
	file.NewSheet(sheet)

	headers := []string{
		"彩票类型",
		"期号",
		"开奖日期",
		"购买时间",
		"红球",
		"蓝球",
		"倍数",
		"追加",
		"金额",
		"备注",
		"图片名",
	}
	rows := [][]string{
		headers,
		{"福彩双色球", "2026031", "2026-03-22", "2026-03-20 19:35", "02,06,10,13,22,31", "04", "2", "否", "4", "同一期同一张票第1注", "ssq-001.jpg"},
		{"福彩双色球", "2026031", "2026-03-22", "2026-03-20 19:35", "03,09,15,19,25,30", "10", "2", "否", "4", "同一期同一张票第2注，会合并到上一行购买记录", "ssq-001.jpg"},
		{"体彩大乐透", "2026030", "2026-03-23", "2026-03-22 18:20", "03,11,18,26,32", "04,09", "2", "是", "6", "大乐透示例", "dlt-001.jpg"},
	}
	writeRows(file, sheet, rows)
	styleSheet(file, sheet, len(headers))
}

func createReadmeSheet(file *excelize.File) {
	sheet := "填写说明"
	file.NewSheet(sheet)

	rows := [][]string{
		{"说明项", "内容"},
		{"用途", "用于批量导入历史票据，可不带图片，也可配合图片 ZIP 一起导入。"},
		{"导入规则", "一行就是一注号码，不再区分单注模板和多注模板。"},
		{"合并规则", "同一用户、同一彩票类型、同一期号的多行，会自动合并为一次购买记录。"},
		{"彩票类型", "支持：福彩双色球 / 体彩大乐透，也支持 ssq / dlt。"},
		{"开奖日期", "建议格式：2026-03-22。"},
		{"购买时间", "建议格式：2026-03-20 19:35 或 RFC3339。"},
		{"图片名", "如果同时上传图片 ZIP，这里填 ZIP 内文件名，例如 ssq-001.jpg。"},
		{"追加列", "支持：是/否、true/false、1/0、追加/不追加。"},
		{"推荐关联", "无需手填推荐ID。系统会按彩票类型、期号和号码自动匹配推荐；只比较红球和蓝球，不比较倍数和追加。"},
		{"号码匹配规则", "如果导入记录包含某条推荐的全部号码，即使购买记录里还有其他号码，也会自动关联这条推荐。"},
	}
	writeRows(file, sheet, rows)
	styleSheet(file, sheet, 2)
}

func writeRows(file *excelize.File, sheet string, rows [][]string) {
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				panic(err)
			}
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				panic(err)
			}
		}
	}
}

func styleSheet(file *excelize.File, sheet string, columnCount int) {
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		panic(err)
	}

	bodyStyle, err := file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		panic(err)
	}

	lastColumn, err := excelize.ColumnNumberToName(columnCount)
	if err != nil {
		panic(err)
	}

	if err := file.SetCellStyle(sheet, "A1", lastColumn+"1", headerStyle); err != nil {
		panic(err)
	}
	if err := file.SetCellStyle(sheet, "A2", lastColumn+"200", bodyStyle); err != nil {
		panic(err)
	}
	if err := file.SetRowHeight(sheet, 1, 24); err != nil {
		panic(err)
	}
	if err := file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		panic(err)
	}

	widths := map[string]float64{
		"A": 16,
		"B": 12,
		"C": 14,
		"D": 18,
		"E": 28,
		"F": 16,
		"G": 10,
		"H": 10,
		"I": 10,
		"J": 24,
		"K": 18,
	}
	for column, width := range widths {
		if err := file.SetColWidth(sheet, column, column, width); err != nil {
			panic(err)
		}
	}
}
