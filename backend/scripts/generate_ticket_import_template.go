package main

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	file := excelize.NewFile()
	defer file.Close()

	createSingleEntrySheet(file)
	createMultiEntrySheet(file)
	createReadmeSheet(file)

	file.DeleteSheet("Sheet1")
	sheetIndex, err := file.GetSheetIndex("单注分列模板")
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

func createSingleEntrySheet(file *excelize.File) {
	sheet := "单注分列模板"
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
		"推荐ID",
	}
	rows := [][]string{
		headers,
		{"福彩双色球", "2026031", "2026-03-22", "2026-03-20 19:35", "02,06,10,13,22,31", "04", "2", "否", "4", "历史补录示例", "ssq-001.jpg", ""},
		{"体彩大乐透", "2026030", "2026-03-23", "2026-03-22 18:20", "03,11,18,26,32", "04,09", "2", "是", "6", "追加示例", "dlt-001.jpg", ""},
	}
	writeRows(file, sheet, rows)
	styleSheet(file, sheet, len(headers))
}

func createMultiEntrySheet(file *excelize.File) {
	sheet := "多注整列模板"
	file.NewSheet(sheet)

	headers := []string{
		"彩票类型",
		"期号",
		"开奖日期",
		"购买时间",
		"号码",
		"金额",
		"备注",
		"图片名",
		"推荐ID",
	}
	rows := [][]string{
		headers,
		{
			"福彩双色球",
			"2026031",
			"2026-03-22",
			"2026-03-20 19:35",
			"02,06,10,13,22,31+04 (2)\n03,09,15,19,25,30+10 (2)",
			"8",
			"同一张票两注",
			"ssq-batch-001.jpg",
			"",
		},
		{
			"体彩大乐透",
			"2026030",
			"2026-03-23",
			"2026-03-22 18:20",
			"03,11,18,26,32+04,09 追加 (2)\n06,14,21,29,34+02,11 追加 (2)",
			"12",
			"大乐透多注示例",
			"dlt-batch-001.jpg",
			"",
		},
	}
	writeRows(file, sheet, rows)
	styleSheet(file, sheet, len(headers))
	if err := file.SetRowHeight(sheet, 2, 42); err != nil {
		panic(err)
	}
	if err := file.SetRowHeight(sheet, 3, 42); err != nil {
		panic(err)
	}
}

func createReadmeSheet(file *excelize.File) {
	sheet := "填写说明"
	file.NewSheet(sheet)

	rows := [][]string{
		{"说明项", "内容"},
		{"用途", "用于批量导入历史票据，可不带图片，也可配合图片 ZIP 一起导入。"},
		{"单注分列模板", "适合一行只录一注号码；大乐透可单独填写追加。"},
		{"多注整列模板", "适合同一张票里有多注号码，号码列里换行分隔。"},
		{"彩票类型", "支持：福彩双色球 / 体彩大乐透，也支持 ssq / dlt。"},
		{"开奖日期", "建议格式：2026-03-22。"},
		{"购买时间", "建议格式：2026-03-20 19:35 或 RFC3339。"},
		{"图片名", "如果同时上传图片 ZIP，这里填 ZIP 内文件名，例如 ssq-001.jpg。"},
		{"追加列", "支持：是/否、true/false、1/0、追加/不追加。"},
		{"推荐ID", "可选，填入后会把这条购买记录关联到推荐。"},
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
		"L": 38,
	}
	for column, width := range widths {
		if err := file.SetColWidth(sheet, column, column, width); err != nil {
			panic(err)
		}
	}
}
