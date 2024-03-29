package xlsx

import . "gopkg.in/check.v1"

type GoogleDocsExcelSuite struct{}

var _ = Suite(&GoogleDocsExcelSuite{})

// Test that we can successfully read an XLSX file generated by
// Google Docs.
func (g *GoogleDocsExcelSuite) TestGoogleDocsExcel(c *C) {
	xlsxFile, err := OpenFile("./testdocs/googleDocsTest.xlsx")
	c.Assert(err, IsNil)
	c.Assert(xlsxFile, NotNil)
}

type MacExcelSuite struct{}

var _ = Suite(&MacExcelSuite{})

// Test that we can successfully read an XLSX file generated by
// Microsoft Excel for Mac.  In particular this requires that we
// respect the contents of workbook.xml.rels, which maps the sheet IDs
// to their internal file names.
func (m *MacExcelSuite) TestMacExcel(c *C) {
	xlsxFile, err := OpenFile("./testdocs/macExcelTest.xlsx")
	c.Assert(err, IsNil)
	c.Assert(xlsxFile, NotNil)
	s := xlsxFile.Sheet["普通技能"].Cell(0, 0).String()
	c.Assert(s, Equals, "编号")
}

type MacNumbersSuite struct{}

var _ = Suite(&MacNumbersSuite{})

// Test that we can successfully read an XLSX file generated by
// Numbers for Mac.
func (m *MacNumbersSuite) TestMacNumbers(c *C) {
	xlsxFile, err := OpenFile("./testdocs/macNumbersTest.xlsx")
	c.Assert(err, IsNil)
	c.Assert(xlsxFile, NotNil)
	sheet, ok := xlsxFile.Sheet["主动技能"]
	c.Assert(ok, Equals, true)
	s := sheet.Cell(0, 0).String()
	c.Assert(s, Equals, "编号")
}

type WpsBlankLineSuite struct{}

var _ = Suite(&WorksheetSuite{})

// Test that we can successfully read an XLSX file generated by
// Wps on windows. you can download it freely from http://www.wps.cn/
func (w *WpsBlankLineSuite) TestWpsBlankLine(c *C) {
	xlsxFile, err := OpenFile("./testdocs/wpsBlankLineTest.xlsx")
	c.Assert(err, IsNil)
	c.Assert(xlsxFile, NotNil)
	sheet := xlsxFile.Sheet["Sheet1"]
	row := sheet.Rows[0]
	cell := row.Cells[0]
	s := cell.String()
	expected := "编号"
	c.Assert(s, Equals, expected)

	row = sheet.Rows[2]
	cell = row.Cells[0]
	s = cell.String()
	c.Assert(s, Equals, expected)

	row = sheet.Rows[4]
	cell = row.Cells[1]
	s = cell.String()
	c.Assert(s, Equals, "")

	s = sheet.Rows[4].Cells[2].String()
	c.Assert(s, Equals, expected)
}
